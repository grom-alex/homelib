import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest'
import { ref } from 'vue'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('vue', async () => {
  const actual = await vi.importActual<typeof import('vue')>('vue')
  return {
    ...actual,
    onMounted: (fn: () => void) => fn(),
    onUnmounted: vi.fn(),
  }
})

vi.mock('@/stores/reader', () => ({
  useReaderStore: vi.fn(),
}))

import { useReaderGestures } from '../useReaderGestures'
import { useReaderStore } from '@/stores/reader'

const mockedUseReaderStore = vi.mocked(useReaderStore)

function createMockElement(): HTMLElement {
  const listeners: Record<string, EventListenerOrEventListenerObject[]> = {}
  return {
    getBoundingClientRect: () => ({ left: 100, top: 0, width: 800, height: 600, right: 900, bottom: 600 }),
    clientWidth: 800,
    addEventListener: vi.fn((type: string, handler: EventListenerOrEventListenerObject, _opts?: unknown) => {
      if (!listeners[type]) listeners[type] = []
      listeners[type].push(handler)
    }),
    removeEventListener: vi.fn(),
    __listeners: listeners,
  } as unknown as HTMLElement
}

function dispatchEvent(el: HTMLElement, type: string, event: unknown) {
  const listeners = (el as unknown as { __listeners: Record<string, EventListenerOrEventListenerObject[]> }).__listeners
  const handlers = listeners[type] || []
  for (const h of handlers) {
    if (typeof h === 'function') h(event as Event)
    else h.handleEvent(event as Event)
  }
}

