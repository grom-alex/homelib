<template>
  <div class="genres-tab">
    <div class="genres-tab__header">Дерево жанров</div>

    <div v-if="loading" class="genres-tab__status">
      <v-progress-circular indeterminate size="24" />
    </div>

    <div v-else-if="genres.length === 0" class="genres-tab__status genres-tab__status--empty">
      Жанры не найдены
    </div>

    <div v-else class="genres-tab__list">
      <div v-for="group in groupedGenres" :key="group.name">
        <div
          class="genres-tab__group"
          :class="{ 'genres-tab__group--selected': false }"
          @click="toggleGroup(group.name)"
        >
          <svg
            width="12" height="12" viewBox="0 0 24 24"
            fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"
            class="genres-tab__chevron"
            :class="{ 'genres-tab__chevron--open': expandedGroups.has(group.name) }"
          >
            <polyline points="9 18 15 12 9 6" />
          </svg>
          <svg
            width="14" height="14" viewBox="0 0 24 24"
            :fill="expandedGroups.has(group.name) ? 'rgb(var(--v-theme-primary))' : 'none'"
            stroke="currentColor" stroke-width="2"
          >
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
          </svg>
          <span class="genres-tab__group-name">{{ group.name }}</span>
          <span class="genres-tab__group-count">{{ group.items.length }}</span>
        </div>

        <template v-if="expandedGroups.has(group.name)">
          <div
            v-for="genre in group.items"
            :key="genre.id"
            class="genres-tab__child"
            :class="{ 'genres-tab__child--selected': catalog.navigationFilter?.type === 'genre' && catalog.navigationFilter?.id === genre.id }"
            @click="selectGenre(genre.id, genre.name)"
          >
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
              <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
            </svg>
            <span>{{ genre.name }}</span>
          </div>
        </template>
      </div>
    </div>
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
    genres.value = (await getGenres()) ?? []
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
  overflow: hidden;
}

.genres-tab__header {
  padding: 8px 6px;
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  text-transform: uppercase;
  letter-spacing: 0.5px;
  font-weight: 600;
  flex-shrink: 0;
}

.genres-tab__status {
  padding: 20px;
  text-align: center;
}

.genres-tab__status--empty {
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
}

.genres-tab__list {
  flex: 1;
  overflow-y: auto;
  padding: 6px;
}

.genres-tab__group {
  cursor: pointer;
  padding: 3px 4px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  gap: 4px;
  user-select: none;
  color: rgb(var(--v-theme-on-surface));
}

.genres-tab__group:hover {
  background: rgb(var(--v-theme-table-row-hover));
}

.genres-tab__chevron {
  transform: rotate(0deg);
  transition: transform 0.15s;
  flex-shrink: 0;
}

.genres-tab__chevron--open {
  transform: rotate(90deg);
}

.genres-tab__group-name {
  font-weight: 500;
  font-size: 13px;
}

.genres-tab__group-count {
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  margin-left: auto;
}

.genres-tab__child {
  cursor: pointer;
  padding: 3px 4px 3px 22px;
  border-radius: 3px;
  display: flex;
  align-items: center;
  gap: 5px;
  user-select: none;
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
}

.genres-tab__child:hover {
  background: rgb(var(--v-theme-table-row-hover));
}

.genres-tab__child--selected {
  background: rgba(var(--v-theme-primary), 0.12);
  color: rgb(var(--v-theme-primary));
}
</style>
