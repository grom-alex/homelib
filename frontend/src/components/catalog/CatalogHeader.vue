<template>
  <div class="catalog-header">
    <div class="catalog-header__logo">
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20" />
        <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z" />
      </svg>
      <span class="catalog-header__logo-name">MyHomeLib</span>
      <span class="catalog-header__logo-suffix">web</span>
    </div>

    <nav class="catalog-header__tabs">
      <button
        v-for="tab in tabs"
        :key="tab.value"
        class="catalog-header__tab"
        :class="{ 'catalog-header__tab--active': isCatalogRoute && catalog.activeTab === tab.value }"
        @click="onTabClick(tab.value)"
      >
        <v-icon size="14">{{ tab.icon }}</v-icon>
        {{ tab.label }}
      </button>
    </nav>

    <div class="catalog-header__spacer" />

    <span v-if="booksCount > 0" class="catalog-header__count">
      Книг: <span class="catalog-header__count-value">{{ formatCount(booksCount) }}</span>
    </span>

    <template v-if="parentalStore.pinSet">
      <div class="catalog-header__divider" />
      <button
        class="catalog-header__parental-btn"
        :class="{ 'catalog-header__parental-btn--unlocked': parentalStore.adultContentEnabled }"
        :title="parentalStore.adultContentEnabled ? 'Контент 18+ разблокирован. Нажмите для блокировки' : 'Контент 18+ заблокирован. Нажмите для разблокировки'"
        @click="onParentalToggle"
      >
        <v-icon size="16">{{ parentalStore.adultContentEnabled ? 'mdi-lock-open-variant' : 'mdi-lock' }}</v-icon>
        <span class="catalog-header__parental-label">18+</span>
      </button>
    </template>

    <div class="catalog-header__divider" />

    <div class="catalog-header__user-area">
      <button
        class="catalog-header__user-btn"
        :class="{ 'catalog-header__user-btn--open': userMenuOpen }"
        @click="userMenuOpen = !userMenuOpen"
      >
        <div class="catalog-header__avatar">{{ userInitials }}</div>
        <span class="catalog-header__user-name">{{ displayName }}</span>
        <svg
          width="12" height="12" viewBox="0 0 24 24"
          fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"
          class="catalog-header__chevron"
          :class="{ 'catalog-header__chevron--open': userMenuOpen }"
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>

      <template v-if="userMenuOpen">
        <div class="catalog-header__overlay" @click="userMenuOpen = false" />
        <div class="catalog-header__dropdown">
          <div class="catalog-header__dropdown-header">
            <div class="catalog-header__avatar catalog-header__avatar--large">{{ userInitials }}</div>
            <div>
              <div class="catalog-header__dropdown-name">{{ displayName }}</div>
              <div class="catalog-header__dropdown-email">{{ userEmail }}</div>
            </div>
          </div>
          <div class="catalog-header__dropdown-items">
            <button class="catalog-header__dropdown-item" disabled>
              <v-icon size="15">mdi-account</v-icon>
              Мой профиль
            </button>
            <button class="catalog-header__dropdown-item" @click="onOpenSettings">
              <v-icon size="15">mdi-cog</v-icon>
              Настройки
            </button>
            <button class="catalog-header__dropdown-item" disabled>
              <v-icon size="15">mdi-bookmark-multiple</v-icon>
              Мои коллекции
            </button>
            <button class="catalog-header__dropdown-item" disabled>
              <v-icon size="15">mdi-upload</v-icon>
              Загрузить книги
            </button>
            <button
              v-if="auth.user?.role === 'admin'"
              class="catalog-header__dropdown-item"
              @click="onOpenImport"
            >
              <v-icon size="15">mdi-database-import</v-icon>
              Импорт
            </button>
            <button
              v-if="auth.user?.role === 'admin'"
              class="catalog-header__dropdown-item"
              @click="onOpenParentalAdmin"
            >
              <v-icon size="15">mdi-shield-lock</v-icon>
              Родительский контроль
            </button>
          </div>
          <div class="catalog-header__dropdown-divider" />
          <div class="catalog-header__dropdown-section">
            <ThemeSwitcher />
          </div>
          <div class="catalog-header__dropdown-divider" />
          <div class="catalog-header__dropdown-items">
            <button class="catalog-header__dropdown-item catalog-header__dropdown-item--danger" @click="onLogout">
              <v-icon size="15">mdi-logout</v-icon>
              Выйти
            </button>
          </div>
        </div>
      </template>
    </div>

    <SettingsDialog v-model="showSettings" />
    <PinUnlockDialog v-model="showPinDialog" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useCatalogStore } from '@/stores/catalog'
import { useAuthStore } from '@/stores/auth'
import { useParentalStore } from '@/stores/parental'
import { getStats } from '@/api/books'
import type { TabType } from '@/types/catalog'
import SettingsDialog from '@/components/catalog/SettingsDialog.vue'
import ThemeSwitcher from '@/components/catalog/ThemeSwitcher.vue'
import PinUnlockDialog from '@/components/common/PinUnlockDialog.vue'

const catalog = useCatalogStore()
const auth = useAuthStore()
const parentalStore = useParentalStore()
const router = useRouter()
const route = useRoute()

const isCatalogRoute = computed(() => route.name === 'catalog')

const showPinDialog = ref(false)

const showSettings = ref(false)
const booksCount = ref(0)
const userMenuOpen = ref(false)

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

function onTabClick(tab: TabType) {
  catalog.setActiveTab(tab)
  if (!isCatalogRoute.value) {
    router.push({ name: 'catalog' })
  }
}

function onOpenSettings() {
  userMenuOpen.value = false
  showSettings.value = true
}

