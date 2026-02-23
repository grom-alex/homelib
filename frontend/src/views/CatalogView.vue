<template>
  <v-container>
    <v-row>
      <v-col cols="12" md="3">
        <BookFilters @update="onFiltersUpdate" />
      </v-col>
      <v-col cols="12" md="9">
        <v-progress-linear v-if="catalog.loading" indeterminate color="primary" class="mb-4" />
        <v-alert v-if="catalog.error" type="error" class="mb-4">{{ catalog.error }}</v-alert>

        <div v-if="!catalog.loading && catalog.books.length === 0" class="text-center pa-8">
          <v-icon size="64" color="grey">mdi-book-open-blank-variant</v-icon>
          <p class="text-h6 mt-4 text-grey">Книги не найдены</p>
        </div>

        <v-row v-else>
          <v-col v-for="book in catalog.books" :key="book.id" cols="12" sm="6" lg="4">
            <BookCard :book="book" :progress="progressMap[book.id] ?? 0" />
          </v-col>
        </v-row>

        <PaginationBar
          v-if="catalog.total > 0"
          :page="catalog.filters.page || 1"
          :total-pages="catalog.totalPages"
          :limit="catalog.filters.limit || 20"
          @update:page="catalog.setPage($event)"
          @update:limit="catalog.updateFilters({ limit: $event })"
        />
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useCatalogStore } from '@/stores/catalog'
import { getAllReadingProgress } from '@/api/reader'
import type { BookFilters as BookFiltersType } from '@/api/books'
import BookCard from '@/components/common/BookCard.vue'
import BookFilters from '@/components/common/BookFilters.vue'
import PaginationBar from '@/components/common/PaginationBar.vue'

const catalog = useCatalogStore()
const progressMap = ref<Record<number, number>>({})

async function loadProgress() {
  try {
    progressMap.value = await getAllReadingProgress()
  } catch {
    // Progress is optional — don't block the catalog
  }
}

function onFiltersUpdate(filters: Partial<BookFiltersType>) {
  catalog.updateFilters(filters)
}

onMounted(async () => {
  await catalog.fetchBooks()
  loadProgress()
})

// Reload progress when books change (e.g. page navigation)
watch(() => catalog.books, () => {
  loadProgress()
})
</script>
