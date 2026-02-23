import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createVuetify } from 'vuetify'
import BookDetailPanel from '../BookDetailPanel.vue'
import { useCatalogStore } from '@/stores/catalog'

vi.mock('@/api/books', () => ({
  getBooks: vi.fn(),
  getBook: vi.fn(),
  downloadBook: vi.fn(),
}))

const vuetify = createVuetify()

function mountBookDetailPanel() {
  return mount(BookDetailPanel, {
    global: {
      plugins: [vuetify],
    },
  })
}

const mockBookDetail = {
  id: 1,
  title: 'Основание',
  lang: 'ru',
  year: 1951,
  format: 'fb2',
  file_size: 524288,
  is_deleted: false,
  description: 'Великолепная книга о будущем.',
  authors: [{ id: 1, name: 'Азимов, Айзек' }],
  genres: [{ id: 1, code: 'sf', name: 'Фантастика' }],
  series: { id: 1, name: 'Основание', num: 1 },
}

describe('BookDetailPanel', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('shows empty state when no book selected', () => {
    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('Выберите книгу для просмотра подробной информации')
  })

  it('shows loading spinner when book loading', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.bookLoading = true

    const wrapper = mountBookDetailPanel()
    expect(wrapper.find('.book-detail-panel__loading').exists()).toBe(true)
  })

  it('shows book details when book loaded', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = mockBookDetail as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('Основание')
    expect(wrapper.text()).toContain('Азимов, Айзек')
    expect(wrapper.text()).toContain('Фантастика')
    expect(wrapper.text()).toContain('1951')
    expect(wrapper.text()).toContain('fb2')
    expect(wrapper.text()).toContain('512 KB')
    expect(wrapper.text()).toContain('ru')
  })

  it('shows series with number', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = mockBookDetail as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('Основание')
    expect(wrapper.text()).toContain('#1')
  })

  it('shows dash when no series', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, series: undefined } as never

    const wrapper = mountBookDetailPanel()
    const text = wrapper.text()
    expect(text).toContain('Серия:')
  })

  it('shows description', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = mockBookDetail as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('Великолепная книга о будущем.')
  })

  it('shows empty annotation message when no description', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, description: undefined } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('Аннотация отсутствует')
  })

  it('shows read button for fb2 books', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = mockBookDetail as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('Читать')
  })

  it('hides read button for non-fb2 books', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, format: 'epub' } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.find('.book-detail-panel__btn--primary').exists()).toBe(false)
  })

  it('shows download button with format', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = mockBookDetail as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('Скачать')
  })

  it('formats multiple authors with comma', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = {
      ...mockBookDetail,
      authors: [
        { id: 1, name: 'Азимов, Айзек' },
        { id: 2, name: 'Кларк, Артур' },
      ],
    } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('Азимов, Айзек, Кларк, Артур')
  })

  it('shows dash for no authors', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, authors: [] } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.find('.book-detail-panel__author').text()).toBe('—')
  })

  it('shows dash for no genres', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, genres: [] } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('—')
  })

  it('formats file size in bytes', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, file_size: 500 } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('500 B')
  })

  it('formats file size in MB', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, file_size: 2097152 } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('2.0 MB')
  })

  it('shows dash for no file size', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, file_size: undefined } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).toContain('—')
  })

  it('calls downloadBook on download click', async () => {
    const { downloadBook } = await import('@/api/books')
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = mockBookDetail as never

    const wrapper = mountBookDetailPanel()
    const downloadBtn = wrapper.find('.book-detail-panel__btn--secondary')
    await downloadBtn.trigger('click')

    expect(downloadBook).toHaveBeenCalledWith(1)
  })

  it('hides year when not present', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, year: undefined } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).not.toContain('Год:')
  })

  it('hides lang when not present', () => {
    const store = useCatalogStore()
    store.selectedBookId = 1
    store.currentBook = { ...mockBookDetail, lang: '' } as never

    const wrapper = mountBookDetailPanel()
    expect(wrapper.text()).not.toContain('Язык:')
  })
})
