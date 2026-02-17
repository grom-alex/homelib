<template>
  <v-text-field
    v-model="query"
    label="Поиск по каталогу"
    prepend-inner-icon="mdi-magnify"
    clearable
    density="compact"
    hide-details
    variant="outlined"
    @update:model-value="onInput"
    @click:clear="query = ''; emit('search', '')"
  />
</template>

<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{
  (e: 'search', query: string): void
}>()

const query = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function onInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    emit('search', query.value)
  }, 300)
}
</script>
