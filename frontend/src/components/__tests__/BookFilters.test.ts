import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import BookFilters from '../BookFilters.vue'

vi.mock('@/services/books', () => ({
  getStats: vi.fn().mockResolvedValue({
    books_count: 100,
    authors_count: 50,
    genres_count: 20,
    series_count: 10,
    languages: ['ru', 'en'],
    formats: ['fb2', 'epub'],
  }),
}))

const vuetify = createVuetify()

function mountFilters() {
  return mount(BookFilters, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('BookFilters', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
  })

  it('renders filter fields', () => {
    const wrapper = mountFilters()
    expect(wrapper.text()).toContain('Фильтры')
    expect(wrapper.text()).toContain('Сбросить фильтры')
  })

  it('emits update event on filter change', async () => {
    const wrapper = mountFilters()
    const input = wrapper.find('input')
    await input.setValue('test query')
    vi.advanceTimersByTime(300)
    await wrapper.vm.$nextTick()
    const emitted = wrapper.emitted('update')
    expect(emitted).toBeTruthy()
  })

  it('emits update on clear filters click', async () => {
    const wrapper = mountFilters()
    const btn = wrapper.findAll('button').find(b => b.text().includes('Сбросить'))
    if (btn) {
      await btn.trigger('click')
      const emitted = wrapper.emitted('update')
      expect(emitted).toBeTruthy()
    }
  })
})
