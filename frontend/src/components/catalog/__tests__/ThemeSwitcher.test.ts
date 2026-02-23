import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import ThemeSwitcher from '../ThemeSwitcher.vue'
import { useThemeStore } from '@/stores/theme'

vi.mock('vuetify', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vuetify')>()
  return {
    ...actual,
    useTheme: vi.fn(() => ({
      global: { name: { value: 'light' } },
    })),
  }
})

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
  },
}))

const vuetify = createVuetify()

function mountThemeSwitcher() {
  return mount(ThemeSwitcher, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('ThemeSwitcher', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders 4 theme options', () => {
    const wrapper = mountThemeSwitcher()
    expect(wrapper.text()).toContain('Светлая')
    expect(wrapper.text()).toContain('Тёмная')
    expect(wrapper.text()).toContain('Сепия')
    expect(wrapper.text()).toContain('Ночная')
  })

  it('switches theme on click', async () => {
    const wrapper = mountThemeSwitcher()
    const store = useThemeStore()

    const darkBtn = wrapper.findAll('.v-btn').find((b) => b.text().includes('Тёмная'))
    await darkBtn!.trigger('click')

    expect(store.catalogTheme).toBe('dark')
  })
})
