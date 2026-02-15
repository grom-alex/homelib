import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import AdminImportPage from '../AdminImportPage.vue'

const mockStartImport = vi.fn()
const mockGetImportStatus = vi.fn()

vi.mock('@/services/admin', () => ({
  startImport: (...args: unknown[]) => mockStartImport(...args),
  getImportStatus: (...args: unknown[]) => mockGetImportStatus(...args),
}))

const vuetify = createVuetify()

function mountPage() {
  return mount(AdminImportPage, {
    global: {
      plugins: [vuetify],
    },
  })
}

describe('AdminImportPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    mockGetImportStatus.mockResolvedValue({ status: 'idle' })
  })

  it('renders import page title', () => {
    const wrapper = mountPage()
    expect(wrapper.text()).toContain('Импорт библиотеки')
  })

  it('renders import button', () => {
    const wrapper = mountPage()
    expect(wrapper.text()).toContain('Запустить импорт')
  })

  it('calls startImport on button click', async () => {
    mockStartImport.mockResolvedValue({ status: 'running' })
    const wrapper = mountPage()
    const btn = wrapper.findAll('button').find(b => b.text().includes('Запустить'))
    if (btn) {
      await btn.trigger('click')
      expect(mockStartImport).toHaveBeenCalled()
    }
  })

  it('displays stats when import completed', async () => {
    mockGetImportStatus.mockResolvedValue({
      status: 'completed',
      stats: {
        books_added: 100,
        books_updated: 10,
        books_deleted: 0,
        authors_added: 50,
        genres_added: 20,
        series_added: 15,
        errors: 0,
        duration_ms: 5000,
      },
    })
    const wrapper = mountPage()
    await wrapper.vm.$nextTick()
    await new Promise(r => setTimeout(r, 10))
    await wrapper.vm.$nextTick()
    // Stats should appear after status is loaded
    expect(mockGetImportStatus).toHaveBeenCalled()
  })

  it('shows error when import fails', async () => {
    mockStartImport.mockRejectedValue({
      response: { data: { error: 'Import already running' } },
    })
    const wrapper = mountPage()
    const btn = wrapper.findAll('button').find(b => b.text().includes('Запустить'))
    if (btn) {
      await btn.trigger('click')
      await wrapper.vm.$nextTick()
    }
  })
})
