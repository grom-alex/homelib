import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockPost = vi.fn()
const mockGet = vi.fn()
vi.mock('../client', () => ({
  default: {
    post: (...args: unknown[]) => mockPost(...args),
    get: (...args: unknown[]) => mockGet(...args),
  },
}))

import { startImport, getImportStatus } from '../admin'

describe('admin service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('startImport calls POST /admin/import', async () => {
    const status = { status: 'running' }
    mockPost.mockResolvedValue({ data: status })
    const result = await startImport()
    expect(mockPost).toHaveBeenCalledWith('/admin/import')
    expect(result).toEqual(status)
  })

  it('getImportStatus calls GET /admin/import/status', async () => {
    const status = { status: 'completed', stats: { books_added: 10 } }
    mockGet.mockResolvedValue({ data: status })
    const result = await getImportStatus()
    expect(mockGet).toHaveBeenCalledWith('/admin/import/status')
    expect(result).toEqual(status)
  })
})
