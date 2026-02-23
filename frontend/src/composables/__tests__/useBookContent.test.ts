import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('@/api/reader', () => ({
  getBookContent: vi.fn(),
  getChapter: vi.fn(),
}))

vi.mock('@/types/reader', () => ({
  defaultSettings: {
    fontSize: 18, fontFamily: 'System', fontWeight: 400, lineHeight: 1.6,
    theme: 'light', paragraphSpacing: 0.5, letterSpacing: 0,
    marginHorizontal: 5, marginVertical: 2, firstLineIndent: 1.5,
    textAlign: 'justify', hyphenation: true, pageAnimation: 'slide', tapZones: 'lrc',
  },
}))

import * as readerApi from '@/api/reader'
import { useBookContent } from '../useBookContent'
import { useReaderStore } from '@/stores/reader'

const mockContent = {
  metadata: { title: 'Test', author: 'Author', cover: null, language: 'en', format: 'fb2' },
  toc: [
    { id: 'ch1', title: 'Chapter 1', level: 0 },
    { id: 'ch2', title: 'Chapter 2', level: 0 },
    { id: 'ch3', title: 'Chapter 3', level: 0 },
  ],
  chapters: ['ch1', 'ch2', 'ch3'],
  totalChapters: 3,
  chapterSizes: { ch1: 100, ch2: 200, ch3: 150 },
}

const mockChapter = {
  id: 'ch1',
  title: 'Chapter 1',
  html: '<p>Content</p>',
}

