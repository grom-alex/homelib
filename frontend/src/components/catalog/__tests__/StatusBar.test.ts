import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import StatusBar from '../StatusBar.vue'
import { useCatalogStore } from '@/stores/catalog'

vi.mock('@/api/books', () => ({
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

const vuetify = createVuetify()

function mountStatusBar() {
  return mount(StatusBar, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('StatusBar', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows "Готов" when no navigation filter', () => {
    const wrapper = mountStatusBar()
    expect(wrapper.text()).toContain('Готов')
  })

  it('shows author filter context', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1, label: 'Азимов, Айзек' }

    const wrapper = mountStatusBar()
    expect(wrapper.text()).toContain('Автор: Азимов, Айзек')
  })

  it('shows author id when no label', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 42 }

    const wrapper = mountStatusBar()
    expect(wrapper.text()).toContain('Автор: 42')
  })

  it('shows series filter context', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'series', id: 1, label: 'Основание' }

    const wrapper = mountStatusBar()
    expect(wrapper.text()).toContain('Серия: Основание')
  })

  it('shows genre filter context', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'genre', id: 1, label: 'Фантастика' }

    const wrapper = mountStatusBar()
    expect(wrapper.text()).toContain('Жанр: Фантастика')
  })

  it('shows search filter context', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'search', label: 'Дюна' }

    const wrapper = mountStatusBar()
    expect(wrapper.text()).toContain('Поиск: Дюна')
  })

  it('shows search without label', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'search' }

    const wrapper = mountStatusBar()
    expect(wrapper.text()).toContain('Поиск:')
  })

  it('shows book count when total > 0', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = [{ id: 1 }, { id: 2 }] as never[]
    store.total = 10

    const wrapper = mountStatusBar()
    expect(wrapper.text()).toContain('Показано книг: 2 из 10')
  })

  it('hides book count when total is 0', () => {
    const store = useCatalogStore()
    store.total = 0

    const wrapper = mountStatusBar()
    expect(wrapper.text()).not.toContain('Показано книг')
  })
})
