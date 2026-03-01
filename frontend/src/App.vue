<template>
  <v-app class="app-root">
    <CatalogHeader v-if="showHeader" />
    <div class="app-content" :class="{ 'app-content--scrollable': !isCatalogRoute }">
      <router-view v-slot="{ Component }">
        <keep-alive include="CatalogView">
          <component :is="Component" />
        </keep-alive>
      </router-view>
    </div>
  </v-app>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import CatalogHeader from '@/components/catalog/CatalogHeader.vue'

const route = useRoute()
const isReaderRoute = computed(() => route.name === 'reader')
const isLoginRoute = computed(() => route.name === 'login')
const isCatalogRoute = computed(() => route.name === 'catalog')
const showHeader = computed(() => !isReaderRoute.value && !isLoginRoute.value)
</script>

<style>
.v-application,
.v-application .v-btn,
.v-application .v-card,
.v-application .v-list,
.v-application .v-field {
  font-family: 'Source Sans 3 Variable', 'Source Sans 3', sans-serif;
}

/* Typography scale: base 13px (proportional from Vuetify's 16px, ×0.8125) */
.v-application { font-size: 13px; }
.v-application .text-h3 { font-size: 36px !important; line-height: 1.2 !important; }
.v-application .text-h4 { font-size: 22px !important; line-height: 1.3 !important; }
.v-application .text-h5 { font-size: 18px !important; line-height: 1.4 !important; }
.v-application .text-h6 { font-size: 15px !important; line-height: 1.4 !important; }
.v-application .text-body-1 { font-size: 13px !important; }
.v-application .text-body-2 { font-size: 12px !important; }
.v-application .text-subtitle-1 { font-size: 13px !important; }
.v-application .text-subtitle-2 { font-size: 12px !important; }
.v-application .text-caption { font-size: 10px !important; }

.v-application .v-card-title { font-size: 15px; line-height: 1.4; }
.v-application .v-card-subtitle { font-size: 12px; }
.v-application .v-btn { font-size: 13px; }
.v-application .v-chip { font-size: 12px; }
.v-application .v-alert { font-size: 13px; }
.v-application .v-tab { font-size: 13px; }
.v-application .v-table { font-size: 13px; }
.v-application .v-table th { font-size: 11px; }
.v-application .v-list-item-title { font-size: 13px; }
.v-application .v-list-item-subtitle { font-size: 12px; }
.v-application .v-field__input,
.v-application .v-label { font-size: 14px; }
.v-application .v-treeview-item { font-size: 13px; }

.app-root {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.app-content {
  flex: 1;
  overflow: hidden;
}

.app-content--scrollable {
  overflow-y: auto;
}
</style>
