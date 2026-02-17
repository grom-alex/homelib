<template>
  <v-container>
    <h1 class="text-h4 mb-4">Импорт библиотеки</h1>

    <v-card variant="outlined" class="mb-4">
      <v-card-text>
        <p class="mb-4">Запустить импорт INPX-файла из директории библиотеки.</p>
        <v-btn
          color="primary"
          :loading="importing"
          :disabled="status?.status === 'running'"
          prepend-icon="mdi-import"
          @click="handleImport"
        >
          Запустить импорт
        </v-btn>
      </v-card-text>
    </v-card>

    <v-alert v-if="error" type="error" class="mb-4" closable @click:close="error = ''">
      {{ error }}
    </v-alert>

    <v-card v-if="status" variant="outlined">
      <v-card-title>Статус импорта</v-card-title>
      <v-card-text>
        <v-chip
          :color="statusColor"
          class="mb-3"
        >
          {{ statusText }}
        </v-chip>

        <div v-if="status.status === 'running'" class="mb-3">
          <v-progress-linear
            v-if="(status.total_batches ?? 0) > 0"
            :model-value="((status.processed_batch ?? 0) / (status.total_batches ?? 1)) * 100"
            color="primary"
            height="20"
            rounded
          >
            <template #default>
              <strong>{{ status.processed_batch ?? 0 }} / {{ status.total_batches ?? 0 }}</strong>
            </template>
          </v-progress-linear>
          <v-progress-linear
            v-else
            indeterminate
            color="primary"
          />
          <div v-if="(status.total_records ?? 0) > 0" class="text-caption mt-1">
            Записей в INPX: {{ (status.total_records ?? 0).toLocaleString('ru-RU') }}
          </div>
        </div>

        <div v-if="status.started_at" class="mb-2">
          <strong>Начало:</strong> {{ formatDate(status.started_at) }}
        </div>
        <div v-if="status.finished_at" class="mb-2">
          <strong>Завершение:</strong> {{ formatDate(status.finished_at) }}
        </div>

        <v-table v-if="status.stats" density="compact" class="mt-3">
          <thead>
            <tr>
              <th>Параметр</th>
              <th class="text-right">Значение</th>
            </tr>
          </thead>
          <tbody>
            <tr><td>Книг добавлено</td><td class="text-right">{{ status.stats.books_added }}</td></tr>
            <tr><td>Книг обновлено</td><td class="text-right">{{ status.stats.books_updated }}</td></tr>
            <tr><td>Книг удалено</td><td class="text-right">{{ status.stats.books_deleted }}</td></tr>
            <tr><td>Авторов добавлено</td><td class="text-right">{{ status.stats.authors_added }}</td></tr>
            <tr><td>Жанров добавлено</td><td class="text-right">{{ status.stats.genres_added }}</td></tr>
            <tr><td>Серий добавлено</td><td class="text-right">{{ status.stats.series_added }}</td></tr>
            <tr><td>Ошибок</td><td class="text-right">{{ status.stats.errors }}</td></tr>
            <tr><td>Длительность</td><td class="text-right">{{ (status.stats.duration_ms / 1000).toFixed(1) }} сек</td></tr>
          </tbody>
        </v-table>

        <v-alert v-if="status.error" type="error" class="mt-3">
          {{ status.error }}
        </v-alert>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { startImport, getImportStatus, type ImportStatus } from '@/api/admin'

const status = ref<ImportStatus | null>(null)
const importing = ref(false)
const error = ref('')
let pollTimer: ReturnType<typeof setInterval> | null = null

const statusColor = computed(() => {
  switch (status.value?.status) {
    case 'running': return 'info'
    case 'completed': return 'success'
    case 'failed': return 'error'
    default: return 'grey'
  }
})

const statusText = computed(() => {
  switch (status.value?.status) {
    case 'idle': return 'Ожидание'
    case 'running': return 'Выполняется'
    case 'completed': return 'Завершён'
    case 'failed': return 'Ошибка'
    default: return 'Неизвестно'
  }
})

function formatDate(iso: string): string {
  return new Date(iso).toLocaleString('ru-RU')
}

async function handleImport() {
  importing.value = true
  error.value = ''
  try {
    status.value = await startImport()
    startPolling()
  } catch (e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Ошибка запуска импорта'
    } else {
      error.value = 'Ошибка запуска импорта'
    }
  } finally {
    importing.value = false
  }
}

function startPolling() {
  stopPolling()
  pollTimer = setInterval(async () => {
    try {
      status.value = await getImportStatus()
      if (status.value.status !== 'running') {
        stopPolling()
      }
    } catch {
      stopPolling()
    }
  }, 2000)
}

function stopPolling() {
  if (pollTimer) {
    clearInterval(pollTimer)
    pollTimer = null
  }
}

onMounted(async () => {
  try {
    status.value = await getImportStatus()
    if (status.value.status === 'running') {
      startPolling()
    }
  } catch {
    // No previous import
  }
})

onUnmounted(stopPolling)
</script>
