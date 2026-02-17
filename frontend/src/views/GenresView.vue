<template>
  <v-container>
    <h1 class="text-h4 mb-4">Жанры</h1>

    <v-alert v-if="error" type="error" class="mb-4" closable @click:close="error = ''">
      {{ error }}
    </v-alert>

    <v-progress-linear v-if="loading" indeterminate color="primary" class="mb-4" />

    <v-treeview
      v-if="genres.length"
      :items="treeItems"
      item-title="title"
      item-value="id"
      open-on-click
      activatable
    />

    <div v-if="!loading && genres.length === 0" class="text-center pa-8 text-grey">
      Жанры не найдены
    </div>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getGenres, type GenreTreeItem } from '@/api/books'

const genres = ref<GenreTreeItem[]>([])
const loading = ref(false)
const error = ref('')

interface TreeItem {
  id: number
  title: string
  children?: TreeItem[]
}

function mapGenreToTree(g: GenreTreeItem): TreeItem {
  return {
    id: g.id,
    title: `${g.name} (${g.books_count})`,
    children: g.children?.map(mapGenreToTree),
  }
}

const treeItems = computed(() => genres.value.map(mapGenreToTree))

onMounted(async () => {
  loading.value = true
  error.value = ''
  try {
    genres.value = await getGenres()
  } catch {
    error.value = 'Не удалось загрузить список жанров'
  } finally {
    loading.value = false
  }
})
</script>
