import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { BookListItem, BookDetail, BookFilters } from '@/api/books'
import * as booksApi from '@/api/books'

export const useCatalogStore = defineStore('catalog', () => {
  const books = ref<BookListItem[]>([])
  const total = ref(0)
  const currentBook = ref<BookDetail | null>(null)
  const loading = ref(false)
  const bookLoading = ref(false)
  const error = ref<string | null>(null)

  const filters = ref<BookFilters>({
    page: 1,
    limit: 20,
    sort: 'title',
    order: 'asc',
  })

  const totalPages = computed(() => Math.ceil(total.value / (filters.value.limit || 20)))

  let fetchBooksController: AbortController | null = null

  async function fetchBooks() {
    if (fetchBooksController) {
      fetchBooksController.abort()
    }
    fetchBooksController = new AbortController()
    const controller = fetchBooksController

    loading.value = true
    error.value = null
    try {
      const result = await booksApi.getBooks(filters.value, controller.signal)
      if (!controller.signal.aborted) {
        books.value = result.items
        total.value = result.total
      }
    } catch (e: unknown) {
      if (controller.signal.aborted) return
      error.value = e instanceof Error ? e.message : 'Failed to load books'
    } finally {
      if (!controller.signal.aborted) {
        loading.value = false
      }
    }
  }

  async function fetchBook(id: number) {
    bookLoading.value = true
    error.value = null
    currentBook.value = null
    try {
      currentBook.value = await booksApi.getBook(id)
    } catch (e: unknown) {
      error.value = e instanceof Error ? e.message : 'Failed to load book'
    } finally {
      bookLoading.value = false
    }
  }

  function updateFilters(newFilters: Partial<BookFilters>) {
    filters.value = { ...filters.value, ...newFilters, page: 1 }
    return fetchBooks()
  }

  function setPage(page: number) {
    filters.value.page = page
    return fetchBooks()
  }

  function resetFilters() {
    filters.value = { page: 1, limit: 20, sort: 'title', order: 'asc' }
    return fetchBooks()
  }

  return {
    books,
    total,
    currentBook,
    loading,
    bookLoading,
    error,
    filters,
    totalPages,
    fetchBooks,
    fetchBook,
    updateFilters,
    setPage,
    resetFilters,
  }
})
