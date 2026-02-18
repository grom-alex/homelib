import { onMounted, onUnmounted } from 'vue'
import { useReaderStore } from '@/stores/reader'

interface KeyboardActions {
  nextPage: () => void
  prevPage: () => void
  goToStart: () => void
  goToEnd: () => void
  changeFontSize: (delta: number) => void
  exitReader: () => void
}

export function useReaderKeyboard(actions: KeyboardActions) {
  const store = useReaderStore()

  function handleKeydown(e: KeyboardEvent) {
    // Don't handle when input is focused
    const tag = (e.target as HTMLElement)?.tagName
    if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return

    switch (e.key) {
      case 'ArrowRight':
      case ' ':
      case 'PageDown':
        e.preventDefault()
        actions.nextPage()
        break

      case 'ArrowLeft':
      case 'PageUp':
        e.preventDefault()
        actions.prevPage()
        break

      case 'Home':
        e.preventDefault()
        actions.goToStart()
        break

      case 'End':
        e.preventDefault()
        actions.goToEnd()
        break

      case 't':
      case 'T':
        e.preventDefault()
        store.toggleTOC()
        break

      case 'f':
      case 'F':
        e.preventDefault()
        toggleFullscreen()
        break

      case '+':
      case '=':
        e.preventDefault()
        actions.changeFontSize(1)
        break

      case '-':
        e.preventDefault()
        actions.changeFontSize(-1)
        break

      case 'n':
      case 'N':
        e.preventDefault()
        cycleTheme()
        break

      case 'Escape':
        e.preventDefault()
        if (store.tocVisible) {
          store.tocVisible = false
        } else if (store.settingsVisible) {
          store.settingsVisible = false
        } else {
          actions.exitReader()
        }
        break
    }
  }

  function cycleTheme() {
    const themes = ['light', 'sepia', 'dark', 'night'] as const
    const idx = themes.indexOf(store.settings.theme as typeof themes[number])
    const next = themes[(idx + 1) % themes.length]
    store.updateSettings({ theme: next })
  }

  function toggleFullscreen() {
    if (document.fullscreenElement) {
      document.exitFullscreen().catch(() => {})
    } else {
      document.documentElement.requestFullscreen().catch(() => {})
    }
  }

  onMounted(() => {
    document.addEventListener('keydown', handleKeydown)
  })

  onUnmounted(() => {
    document.removeEventListener('keydown', handleKeydown)
  })
}
