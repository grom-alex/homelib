<template>
  <header class="reader-header" :class="{ hidden: !store.uiVisible }">
    <button class="reader-header-btn" title="Назад в каталог" @click="$router.back()">
      ← Каталог
    </button>

    <div class="reader-header-title">
      <div class="reader-header-book">{{ store.bookContent?.metadata.title }}</div>
      <div class="reader-header-chapter">{{ store.currentChapterContent?.title }}</div>
    </div>

    <div class="reader-header-actions">
      <button class="reader-header-btn" title="Оглавление (T)" @click="store.toggleTOC()">
        ☰
      </button>
      <button class="reader-header-btn" title="Настройки" @click="store.toggleSettings()">
        ⚙
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useReaderStore } from '@/stores/reader'

const store = useReaderStore()
</script>

<style scoped>
.reader-header-title {
  flex: 1;
  text-align: center;
  overflow: hidden;
  min-width: 0;
}

.reader-header-book {
  font-weight: 600;
  font-size: 0.9em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.reader-header-chapter {
  font-size: 0.8em;
  opacity: 0.7;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.reader-header-actions {
  display: flex;
  gap: 8px;
}

.reader-header-btn {
  background: none;
  border: none;
  color: inherit;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 1em;
}

.reader-header-btn:hover {
  background: var(--reader-selection);
}
</style>
