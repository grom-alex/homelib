import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfo, RegisterInput, LoginInput } from '@/api/auth'
import * as authApi from '@/api/auth'
import { setAccessToken } from '@/api/client'

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
    user.value = null
    accessToken.value = null
    setAccessToken(null)
  }

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
    return initPromise
  }

  async function doInit() {
    try {
      const data = await authApi.refresh()
      setAuth(data)
    } catch {
      clearAuth()
    } finally {
      initialized.value = true
      initPromise = null
    }
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
