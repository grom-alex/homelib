<template>
  <div class="series-tab">
    <div class="series-tab__search">
      <v-text-field
        v-model="searchQuery"
        placeholder="Поиск серии..."
        density="compact"
        variant="outlined"
        hide-details
        clearable
        prepend-inner-icon="mdi-magnify"
        @update:model-value="onSearchInput"
      />
    </div>

    <div v-if="loading && series.length === 0" class="series-tab__status">
      <v-progress-circular indeterminate size="24" />
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
