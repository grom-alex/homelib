<template>
  <div class="status-bar">
    <span class="status-bar__context text-caption">
      {{ statusText }}
    </span>
    <span class="status-bar__count text-caption">
      <template v-if="catalog.total > 0">
        Показано книг: {{ catalog.books.length }} из {{ catalog.total }}
      </template>
    </span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useCatalogStore } from '@/stores/catalog'

const catalog = useCatalogStore()

const statusText = computed(() => {
  if (!catalog.navigationFilter) return 'Готов'

  const label = catalog.navigationFilter.label
  switch (catalog.navigationFilter.type) {
    case 'author':
      return `Автор: ${label || catalog.navigationFilter.id}`
    case 'series':
      return `Серия: ${label || catalog.navigationFilter.id}`
    case 'genre':
      return `Жанр: ${label || catalog.navigationFilter.id}`
    case 'search':
      return `Поиск: ${label || ''}`
    default:
      return 'Готов'
  }
})
</script>

<style scoped>
.status-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 24px;
  padding: 0 12px;
  background: rgb(var(--v-theme-status-bar));
  border-top: 1px solid rgb(var(--v-theme-surface-variant));
  flex-shrink: 0;
}

.status-bar__context,
.status-bar__count {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
