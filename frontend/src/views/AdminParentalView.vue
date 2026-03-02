<template>
  <!-- PIN gate: require PIN before showing settings -->
  <v-container v-if="showPinGate" class="d-flex align-center justify-center" style="min-height: 400px">
    <PinUnlockDialog
      v-model="showPinDialog"
      @cancel="onPinCancel"
    />
    <v-card variant="outlined" max-width="400" class="text-center pa-6">
      <v-icon size="64" color="grey" class="mb-4">mdi-shield-lock</v-icon>
      <h2 class="text-h5 mb-2">Родительский контроль</h2>
      <p class="text-body-2 text-grey mb-4">Введите PIN-код для доступа к настройкам</p>
      <v-btn color="primary" @click="showPinDialog = true">
        Ввести PIN
      </v-btn>
    </v-card>
  </v-container>

  <!-- Loading state while checking parental status -->
  <v-container v-else-if="!parentalLoaded" class="d-flex justify-center pa-8">
    <v-progress-circular indeterminate />
  </v-container>

  <!-- Main settings content -->
  <v-container v-else>
    <h1 class="text-h4 mb-4">Родительский контроль</h1>

    <v-alert v-if="error" type="error" class="mb-4" closable @click:close="error = ''">
      {{ error }}
    </v-alert>

    <v-alert v-if="successMsg" type="success" class="mb-4" closable @click:close="successMsg = ''">
      {{ successMsg }}
    </v-alert>

    <!-- PIN Section -->
    <v-card variant="outlined" class="mb-4">
      <v-card-title>PIN-код</v-card-title>
      <v-card-text>
        <v-chip :color="status?.pin_set ? 'success' : 'grey'" class="mb-3">
          {{ status?.pin_set ? 'PIN установлен' : 'PIN не установлен' }}
        </v-chip>

        <div class="d-flex align-center ga-3">
          <v-text-field
            v-model="newPin"
            label="Новый PIN (4-6 цифр)"
            type="password"
            maxlength="6"
            counter
            density="compact"
            style="max-width: 240px"
            hide-details
          />
          <v-btn
            color="primary"
            :loading="savingPin"
            :disabled="newPin.length < 4 || newPin.length > 6"
            @click="handleSetPin"
          >
            {{ status?.pin_set ? 'Изменить' : 'Установить' }}
          </v-btn>
          <v-btn
            v-if="status?.pin_set"
            color="error"
            variant="outlined"
            :loading="removingPin"
            @click="handleRemovePin"
          >
            Удалить
          </v-btn>
        </div>
      </v-card-text>
    </v-card>

    <!-- Restricted Genres Section -->
    <v-card variant="outlined" class="mb-4">
      <v-card-title>Ограниченные жанры</v-card-title>
      <v-card-subtitle>
        Книги этих жанров будут скрыты от пользователей без доступа к контенту 18+
      </v-card-subtitle>
      <v-card-text>
        <div v-if="loadingGenres" class="d-flex justify-center pa-4">
          <v-progress-circular indeterminate />
        </div>
        <template v-else>
          <v-treeview
            v-if="allGenres.length > 0"
            :items="allGenres"
            item-value="code"
            item-title="name"
            item-children="children"
            selectable
            select-strategy="independent"
            density="compact"
            slim
            open-on-click
            :model-value="selectedCodes"
            @update:model-value="selectedCodes = $event as string[]"
          />
          <v-alert v-else type="info" variant="tonal">
            Дерево жанров не загружено. Выполните загрузку жанров в разделе «Импорт».
          </v-alert>

          <v-btn
            v-if="allGenres.length > 0"
            color="primary"
            class="mt-3"
            :loading="savingGenres"
            @click="handleSaveGenres"
          >
            Сохранить
          </v-btn>
        </template>
      </v-card-text>
    </v-card>

    <!-- Users Section -->
    <v-card variant="outlined">
      <v-card-title>Пользователи</v-card-title>
      <v-card-subtitle>Управление доступом к контенту 18+ для каждого пользователя</v-card-subtitle>
      <v-card-text>
        <div v-if="loadingUsers" class="d-flex justify-center pa-4">
          <v-progress-circular indeterminate />
        </div>
        <v-table v-else-if="users.length > 0" density="compact">
          <thead>
            <tr>
              <th>Пользователь</th>
              <th>Роль</th>
              <th class="text-center">Контент 18+</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.user_id">
              <td>{{ u.display_name || u.username }}</td>
              <td>{{ u.role }}</td>
              <td class="text-center">
                <v-switch
                  :model-value="u.adult_content_enabled"
                  color="primary"
                  density="compact"
                  hide-details
                  @update:model-value="handleToggleUser(u, $event as boolean)"
                />
              </td>
            </tr>
          </tbody>
        </v-table>
        <v-alert v-else type="info" variant="tonal">
          Нет пользователей
        </v-alert>
      </v-card-text>
    </v-card>
  </v-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import * as parentalApi from '@/api/parental'
