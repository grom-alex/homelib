<template>
  <div class="book-table">
    <div v-if="catalog.loading" class="book-table__loading">
      <v-progress-linear indeterminate color="primary" />
    </div>

    <div v-if="catalog.error" class="book-table__error pa-4">
      <v-alert type="error" density="compact">{{ catalog.error }}</v-alert>
    </div>

    <div v-if="!catalog.loading && !catalog.navigationFilter" class="book-table__empty">
      <v-icon size="48" color="grey">mdi-book-open-blank-variant</v-icon>
      <p class="book-table__empty-text">Выберите элемент навигации</p>
    </div>

    <div v-else-if="!catalog.loading && catalog.books?.length === 0 && catalog.navigationFilter" class="book-table__empty">
      <v-icon size="48" color="grey">mdi-book-off-outline</v-icon>
      <p class="book-table__empty-text">Книги не найдены</p>
    </div>

    <template v-else>
      <div class="book-table__grid" role="table">
        <div class="book-table__header" role="row">
          <div
            v-for="col in columns"
            :key="col.field"
            class="book-table__header-cell"
            :class="{
              'book-table__header-cell--sortable': col.sortable,
              'book-table__header-cell--active': col.sortable && catalog.filters.sort === col.field,
            }"
            :style="{ width: col.width }"
            role="columnheader"
            @click="col.sortable && onSortClick(col.field)"
          >
            {{ col.label }}
            <span
              v-if="col.sortable && catalog.filters.sort === col.field"
              class="book-table__sort-arrow"
            >{{ catalog.filters.order === 'asc' ? '▲' : '▼' }}</span>
          </div>
        </div>

        <div class="book-table__body" role="rowgroup">
          <div
            v-for="book in catalog.books"
            :key="book.id"
            class="book-table__row"
            :class="{ 'book-table__row--selected': catalog.selectedBookId === book.id }"
            role="row"
            tabindex="0"
            :data-book-id="book.id"
            @click="selectBook(book.id)"
            @keydown.enter="onEnterKey"
            @keydown.up.prevent="navigateRow(-1, $event)"
            @keydown.down.prevent="navigateRow(1, $event)"
          >
            <div class="book-table__cell" :style="{ width: columns[0].width }">
              {{ book.title }}
            </div>
            <div class="book-table__cell" :style="{ width: columns[1].width }">
              {{ formatAuthors(book.authors) }}
            </div>
            <div
              class="book-table__cell"
              :class="{ 'book-table__cell--muted': !book.series }"
              :style="{ width: columns[2].width }"
            >
              {{ formatSeries(book) }}
            </div>
            <div class="book-table__cell" :style="{ width: columns[3].width }">
              {{ formatGenres(book.genres) }}
            </div>
            <div class="book-table__cell book-table__cell--mono" :style="{ width: columns[4].width }">
              {{ formatFileSize(book.file_size) }}
            </div>
          </div>
        </div>
      </div>

      <div v-if="catalog.totalPages > 1" class="book-table__pagination">
        <div class="pagination__size">
          <select
            :value="catalog.filters.limit"
            class="pagination__size-select"
            @change="onPageSizeChange($event)"
          >
            <option v-for="s in pageSizes" :key="s" :value="s">
              {{ s }}
            </option>
          </select>
        </div>

        <div class="pagination__nav">
          <button
            class="pagination__btn"
            :disabled="currentPage <= 1"
            title="В начало"
            @click="catalog.setPage(1)"
          >
            &laquo;&laquo;
          </button>
          <button
            class="pagination__btn"
            :disabled="currentPage <= 1"
            title="-10 страниц"
            @click="catalog.setPage(Math.max(1, currentPage - 10))"
          >
            &laquo;
          </button>
          <button
            class="pagination__btn"
            :disabled="currentPage <= 1"
            title="Предыдущая"
            @click="catalog.setPage(currentPage - 1)"
          >
            &lsaquo;
          </button>

          <button
            v-for="p in visiblePages"
            :key="p"
            class="pagination__page"
            :class="{ 'pagination__page--active': p === currentPage }"
            @click="catalog.setPage(p)"
          >
            {{ p }}
          </button>

          <button
            class="pagination__btn"
            :disabled="currentPage >= catalog.totalPages"
            title="Следующая"
            @click="catalog.setPage(currentPage + 1)"
          >
            &rsaquo;
          </button>
          <button
            class="pagination__btn"
            :disabled="currentPage >= catalog.totalPages"
            title="+10 страниц"
            @click="catalog.setPage(Math.min(catalog.totalPages, currentPage + 10))"
          >
            &raquo;
          </button>
          <button
            class="pagination__btn"
            :disabled="currentPage >= catalog.totalPages"
            title="В конец"
            @click="catalog.setPage(catalog.totalPages)"
          >
            &raquo;&raquo;
          </button>
        </div>

        <div class="pagination__info">
          Стр. {{ currentPage }} из {{ catalog.totalPages }}
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useCatalogStore } from '@/stores/catalog'
import type { PageSize } from '@/types/catalog'

const router = useRouter()

const catalog = useCatalogStore()

const pageSizes: PageSize[] = [25, 50, 75, 100]

const currentPage = computed(() => catalog.filters.page || 1)

const visiblePages = computed(() => {
  const total = catalog.totalPages
  const current = currentPage.value
  const windowSize = 10

  if (total <= windowSize) {
    return Array.from({ length: total }, (_, i) => i + 1)
  }

  let start = Math.max(1, current - Math.floor(windowSize / 2))
  let end = start + windowSize - 1

  if (end > total) {
    end = total
    start = Math.max(1, end - windowSize + 1)
  }

  return Array.from({ length: end - start + 1 }, (_, i) => start + i)
})

