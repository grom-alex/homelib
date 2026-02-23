// Catalog types per spec 007-catalog-redesign

export type CatalogThemeName = 'light' | 'dark' | 'sepia' | 'night'

export interface CatalogThemeDefinition {
  name: CatalogThemeName
  label: string // «Светлая», «Тёмная», «Сепия», «Ночная»
  dark: boolean // Vuetify dark flag
  colors: Record<string, string>
  variables?: Record<string, string | number>
}

export type TabType = 'authors' | 'series' | 'genres' | 'search'

export type SortField = 'title' | 'year' | 'file_size'
export type SortOrder = 'asc' | 'desc'
export type PageSize = 25 | 50 | 75 | 100

export interface PanelSizes {
  leftWidth: number // 0-100, default 25
  tableHeight: number // 0-100, default 60
}

export interface TableSort {
  field: SortField
  order: SortOrder
}

export interface CatalogSettings {
  theme: CatalogThemeName
  panelSizes: PanelSizes
  activeTab: TabType
  tableSort: TableSort
  pageSize: PageSize
}

export const defaultCatalogSettings: CatalogSettings = {
  theme: 'light',
  panelSizes: {
    leftWidth: 25,
    tableHeight: 60,
  },
  activeTab: 'authors',
  tableSort: {
    field: 'title',
    order: 'asc',
  },
  pageSize: 25,
}

export interface NavigationFilter {
  type: 'author' | 'series' | 'genre' | 'search'
  id?: number
  label?: string
  params?: Record<string, string>
}

export interface BookTableRow {
  id: number
  title: string
  authorName: string // First author + «и др.» if multiple
  seriesName?: string // «Series #N» or undefined
  genreName: string // First genre
  fileSize: string // Formatted (e.g., «1.2 MB»)
  year?: number
  format: string
}
