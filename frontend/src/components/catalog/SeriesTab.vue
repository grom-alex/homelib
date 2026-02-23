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

    <div v-if="loading && series.length === 0" class="pa-4 text-center">
      <v-progress-circular indeterminate size="24" />
    </div>

    <div v-else-if="!loading && series.length === 0" class="pa-4 text-center text-body-2 text-medium-emphasis">
      Ничего не найдено
    </div>

    <v-list v-else density="compact" class="series-tab__list">
      <v-list-item
        v-for="item in series"
        :key="item.id"
        :active="catalog.navigationFilter?.type === 'series' && catalog.navigationFilter?.id === item.id"
        @click="selectSeries(item.id, item.name)"
      >
        <v-list-item-title class="text-body-2">{{ item.name }}</v-list-item-title>
        <template #append>
          <v-badge :content="String(item.books_count)" color="primary" inline />
        </template>
      </v-list-item>

      <div v-if="hasMore" class="text-center pa-2">
        <v-btn
          variant="text"
          size="small"
          :loading="loading"
          @click="loadMore"
        >
          Загрузить ещё
        </v-btn>
      </div>
    </v-list>
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
      series.value = result.items
    } else {
      series.value = [...series.value, ...result.items]
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
  padding: 8px;
  flex-shrink: 0;
}

.series-tab__list {
  flex: 1;
  overflow-y: auto;
}
</style>
