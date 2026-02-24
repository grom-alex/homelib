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
    const listItems = wrapper.findAll('.series-tab__item')
    await listItems[0].trigger('click')

    expect(store.navigationFilter?.type).toBe('series')
    expect(store.navigationFilter?.id).toBe(1)
  })

  it('shows load more button when more items available', async () => {
    vi.mocked(booksApi.getSeries).mockResolvedValue({
      items: mockSeries,
      total: 100,
      page: 1,
      limit: 50,
    })

    const wrapper = mountSeriesTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Загрузить ещё')
  })

  it('loads more series appending to list', async () => {
    vi.mocked(booksApi.getSeries)
      .mockResolvedValueOnce({
        items: mockSeries,
        total: 100,
        page: 1,
        limit: 50,
      })
      .mockResolvedValueOnce({
        items: [{ id: 3, name: 'Гарри Поттер', books_count: 7, authors: 'Роулинг' }],
        total: 100,
        page: 2,
        limit: 50,
      })

    const wrapper = mountSeriesTab()
    await flushPromises()

    const loadMoreBtn = wrapper.find('.series-tab__load-more-btn')
    await loadMoreBtn.trigger('click')
    await flushPromises()

    expect(booksApi.getSeries).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Гарри Поттер')
    expect(wrapper.text()).toContain('Основание')
  })

  it('clears search resets list', async () => {
    vi.mocked(booksApi.getSeries).mockResolvedValue({
      items: mockSeries,
      total: 2,
      page: 1,
      limit: 50,
    })

    const wrapper = mountSeriesTab()
    await flushPromises()

    const input = wrapper.find('input')
    await input.setValue('test')

    const clearBtn = wrapper.find('.search-input-clear')
    if (clearBtn.exists()) {
      await clearBtn.trigger('click')
      await flushPromises()
      expect(input.element.value).toBe('')
    }
  })

  it('handles API error gracefully', async () => {
    vi.mocked(booksApi.getSeries).mockRejectedValue(new Error('Network'))

    const wrapper = mountSeriesTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Ничего не найдено')
  })

  it('debounces search input', async () => {
    vi.useFakeTimers()
    vi.mocked(booksApi.getSeries).mockResolvedValue({
      items: mockSeries,
      total: 2,
      page: 1,
      limit: 50,
    })

    const wrapper = mountSeriesTab()
    await flushPromises()
    vi.clearAllMocks()

    const input = wrapper.find('input')
    await input.setValue('Осн')
    await input.trigger('input')

    expect(booksApi.getSeries).not.toHaveBeenCalled()

    vi.advanceTimersByTime(300)
    await flushPromises()

    expect(booksApi.getSeries).toHaveBeenCalledWith({
      q: 'Осн',
      page: 1,
      limit: 50,
    })

    vi.useRealTimers()
  })

  it('renders author names when available', async () => {
    vi.mocked(booksApi.getSeries).mockResolvedValue({
      items: [{ id: 1, name: 'Основание', books_count: 7, authors: 'Азимов, Айзек' }],
      total: 1,
      page: 1,
      limit: 50,
    })

    const wrapper = mountSeriesTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Азимов, Айзек')
  })
})
