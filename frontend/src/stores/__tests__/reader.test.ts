import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('@/types/reader', () => ({
  defaultSettings: {
    fontSize: 18,
    fontFamily: 'System',
    fontWeight: 400,
    lineHeight: 1.6,
    theme: 'light',
    paragraphSpacing: 0.5,
    letterSpacing: 0,
    marginHorizontal: 5,
    marginVertical: 2,
    firstLineIndent: 1.5,
    textAlign: 'justify',
    hyphenation: true,
    pageAnimation: 'slide',
    tapZones: 'lrc',
  },
}))

import { useReaderStore } from '../reader'

// Helpers to build test data
function makeBookContent(chapterIds: string[], chapterSizes?: Record<string, number>) {
  return {
    bookId: 42,
    title: 'Test Book',
    chapters: chapterIds,
    toc: chapterIds.map((id) => ({ id, title: `Chapter ${id}`, level: 0 })),
    chapterSizes: chapterSizes ?? {},
  }
}

function makeChapter(id: string, html = '<p>content</p>') {
  return { id, html }
}

describe('reader store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  // -------------------------------------------------------------------
  // 1. Initial state
  // -------------------------------------------------------------------
  describe('initial state', () => {
    it('has correct defaults', () => {
      const store = useReaderStore()
      expect(store.bookContent).toBeNull()
      expect(store.currentChapterId).toBeNull()
      expect(store.currentChapterContent).toBeNull()
      expect(store.currentPage).toBe(1)
      expect(store.totalPages).toBe(1)
      expect(store.chapterPageCounts).toEqual(new Map())
      expect(store.loading).toBe(false)
      expect(store.error).toBeNull()
      expect(store.tocVisible).toBe(false)
      expect(store.uiVisible).toBe(true)
      expect(store.settingsVisible).toBe(false)
      expect(store.navigationDirection).toBe('forward')
    })

    it('initializes settings from defaultSettings', () => {
      const store = useReaderStore()
      expect(store.settings).toEqual({
        fontSize: 18,
        fontFamily: 'System',
        fontWeight: 400,
        lineHeight: 1.6,
        theme: 'light',
        paragraphSpacing: 0.5,
        letterSpacing: 0,
        marginHorizontal: 5,
        marginVertical: 2,
        firstLineIndent: 1.5,
        textAlign: 'justify',
        hyphenation: true,
        pageAnimation: 'slide',
        tapZones: 'lrc',
      })
    })
  })

  // -------------------------------------------------------------------
  // 2. setBookContent
  // -------------------------------------------------------------------
  describe('setBookContent', () => {
    it('sets book content', () => {
      const store = useReaderStore()
      const bc = makeBookContent(['ch1', 'ch2'])
      store.setBookContent(bc)

      expect(store.bookContent).toEqual(bc)
    })

    it('clears chapterPageCounts', () => {
      const store = useReaderStore()
      // Pre-populate some counts
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(5)
      expect(store.chapterPageCounts.size).toBeGreaterThan(0)

      // Setting new book content should clear counts
      store.setBookContent(makeBookContent(['ch2', 'ch3']))
      expect(store.chapterPageCounts).toEqual(new Map())
    })

    it('clears error', () => {
      const store = useReaderStore()
      store.setError('something went wrong')
      expect(store.error).toBe('something went wrong')

      store.setBookContent(makeBookContent(['ch1']))
      expect(store.error).toBeNull()
    })
  })

  // -------------------------------------------------------------------
  // 3. setChapter
  // -------------------------------------------------------------------
  describe('setChapter', () => {
    it('sets chapter id and content', () => {
      const store = useReaderStore()
      const chapter = makeChapter('ch1', '<p>hello</p>')
      store.setChapter(chapter)

      expect(store.currentChapterId).toBe('ch1')
      expect(store.currentChapterContent).toEqual(chapter)
    })

    it('resets currentPage to 1', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)
      store.setPage(7)
      expect(store.currentPage).toBe(7)

      store.setChapter(makeChapter('ch2'))
      expect(store.currentPage).toBe(1)
    })
  })

  // -------------------------------------------------------------------
  // 4. setPage
  // -------------------------------------------------------------------
  describe('setPage', () => {
    it('sets page when within range', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(5)

      store.setPage(3)
      expect(store.currentPage).toBe(3)

      store.setPage(1)
      expect(store.currentPage).toBe(1)

      store.setPage(5)
      expect(store.currentPage).toBe(5)
    })

    it('ignores page below 1', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(5)
      store.setPage(3)

      store.setPage(0)
      expect(store.currentPage).toBe(3)

      store.setPage(-1)
      expect(store.currentPage).toBe(3)
    })

    it('ignores page above totalPages', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(5)
      store.setPage(3)

      store.setPage(6)
      expect(store.currentPage).toBe(3)

      store.setPage(100)
      expect(store.currentPage).toBe(3)
    })
  })

  // -------------------------------------------------------------------
  // 5. setTotalPages
  // -------------------------------------------------------------------
  describe('setTotalPages', () => {
    it('clamps to minimum of 1', () => {
      const store = useReaderStore()
      store.setTotalPages(0)
      expect(store.totalPages).toBe(1)

      store.setTotalPages(-5)
      expect(store.totalPages).toBe(1)
    })

    it('sets totalPages', () => {
      const store = useReaderStore()
      store.setTotalPages(10)
      expect(store.totalPages).toBe(10)
    })

    it('adjusts currentPage when it exceeds new totalPages', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)
      store.setPage(8)
      expect(store.currentPage).toBe(8)

      store.setTotalPages(5)
      expect(store.currentPage).toBe(5)
    })

    it('tracks page count in chapterPageCounts for current chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(7)

      expect(store.chapterPageCounts.get('ch1')).toBe(7)
    })

    it('calls estimateChapterPages when chapterSizes are available', () => {
      const store = useReaderStore()
      // ch1 has size 1000, ch2 has size 2000 -- ch2 should be estimated at ~double
      store.setBookContent(makeBookContent(['ch1', 'ch2'], { ch1: 1000, ch2: 2000 }))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)

      // ch1 measured at 10 pages for 1000 chars -> ratio=0.01
      // ch2 estimated: 2000 * 0.01 = 20
      expect(store.chapterPageCounts.get('ch1')).toBe(10)
      expect(store.chapterPageCounts.get('ch2')).toBe(20)
    })

    it('does not overwrite already measured chapters during estimation', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3'], { ch1: 1000, ch2: 500, ch3: 2000 }))

      // Visit ch2 first, measure it
      store.setChapter(makeChapter('ch2'))
      store.setTotalPages(5)
      expect(store.chapterPageCounts.get('ch2')).toBe(5)

      // Now visit ch1, measure it
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)

      // ch2 should still be 5 (measured), not re-estimated
      expect(store.chapterPageCounts.get('ch2')).toBe(5)
      // ch3 should be estimated from ch1: ratio = 10/1000 = 0.01, ch3 = 2000*0.01 = 20
      expect(store.chapterPageCounts.get('ch3')).toBe(20)
    })

    it('handles zero/negative chapterSizes gracefully', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3'], { ch1: 1000, ch2: 0, ch3: -100 }))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)

      // ch2/ch3 have invalid sizes, should not be estimated
      expect(store.chapterPageCounts.has('ch2')).toBe(false)
      expect(store.chapterPageCounts.has('ch3')).toBe(false)
    })

    it('estimation clamps to minimum of 1 page per chapter', () => {
      const store = useReaderStore()
      // Very small chapter: size 1 -> ratio from ch1 = 10/10000 = 0.001, 1*0.001 = ~0 -> clamped to 1
      store.setBookContent(makeBookContent(['ch1', 'ch2'], { ch1: 10000, ch2: 1 }))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)

      expect(store.chapterPageCounts.get('ch2')).toBe(1)
    })
  })

  // -------------------------------------------------------------------
  // 6. currentChapterIndex, hasNextChapter, hasPrevChapter
  // -------------------------------------------------------------------
  describe('currentChapterIndex', () => {
    it('returns -1 when no book content', () => {
      const store = useReaderStore()
      expect(store.currentChapterIndex).toBe(-1)
    })

    it('returns -1 when no chapter selected', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2']))
      expect(store.currentChapterIndex).toBe(-1)
    })

    it('returns correct index', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))
      store.setChapter(makeChapter('ch2'))
      expect(store.currentChapterIndex).toBe(1)
    })
  })

  describe('hasNextChapter', () => {
    it('returns false when no book content', () => {
      const store = useReaderStore()
      expect(store.hasNextChapter).toBe(false)
    })

    it('returns true when not on last chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))
      store.setChapter(makeChapter('ch1'))
      expect(store.hasNextChapter).toBe(true)
    })

    it('returns false when on last chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))
      store.setChapter(makeChapter('ch3'))
      expect(store.hasNextChapter).toBe(false)
    })
  })

  describe('hasPrevChapter', () => {
    it('returns false when no chapter selected', () => {
      const store = useReaderStore()
      expect(store.hasPrevChapter).toBe(false)
    })

    it('returns false when on first chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))
      store.setChapter(makeChapter('ch1'))
      expect(store.hasPrevChapter).toBe(false)
    })

    it('returns true when not on first chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))
      store.setChapter(makeChapter('ch2'))
      expect(store.hasPrevChapter).toBe(true)
    })
  })

  // -------------------------------------------------------------------
  // 7. chapterProgress
  // -------------------------------------------------------------------
  describe('chapterProgress', () => {
    it('returns 100 when totalPages is 1', () => {
      const store = useReaderStore()
      expect(store.totalPages).toBe(1)
      expect(store.chapterProgress).toBe(100)
    })

    it('returns 0 on first page of multi-page chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)
      store.setPage(1)
      expect(store.chapterProgress).toBe(0)
    })

    it('returns 100 on last page of multi-page chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)
      store.setPage(10)
      expect(store.chapterProgress).toBe(100)
    })

    it('returns 50 at midpoint', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(11) // pages 1..11, midpoint is page 6 -> (6-1)/(11-1) = 50%
      store.setPage(6)
      expect(store.chapterProgress).toBe(50)
    })

    it('rounds to nearest integer', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(3) // page 2: (2-1)/(3-1) = 0.5 -> 50
      store.setPage(2)
      expect(store.chapterProgress).toBe(50)
    })
  })

  // -------------------------------------------------------------------
  // 8. bookTotalPages and bookCurrentPage
  // -------------------------------------------------------------------
  describe('bookTotalPages', () => {
    it('returns 1 when no book content', () => {
      const store = useReaderStore()
      expect(store.bookTotalPages).toBe(1)
    })

    it('defaults to 1 per chapter when no page counts are known', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))
      // No chapters visited, so each defaults to 1
      expect(store.bookTotalPages).toBe(3)
    })

    it('sums known and default page counts', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)

      // ch1=10, ch2=1 (default), ch3=1 (default) -> 12
      // (unless estimation changes ch2/ch3 -- no chapterSizes so no estimation)
      expect(store.bookTotalPages).toBe(12)
    })

    it('uses estimated page counts when available', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2'], { ch1: 500, ch2: 1000 }))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(5)

      // ch1=5 (measured), ch2=10 (estimated: 1000 * (5/500))
      expect(store.bookTotalPages).toBe(15)
    })
  })

  describe('bookCurrentPage', () => {
    it('returns 1 when no book content', () => {
      const store = useReaderStore()
      expect(store.bookCurrentPage).toBe(1)
    })

    it('returns 1 when no chapter selected', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2']))
      expect(store.bookCurrentPage).toBe(1)
    })

    it('equals currentPage when on first chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)
      store.setPage(5)

      expect(store.bookCurrentPage).toBe(5)
    })

    it('sums previous chapter pages plus currentPage', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))

      // Visit ch1 and measure it
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)

      // Visit ch2 and measure it
      store.setChapter(makeChapter('ch2'))
      store.setTotalPages(8)
      store.setPage(3)

      // bookCurrentPage = ch1(10) + currentPage(3) = 13
      expect(store.bookCurrentPage).toBe(13)
    })

    it('uses default of 1 for unmeasured previous chapters', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))

      // Jump directly to ch3 without visiting earlier ones
      store.setChapter(makeChapter('ch3'))
      store.setTotalPages(5)
      store.setPage(3)

      // bookCurrentPage = ch1(1 default) + ch2(1 default) + 3 = 5
      // (no chapterSizes, so no estimation changes ch1/ch2)
      expect(store.bookCurrentPage).toBe(5)
    })
  })

  // -------------------------------------------------------------------
  // 9. totalProgress
  // -------------------------------------------------------------------
  describe('totalProgress', () => {
    it('returns 100 when bookTotalPages is 1', () => {
      const store = useReaderStore()
      // No book content -> bookTotalPages=1
      expect(store.totalProgress).toBe(100)
    })

    it('returns 0 on first page of first chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(5)
      store.setPage(1)

      // bookCurrentPage=1, bookTotalPages=5+1=6
      // (1-1)/(6-1) = 0/5 = 0
      expect(store.totalProgress).toBe(0)
    })

    it('returns 100 on last page of last chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2']))

      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(5)

      store.setChapter(makeChapter('ch2'))
      store.setTotalPages(3)
      store.setPage(3)

      // bookCurrentPage = ch1(5) + 3 = 8, bookTotalPages = 5+3 = 8
      // (8-1)/(8-1) = 1 -> 100
      expect(store.totalProgress).toBe(100)
    })

    it('calculates intermediate progress correctly', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2', 'ch3']))

      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)

      store.setChapter(makeChapter('ch2'))
      store.setTotalPages(10)

      store.setChapter(makeChapter('ch3'))
      store.setTotalPages(10)

      // All chapters have 10 pages, totalBookPages = 30
      // Go to ch2, page 1 -> bookCurrentPage = 10 + 1 = 11
      store.setChapter(makeChapter('ch2'))
      store.setPage(1)
      // (11-1)/(30-1) = 10/29 ~ 34.48 -> 34
      expect(store.totalProgress).toBe(34)
    })
  })

  // -------------------------------------------------------------------
  // 10. Toggle actions
  // -------------------------------------------------------------------
  describe('toggleTOC', () => {
    it('toggles tocVisible', () => {
      const store = useReaderStore()
      expect(store.tocVisible).toBe(false)
      store.toggleTOC()
      expect(store.tocVisible).toBe(true)
      store.toggleTOC()
      expect(store.tocVisible).toBe(false)
    })
  })

  describe('toggleUI', () => {
    it('toggles uiVisible', () => {
      const store = useReaderStore()
      expect(store.uiVisible).toBe(true)
      store.toggleUI()
      expect(store.uiVisible).toBe(false)
      store.toggleUI()
      expect(store.uiVisible).toBe(true)
    })
  })

  describe('toggleSettings', () => {
    it('toggles settingsVisible', () => {
      const store = useReaderStore()
      expect(store.settingsVisible).toBe(false)
      store.toggleSettings()
      expect(store.settingsVisible).toBe(true)
      store.toggleSettings()
      expect(store.settingsVisible).toBe(false)
    })
  })

  // -------------------------------------------------------------------
  // 11. updateSettings
  // -------------------------------------------------------------------
  describe('updateSettings', () => {
    it('merges partial settings', () => {
      const store = useReaderStore()
      store.updateSettings({ fontSize: 24, theme: 'dark' })

      expect(store.settings.fontSize).toBe(24)
      expect(store.settings.theme).toBe('dark')
      // Other settings remain default
      expect(store.settings.fontFamily).toBe('System')
      expect(store.settings.lineHeight).toBe(1.6)
    })

    it('overwrites only provided keys', () => {
      const store = useReaderStore()
      store.updateSettings({ fontSize: 24 })
      store.updateSettings({ lineHeight: 2.0 })

      expect(store.settings.fontSize).toBe(24)
      expect(store.settings.lineHeight).toBe(2.0)
    })

    it('handles empty partial', () => {
      const store = useReaderStore()
      const before = { ...store.settings }
      store.updateSettings({})
      expect(store.settings).toEqual(before)
    })
  })

  // -------------------------------------------------------------------
  // 12. setError
  // -------------------------------------------------------------------
  describe('setError', () => {
    it('sets error message', () => {
      const store = useReaderStore()
      store.setError('Something went wrong')
      expect(store.error).toBe('Something went wrong')
    })

    it('sets loading to false', () => {
      const store = useReaderStore()
      store.loading = true
      store.setError('fail')
      expect(store.loading).toBe(false)
    })
  })

  // -------------------------------------------------------------------
  // 13. reset
  // -------------------------------------------------------------------
  describe('reset', () => {
    it('resets all state to defaults', () => {
      const store = useReaderStore()

      // Populate the store with non-default values
      store.setBookContent(makeBookContent(['ch1', 'ch2'], { ch1: 100, ch2: 200 }))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)
      store.setPage(5)
      store.loading = true
      store.setError('err')
      store.toggleTOC()
      store.toggleUI()
      store.toggleSettings()
      store.navigationDirection = 'backward'

      store.reset()

      expect(store.bookContent).toBeNull()
      expect(store.currentChapterId).toBeNull()
      expect(store.currentChapterContent).toBeNull()
      expect(store.currentPage).toBe(1)
      expect(store.totalPages).toBe(1)
      expect(store.chapterPageCounts).toEqual(new Map())
      expect(store.loading).toBe(false)
      expect(store.error).toBeNull()
      expect(store.tocVisible).toBe(false)
      expect(store.uiVisible).toBe(true)
      expect(store.settingsVisible).toBe(false)
      expect(store.navigationDirection).toBe('forward')
    })

    it('does not reset settings', () => {
      const store = useReaderStore()
      store.updateSettings({ fontSize: 30, theme: 'dark' })
      store.reset()

      // Settings should remain after reset (the store code does not reset settings)
      expect(store.settings.fontSize).toBe(30)
      expect(store.settings.theme).toBe('dark')
    })
  })

  // -------------------------------------------------------------------
  // Edge cases
  // -------------------------------------------------------------------
  describe('edge cases', () => {
    it('setTotalPages with no current chapter does not crash', () => {
      const store = useReaderStore()
      expect(() => store.setTotalPages(5)).not.toThrow()
      expect(store.totalPages).toBe(5)
      // No chapter ID set, so nothing should be tracked
      expect(store.chapterPageCounts.size).toBe(0)
    })

    it('estimateChapterPages skips when no chapterSizes', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2'])) // no chapterSizes
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(10)

      // ch1 measured, ch2 not estimated (no sizes data)
      expect(store.chapterPageCounts.get('ch1')).toBe(10)
      expect(store.chapterPageCounts.has('ch2')).toBe(false)
    })

    it('single chapter book progress works correctly', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1']))
      store.setChapter(makeChapter('ch1'))
      store.setTotalPages(5)

      store.setPage(3)
      // chapterProgress: (3-1)/(5-1) = 50%
      expect(store.chapterProgress).toBe(50)
      // bookTotalPages = 5, bookCurrentPage = 3
      expect(store.bookTotalPages).toBe(5)
      expect(store.bookCurrentPage).toBe(3)
      // totalProgress: (3-1)/(5-1) = 50%
      expect(store.totalProgress).toBe(50)
    })

    it('hasNextChapter and hasPrevChapter with single chapter', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['only']))
      store.setChapter(makeChapter('only'))

      expect(store.hasNextChapter).toBe(false)
      expect(store.hasPrevChapter).toBe(false)
    })

    it('currentChapterIndex returns -1 for unknown chapter id', () => {
      const store = useReaderStore()
      store.setBookContent(makeBookContent(['ch1', 'ch2']))
      store.setChapter(makeChapter('unknown'))
      expect(store.currentChapterIndex).toBe(-1)
    })
  })
})
