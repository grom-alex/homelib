import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import SearchBar from '../SearchBar.vue'

const vuetify = createVuetify()

function mountSearchBar() {
  return mount(SearchBar, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('SearchBar', () => {
  beforeEach(() => {
    vi.useFakeTimers()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('renders search input', () => {
    const wrapper = mountSearchBar()
    expect(wrapper.find('input').exists()).toBe(true)
  })

  it('emits search event after debounce', async () => {
    const wrapper = mountSearchBar()
    const input = wrapper.find('input')
    await input.setValue('hello')
    vi.advanceTimersByTime(300)
    await wrapper.vm.$nextTick()
    const emitted = wrapper.emitted('search')
    expect(emitted).toBeTruthy()
    expect(emitted![emitted!.length - 1]).toEqual(['hello'])
  })

  it('debounces rapid input — only emits last value', async () => {
    const wrapper = mountSearchBar()
    const input = wrapper.find('input')
    await input.setValue('h')
    vi.advanceTimersByTime(100)
    await input.setValue('he')
    vi.advanceTimersByTime(100)
    await input.setValue('hel')
    vi.advanceTimersByTime(300)
    await wrapper.vm.$nextTick()
    const emitted = wrapper.emitted('search')
    expect(emitted).toBeTruthy()
    const lastEmit = emitted![emitted!.length - 1]
    expect(lastEmit).toEqual(['hel'])
  })

  it('emits empty string on clear', async () => {
    const wrapper = mountSearchBar()
    const input = wrapper.find('input')
    await input.setValue('test')
    vi.advanceTimersByTime(300)
    // Trigger click:clear
    const clearBtn = wrapper.find('.v-field__clearable button')
    if (clearBtn.exists()) {
      await clearBtn.trigger('click')
      await wrapper.vm.$nextTick()
      const emitted = wrapper.emitted('search')
      expect(emitted).toBeTruthy()
      expect(emitted![emitted!.length - 1]).toEqual([''])
    }
  })
})
