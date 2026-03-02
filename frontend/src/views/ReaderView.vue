<template>
  <div v-if="store.loading && !store.bookContent" class="reader-loading">
    <v-progress-circular indeterminate size="48" />
    <p>Загрузка книги…</p>
  </div>

  <div v-else-if="store.error" class="reader-error">
    <v-icon size="64" :color="isContentRestricted ? 'warning' : 'error'">
      {{ isContentRestricted ? 'mdi-lock' : 'mdi-alert-circle-outline' }}
    </v-icon>
    <h2>{{ errorTitle }}</h2>
    <p>{{ isContentRestricted ? 'Доступ к этой книге ограничен родительским контролем' : store.error }}</p>
    <v-btn color="primary" @click="$router.back()">Назад в каталог</v-btn>
  </div>

  <BookReader v-else-if="store.bookContent" :book-id="bookId" :key="bookId" />
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useReaderStore } from '@/stores/reader'
import { useBookContent } from '@/composables/useBookContent'
import BookReader from '@/components/reader/BookReader.vue'
import '@/assets/styles/reader-themes.css'

const route = useRoute()
const store = useReaderStore()
const { loadBookContent } = useBookContent()

const bookId = computed(() => Number(route.params.id))

const isContentRestricted = computed(() => {
  const err = store.error ?? ''
  return err.includes('content_restricted') || err.includes('Контент ограничен')
})

const errorTitle = computed(() => {
  if (isContentRestricted.value) return 'Контент ограничен'
  const err = store.error ?? ''
  if (err.includes('не найдена')) return 'Книга не найдена'
  if (err.includes('не поддерживается')) return 'Формат не поддерживается'
  if (err.includes('повреждён')) return 'Ошибка чтения'
  return 'Ошибка'
})

onMounted(() => {
  loadBookContent(bookId.value)
})

onUnmounted(() => {
  store.reset()
})
</script>

<style scoped>
.reader-loading {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: 16px;
}

.reader-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  gap: 16px;
  padding: 24px;
  text-align: center;
}
</style>
