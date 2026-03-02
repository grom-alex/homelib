import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import GenresTab from '../GenresTab.vue'
import { useCatalogStore } from '@/stores/catalog'

vi.mock('@/api/books', () => ({
  getGenres: vi.fn(),
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

import * as booksApi from '@/api/books'

const vuetify = createVuetify()

const mockGenreTree = [
  {
    id: 1, code: 'sf_all', name: 'Фантастика', position: '0.1', books_count: 250,
    children: [
      { id: 10, code: 'sf_history', name: 'Альтернативная история', position: '0.1.1', books_count: 40 },
      { id: 11, code: 'sf_action', name: 'Боевая фантастика', position: '0.1.2', books_count: 60 },
    ],
  },
  {
    id: 2, code: 'det_all', name: 'Детективы', position: '0.2', books_count: 120,
    children: [
      { id: 20, code: 'det_classic', name: 'Классический детектив', position: '0.2.0', books_count: 30 },
    ],
  },
  { id: 3, code: 'sci_all', name: 'Наука, Образование', position: '0.3', books_count: 80 },
]

function mountGenresTab() {
  return mount(GenresTab, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('GenresTab', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads genres on mount', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    mountGenresTab()
    await flushPromises()

    expect(booksApi.getGenres).toHaveBeenCalled()
  })

  it('renders root categories from tree data', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountGenresTab()
    await flushPromises()

    const text = wrapper.text()
    expect(text).toContain('Фантастика')
    expect(text).toContain('Детективы')
    expect(text).toContain('Наука, Образование')
  })

  it('shows empty state when no genres', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue([])

    const wrapper = mountGenresTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Жанры не найдены')
  })

  it('displays books_count badge for genres', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountGenresTab()
    await flushPromises()

    const countBadges = wrapper.findAll('.genres-tab__count')
    expect(countBadges.length).toBeGreaterThan(0)

    // Root items should show counts
    const text = wrapper.text()
    expect(text).toContain('250')
    expect(text).toContain('120')
  })

  it('selects genre on activation and updates catalog store', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 25,
    })

    const wrapper = mountGenresTab()
    await flushPromises()

    const store = useCatalogStore()

    // Find the VTreeview and activate an item
    const treeview = wrapper.findComponent({ name: 'VTreeview' })
    expect(treeview.exists()).toBe(true)

    // Simulate activation via emitted event
    treeview.vm.$emit('update:activated', [1])
    await flushPromises()

    expect(store.navigationFilter?.type).toBe('genre')
    expect(store.navigationFilter?.id).toBe(1)
    expect(store.navigationFilter?.label).toBe('Фантастика')
  })

  it('handles API error gracefully', async () => {
    vi.mocked(booksApi.getGenres).mockRejectedValue(new Error('Network'))

    const wrapper = mountGenresTab()
    await flushPromises()

    // Should not crash, shows empty state
    expect(wrapper.text()).toContain('Жанры не найдены')
  })

  it('renders VTreeview component', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountGenresTab()
    await flushPromises()

    const treeview = wrapper.findComponent({ name: 'VTreeview' })
    expect(treeview.exists()).toBe(true)
  })

  it('has search input for filtering genres', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountGenresTab()
    await flushPromises()

    const input = wrapper.find('input')
    expect(input.exists()).toBe(true)
    expect(input.attributes('placeholder')).toBe('Поиск жанра...')
  })

  it('passes search query to VTreeview with debounce', async () => {
    vi.useFakeTimers()
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountGenresTab()
    await flushPromises()

    const input = wrapper.find('input')
    await input.setValue('физика')
    await input.trigger('input')

    // Before debounce: search not yet applied
    const treeview = wrapper.findComponent({ name: 'VTreeview' })
    expect(treeview.props('search')).toBe('')

    // After debounce
    vi.advanceTimersByTime(300)
    await wrapper.vm.$nextTick()

    expect(treeview.props('search')).toBe('физика')

    vi.useRealTimers()
  })

  it('shows empty message when search has no matches', async () => {
    vi.useFakeTimers()
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountGenresTab()
    await flushPromises()

    const input = wrapper.find('input')
    await input.setValue('несуществующий_жанр_xyz')
    await input.trigger('input')

    vi.advanceTimersByTime(300)
    await wrapper.vm.$nextTick()

    // "Жанры не найдены" should appear for no matches
    expect(wrapper.text()).toContain('Жанры не найдены')

    vi.useRealTimers()
  })

  it('clears search and restores full tree', async () => {
    vi.useFakeTimers()
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenreTree)

    const wrapper = mountGenresTab()
    await flushPromises()

    // Type search
    const input = wrapper.find('input')
    await input.setValue('Фант')
    await input.trigger('input')
    vi.advanceTimersByTime(300)
    await wrapper.vm.$nextTick()

    // Clear search
    const clearBtn = wrapper.find('.search-input-clear')
    expect(clearBtn.exists()).toBe(true)
    await clearBtn.trigger('click')
    await wrapper.vm.$nextTick()

    // Search should be cleared
    expect(input.element.value).toBe('')
    const treeview = wrapper.findComponent({ name: 'VTreeview' })
    expect(treeview.props('search')).toBe('')

    vi.useRealTimers()
  })
})
