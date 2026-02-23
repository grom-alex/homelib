import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useTheme } from 'vuetify'
import type { CatalogThemeName, CustomCatalogColors } from '@/types/catalog'
import { defaultCatalogSettings } from '@/types/catalog'
import api from '@/api/client'

/** Определяем яркость цвета (0-255) для auto-detect dark/light */
function luminance(hex: string): number {
  const c = hex.replace('#', '')
  const r = parseInt(c.substring(0, 2), 16)
  const g = parseInt(c.substring(2, 4), 16)
  const b = parseInt(c.substring(4, 6), 16)
  return 0.299 * r + 0.587 * g + 0.114 * b
}

/** Осветляет/затемняет hex-цвет на amount (-255..+255) */
function adjustColor(hex: string, amount: number): string {
  const c = hex.replace('#', '')
  const r = Math.max(0, Math.min(255, parseInt(c.substring(0, 2), 16) + amount))
  const g = Math.max(0, Math.min(255, parseInt(c.substring(2, 4), 16) + amount))
  const b = Math.max(0, Math.min(255, parseInt(c.substring(4, 6), 16) + amount))
  return `#${Math.round(r).toString(16).padStart(2, '0')}${Math.round(g).toString(16).padStart(2, '0')}${Math.round(b).toString(16).padStart(2, '0')}`
}

/** Строим полную палитру Vuetify из 4 пользовательских цветов */
function buildCustomPalette(colors: CustomCatalogColors): Record<string, string> {
  const isDark = luminance(colors.background) < 128
  const shift = isDark ? 12 : -8
  const shiftMore = isDark ? 20 : -15

  return {
    background: colors.background,
    surface: adjustColor(colors.background, shift),
    'surface-variant': adjustColor(colors.background, shiftMore),
    primary: colors.link,
    secondary: colors.text,
    accent: colors.link,
    error: isDark ? '#EF5350' : '#D32F2F',
    success: isDark ? '#66BB6A' : '#388E3C',
    warning: isDark ? '#FFA726' : '#F57C00',
    info: colors.link,
    'on-background': colors.text,
    'on-surface': colors.text,
    'table-row-hover': adjustColor(colors.background, shiftMore),
    'table-row-selected': colors.selection,
    'nav-item-active': colors.selection,
    'status-bar': adjustColor(colors.background, shiftMore),
  }
}

const defaultCustomColors: CustomCatalogColors = {
  background: '#FFFFFF',
  text: '#212121',
  link: '#1565C0',
  selection: '#BBDEFB',
}

export const useThemeStore = defineStore('theme', () => {
  const catalogTheme = ref<CatalogThemeName>(defaultCatalogSettings.theme)
  const readerThemeOverride = ref<CatalogThemeName | null>(null)
  const customCatalogColors = ref<CustomCatalogColors>({ ...defaultCustomColors })
  const loaded = ref(false)

  // Захватываем ссылку на Vuetify theme в setup-контексте стора,
  // где inject() ещё доступен (стор создаётся внутри компонента).
  let vuetifyTheme: ReturnType<typeof useTheme> | null = null
  try {
    vuetifyTheme = useTheme()
  } catch {
    // useTheme() не доступен вне setup-контекста (тесты)
  }

  const effectiveReaderTheme = computed<CatalogThemeName>(
    () => readerThemeOverride.value ?? catalogTheme.value,
  )

  function applyCustomColorsToVuetify(colors: CustomCatalogColors) {
    if (!vuetifyTheme) return
    const palette = buildCustomPalette(colors)
    const isDark = luminance(colors.background) < 128
    vuetifyTheme.themes.value.custom.dark = isDark
    for (const [key, value] of Object.entries(palette)) {
      vuetifyTheme.themes.value.custom.colors[key] = value
    }
  }

  function setCatalogTheme(theme: CatalogThemeName) {
    catalogTheme.value = theme
    if (theme === 'custom') {
      applyCustomColorsToVuetify(customCatalogColors.value)
    }
    applyVuetifyTheme(theme)
    scheduleSave()
  }

  function setCatalogCustomColors(colors: CustomCatalogColors) {
    customCatalogColors.value = { ...colors }
    applyCustomColorsToVuetify(colors)
    if (catalogTheme.value === 'custom') {
      applyVuetifyTheme('custom')
    }
    scheduleSave()
  }

  function setReaderTheme(theme: CatalogThemeName) {
    readerThemeOverride.value = theme
    scheduleSave()
  }

  function resetReaderTheme() {
    readerThemeOverride.value = null
    scheduleSave()
  }

  function applyVuetifyTheme(theme: CatalogThemeName) {
    if (vuetifyTheme) {
      vuetifyTheme.global.name.value = theme
    }
  }

  async function loadSettings() {
    try {
      const { data } = await api.get<Record<string, unknown>>('/me/settings')

      const catalog = data.catalog as Record<string, unknown> | undefined
      if (catalog?.theme) {
        catalogTheme.value = catalog.theme as CatalogThemeName
      }
      if (catalog?.customColors) {
        customCatalogColors.value = catalog.customColors as CustomCatalogColors
      }
      if (catalogTheme.value === 'custom') {
        applyCustomColorsToVuetify(customCatalogColors.value)
      }
      applyVuetifyTheme(catalogTheme.value)

      const reader = data.reader as Record<string, unknown> | undefined
      if (reader) {
        readerThemeOverride.value = (reader.theme as CatalogThemeName | null) ?? null
      }

      loaded.value = true
    } catch {
      // Используем значения по умолчанию
      loaded.value = true
    }
  }

  let saveTimer: ReturnType<typeof setTimeout> | null = null

  function scheduleSave() {
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => saveSettings(), 1000)
  }

  async function saveSettings() {
    try {
      await api.put('/me/settings', {
        catalog: {
          theme: catalogTheme.value,
          customColors: customCatalogColors.value,
        },
        reader: { theme: readerThemeOverride.value },
      })
    } catch {
      // Ошибка сохранения — тема уже применена локально
    }
  }

  return {
    catalogTheme,
    readerThemeOverride,
    customCatalogColors,
    effectiveReaderTheme,
    loaded,
    setCatalogTheme,
    setCatalogCustomColors,
    setReaderTheme,
    resetReaderTheme,
    loadSettings,
    saveSettings,
  }
})
