<template>
  <div class="genres-tab">
    <div v-if="loading" class="pa-4 text-center">
      <v-progress-circular indeterminate size="24" />
    </div>

    <div v-else-if="genres.length === 0" class="pa-4 text-center text-body-2 text-medium-emphasis">
      Жанры не найдены
    </div>

    <v-list v-else density="compact" class="genres-tab__list">
      <template v-for="group in groupedGenres" :key="group.name">
        <v-list-item
          class="genres-tab__group-header"
          @click="toggleGroup(group.name)"
        >
          <template #prepend>
            <v-icon size="16">
              {{ expandedGroups.has(group.name) ? 'mdi-chevron-down' : 'mdi-chevron-right' }}
            </v-icon>
          </template>
          <v-list-item-title class="text-body-2 font-weight-medium">
            {{ group.name }}
          </v-list-item-title>
        </v-list-item>

        <v-expand-transition>
          <div v-show="expandedGroups.has(group.name)">
            <v-list-item
              v-for="genre in group.items"
              :key="genre.id"
              :active="catalog.navigationFilter?.type === 'genre' && catalog.navigationFilter?.id === genre.id"
              class="pl-8"
              @click="selectGenre(genre.id, genre.name)"
            >
              <v-list-item-title class="text-body-2">{{ genre.name }}</v-list-item-title>
              <template #append>
                <span class="text-caption text-medium-emphasis">{{ genre.books_count }}</span>
              </template>
            </v-list-item>
          </div>
        </v-expand-transition>
      </template>
    </v-list>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useCatalogStore } from '@/stores/catalog'
import { getGenres, type GenreTreeItem } from '@/api/books'

const catalog = useCatalogStore()

const genres = ref<GenreTreeItem[]>([])
const loading = ref(false)
const expandedGroups = ref<Set<string>>(new Set())

interface GenreGroup {
  name: string
  items: GenreTreeItem[]
}

const groupedGenres = computed<GenreGroup[]>(() => {
  const groups = new Map<string, GenreTreeItem[]>()

  for (const genre of genres.value) {
    const groupName = genre.meta_group || 'Другое'
    if (!groups.has(groupName)) {
      groups.set(groupName, [])
    }
    groups.get(groupName)!.push(genre)

    if (genre.children) {
      for (const child of genre.children) {
        const childGroup = child.meta_group || groupName
        if (!groups.has(childGroup)) {
          groups.set(childGroup, [])
        }
        groups.get(childGroup)!.push(child)
      }
    }
  }

  return Array.from(groups.entries())
    .map(([name, items]) => ({ name, items }))
    .sort((a, b) => a.name.localeCompare(b.name, 'ru'))
})

function toggleGroup(name: string) {
  if (expandedGroups.value.has(name)) {
    expandedGroups.value.delete(name)
  } else {
    expandedGroups.value.add(name)
  }
  expandedGroups.value = new Set(expandedGroups.value)
}

function selectGenre(genreId: number, name: string) {
  catalog.selectNavItem('genre', genreId, undefined, name)
}

async function fetchGenres() {
  loading.value = true
  try {
    genres.value = await getGenres()
  } catch {
    // Ошибка загрузки жанров
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchGenres()
})
</script>

<style scoped>
.genres-tab {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow-y: auto;
}

.genres-tab__list {
  flex: 1;
}

.genres-tab__group-header {
  cursor: pointer;
}

.genres-tab__group-header:hover {
  background: rgb(var(--v-theme-table-row-hover));
}
</style>