describe('useBookContent', () => {
  let store: ReturnType<typeof useReaderStore>

  beforeEach(() => {
    setActivePinia(createPinia())
    store = useReaderStore()
    vi.clearAllMocks()
  })

  describe('loadBookContent', () => {
    it('loads content and first chapter on success', async () => {
      const firstChapter = { id: 'ch1', title: 'Chapter 1', html: '<p>First</p>' }
      vi.mocked(readerApi.getBookContent).mockResolvedValue(mockContent)
      vi.mocked(readerApi.getChapter).mockResolvedValue(firstChapter)

      const { loadBookContent } = useBookContent()
      await loadBookContent(42)

      expect(readerApi.getBookContent).toHaveBeenCalledWith(42)
      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch1')
      expect(store.bookContent).toEqual(mockContent)
      expect(store.currentChapterId).toBe('ch1')
      expect(store.currentChapterContent).toEqual(firstChapter)
    })

    it('sets loading true during request and false after', async () => {
      vi.mocked(readerApi.getBookContent).mockResolvedValue(mockContent)
      vi.mocked(readerApi.getChapter).mockResolvedValue(mockChapter)

      const { loadBookContent } = useBookContent()

      expect(store.loading).toBe(false)

      let loadingDuringRequest = false
      vi.mocked(readerApi.getBookContent).mockImplementation(async () => {
        loadingDuringRequest = store.loading
        return mockContent
      })

      await loadBookContent(1)

      expect(loadingDuringRequest).toBe(true)
      expect(store.loading).toBe(false)
    })

    it('handles 404 error', async () => {
      vi.mocked(readerApi.getBookContent).mockRejectedValue({
        response: { status: 404, data: {} },
      })

      const { loadBookContent } = useBookContent()
      await loadBookContent(999)

      expect(store.error).toBe('Книга не найдена')
      expect(store.loading).toBe(false)
    })

    it('handles 415 error', async () => {
      vi.mocked(readerApi.getBookContent).mockRejectedValue({
        response: { status: 415, data: {} },
      })

      const { loadBookContent } = useBookContent()
      await loadBookContent(1)

      expect(store.error).toBe('Формат книги не поддерживается для чтения в браузере')
      expect(store.loading).toBe(false)
    })

    it('handles 422 error', async () => {
      vi.mocked(readerApi.getBookContent).mockRejectedValue({
        response: { status: 422, data: {} },
      })

      const { loadBookContent } = useBookContent()
      await loadBookContent(1)

      expect(store.error).toBe('Файл книги повреждён или имеет некорректный формат')
      expect(store.loading).toBe(false)
    })

    it('handles generic server error with message from response', async () => {
      vi.mocked(readerApi.getBookContent).mockRejectedValue({
        response: { status: 500, data: { message: 'Internal server error' } },
      })

      const { loadBookContent } = useBookContent()
      await loadBookContent(1)

      expect(store.error).toBe('Internal server error')
      expect(store.loading).toBe(false)
    })

    it('handles generic Error object', async () => {
      vi.mocked(readerApi.getBookContent).mockRejectedValue(new Error('Network failure'))

      const { loadBookContent } = useBookContent()
      await loadBookContent(1)

      expect(store.error).toBe('Network failure')
      expect(store.loading).toBe(false)
    })

    it('handles non-Error non-response rejection with fallback message', async () => {
      vi.mocked(readerApi.getBookContent).mockRejectedValue('unexpected')

      const { loadBookContent } = useBookContent()
      await loadBookContent(1)

      expect(store.error).toBe('Не удалось загрузить книгу. Проверьте подключение к сети.')
      expect(store.loading).toBe(false)
    })

    it('does not load first chapter when content has no chapters', async () => {
      const emptyContent = { ...mockContent, chapters: [], totalChapters: 0 }
      vi.mocked(readerApi.getBookContent).mockResolvedValue(emptyContent)

      const { loadBookContent } = useBookContent()
      await loadBookContent(1)

      expect(readerApi.getChapter).not.toHaveBeenCalled()
      expect(store.bookContent).toEqual(emptyContent)
    })

    it('handles generic server error without message, falls back to default', async () => {
      vi.mocked(readerApi.getBookContent).mockRejectedValue({
        response: { status: 500, data: {} },
      })

      const { loadBookContent } = useBookContent()
      await loadBookContent(1)

      expect(store.error).toBe('Не удалось загрузить книгу. Проверьте подключение к сети.')
    })
  })

  describe('loadChapter', () => {
    it('loads chapter and sets it in store', async () => {
      const chapter = { id: 'ch2', title: 'Chapter 2', html: '<p>Two</p>' }
      vi.mocked(readerApi.getChapter).mockResolvedValue(chapter)

      const { loadChapter } = useBookContent()
      await loadChapter(42, 'ch2')

      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch2')
      expect(store.currentChapterId).toBe('ch2')
      expect(store.currentChapterContent).toEqual(chapter)
    })

    it('handles error when loading chapter', async () => {
      vi.mocked(readerApi.getChapter).mockRejectedValue({
        response: { status: 404, data: {} },
      })

      const { loadChapter } = useBookContent()
      await loadChapter(1, 'missing')

      expect(store.error).toBe('Книга не найдена')
    })
  })

  describe('navigateToChapter', () => {
    it('loads chapter, closes TOC, and manages loading state', async () => {
      const chapter = { id: 'ch2', title: 'Chapter 2', html: '<p>Two</p>' }
      vi.mocked(readerApi.getChapter).mockResolvedValue(chapter)
      store.tocVisible = true

      const { navigateToChapter } = useBookContent()
      await navigateToChapter(42, 'ch2')

      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch2')
      expect(store.currentChapterId).toBe('ch2')
      expect(store.tocVisible).toBe(false)
      expect(store.loading).toBe(false)
    })

    it('sets loading true during navigation', async () => {
      let loadingDuringNav = false
      vi.mocked(readerApi.getChapter).mockImplementation(async () => {
        loadingDuringNav = store.loading
        return mockChapter
      })

      const { navigateToChapter } = useBookContent()
      await navigateToChapter(1, 'ch1')

      expect(loadingDuringNav).toBe(true)
      expect(store.loading).toBe(false)
    })

    it('handles error during navigation and still closes TOC', async () => {
      vi.mocked(readerApi.getChapter).mockRejectedValue(new Error('fail'))
      store.tocVisible = true

      const { navigateToChapter } = useBookContent()
      await navigateToChapter(1, 'ch1')

      expect(store.error).toBe('fail')
      expect(store.loading).toBe(false)
      // loadChapter catches error internally and does not re-throw,
      // so navigateToChapter proceeds to set tocVisible = false
      expect(store.tocVisible).toBe(false)
    })
  })

  describe('nextChapter', () => {
    it('advances to next chapter and sets direction to forward', async () => {
      store.setBookContent(mockContent)
      store.setChapter(mockChapter) // current = ch1, index = 0

      const nextChapterData = { id: 'ch2', title: 'Chapter 2', html: '<p>Two</p>' }
      vi.mocked(readerApi.getChapter).mockResolvedValue(nextChapterData)

      const { nextChapter } = useBookContent()
      await nextChapter(42)

      expect(store.navigationDirection).toBe('forward')
      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch2')
      expect(store.currentChapterId).toBe('ch2')
    })

    it('does nothing when there is no next chapter', async () => {
      store.setBookContent(mockContent)
      store.setChapter({ id: 'ch3', title: 'Chapter 3', html: '<p>Three</p>' }) // last chapter

      const { nextChapter } = useBookContent()
      await nextChapter(42)

      expect(readerApi.getChapter).not.toHaveBeenCalled()
      expect(store.currentChapterId).toBe('ch3')
    })

    it('does nothing when bookContent is null', async () => {
      const { nextChapter } = useBookContent()
      await nextChapter(42)

      expect(readerApi.getChapter).not.toHaveBeenCalled()
    })
  })

  describe('prevChapter', () => {
    it('goes to previous chapter and sets direction to backward', async () => {
      store.setBookContent(mockContent)
      store.setChapter({ id: 'ch2', title: 'Chapter 2', html: '<p>Two</p>' }) // index = 1

      const prevChapterData = { id: 'ch1', title: 'Chapter 1', html: '<p>One</p>' }
      vi.mocked(readerApi.getChapter).mockResolvedValue(prevChapterData)

      const { prevChapter } = useBookContent()
      await prevChapter(42)

      expect(store.navigationDirection).toBe('backward')
      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch1')
      expect(store.currentChapterId).toBe('ch1')
    })

    it('does nothing when there is no previous chapter', async () => {
      store.setBookContent(mockContent)
      store.setChapter(mockChapter) // ch1, index = 0

      const { prevChapter } = useBookContent()
      await prevChapter(42)

      expect(readerApi.getChapter).not.toHaveBeenCalled()
      expect(store.currentChapterId).toBe('ch1')
    })

    it('does nothing when bookContent is null', async () => {
      const { prevChapter } = useBookContent()
      await prevChapter(42)

      expect(readerApi.getChapter).not.toHaveBeenCalled()
    })
  })

  describe('prefetchAdjacentChapters', () => {
    it('prefetches next and previous chapters', () => {
      store.setBookContent(mockContent)
      store.setChapter({ id: 'ch2', title: 'Chapter 2', html: '<p>Two</p>' }) // index = 1

      vi.mocked(readerApi.getChapter).mockResolvedValue(mockChapter)

      const { prefetchAdjacentChapters } = useBookContent()
      prefetchAdjacentChapters(42)

      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch3') // next
      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch1') // prev
      expect(readerApi.getChapter).toHaveBeenCalledTimes(2)
    })

    it('prefetches only next chapter when at first chapter', () => {
      store.setBookContent(mockContent)
      store.setChapter(mockChapter) // ch1, index = 0

      vi.mocked(readerApi.getChapter).mockResolvedValue(mockChapter)

      const { prefetchAdjacentChapters } = useBookContent()
      prefetchAdjacentChapters(42)

      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch2')
      expect(readerApi.getChapter).toHaveBeenCalledTimes(1)
    })

    it('prefetches only previous chapter when at last chapter', () => {
      store.setBookContent(mockContent)
      store.setChapter({ id: 'ch3', title: 'Chapter 3', html: '<p>Three</p>' }) // index = 2

      vi.mocked(readerApi.getChapter).mockResolvedValue(mockChapter)

      const { prefetchAdjacentChapters } = useBookContent()
      prefetchAdjacentChapters(42)

      expect(readerApi.getChapter).toHaveBeenCalledWith(42, 'ch2')
      expect(readerApi.getChapter).toHaveBeenCalledTimes(1)
    })

    it('does nothing when bookContent is null', () => {
      const { prefetchAdjacentChapters } = useBookContent()
      prefetchAdjacentChapters(42)

      expect(readerApi.getChapter).not.toHaveBeenCalled()
    })

    it('silently ignores prefetch errors', async () => {
      store.setBookContent(mockContent)
      store.setChapter({ id: 'ch2', title: 'Chapter 2', html: '<p>Two</p>' })

      vi.mocked(readerApi.getChapter).mockRejectedValue(new Error('prefetch fail'))

      const { prefetchAdjacentChapters } = useBookContent()
      prefetchAdjacentChapters(42)

      // Should not throw; errors are caught by .catch(() => {})
      await vi.waitFor(() => {
        expect(readerApi.getChapter).toHaveBeenCalledTimes(2)
      })
      expect(store.error).toBeNull()
    })
  })
})
