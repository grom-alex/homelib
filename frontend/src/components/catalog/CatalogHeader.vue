<template>
  <div class="catalog-header">
    <div class="catalog-header__logo">
      <v-icon size="20" class="mr-1">mdi-bookshelf</v-icon>
      <span class="text-subtitle-2 font-weight-bold">HomeLib</span>
    </div>

    <div class="catalog-header__tabs">
      <v-btn
        v-for="tab in tabs"
        :key="tab.value"
        :variant="catalog.activeTab === tab.value ? 'flat' : 'text'"
        :color="catalog.activeTab === tab.value ? 'primary' : undefined"
        size="small"
        @click="catalog.setActiveTab(tab.value)"
      >
        <v-icon start size="16">{{ tab.icon }}</v-icon>
        {{ tab.label }}
      </v-btn>
    </div>

    <div class="catalog-header__spacer" />

    <span v-if="booksCount > 0" class="text-caption text-medium-emphasis">
      {{ formatCount(booksCount) }} книг в библиотеке
    </span>

    <v-menu location="bottom end" :close-on-content-click="false">
      <template #activator="{ props }">
        <v-btn
          v-bind="props"
          variant="text"
          size="small"
          class="catalog-header__avatar"
        >
          <v-avatar size="28" color="primary">
            <span class="text-caption font-weight-bold text-white">{{ userInitials }}</span>
          </v-avatar>
        </v-btn>
      </template>

      <v-card min-width="200">
        <v-card-text class="pb-0">
          <div class="text-subtitle-2">{{ displayName }}</div>
          <div class="text-caption text-medium-emphasis">{{ userEmail }}</div>
        </v-card-text>

        <v-divider class="my-2" />

        <v-list density="compact">
          <v-list-item prepend-icon="mdi-account" disabled>
            <v-list-item-title>Мой профиль</v-list-item-title>
          </v-list-item>
          <v-list-item prepend-icon="mdi-cog" @click="showSettings = true">
            <v-list-item-title>Настройки</v-list-item-title>
          </v-list-item>
          <v-list-item prepend-icon="mdi-bookmark-multiple" disabled>
            <v-list-item-title>Мои коллекции</v-list-item-title>
          </v-list-item>
          <v-list-item prepend-icon="mdi-upload" disabled>
            <v-list-item-title>Загрузить книги</v-list-item-title>
          </v-list-item>
        </v-list>

        <v-divider />

        <div class="pa-2">
          <ThemeSwitcher />
        </div>

        <v-divider />

        <v-list density="compact">
          <v-list-item prepend-icon="mdi-logout" @click="onLogout">
            <v-list-item-title>Выйти</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-card>
    </v-menu>

    <SettingsDialog v-model="showSettings" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useCatalogStore } from '@/stores/catalog'
import { useAuthStore } from '@/stores/auth'
import { getStats } from '@/api/books'
import type { TabType } from '@/types/catalog'
import SettingsDialog from '@/components/catalog/SettingsDialog.vue'
import ThemeSwitcher from '@/components/catalog/ThemeSwitcher.vue'

const catalog = useCatalogStore()
const auth = useAuthStore()
const router = useRouter()

const showSettings = ref(false)
const booksCount = ref(0)

const tabs: Array<{ value: TabType; label: string; icon: string }> = [
  { value: 'authors', label: 'Авторы', icon: 'mdi-account-group' },
  { value: 'series', label: 'Серии', icon: 'mdi-bookshelf' },
  { value: 'genres', label: 'Жанры', icon: 'mdi-tag-multiple' },
  { value: 'search', label: 'Поиск', icon: 'mdi-magnify' },
]

const displayName = computed(() => auth.user?.display_name || auth.user?.username || '')
const userEmail = computed(() => auth.user?.email || '')
const userInitials = computed(() => {
  const name = displayName.value
  if (!name) return '?'
  const parts = name.split(/\s+/)
  return parts.length >= 2
    ? (parts[0][0] + parts[1][0]).toUpperCase()
    : name.substring(0, 2).toUpperCase()
})

function formatCount(count: number): string {
  if (count >= 1000000) return `${(count / 1000000).toFixed(1)}M`
  if (count >= 1000) return `${(count / 1000).toFixed(0)}K`
  return String(count)
}

async function onLogout() {
  await auth.logout()
  router.push({ name: 'login' })
}

onMounted(async () => {
  try {
    const stats = await getStats()
    booksCount.value = stats.books_count
  } catch {
    // Статистика опциональна
  }
})
</script>

<style scoped>
.catalog-header {
  display: flex;
  align-items: center;
  height: 40px;
  padding: 0 12px;
  background: rgb(var(--v-theme-surface));
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  gap: 8px;
}

.catalog-header__logo {
  display: flex;
  align-items: center;
  flex-shrink: 0;
}

.catalog-header__tabs {
  display: flex;
  gap: 2px;
}

.catalog-header__spacer {
  flex: 1;
}

.catalog-header__avatar {
  min-width: 0 !important;
  padding: 0 !important;
}
</style>
