<template>
  <v-container>
    <h1 class="text-h4 mb-4">Серии</h1>

    <v-text-field
      v-model="query"
      label="Поиск по сериям"
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
      <v-list-item v-for="s in seriesList" :key="s.id">
        <v-list-item-title>{{ s.name }}</v-list-item-title>
        <template #append>
          <v-chip size="small">{{ s.books_count }} книг</v-chip>
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
import { getSeries, type SeriesListItem } from '@/api/books'

const seriesList = ref<SeriesListItem[]>([])
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
    const result = await getSeries({ q: query.value || undefined, page: page.value, limit })
    seriesList.value = result.items
    total.value = result.total
    totalPages.value = Math.ceil(result.total / limit)
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>
