<template>
  <div class="book-detail-panel">
    <div v-if="!catalog.selectedBookId" class="book-detail-panel__empty">
      <v-icon size="48" color="grey">mdi-book-information-variant</v-icon>
      <p class="book-detail-panel__empty-text">Выберите книгу для просмотра подробной информации</p>
    </div>

    <div v-else-if="catalog.bookLoading" class="book-detail-panel__loading">
      <v-progress-circular indeterminate size="32" />
    </div>

    <div v-else-if="catalog.currentBook" class="book-detail-panel__content">
      <div class="book-detail-panel__top">
        <div class="book-detail-panel__cover">
          📖
        </div>
        <div class="book-detail-panel__info">
          <h2 class="book-detail-panel__title">{{ catalog.currentBook.title }}</h2>
          <div class="book-detail-panel__author">{{ formatAuthors(catalog.currentBook.authors) }}</div>

          <div class="book-detail-panel__meta">
            <div class="book-detail-panel__meta-item">
              <span class="book-detail-panel__meta-label">Серия: </span>
              {{ catalog.currentBook.series ? catalog.currentBook.series.name : '—' }}
              <template v-if="catalog.currentBook.series?.num"> (#{{ catalog.currentBook.series.num }})</template>
            </div>
            <div class="book-detail-panel__meta-item">
              <span class="book-detail-panel__meta-label">Жанр: </span>
              {{ formatGenres(catalog.currentBook.genres) }}
            </div>
            <div v-if="catalog.currentBook.year" class="book-detail-panel__meta-item">
              <span class="book-detail-panel__meta-label">Год: </span>
              {{ catalog.currentBook.year }}
            </div>
            <div class="book-detail-panel__meta-item">
              <span class="book-detail-panel__meta-label">Формат: </span>
              <span class="book-detail-panel__mono">{{ catalog.currentBook.format }}</span>
            </div>
            <div class="book-detail-panel__meta-item">
              <span class="book-detail-panel__meta-label">Размер: </span>
              <span class="book-detail-panel__mono">{{ formatFileSize(catalog.currentBook.file_size) }}</span>
            </div>
            <div v-if="catalog.currentBook.lang" class="book-detail-panel__meta-item">
              <span class="book-detail-panel__meta-label">Язык: </span>
              {{ catalog.currentBook.lang }}
            </div>
          </div>

          <div class="book-detail-panel__actions">
            <button
              v-if="catalog.currentBook.format === 'fb2'"
              class="book-detail-panel__btn book-detail-panel__btn--primary"
              @click="readBook"
            >
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M2 3h6a4 4 0 0 1 4 4v14a3 3 0 0 0-3-3H2z" />
                <path d="M22 3h-6a4 4 0 0 0-4 4v14a3 3 0 0 1 3-3h7z" />
              </svg>
              Читать
            </button>
            <button
              class="book-detail-panel__btn book-detail-panel__btn--secondary"
              @click="downloadCurrentBook"
            >
              <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                <polyline points="7 10 12 15 17 10" />
                <line x1="12" y1="15" x2="12" y2="3" />
              </svg>
              Скачать <span class="book-detail-panel__mono" style="font-size: 11px; opacity: 0.7">({{ catalog.currentBook.format }})</span>
            </button>
          </div>

          <div class="book-detail-panel__annotation">
            <span class="book-detail-panel__annotation-label">Аннотация</span>
            <p v-if="catalog.currentBook.description" class="book-detail-panel__annotation-text">
              {{ catalog.currentBook.description }}
            </p>
            <p v-else class="book-detail-panel__annotation-text book-detail-panel__annotation-text--empty">
              Аннотация отсутствует
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useCatalogStore } from '@/stores/catalog'
import { downloadBook } from '@/api/books'

const catalog = useCatalogStore()
const router = useRouter()

function formatAuthors(authors: Array<{ id: number; name: string }>): string {
  if (!authors || authors.length === 0) return '—'
  return authors.map((a) => a.name).join(', ')
}

function formatGenres(genres: Array<{ id: number; name: string }>): string {
  if (!genres || genres.length === 0) return '—'
  return genres.map((g) => g.name).join(', ')
}

function formatFileSize(bytes?: number): string {
  if (!bytes) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function readBook() {
  if (catalog.currentBook) {
    router.push(`/books/${catalog.currentBook.id}/read`)
  }
}

function downloadCurrentBook() {
  if (catalog.currentBook) {
    downloadBook(catalog.currentBook.id)
  }
}
</script>

<style scoped>
.book-detail-panel {
  height: 100%;
  overflow-y: auto;
  padding: 16px;
  background: rgb(var(--v-theme-surface));
}

.book-detail-panel__empty,
.book-detail-panel__loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
}

.book-detail-panel__empty-text {
  font-size: 13px;
  margin-top: 8px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
}

.book-detail-panel__top {
  display: flex;
  gap: 20px;
  align-items: flex-start;
}

.book-detail-panel__cover {
  width: 80px;
  height: 110px;
  background: rgb(var(--v-theme-surface-variant));
  border-radius: 4px;
  border: 1px solid rgb(var(--v-theme-surface-variant));
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  font-size: 28px;
}

.book-detail-panel__info {
  flex: 1;
  min-width: 0;
}

.book-detail-panel__title {
  font-size: 18px;
  font-weight: 700;
  color: rgb(var(--v-theme-on-surface));
  margin-bottom: 4px;
  line-height: 1.3;
}

.book-detail-panel__author {
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.6;
  margin-bottom: 10px;
}

.book-detail-panel__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 20px;
  font-size: 12px;
  margin-bottom: 12px;
  color: rgb(var(--v-theme-on-surface));
}

.book-detail-panel__meta-label {
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
}

.book-detail-panel__meta-item {
  white-space: nowrap;
}

.book-detail-panel__mono {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}

.book-detail-panel__actions {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.book-detail-panel__btn {
  display: flex;
  align-items: center;
  gap: 7px;
  padding: 8px 18px;
  border-radius: 5px;
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 600;
  transition: all 0.15s;
  border: none;
  white-space: nowrap;
}

.book-detail-panel__btn--primary {
  background: rgb(var(--v-theme-primary));
  color: #1a1d23;
}

.book-detail-panel__btn--primary:hover {
  filter: brightness(1.1);
  box-shadow: 0 2px 12px rgba(var(--v-theme-primary), 0.25);
}

.book-detail-panel__btn--secondary {
  background: transparent;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.6;
  border: 1px solid rgb(var(--v-theme-surface-variant));
}

.book-detail-panel__btn--secondary:hover {
  opacity: 1;
  background: rgb(var(--v-theme-table-row-hover));
}

.book-detail-panel__annotation {
  border-top: 1px solid rgb(var(--v-theme-surface-variant));
  padding-top: 10px;
}

.book-detail-panel__annotation-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 600;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  display: block;
  margin-bottom: 4px;
}

.book-detail-panel__annotation-text {
  font-size: 13px;
  line-height: 1.65;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.6;
}

.book-detail-panel__annotation-text--empty {
  font-style: italic;
  opacity: 0.35;
}
</style>