function onOpenImport() {
  userMenuOpen.value = false
  router.push('/admin/import')
}

function onOpenParentalAdmin() {
  userMenuOpen.value = false
  router.push('/admin/parental')
}

function onParentalToggle() {
  if (parentalStore.adultContentEnabled) {
    parentalStore.lock()
  } else {
    showPinDialog.value = true
  }
}

async function onLogout() {
  userMenuOpen.value = false
  parentalStore.reset()
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
  height: 48px;
  padding: 0 16px;
  background: rgb(var(--v-theme-surface));
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  flex-shrink: 0;
}

.catalog-header__logo {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-right: 32px;
  flex-shrink: 0;
  color: rgb(var(--v-theme-on-surface));
}

.catalog-header__logo-name {
  font-weight: 700;
  font-size: 15px;
  letter-spacing: -0.3px;
}

.catalog-header__logo-suffix {
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  font-weight: 400;
  margin-left: 2px;
}

.catalog-header__tabs {
  display: flex;
  gap: 0;
  height: 100%;
}

.catalog-header__tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 0 18px;
  background: none;
  border: none;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.5;
  cursor: pointer;
  font-family: inherit;
  font-size: 13px;
  font-weight: 400;
  border-bottom: 2px solid transparent;
  transition: all 0.15s;
  height: 100%;
}

.catalog-header__tab:hover {
  opacity: 0.8;
}

.catalog-header__tab--active {
  color: rgb(var(--v-theme-primary));
  opacity: 1;
  font-weight: 600;
  border-bottom-color: rgb(var(--v-theme-primary));
}

.catalog-header__spacer {
  flex: 1;
}

.catalog-header__count {
  font-size: 11px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
}

.catalog-header__count-value {
  color: rgb(var(--v-theme-primary));
  font-weight: 600;
  opacity: 1;
}

.catalog-header__parental-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: none;
  border: 1px solid rgb(var(--v-theme-surface-variant));
  border-radius: 6px;
  cursor: pointer;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.5;
  font-family: inherit;
  font-size: 11px;
  font-weight: 600;
  transition: all 0.15s;
}

.catalog-header__parental-btn:hover {
  opacity: 0.8;
  background: rgb(var(--v-theme-table-row-hover));
}

.catalog-header__parental-btn--unlocked {
  opacity: 0.8;
  color: rgb(var(--v-theme-warning));
  border-color: rgb(var(--v-theme-warning));
}

.catalog-header__parental-label {
  line-height: 1;
}

.catalog-header__divider {
  width: 1px;
  height: 20px;
  background: rgb(var(--v-theme-surface-variant));
  margin: 0 14px;
}

.catalog-header__user-area {
  position: relative;
}

.catalog-header__user-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 8px 4px 4px;
  background: none;
  border: 1px solid transparent;
  border-radius: 6px;
  cursor: pointer;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.6;
  font-family: inherit;
  transition: all 0.15s;
}

.catalog-header__user-btn:hover {
  background: rgb(var(--v-theme-table-row-hover));
  border-color: rgb(var(--v-theme-surface-variant));
  opacity: 1;
}

.catalog-header__user-btn--open {
  background: rgb(var(--v-theme-table-row-hover));
  border-color: rgb(var(--v-theme-surface-variant));
  opacity: 1;
}

.catalog-header__avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: linear-gradient(135deg, #d4a017 0%, #b8860b 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: #1a1d23;
  flex-shrink: 0;
  letter-spacing: -0.5px;
}

.catalog-header__avatar--large {
  width: 36px;
  height: 36px;
  font-size: 14px;
}

.catalog-header__user-name {
  font-size: 13px;
  font-weight: 500;
}

.catalog-header__chevron {
  transition: transform 0.15s;
}

.catalog-header__chevron--open {
  transform: rotate(180deg);
}

.catalog-header__overlay {
  position: fixed;
  inset: 0;
  z-index: 99;
}

.catalog-header__dropdown {
  position: absolute;
  top: calc(100% + 6px);
  right: 0;
  z-index: 100;
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgb(var(--v-theme-surface-variant));
  border-radius: 8px;
  min-width: 220px;
  box-shadow: 0 12px 32px rgba(0, 0, 0, 0.25);
  overflow: hidden;
}

.catalog-header__dropdown-header {
  padding: 14px 14px 12px;
  border-bottom: 1px solid rgb(var(--v-theme-surface-variant));
  display: flex;
  align-items: center;
  gap: 10px;
}

.catalog-header__dropdown-name {
  font-weight: 600;
  font-size: 14px;
  color: rgb(var(--v-theme-on-surface));
}

.catalog-header__dropdown-email {
  font-size: 12px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.4;
  margin-top: 1px;
}

.catalog-header__dropdown-items {
  padding: 4px 0;
}

.catalog-header__dropdown-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 14px;
  font-size: 13px;
  color: rgb(var(--v-theme-on-surface));
  opacity: 0.6;
  cursor: pointer;
  transition: all 0.12s;
  border: none;
  background: none;
  width: 100%;
  font-family: inherit;
  text-align: left;
}

.catalog-header__dropdown-item:hover:not(:disabled) {
  background: rgb(var(--v-theme-table-row-hover));
  opacity: 1;
}

.catalog-header__dropdown-item:disabled {
  opacity: 0.3;
  cursor: default;
}

.catalog-header__dropdown-item--danger:hover:not(:disabled) {
  background: rgba(220, 60, 60, 0.1);
  color: #e05555;
  opacity: 1;
}

.catalog-header__dropdown-divider {
  height: 1px;
  background: rgb(var(--v-theme-surface-variant));
  margin: 4px 0;
}

.catalog-header__dropdown-section {
  padding: 8px 14px;
}
</style>
