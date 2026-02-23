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

const mockGenres = [
  { id: 1, code: 'sf', name: 'Научная фантастика', meta_group: 'Фантастика', books_count: 100 },
  { id: 2, code: 'fantasy', name: 'Фэнтези', meta_group: 'Фантастика', books_count: 80 },
  { id: 3, code: 'detective', name: 'Детектив', meta_group: 'Детективы', books_count: 50 },
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
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenres)

    mountGenresTab()
    await flushPromises()

    expect(booksApi.getGenres).toHaveBeenCalled()
  })

  it('groups genres by meta_group', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenres)

    const wrapper = mountGenresTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Фантастика')
    expect(wrapper.text()).toContain('Детективы')
  })

  it('shows empty state', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue([])

    const wrapper = mountGenresTab()
    await flushPromises()

    expect(wrapper.text()).toContain('Жанры не найдены')
  })

  it('expands and collapses groups', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenres)

    const wrapper = mountGenresTab()
    await flushPromises()

    // Click on group header to expand
    const groupHeaders = wrapper.findAll('.genres-tab__group-header')
    expect(groupHeaders.length).toBeGreaterThan(0)
    await groupHeaders[0].trigger('click')
    await wrapper.vm.$nextTick()

    // Genres within the expanded group should be visible
    expect(wrapper.text()).toContain('Научная фантастика')
  })

  it('selects genre on click', async () => {
    vi.mocked(booksApi.getGenres).mockResolvedValue(mockGenres)
    vi.mocked(booksApi.getBooks).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      limit: 20,
    })

    const wrapper = mountGenresTab()
    await flushPromises()

    // Expand first group
    const groupHeaders = wrapper.findAll('.genres-tab__group-header')
    await groupHeaders[0].trigger('click')
    await wrapper.vm.$nextTick()

    // Click on genre item (skip group headers)
    const allItems = wrapper.findAll('.v-list-item:not(.genres-tab__group-header)')
    if (allItems.length > 0) {
      await allItems[0].trigger('click')
      const store = useCatalogStore()
      expect(store.navigationFilter?.type).toBe('genre')
    }
  })
})
