import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import SeriesTab from '../SeriesTab.vue'
import { useCatalogStore } from '@/stores/catalog'

vi.mock('@/api/books', () => ({
  getSeries: vi.fn(),
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

import * as booksApi from '@/api/books'

const vuetify = createVuetify()

const mockSeries = [
  { id: 1, name: 'Основание', books_count: 7 },
  { id: 2, name: 'Дюна', books_count: 6 },
]

function mountSeriesTab() {
  return mount(SeriesTab, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('SeriesTab', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads series on mount', async () => {
    vi.mocked(booksApi.getSeries).mockResolvedValue({
      items: mockSeries,
      total: 2,
      page: 1,
      limit: 50,
    })

    mountSeriesTab()
    await flushPromises()

    expect(booksApi.getSeries).toHaveBeenCalledWith({
      q: undefined,
      page: 1,
      limit: 50,
    })
  })

  it('renders series list', async () => {
    vi.mocked(booksApi.getSeries).mockResolvedValue({
      items: mockSeries,
      total: 2,
      page: 1,
      limit: 50,
    })

    const wrapper = mountSeriesTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Основание')
    expect(wrapper.text()).toContain('Дюна')
  })

  it('shows empty state', async () => {
    vi.mocked(booksApi.getSeries).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 50,
    })

    const wrapper = mountSeriesTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Ничего не найдено')
  })

  it('selects series on click', async () => {
    vi.mocked(booksApi.getSeries).mockResolvedValue({
      items: mockSeries,
      total: 2,
      page: 1,
      limit: 50,
    })
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountSeriesTab()
    await flushPromises()

    const store = useCatalogStore()
    const listItems = wrapper.findAll('.v-list-item')
    await listItems[0].trigger('click')

    expect(store.navigationFilter?.type).toBe('series')
    expect(store.navigationFilter?.id).toBe(1)
  })
})
