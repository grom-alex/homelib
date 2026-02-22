import { onMounted, onUnmounted, type Ref } from 'vue'
import { useReaderStore } from '@/stores/reader'

interface GestureActions {
  nextPage: () => void
  prevPage: () => void
  toggleUI: () => void
}

const SWIPE_THRESHOLD = 50
const WHEEL_COOLDOWN_MS = 300

export function useReaderGestures(
  containerRef: Ref<HTMLElement | null>,
  actions: GestureActions,
) {
  const store = useReaderStore()

  let touchStartX = 0
  let touchStartY = 0
  let lastWheelTime = 0

  function getRelativeX(clientX: number): number {
    const el = containerRef.value
    if (!el) return 0.5
    const rect = el.getBoundingClientRect()
    return (clientX - rect.left) / rect.width
  }

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
      navigateByTapZone(getRelativeX(touch.clientX))
    }
  }

  function handleClick(e: MouseEvent) {
    // Skip if it was a footnote or interactive element click
    const target = e.target as HTMLElement
    if (target.closest('a, button, .footnote-ref, .footnote-popup')) return

    navigateByTapZone(getRelativeX(e.clientX))
  }

  function navigateByTapZone(relativeX: number) {
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

  function handleWheel(e: WheelEvent) {
    e.preventDefault()

    const now = Date.now()
    if (now - lastWheelTime < WHEEL_COOLDOWN_MS) return
    lastWheelTime = now

    if (e.deltaY > 0 || e.deltaX > 0) {
      actions.nextPage()
    } else if (e.deltaY < 0 || e.deltaX < 0) {
      actions.prevPage()
    }
  }

  function handleMouseUp(e: MouseEvent) {
    // Mouse back button (button 3) → prev page
    if (e.button === 3) {
      e.preventDefault()
      actions.prevPage()
    }
    // Mouse forward button (button 4) → next page
    if (e.button === 4) {
      e.preventDefault()
      actions.nextPage()
    }
  }

  onMounted(() => {
    const el = containerRef.value
    if (!el) return
    el.addEventListener('touchstart', handleTouchStart, { passive: true })
    el.addEventListener('touchend', handleTouchEnd, { passive: true })
    el.addEventListener('click', handleClick)
    el.addEventListener('wheel', handleWheel, { passive: false })
    el.addEventListener('mouseup', handleMouseUp)
  })

  onUnmounted(() => {
    const el = containerRef.value
    if (!el) return
    el.removeEventListener('touchstart', handleTouchStart)
    el.removeEventListener('touchend', handleTouchEnd)
    el.removeEventListener('click', handleClick)
    el.removeEventListener('wheel', handleWheel)
    el.removeEventListener('mouseup', handleMouseUp)
  })
}
