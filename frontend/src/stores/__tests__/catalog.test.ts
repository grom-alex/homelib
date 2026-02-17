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
    vi.mocked(booksApi.getBooks).mockResolvedValue({ items: [], total: 45, page: 1, limit: 20 })
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
})
