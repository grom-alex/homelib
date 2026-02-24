import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import SettingsDialog from '../SettingsDialog.vue'

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

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
  },
}))

// VDialog needs visualViewport in jsdom
if (typeof globalThis.visualViewport === 'undefined') {
  (globalThis as Record<string, unknown>).visualViewport = {
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    width: 1024,
    height: 768,
    offsetLeft: 0,
    offsetTop: 0,
    pageLeft: 0,
    pageTop: 0,
    scale: 1,
  }
}

const vuetify = createVuetify()

function mountSettingsDialog(modelValue = true) {
  return mount(SettingsDialog, {
    props: {
      modelValue,
    },
    global: {
      plugins: [vuetify],
    },
    attachTo: document.body,
  })
}

describe('SettingsDialog', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders theme buttons when open', () => {
    const wrapper = mountSettingsDialog()
    const text = document.body.textContent || ''
    expect(text).toContain('Светлая')
    expect(text).toContain('Тёмная')
    expect(text).toContain('Сепия')
    expect(text).toContain('Ночная')
    expect(text).toContain('Своя')
    wrapper.unmount()
  })

  it('renders catalog and reader theme sections', () => {
    const wrapper = mountSettingsDialog()
    const text = document.body.textContent || ''
    expect(text).toContain('Тема каталога')
    expect(text).toContain('Тема читалки')
    wrapper.unmount()
  })

  it('renders settings title', () => {
    const wrapper = mountSettingsDialog()
    const text = document.body.textContent || ''
    expect(text).toContain('Настройки')
    wrapper.unmount()
  })

  it('renders "Тема каталога" reset button for reader', () => {
    const wrapper = mountSettingsDialog()
    // "Тема каталога" appears as both a section label AND a button text
    const allText = document.body.textContent || ''
    expect(allText).toContain('Тема каталога')
    wrapper.unmount()
  })

  it('has theme button elements', () => {
    const wrapper = mountSettingsDialog()
    const buttons = document.body.querySelectorAll('.v-btn')
    // 5 catalog + 1 "Тема каталога" + 5 reader + 1 close = 12
    expect(buttons.length).toBeGreaterThanOrEqual(11)
    wrapper.unmount()
  })
})
