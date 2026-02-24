import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useCatalogStore } from '../catalog'

vi.mock('@/api/books', () => ({
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

import * as booksApi from '@/api/books'

const mockBooks = [
  { id: 1, title: 'Book 1', lang: 'ru', format: 'fb2', is_deleted: false, authors: [], genres: [] },
  { id: 2, title: 'Book 2', lang: 'en', format: 'epub', is_deleted: false, authors: [], genres: [] },
]

describe('catalog store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('starts with empty state', () => {
    const store = useCatalogStore()
    expect(store.books).toEqual([])
    expect(store.total).toBe(0)
    expect(store.loading).toBe(false)
    expect(store.error).toBeNull()
  })

  it('fetchBooks updates state', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: mockBooks as never[],
      total: 2,
      page: 1,
      limit: 20,
    })
    const store = useCatalogStore()
    await store.fetchBooks()

    expect(store.books).toEqual(mockBooks)
    expect(store.total).toBe(2)
    expect(store.loading).toBe(false)
  })

  it('fetchBooks sets error on failure', async () => {
    vi.mocked(booksApi.getBooks).mockRejectedValue(new Error('Network error'))
    const store = useCatalogStore()
    await store.fetchBooks()

    expect(store.error).toBe('Network error')
    expect(store.books).toEqual([])
  })

  it('fetchBook loads single book', async () => {
    const detail = { id: 1, title: 'Book 1', lang: 'ru', format: 'fb2', is_deleted: false, authors: [], genres: [] }
    vi.mocked(booksApi.getBook).mockResolvedValue(detail as never)
    const store = useCatalogStore()
    await store.fetchBook(1)

    expect(store.currentBook).toEqual(detail)
  })

  it('updateFilters resets page to 1 and fetches', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    store.filters.page = 5

    await store.updateFilters({ q: 'test' })
    expect(store.filters.page).toBe(1)
    expect(store.filters.q).toBe('test')
    expect(booksApi.getBooks).toHaveBeenCalled()
  })

  it('setPage updates page and fetches', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 100, page: 3, limit: 20 })
    const store = useCatalogStore()
    await store.setPage(3)

    expect(store.filters.page).toBe(3)
  })

  it('resetFilters restores defaults', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    store.filters.q = 'search'
    store.filters.lang = 'ru'

    await store.resetFilters()
    expect(store.filters.q).toBeUndefined()
    expect(store.filters.lang).toBeUndefined()
    expect(store.filters.page).toBe(1)
  })

  it('totalPages computed correctly', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 75, page: 1, limit: 25 })
    const store = useCatalogStore()
    await store.fetchBooks()
    expect(store.totalPages).toBe(3)
  })

  it('fetchBooks aborts previous request', async () => {
    let resolveFirst: (v: unknown) => void
    const firstCall = new Promise((resolve) => { resolveFirst = resolve })
    vi.mocked(booksApi.getBooks)
      .mockImplementationOnce(() => firstCall as never)
      .mockResolvedValueOnce({ items: mockBooks as never[], total: 2, page: 1, limit: 20 })

    const store = useCatalogStore()
    const first = store.fetchBooks()
    const second = store.fetchBooks()

    resolveFirst!({ items: [], total: 0, page: 1, limit: 20 })
    await first
    await second

    expect(store.books).toEqual(mockBooks)
  })

  it('fetchBooks sets fallback error for non-Error rejection', async () => {
    vi.mocked(booksApi.getBooks).mockRejectedValue('string error')
    const store = useCatalogStore()
    await store.fetchBooks()

    expect(store.error).toBe('Failed to load books')
  })

  it('fetchBook sets fallback error for non-Error rejection', async () => {
    vi.mocked(booksApi.getBook).mockRejectedValue('string error')
    const store = useCatalogStore()
    await store.fetchBook(1)

    expect(store.error).toBe('Failed to load book')
  })

  it('fetchBook sets error on failure', async () => {
    vi.mocked(booksApi.getBook).mockRejectedValue(new Error('Not found'))
    const store = useCatalogStore()
    await store.fetchBook(1)

    expect(store.error).toBe('Not found')
    expect(store.currentBook).toBeNull()
    expect(store.bookLoading).toBe(false)
  })

  it('starts with default activeTab and no selection', () => {
    const store = useCatalogStore()
    expect(store.activeTab).toBe('authors')
    expect(store.selectedBookId).toBeNull()
    expect(store.navigationFilter).toBeNull()
  })

  it('selectNavItem sets filter and fetches books', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    await store.selectNavItem('author', 42)

    expect(store.navigationFilter).toEqual({ type: 'author', id: 42, params: undefined })
    expect(store.selectedBookId).toBeNull()
    expect(booksApi.getBooks).toHaveBeenCalledWith(
      expect.objectContaining({ author_id: 42, page: 1 }),
      expect.any(AbortSignal),
    )
  })

  it('selectNavItem with series sets series_id filter', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    await store.selectNavItem('series', 7)

    expect(booksApi.getBooks).toHaveBeenCalledWith(
      expect.objectContaining({ series_id: 7 }),
      expect.any(AbortSignal),
    )
  })

  it('selectNavItem with genre sets genre_id filter', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    await store.selectNavItem('genre', 3)

    expect(booksApi.getBooks).toHaveBeenCalledWith(
      expect.objectContaining({ genre_id: 3 }),
      expect.any(AbortSignal),
    )
  })

  it('selectNavItem with search passes params', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    await store.selectNavItem('search', undefined, { q: 'Dune', format: 'fb2' })

    expect(booksApi.getBooks).toHaveBeenCalledWith(
      expect.objectContaining({ q: 'Dune', format: 'fb2' }),
      expect.any(AbortSignal),
    )
  })

  it('setActiveTab resets state', () => {
    const store = useCatalogStore()
    store.selectedBookId = 5
    store.books = mockBooks as never[]
    store.total = 2

    store.setActiveTab('series')
    expect(store.activeTab).toBe('series')
    expect(store.selectedBookId).toBeNull()
    expect(store.books).toEqual([])
    expect(store.total).toBe(0)
    expect(store.navigationFilter).toBeNull()
  })

  it('setActiveTab does nothing if same tab', () => {
    const store = useCatalogStore()
    store.books = mockBooks as never[]
    store.setActiveTab('authors')
    expect(store.books).toEqual(mockBooks)
  })

  it('setSelectedBook updates id and fetches book', async () => {
    const detail = { id: 1, title: 'Book 1', lang: 'ru', format: 'fb2', is_deleted: false, authors: [], genres: [] }
    vi.mocked(booksApi.getBook).mockResolvedValue(detail as never)
    const store = useCatalogStore()
    await store.setSelectedBook(1)

    expect(store.selectedBookId).toBe(1)
    expect(store.currentBook).toEqual(detail)
  })

  it('setSort updates sort and fetches', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    await store.setSort('year', 'desc')

    expect(store.filters.sort).toBe('year')
    expect(store.filters.order).toBe('desc')
    expect(store.filters.page).toBe(1)
  })

  it('selectNavItem resets selectedBook and currentBook', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    store.selectedBookId = 5
    store.currentBook = { id: 5, title: 'Test' } as never

    await store.selectNavItem('author', 1)
    expect(store.selectedBookId).toBeNull()
    expect(store.currentBook).toBeNull()
  })

  it('resetFilters clears navigation state', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 0, page: 1, limit: 20 })
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.selectedBookId = 3

    await store.resetFilters()
    expect(store.navigationFilter).toBeNull()
    expect(store.selectedBookId).toBeNull()
    expect(store.currentBook).toBeNull()
  })
})
