import axios from 'axios'
import router from '@/router'

// Access token stored in memory only (not sessionStorage/localStorage)
let accessToken: string | null = null
let onAuthExpired: (() => void) | null = null
let authInitPromise: Promise<void> | null = null

export function setAuthInitPromise(promise: Promise<void> | null) {
  authInitPromise = promise
}

export function setAccessToken(token: string | null) {
  accessToken = token
}

export function getAccessToken(): string | null {
  return accessToken
}

export function setOnAuthExpired(callback: () => void) {
  onAuthExpired = callback
}

const api = axios.create({
  baseURL: '/api',
  headers: {
    'Content-Type': 'application/json',
  },
  withCredentials: true,
})

api.interceptors.request.use((config) => {
  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }
  return config
})

let isRefreshing = false
let refreshSubscribers: Array<{
  resolve: (token: string) => void
  reject: (error: unknown) => void
}> = []

function onRefreshed(token: string) {
  refreshSubscribers.forEach((sub) => sub.resolve(token))
  refreshSubscribers = []
}

function onRefreshFailed(error: unknown) {
  refreshSubscribers.forEach((sub) => sub.reject(error))
  refreshSubscribers = []
}

// Separate instance for refresh to avoid interceptor loop
export const refreshClient = axios.create({
  baseURL: '/api',
  withCredentials: true,
})

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config
    if (
      error.response?.status === 401 &&
      !originalRequest._retry &&
      !originalRequest.url?.includes('/auth/refresh')
    ) {
      console.warn('[INTERCEPTOR] 401 for:', originalRequest.url, 'hasInitPromise:', !!authInitPromise, 'isRefreshing:', isRefreshing)

      // If auth is initializing, wait for it instead of triggering a parallel refresh
      // that would race with token rotation
      if (authInitPromise) {
        console.warn('[INTERCEPTOR] waiting for auth init to complete')
        try {
          await authInitPromise
        } catch { /* init handles its own errors */ }
        console.warn('[INTERCEPTOR] auth init done, hasToken:', !!accessToken)
        if (accessToken) {
          originalRequest.headers.Authorization = `Bearer ${accessToken}`
          return api(originalRequest)
        }
        return Promise.reject(error)
      }

      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          refreshSubscribers.push({
            resolve: (token: string) => {
              originalRequest.headers.Authorization = `Bearer ${token}`
              resolve(api(originalRequest))
            },
            reject,
          })
        })
      }

      originalRequest._retry = true
      isRefreshing = true

      try {
        const { data } = await refreshClient.post('/auth/refresh')
        accessToken = data.access_token
        originalRequest.headers.Authorization = `Bearer ${data.access_token}`
        onRefreshed(data.access_token)
        return api(originalRequest)
      } catch (refreshError) {
        console.warn('[INTERCEPTOR] refresh FAILED, redirecting to login. Error:', refreshError)
        accessToken = null
        if (onAuthExpired) onAuthExpired()
        onRefreshFailed(refreshError)
        router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
        return Promise.reject(error)
      } finally {
        isRefreshing = false
      }
    }
    return Promise.reject(error)
  },
)

export default api
