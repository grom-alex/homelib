import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import { createRouter, createWebHistory } from 'vue-router'
import BookCard from '../BookCard.vue'

const vuetify = createVuetify()
const router = createRouter({
  history: createWebHistory(),
  routes: [{ path: '/', component: { template: '<div />' } }],
})

function mountCard(book = {}) {
  const defaultBook = {
    id: 1,
    title: 'Test Book',
    lang: 'ru',
    format: 'fb2',
    is_deleted: false,
    authors: [{ id: 1, name: 'Author One' }],
    genres: [{ id: 1, name: 'Fiction' }],
  }
  return mount(BookCard, {
    props: { book: { ...defaultBook, ...book } },
    global: {
      plugins: [vuetify, router],
    },
  })
}

describe('BookCard', () => {
  it('renders title', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('Test Book')
  })

  it('renders author name', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('Author One')
  })

  it('renders genre chips', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('Fiction')
  })

  it('renders format chip', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('FB2')
  })

  it('renders language', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).toContain('ru')
  })

  it('renders year when present', () => {
    const wrapper = mountCard({ year: 2020 })
    expect(wrapper.text()).toContain('2020')
  })

  it('renders series when present', () => {
    const wrapper = mountCard({ series: { id: 1, name: 'My Series', num: 3 } })
    expect(wrapper.text()).toContain('My Series')
    expect(wrapper.text()).toContain('#3')
  })

  it('does not render series when absent', () => {
    const wrapper = mountCard()
    expect(wrapper.text()).not.toContain('#')
  })

  it('renders file size in MB', () => {
    const wrapper = mountCard({ file_size: 1536000 })
    expect(wrapper.text()).toContain('1.5 MB')
  })

  it('renders file size in KB', () => {
    const wrapper = mountCard({ file_size: 5120 })
    expect(wrapper.text()).toContain('5 KB')
  })

  it('renders file size in bytes', () => {
    const wrapper = mountCard({ file_size: 512 })
    expect(wrapper.text()).toContain('512 B')
  })

  it('renders multiple authors', () => {
    const wrapper = mountCard({
      authors: [
        { id: 1, name: 'Author One' },
        { id: 2, name: 'Author Two' },
      ],
    })
    expect(wrapper.text()).toContain('Author One')
    expect(wrapper.text()).toContain('Author Two')
  })
})
