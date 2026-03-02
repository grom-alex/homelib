import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import SearchTab from '../SearchTab.vue'
import { useCatalogStore } from '@/stores/catalog'
import { useParentalStore } from '@/stores/parental'

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

vi.mock('@/api/parental', () => ({
  getMyParentalStatus: vi.fn().mockResolvedValue({ adult_content_enabled: false, pin_set: false }),
  unlockAdultContent: vi.fn(),
  lockAdultContent: vi.fn(),
}))

import * as booksApi from '@/api/books'

const vuetify = createVuetify()

const mockGenreTree = [
  {
    id: 1, code: 'sf_all', name: 'Фантастика', position: '0.1', books_count: 100,
    children: [
      { id: 3, code: 'sf_space', name: 'Космическая', position: '0.1.1', books_count: 30 },
    ],
  },
  { id: 2, code: 'det_all', name: 'Детектив', position: '0.2', books_count: 50 },
]

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

  it('shows genre dropdown button when genres loaded', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountSearchTab()
    await flushPromises()

    const genreBtn = wrapper.find('.search-field__genre-btn')
    expect(genreBtn.exists()).toBe(true)
    expect(genreBtn.text()).toContain('Все жанры')
  })

  it('renders format select options from stats', async () => {
    const wrapper = mountSearchTab()
    await flushPromises()

    const selects = wrapper.findAll('select')
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

  it('clears genre selection on reset', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountSearchTab()
    await flushPromises()

    // Genre btn should show "Все жанры" after clear
    const clearBtn = wrapper.find('.search-tab__btn-clear')
    await clearBtn.trigger('click')

    const genreBtn = wrapper.find('.search-field__genre-btn')
    expect(genreBtn.text()).toContain('Все жанры')
  })

  it('submits with genre_id when genre selected', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountSearchTab()
    await flushPromises()

    // Simulate genre activation via component internals
    const vm = wrapper.vm as unknown as { onGenreActivated: (ids: number[]) => void; form: { genre_id: number | null } }
    vm.onGenreActivated([1])
    await flushPromises()

    expect(vm.form.genre_id).toBe(1)

    await wrapper.find('form').trigger('submit')

    const store = useCatalogStore()
    expect(store.navigationFilter?.params?.genre_id).toBe('1')
  })

  it('clears genre via clearGenre', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountSearchTab()
    await flushPromises()

    const vm = wrapper.vm as unknown as { onGenreActivated: (ids: number[]) => void; clearGenre: () => void; form: { genre_id: number | null } }
    vm.onGenreActivated([2])
    expect(vm.form.genre_id).toBe(2)

    vm.clearGenre()
    expect(vm.form.genre_id).toBeNull()
  })

  it('genreActivated returns empty array when no genre selected', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountSearchTab()
    await flushPromises()

    const vm = wrapper.vm as unknown as { genreActivated: number[] }
    expect(vm.genreActivated).toEqual([])
  })

  it('reloads options when parental status changes', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue([])

    mountSearchTab()
    await flushPromises()

    expect(booksApi.getGenres).toHaveBeenCalledTimes(1)

    // Change parental status
    const parentalStore = useParentalStore()
    parentalStore.adultContentEnabled = true
    await flushPromises()

    expect(booksApi.getGenres).toHaveBeenCalledTimes(2)
  })

  it('onGenreActivated does nothing with empty array', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountSearchTab()
    await flushPromises()

    const vm = wrapper.vm as unknown as { onGenreActivated: (ids: number[]) => void; form: { genre_id: number | null } }
    vm.onGenreActivated([3])
    expect(vm.form.genre_id).toBe(3)

    // Empty activation should not change genre_id
    vm.onGenreActivated([])
    expect(vm.form.genre_id).toBe(3)
  })
})