import type { AdminParentalStatus, UserAdultStatus } from '@/api/parental'
import { getGenres, type GenreTreeItem } from '@/api/books'
import { useParentalStore } from '@/stores/parental'
import PinUnlockDialog from '@/components/common/PinUnlockDialog.vue'

const router = useRouter()
const parentalStore = useParentalStore()

const showPinDialog = ref(false)
const parentalLoaded = ref(false)

const showPinGate = computed(() =>
  parentalLoaded.value && parentalStore.pinSet && !parentalStore.adultContentEnabled
)

const status = ref<AdminParentalStatus | null>(null)
const newPin = ref('')
const savingPin = ref(false)
const removingPin = ref(false)
const error = ref('')
const successMsg = ref('')

const allGenres = ref<GenreTreeItem[]>([])
const selectedCodes = ref<string[]>([])
const loadingGenres = ref(false)
const savingGenres = ref(false)

const users = ref<UserAdultStatus[]>([])
const loadingUsers = ref(false)

function showSuccess(msg: string) {
  successMsg.value = msg
  setTimeout(() => { successMsg.value = '' }, 3000)
}

function onPinCancel() {
  showPinDialog.value = false
  router.back()
}

// When PIN gate disappears (after unlock), load admin data.
// Using a watcher is more reliable than the @unlocked emit because
// unlock() sets adultContentEnabled=true reactively, which may cause
// Vue to unmount the PinUnlockDialog before the emit reaches the parent.
const dataLoaded = ref(false)
watch(showPinGate, (gate, prevGate) => {
  if (prevGate && !gate && !dataLoaded.value) {
    loadAllData()
  }
})

function loadAllData() {
  dataLoaded.value = true
  loadStatus()
  loadGenres()
  loadUsers()
}

async function loadStatus() {
  try {
    status.value = await parentalApi.getAdminParentalStatus()
    selectedCodes.value = status.value.restricted_genre_codes || []
  } catch {
    error.value = 'Не удалось загрузить статус'
  }
}

async function loadGenres() {
  loadingGenres.value = true
  try {
    allGenres.value = await getGenres()
  } catch {
    // Genres not loaded yet
  } finally {
    loadingGenres.value = false
  }
}

async function loadUsers() {
  loadingUsers.value = true
  try {
    users.value = await parentalApi.listUsersAdultStatus()
  } catch {
    error.value = 'Не удалось загрузить список пользователей'
  } finally {
    loadingUsers.value = false
  }
}

async function handleSetPin() {
  savingPin.value = true
  error.value = ''
  try {
    await parentalApi.setParentalPin(newPin.value)
    newPin.value = ''
    if (status.value) status.value.pin_set = true
    showSuccess('PIN установлен')
  } catch {
    error.value = 'Не удалось установить PIN'
  } finally {
    savingPin.value = false
  }
}

async function handleRemovePin() {
  removingPin.value = true
  error.value = ''
  try {
    await parentalApi.removeParentalPin()
    if (status.value) status.value.pin_set = false
    showSuccess('PIN удалён')
  } catch {
    error.value = 'Не удалось удалить PIN'
  } finally {
    removingPin.value = false
  }
}

async function handleSaveGenres() {
  savingGenres.value = true
  error.value = ''
  try {
    await parentalApi.updateRestrictedGenres(selectedCodes.value)
    showSuccess('Список ограниченных жанров сохранён')
  } catch {
    error.value = 'Не удалось сохранить список жанров'
  } finally {
    savingGenres.value = false
  }
}

async function handleToggleUser(user: UserAdultStatus, enabled: boolean) {
  try {
    await parentalApi.setUserAdultContent(user.user_id, enabled)
    user.adult_content_enabled = enabled
  } catch {
    error.value = `Не удалось изменить доступ для ${user.display_name || user.username}`
  }
}

onMounted(async () => {
  if (!parentalStore.loaded) {
    await parentalStore.loadStatus()
  }
  parentalLoaded.value = true

  // If no PIN set or already unlocked, load data immediately
  if (!parentalStore.pinSet || parentalStore.adultContentEnabled) {
    loadAllData()
  }
})
</script>
