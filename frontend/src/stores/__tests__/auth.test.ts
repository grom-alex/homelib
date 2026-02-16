import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../auth'

vi.mock('@/services/auth', () => ({
  login: vi.fn(),
  register: vi.fn(),
  refresh: vi.fn(),
  logout: vi.fn(),
}))

import * as authApi from '@/services/auth'

const mockUser = {
  id: '1',
  email: 'test@test.com',
  username: 'testuser',
  display_name: 'Test User',
  role: 'user',
}

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    sessionStorage.clear()
    vi.clearAllMocks()
  })

  it('starts unauthenticated', () => {
    const store = useAuthStore()
    expect(store.isAuthenticated).toBe(false)
    expect(store.isAdmin).toBe(false)
    expect(store.user).toBeNull()
  })

  it('login sets user and token', async () => {
    vi.mocked(authApi.login).mockResolvedValue({
      user: mockUser,
      access_token: 'tok123',
    })
    const store = useAuthStore()
    await store.login({ email: 'test@test.com', password: 'pass' })

    expect(store.isAuthenticated).toBe(true)
    expect(store.user).toEqual(mockUser)
    expect(store.accessToken).toBe('tok123')
    expect(sessionStorage.getItem('access_token')).toBe('tok123')
  })

  it('logout clears state', async () => {
    vi.mocked(authApi.logout).mockResolvedValue(undefined)
    const store = useAuthStore()
    store.setAuth({ user: mockUser, access_token: 'tok' })

    await store.logout()
    expect(store.isAuthenticated).toBe(false)
    expect(store.user).toBeNull()
    expect(sessionStorage.getItem('access_token')).toBeNull()
  })

  it('isAdmin returns true for admin role', () => {
    const store = useAuthStore()
    store.setAuth({ user: { ...mockUser, role: 'admin' }, access_token: 'tok' })
    expect(store.isAdmin).toBe(true)
  })

  it('isAdmin returns false for user role', () => {
    const store = useAuthStore()
    store.setAuth({ user: mockUser, access_token: 'tok' })
    expect(store.isAdmin).toBe(false)
  })

  it('register sets user and token', async () => {
    vi.mocked(authApi.register).mockResolvedValue({
      user: mockUser,
      access_token: 'newtok',
    })
    const store = useAuthStore()
    await store.register({
      email: 'test@test.com',
      username: 'testuser',
      display_name: 'Test',
      password: 'password123',
    })
    expect(store.isAuthenticated).toBe(true)
    expect(store.accessToken).toBe('newtok')
  })

  it('refreshToken updates tokens on success', async () => {
    vi.mocked(authApi.refresh).mockResolvedValue({
      user: mockUser,
      access_token: 'refreshed',
    })
    const store = useAuthStore()
    store.setAuth({ user: mockUser, access_token: 'old' })

    await store.refreshToken()
    expect(store.accessToken).toBe('refreshed')
  })

  it('refreshToken clears auth on failure', async () => {
    vi.mocked(authApi.refresh).mockRejectedValue(new Error('expired'))
    const store = useAuthStore()
    store.setAuth({ user: mockUser, access_token: 'old' })

    await expect(store.refreshToken()).rejects.toThrow('Session expired')
    expect(store.isAuthenticated).toBe(false)
  })

  it('init restores token and refreshes session', async () => {
    vi.mocked(authApi.refresh).mockResolvedValue({
      user: mockUser,
      access_token: 'refreshed',
    })
    sessionStorage.setItem('access_token', 'stored')
    const store = useAuthStore()
    await store.init()
    expect(store.accessToken).toBe('refreshed')
    expect(store.user).toEqual(mockUser)
    expect(store.initialized).toBe(true)
  })

  it('init clears auth if refresh fails', async () => {
    vi.mocked(authApi.refresh).mockRejectedValue(new Error('expired'))
    sessionStorage.setItem('access_token', 'stored')
    const store = useAuthStore()
    await store.init()
    expect(store.accessToken).toBeNull()
    expect(store.user).toBeNull()
    expect(store.initialized).toBe(true)
  })

  it('init does nothing on second call', async () => {
    vi.mocked(authApi.refresh).mockResolvedValue({
      user: mockUser,
      access_token: 'refreshed',
    })
    sessionStorage.setItem('access_token', 'stored')
    const store = useAuthStore()
    await store.init()
    vi.mocked(authApi.refresh).mockClear()
    await store.init()
    expect(authApi.refresh).not.toHaveBeenCalled()
  })

  it('init skips refresh if no stored token', async () => {
    const store = useAuthStore()
    await store.init()
    expect(authApi.refresh).not.toHaveBeenCalled()
    expect(store.initialized).toBe(true)
  })

  it('clearAuth removes everything', () => {
    const store = useAuthStore()
    store.setAuth({ user: mockUser, access_token: 'tok' })
    store.clearAuth()
    expect(store.user).toBeNull()
    expect(store.accessToken).toBeNull()
  })
})
