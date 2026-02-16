import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UserInfo, RegisterInput, LoginInput } from '@/services/auth'
import * as authApi from '@/services/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserInfo | null>(null)
  const accessToken = ref<string | null>(null)

  const isAuthenticated = computed(() => !!accessToken.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  function setAuth(data: { user: UserInfo; access_token: string }) {
    user.value = data.user
    accessToken.value = data.access_token
    sessionStorage.setItem('access_token', data.access_token)
  }

  function clearAuth() {
    user.value = null
    accessToken.value = null
    sessionStorage.removeItem('access_token')
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
    const token = sessionStorage.getItem('access_token')
    if (token) {
      accessToken.value = token
      try {
        const data = await authApi.refresh()
        setAuth(data)
      } catch {
        clearAuth()
      }
    }
  }

  return {
    user,
    accessToken,
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
