<template>
  <div class="search-tab">
    <div class="search-tab__header">Критерии поиска</div>
    <div class="search-tab__body">
      <form class="search-tab__form" @submit.prevent="onSubmit">
        <div class="search-field">
          <label>Название</label>
          <input v-model="form.q" placeholder="Введите название..." />
        </div>

        <div class="search-field">
          <label>Автор</label>
          <input v-model="form.author_name" placeholder="Введите автора..." />
        </div>

        <div class="search-field">
          <label>Серия</label>
          <input v-model="form.series_name" placeholder="Введите серию..." />
        </div>

        <div v-if="genreOptions.length > 0" class="search-field">
          <label>Жанр</label>
          <select v-model="form.genre_id">
            <option :value="null">Все жанры</option>
            <option v-for="g in genreOptions" :key="g.id" :value="g.id">{{ g.name }}</option>
          </select>
        </div>

        <div v-if="formatOptions.length > 0" class="search-field">
          <label>Формат</label>
          <select v-model="form.format">
            <option :value="null">Все форматы</option>
            <option v-for="f in formatOptions" :key="f" :value="f">{{ f }}</option>
          </select>
        </div>

        <div v-if="langOptions.length > 0" class="search-field">
          <label>Язык</label>
          <select v-model="form.lang">
            <option :value="null">Все языки</option>
            <option v-for="l in langOptions" :key="l" :value="l">{{ l }}</option>
          </select>
        </div>

        <div class="search-tab__actions">
          <button type="submit" class="search-tab__btn-find">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="11" cy="11" r="8" /><line x1="21" y1="21" x2="16.65" y2="16.65" />
            </svg>
            Найти
          </button>
          <button type="button" class="search-tab__btn-clear" @click="onClear">
            Очистить
          </button>
        </div>
      </form>
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
  author_name: '',
  series_name: '',
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
  const [genresResult, statsResult] = await Promise.allSettled([getGenres(), getStats()])
  if (genresResult.status === 'fulfilled') {
    genreOptions.value = flattenGenres(genresResult.value)
  }
  if (statsResult.status === 'fulfilled') {
    formatOptions.value = statsResult.value.formats
    langOptions.value = statsResult.value.languages
  }
}

function onSubmit() {
  const params: Record<string, string> = {}
  if (form.q) params.q = form.q
  if (form.author_name) params.author_name = form.author_name
  if (form.series_name) params.series_name = form.series_name
  if (form.genre_id) params.genre_id = String(form.genre_id)
  if (form.format) params.format = form.format
  if (form.lang) params.lang = form.lang

  const label = form.q || form.author_name || form.series_name || 'Расширенный поиск'
  catalog.selectNavItem('search', undefined, params, label)
}

function onClear() {
  form.q = ''
  form.author_name = ''
  form.series_name = ''
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
  gap: 12px;
}

.search-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.search-field label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.45;
  font-weight: 600;
}

.search-field input,
.search-field select {
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgb(var(--v-theme-surface-variant));
  color: rgb(var(--v-theme-on-surface));
  padding: 6px 10px;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s;
}

.search-field input:focus,
.search-field select:focus {
  border-color: rgb(var(--v-theme-primary));
}

.search-field select option {
  background: rgb(var(--v-theme-surface));
}

.search-tab__actions {
  margin-top: 4px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.search-tab__btn-find {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
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
