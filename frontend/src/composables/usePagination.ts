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

  // Find the first block-level element visible on the given page.
  // Used as an anchor to preserve reading position across layout changes.
  function findAnchorElement(
    container: HTMLElement,
    pageStart: number,
    pageWidth: number,
  ): HTMLElement | null {
    const elements = container.querySelectorAll(
      'p, h1, h2, h3, h4, h5, h6, img, blockquote, hr',
    )
    for (const el of elements) {
      const htmlEl = el as HTMLElement
      if (htmlEl.offsetLeft >= pageStart && htmlEl.offsetLeft < pageStart + pageWidth) {
        return htmlEl
      }
    }
    return null
  }

  function recalculate() {
    const el = containerRef.value
    if (!el) return

    // Anchor-based page preservation: find the element the user is reading
    const width = getWidth()
    const pageStart = (store.currentPage - 1) * width
    const anchor = width > 0 ? findAnchorElement(el, pageStart, width) : null

    nextTick(() => {
      calculateTotalPages()

      if (anchor) {
        const newWidth = getWidth()
        if (newWidth > 0) {
          const newPage = Math.floor(anchor.offsetLeft / newWidth) + 1
          goToPage(Math.max(1, Math.min(newPage, store.totalPages)))
          return
        }
      }

      // Fallback: stay on current page (clamped to new total)
      goToPage(Math.max(1, Math.min(store.currentPage, store.totalPages)))
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
