import { reactive, onMounted } from 'vue'
import type { PanelSizes } from '@/types/catalog'
import { defaultCatalogSettings } from '@/types/catalog'
import api from '@/api/client'

const STORAGE_KEY = 'homelib-panel-sizes'

const MIN_LEFT_WIDTH = 10
const MAX_LEFT_WIDTH = 50
const MIN_TABLE_HEIGHT = 20
const MAX_TABLE_HEIGHT = 80

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

function loadFromLocalStorage(): PanelSizes | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw)
    if (typeof parsed.leftWidth === 'number' && typeof parsed.tableHeight === 'number') {
      return {
        leftWidth: clamp(parsed.leftWidth, MIN_LEFT_WIDTH, MAX_LEFT_WIDTH),
        tableHeight: clamp(parsed.tableHeight, MIN_TABLE_HEIGHT, MAX_TABLE_HEIGHT),
      }
    }
  } catch {
    // corrupted data
  }
  return null
}

function saveToLocalStorage(sizes: PanelSizes) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(sizes))
  } catch {
    // storage full or disabled
  }
}

export function usePanelResize() {
  const sizes = reactive<PanelSizes>({
    ...defaultCatalogSettings.panelSizes,
  })

  // Load from localStorage immediately for instant restore
  const cached = loadFromLocalStorage()
  if (cached) {
    sizes.leftWidth = cached.leftWidth
    sizes.tableHeight = cached.tableHeight
  }

  let saveTimer: ReturnType<typeof setTimeout> | null = null

  function scheduleSave() {
    if (saveTimer) clearTimeout(saveTimer)
    saveTimer = setTimeout(() => saveToServer(), 1000)
  }

  async function saveToServer() {
    try {
      await api.put('/me/settings', {
        catalog: { panelSizes: { leftWidth: sizes.leftWidth, tableHeight: sizes.tableHeight } },
      })
    } catch {
      // server save failed — localStorage already has the data
    }
  }

  function onVerticalResized(panes: Array<{ size: number }>) {
    if (panes[0]) {
      sizes.leftWidth = clamp(panes[0].size, MIN_LEFT_WIDTH, MAX_LEFT_WIDTH)
      saveToLocalStorage(sizes)
      scheduleSave()
    }
  }

  function onHorizontalResized(panes: Array<{ size: number }>) {
    if (panes[0]) {
      sizes.tableHeight = clamp(panes[0].size, MIN_TABLE_HEIGHT, MAX_TABLE_HEIGHT)
      saveToLocalStorage(sizes)
      scheduleSave()
    }
  }

  async function loadFromServer() {
    try {
      const { data } = await api.get<Record<string, unknown>>('/me/settings')
      const catalog = data.catalog as Record<string, unknown> | undefined
      const panelSizes = catalog?.panelSizes as Record<string, number> | undefined
      if (panelSizes && typeof panelSizes.leftWidth === 'number' && typeof panelSizes.tableHeight === 'number') {
        sizes.leftWidth = clamp(panelSizes.leftWidth, MIN_LEFT_WIDTH, MAX_LEFT_WIDTH)
        sizes.tableHeight = clamp(panelSizes.tableHeight, MIN_TABLE_HEIGHT, MAX_TABLE_HEIGHT)
        saveToLocalStorage(sizes)
      }
    } catch {
      // use cached/default values
    }
  }

  onMounted(() => {
    loadFromServer()
  })

  return {
    sizes,
    onVerticalResized,
    onHorizontalResized,
  }
}
