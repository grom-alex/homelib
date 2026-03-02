import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockGet = vi.fn()
const mockPost = vi.fn()
const mockPut = vi.fn()
const mockDelete = vi.fn()
vi.mock('../client', () => ({
  default: {
    get: (...args: unknown[]) => mockGet(...args),
    post: (...args: unknown[]) => mockPost(...args),
    put: (...args: unknown[]) => mockPut(...args),
    delete: (...args: unknown[]) => mockDelete(...args),
  },
}))

import {
  getMyParentalStatus,
  unlockAdultContent,
  lockAdultContent,
  getAdminParentalStatus,
  getRestrictedGenres,
  updateRestrictedGenres,
  setParentalPin,
  removeParentalPin,
  listUsersAdultStatus,
  setUserAdultContent,
} from '../parental'

describe('parental API', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  // --- User endpoints ---

  it('getMyParentalStatus calls GET /me/parental/status', async () => {
    const status = { adult_content_enabled: true, pin_set: true }
    mockGet.mockResolvedValue({ data: status })
    const result = await getMyParentalStatus()
    expect(mockGet).toHaveBeenCalledWith('/me/parental/status')
    expect(result).toEqual(status)
  })

  it('unlockAdultContent calls POST /me/parental/unlock with pin', async () => {
    mockPost.mockResolvedValue({ data: { status: 'ok' } })
    await unlockAdultContent('1234')
    expect(mockPost).toHaveBeenCalledWith('/me/parental/unlock', { pin: '1234' })
  })

  it('lockAdultContent calls POST /me/parental/lock', async () => {
    mockPost.mockResolvedValue({ data: { status: 'ok' } })
    await lockAdultContent()
    expect(mockPost).toHaveBeenCalledWith('/me/parental/lock')
  })

  // --- Admin endpoints ---

  it('getAdminParentalStatus calls GET /admin/parental/status', async () => {
    const status = { pin_set: true, restricted_genre_codes: ['love'] }
    mockGet.mockResolvedValue({ data: status })
    const result = await getAdminParentalStatus()
    expect(mockGet).toHaveBeenCalledWith('/admin/parental/status')
    expect(result).toEqual(status)
  })

  it('getRestrictedGenres calls GET /admin/parental/genres', async () => {
    const data = { codes: ['love', 'erotica'] }
    mockGet.mockResolvedValue({ data })
    const result = await getRestrictedGenres()
    expect(mockGet).toHaveBeenCalledWith('/admin/parental/genres')
    expect(result).toEqual(data)
  })

  it('updateRestrictedGenres calls PUT /admin/parental/genres', async () => {
    mockPut.mockResolvedValue({ data: { codes: ['love'] } })
    await updateRestrictedGenres(['love'])
    expect(mockPut).toHaveBeenCalledWith('/admin/parental/genres', { codes: ['love'] })
  })

  it('setParentalPin calls POST /admin/parental/pin', async () => {
    mockPost.mockResolvedValue({ data: { status: 'ok' } })
    await setParentalPin('5678')
    expect(mockPost).toHaveBeenCalledWith('/admin/parental/pin', { pin: '5678' })
  })

  it('removeParentalPin calls DELETE /admin/parental/pin', async () => {
    mockDelete.mockResolvedValue({ data: { status: 'ok' } })
    await removeParentalPin()
    expect(mockDelete).toHaveBeenCalledWith('/admin/parental/pin')
  })

  it('listUsersAdultStatus calls GET /admin/parental/users', async () => {
    const users = [
      { user_id: '1', username: 'admin', display_name: 'Admin', role: 'admin', adult_content_enabled: true },
    ]
    mockGet.mockResolvedValue({ data: users })
    const result = await listUsersAdultStatus()
    expect(mockGet).toHaveBeenCalledWith('/admin/parental/users')
    expect(result).toEqual(users)
  })

  it('setUserAdultContent calls PUT /admin/parental/users/:id', async () => {
    mockPut.mockResolvedValue({ data: { status: 'ok' } })
    await setUserAdultContent('user-123', true)
    expect(mockPut).toHaveBeenCalledWith('/admin/parental/users/user-123', { adult_content_enabled: true })
  })
})
