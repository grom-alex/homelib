<template>
  <v-app>
    <AppHeader />
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import AppHeader from '@/components/AppHeader.vue'

const auth = useAuthStore()

onMounted(async () => {
  auth.init()
  if (auth.isAuthenticated) {
    try {
      await auth.refreshToken()
    } catch {
      // Token expired, user will be redirected to login
    }
  }
})
</script>
