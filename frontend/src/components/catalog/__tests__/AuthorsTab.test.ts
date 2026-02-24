import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import AuthorsTab from '../AuthorsTab.vue'
import { useCatalogStore } from '@/stores/catalog'

vi.mock('@/api/books', () => ({
  getAuthors: vi.fn(),
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

import * as booksApi from '@/api/books'

const vuetify = createVuetify()

const mockAuthors = [
  { id: 1, name: 'Азимов, Айзек', books_count: 42 },
  { id: 2, name: 'Герберт, Фрэнк', books_count: 15 },
  { id: 3, name: 'Кларк, Артур', books_count: 30 },
]

function mountAuthorsTab() {
  return mount(AuthorsTab, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('AuthorsTab', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads authors on mount', async () => {
    vi.mocked(booksApi.getAuthors).mockResolvedValue({
      items: mockAuthors,
      total: 3,
      page: 1,
      limit: 50,
    })

    mountAuthorsTab()
    await flushPromises()

    expect(booksApi.getAuthors).toHaveBeenCalledWith({
      q: undefined,
      page: 1,
      limit: 50,
    })
  })

  it('renders author list', async () => {
    vi.mocked(booksApi.getAuthors).mockResolvedValue({
      items: mockAuthors,
      total: 3,
      page: 1,
      limit: 50,
    })

    const wrapper = mountAuthorsTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Азимов, Айзек')
    expect(wrapper.text()).toContain('Герберт, Фрэнк')
  })

  it('shows empty state when no results', async () => {
    vi.mocked(booksApi.getAuthors).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 50,
    })

    const wrapper = mountAuthorsTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Ничего не найдено')
  })

  it('shows load more button when more items available', async () => {
    vi.mocked(booksApi.getAuthors).mockResolvedValue({
      items: mockAuthors,
      total: 100,
      page: 1,
      limit: 50,
    })

    const wrapper = mountAuthorsTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Загрузить ещё')
  })

  it('does not show load more when all loaded', async () => {
    vi.mocked(booksApi.getAuthors).mockResolvedValue({
      items: mockAuthors,
      total: 3,
      page: 1,
      limit: 50,
    })

    const wrapper = mountAuthorsTab()
    await flushPromises()

    expect(wrapper.text()).not.toContain('Загрузить ещё')
  })

  it('selects author on click and updates store', async () => {
    vi.mocked(booksApi.getAuthors).mockResolvedValue({
      items: mockAuthors,
      total: 3,
      page: 1,
      limit: 50,
    })
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountAuthorsTab()
    await flushPromises()

    const store = useCatalogStore()
    const listItems = wrapper.findAll('.authors-tab__item')
    await listItems[0].trigger('click')

    expect(store.navigationFilter?.type).toBe('author')
    expect(store.navigationFilter?.id).toBe(1)
  })

  it('loads more authors appending to list', async () => {
    vi.mocked(booksApi.getAuthors)
      .mockResolvedValueOnce({
        items: mockAuthors,
        total: 100,
        page: 1,
        limit: 50,
      })
      .mockResolvedValueOnce({
        items: [{ id: 4, name: 'Толкин, Джон', books_count: 20 }],
        total: 100,
        page: 2,
        limit: 50,
      })

    const wrapper = mountAuthorsTab()
    await flushPromises()

    const loadMoreBtn = wrapper.find('.authors-tab__load-more-btn')
    await loadMoreBtn.trigger('click')
    await flushPromises()

    expect(booksApi.getAuthors).toHaveBeenCalledTimes(2)
    expect(wrapper.text()).toContain('Толкин, Джон')
    // Original authors should still be there
    expect(wrapper.text()).toContain('Азимов, Айзек')
  })

  it('clears search resets list', async () => {
    vi.mocked(booksApi.getAuthors).mockResolvedValue({
      items: mockAuthors,
      total: 3,
      page: 1,
      limit: 50,
    })

    const wrapper = mountAuthorsTab()
    await flushPromises()

    // Type something and then clear
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
    vi.mocked(booksApi.getAuthors).mockRejectedValue(new Error('Network'))

    const wrapper = mountAuthorsTab()
    await flushPromises()

    // Should not crash, shows empty state
    expect(wrapper.text()).toContain('Ничего не найдено')
  })

  it('debounces search input', async () => {
    vi.useFakeTimers()
    vi.mocked(booksApi.getAuthors).mockResolvedValue({
      items: mockAuthors,
      total: 3,
      page: 1,
      limit: 50,
    })

    const wrapper = mountAuthorsTab()
    await flushPromises()
    vi.clearAllMocks()

    const input = wrapper.find('input')
    await input.setValue('Ази')
    await input.trigger('input')

    expect(booksApi.getAuthors).not.toHaveBeenCalled()

    vi.advanceTimersByTime(300)
    await flushPromises()

    expect(booksApi.getAuthors).toHaveBeenCalledWith({
      q: 'Ази',
      page: 1,
      limit: 50,
    })

    vi.useRealTimers()
  })
})
