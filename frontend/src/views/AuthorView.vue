<template>
  <v-container>
    <v-btn variant="text" prepend-icon="mdi-arrow-left" @click="$router.back()">Назад</v-btn>

    <v-progress-linear v-if="loading" indeterminate color="primary" class="my-4" />

    <template v-if="author">
      <h1 class="text-h4 my-4">{{ author.name }}</h1>
      <p class="text-body-1 mb-4">Книг: {{ author.books_count }}</p>

      <v-row>
        <v-col v-for="book in author.books" :key="book.id" cols="12" sm="6" lg="4">
          <BookCard :book="book" :progress="progressMap[book.id] ?? 0" />
        </v-col>
      </v-row>
    </template>
  </v-container>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getAuthor, type AuthorDetail } from '@/api/books'
import { getAllReadingProgress } from '@/api/reader'
import BookCard from '@/components/common/BookCard.vue'

const route = useRoute()
const author = ref<AuthorDetail | null>(null)
const loading = ref(false)
const progressMap = ref<Record<number, number>>({})

watch(
  () => route.params.id,
  async (newId) => {
    const id = Number(newId)
    if (!id) return
    loading.value = true
    try {
      author.value = await getAuthor(id)
      progressMap.value = await getAllReadingProgress()
    } catch {
      // Author load may fail; progress is optional
    } finally {
      loading.value = false
    }
  },
  { immediate: true },
)
</script>
