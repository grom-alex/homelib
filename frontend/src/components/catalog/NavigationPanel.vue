<template>
  <div class="navigation-panel">
    <AuthorsTab v-if="catalog.activeTab === 'authors'" />
    <SeriesTab v-else-if="catalog.activeTab === 'series'" />
    <GenresTab v-else-if="catalog.activeTab === 'genres'" />
    <SearchTab v-else-if="catalog.activeTab === 'search'" />
  </div>
</template>

<script setup lang="ts">
import { useCatalogStore } from '@/stores/catalog'
import AuthorsTab from '@/components/catalog/AuthorsTab.vue'
import SeriesTab from '@/components/catalog/SeriesTab.vue'
import GenresTab from '@/components/catalog/GenresTab.vue'
import SearchTab from '@/components/catalog/SearchTab.vue'

const catalog = useCatalogStore()
</script>

<style scoped>
.navigation-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: rgb(var(--v-theme-surface));
  overflow: hidden;
}
</style>
