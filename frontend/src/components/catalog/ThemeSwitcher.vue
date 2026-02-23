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
  </div>
</template>

<script setup lang="ts">
import { useThemeStore } from '@/stores/theme'
import type { CatalogThemeName } from '@/types/catalog'

const themeStore = useThemeStore()

const themes: Array<{ name: CatalogThemeName; label: string; previewColor: string; dark: boolean }> = [
  { name: 'light', label: 'Светлая', previewColor: '#FFFFFF', dark: false },
  { name: 'dark', label: 'Тёмная', previewColor: '#1E1E1E', dark: true },
  { name: 'sepia', label: 'Сепия', previewColor: '#F5F0E8', dark: false },
  { name: 'night', label: 'Ночная', previewColor: '#0D1117', dark: true },
]
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
</style>
