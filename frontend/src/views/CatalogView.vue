<template>
  <div class="catalog-view">
    <div class="catalog-header-slot">
      <CatalogHeader />
    </div>

    <div class="catalog-content">
      <Splitpanes class="default-theme" @resized="onVerticalResized">
        <Pane :size="panelSizes.leftWidth" :min-size="10" :max-size="50">
          <NavigationPanel />
        </Pane>
        <Pane :size="100 - panelSizes.leftWidth">
          <Splitpanes horizontal @resized="onHorizontalResized">
            <Pane :size="panelSizes.tableHeight" :min-size="20" :max-size="80">
              <BookTable />
            </Pane>
            <Pane :size="100 - panelSizes.tableHeight" :min-size="15">
              <BookDetailPanel />
            </Pane>
          </Splitpanes>
        </Pane>
      </Splitpanes>
    </div>

    <StatusBar />
  </div>
</template>

<script setup lang="ts">
defineOptions({ name: 'CatalogView' })

import { onMounted } from 'vue'
import { Splitpanes, Pane } from 'splitpanes'
import 'splitpanes/dist/splitpanes.css'
import { useThemeStore } from '@/stores/theme'
import { usePanelResize } from '@/composables/usePanelResize'
import CatalogHeader from '@/components/catalog/CatalogHeader.vue'
import NavigationPanel from '@/components/catalog/NavigationPanel.vue'
import BookTable from '@/components/catalog/BookTable.vue'
import BookDetailPanel from '@/components/catalog/BookDetailPanel.vue'
import StatusBar from '@/components/catalog/StatusBar.vue'

const themeStore = useThemeStore()
const { sizes: panelSizes, onVerticalResized, onHorizontalResized } = usePanelResize()

onMounted(() => {
  themeStore.loadSettings()
})
</script>

<style scoped>
.catalog-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  overflow: hidden;
  background: rgb(var(--v-theme-background));
  color: rgb(var(--v-theme-on-background));
}

.catalog-header-slot {
  flex-shrink: 0;
}

.catalog-content {
  flex: 1;
  overflow: hidden;
}

.catalog-content :deep(.splitpanes__splitter) {
  background: rgb(var(--v-theme-surface-variant));
  position: relative;
}

.catalog-content :deep(.splitpanes--vertical > .splitpanes__splitter) {
  width: 4px;
  cursor: col-resize;
}

.catalog-content :deep(.splitpanes--horizontal > .splitpanes__splitter) {
  height: 4px;
  cursor: row-resize;
}

.catalog-content :deep(.splitpanes__splitter:hover) {
  background: rgb(var(--v-theme-primary));
}
</style>
