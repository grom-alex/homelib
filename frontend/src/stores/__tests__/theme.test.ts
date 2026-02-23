import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useThemeStore } from '../theme'

vi.mock('vuetify', () => ({
  useTheme: vi.fn(() => ({
    global: { name: { value: 'light' } },
  })),
}))

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
  },
}))

import api from '@/api/client'

describe('theme store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('starts with default values', () => {
    const store = useThemeStore()
    expect(store.catalogTheme).toBe('light')
    expect(store.readerThemeOverride).toBeNull()
    expect(store.loaded).toBe(false)
  })

  it('effectiveReaderTheme returns catalogTheme when no override', () => {
    const store = useThemeStore()
    store.catalogTheme = 'dark'
    expect(store.effectiveReaderTheme).toBe('dark')
  })

  it('effectiveReaderTheme returns override when set', () => {
    const store = useThemeStore()
    store.catalogTheme = 'dark'
    store.readerThemeOverride = 'sepia'
    expect(store.effectiveReaderTheme).toBe('sepia')
  })

  it('setCatalogTheme updates theme and schedules save', () => {
    const store = useThemeStore()
    store.setCatalogTheme('dark')
    expect(store.catalogTheme).toBe('dark')

    vi.advanceTimersByTime(1000)
    expect(api.put).toHaveBeenCalledWith('/me/settings', {
      catalog: expect.objectContaining({ theme: 'dark' }),
      reader: { theme: null },
    })
  })

  it('setReaderTheme updates override and schedules save', () => {
    const store = useThemeStore()
    store.setReaderTheme('night')
    expect(store.readerThemeOverride).toBe('night')

    vi.advanceTimersByTime(1000)
    expect(api.put).toHaveBeenCalledWith('/me/settings', {
      catalog: expect.objectContaining({ theme: 'light' }),
      reader: { theme: 'night' },
    })
  })

  it('resetReaderTheme sets override to null', () => {
    const store = useThemeStore()
    store.readerThemeOverride = 'sepia'
    store.resetReaderTheme()
    expect(store.readerThemeOverride).toBeNull()
    expect(store.effectiveReaderTheme).toBe('light')
  })

  it('loadSettings loads from server', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        catalog: { theme: 'night' },
        reader: { theme: 'sepia' },
      },
    })

    const store = useThemeStore()
    await store.loadSettings()

    expect(store.catalogTheme).toBe('night')
    expect(store.readerThemeOverride).toBe('sepia')
    expect(store.loaded).toBe(true)
  })

  it('loadSettings uses defaults on error', async () => {
    vi.mocked(api.get).mockRejectedValue(new Error('Network'))

    const store = useThemeStore()
    await store.loadSettings()

    expect(store.catalogTheme).toBe('light')
    expect(store.readerThemeOverride).toBeNull()
    expect(store.loaded).toBe(true)
  })

  it('loadSettings handles null reader theme as no override', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        catalog: { theme: 'dark' },
        reader: { theme: null },
      },
    })

    const store = useThemeStore()
    await store.loadSettings()

    expect(store.readerThemeOverride).toBeNull()
    expect(store.effectiveReaderTheme).toBe('dark')
  })

  it('debounces save — only last call fires', () => {
    const store = useThemeStore()
    store.setCatalogTheme('dark')
    store.setCatalogTheme('sepia')
    store.setCatalogTheme('night')

    vi.advanceTimersByTime(1000)
    expect(api.put).toHaveBeenCalledTimes(1)
    expect(api.put).toHaveBeenCalledWith('/me/settings', {
      catalog: expect.objectContaining({ theme: 'night' }),
      reader: { theme: null },
    })
  })

  it('saveSettings handles error silently', async () => {
    vi.mocked(api.put).mockRejectedValue(new Error('Disk full'))

    const store = useThemeStore()
    // Should not throw
    await store.saveSettings()
  })

  it('loadSettings handles missing catalog section', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: { reader: { fontSize: 18 } },
    })

    const store = useThemeStore()
    await store.loadSettings()

    expect(store.catalogTheme).toBe('light')
    expect(store.loaded).toBe(true)
  })
})
