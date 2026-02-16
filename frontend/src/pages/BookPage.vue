<template>
  <v-container>
    <v-progress-linear v-if="catalog.loading" indeterminate color="primary" />

    <v-alert v-if="catalog.error" type="error" class="mb-4">{{ catalog.error }}</v-alert>

    <template v-if="book">
      <v-row>
        <v-col cols="12">
          <v-btn variant="text" prepend-icon="mdi-arrow-left" @click="$router.back()">Назад</v-btn>
        </v-col>
      </v-row>
      <v-row>
        <v-col cols="12" md="8">
          <h1 class="text-h4 mb-2">{{ book.title }}</h1>
          <div class="mb-4">
            <span v-for="(author, i) in book.authors" :key="author.id">
              <router-link :to="`/authors/${author.id}`">{{ author.name }}</router-link>
              <span v-if="i < book.authors.length - 1">, </span>
            </span>
          </div>

          <div v-if="book.description" class="text-body-1 mb-4">{{ book.description }}</div>

          <div v-if="book.keywords?.length" class="mb-4">
            <v-chip v-for="kw in book.keywords" :key="kw" size="small" class="mr-1 mb-1" label>
              {{ kw }}
            </v-chip>
          </div>
        </v-col>
        <v-col cols="12" md="4">
          <v-card variant="outlined">
            <v-card-text>
              <v-list density="compact">
                <v-list-item>
                  <template #prepend><v-icon>mdi-file-document</v-icon></template>
                  <v-list-item-title>Формат: {{ book.format.toUpperCase() }}</v-list-item-title>
                </v-list-item>
                <v-list-item>
                  <template #prepend><v-icon>mdi-translate</v-icon></template>
                  <v-list-item-title>Язык: {{ book.lang }}</v-list-item-title>
                </v-list-item>
                <v-list-item v-if="book.year">
                  <template #prepend><v-icon>mdi-calendar</v-icon></template>
                  <v-list-item-title>Год: {{ book.year }}</v-list-item-title>
                </v-list-item>
                <v-list-item v-if="book.file_size">
                  <template #prepend><v-icon>mdi-harddisk</v-icon></template>
                  <v-list-item-title>Размер: {{ formatSize(book.file_size) }}</v-list-item-title>
                </v-list-item>
                <v-list-item v-if="book.series">
                  <template #prepend><v-icon>mdi-bookshelf</v-icon></template>
                  <v-list-item-title>
                    Серия: {{ book.series.name }}
                    <span v-if="book.series.num"> #{{ book.series.num }}</span>
                  </v-list-item-title>
                </v-list-item>
                <v-list-item v-if="book.collection">
                  <template #prepend><v-icon>mdi-folder</v-icon></template>
                  <v-list-item-title>Коллекция: {{ book.collection.name }}</v-list-item-title>
                </v-list-item>
              </v-list>
              <v-divider class="my-2" />
              <div class="d-flex flex-wrap ga-1 mb-3">
                <v-chip v-for="genre in book.genres" :key="genre.id" size="small" label>
                  {{ genre.name }}
                </v-chip>
              </div>
              <v-btn color="primary" block prepend-icon="mdi-download" @click="handleDownload">
                Скачать
              </v-btn>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useCatalogStore } from '@/stores/catalog'
import { downloadBook } from '@/services/books'

const route = useRoute()
const catalog = useCatalogStore()

const book = computed(() => catalog.currentBook)

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(0) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

async function handleDownload() {
  if (book.value) {
    await downloadBook(book.value.id)
  }
}

watch(
  () => route.params.id,
  (newId) => {
    const id = Number(newId)
    if (id) catalog.fetchBook(id)
  },
  { immediate: true },
)
</script>
