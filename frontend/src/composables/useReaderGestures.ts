import { onMounted, onUnmounted, type Ref } from 'vue'
import { useReaderStore } from '@/stores/reader'

interface GestureActions {
  nextPage: () => void
  prevPage: () => void
  toggleUI: () => void
}

const SWIPE_THRESHOLD = 50

export function useReaderGestures(
  containerRef: Ref<HTMLElement | null>,
  actions: GestureActions,
) {
  const store = useReaderStore()

  let touchStartX = 0
  let touchStartY = 0

  function handleTouchStart(e: TouchEvent) {
    const touch = e.touches[0]
    touchStartX = touch.clientX
    touchStartY = touch.clientY
  }

  function handleTouchEnd(e: TouchEvent) {
    const touch = e.changedTouches[0]
    const deltaX = touch.clientX - touchStartX
    const deltaY = touch.clientY - touchStartY

    // Check if horizontal swipe (deltaX > deltaY)
    if (Math.abs(deltaX) > Math.abs(deltaY) && Math.abs(deltaX) > SWIPE_THRESHOLD) {
      if (deltaX > 0) {
        // Swipe right → prev page
        actions.prevPage()
      } else {
        // Swipe left → next page
        actions.nextPage()
      }
      return
    }

    // Tap zone detection (only for taps, not swipes)
    if (Math.abs(deltaX) < 10 && Math.abs(deltaY) < 10) {
      handleTap(touch.clientX)
    }
  }

  function handleTap(x: number) {
    const el = containerRef.value
    if (!el) return

    const width = el.clientWidth
    const relativeX = x / width

    const tapZones = store.settings.tapZones

    if (tapZones === 'lrc') {
      // Left 25% = prev, Center 50% = toggleUI, Right 25% = next
      if (relativeX < 0.25) {
        actions.prevPage()
      } else if (relativeX > 0.75) {
        actions.nextPage()
      } else {
        actions.toggleUI()
      }
    } else {
      // 'lr': Left 40% = prev, Right 60% = next
      if (relativeX < 0.4) {
        actions.prevPage()
      } else {
        actions.nextPage()
      }
    }
  }

  onMounted(() => {
    const el = containerRef.value
    if (!el) return
    el.addEventListener('touchstart', handleTouchStart, { passive: true })
    el.addEventListener('touchend', handleTouchEnd, { passive: true })
  })

  onUnmounted(() => {
    const el = containerRef.value
    if (!el) return
    el.removeEventListener('touchstart', handleTouchStart)
    el.removeEventListener('touchend', handleTouchEnd)
  })
}
