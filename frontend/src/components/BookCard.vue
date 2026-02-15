<template>
  <v-card class="book-card" hover @click="$router.push(`/books/${book.id}`)">
    <v-card-title class="text-subtitle-1 font-weight-bold">
      {{ book.title }}
    </v-card-title>
    <v-card-subtitle>
      <span v-for="(author, i) in book.authors" :key="author.id">
        <router-link
          :to="`/authors/${author.id}`"
          class="text-decoration-none"
          @click.stop
        >{{ author.name }}</router-link>
        <span v-if="i < book.authors.length - 1">, </span>
      </span>
    </v-card-subtitle>
    <v-card-text>
      <div class="d-flex flex-wrap ga-1 mb-2">
        <v-chip v-for="genre in book.genres" :key="genre.id" size="x-small" label>
          {{ genre.name }}
        </v-chip>
      </div>
      <div class="d-flex align-center ga-2 text-body-2 text-medium-emphasis">
        <v-chip size="small" color="primary" label>{{ book.format.toUpperCase() }}</v-chip>
        <span v-if="book.lang">{{ book.lang }}</span>
        <span v-if="book.year">{{ book.year }}</span>
        <span v-if="book.file_size">{{ formatSize(book.file_size) }}</span>
        <v-chip v-if="book.series" size="small" variant="outlined" label>
          {{ book.series.name }}<span v-if="book.series.num"> #{{ book.series.num }}</span>
        </v-chip>
      </div>
    </v-card-text>
  </v-card>
</template>

<script setup lang="ts">
import type { BookListItem } from '@/services/books'

defineProps<{ book: BookListItem }>()

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(0) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}
</script>
