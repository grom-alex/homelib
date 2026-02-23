import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import BookTable from '../BookTable.vue'
import { useCatalogStore } from '@/stores/catalog'

vi.mock('@/api/books', () => ({
  getBooks: vi.fn(),
  getBook: vi.fn(),
}))

const vuetify = createVuetify()

function mountBookTable() {
  return mount(BookTable, {
    global: {
      plugins: [vuetify],
    },
  })
}

const mockBooks = [
  {
    id: 1,
    title: 'Основание',
    lang: 'ru',
    format: 'fb2',
    file_size: 524288,
    is_deleted: false,
    authors: [{ id: 1, name: 'Азимов, Айзек' }],
    genres: [{ id: 1, name: 'Фантастика' }],
    series: { id: 1, name: 'Основание', num: 1 },
  },
  {
    id: 2,
    title: 'Дюна',
    lang: 'ru',
    format: 'fb2',
    file_size: 1048576,
    is_deleted: false,
    authors: [{ id: 2, name: 'Герберт, Фрэнк' }, { id: 3, name: 'Другой Автор' }],
    genres: [{ id: 1, name: 'Фантастика' }],
  },
]

describe('BookTable', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows empty state when no navigation filter', () => {
    const wrapper = mountBookTable()
    expect(wrapper.text()).toContain('Выберите элемент навигации')
  })

  it('shows "not found" when filter active but no books', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = []

    const wrapper = mountBookTable()
    expect(wrapper.text()).toContain('Книги не найдены')
  })

  it('renders book rows', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]

    const wrapper = mountBookTable()
    const rows = wrapper.findAll('.book-table__row')
    expect(rows).toHaveLength(2)
    expect(rows[0].text()).toContain('Основание')
    expect(rows[1].text()).toContain('Дюна')
  })

  it('formats authors with "и др." for multiple', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]

    const wrapper = mountBookTable()
    const rows = wrapper.findAll('.book-table__row')
    expect(rows[0].text()).toContain('Азимов, Айзек')
    expect(rows[1].text()).toContain('Герберт, Фрэнк и др.')
  })

  it('formats series with number', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]

    const wrapper = mountBookTable()
    const rows = wrapper.findAll('.book-table__row')
    expect(rows[0].text()).toContain('Основание #1')
  })

  it('formats file size', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]

    const wrapper = mountBookTable()
    expect(wrapper.text()).toContain('512 KB')
    expect(wrapper.text()).toContain('1.0 MB')
  })

  it('highlights selected row', async () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.selectedBookId = 1

    const wrapper = mountBookTable()
    const selected = wrapper.find('.book-table__row--selected')
    expect(selected.exists()).toBe(true)
    expect(selected.text()).toContain('Основание')
  })

  it('renders column headers', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]

    const wrapper = mountBookTable()
    const headers = wrapper.findAll('.book-table__header-cell')
    expect(headers).toHaveLength(5)
    expect(headers[0].text()).toContain('Название')
    expect(headers[1].text()).toContain('Автор')
    expect(headers[4].text()).toContain('Размер')
  })

  it('shows pagination when multiple pages', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 100

    const wrapper = mountBookTable()
    expect(wrapper.find('.book-table__pagination').exists()).toBe(true)
  })

  it('does not show pagination for single page', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 2

    const wrapper = mountBookTable()
    expect(wrapper.find('.book-table__pagination').exists()).toBe(false)
  })
})
