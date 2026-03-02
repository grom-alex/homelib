import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as parentalApi from '@/api/parental'

const AUTO_LOCK_TIMEOUT = 15 * 60 * 1000 // 15 minutes
const UNLOCK_TS_KEY = 'parental_unlock_ts'

export const useParentalStore = defineStore('parental', () => {
  const adultContentEnabled = ref(false)
  const pinSet = ref(false)
  const loaded = ref(false)
  let unlockTimer: ReturnType<typeof setTimeout> | null = null

  const isRestricted = computed(() => !adultContentEnabled.value && pinSet.value)

  function clearAutoLockTimer() {
    if (unlockTimer !== null) {
      clearTimeout(unlockTimer)
      unlockTimer = null
    }
  }

  function startAutoLockTimer(delay: number) {
    clearAutoLockTimer()
    unlockTimer = setTimeout(() => {
      unlockTimer = null
      lock()
    }, delay)
  }

  function checkAutoLock() {
    const ts = localStorage.getItem(UNLOCK_TS_KEY)
    if (!ts) return
    const elapsed = Date.now() - Number(ts)
    if (elapsed >= AUTO_LOCK_TIMEOUT) {
      localStorage.removeItem(UNLOCK_TS_KEY)
      lock()
    } else {
      startAutoLockTimer(AUTO_LOCK_TIMEOUT - elapsed)
    }
  }

  async function loadStatus() {
    try {
      const status = await parentalApi.getMyParentalStatus()
      adultContentEnabled.value = status.adult_content_enabled
      pinSet.value = status.pin_set
      loaded.value = true
      if (adultContentEnabled.value && pinSet.value) {
        checkAutoLock()
      }
    } catch {
      loaded.value = true
    }
  }

  async function unlock(pin: string) {
    await parentalApi.unlockAdultContent(pin)
    adultContentEnabled.value = true
    localStorage.setItem(UNLOCK_TS_KEY, String(Date.now()))
    startAutoLockTimer(AUTO_LOCK_TIMEOUT)
  }

  async function lock() {
    clearAutoLockTimer()
    localStorage.removeItem(UNLOCK_TS_KEY)
    try {
      await parentalApi.lockAdultContent()
    } catch {
      // ignore — may fail if session expired
    }
    adultContentEnabled.value = false
  }

  function reset() {
    clearAutoLockTimer()
    localStorage.removeItem(UNLOCK_TS_KEY)
    adultContentEnabled.value = false
    pinSet.value = false
    loaded.value = false
  }

  return {
    adultContentEnabled,
    pinSet,
    loaded,
    isRestricted,
    loadStatus,
    unlock,
    lock,
    reset,
  }
})