describe('useReaderGestures', () => {
  let actions: { nextPage: ReturnType<typeof vi.fn>; prevPage: ReturnType<typeof vi.fn>; toggleUI: ReturnType<typeof vi.fn> }
  let el: HTMLElement
  let containerRef: ReturnType<typeof ref<HTMLElement | null>>

  beforeEach(() => {
    setActivePinia(createPinia())
    actions = {
      nextPage: vi.fn(),
      prevPage: vi.fn(),
      toggleUI: vi.fn(),
    }
    el = createMockElement()
    containerRef = ref(el) as ReturnType<typeof ref<HTMLElement | null>>
    mockedUseReaderStore.mockReturnValue({
      settings: { tapZones: 'lrc' },
    } as ReturnType<typeof useReaderStore>)
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  it('registers all event listeners on mount', () => {
    useReaderGestures(containerRef, actions)
    expect(el.addEventListener).toHaveBeenCalledWith('touchstart', expect.any(Function), { passive: true })
    expect(el.addEventListener).toHaveBeenCalledWith('touchend', expect.any(Function), { passive: true })
    expect(el.addEventListener).toHaveBeenCalledWith('click', expect.any(Function))
    expect(el.addEventListener).toHaveBeenCalledWith('wheel', expect.any(Function), { passive: false })
    expect(el.addEventListener).toHaveBeenCalledWith('mouseup', expect.any(Function))
  })

  describe('tap zones (lrc mode)', () => {
    it('click on left 25% triggers prevPage', () => {
      useReaderGestures(containerRef, actions)
      // left edge: clientX=100 (el starts at 100), relativeX = 0/800 = 0
      dispatchEvent(el, 'click', { clientX: 100, target: { closest: () => null } })
      expect(actions.prevPage).toHaveBeenCalled()
    })

    it('click on right 25% triggers nextPage', () => {
      useReaderGestures(containerRef, actions)
      // right edge: clientX=850, relativeX = 750/800 = 0.9375
      dispatchEvent(el, 'click', { clientX: 850, target: { closest: () => null } })
      expect(actions.nextPage).toHaveBeenCalled()
    })

    it('click on center 50% triggers toggleUI', () => {
      useReaderGestures(containerRef, actions)
      // center: clientX=500, relativeX = 400/800 = 0.5
      dispatchEvent(el, 'click', { clientX: 500, target: { closest: () => null } })
      expect(actions.toggleUI).toHaveBeenCalled()
    })
  })

  describe('tap zones (lr mode)', () => {
    beforeEach(() => {
      mockedUseReaderStore.mockReturnValue({
        settings: { tapZones: 'lr' },
      } as ReturnType<typeof useReaderStore>)
    })

    it('click on left 40% triggers prevPage', () => {
      useReaderGestures(containerRef, actions)
      // clientX=300, relativeX = 200/800 = 0.25
      dispatchEvent(el, 'click', { clientX: 300, target: { closest: () => null } })
      expect(actions.prevPage).toHaveBeenCalled()
    })

    it('click on right 60% triggers nextPage', () => {
      useReaderGestures(containerRef, actions)
      // clientX=600, relativeX = 500/800 = 0.625
      dispatchEvent(el, 'click', { clientX: 600, target: { closest: () => null } })
      expect(actions.nextPage).toHaveBeenCalled()
    })
  })

  describe('click skips interactive elements', () => {
    it('does not navigate when clicking a link', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'click', { clientX: 100, target: { closest: (sel: string) => sel.includes('a') ? {} : null } })
      expect(actions.prevPage).not.toHaveBeenCalled()
      expect(actions.nextPage).not.toHaveBeenCalled()
      expect(actions.toggleUI).not.toHaveBeenCalled()
    })

    it('does not navigate when clicking footnote-ref', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'click', { clientX: 850, target: { closest: (sel: string) => sel.includes('footnote-ref') ? {} : null } })
      expect(actions.nextPage).not.toHaveBeenCalled()
    })
  })

  describe('touch swipe', () => {
    it('swipe left triggers nextPage', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'touchstart', { touches: [{ clientX: 400, clientY: 300 }] })
      dispatchEvent(el, 'touchend', { changedTouches: [{ clientX: 300, clientY: 300 }] })
      expect(actions.nextPage).toHaveBeenCalled()
    })

    it('swipe right triggers prevPage', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'touchstart', { touches: [{ clientX: 300, clientY: 300 }] })
      dispatchEvent(el, 'touchend', { changedTouches: [{ clientX: 400, clientY: 300 }] })
      expect(actions.prevPage).toHaveBeenCalled()
    })

    it('small movement triggers tap zone instead of swipe', () => {
      useReaderGestures(containerRef, actions)
      // Small movement at center area
      dispatchEvent(el, 'touchstart', { touches: [{ clientX: 500, clientY: 300 }] })
      dispatchEvent(el, 'touchend', { changedTouches: [{ clientX: 503, clientY: 302 }] })
      expect(actions.toggleUI).toHaveBeenCalled()
    })

    it('vertical swipe is ignored', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'touchstart', { touches: [{ clientX: 400, clientY: 200 }] })
      dispatchEvent(el, 'touchend', { changedTouches: [{ clientX: 410, clientY: 400 }] })
      expect(actions.nextPage).not.toHaveBeenCalled()
      expect(actions.prevPage).not.toHaveBeenCalled()
    })
  })

  describe('mouse wheel', () => {
    it('wheel down triggers nextPage', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'wheel', { deltaY: 100, deltaX: 0, preventDefault: vi.fn() })
      expect(actions.nextPage).toHaveBeenCalled()
    })

    it('wheel up triggers prevPage', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'wheel', { deltaY: -100, deltaX: 0, preventDefault: vi.fn() })
      expect(actions.prevPage).toHaveBeenCalled()
    })

    it('horizontal wheel right triggers nextPage', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'wheel', { deltaY: 0, deltaX: 50, preventDefault: vi.fn() })
      expect(actions.nextPage).toHaveBeenCalled()
    })

    it('wheel events are throttled by cooldown', () => {
      vi.useFakeTimers()
      useReaderGestures(containerRef, actions)

      dispatchEvent(el, 'wheel', { deltaY: 100, deltaX: 0, preventDefault: vi.fn() })
      expect(actions.nextPage).toHaveBeenCalledTimes(1)

      // Fire again within cooldown — should be ignored
      vi.advanceTimersByTime(100)
      dispatchEvent(el, 'wheel', { deltaY: 100, deltaX: 0, preventDefault: vi.fn() })
      expect(actions.nextPage).toHaveBeenCalledTimes(1)

      // Fire after cooldown — should work
      vi.advanceTimersByTime(300)
      dispatchEvent(el, 'wheel', { deltaY: 100, deltaX: 0, preventDefault: vi.fn() })
      expect(actions.nextPage).toHaveBeenCalledTimes(2)

      vi.useRealTimers()
    })
  })

  describe('mouse back/forward buttons', () => {
    it('mouse back button (3) triggers prevPage', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'mouseup', { button: 3, preventDefault: vi.fn() })
      expect(actions.prevPage).toHaveBeenCalled()
    })

    it('mouse forward button (4) triggers nextPage', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'mouseup', { button: 4, preventDefault: vi.fn() })
      expect(actions.nextPage).toHaveBeenCalled()
    })

    it('regular mouse buttons do not trigger navigation', () => {
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'mouseup', { button: 0, preventDefault: vi.fn() })
      dispatchEvent(el, 'mouseup', { button: 1, preventDefault: vi.fn() })
      dispatchEvent(el, 'mouseup', { button: 2, preventDefault: vi.fn() })
      expect(actions.prevPage).not.toHaveBeenCalled()
      expect(actions.nextPage).not.toHaveBeenCalled()
    })
  })

  describe('relativeX calculation accounts for element offset', () => {
    it('uses getBoundingClientRect.left for correct position', () => {
      // Element starts at left:100, width:800
      // Click at clientX=100 should be relativeX=0 (leftmost edge)
      useReaderGestures(containerRef, actions)
      dispatchEvent(el, 'click', { clientX: 100, target: { closest: () => null } })
      // relativeX = (100-100)/800 = 0, which is < 0.25, so prevPage
      expect(actions.prevPage).toHaveBeenCalled()
    })

    it('correctly identifies right zone with offset', () => {
      useReaderGestures(containerRef, actions)
      // Click at clientX=880, relativeX = (880-100)/800 = 0.975
      dispatchEvent(el, 'click', { clientX: 880, target: { closest: () => null } })
      expect(actions.nextPage).toHaveBeenCalled()
    })
  })

  it('does nothing when containerRef is null', () => {
    const nullRef = ref(null) as ReturnType<typeof ref<HTMLElement | null>>
    // Should not throw
    useReaderGestures(nullRef, actions)
    expect(actions.nextPage).not.toHaveBeenCalled()
  })
})
