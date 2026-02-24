import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfo, RegisterInput, LoginInput } from '@/api/auth'
import * as authApi from '@/api/auth'
import { setAccessToken, setOnAuthExpired, setAuthInitPromise } from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const accessToken = ref<string | null>(null)
  const initialized = ref(false)
  let initPromise: Promise<void> | null = null

  const isAuthenticated = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  function setAuth(data: { user: UserInfo; access_token: string }) {
    user.value = data.user
    accessToken.value = data.access_token
    setAccessToken(data.access_token)
  }

  function clearAuth() {
    console.warn('[AUTH] clearAuth called, stack:', new Error().stack?.split('\n').slice(1, 4).join(' <- '))
    user.value = null
    accessToken.value = null
    setAccessToken(null)
  }

  // Sync store state when interceptor detects expired session
  setOnAuthExpired(() => clearAuth())

  async function login(input: LoginInput) {
    const data = await authApi.login(input)
    setAuth(data)
    return data
  }

  async function register(input: RegisterInput) {
    const data = await authApi.register(input)
    setAuth(data)
    return data
  }

  async function refreshToken() {
    try {
      const data = await authApi.refresh()
      setAuth(data)
      return data
    } catch {
      clearAuth()
      throw new Error('Session expired')
    }
  }

  async function logout() {
    try {
      await authApi.logout()
    } finally {
      clearAuth()
    }
  }

  async function init() {
    if (initialized.value) return
    if (initPromise) return initPromise
    initPromise = doInit()
    setAuthInitPromise(initPromise)
    return initPromise
  }

  async function doInit() {
    console.warn('[AUTH] doInit: starting refresh')
    try {
      const data = await refreshWithRetry()
      console.warn('[AUTH] doInit: refresh OK, user:', data.user?.email, 'token length:', data.access_token?.length)
      setAuth(data)
    } catch (e) {
      console.warn('[AUTH] doInit: refresh FAILED:', e)
      clearAuth()
    } finally {
      initialized.value = true
      initPromise = null
      setAuthInitPromise(null)
      console.warn('[AUTH] doInit: done, isAuthenticated:', !!accessToken.value)
    }
  }

  /** Retry refresh on transient errors (network, 5xx); give up immediately on 401/403. */
  async function refreshWithRetry(): Promise<{ user: UserInfo; access_token: string }> {
    const maxAttempts = 3
    let lastError: unknown
    for (let attempt = 1; attempt <= maxAttempts; attempt++) {
      try {
        console.warn(`[AUTH] refreshWithRetry: attempt ${attempt}/${maxAttempts}`)
        return await authApi.refresh()
      } catch (error: unknown) {
        lastError = error
        const status = (error as { response?: { status?: number } })?.response?.status
        console.warn(`[AUTH] refreshWithRetry: attempt ${attempt} failed, status:`, status, 'error:', error)
        if (isAuthError(error)) throw error
        if (attempt < maxAttempts) {
          await new Promise(resolve => setTimeout(resolve, 500 * attempt))
        }
      }
    }
    throw lastError
  }

  function isAuthError(error: unknown): boolean {
    if (error && typeof error === 'object' && 'response' in error) {
      const resp = (error as { response?: { status?: number } }).response
      return resp?.status === 401 || resp?.status === 403
    }
    return false
  }

  return {
    user,
    accessToken,
    initialized,
    isAuthenticated,
    isAdmin,
    login,
    register,
    refreshToken,
    logout,
    init,
    setAuth,
    clearAuth,
  }
})
