import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPut = vi.fn()
vi.mock('../client', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    put: (...args: unknown[]) => mockPut(...args),
  },
}))

import {
  getBookContent,
  getChapter,
  getBookImageUrl,
  getReadingProgress,
  saveReadingProgress,
  getUserSettings,
  updateUserSettings,
} from '../reader'

import type { BookContent, ChapterContent, ReadingPosition, ReaderSettings } from '@/types/reader'

describe('reader API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getBookContent calls GET /books/:id/content', async () => {
    const content: BookContent = {
      metadata: {
        title: 'Test Book',
        author: 'Author Name',
        cover: null,
        language: 'ru',
        format: 'fb2',
      },
      toc: [{ id: 'ch1', title: 'Chapter 1', level: 0 }],
      chapters: ['ch1', 'ch2'],
      totalChapters: 2,
    }
    mockGet.mockResolvedValue({ data: content })

    const result = await getBookContent(42)

    expect(mockGet).toHaveBeenCalledWith('/books/42/content')
    expect(result).toEqual(content)
  })

  it('getChapter calls GET /books/:id/chapter/:chapterId', async () => {
    const chapter: ChapterContent = {
      id: 'ch1',
      title: 'Chapter 1',
      html: '<p>Hello world</p>',
    }
    mockGet.mockResolvedValue({ data: chapter })

    const result = await getChapter(5, 'ch1')

    expect(mockGet).toHaveBeenCalledWith('/books/5/chapter/ch1')
    expect(result).toEqual(chapter)
  })

  it('getBookImageUrl returns correct URL string', () => {
    const url = getBookImageUrl(10, 'img_cover')

    expect(url).toBe('/api/books/10/image/img_cover')
  })

  it('getReadingProgress returns null on 204', async () => {
    mockGet.mockResolvedValue({ status: 204, data: '' })

    const result = await getReadingProgress(7)

    expect(mockGet).toHaveBeenCalledWith('/me/books/7/progress')
    expect(result).toBeNull()
  })

  it('getReadingProgress returns data on 200', async () => {
    const position: ReadingPosition = {
      chapterId: 'ch3',
      chapterProgress: 45,
      totalProgress: 30,
      device: 'desktop',
      updatedAt: '2026-01-15T10:00:00Z',
    }
    mockGet.mockResolvedValue({ status: 200, data: position })

    const result = await getReadingProgress(7)

    expect(mockGet).toHaveBeenCalledWith('/me/books/7/progress')
    expect(result).toEqual(position)
  })

  it('saveReadingProgress calls PUT /me/books/:id/progress', async () => {
    const position = {
      chapterId: 'ch2',
      chapterProgress: 80,
      totalProgress: 55,
      device: 'mobile',
    }
    const saved: ReadingPosition = {
      ...position,
      updatedAt: '2026-02-19T12:00:00Z',
    }
    mockPut.mockResolvedValue({ data: saved })

    const result = await saveReadingProgress(3, position)

    expect(mockPut).toHaveBeenCalledWith('/me/books/3/progress', position)
    expect(result).toEqual(saved)
  })

  it('getUserSettings calls GET /me/settings', async () => {
    const settings = {
      reader: {
        fontSize: 20,
        theme: 'dark' as const,
      },
    }
    mockGet.mockResolvedValue({ data: settings })

    const result = await getUserSettings()

    expect(mockGet).toHaveBeenCalledWith('/me/settings')
    expect(result).toEqual(settings)
  })

  it('updateUserSettings calls PUT /me/settings', async () => {
    const input = {
      reader: {
        fontSize: 22,
        theme: 'sepia' as const,
      } as Partial<ReaderSettings>,
    }
    const saved = {
      reader: {
        fontSize: 22,
        fontFamily: 'Georgia',
        fontWeight: 400,
        lineHeight: 1.6,
        paragraphSpacing: 0.5,
        letterSpacing: 0,
        marginHorizontal: 5,
        marginVertical: 3,
        firstLineIndent: 1.5,
        textAlign: 'justify',
        hyphenation: true,
        theme: 'sepia',
        customColors: null,
        viewMode: 'paginated',
        pageAnimation: 'slide',
        showProgress: true,
        showClock: false,
        tapZones: 'lrc',
      },
    }
    mockPut.mockResolvedValue({ data: saved })

    const result = await updateUserSettings(input)

    expect(mockPut).toHaveBeenCalledWith('/me/settings', input)
    expect(result).toEqual(saved)
  })
})
