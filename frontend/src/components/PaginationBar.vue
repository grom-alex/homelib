<template>
  <div class="d-flex align-center justify-center ga-4 mt-4">
    <v-pagination
      v-model="currentPage"
      :length="totalPages"
      :total-visible="7"
      density="comfortable"
      @update:model-value="emit('update:page', $event)"
    />
    <v-select
      v-model="currentLimit"
      :items="[10, 20, 50, 100]"
      label="На странице"
      density="compact"
      hide-details
      style="max-width: 120px"
      @update:model-value="emit('update:limit', $event)"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  page: number
  totalPages: number
  limit: number
}>()

const emit = defineEmits<{
  (e: 'update:page', page: number): void
  (e: 'update:limit', limit: number): void
}>()

const currentPage = ref(props.page)
const currentLimit = ref(props.limit)

watch(() => props.page, (v) => { currentPage.value = v })
watch(() => props.limit, (v) => { currentLimit.value = v })
</script>
