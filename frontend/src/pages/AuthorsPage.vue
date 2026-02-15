<template>
  <v-container>
    <h1 class="text-h4 mb-4">Авторы</h1>

    <v-text-field
      v-model="query"
      label="Поиск по авторам"
      prepend-inner-icon="mdi-magnify"
      clearable
      density="compact"
      hide-details
      variant="outlined"
      class="mb-4"
      @update:model-value="onSearch"
      @click:clear="query = ''; fetchData()"
    />

    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

    <v-list>
      <v-list-item
        v-for="author in authors"
        :key="author.id"
        :to="`/authors/${author.id}`"
      >
        <v-list-item-title>{{ author.name }}</v-list-item-title>
        <template #append>
          <v-chip size="small">{{ author.books_count }} книг</v-chip>
        </template>
      </v-list-item>
    </v-list>

    <v-pagination
      v-if="totalPages > 1"
      v-model="page"
      :length="totalPages"
      class="mt-4"
      @update:model-value="fetchData"
    />
  </v-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAuthors, type AuthorListItem } from '@/services/books'

const authors = ref<AuthorListItem[]>([])
const loading = ref(false)
const query = ref('')
const page = ref(1)
const total = ref(0)
const limit = 50
const totalPages = ref(0)

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onSearch() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    page.value = 1
    fetchData()
  }, 300)
}

async function fetchData() {
  loading.value = true
  try {
    const result = await getAuthors({ q: query.value || undefined, page: page.value, limit })
    authors.value = result.items
    total.value = result.total
    totalPages.value = Math.ceil(result.total / limit)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>
