import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useTheme } from 'vuetify'
import type { CatalogThemeName } from '@/types/catalog'
import { defaultCatalogSettings } from '@/types/catalog'
import api from '@/api/client'

export const useThemeStore = defineStore('theme', () => {
  const catalogTheme = ref<CatalogThemeName>(defaultCatalogSettings.theme)
  const readerThemeOverride = ref<CatalogThemeName | null>(null)
  const loaded = ref(false)

  const effectiveReaderTheme = computed<CatalogThemeName>(
    () => readerThemeOverride.value ?? catalogTheme.value,
  )

  function setCatalogTheme(theme: CatalogThemeName) {
    catalogTheme.value = theme
    applyVuetifyTheme(theme)
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
    try {
      const vuetifyTheme = useTheme()
      vuetifyTheme.global.name.value = theme
    } catch {
      // useTheme() не доступен вне setup-контекста (тесты)
    }
  }

  async function loadSettings() {
    try {
      const { data } = await api.get<Record<string, unknown>>('/me/settings')

      const catalog = data.catalog as Record<string, unknown> | undefined
      if (catalog?.theme) {
        catalogTheme.value = catalog.theme as CatalogThemeName
        applyVuetifyTheme(catalogTheme.value)
      }

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
        catalog: { theme: catalogTheme.value },
        reader: { theme: readerThemeOverride.value },
      })
    } catch {
      // Ошибка сохранения — тема уже применена локально
    }
  }

  return {
    catalogTheme,
    readerThemeOverride,
    effectiveReaderTheme,
    loaded,
    setCatalogTheme,
    setReaderTheme,
    resetReaderTheme,
    loadSettings,
    saveSettings,
  }
})
