<template>
  <v-app>
    <AppHeader v-if="!isReaderRoute && !isCatalogRoute" />
    <v-main>
      <router-view />
    </v-main>
  </v-app>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppHeader from '@/components/AppHeader.vue'

const route = useRoute()
const isReaderRoute = computed(() => route.name === 'reader')
const isCatalogRoute = computed(() => route.name === 'catalog')
</script>

<style>
.v-application {
  font-family: 'Source Sans 3', sans-serif !important;
}
</style>
