<template>
  <v-app-bar color="primary" density="comfortable">
    <v-app-bar-title>
      <router-link to="/books" class="text-white text-decoration-none">HomeLib</router-link>
    </v-app-bar-title>

    <template v-if="auth.isAuthenticated">
      <v-btn to="/books" variant="text">Каталог</v-btn>
      <v-btn to="/authors" variant="text">Авторы</v-btn>
      <v-btn to="/genres" variant="text">Жанры</v-btn>
      <v-btn to="/series" variant="text">Серии</v-btn>
      <v-btn v-if="auth.isAdmin" to="/admin/import" variant="text">Импорт</v-btn>

      <v-spacer />

      <v-menu>
        <template #activator="{ props }">
          <v-btn v-bind="props" icon="mdi-account-circle" />
        </template>
        <v-list>
          <v-list-item>
            <v-list-item-title>{{ auth.user?.display_name }}</v-list-item-title>
            <v-list-item-subtitle>{{ auth.user?.email }}</v-list-item-subtitle>
          </v-list-item>
          <v-divider />
          <v-list-item @click="handleLogout">
            <template #prepend>
              <v-icon icon="mdi-logout" />
            </template>
            <v-list-item-title>Выйти</v-list-item-title>
          </v-list-item>
        </v-list>
      </v-menu>
    </template>
  </v-app-bar>
</template>

<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const auth = useAuthStore()
const router = useRouter()

async function handleLogout() {
  await auth.logout()
  router.push('/login')
}
</script>
