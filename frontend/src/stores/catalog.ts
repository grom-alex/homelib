import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { BookListItem, BookDetail, BookFilters } from '@/api/books'
import * as booksApi from '@/api/books'
import type { TabType, NavigationFilter, SortField, SortOrder, PageSize } from '@/types/catalog'
import { defaultCatalogSettings } from '@/types/catalog'

interface TabState {
  books: BookListItem[]
  total: number
  filters: BookFilters
  navigationFilter: NavigationFilter | null
  selectedBookId: number | null
  currentBook: BookDetail | null
}

function createEmptyTabState(): TabState {
  return {
    books: [],
    total: 0,
    filters: {
      page: 1,
      limit: defaultCatalogSettings.pageSize,
      sort: defaultCatalogSettings.tableSort.field,
      order: defaultCatalogSettings.tableSort.order,
    },
    navigationFilter: null,
    selectedBookId: null,
    currentBook: null,
  }
}

export const useCatalogStore = defineStore('catalog', () => {
  const books = ref<BookListItem[]>([])
  const total = ref(0)
  const currentBook = ref<BookDetail | null>(null)
  const loading = ref(false)
  const bookLoading = ref(false)
  const error = ref<string | null>(null)

  const activeTab = ref<TabType>(defaultCatalogSettings.activeTab)
  const selectedBookId = ref<number | null>(null)
  const navigationFilter = ref<NavigationFilter | null>(null)

  const filters = ref<BookFilters>({
    page: 1,
    limit: defaultCatalogSettings.pageSize,
    sort: defaultCatalogSettings.tableSort.field,
    order: defaultCatalogSettings.tableSort.order,
  })

  const tabStates = new Map<TabType, TabState>()

  const totalPages = computed(() => Math.ceil(total.value / (filters.value.limit || defaultCatalogSettings.pageSize)))

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
        books.value = result.items ?? []
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

  function selectNavItem(type: NavigationFilter['type'], id?: number, params?: Record<string, string>, label?: string) {
    navigationFilter.value = { type, id, label, params }
    selectedBookId.value = null
    currentBook.value = null

    // Clear all navigation-specific filters before applying new ones
    const apiFilters: Partial<BookFilters> = {
      author_id: undefined,
      author_name: undefined,
      series_id: undefined,
      series_name: undefined,
      genre_id: undefined,
      q: undefined,
      format: undefined,
      lang: undefined,
    }
    if (type === 'author' && id) apiFilters.author_id = id
    else if (type === 'series' && id) apiFilters.series_id = id
    else if (type === 'genre' && id) apiFilters.genre_id = id
    else if (type === 'search' && params) {
      if (params.q) apiFilters.q = params.q
      if (params.author_name) apiFilters.author_name = params.author_name
      if (params.genre_id) apiFilters.genre_id = Number(params.genre_id)
      if (params.series_name) apiFilters.series_name = params.series_name
      if (params.format) apiFilters.format = params.format
      if (params.lang) apiFilters.lang = params.lang
    }

    updateFilters(apiFilters)
  }

  function saveCurrentTabState() {
    tabStates.set(activeTab.value, {
      books: books.value,
      total: total.value,
      filters: { ...filters.value },
      navigationFilter: navigationFilter.value,
      selectedBookId: selectedBookId.value,
      currentBook: currentBook.value,
    })
  }

  function restoreTabState(tab: TabType) {
    const state = tabStates.get(tab)
    if (state) {
      books.value = state.books
      total.value = state.total
      filters.value = { ...state.filters }
      navigationFilter.value = state.navigationFilter
      selectedBookId.value = state.selectedBookId
      currentBook.value = state.currentBook
    } else {
      const empty = createEmptyTabState()
      books.value = empty.books
      total.value = empty.total
      filters.value = empty.filters
      navigationFilter.value = empty.navigationFilter
      selectedBookId.value = empty.selectedBookId
      currentBook.value = empty.currentBook
    }
  }

  function setActiveTab(tab: TabType) {
    if (activeTab.value === tab) return
    saveCurrentTabState()
    activeTab.value = tab
    restoreTabState(tab)
  }

  function setSelectedBook(id: number) {
    selectedBookId.value = id
    return fetchBook(id)
  }

  function setSort(field: SortField, order: SortOrder) {
    filters.value = { ...filters.value, sort: field, order, page: 1 }
    return fetchBooks()
  }

  function updateFilters(newFilters: Partial<BookFilters>) {
    filters.value = { ...filters.value, ...newFilters, page: 1 }
    return fetchBooks()
  }

  function setPage(page: number) {
    filters.value.page = page
    return fetchBooks()
  }

  function setPageSize(size: PageSize) {
    filters.value = { ...filters.value, limit: size, page: 1 }
    return fetchBooks()
  }

  function resetFilters() {
    filters.value = { page: 1, limit: defaultCatalogSettings.pageSize, sort: 'title', order: 'asc' }
    navigationFilter.value = null
    selectedBookId.value = null
    currentBook.value = null
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
    activeTab,
    selectedBookId,
    navigationFilter,
    fetchBooks,
    fetchBook,
    selectNavItem,
    setActiveTab,
    setSelectedBook,
    setSort,
    updateFilters,
    setPage,
    setPageSize,
    resetFilters,
  }
})
