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
  setAuthInitPromise: vi.fn(),
}))

import * as authApi from '@/api/auth'
import { setOnAuthExpired, setAuthInitPromise } from '@/api/client'

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

  it('init clears auth if refresh fails with 401', async () => {
    const axiosError = Object.assign(new Error('Unauthorized'), {
      response: { status: 401 },
    })
    vi.mocked(authApi.refresh).mockRejectedValue(axiosError)
    const store = useAuthStore()
    await store.init()
    expect(authApi.refresh).toHaveBeenCalledTimes(1)
    expect(store.accessToken).toBeNull()
    expect(store.user).toBeNull()
    expect(store.initialized).toBe(true)
  })

  it('init retries on network error then succeeds', async () => {
    vi.useFakeTimers()
    vi.mocked(authApi.refresh)
      .mockRejectedValueOnce(new Error('Network Error'))
      .mockResolvedValueOnce({ user: mockUser, access_token: 'tok' })

    const store = useAuthStore()
    const p = store.init()
    await vi.advanceTimersByTimeAsync(2000)
    await p

    expect(authApi.refresh).toHaveBeenCalledTimes(2)
    expect(store.isAuthenticated).toBe(true)
    expect(store.accessToken).toBe('tok')
    expect(store.initialized).toBe(true)
    vi.useRealTimers()
  })

  it('init gives up after max retries on network errors', async () => {
    vi.useFakeTimers()
    vi.mocked(authApi.refresh).mockRejectedValue(new Error('Network Error'))

    const store = useAuthStore()
    const p = store.init()
    await vi.advanceTimersByTimeAsync(5000)
    await p

    expect(authApi.refresh).toHaveBeenCalledTimes(3)
    expect(store.isAuthenticated).toBe(false)
    expect(store.initialized).toBe(true)
    vi.useRealTimers()
  })

  it('init does not retry on 403 auth error', async () => {
    const axiosError = Object.assign(new Error('Forbidden'), {
      response: { status: 403 },
    })
    vi.mocked(authApi.refresh).mockRejectedValue(axiosError)
    const store = useAuthStore()
    await store.init()
    expect(authApi.refresh).toHaveBeenCalledTimes(1)
    expect(store.isAuthenticated).toBe(false)
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

  it('init sets and clears authInitPromise on client module', async () => {
    vi.mocked(authApi.refresh).mockResolvedValue({
      user: mockUser,
      access_token: 'tok',
    })
    const store = useAuthStore()
    vi.mocked(setAuthInitPromise).mockClear()

    await store.init()

    // Should have been called with a promise, then with null
    expect(setAuthInitPromise).toHaveBeenCalledTimes(2)
    expect(setAuthInitPromise).toHaveBeenNthCalledWith(1, expect.any(Promise))
    expect(setAuthInitPromise).toHaveBeenNthCalledWith(2, null)
    expect(store.initialized).toBe(true)
  })
})
