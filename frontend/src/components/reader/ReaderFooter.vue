<template>
  <footer
    v-if="store.settings.showProgress"
    class="reader-footer"
    :class="{ hidden: !store.uiVisible }"
  >
    <span class="reader-footer-page">
      {{ store.bookCurrentPage }} / {{ store.bookTotalPages }}
    </span>

    <div class="progress-bar">
      <div class="progress-bar-fill" :style="{ width: store.totalProgress + '%' }" />
    </div>

    <span class="reader-footer-info">
      <span>{{ store.totalProgress }}%</span>
      <span v-if="store.settings.showClock" class="reader-footer-clock">
        {{ currentTime }}
      </span>
    </span>
  </footer>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useReaderStore } from '@/stores/reader'

const store = useReaderStore()
const currentTime = ref('')
let clockInterval: ReturnType<typeof setInterval> | null = null

function updateClock() {
  const now = new Date()
  currentTime.value = now.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

onMounted(() => {
  updateClock()
  clockInterval = setInterval(updateClock, 30000)
})

onUnmounted(() => {
  if (clockInterval) clearInterval(clockInterval)
})
</script>

<style scoped>
.reader-footer-page {
  white-space: nowrap;
  min-width: 60px;
}

.reader-footer-info {
  display: flex;
  gap: 8px;
  white-space: nowrap;
  min-width: 60px;
  justify-content: flex-end;
}

.reader-footer-clock {
  opacity: 0.7;
}
</style>
