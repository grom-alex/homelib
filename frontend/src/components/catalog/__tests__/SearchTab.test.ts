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

  it('renders genre options from API', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue([
      { id: 1, code: 'sf', name: 'Фантастика', books_count: 100 },
      { id: 2, code: 'det', name: 'Детектив', books_count: 50 },
    ])

    const wrapper = mountSearchTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Фантастика')
    expect(wrapper.text()).toContain('Детектив')
  })

  it('renders genres with meta_group prefix', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue([
      { id: 1, code: 'sf', name: 'Фантастика', meta_group: 'Проза', books_count: 100 },
    ])

    const wrapper = mountSearchTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Проза / Фантастика')
  })

  it('renders genre children', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue([
      {
        id: 1, code: 'sf', name: 'Фантастика', meta_group: 'Проза', books_count: 100,
        children: [
          { id: 3, code: 'sf_space', name: 'Космическая', books_count: 30 },
        ],
      },
    ])

    const wrapper = mountSearchTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Космическая')
  })

  it('renders format select options from stats', async () => {
    const wrapper = mountSearchTab()
    await flushPromises()

    const selects = wrapper.findAll('select')
    // Find format select (second one after genre)
    const formatSelect = selects.find(s => s.text().includes('fb2'))
    expect(formatSelect).toBeDefined()
  })

  it('renders language select options from stats', async () => {
    const wrapper = mountSearchTab()
    await flushPromises()

    const selects = wrapper.findAll('select')
    const langSelect = selects.find(s => s.text().includes('ru'))
    expect(langSelect).toBeDefined()
  })

  it('submits search with author_name param', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountSearchTab()
    await flushPromises()

    const inputs = wrapper.findAll('input')
    // Second input is author_name
    await inputs[1].setValue('Азимов')

    await wrapper.find('form').trigger('submit')

    const store = useCatalogStore()
    expect(store.navigationFilter?.params?.author_name).toBe('Азимов')
  })

  it('submits search with series_name param', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountSearchTab()
    await flushPromises()

    const inputs = wrapper.findAll('input')
    // Third input is series_name
    await inputs[2].setValue('Основание')

    await wrapper.find('form').trigger('submit')

    const store = useCatalogStore()
    expect(store.navigationFilter?.params?.series_name).toBe('Основание')
  })

  it('uses "Расширенный поиск" label when no text fields filled', async () => {
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountSearchTab()
    await flushPromises()

    await wrapper.find('form').trigger('submit')

    const store = useCatalogStore()
    expect(store.navigationFilter?.label).toBe('Расширенный поиск')
  })

  it('handles loadOptions errors gracefully', async () => {
    vi.mocked(booksApi.getGenres).mockRejectedValue(new Error('Network'))
    vi.mocked(booksApi.getStats).mockRejectedValue(new Error('Network'))

    const wrapper = mountSearchTab()
    await flushPromises()

    // Should not crash, form still renders
    expect(wrapper.text()).toContain('Найти')
  })
})
