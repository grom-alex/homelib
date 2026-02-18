import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAuthStore } from '../auth'

vi.mock('@/api/auth', () => ({
  login: vi.fn(),
  register: vi.fn(),
  refresh: vi.fn(),
  logout: vi.fn(),
}))

vi.mock('@/api/client', () => ({
  default: {},
  setAccessToken: vi.fn(),
  setOnAuthExpired: vi.fn(),
}))

import * as authApi from '@/api/auth'
import { setOnAuthExpired } from '@/api/client'

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
  })

  it('logout clears state', async () => {
    vi.mocked(authApi.logout).mockResolvedValue(undefined)
    const store = useAuthStore()
    store.setAuth({ user: mockUser, access_token: 'tok' })

    await store.logout()
    expect(store.isAuthenticated).toBe(false)
    expect(store.user).toBeNull()
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

  it('init refreshes session via httpOnly cookie', async () => {
    vi.mocked(authApi.refresh).mockResolvedValue({
      user: mockUser,
      access_token: 'refreshed',
    })
    const store = useAuthStore()
    await store.init()
    expect(authApi.refresh).toHaveBeenCalled()
    expect(store.accessToken).toBe('refreshed')
    expect(store.user).toEqual(mockUser)
    expect(store.initialized).toBe(true)
  })

  it('init clears auth if refresh fails', async () => {
    vi.mocked(authApi.refresh).mockRejectedValue(new Error('expired'))
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
    const store = useAuthStore()
    await store.init()
    vi.mocked(authApi.refresh).mockClear()
    await store.init()
    expect(authApi.refresh).not.toHaveBeenCalled()
  })

  it('clearAuth removes everything', () => {
    const store = useAuthStore()
    store.setAuth({ user: mockUser, access_token: 'tok' })
    store.clearAuth()
    expect(store.user).toBeNull()
    expect(store.accessToken).toBeNull()
  })

  it('registers onAuthExpired callback that clears auth state', () => {
    const store = useAuthStore()
    store.setAuth({ user: mockUser, access_token: 'tok' })

    expect(setOnAuthExpired).toHaveBeenCalledTimes(1)
    expect(setOnAuthExpired).toHaveBeenCalledWith(expect.any(Function))

    // Simulate the callback being invoked (as interceptor would do)
    const callback = vi.mocked(setOnAuthExpired).mock.calls[0][0]
    callback()

    expect(store.isAuthenticated).toBe(false)
    expect(store.user).toBeNull()
  })
})
