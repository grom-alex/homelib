import { vi } from 'vitest'
import 'vuetify/styles'

// Polyfill ResizeObserver for jsdom
global.ResizeObserver = class ResizeObserver {
  observe = vi.fn()
  unobserve = vi.fn()
  disconnect = vi.fn()
}
