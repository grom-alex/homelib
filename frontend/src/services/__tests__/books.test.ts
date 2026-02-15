import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
vi.mock('../api', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
  },
}))

import { getBooks, getBook, downloadBook, getAuthors, getAuthor, getGenres, getSeries, getStats } from '../books'

describe('books service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('getBooks calls GET /books with filters', async () => {
    const resp = { items: [], total: 0, page: 1, limit: 20 }
    mockGet.mockResolvedValue({ data: resp })
    const result = await getBooks({ q: 'test', page: 2 })
    expect(mockGet).toHaveBeenCalledWith('/books', { params: { q: 'test', page: 2 } })
    expect(result).toEqual(resp)
  })

  it('getBooks filters out empty values', async () => {
    mockGet.mockResolvedValue({ data: { items: [], total: 0, page: 1, limit: 20 } })
    await getBooks({ q: '', page: 1 })
    expect(mockGet).toHaveBeenCalledWith('/books', { params: { page: 1 } })
  })

  it('getBook calls GET /books/:id', async () => {
    const book = { id: 1, title: 'Book' }
    mockGet.mockResolvedValue({ data: book })
    const result = await getBook(1)
    expect(mockGet).toHaveBeenCalledWith('/books/1')
    expect(result).toEqual(book)
  })

  it('getAuthors calls GET /authors', async () => {
    const resp = { items: [], total: 0, page: 1, limit: 20 }
    mockGet.mockResolvedValue({ data: resp })
    const result = await getAuthors({ q: 'pushkin' })
    expect(mockGet).toHaveBeenCalledWith('/authors', { params: { q: 'pushkin' } })
    expect(result).toEqual(resp)
  })

  it('getAuthor calls GET /authors/:id', async () => {
    const author = { id: 1, name: 'Author', books: [], books_count: 0 }
    mockGet.mockResolvedValue({ data: author })
    const result = await getAuthor(1)
    expect(mockGet).toHaveBeenCalledWith('/authors/1')
    expect(result).toEqual(author)
  })

  it('getGenres calls GET /genres', async () => {
    const genres = [{ id: 1, code: 'sf', name: 'Sci-fi', books_count: 10 }]
    mockGet.mockResolvedValue({ data: genres })
    const result = await getGenres()
    expect(mockGet).toHaveBeenCalledWith('/genres')
    expect(result).toEqual(genres)
  })

  it('getSeries calls GET /series', async () => {
    const resp = { items: [], total: 0, page: 1, limit: 20 }
    mockGet.mockResolvedValue({ data: resp })
    const result = await getSeries({ q: 'name', page: 1, limit: 50 })
    expect(mockGet).toHaveBeenCalledWith('/series', { params: { q: 'name', page: 1, limit: 50 } })
    expect(result).toEqual(resp)
  })

  it('downloadBook creates blob download', async () => {
    const blobData = new Blob(['file content'])
    mockGet.mockResolvedValue({
      data: blobData,
      headers: { 'content-disposition': 'attachment; filename="test.fb2"' },
    })

    // Mock URL and DOM
    const createObjectURL = vi.fn().mockReturnValue('blob:test')
    const revokeObjectURL = vi.fn()
    global.URL.createObjectURL = createObjectURL
    global.URL.revokeObjectURL = revokeObjectURL

    const clickSpy = vi.fn()
    const appendSpy = vi.spyOn(document.body, 'appendChild').mockImplementation(vi.fn())
    const removeSpy = vi.spyOn(document.body, 'removeChild').mockImplementation(vi.fn())
    vi.spyOn(document, 'createElement').mockReturnValue({
      set href(_: string) {},
      set download(_: string) {},
      click: clickSpy,
    } as unknown as HTMLAnchorElement)

    await downloadBook(1)
    expect(mockGet).toHaveBeenCalledWith('/books/1/download', { responseType: 'blob' })
    expect(createObjectURL).toHaveBeenCalled()
    expect(clickSpy).toHaveBeenCalled()
    expect(revokeObjectURL).toHaveBeenCalled()

    appendSpy.mockRestore()
    removeSpy.mockRestore()
  })

  it('downloadBook uses fallback filename without disposition header', async () => {
    const blobData = new Blob(['file content'])
    mockGet.mockResolvedValue({
      data: blobData,
      headers: {},
    })

    global.URL.createObjectURL = vi.fn().mockReturnValue('blob:test')
    global.URL.revokeObjectURL = vi.fn()
    vi.spyOn(document.body, 'appendChild').mockImplementation(vi.fn())
    vi.spyOn(document.body, 'removeChild').mockImplementation(vi.fn())

    let downloadValue = ''
    vi.spyOn(document, 'createElement').mockReturnValue({
      set href(_: string) {},
      set download(v: string) { downloadValue = v },
      get download() { return downloadValue },
      click: vi.fn(),
    } as unknown as HTMLAnchorElement)

    await downloadBook(42)
    expect(downloadValue).toBe('book_42')
  })

  it('getStats calls GET /stats', async () => {
    const stats = { books_count: 100, authors_count: 50, genres_count: 20, series_count: 10, languages: [], formats: [] }
    mockGet.mockResolvedValue({ data: stats })
    const result = await getStats()
    expect(mockGet).toHaveBeenCalledWith('/stats')
    expect(result).toEqual(stats)
  })
})
