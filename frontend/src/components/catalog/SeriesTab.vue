<template>
  <div class="series-tab">
    <div class="series-tab__search">
      <div class="search-input-wrapper">
        <svg class="search-input-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        <input
          v-model="searchQuery"
          placeholder="Поиск серии..."
          @input="onSearchInput"
        />
        <button v-if="searchQuery" class="search-input-clear" @click="clearSearch">&times;</button>
      </div>
    </div>

    <div v-if="loading && series.length === 0" class="series-tab__status">
      <span class="spinner" />
    </div>

    <div v-else-if="!loading && series.length === 0" class="series-tab__status series-tab__status--empty">
      Ничего не найдено
    </div>

    <div v-else class="series-tab__list">
      <div
        v-for="item in series"
        :key="item.id"
        class="series-tab__item"
        :class="{ 'series-tab__item--selected': catalog.navigationFilter?.type === 'series' && catalog.navigationFilter?.id === item.id }"
        @click="selectSeries(item.id, item.name)"
      >
        <div class="series-tab__item-info">
          <span class="series-tab__item-name">{{ item.name }}</span>
          <span v-if="item.authors" class="series-tab__item-authors">{{ item.authors }}</span>
        </div>
        <span class="series-tab__item-count">{{ item.books_count }}</span>
      </div>

      <div v-if="hasMore" class="series-tab__load-more">
        <button
          class="series-tab__load-more-btn"
          :disabled="loading"
          @click="loadMore"
        >
          {{ loading ? 'Загрузка...' : 'Загрузить ещё' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useCatalogStore } from '@/stores/catalog'
import { getSeries, type SeriesListItem } from '@/api/books'

const catalog = useCatalogStore()

const searchQuery = ref('')
const series = ref<SeriesListItem[]>([])
const loading = ref(false)
const page = ref(1)
const hasMore = ref(false)
const limit = 50

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function clearSearch() {
  searchQuery.value = ''
  page.value = 1
  series.value = []
  fetchSeries()
}

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    page.value = 1
    series.value = []
    fetchSeries()
  }, 300)
}

async function fetchSeries() {
  loading.value = true
  try {
    const result = await getSeries({
      q: searchQuery.value || undefined,
      page: page.value,
      limit,
    })
    if (page.value === 1) {
      series.value = result.items ?? []
    } else {
      series.value = [...series.value, ...(result.items ?? [])]
    }
    hasMore.value = series.value.length < result.total
  } catch {
    // Ошибка загрузки серий
  } finally {
    loading.value = false
  }
}

function loadMore() {
  page.value++
  fetchSeries()
}

function selectSeries(seriesId: number, name: string) {
  catalog.selectNavItem('series', seriesId, undefined, name)
}

onMounted(() => {
  fetchSeries()
})
</script>

<style scoped>
.series-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.series-tab__search {
  padding: 10px 10px 8px;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  flex-shrink: 0;
}

.search-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.search-input-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.35;
  pointer-events: none;
}

.search-input-wrapper input {
  width: 100%;
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgb(var(--v-theme-surface-variant));
  color: rgb(var(--v-theme-on-surface));
  padding: 6px 28px 6px 32px;
  border-radius: 4px;
  font-size: 13px;
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s;
}

.search-input-wrapper input:focus {
  border-color: rgb(var(--v-theme-primary));
}

.search-input-wrapper input::placeholder {
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.3;
}

.search-input-clear {
  position: absolute;
  right: 6px;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  cursor: pointer;
  font-size: 16px;
  line-height: 1;
  padding: 0 4px;
}

.search-input-clear:hover {
  opacity: 0.7;
}

.spinner {
  display: inline-block;
  width: 24px;
  height: 24px;
  border: 2.5px solid rgb(var(--v-theme-surface-variant));
  border-top-color: rgb(var(--v-theme-primary));
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.series-tab__status {
  padding: 20px;
  text-align: center;
}

.series-tab__status--empty {
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
}

.series-tab__list {
  flex: 1;
  overflow-y: auto;
}

.series-tab__item {
  cursor: pointer;
  padding: 5px 10px;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
}

.series-tab__item:hover {
  background: rgb(var(--v-theme-table-row-hover));
}

.series-tab__item--selected {
  background: rgba(var(--v-theme-primary), 0.12);
  color: rgb(var(--v-theme-primary));
}

.series-tab__item-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.series-tab__item-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.series-tab__item-authors {
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.45;
  margin-top: 1px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.series-tab__item-count {
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  background: rgb(var(--v-theme-surface-variant));
  padding: 1px 7px;
  border-radius: 8px;
  flex-shrink: 0;
  margin-left: 8px;
}

.series-tab__load-more {
  text-align: center;
  padding: 8px;
}

.series-tab__load-more-btn {
  background: none;
  border: none;
  color: rgb(var(--v-theme-primary));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  padding: 4px 12px;
}

.series-tab__load-more-btn:hover {
  text-decoration: underline;
}

.series-tab__load-more-btn:disabled {
  opacity: 0.5;
  cursor: default;
  text-decoration: none;
}
</style>
