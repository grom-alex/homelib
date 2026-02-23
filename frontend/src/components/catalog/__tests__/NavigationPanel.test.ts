import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import NavigationPanel from '../NavigationPanel.vue'
import { useCatalogStore } from '@/stores/catalog'

vi.mock('@/api/books', () => ({
  getAuthors: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, limit: 50 }),
  getSeries: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, limit: 50 }),
  getGenres: vi.fn().mockResolvedValue([]),
  getStats: vi.fn().mockResolvedValue({ books_count: 0, authors_count: 0, genres_count: 0, series_count: 0, languages: [], formats: [] }),
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

const vuetify = createVuetify()

function mountNavigationPanel() {
  return mount(NavigationPanel, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('NavigationPanel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows AuthorsTab by default', async () => {
    const wrapper = mountNavigationPanel()
    await flushPromises()
    expect(wrapper.findComponent({ name: 'AuthorsTab' }).exists()).toBe(true)
  })

  it('shows AuthorsTab when activeTab is authors', async () => {
    const wrapper = mountNavigationPanel()
    await flushPromises()
    expect(wrapper.findComponent({ name: 'AuthorsTab' }).exists()).toBe(true)
  })

  it('shows SeriesTab when activeTab is series', async () => {
    const store = useCatalogStore()
    store.activeTab = 'series'

    const wrapper = mountNavigationPanel()
    await flushPromises()
    expect(wrapper.findComponent({ name: 'SeriesTab' }).exists()).toBe(true)
  })

  it('switches components when tab changes', async () => {
    const store = useCatalogStore()
    const wrapper = mountNavigationPanel()
    await flushPromises()

    expect(wrapper.findComponent({ name: 'AuthorsTab' }).exists()).toBe(true)

    store.activeTab = 'genres'
    await wrapper.vm.$nextTick()
    expect(wrapper.findComponent({ name: 'GenresTab' }).exists()).toBe(true)
  })
})
