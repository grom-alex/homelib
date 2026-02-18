import { describe, it, expect, vi, beforeEach } from 'vitest'

// Mock router
vi.mock('@/router', () => ({
  default: {
    push: vi.fn(),
    currentRoute: { value: { fullPath: '/books' } },
  },
}))

describe('api service', () => {
  beforeEach(() => {
    vi.resetModules()
  })

  it('creates axios instance with /api baseURL', async () => {
    const createSpy = vi.fn().mockReturnValue({
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))

    await import('../client')
    expect(createSpy).toHaveBeenCalledWith(
      expect.objectContaining({ baseURL: '/api' }),
    )
  })

  it('request interceptor adds Bearer token from memory', async () => {
    let requestInterceptor: (config: { headers: Record<string, string> }) => { headers: Record<string, string> }
    const createSpy = vi.fn().mockReturnValue({
      interceptors: {
        request: {
          use: vi.fn((fn: typeof requestInterceptor) => { requestInterceptor = fn }),
        },
        response: { use: vi.fn() },
      },
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))
    const mod = await import('../client')

    mod.setAccessToken('my-token')
    const config = { headers: {} as Record<string, string> }
    const result = requestInterceptor!(config)
    expect(result.headers.Authorization).toBe('Bearer my-token')
  })

  it('request interceptor skips header without token', async () => {
    let requestInterceptor: (config: { headers: Record<string, string> }) => { headers: Record<string, string> }
    const createSpy = vi.fn().mockReturnValue({
      interceptors: {
        request: {
          use: vi.fn((fn: typeof requestInterceptor) => { requestInterceptor = fn }),
        },
        response: { use: vi.fn() },
      },
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))
    await import('../client')

    const config = { headers: {} as Record<string, string> }
    const result = requestInterceptor!(config)
    expect(result.headers.Authorization).toBeUndefined()
  })

  it('response interceptor retries on 401 with refresh', async () => {
    let responseRejector: (error: unknown) => Promise<unknown>
    const mockApiCall = vi.fn().mockResolvedValue({ data: 'retry-success' })
    const mockRefreshPost = vi.fn().mockResolvedValue({ data: { access_token: 'new-tok' } })

    const apiInstance = Object.assign(mockApiCall, {
      interceptors: {
        request: { use: vi.fn() },
        response: {
          use: vi.fn((_: unknown, rej: typeof responseRejector) => { responseRejector = rej }),
        },
      },
    })
    const refreshInstance = { post: mockRefreshPost }

    let callCount = 0
    const createSpy = vi.fn(() => {
      callCount++
      return callCount === 1 ? apiInstance : refreshInstance
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))
    await import('../client')

    const error = {
      response: { status: 401 },
      config: { headers: {} as Record<string, string>, _retry: false, url: '/books' },
    }

    const result = await responseRejector!(error)
    expect(mockRefreshPost).toHaveBeenCalledWith('/auth/refresh')
    expect(result).toEqual({ data: 'retry-success' })
  })

  it('response interceptor navigates to /login with redirect query on refresh failure', async () => {
    let responseRejector: (error: unknown) => Promise<unknown>
    const mockRefreshPost = vi.fn().mockRejectedValue(new Error('refresh failed'))

    const apiInstance = {
      interceptors: {
        request: { use: vi.fn() },
        response: {
          use: vi.fn((_: unknown, rej: typeof responseRejector) => { responseRejector = rej }),
        },
      },
    }
    const refreshInstance = { post: mockRefreshPost }

    let callCount = 0
    const createSpy = vi.fn(() => {
      callCount++
      return callCount === 1 ? apiInstance : refreshInstance
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))

    await import('../client')
    const router = (await import('@/router')).default

    const error = {
      response: { status: 401 },
      config: { headers: {}, _retry: false, url: '/books' },
    }

    await expect(responseRejector!(error)).rejects.toBeDefined()
    expect(router.push).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'login', query: expect.objectContaining({ redirect: expect.any(String) }) }),
    )
  })

  it('response interceptor calls onAuthExpired callback before navigating to login', async () => {
    let responseRejector: (error: unknown) => Promise<unknown>
    const mockRefreshPost = vi.fn().mockRejectedValue(new Error('refresh failed'))

    const apiInstance = {
      interceptors: {
        request: { use: vi.fn() },
        response: {
          use: vi.fn((_: unknown, rej: typeof responseRejector) => { responseRejector = rej }),
        },
      },
    }
    const refreshInstance = { post: mockRefreshPost }

    let callCount = 0
    const createSpy = vi.fn(() => {
      callCount++
      return callCount === 1 ? apiInstance : refreshInstance
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))

    const mod = await import('../client')
    const onExpired = vi.fn()
    mod.setOnAuthExpired(onExpired)

    const error = {
      response: { status: 401 },
      config: { headers: {}, _retry: false, url: '/books' },
    }

    await expect(responseRejector!(error)).rejects.toBeDefined()
    expect(onExpired).toHaveBeenCalledTimes(1)
  })

  it('response interceptor passes through non-401 errors', async () => {
    let responseRejector: (error: unknown) => Promise<unknown>
    const createSpy = vi.fn().mockReturnValue({
      interceptors: {
        request: { use: vi.fn() },
        response: {
          use: vi.fn((_: unknown, rej: typeof responseRejector) => { responseRejector = rej }),
        },
      },
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))
    await import('../client')

    const error = { response: { status: 500 }, config: {} }
    await expect(responseRejector!(error)).rejects.toEqual(error)
  })

  it('response interceptor skips refresh for /auth/refresh URL', async () => {
    let responseRejector: (error: unknown) => Promise<unknown>
    const createSpy = vi.fn().mockReturnValue({
      interceptors: {
        request: { use: vi.fn() },
        response: {
          use: vi.fn((_: unknown, rej: typeof responseRejector) => { responseRejector = rej }),
        },
      },
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))
    await import('../client')

    const error = {
      response: { status: 401 },
      config: { headers: {}, _retry: false, url: '/auth/refresh' },
    }
    await expect(responseRejector!(error)).rejects.toEqual(error)
  })
})
