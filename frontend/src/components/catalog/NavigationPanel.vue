<template>
  <div class="navigation-panel">
    <div class="navigation-panel__header">
      <span class="text-subtitle-2">{{ tabLabel }}</span>
    </div>
    <div class="navigation-panel__content">
      <AuthorsTab v-if="catalog.activeTab === 'authors'" />
      <SeriesTab v-else-if="catalog.activeTab === 'series'" />
      <GenresTab v-else-if="catalog.activeTab === 'genres'" />
      <SearchTab v-else-if="catalog.activeTab === 'search'" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCatalogStore } from '@/stores/catalog'
import AuthorsTab from '@/components/catalog/AuthorsTab.vue'
import SeriesTab from '@/components/catalog/SeriesTab.vue'
import GenresTab from '@/components/catalog/GenresTab.vue'
import SearchTab from '@/components/catalog/SearchTab.vue'

const catalog = useCatalogStore()

const tabLabels: Record<string, string> = {
  authors: 'Авторы',
  series: 'Серии',
  genres: 'Жанры',
  search: 'Поиск',
}

const tabLabel = computed(() => tabLabels[catalog.activeTab] || catalog.activeTab)
</script>

<style scoped>
.navigation-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: rgb(var(--v-theme-surface));
  overflow: hidden;
}

.navigation-panel__header {
  padding: 8px 12px;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  flex-shrink: 0;
}

.navigation-panel__content {
  flex: 1;
  overflow-y: auto;
}
</style>
