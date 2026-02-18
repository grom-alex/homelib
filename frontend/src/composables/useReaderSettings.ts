import { watch } from 'vue'
import { useReaderStore } from '@/stores/reader'
import { getUserSettings, updateUserSettings } from '@/api/reader'
import { defaultSettings, type ReaderSettings } from '@/types/reader'

let saveTimer: ReturnType<typeof setTimeout> | null = null
const DEBOUNCE_MS = 1000

export function useReaderSettings() {
  const store = useReaderStore()

  async function loadSettings() {
    try {
      const data = await getUserSettings()
      if (data.reader) {
        const merged = { ...defaultSettings, ...data.reader }
        store.updateSettings(merged)
      }
    } catch {
      // Use defaults on error
    }
  }

  function applySettings(el: HTMLElement) {
    const s = store.settings
    el.style.setProperty('--font-size', `${s.fontSize}px`)
    el.style.setProperty('--font-family', s.fontFamily === 'System' ? 'system-ui, sans-serif' : `"${s.fontFamily}", serif`)
    el.style.setProperty('--font-weight', String(s.fontWeight))
    el.style.setProperty('--line-height', String(s.lineHeight))
    el.style.setProperty('--paragraph-spacing', `${s.paragraphSpacing}em`)
    el.style.setProperty('--letter-spacing', `${s.letterSpacing}em`)
    el.style.setProperty('--margin-h', `${s.marginHorizontal}%`)
    el.style.setProperty('--margin-v', `${s.marginVertical}%`)
    el.style.setProperty('--first-line-indent', `${s.firstLineIndent}em`)
    el.style.setProperty('--text-align', s.textAlign)
    el.style.setProperty('--hyphenation', s.hyphenation ? 'auto' : 'manual')
  }

  function watchSettings(el: HTMLElement) {
    watch(
      () => store.settings,
      () => {
        applySettings(el)
        scheduleSave()
      },
      { deep: true },
    )
    // Apply immediately
    applySettings(el)
  }

  async function saveSettings() {
    try {
      const partial: Partial<ReaderSettings> = { ...store.settings }
      await updateUserSettings({ reader: partial })
    } catch {
      // Ignore save errors
    }
  }

  function scheduleSave() {
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(saveSettings, DEBOUNCE_MS)
  }

  return {
    loadSettings,
    applySettings,
    watchSettings,
    saveSettings,
  }
}
