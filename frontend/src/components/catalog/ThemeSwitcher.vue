<template>
  <div class="theme-switcher">
    <div class="text-caption text-medium-emphasis mb-1 px-2">Тема</div>
    <div class="theme-switcher__options">
      <v-btn
        v-for="theme in themes"
        :key="theme.name"
        :variant="themeStore.catalogTheme === theme.name ? 'flat' : 'text'"
        :color="themeStore.catalogTheme === theme.name ? 'primary' : undefined"
        size="x-small"
        @click="themeStore.setCatalogTheme(theme.name)"
      >
        <span
          class="theme-switcher__preview mr-1"
          :style="{ background: theme.previewColor, border: theme.dark ? '1px solid #555' : '1px solid #ccc' }"
        />
        {{ theme.label }}
      </v-btn>
    </div>

    <div v-if="themeStore.catalogTheme === 'custom'" class="theme-switcher__custom">
      <label v-for="field in colorFields" :key="field.key" class="theme-switcher__color-field">
        <input
          type="color"
          :value="themeStore.customCatalogColors[field.key]"
          @input="onColorChange(field.key, ($event.target as HTMLInputElement).value)"
        />
        <span>{{ field.label }}</span>
      </label>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useThemeStore } from '@/stores/theme'
import type { CatalogThemeName, CustomCatalogColors } from '@/types/catalog'

const themeStore = useThemeStore()

const themes: Array<{ name: CatalogThemeName; label: string; previewColor: string; dark: boolean }> = [
  { name: 'light', label: 'Светлая', previewColor: '#FFFFFF', dark: false },
  { name: 'dark', label: 'Тёмная', previewColor: '#1E1E1E', dark: true },
  { name: 'sepia', label: 'Сепия', previewColor: '#f5e6d3', dark: false },
  { name: 'night', label: 'Ночная', previewColor: '#000000', dark: true },
  { name: 'custom', label: 'Своя', previewColor: 'linear-gradient(135deg, #ff6b6b, #4ecdc4)', dark: false },
]

const colorFields: Array<{ key: keyof CustomCatalogColors; label: string }> = [
  { key: 'background', label: 'Фон' },
  { key: 'text', label: 'Текст' },
  { key: 'link', label: 'Акцент' },
  { key: 'selection', label: 'Выделение' },
]

function onColorChange(key: keyof CustomCatalogColors, value: string) {
  themeStore.setCatalogCustomColors({
    ...themeStore.customCatalogColors,
    [key]: value,
  })
}
</script>

<style scoped>
.theme-switcher__options {
  display: flex;
  gap: 2px;
  padding: 0 4px;
}

.theme-switcher__preview {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.theme-switcher__custom {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  padding: 6px 6px 0;
}

.theme-switcher__color-field {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  cursor: pointer;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.7;
}

.theme-switcher__color-field input[type="color"] {
  width: 20px;
  height: 20px;
  border: 1px solid rgb(var(--v-theme-surface-variant));
  border-radius: 3px;
  padding: 0;
  cursor: pointer;
  background: none;
}
</style>
