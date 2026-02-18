import { ref, watch, nextTick, type Ref } from 'vue'
import { useReaderStore } from '@/stores/reader'

export function usePagination(containerRef: Ref<HTMLElement | null>) {
  const store = useReaderStore()
  const translateX = ref(0)

  // Sub-pixel precision to avoid cumulative drift on distant pages
  function getWidth(): number {
    const el = containerRef.value
    if (!el) return 0
    return el.getBoundingClientRect().width
  }

  function applyColumnWidth() {
    const el = containerRef.value
    if (!el) return
    const w = getWidth()
    if (w > 0) {
      el.style.columnWidth = w + 'px'
    }
  }

  function calculateTotalPages() {
    const el = containerRef.value
    if (!el) {
      store.setTotalPages(1)
      return
    }
    applyColumnWidth()
    // Force layout recalc after setting column-width
    void el.offsetHeight
    const scrollW = el.scrollWidth
    const clientW = getWidth()
    if (clientW <= 0) {
      store.setTotalPages(1)
      return
    }
    const pages = Math.max(1, Math.round(scrollW / clientW))
    store.setTotalPages(pages)
  }

  function goToPage(page: number) {
    const el = containerRef.value
    if (!el) return

    const clamped = Math.max(1, Math.min(page, store.totalPages))
    store.setPage(clamped)
    translateX.value = -(clamped - 1) * getWidth()
  }

  function nextPage() {
    if (store.currentPage < store.totalPages) {
      goToPage(store.currentPage + 1)
    }
  }

  function prevPage() {
    if (store.currentPage > 1) {
      goToPage(store.currentPage - 1)
    }
  }

  function recalculate() {
    nextTick(() => {
      // Proportional page preservation
      const prevTotal = store.totalPages
      const ratio = prevTotal > 1
        ? (store.currentPage - 1) / (prevTotal - 1)
        : 0
      calculateTotalPages()
      const newPage = Math.max(1, Math.round(ratio * (store.totalPages - 1)) + 1)
      goToPage(newPage)
    })
  }

  // Watch for settings changes that affect layout
  watch(
    () => [
      store.settings.fontSize,
      store.settings.fontFamily,
      store.settings.lineHeight,
      store.settings.paragraphSpacing,
      store.settings.marginHorizontal,
      store.settings.marginVertical,
      store.settings.firstLineIndent,
    ],
    () => {
      nextTick(() => recalculate())
    },
  )

  // Handle resize
  let resizeObserver: ResizeObserver | null = null

  function setupResizeObserver() {
    if (resizeObserver) {
      resizeObserver.disconnect()
    }
    const el = containerRef.value
    if (!el) return

    resizeObserver = new ResizeObserver(() => {
      recalculate()
    })
    resizeObserver.observe(el)
  }

  function cleanup() {
    if (resizeObserver) {
      resizeObserver.disconnect()
      resizeObserver = null
    }
  }

  return {
    translateX,
    calculateTotalPages,
    goToPage,
    nextPage,
    prevPage,
    recalculate,
    setupResizeObserver,
    cleanup,
  }
}
