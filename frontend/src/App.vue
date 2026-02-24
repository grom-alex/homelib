<template>
  <v-app>
    <AppHeader v-if="!isReaderRoute && !isCatalogRoute" />
    <v-main>
      <router-view v-slot="{ Component }">
        <keep-alive include="CatalogView">
          <component :is="Component" />
        </keep-alive>
      </router-view>
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
.v-application,
.v-application .v-btn,
.v-application .v-card,
.v-application .v-list,
.v-application .v-field {
  font-family: 'Source Sans 3 Variable', 'Source Sans 3', sans-serif;
}
</style>
