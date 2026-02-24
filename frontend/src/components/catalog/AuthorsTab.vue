<template>
  <div class="authors-tab">
    <div class="authors-tab__search">
      <div class="search-input-wrapper">
        <svg class="search-input-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
          <circle cx="11" cy="11" r="8" />
          <line x1="21" y1="21" x2="16.65" y2="16.65" />
        </svg>
        <input
          v-model="searchQuery"
          placeholder="Поиск автора..."
          @input="onSearchInput"
        />
        <button v-if="searchQuery" class="search-input-clear" @click="clearSearch">&times;</button>
      </div>
    </div>

    <div v-if="loading && authors.length === 0" class="authors-tab__status">
      <span class="spinner" />
    </div>

    <div v-else-if="!loading && authors.length === 0" class="authors-tab__status authors-tab__status--empty">
      Ничего не найдено
    </div>

    <div v-else class="authors-tab__list">
      <div
        v-for="author in authors"
        :key="author.id"
        class="authors-tab__item"
        :class="{ 'authors-tab__item--selected': catalog.navigationFilter?.type === 'author' && catalog.navigationFilter?.id === author.id }"
        @click="selectAuthor(author.id, author.name)"
      >
        <span class="authors-tab__item-name">{{ author.name }}</span>
        <span class="authors-tab__item-count">{{ author.books_count }}</span>
      </div>

      <div v-if="hasMore" class="authors-tab__load-more">
        <button
          class="authors-tab__load-more-btn"
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
import { ref, onMounted, onUnmounted } from 'vue'
import { useCatalogStore } from '@/stores/catalog'
import { getAuthors, type AuthorListItem } from '@/api/books'

const catalog = useCatalogStore()

const searchQuery = ref('')
const authors = ref<AuthorListItem[]>([])
const loading = ref(false)
const page = ref(1)
const hasMore = ref(false)
const limit = 50

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function clearSearch() {
  searchQuery.value = ''
  page.value = 1
  authors.value = []
  fetchAuthors()
}

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    page.value = 1
    authors.value = []
    fetchAuthors()
  }, 300)
}

async function fetchAuthors() {
  loading.value = true
  try {
    const result = await getAuthors({
      q: searchQuery.value || undefined,
      page: page.value,
      limit,
    })
    if (page.value === 1) {
      authors.value = result.items ?? []
    } else {
      authors.value = [...authors.value, ...(result.items ?? [])]
    }
    hasMore.value = authors.value.length < result.total
  } catch {
    // Ошибка загрузки авторов
  } finally {
    loading.value = false
  }
}

function loadMore() {
  page.value++
  fetchAuthors()
}

function selectAuthor(authorId: number, name: string) {
  catalog.selectNavItem('author', authorId, undefined, name)
}

onMounted(() => {
  fetchAuthors()
})

onUnmounted(() => {
  if (debounceTimer) clearTimeout(debounceTimer)
})
</script>

<style src="@/assets/nav-tab.css"></style>

<style scoped>
.authors-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.authors-tab__search {
  padding: 10px 10px 8px;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  flex-shrink: 0;
}

.authors-tab__status {
  padding: 20px;
  text-align: center;
}

.authors-tab__status--empty {
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
}

.authors-tab__list {
  flex: 1;
  overflow-y: auto;
}

.authors-tab__item {
  cursor: pointer;
  padding: 5px 10px;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
}

.authors-tab__item:hover {
  background: rgb(var(--v-theme-table-row-hover));
}

.authors-tab__item--selected {
  background: rgba(var(--v-theme-primary), 0.12);
  color: rgb(var(--v-theme-primary));
}

.authors-tab__item-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.authors-tab__item-count {
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  background: rgb(var(--v-theme-surface-variant));
  padding: 1px 7px;
  border-radius: 8px;
  flex-shrink: 0;
  margin-left: 8px;
}

.authors-tab__load-more {
  text-align: center;
  padding: 8px;
}

.authors-tab__load-more-btn {
  background: none;
  border: none;
  color: rgb(var(--v-theme-primary));
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  padding: 4px 12px;
}

.authors-tab__load-more-btn:hover {
  text-decoration: underline;
}

.authors-tab__load-more-btn:disabled {
  opacity: 0.5;
  cursor: default;
  text-decoration: none;
}
</style>
