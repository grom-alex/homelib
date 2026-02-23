import { onMounted, onUnmounted } from 'vue'
import { useReaderStore } from '@/stores/reader'
import { getReadingProgress, saveReadingProgress } from '@/api/reader'
import { getAccessToken } from '@/api/client'

const DEBOUNCE_MS = 2000

function getDeviceType(): string {
  const width = window.innerWidth
  if (width < 768) return 'mobile'
  if (width < 1024) return 'tablet'
  return 'desktop'
}

export function useReadingProgress(bookId: number) {
  const store = useReaderStore()
  let pendingSave = false
  let saveTimer: ReturnType<typeof setTimeout> | null = null

  async function loadProgress(): Promise<{ chapterId: string; chapterProgress: number } | null> {
    try {
      const progress = await getReadingProgress(bookId)
      return progress
    } catch {
      return null
    }
  }

  function calculateTotalProgress(): number {
    const content = store.bookContent
    if (!content || !content.chapters.length) return 0

    const chapterIndex = store.currentChapterIndex
    const totalChapters = content.totalChapters
    if (totalChapters === 0) return 0

    const chapterWeight = 100 / totalChapters
    const completedChapters = chapterIndex * chapterWeight
    const currentChapterProgress = store.chapterProgress * chapterWeight / 100

    return Math.min(100, Math.max(0, Math.round(completedChapters + currentChapterProgress)))
  }

  async function doSave() {
    if (!store.currentChapterId) return

    const totalProgress = calculateTotalProgress()
    try {
      await saveReadingProgress(bookId, {
        chapterId: store.currentChapterId,
        chapterProgress: store.chapterProgressInt,
        totalProgress,
        device: getDeviceType(),
      })
    } catch {
      // Ошибки сохранения игнорируются — не блокируем чтение
    }
    pendingSave = false
  }

  function scheduleSave() {
    pendingSave = true
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(doSave, DEBOUNCE_MS)
  }

  function saveNow() {
    if (saveTimer) {
      clearTimeout(saveTimer)
      saveTimer = null
    }
    if (!pendingSave) return

    // Use fetch with keepalive for reliable delivery on page unload.
    // Unlike sendBeacon, fetch supports custom headers (Authorization) and PUT method.
    const totalProgress = calculateTotalProgress()
    const body = JSON.stringify({
      chapterId: store.currentChapterId,
      chapterProgress: store.chapterProgressInt,
      totalProgress,
      device: getDeviceType(),
    })
    const token = getAccessToken()
    fetch(`/api/me/books/${bookId}/progress`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body,
      keepalive: true,
    }).catch(() => {})
    pendingSave = false
  }

  function handleBeforeUnload() {
    saveNow()
  }

  onMounted(() => {
    window.addEventListener('beforeunload', handleBeforeUnload)
  })

  onUnmounted(() => {
    saveNow()
    window.removeEventListener('beforeunload', handleBeforeUnload)
  })

  return {
    loadProgress,
    scheduleSave,
    saveNow,
    calculateTotalProgress,
  }
}
