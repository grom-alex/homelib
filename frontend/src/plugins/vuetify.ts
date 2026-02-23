import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createVuetify } from 'vuetify'
import { aliases, mdi } from 'vuetify/iconsets/mdi'
import type { ThemeDefinition } from 'vuetify'

const light: ThemeDefinition = {
  dark: false,
  colors: {
    background: '#FFFFFF',
    surface: '#FFFFFF',
    'surface-variant': '#F5F5F5',
    primary: '#1565C0',
    secondary: '#424242',
    accent: '#FF6F00',
    error: '#D32F2F',
    success: '#388E3C',
    warning: '#F57C00',
    info: '#0288D1',
    'on-background': '#212121',
    'on-surface': '#212121',
    'table-row-hover': '#E3F2FD',
    'table-row-selected': '#BBDEFB',
    'nav-item-active': '#E3F2FD',
    'status-bar': '#F5F5F5',
  },
}

const dark: ThemeDefinition = {
  dark: true,
  colors: {
    background: '#1E1E1E',
    surface: '#252526',
    'surface-variant': '#2D2D2D',
    primary: '#d4a017',
    secondary: '#BDBDBD',
    accent: '#d4a017',
    error: '#EF5350',
    success: '#66BB6A',
    warning: '#FFA726',
    info: '#29B6F6',
    'on-background': '#E0E0E0',
    'on-surface': '#E0E0E0',
    'table-row-hover': '#2A2D2E',
    'table-row-selected': '#37373D',
    'nav-item-active': '#37373D',
    'status-bar': '#007ACC',
  },
}

const sepia: ThemeDefinition = {
  dark: false,
  colors: {
    background: '#f5e6d3',
    surface: '#faf0e4',
    'surface-variant': '#ede0cf',
    primary: '#8b5a2b',
    secondary: '#5c4b37',
    accent: '#8b5a2b',
    error: '#C62828',
    success: '#558B2F',
    warning: '#E65100',
    info: '#01579B',
    'on-background': '#5c4b37',
    'on-surface': '#5c4b37',
    'table-row-hover': '#ede0cf',
    'table-row-selected': '#d4c4b0',
    'nav-item-active': '#d4c4b0',
    'status-bar': '#ede0cf',
  },
}

const night: ThemeDefinition = {
  dark: true,
  colors: {
    background: '#000000',
    surface: '#0a0a0a',
    'surface-variant': '#141414',
    primary: '#4a90d9',
    secondary: '#666666',
    accent: '#4a90d9',
    error: '#993333',
    success: '#2d6630',
    warning: '#8a6d2b',
    info: '#4a90d9',
    'on-background': '#666666',
    'on-surface': '#666666',
    'table-row-hover': '#141414',
    'table-row-selected': '#1a1a1a',
    'nav-item-active': '#1a1a1a',
    'status-bar': '#0a0a0a',
  },
}

export default createVuetify({
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: { mdi },
  },
  theme: {
    defaultTheme: 'light',
    themes: { light, dark, sepia, night },
  },
})
