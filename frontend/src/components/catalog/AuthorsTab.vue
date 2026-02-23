<template>
  <div class="authors-tab">
    <div class="authors-tab__search">
      <v-text-field
        v-model="searchQuery"
        placeholder="Поиск автора..."
        density="compact"
        variant="outlined"
        hide-details
        clearable
        prepend-inner-icon="mdi-magnify"
        @update:model-value="onSearchInput"
      />
    </div>

    <div v-if="loading && authors.length === 0" class="pa-4 text-center">
      <v-progress-circular indeterminate size="24" />
    </div>

    <div v-else-if="!loading && authors.length === 0" class="pa-4 text-center text-body-2 text-medium-emphasis">
      Ничего не найдено
    </div>

    <v-list v-else density="compact" class="authors-tab__list">
      <v-list-item
        v-for="author in authors"
        :key="author.id"
        :active="catalog.navigationFilter?.type === 'author' && catalog.navigationFilter?.id === author.id"
        @click="selectAuthor(author.id, author.name)"
      >
        <v-list-item-title class="text-body-2">{{ author.name }}</v-list-item-title>
        <template #append>
          <v-badge :content="String(author.books_count)" color="primary" inline />
        </template>
      </v-list-item>

      <div v-if="hasMore" class="text-center pa-2">
        <v-btn
          variant="text"
          size="small"
          :loading="loading"
          @click="loadMore"
        >
          Загрузить ещё
        </v-btn>
      </div>
    </v-list>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useCatalogStore } from '@/stores/catalog'
import { getAuthors, type AuthorListItem } from '@/api/books'

const catalog = useCatalogStore()

const searchQuery = ref('')
const authors = ref<AuthorListItem[]>([])
const loading = ref(false)
const page = ref(1)
const hasMore = ref(false)
const limit = 50

let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    page.value = 1
    authors.value = []
    fetchAuthors()
  }, 300)
}

async function fetchAuthors() {
  loading.value = true
  try {
    const result = await getAuthors({
      q: searchQuery.value || undefined,
      page: page.value,
      limit,
    })
    if (page.value === 1) {
      authors.value = result.items
    } else {
      authors.value = [...authors.value, ...result.items]
    }
    hasMore.value = authors.value.length < result.total
  } catch {
    // Ошибка загрузки авторов
  } finally {
    loading.value = false
  }
}

function loadMore() {
  page.value++
  fetchAuthors()
}

function selectAuthor(authorId: number, name: string) {
  catalog.selectNavItem('author', authorId, undefined, name)
}

onMounted(() => {
  fetchAuthors()
})
</script>

<style scoped>
.authors-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.authors-tab__search {
  padding: 8px;
  flex-shrink: 0;
}

.authors-tab__list {
  flex: 1;
  overflow-y: auto;
}
</style>
