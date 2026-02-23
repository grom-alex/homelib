<template>
  <v-dialog :model-value="modelValue" max-width="500" @update:model-value="$emit('update:modelValue', $event)">
    <v-card>
      <v-card-title class="d-flex align-center">
        Настройки
        <v-spacer />
        <v-btn icon="mdi-close" variant="text" size="small" @click="$emit('update:modelValue', false)" />
      </v-card-title>

      <v-card-text>
        <div class="text-subtitle-2 mb-2">Тема каталога</div>
        <div class="d-flex gap-2">
          <v-btn
            v-for="theme in themes"
            :key="theme.name"
            :variant="themeStore.catalogTheme === theme.name ? 'flat' : 'outlined'"
            :color="themeStore.catalogTheme === theme.name ? 'primary' : undefined"
            size="small"
            @click="themeStore.setCatalogTheme(theme.name)"
          >
            {{ theme.label }}
          </v-btn>
        </div>

        <div class="text-subtitle-2 mt-4 mb-2">Тема читалки</div>
        <div class="d-flex gap-2 flex-wrap">
          <v-btn
            :variant="themeStore.readerThemeOverride === null ? 'flat' : 'outlined'"
            :color="themeStore.readerThemeOverride === null ? 'primary' : undefined"
            size="small"
            @click="themeStore.resetReaderTheme()"
          >
            Тема каталога
          </v-btn>
          <v-btn
            v-for="theme in themes"
            :key="theme.name"
            :variant="themeStore.readerThemeOverride === theme.name ? 'flat' : 'outlined'"
            :color="themeStore.readerThemeOverride === theme.name ? 'primary' : undefined"
            size="small"
            @click="themeStore.setReaderTheme(theme.name)"
          >
            {{ theme.label }}
          </v-btn>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { useThemeStore } from '@/stores/theme'
import type { CatalogThemeName } from '@/types/catalog'

defineProps<{ modelValue: boolean }>()
defineEmits<{ 'update:modelValue': [value: boolean] }>()

const themeStore = useThemeStore()

const themes: Array<{ name: CatalogThemeName; label: string }> = [
  { name: 'light', label: 'Светлая' },
  { name: 'dark', label: 'Тёмная' },
  { name: 'sepia', label: 'Сепия' },
  { name: 'night', label: 'Ночная' },
  { name: 'custom', label: 'Своя' },
]
</script>
