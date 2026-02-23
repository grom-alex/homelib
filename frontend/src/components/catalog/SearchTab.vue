<template>
  <div class="search-tab">
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
        <v-btn type="submit" color="primary" variant="flat" size="small" block>
          Найти
        </v-btn>
        <v-btn variant="outlined" size="small" block class="mt-1" @click="onClear">
          Очистить
        </v-btn>
      </div>
    </v-form>
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
  padding: 8px;
  height: 100%;
  overflow-y: auto;
}

.search-tab__form {
  display: flex;
  flex-direction: column;
}

.search-tab__actions {
  margin-top: 8px;
}
</style>
