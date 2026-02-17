<template>
  <v-card variant="outlined">
    <v-card-title class="text-subtitle-1">Фильтры</v-card-title>
    <v-card-text>
      <v-text-field
        v-model="localFilters.q"
        label="Поиск"
        prepend-inner-icon="mdi-magnify"
        clearable
        density="compact"
        hide-details
        class="mb-3"
        @update:model-value="onFilterChange"
        @click:clear="localFilters.q = ''; onFilterChange()"
      />
      <v-select
        v-model="localFilters.lang"
        :items="languages"
        label="Язык"
        clearable
        density="compact"
        hide-details
        class="mb-3"
        @update:model-value="onFilterChange"
      />
      <v-select
        v-model="localFilters.format"
        :items="formats"
        label="Формат"
        clearable
        density="compact"
        hide-details
        class="mb-3"
        @update:model-value="onFilterChange"
      />
      <v-select
        v-model="localFilters.sort"
        :items="sortOptions"
        label="Сортировка"
        density="compact"
        hide-details
        class="mb-3"
        @update:model-value="onFilterChange"
      />
      <v-btn variant="text" size="small" @click="clearFilters">Сбросить фильтры</v-btn>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import { reactive, onMounted, onUnmounted } from 'vue'
import { getStats } from '@/api/books'
import type { BookFilters } from '@/api/books'

const emit = defineEmits<{
  (e: 'update', filters: Partial<BookFilters>): void
}>()

const languages = reactive<string[]>([])
const formats = reactive<string[]>([])
const sortOptions = [
  { title: 'По названию', value: 'title' },
  { title: 'По году', value: 'year' },
  { title: 'По дате добавления', value: 'added_at' },
]

const localFilters = reactive<Partial<BookFilters>>({
  q: '',
  lang: undefined,
  format: undefined,
  sort: 'title',
})

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onFilterChange() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    emit('update', { ...localFilters })
  }, 300)
}

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
})

function clearFilters() {
  localFilters.q = ''
  localFilters.lang = undefined
  localFilters.format = undefined
  localFilters.sort = 'title'
  emit('update', { ...localFilters })
}

onMounted(async () => {
  try {
    const stats = await getStats()
    languages.push(...stats.languages)
    formats.push(...stats.formats)
  } catch {
    // Stats not critical for filters
  }
})
</script>
