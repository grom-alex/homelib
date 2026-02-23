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
    expect(headers).toHaveLength(7)
    expect(headers[0].text()).toContain('Название')
    expect(headers[1].text()).toContain('Автор')
    expect(headers[4].text()).toContain('Язык')
    expect(headers[5].text()).toContain('Формат')
    expect(headers[6].text()).toContain('Размер')
  })

  it('shows pagination when multiple pages', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 100

    const wrapper = mountBookTable()
    expect(wrapper.find('.book-table__pagination').exists()).toBe(true)
  })

  it('shows pagination bar but hides nav for single page', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 2

    const wrapper = mountBookTable()
    expect(wrapper.find('.book-table__pagination').exists()).toBe(true)
    expect(wrapper.find('.pagination__size-select').exists()).toBe(true)
    expect(wrapper.find('.pagination__nav').exists()).toBe(false)
  })

  it('renders page size selector with options', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 100

    const wrapper = mountBookTable()
    const select = wrapper.find('.pagination__size-select')
    expect(select.exists()).toBe(true)

    const options = select.findAll('option')
    expect(options).toHaveLength(4)
    expect(options.map(o => Number(o.element.value))).toEqual([25, 50, 75, 100])
  })

  it('renders navigation buttons', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 100

    const wrapper = mountBookTable()
    const nav = wrapper.find('.pagination__nav')
    expect(nav.exists()).toBe(true)

    const buttons = nav.findAll('.pagination__btn')
    // first, -10, prev, next, +10, last = 6 buttons
    expect(buttons).toHaveLength(6)
  })

  it('renders page number buttons (max 10)', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 1000 // 40 pages with limit 25

    const wrapper = mountBookTable()
    const pages = wrapper.findAll('.pagination__page')
    expect(pages).toHaveLength(10)
    expect(pages[0].text()).toBe('1')
    expect(pages[9].text()).toBe('10')
  })

  it('shows all page buttons when total pages <= 10', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 75 // 3 pages with limit 25

    const wrapper = mountBookTable()
    const pages = wrapper.findAll('.pagination__page')
    expect(pages).toHaveLength(3)
  })

  it('highlights current page', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 100

    const wrapper = mountBookTable()
    const activePage = wrapper.find('.pagination__page--active')
    expect(activePage.exists()).toBe(true)
    expect(activePage.text()).toBe('1')
  })

  it('disables prev/first buttons on first page', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 100
    store.filters.page = 1

    const wrapper = mountBookTable()
    const buttons = wrapper.findAll('.pagination__btn')
    // first 3 buttons: first, -10, prev should be disabled
    expect((buttons[0].element as HTMLButtonElement).disabled).toBe(true)
    expect((buttons[1].element as HTMLButtonElement).disabled).toBe(true)
    expect((buttons[2].element as HTMLButtonElement).disabled).toBe(true)
  })

  it('disables next/last buttons on last page', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 50 // 2 pages
    store.filters.page = 2

    const wrapper = mountBookTable()
    const buttons = wrapper.findAll('.pagination__btn')
    // last 3 buttons: next, +10, last should be disabled
    expect((buttons[3].element as HTMLButtonElement).disabled).toBe(true)
    expect((buttons[4].element as HTMLButtonElement).disabled).toBe(true)
    expect((buttons[5].element as HTMLButtonElement).disabled).toBe(true)
  })

  it('shows page info text', () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 100

    const wrapper = mountBookTable()
    const info = wrapper.find('.pagination__info')
    expect(info.exists()).toBe(true)
    expect(info.text()).toContain('Стр. 1 из 4')
  })

  it('changes page size on select change', async () => {
    const store = useCatalogStore()
    store.navigationFilter = { type: 'author', id: 1 }
    store.books = mockBooks as never[]
    store.total = 100
    store.setPageSize = vi.fn()

    const wrapper = mountBookTable()
    const select = wrapper.find('.pagination__size-select')
    await select.setValue('50')

    expect(store.setPageSize).toHaveBeenCalledWith(50)
  })
})
