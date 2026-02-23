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

    <div v-else-if="!catalog.loading && catalog.books.length === 0 && catalog.navigationFilter" class="book-table__empty">
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
        <v-pagination
          :model-value="catalog.filters.page || 1"
          :length="catalog.totalPages"
          density="compact"
          size="small"
          @update:model-value="catalog.setPage($event)"
        />
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useCatalogStore } from '@/stores/catalog'

const router = useRouter()

const catalog = useCatalogStore()

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
  padding: 4px;
  flex-shrink: 0;
  display: flex;
  justify-content: center;
}
</style>
