import { defineStore } from 'pinia'
import { ref, computed, triggerRef } from 'vue'
import type { BookContent, ChapterContent, ReaderSettings } from '@/types/reader'
import { defaultSettings } from '@/types/reader'

export const useReaderStore = defineStore('reader', () => {
  // Book data
  const bookContent = ref<BookContent | null>(null)
  const currentChapterId = ref<string | null>(null)
  const currentChapterContent = ref<ChapterContent | null>(null)

  // Pagination (current chapter)
  const currentPage = ref(1)
  const totalPages = ref(1)

  // Per-chapter page counts for book-level progress
  const chapterPageCounts = ref<Map<string, number>>(new Map())
  // Chapters that have been actually rendered and measured (vs estimated)
  const measuredChapters = ref<Set<string>>(new Set())

  // UI state
  const loading = ref(false)
  const error = ref<string | null>(null)
  const tocVisible = ref(false)
  const uiVisible = ref(true)
  const settingsVisible = ref(false)

  // Navigation direction for chapter transitions ('backward' → open last page)
  const navigationDirection = ref<'forward' | 'backward'>('forward')

  // Settings
  const settings = ref<ReaderSettings>({ ...defaultSettings })

  // Computed
  const currentChapterIndex = computed(() => {
    if (!bookContent.value || !currentChapterId.value) return -1
    return bookContent.value.chapters.indexOf(currentChapterId.value)
  })

  const hasNextChapter = computed(() => {
    if (!bookContent.value) return false
    return currentChapterIndex.value < bookContent.value.chapters.length - 1
  })

  const hasPrevChapter = computed(() => {
    return currentChapterIndex.value > 0
  })

  const chapterProgress = computed(() => {
    if (totalPages.value <= 1) return 100
    return Math.round(((currentPage.value - 1) / (totalPages.value - 1)) * 1000) / 10
  })

  // Book-level page tracking
  const bookTotalPages = computed(() => {
    if (!bookContent.value) return 1
    let sum = 0
    for (const chId of bookContent.value.chapters) {
      sum += chapterPageCounts.value.get(chId) ?? 1
    }
    return Math.max(1, sum)
  })

  const bookCurrentPage = computed(() => {
    if (!bookContent.value || currentChapterIndex.value < 0) return 1
    let page = 0
    for (let i = 0; i < currentChapterIndex.value; i++) {
      const chId = bookContent.value.chapters[i]
      page += chapterPageCounts.value.get(chId) ?? 1
    }
    return page + currentPage.value
  })

  const totalProgress = computed(() => {
    if (bookTotalPages.value <= 1) return 100
    return Math.round(((bookCurrentPage.value - 1) / (bookTotalPages.value - 1)) * 1000) / 10
  })

  // Actions
  function setBookContent(content: BookContent) {
    bookContent.value = content
    chapterPageCounts.value = new Map()
    measuredChapters.value = new Set()
    error.value = null
  }

  function setChapter(chapter: ChapterContent) {
    currentChapterId.value = chapter.id
    currentChapterContent.value = chapter
    currentPage.value = 1
  }

  function setPage(page: number) {
    if (page >= 1 && page <= totalPages.value) {
      currentPage.value = page
    }
  }

  function setTotalPages(count: number) {
    const clamped = Math.max(1, count)
    totalPages.value = clamped
    if (currentPage.value > clamped) {
      currentPage.value = clamped
    }
    // Track page count for current chapter (actual measurement)
    if (currentChapterId.value) {
      chapterPageCounts.value.set(currentChapterId.value, clamped)
      measuredChapters.value.add(currentChapterId.value)
    }
    // Re-estimate page counts for unvisited chapters using weighted average
    estimateChapterPages()
    // Trigger reactivity for Map/Set mutations
    triggerRef(chapterPageCounts)
    triggerRef(measuredChapters)
  }

  function estimateChapterPages() {
    const bc = bookContent.value
    if (!bc?.chapterSizes) return

    // Weighted average ratio from all actually measured chapters
    let totalMeasuredPages = 0
    let totalMeasuredSize = 0
    for (const chId of measuredChapters.value) {
      const pages = chapterPageCounts.value.get(chId)
      const size = bc.chapterSizes[chId]
      if (pages && pages > 0 && size && size > 0) {
        totalMeasuredPages += pages
        totalMeasuredSize += size
      }
    }
    if (totalMeasuredSize <= 0 || totalMeasuredPages <= 0) return

    const ratio = totalMeasuredPages / totalMeasuredSize

    // Re-estimate ALL non-measured chapters (overwrite old estimates)
    for (const chId of bc.chapters) {
      if (measuredChapters.value.has(chId)) continue
      const size = bc.chapterSizes[chId]
      if (size && size > 0) {
        chapterPageCounts.value.set(chId, Math.max(1, Math.round(size * ratio)))
      }
    }
  }

  function toggleTOC() {
    tocVisible.value = !tocVisible.value
  }

  function toggleUI() {
    uiVisible.value = !uiVisible.value
  }

  function toggleSettings() {
    settingsVisible.value = !settingsVisible.value
  }

  function updateSettings(partial: Partial<ReaderSettings>) {
    settings.value = { ...settings.value, ...partial }
  }

  function setError(msg: string) {
    error.value = msg
    loading.value = false
  }

  function reset() {
    bookContent.value = null
    currentChapterId.value = null
    currentChapterContent.value = null
    currentPage.value = 1
    totalPages.value = 1
    chapterPageCounts.value = new Map()
    measuredChapters.value = new Set()
    loading.value = false
    error.value = null
    tocVisible.value = false
    uiVisible.value = true
    settingsVisible.value = false
    navigationDirection.value = 'forward'
  }

  return {
    // State
    bookContent,
    currentChapterId,
    currentChapterContent,
    currentPage,
    totalPages,
    chapterPageCounts,
    measuredChapters,
    loading,
    error,
    tocVisible,
    uiVisible,
    settingsVisible,
    navigationDirection,
    settings,

    // Computed
    currentChapterIndex,
    hasNextChapter,
    hasPrevChapter,
    chapterProgress,
    bookTotalPages,
    bookCurrentPage,
    totalProgress,

    // Actions
    setBookContent,
    setChapter,
    setPage,
    setTotalPages,
    toggleTOC,
    toggleUI,
    toggleSettings,
    updateSettings,
    setError,
    reset,
  }
})
