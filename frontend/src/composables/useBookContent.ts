import { useReaderStore } from '@/stores/reader'
import * as readerApi from '@/api/reader'

export function useBookContent() {
  const store = useReaderStore()

  async function loadBookContent(bookId: number) {
    store.loading = true
    store.error = null

    try {
      const content = await readerApi.getBookContent(bookId)
      store.setBookContent(content)

      // Load first chapter
      if (content.chapters.length > 0) {
        await loadChapter(bookId, content.chapters[0])
      }
    } catch (e: unknown) {
      handleError(e)
    } finally {
      store.loading = false
    }
  }

  async function loadChapter(bookId: number, chapterId: string) {
    try {
      const chapter = await readerApi.getChapter(bookId, chapterId)
      store.setChapter(chapter)
    } catch (e: unknown) {
      handleError(e)
    }
  }

  async function navigateToChapter(bookId: number, chapterId: string) {
    store.loading = true
    try {
      await loadChapter(bookId, chapterId)
      store.tocVisible = false
    } catch (e: unknown) {
      handleError(e)
    } finally {
      store.loading = false
    }
  }

  async function nextChapter(bookId: number) {
    if (!store.bookContent || !store.hasNextChapter) return
    store.navigationDirection = 'forward'
    const nextId = store.bookContent.chapters[store.currentChapterIndex + 1]
    await navigateToChapter(bookId, nextId)
  }

  async function prevChapter(bookId: number) {
    if (!store.bookContent || !store.hasPrevChapter) return
    store.navigationDirection = 'backward'
    const prevId = store.bookContent.chapters[store.currentChapterIndex - 1]
    await navigateToChapter(bookId, prevId)
  }

  function prefetchAdjacentChapters(bookId: number) {
    if (!store.bookContent) return

    const chapters = store.bookContent.chapters
    const idx = store.currentChapterIndex

    // Prefetch next chapter
    if (idx < chapters.length - 1) {
      readerApi.getChapter(bookId, chapters[idx + 1]).catch(() => {})
    }
    // Prefetch prev chapter
    if (idx > 0) {
      readerApi.getChapter(bookId, chapters[idx - 1]).catch(() => {})
    }
  }

  function handleError(e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const resp = (e as { response?: { status?: number; data?: { error?: string; message?: string } } }).response
      if (resp?.status === 404) {
        store.setError('Книга не найдена')
      } else if (resp?.status === 415) {
        store.setError('Формат книги не поддерживается для чтения в браузере')
      } else if (resp?.status === 422) {
        store.setError('Файл книги повреждён или имеет некорректный формат')
      } else {
        store.setError(resp?.data?.message || 'Не удалось загрузить книгу. Проверьте подключение к сети.')
      }
    } else {
      store.setError(e instanceof Error ? e.message : 'Не удалось загрузить книгу. Проверьте подключение к сети.')
    }
  }

  return {
    loadBookContent,
    loadChapter,
    navigateToChapter,
    nextChapter,
    prevChapter,
    prefetchAdjacentChapters,
  }
}