function onPageSizeChange(event: Event) {
  const value = Number((event.target as HTMLSelectElement).value) as PageSize
  catalog.setPageSize(value)
}

interface Column {
  field: string
  label: string
  width: string
  sortable: boolean
}

const columns: Column[] = [
  { field: 'title', label: 'Название', width: '35%', sortable: true },
  { field: 'author', label: 'Автор', width: '22%', sortable: false },
  { field: 'series', label: 'Серия', width: '18%', sortable: false },
  { field: 'genre', label: 'Жанр', width: '15%', sortable: false },
  { field: 'file_size', label: 'Размер', width: '10%', sortable: true },
]

function onSortClick(field: string) {
  if (field !== 'title' && field !== 'file_size' && field !== 'year') return
  const order = catalog.filters.sort === field && catalog.filters.order === 'asc' ? 'desc' : 'asc'
  catalog.setSort(field as 'title' | 'year' | 'file_size', order)
}

function selectBook(id: number) {
  catalog.setSelectedBook(id)
}

function navigateRow(direction: number, event: KeyboardEvent) {
  const target = event.target as HTMLElement
  const sibling = direction > 0
    ? target.nextElementSibling as HTMLElement
    : target.previousElementSibling as HTMLElement
  if (sibling) {
    sibling.focus()
    const bookId = Number(sibling.dataset?.bookId)
    if (bookId) selectBook(bookId)
  }
}

function onEnterKey() {
  const book = catalog.currentBook
  if (book && book.format === 'fb2') {
    router.push(`/books/${book.id}/read`)
  }
}

function formatAuthors(authors: Array<{ id: number; name: string }>): string {
  if (!authors || authors.length === 0) return '—'
  if (authors.length === 1) return authors[0].name
  return `${authors[0].name} и др.`
}

function formatSeries(book: { series?: { name: string; num?: number } }): string {
  if (!book.series) return '—'
  return book.series.num
    ? `${book.series.name} #${book.series.num}`
    : book.series.name
}

function formatGenres(genres: Array<{ id: number; name: string }>): string {
  if (!genres || genres.length === 0) return '—'
  return genres[0].name
}

function formatFileSize(bytes?: number): string {
  if (!bytes) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
</script>

<style scoped>
.book-table {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: rgb(var(--v-theme-surface));
  overflow: hidden;
}

.book-table__loading {
  flex-shrink: 0;
}

.book-table__error {
  flex-shrink: 0;
}

.book-table__empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.book-table__empty-text {
  font-size: 13px;
  margin-top: 8px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
}

.book-table__grid {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.book-table__header {
  display: flex;
  background: rgb(var(--v-theme-surface));
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  flex-shrink: 0;
}

.book-table__header-cell {
  padding: 7px 10px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.35;
  user-select: none;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  border-right: 1px solid rgb(var(--v-theme-surface-variant));
}

.book-table__header-cell--sortable {
  cursor: pointer;
  transition: color 0.15s;
}

.book-table__header-cell--sortable:hover {
  opacity: 0.7;
}

.book-table__header-cell--active {
  color: rgb(var(--v-theme-primary));
  opacity: 1;
}

.book-table__sort-arrow {
  margin-left: 4px;
  font-size: 10px;
  opacity: 0.7;
}

.book-table__body {
  flex: 1;
  overflow-y: auto;
}

.book-table__row {
  display: flex;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  cursor: pointer;
  outline: none;
  transition: background-color 0.1s;
  font-size: 13px;
}

.book-table__row:hover {
  background: rgb(var(--v-theme-table-row-hover));
}

.book-table__row--selected {
  background: rgb(var(--v-theme-table-row-selected));
}

.book-table__row:focus {
  background: rgb(var(--v-theme-table-row-selected));
}

.book-table__cell {
  padding: 6px 10px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.6;
  border-right: 1px solid rgb(var(--v-theme-surface-variant));
}

.book-table__cell--muted {
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.35;
  font-style: italic;
}

.book-table__cell--mono {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 12px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.6;
}

.book-table__pagination {
  border-top: 1px solid rgb(var(--v-theme-surface-variant));
  padding: 4px 12px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
}

.pagination__size {
  flex-shrink: 0;
}

.pagination__size-select {
  font-family: inherit;
  font-size: 11px;
  padding: 2px 4px;
  border: 1px solid rgb(var(--v-theme-surface-variant));
  border-radius: 3px;
  background: rgb(var(--v-theme-surface));
  color: rgb(var(--v-theme-on-surface));
  cursor: pointer;
  outline: none;
}

.pagination__size-select:focus {
  border-color: rgb(var(--v-theme-primary));
}

.pagination__nav {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2px;
}

.pagination__btn {
  font-family: inherit;
  font-size: 12px;
  padding: 2px 6px;
  border: none;
  background: transparent;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.6;
  cursor: pointer;
  border-radius: 3px;
  line-height: 1.4;
}

.pagination__btn:hover:not(:disabled) {
  opacity: 1;
  background: rgb(var(--v-theme-table-row-hover));
}

.pagination__btn:disabled {
  opacity: 0.2;
  cursor: default;
}

.pagination__page {
  font-family: inherit;
  font-size: 12px;
  min-width: 24px;
  padding: 2px 4px;
  border: none;
  background: transparent;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.6;
  cursor: pointer;
  border-radius: 3px;
  text-align: center;
  line-height: 1.4;
}

.pagination__page:hover {
  opacity: 1;
  background: rgb(var(--v-theme-table-row-hover));
}

.pagination__page--active {
  background: rgb(var(--v-theme-primary));
  color: #fff;
  opacity: 1;
}

.pagination__page--active:hover {
  background: rgb(var(--v-theme-primary));
}

.pagination__info {
  flex-shrink: 0;
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.45;
  white-space: nowrap;
}
</style>
