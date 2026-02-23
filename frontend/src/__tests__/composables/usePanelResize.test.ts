import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { defineComponent } from 'vue'
import { usePanelResize } from '@/composables/usePanelResize'

vi.mock('@/api/client', () => ({
  default: {
    get: vi.fn(),
    put: vi.fn(),
  },
}))

import api from '@/api/client'

const STORAGE_KEY = 'homelib-panel-sizes'

function createTestComponent() {
  return defineComponent({
    setup() {
      const { sizes, onVerticalResized, onHorizontalResized } = usePanelResize()
      return { sizes, onVerticalResized, onHorizontalResized }
    },
    template: '<div />',
  })
}

describe('usePanelResize', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    vi.useFakeTimers()
    vi.mocked(api.get).mockResolvedValue({ data: {} })
    vi.mocked(api.put).mockResolvedValue({ data: {} })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('uses default sizes when no cached data', () => {
    const wrapper = mount(createTestComponent())
    expect(wrapper.vm.sizes.leftWidth).toBe(25)
    expect(wrapper.vm.sizes.tableHeight).toBe(60)
  })

  it('loads sizes from localStorage on init', () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ leftWidth: 30, tableHeight: 50 }))

    const wrapper = mount(createTestComponent())
    expect(wrapper.vm.sizes.leftWidth).toBe(30)
    expect(wrapper.vm.sizes.tableHeight).toBe(50)
  })

  it('clamps localStorage values within min/max bounds', () => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify({ leftWidth: 5, tableHeight: 95 }))

    const wrapper = mount(createTestComponent())
    // min leftWidth=10, max tableHeight=80
    expect(wrapper.vm.sizes.leftWidth).toBe(10)
    expect(wrapper.vm.sizes.tableHeight).toBe(80)
  })

  it('handles corrupted localStorage data', () => {
    localStorage.setItem(STORAGE_KEY, 'not json')

    const wrapper = mount(createTestComponent())
    expect(wrapper.vm.sizes.leftWidth).toBe(25)
    expect(wrapper.vm.sizes.tableHeight).toBe(60)
  })

  it('updates leftWidth on vertical resize', () => {
    const wrapper = mount(createTestComponent())
    wrapper.vm.onVerticalResized([{ size: 35 }, { size: 65 }])

    expect(wrapper.vm.sizes.leftWidth).toBe(35)
  })

  it('updates tableHeight on horizontal resize', () => {
    const wrapper = mount(createTestComponent())
    wrapper.vm.onHorizontalResized([{ size: 70 }, { size: 30 }])

    expect(wrapper.vm.sizes.tableHeight).toBe(70)
  })

  it('saves to localStorage immediately on resize', () => {
    const wrapper = mount(createTestComponent())
    wrapper.vm.onVerticalResized([{ size: 40 }, { size: 60 }])

    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY)!)
    expect(stored.leftWidth).toBe(40)
  })

  it('debounces server save on resize', async () => {
    const wrapper = mount(createTestComponent())

    wrapper.vm.onVerticalResized([{ size: 30 }, { size: 70 }])
    wrapper.vm.onVerticalResized([{ size: 32 }, { size: 68 }])
    wrapper.vm.onVerticalResized([{ size: 35 }, { size: 65 }])

    // No server call yet
    expect(api.put).not.toHaveBeenCalled()

    // After debounce
    vi.advanceTimersByTime(1000)
    await flushPromises()

    expect(api.put).toHaveBeenCalledTimes(1)
    expect(api.put).toHaveBeenCalledWith('/me/settings', {
      catalog: { panelSizes: { leftWidth: 35, tableHeight: 60 } },
    })
  })

  it('loads sizes from server on mount', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        catalog: { panelSizes: { leftWidth: 33, tableHeight: 55 } },
      },
    })

    const wrapper = mount(createTestComponent())
    await flushPromises()

    expect(wrapper.vm.sizes.leftWidth).toBe(33)
    expect(wrapper.vm.sizes.tableHeight).toBe(55)
  })

  it('clamps server values within bounds', async () => {
    vi.mocked(api.get).mockResolvedValue({
      data: {
        catalog: { panelSizes: { leftWidth: 2, tableHeight: 90 } },
      },
    })

    const wrapper = mount(createTestComponent())
    await flushPromises()

    expect(wrapper.vm.sizes.leftWidth).toBe(10)
    expect(wrapper.vm.sizes.tableHeight).toBe(80)
  })

  it('handles server load error gracefully', async () => {
    vi.mocked(api.get).mockRejectedValue(new Error('Network error'))

    const wrapper = mount(createTestComponent())
    await flushPromises()

    // Falls back to default
    expect(wrapper.vm.sizes.leftWidth).toBe(25)
    expect(wrapper.vm.sizes.tableHeight).toBe(60)
  })

  it('handles server save error gracefully', async () => {
    vi.mocked(api.put).mockRejectedValue(new Error('Network error'))

    const wrapper = mount(createTestComponent())
    wrapper.vm.onVerticalResized([{ size: 40 }, { size: 60 }])

    vi.advanceTimersByTime(1000)
    await flushPromises()

    // No error thrown, localStorage still has the value
    expect(wrapper.vm.sizes.leftWidth).toBe(40)
    const stored = JSON.parse(localStorage.getItem(STORAGE_KEY)!)
    expect(stored.leftWidth).toBe(40)
  })
})
