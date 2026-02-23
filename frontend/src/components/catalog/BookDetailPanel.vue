<template>
  <div class="book-detail-panel">
    <div v-if="!catalog.selectedBookId" class="book-detail-panel__empty">
      <v-icon size="48" color="grey">mdi-book-information-variant</v-icon>
      <p class="text-body-1 mt-2 text-medium-emphasis">Выберите книгу для просмотра подробной информации</p>
    </div>

    <div v-else-if="catalog.bookLoading" class="book-detail-panel__loading">
      <v-progress-circular indeterminate size="32" />
    </div>

    <div v-else-if="catalog.currentBook" class="book-detail-panel__content">
      <h3 class="text-h6 mb-2">{{ catalog.currentBook.title }}</h3>

      <div class="book-detail-panel__meta">
        <div class="book-detail-panel__field">
          <span class="text-caption text-medium-emphasis">Автор</span>
          <span class="text-body-2">{{ formatAuthors(catalog.currentBook.authors) }}</span>
        </div>

        <div v-if="catalog.currentBook.series" class="book-detail-panel__field">
          <span class="text-caption text-medium-emphasis">Серия</span>
          <span class="text-body-2">
            {{ catalog.currentBook.series.name }}
            <template v-if="catalog.currentBook.series.num">#{{ catalog.currentBook.series.num }}</template>
          </span>
        </div>

        <div class="book-detail-panel__field">
          <span class="text-caption text-medium-emphasis">Жанр</span>
          <span class="text-body-2">{{ formatGenres(catalog.currentBook.genres) }}</span>
        </div>

        <div class="book-detail-panel__field-row">
          <div class="book-detail-panel__field">
            <span class="text-caption text-medium-emphasis">Формат</span>
            <span class="text-body-2">{{ catalog.currentBook.format?.toUpperCase() }}</span>
          </div>
          <div class="book-detail-panel__field">
            <span class="text-caption text-medium-emphasis">Размер</span>
            <span class="text-body-2 font-mono">{{ formatFileSize(catalog.currentBook.file_size) }}</span>
          </div>
          <div v-if="catalog.currentBook.year" class="book-detail-panel__field">
            <span class="text-caption text-medium-emphasis">Год</span>
            <span class="text-body-2">{{ catalog.currentBook.year }}</span>
          </div>
          <div v-if="catalog.currentBook.lang" class="book-detail-panel__field">
            <span class="text-caption text-medium-emphasis">Язык</span>
            <span class="text-body-2">{{ catalog.currentBook.lang }}</span>
          </div>
        </div>
      </div>

      <div v-if="catalog.currentBook.description" class="book-detail-panel__annotation mt-3">
        <span class="text-caption text-medium-emphasis">Аннотация</span>
        <p class="text-body-2 mt-1">{{ catalog.currentBook.description }}</p>
      </div>
      <div v-else class="mt-3">
        <span class="text-caption text-medium-emphasis">Аннотация</span>
        <p class="text-body-2 mt-1 text-medium-emphasis font-italic">Аннотация отсутствует</p>
      </div>

      <div class="book-detail-panel__actions mt-4">
        <v-btn
          v-if="catalog.currentBook.format === 'fb2'"
          color="primary"
          variant="flat"
          size="small"
          prepend-icon="mdi-book-open-page-variant"
          @click="readBook"
        >
          Читать
        </v-btn>
        <v-btn
          variant="outlined"
          size="small"
          prepend-icon="mdi-download"
          @click="downloadCurrentBook"
        >
          Скачать
        </v-btn>
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
  padding: 12px 16px;
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

.book-detail-panel__meta {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.book-detail-panel__field {
  display: flex;
  flex-direction: column;
}

.book-detail-panel__field-row {
  display: flex;
  gap: 16px;
  flex-wrap: wrap;
  margin-top: 4px;
}

.book-detail-panel__annotation {
  max-height: 200px;
  overflow-y: auto;
}

.book-detail-panel__actions {
  display: flex;
  gap: 8px;
}

.font-mono {
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
}
</style>
