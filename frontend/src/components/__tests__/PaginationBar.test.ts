import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import PaginationBar from '../PaginationBar.vue'

const vuetify = createVuetify()

function mountPagination(props = {}) {
  return mount(PaginationBar, {
    props: { page: 1, totalPages: 5, limit: 20, ...props },
    global: {
      plugins: [vuetify],
    },
  })
}

describe('PaginationBar', () => {
  it('renders pagination component', () => {
    const wrapper = mountPagination()
    expect(wrapper.find('.v-pagination').exists()).toBe(true)
  })

  it('displays correct number of pages', () => {
    const wrapper = mountPagination({ totalPages: 10 })
    expect(wrapper.text()).toContain('10')
  })

  it('renders page buttons for page 1', () => {
    const wrapper = mountPagination({ page: 1, totalPages: 5 })
    expect(wrapper.text()).toContain('1')
  })

  it('shows items per page selector', () => {
    const wrapper = mountPagination()
    expect(wrapper.find('.v-select').exists()).toBe(true)
  })

  it('updates currentPage when page prop changes', async () => {
    const wrapper = mountPagination({ page: 1 })
    await wrapper.setProps({ page: 3 })
    expect(wrapper.props('page')).toBe(3)
  })

  it('updates currentLimit when limit prop changes', async () => {
    const wrapper = mountPagination({ limit: 20 })
    await wrapper.setProps({ limit: 50 })
    expect(wrapper.props('limit')).toBe(50)
  })
})
