<template>
  <div class="navigation-panel">
    <keep-alive>
      <component :is="activeTabComponent" />
    </keep-alive>
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

const tabComponents = {
  authors: AuthorsTab,
  series: SeriesTab,
  genres: GenresTab,
  search: SearchTab,
} as const

const activeTabComponent = computed(() => tabComponents[catalog.activeTab])
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
