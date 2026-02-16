import { describe, it, expect, vi, beforeEach } from 'vitest'

const mockPost = vi.fn()
vi.mock('../client', () => ({
  default: {
    post: (...args: unknown[]) => mockPost(...args),
  },
}))

import { login, register, refresh, logout } from '../auth'

describe('auth service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('login calls POST /auth/login', async () => {
    const response = { user: { id: '1', email: 'a@b.com', username: 'u', display_name: 'U', role: 'user' }, access_token: 'tok' }
    mockPost.mockResolvedValue({ data: response })
    const result = await login({ email: 'a@b.com', password: 'pass' })
    expect(mockPost).toHaveBeenCalledWith('/auth/login', { email: 'a@b.com', password: 'pass' })
    expect(result).toEqual(response)
  })

  it('register calls POST /auth/register', async () => {
    const response = { user: { id: '1', email: 'a@b.com', username: 'u', display_name: 'U', role: 'user' }, access_token: 'tok' }
    mockPost.mockResolvedValue({ data: response })
    const result = await register({ email: 'a@b.com', username: 'u', display_name: 'U', password: 'pass' })
    expect(mockPost).toHaveBeenCalledWith('/auth/register', { email: 'a@b.com', username: 'u', display_name: 'U', password: 'pass' })
    expect(result).toEqual(response)
  })

  it('refresh calls POST /auth/refresh', async () => {
    const response = { user: { id: '1', email: 'a@b.com', username: 'u', display_name: 'U', role: 'user' }, access_token: 'new' }
    mockPost.mockResolvedValue({ data: response })
    const result = await refresh()
    expect(mockPost).toHaveBeenCalledWith('/auth/refresh')
    expect(result).toEqual(response)
  })

  it('logout calls POST /auth/logout', async () => {
    mockPost.mockResolvedValue({ data: {} })
    await logout()
    expect(mockPost).toHaveBeenCalledWith('/auth/logout')
  })
})
