import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import SearchTab from '../SearchTab.vue'
import { useCatalogStore } from '@/stores/catalog'

vi.mock('@/api/books', () => ({
  getGenres: vi.fn().mockResolvedValue([]),
  getStats: vi.fn().mockResolvedValue({
    books_count: 100,
    authors_count: 50,
    genres_count: 20,
    series_count: 10,
    languages: ['ru', 'en'],
    formats: ['fb2', 'epub', 'pdf'],
  }),
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

import * as booksApi from '@/api/books'

const vuetify = createVuetify()

function mountSearchTab() {
  return mount(SearchTab, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('SearchTab', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.mocked(booksApi.getGenres).mockResolvedValue([])
    vi.mocked(booksApi.getStats).mockResolvedValue({
      books_count: 100,
      authors_count: 50,
      genres_count: 20,
      series_count: 10,
      languages: ['ru', 'en'],
      formats: ['fb2', 'epub', 'pdf'],
    })
  })

  it('loads options on mount', async () => {
    mountSearchTab()
    await flushPromises()

    expect(booksApi.getGenres).toHaveBeenCalled()
    expect(booksApi.getStats).toHaveBeenCalled()
  })

  it('renders search form', async () => {
    const wrapper = mountSearchTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Найти')
    expect(wrapper.text()).toContain('Очистить')
  })

  it('submits search with params', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountSearchTab()
    await flushPromises()

    // Set query field
    const input = wrapper.find('input')
    await input.setValue('Дюна')

    // Submit form
    await wrapper.find('form').trigger('submit')

    const store = useCatalogStore()
    expect(store.navigationFilter?.type).toBe('search')
    expect(store.navigationFilter?.params?.q).toBe('Дюна')
  })

  it('clears form on reset', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountSearchTab()
    await flushPromises()

    const input = wrapper.find('input')
    await input.setValue('Дюна')

    // Click clear button
    const clearBtn = wrapper.find('.search-tab__btn-clear')
    await clearBtn.trigger('click')

    expect(input.element.value).toBe('')
  })
})
