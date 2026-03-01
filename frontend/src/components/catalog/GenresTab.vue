<template>
  <div class="genres-tab">
    <div class="genres-tab__search">
      <div class="search-input-wrapper">
        <svg class="search-input-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        <input
          v-model="searchQuery"
          placeholder="Поиск жанра..."
          @input="onSearchInput"
        />
        <button v-if="searchQuery" class="search-input-clear" @click="clearSearch">&times;</button>
      </div>
      <button
        class="genres-tab__sort-btn"
        :title="themeStore.genreSortOrder === 'original' ? 'Порядок из файла' : 'По алфавиту'"
        @click="toggleSort"
      >
        <svg v-if="themeStore.genreSortOrder === 'alphabetical'" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <text x="3" y="10" font-size="8" stroke="none" fill="currentColor" font-weight="bold">Aa</text>
          <line x1="17" y1="4" x2="17" y2="20" /><polyline points="13 16 17 20 21 16" />
        </svg>
        <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <line x1="4" y1="6" x2="14" y2="6" /><line x1="4" y1="12" x2="11" y2="12" /><line x1="4" y1="18" x2="8" y2="18" />
          <line x1="17" y1="4" x2="17" y2="20" /><polyline points="13 16 17 20 21 16" />
        </svg>
      </button>
    </div>

    <div v-if="loading" class="genres-tab__status">
      <span class="spinner" />
    </div>

    <div v-else-if="genres.length === 0" class="genres-tab__status genres-tab__status--empty">
      Жанры не найдены
    </div>

    <div v-else class="genres-tab__tree">
      <div v-if="debouncedSearch && filteredCount === 0" class="genres-tab__status genres-tab__status--empty">
        Жанры не найдены
      </div>
      <v-treeview
        v-show="!debouncedSearch || filteredCount > 0"
        :items="sortedGenres"
        item-value="id"
        item-title="name"
        item-children="children"
        activatable
        open-on-click
        density="compact"
        slim
        :search="debouncedSearch"
        :activated="activated"
        @update:activated="onActivated"
      >
        <template #append="{ item }">
          <span class="genres-tab__count">{{ item.books_count }}</span>
        </template>
      </v-treeview>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { useCatalogStore } from '@/stores/catalog'
import { useThemeStore } from '@/stores/theme'
import { useParentalStore } from '@/stores/parental'
import { getGenres, type GenreTreeItem } from '@/api/books'

const catalog = useCatalogStore()
const themeStore = useThemeStore()
const parentalStore = useParentalStore()

const genres = ref<GenreTreeItem[]>([])
const loading = ref(false)
const searchQuery = ref('')
const debouncedSearch = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function sortGenreTree(items: GenreTreeItem[]): GenreTreeItem[] {
  const sorted = [...items].sort((a, b) => a.name.localeCompare(b.name, 'ru'))
  return sorted.map(item => ({
    ...item,
    children: item.children ? sortGenreTree(item.children) : undefined,
  }))
}

const sortedGenres = computed(() => {
  if (themeStore.genreSortOrder === 'alphabetical') {
    return sortGenreTree(genres.value)
  }
  return genres.value
})

function toggleSort() {
  themeStore.setGenreSortOrder(
    themeStore.genreSortOrder === 'original' ? 'alphabetical' : 'original',
  )
}

// Build a flat lookup map for id → genre
const genreMap = computed(() => {
  const map = new Map<number, GenreTreeItem>()
  function walk(items: GenreTreeItem[]) {
    for (const item of items) {
      map.set(item.id, item)
      if (item.children) walk(item.children)
    }
  }
  walk(genres.value)
  return map
})

// Count how many genres match the search (for empty state)
const filteredCount = computed(() => {
  if (!debouncedSearch.value) return genreMap.value.size
  const q = debouncedSearch.value.toLowerCase()
  let count = 0
  for (const genre of genreMap.value.values()) {
    if (genre.name.toLowerCase().includes(q)) count++
  }
  return count
})

// Sync activated state with catalog store
const activated = computed(() => {
  if (catalog.navigationFilter?.type === 'genre' && catalog.navigationFilter?.id) {
    return [catalog.navigationFilter.id]
  }
  return []
})

function onActivated(ids: number[]) {
  if (ids.length > 0) {
    const id = ids[0]
    const genre = genreMap.value.get(id)
    catalog.selectNavItem('genre', id, undefined, genre?.name)
  }
}

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    debouncedSearch.value = searchQuery.value
  }, 300)
}

function clearSearch() {
  searchQuery.value = ''
  debouncedSearch.value = ''
  if (debounceTimer) clearTimeout(debounceTimer)
}

async function fetchGenres() {
  loading.value = true
  try {
    genres.value = (await getGenres()) ?? []
  } catch {
    // Ошибка загрузки жанров
  } finally {
    loading.value = false
  }
}

// Re-fetch genres when parental content status changes (backend filters server-side)
watch(() => parentalStore.adultContentEnabled, () => {
  fetchGenres()
})

onMounted(() => {
  fetchGenres()
})

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style src="@/assets/nav-tab.css"></style>

<style scoped>
.genres-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
}

.genres-tab__search {
  padding: 10px 10px 8px;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 6px;
}

.genres-tab__sort-btn {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border: 1px solid rgb(var(--v-theme-surface-variant));
  border-radius: 4px;
  background: transparent;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.5;
  cursor: pointer;
  transition: all 0.2s;
}

.genres-tab__sort-btn:hover {
  opacity: 1;
  border-color: rgb(var(--v-theme-primary));
  color: rgb(var(--v-theme-primary));
}

.genres-tab__status {
  padding: 20px;
  text-align: center;
}

.genres-tab__status--empty {
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
}

.genres-tab__tree {
  flex: 1;
  overflow-y: auto;
  padding: 2px 0;
}

.genres-tab__count {
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  background: rgb(var(--v-theme-surface-variant));
  padding: 1px 7px;
  border-radius: 8px;
  flex-shrink: 0;
}

/* Compact VTreeview styling to match navigation panel */
.genres-tab__tree :deep(.v-treeview-item) {
  min-height: 28px;
}

.genres-tab__tree :deep(.v-list-item-title) {
  font-size: 13px;
}

.genres-tab__tree :deep(.v-list-item--active) {
  background: rgba(var(--v-theme-primary), 0.12);
  color: rgb(var(--v-theme-primary));
}
</style>
