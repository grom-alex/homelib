<template>
  <div class="search-tab">
    <div class="search-tab__header">Критерии поиска</div>
    <div class="search-tab__body">
      <v-form class="search-tab__form" @submit.prevent="onSubmit">
        <v-text-field
          v-model="form.q"
          label="Название"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          class="mb-2"
        />

        <v-select
          v-if="genreOptions.length > 0"
          v-model="form.genre_id"
          :items="genreOptions"
          item-title="name"
          item-value="id"
          label="Жанр"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          class="mb-2"
        />

        <v-select
          v-if="formatOptions.length > 0"
          v-model="form.format"
          :items="formatOptions"
          label="Формат"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          class="mb-2"
        />

        <v-select
          v-if="langOptions.length > 0"
          v-model="form.lang"
          :items="langOptions"
          label="Язык"
          density="compact"
          variant="outlined"
          hide-details
          clearable
          class="mb-2"
        />

        <div class="search-tab__actions">
          <button type="submit" class="search-tab__btn-find">
            <v-icon size="14" class="mr-1">mdi-magnify</v-icon>
            Найти
          </button>
          <button type="button" class="search-tab__btn-clear" @click="onClear">
            Очистить
          </button>
        </div>
      </v-form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useCatalogStore } from '@/stores/catalog'
import { getGenres, getStats, type GenreTreeItem } from '@/api/books'

const catalog = useCatalogStore()

const form = reactive({
  q: '',
  genre_id: null as number | null,
  format: null as string | null,
  lang: null as string | null,
})

const genreOptions = ref<Array<{ id: number; name: string }>>([])
const formatOptions = ref<string[]>([])
const langOptions = ref<string[]>([])

function flattenGenres(genres: GenreTreeItem[]): Array<{ id: number; name: string }> {
  const result: Array<{ id: number; name: string }> = []
  for (const genre of genres) {
    const prefix = genre.meta_group ? `${genre.meta_group} / ` : ''
    result.push({ id: genre.id, name: `${prefix}${genre.name}` })
    if (genre.children) {
      for (const child of genre.children) {
        result.push({ id: child.id, name: `${prefix}${child.name}` })
      }
    }
  }
  return result.sort((a, b) => a.name.localeCompare(b.name, 'ru'))
}

async function loadOptions() {
  try {
    const [genres, stats] = await Promise.all([getGenres(), getStats()])
    genreOptions.value = flattenGenres(genres)
    formatOptions.value = stats.formats
    langOptions.value = stats.languages
  } catch {
    // Ошибка загрузки опций
  }
}

function onSubmit() {
  const params: Record<string, string> = {}
  if (form.q) params.q = form.q
  if (form.genre_id) params.genre_id = String(form.genre_id)
  if (form.format) params.format = form.format
  if (form.lang) params.lang = form.lang

  catalog.selectNavItem('search', undefined, params, form.q || 'Расширенный поиск')
}

function onClear() {
  form.q = ''
  form.genre_id = null
  form.format = null
  form.lang = null
  catalog.resetFilters()
}

onMounted(() => {
  loadOptions()
})
</script>

<style scoped>
.search-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.search-tab__header {
  padding: 8px 6px;
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 600;
  flex-shrink: 0;
}

.search-tab__body {
  flex: 1;
  overflow-y: auto;
  padding: 12px 10px;
}

.search-tab__form {
  display: flex;
  flex-direction: column;
}

.search-tab__actions {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.search-tab__btn-find {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 9px 0;
  background: rgb(var(--v-theme-primary));
  color: #1a1d23;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  font-family: inherit;
  transition: filter 0.2s;
  width: 100%;
}

.search-tab__btn-find:hover {
  filter: brightness(1.1);
}

.search-tab__btn-clear {
  padding: 7px 0;
  background: transparent;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.5;
  border: 1px solid rgb(var(--v-theme-surface-variant));
  border-radius: 4px;
  cursor: pointer;
  font-size: 12px;
  font-family: inherit;
  transition: all 0.2s;
  width: 100%;
}

.search-tab__btn-clear:hover {
  opacity: 0.8;
  border-color: rgb(var(--v-theme-on-surface));
}
</style>
