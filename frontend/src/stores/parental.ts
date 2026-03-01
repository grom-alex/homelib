import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as parentalApi from '@/api/parental'

export const useParentalStore = defineStore('parental', () => {
  const adultContentEnabled = ref(false)
  const pinSet = ref(false)
  const loaded = ref(false)

  const isRestricted = computed(() => !adultContentEnabled.value && pinSet.value)

  async function loadStatus() {
    try {
      const status = await parentalApi.getMyParentalStatus()
      adultContentEnabled.value = status.adult_content_enabled
      pinSet.value = status.pin_set
      loaded.value = true
    } catch {
      loaded.value = true
    }
  }

  async function unlock(pin: string) {
    await parentalApi.unlockAdultContent(pin)
    adultContentEnabled.value = true
  }

  async function lock() {
    await parentalApi.lockAdultContent()
    adultContentEnabled.value = false
  }

  function reset() {
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
