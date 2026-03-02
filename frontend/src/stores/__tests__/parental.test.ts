import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useParentalStore } from '../parental'

vi.mock('@/api/parental', () => ({
  getMyParentalStatus: vi.fn(),
  unlockAdultContent: vi.fn(),
  lockAdultContent: vi.fn(),
}))

import * as parentalApi from '@/api/parental'

describe('parental store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    vi.useFakeTimers()
    localStorage.clear()
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('has correct initial state', () => {
    const store = useParentalStore()
    expect(store.adultContentEnabled).toBe(false)
    expect(store.pinSet).toBe(false)
    expect(store.loaded).toBe(false)
    expect(store.isRestricted).toBe(false)
  })

  it('isRestricted is true when content disabled and pin set', () => {
    const store = useParentalStore()
    store.pinSet = true
    store.adultContentEnabled = false
    expect(store.isRestricted).toBe(true)
  })

  it('isRestricted is false when content enabled', () => {
    const store = useParentalStore()
    store.pinSet = true
    store.adultContentEnabled = true
    expect(store.isRestricted).toBe(false)
  })

  describe('loadStatus', () => {
    it('loads status from API', async () => {
      vi.mocked(parentalApi.getMyParentalStatus).mockResolvedValue({
        adult_content_enabled: true,
        pin_set: true,
      })

      const store = useParentalStore()
      await store.loadStatus()

      expect(store.adultContentEnabled).toBe(true)
      expect(store.pinSet).toBe(true)
      expect(store.loaded).toBe(true)
    })

    it('sets loaded on API error', async () => {
      vi.mocked(parentalApi.getMyParentalStatus).mockRejectedValue(new Error('Network'))

      const store = useParentalStore()
      await store.loadStatus()

      expect(store.loaded).toBe(true)
      expect(store.adultContentEnabled).toBe(false)
    })

    it('checks auto-lock when content enabled and pin set', async () => {
      const now = Date.now()
      // Simulate unlock 5 minutes ago
      localStorage.setItem('parental_unlock_ts', String(now - 5 * 60 * 1000))

      vi.mocked(parentalApi.getMyParentalStatus).mockResolvedValue({
        adult_content_enabled: true,
        pin_set: true,
      })

      const store = useParentalStore()
      await store.loadStatus()

      // Should still be enabled (only 5 min passed, threshold is 15)
      expect(store.adultContentEnabled).toBe(true)
    })

    it('auto-locks when timestamp is expired', async () => {
      const now = Date.now()
      // Simulate unlock 20 minutes ago (over 15 min threshold)
      localStorage.setItem('parental_unlock_ts', String(now - 20 * 60 * 1000))

      vi.mocked(parentalApi.getMyParentalStatus).mockResolvedValue({
        adult_content_enabled: true,
        pin_set: true,
      })
      vi.mocked(parentalApi.lockAdultContent).mockResolvedValue()

      const store = useParentalStore()
      await store.loadStatus()

      // Should auto-lock immediately
      expect(store.adultContentEnabled).toBe(false)
      expect(localStorage.getItem('parental_unlock_ts')).toBeNull()
    })

    it('does not check auto-lock when no timestamp in localStorage', async () => {
      vi.mocked(parentalApi.getMyParentalStatus).mockResolvedValue({
        adult_content_enabled: true,
        pin_set: true,
      })

      const store = useParentalStore()
      await store.loadStatus()

      // Enabled via admin panel (no timestamp), should stay enabled
      expect(store.adultContentEnabled).toBe(true)
    })
  })

  describe('unlock', () => {
    it('calls API and sets state', async () => {
      vi.mocked(parentalApi.unlockAdultContent).mockResolvedValue()

      const store = useParentalStore()
      await store.unlock('1234')

      expect(parentalApi.unlockAdultContent).toHaveBeenCalledWith('1234')
      expect(store.adultContentEnabled).toBe(true)
      expect(localStorage.getItem('parental_unlock_ts')).toBeTruthy()
    })

    it('starts auto-lock timer after unlock', async () => {
      vi.mocked(parentalApi.unlockAdultContent).mockResolvedValue()
      vi.mocked(parentalApi.lockAdultContent).mockResolvedValue()

      const store = useParentalStore()
      await store.unlock('1234')

      expect(store.adultContentEnabled).toBe(true)

      // Fast-forward 15 minutes and flush async lock()
      vi.advanceTimersByTime(15 * 60 * 1000)
      await vi.runAllTimersAsync()

      expect(store.adultContentEnabled).toBe(false)
    })
  })

  describe('lock', () => {
    it('calls API and resets state', async () => {
      vi.mocked(parentalApi.lockAdultContent).mockResolvedValue()

      const store = useParentalStore()
      store.adultContentEnabled = true
      localStorage.setItem('parental_unlock_ts', String(Date.now()))

      await store.lock()

      expect(parentalApi.lockAdultContent).toHaveBeenCalled()
      expect(store.adultContentEnabled).toBe(false)
      expect(localStorage.getItem('parental_unlock_ts')).toBeNull()
    })

    it('still resets state if API fails', async () => {
      vi.mocked(parentalApi.lockAdultContent).mockRejectedValue(new Error('expired'))

      const store = useParentalStore()
      store.adultContentEnabled = true

      await store.lock()

      expect(store.adultContentEnabled).toBe(false)
    })
  })

  describe('reset', () => {
    it('resets all state and clears localStorage', async () => {
      vi.mocked(parentalApi.unlockAdultContent).mockResolvedValue()

      const store = useParentalStore()
      store.adultContentEnabled = true
      store.pinSet = true
      store.loaded = true
      localStorage.setItem('parental_unlock_ts', String(Date.now()))

      store.reset()

      expect(store.adultContentEnabled).toBe(false)
      expect(store.pinSet).toBe(false)
      expect(store.loaded).toBe(false)
      expect(localStorage.getItem('parental_unlock_ts')).toBeNull()
    })
  })

  describe('auto-lock timer on loadStatus', () => {
    it('starts timer for remaining time when partially elapsed', async () => {
      const now = Date.now()
      // Unlocked 10 minutes ago, 5 minutes remaining
      localStorage.setItem('parental_unlock_ts', String(now - 10 * 60 * 1000))

      vi.mocked(parentalApi.getMyParentalStatus).mockResolvedValue({
        adult_content_enabled: true,
        pin_set: true,
      })
      vi.mocked(parentalApi.lockAdultContent).mockResolvedValue()

      const store = useParentalStore()
      await store.loadStatus()

      expect(store.adultContentEnabled).toBe(true)

      // Advance 5 minutes — should auto-lock
      vi.advanceTimersByTime(5 * 60 * 1000)
      await vi.runAllTimersAsync()

      expect(store.adultContentEnabled).toBe(false)
    })
  })
})
