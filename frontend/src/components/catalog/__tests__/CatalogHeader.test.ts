import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import CatalogHeader from '../CatalogHeader.vue'
import { useCatalogStore } from '@/stores/catalog'
import { useAuthStore } from '@/stores/auth'

vi.mock('vuetify', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vuetify')>()
  return {
    ...actual,
    useTheme: vi.fn(() => ({
      global: { name: { value: 'light' } },
      themes: { value: { custom: { dark: false, colors: {} } } },
    })),
  }
})

vi.mock('@/api/books', () => ({
  getStats: vi.fn().mockResolvedValue({ books_count: 0 }),
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
    post: vi.fn(),
    delete: vi.fn(),
  },
  setOnAuthExpired: vi.fn(),
}))

import * as booksApi from '@/api/books'

const vuetify = createVuetify()

function mountCatalogHeader() {
  return mount(CatalogHeader, {
    global: {
      plugins: [vuetify],
      stubs: {
        ThemeSwitcher: { template: '<div class="theme-switcher-stub" />' },
        SettingsDialog: { template: '<div class="settings-dialog-stub" />' },
      },
    },
  })
}

describe('CatalogHeader', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.mocked(booksApi.getStats).mockResolvedValue({
      books_count: 0,
      authors_count: 0,
      genres_count: 0,
      series_count: 0,
      languages: [],
      formats: [],
    })
  })

  it('renders logo', () => {
    const wrapper = mountCatalogHeader()
    expect(wrapper.text()).toContain('MyHomeLib')
    expect(wrapper.text()).toContain('web')
  })

  it('renders all navigation tabs', () => {
    const wrapper = mountCatalogHeader()
    expect(wrapper.text()).toContain('Авторы')
    expect(wrapper.text()).toContain('Серии')
    expect(wrapper.text()).toContain('Жанры')
    expect(wrapper.text()).toContain('Поиск')
  })

  it('sets active tab class on current tab', () => {
    const store = useCatalogStore()
    store.activeTab = 'series'

    const wrapper = mountCatalogHeader()
    const activeTab = wrapper.find('.catalog-header__tab--active')
    expect(activeTab.exists()).toBe(true)
    expect(activeTab.text()).toContain('Серии')
  })

  it('switches tab on click', async () => {
    const store = useCatalogStore()
    const wrapper = mountCatalogHeader()

    const tabs = wrapper.findAll('.catalog-header__tab')
    await tabs[1].trigger('click') // Series tab

    expect(store.activeTab).toBe('series')
  })

  it('shows books count when available', async () => {
    vi.mocked(booksApi.getStats).mockResolvedValue({
      books_count: 1500,
      authors_count: 50,
      genres_count: 20,
      series_count: 10,
      languages: [],
      formats: [],
    })

    const wrapper = mountCatalogHeader()
    await flushPromises()

    expect(wrapper.text()).toContain('Книг:')
    expect(wrapper.text()).toContain('2K')
  })

  it('hides books count when 0', async () => {
    const wrapper = mountCatalogHeader()
    await flushPromises()

    expect(wrapper.find('.catalog-header__count').exists()).toBe(false)
  })

  it('displays user name from auth store', () => {
    const auth = useAuthStore()
    auth.user = { id: 1, username: 'testuser', display_name: 'Test User', email: 'test@test.com', role: 'user' } as never

    const wrapper = mountCatalogHeader()
    expect(wrapper.text()).toContain('Test User')
  })

  it('shows user initials', () => {
    const auth = useAuthStore()
    auth.user = { id: 1, username: 'testuser', display_name: 'Test User', email: 'test@test.com', role: 'user' } as never

    const wrapper = mountCatalogHeader()
    expect(wrapper.text()).toContain('TU')
  })

  it('shows single-name initials (first 2 chars)', () => {
    const auth = useAuthStore()
    auth.user = { id: 1, username: 'admin', display_name: 'Admin', email: 'a@b.com', role: 'admin' } as never

    const wrapper = mountCatalogHeader()
    expect(wrapper.text()).toContain('AD')
  })

  it('shows ? when no display name', () => {
    const auth = useAuthStore()
    auth.user = { id: 1, username: '', display_name: '', email: '', role: 'user' } as never

    const wrapper = mountCatalogHeader()
    expect(wrapper.find('.catalog-header__avatar').text()).toBe('?')
  })

  it('opens and closes user menu on click', async () => {
    const auth = useAuthStore()
    auth.user = { id: 1, username: 'testuser', display_name: 'Test User', email: 'test@test.com', role: 'user' } as never

    const wrapper = mountCatalogHeader()

    expect(wrapper.find('.catalog-header__dropdown').exists()).toBe(false)

    await wrapper.find('.catalog-header__user-btn').trigger('click')
    expect(wrapper.find('.catalog-header__dropdown').exists()).toBe(true)

    // Close via overlay
    await wrapper.find('.catalog-header__overlay').trigger('click')
    expect(wrapper.find('.catalog-header__dropdown').exists()).toBe(false)
  })

  it('menu contains settings and logout items', async () => {
    const auth = useAuthStore()
    auth.user = { id: 1, username: 'testuser', display_name: 'Test User', email: 'test@test.com', role: 'user' } as never

    const wrapper = mountCatalogHeader()
    await wrapper.find('.catalog-header__user-btn').trigger('click')

    expect(wrapper.text()).toContain('Настройки')
    expect(wrapper.text()).toContain('Выйти')
  })

  it('formatCount formats thousands', async () => {
    vi.mocked(booksApi.getStats).mockResolvedValue({
      books_count: 650000,
      authors_count: 0,
      genres_count: 0,
      series_count: 0,
      languages: [],
      formats: [],
    })

    const wrapper = mountCatalogHeader()
    await flushPromises()

    expect(wrapper.text()).toContain('650K')
  })

  it('formatCount formats millions', async () => {
    vi.mocked(booksApi.getStats).mockResolvedValue({
      books_count: 2500000,
      authors_count: 0,
      genres_count: 0,
      series_count: 0,
      languages: [],
      formats: [],
    })

    const wrapper = mountCatalogHeader()
    await flushPromises()

    expect(wrapper.text()).toContain('2.5M')
  })

  it('handles stats API failure gracefully', async () => {
    vi.mocked(booksApi.getStats).mockRejectedValue(new Error('Network'))

    const wrapper = mountCatalogHeader()
    await flushPromises()

    // Should not crash, count not shown
    expect(wrapper.find('.catalog-header__count').exists()).toBe(false)
  })

  it('opens settings dialog via menu item', async () => {
    const auth = useAuthStore()
    auth.user = { id: 1, username: 'testuser', display_name: 'Test User', email: 'test@test.com', role: 'user' } as never

    const wrapper = mountCatalogHeader()
    await wrapper.find('.catalog-header__user-btn').trigger('click')

    const items = wrapper.findAll('.catalog-header__dropdown-item')
    const settingsItem = items.find(i => i.text().includes('Настройки'))
    expect(settingsItem).toBeDefined()
    await settingsItem!.trigger('click')

    // Menu should close
    expect(wrapper.find('.catalog-header__dropdown').exists()).toBe(false)
  })
})
