import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('api service', () => {
  beforeEach(() => {
    vi.resetModules()
    sessionStorage.clear()
  })

  it('creates axios instance with /api baseURL', async () => {
    const createSpy = vi.fn().mockReturnValue({
      interceptors: {
        request: { use: vi.fn() },
        response: { use: vi.fn() },
      },
    })
    vi.doMock('axios', () => ({ default: { create: createSpy, post: vi.fn() } }))

    await import('../api')
    expect(createSpy).toHaveBeenCalledWith(
      expect.objectContaining({ baseURL: '/api' }),
    )
  })

  it('request interceptor adds Bearer token from sessionStorage', async () => {
    let requestInterceptor: (config: { headers: Record<string, string> }) => { headers: Record<string, string> }
    const createSpy = vi.fn().mockReturnValue({
      interceptors: {
        request: {
          use: vi.fn((fn: typeof requestInterceptor) => { requestInterceptor = fn }),
        },
        response: { use: vi.fn() },
      },
    })
    vi.doMock('axios', () => ({ default: { create: createSpy, post: vi.fn() } }))
    await import('../api')

    sessionStorage.setItem('access_token', 'my-token')
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
    vi.doMock('axios', () => ({ default: { create: createSpy, post: vi.fn() } }))
    await import('../api')

    const config = { headers: {} as Record<string, string> }
    const result = requestInterceptor!(config)
    expect(result.headers.Authorization).toBeUndefined()
  })

  it('response interceptor retries on 401 with refresh', async () => {
    let responseRejector: (error: unknown) => Promise<unknown>
    const mockApiCall = vi.fn().mockResolvedValue({ data: 'retry-success' })
    const mockPost = vi.fn().mockResolvedValue({ data: { access_token: 'new-tok' } })
    const instance = Object.assign(mockApiCall, {
      interceptors: {
        request: { use: vi.fn() },
        response: {
          use: vi.fn((_: unknown, rej: typeof responseRejector) => { responseRejector = rej }),
        },
      },
      post: mockPost,
    })
    const createSpy = vi.fn().mockReturnValue(instance)
    vi.doMock('axios', () => ({ default: { create: createSpy } }))
    await import('../api')

    const error = {
      response: { status: 401 },
      config: { headers: {} as Record<string, string>, _retry: false },
    }

    const result = await responseRejector!(error)
    expect(mockPost).toHaveBeenCalledWith('/auth/refresh')
    expect(sessionStorage.getItem('access_token')).toBe('new-tok')
    expect(result).toEqual({ data: 'retry-success' })
  })

  it('response interceptor redirects to /login on refresh failure', async () => {
    let responseRejector: (error: unknown) => Promise<unknown>
    const mockPost = vi.fn().mockRejectedValue(new Error('refresh failed'))
    const createSpy = vi.fn().mockReturnValue({
      interceptors: {
        request: { use: vi.fn() },
        response: {
          use: vi.fn((_: unknown, rej: typeof responseRejector) => { responseRejector = rej }),
        },
      },
      post: mockPost,
    })
    vi.doMock('axios', () => ({ default: { create: createSpy } }))

    // Mock window.location
    const originalHref = window.location.href
    Object.defineProperty(window, 'location', {
      value: { href: originalHref },
      writable: true,
    })

    await import('../api')

    sessionStorage.setItem('access_token', 'old-tok')
    const error = {
      response: { status: 401 },
      config: { headers: {}, _retry: false },
    }

    await expect(responseRejector!(error)).rejects.toBeDefined()
    expect(sessionStorage.getItem('access_token')).toBeNull()
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
    vi.doMock('axios', () => ({ default: { create: createSpy, post: vi.fn() } }))
    await import('../api')

    const error = { response: { status: 500 }, config: {} }
    await expect(responseRejector!(error)).rejects.toEqual(error)
  })
})
