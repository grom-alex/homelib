<template>
  <footer
    v-if="store.settings.showProgress"
    class="reader-footer"
  >
    <span class="reader-footer-page">
      {{ store.bookCurrentPage }} / {{ store.bookTotalPages }}
    </span>

    <div
      ref="progressBarRef"
      class="progress-bar"
      @dblclick="handleProgressDblClick"
      @touchend="handleProgressTouchEnd"
    >
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

const emit = defineEmits<{
  navigateToProgress: [percent: number]
}>()

const store = useReaderStore()
const progressBarRef = ref<HTMLElement | null>(null)
const currentTime = ref('')
let clockInterval: ReturnType<typeof setInterval> | null = null

let lastTapTime = 0

function getProgressPercent(clientX: number): number {
  const el = progressBarRef.value
  if (!el) return 0
  const rect = el.getBoundingClientRect()
  return Math.max(0, Math.min(100, ((clientX - rect.left) / rect.width) * 100))
}

function handleProgressDblClick(e: MouseEvent) {
  emit('navigateToProgress', getProgressPercent(e.clientX))
}

function handleProgressTouchEnd(e: TouchEvent) {
  const now = Date.now()
  if (now - lastTapTime < 300) {
    const touch = e.changedTouches[0]
    emit('navigateToProgress', getProgressPercent(touch.clientX))
  }
  lastTapTime = now
}

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

.progress-bar {
  cursor: pointer;
}
</style>
