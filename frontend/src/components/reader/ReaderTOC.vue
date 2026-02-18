<template>
  <Teleport to="body">
    <template v-if="store.tocVisible">
      <div class="reader-toc-overlay" @click="store.tocVisible = false" />
      <nav class="reader-toc">
        <div class="reader-toc-title">Оглавление</div>
        <button
          v-for="entry in store.bookContent?.toc ?? []"
          :key="entry.id"
          class="reader-toc-item"
          :class="[
            `level-${entry.level}`,
            { active: entry.id === store.currentChapterId },
          ]"
          @click="selectChapter(entry.id)"
        >
          {{ entry.title }}
        </button>
      </nav>
    </template>
  </Teleport>
</template>

<script setup lang="ts">
import { useReaderStore } from '@/stores/reader'

const emit = defineEmits<{
  navigate: [chapterId: string]
}>()

const store = useReaderStore()

function selectChapter(chapterId: string) {
  emit('navigate', chapterId)
}
</script>
