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
    background: '#F5F0E8',
    surface: '#FAF6EE',
    'surface-variant': '#EDE6D6',
    primary: '#8B6914',
    secondary: '#5D4E37',
    accent: '#A67C00',
    error: '#C62828',
    success: '#558B2F',
    warning: '#E65100',
    info: '#01579B',
    'on-background': '#3E2723',
    'on-surface': '#3E2723',
    'table-row-hover': '#EDE6D6',
    'table-row-selected': '#E0D5BE',
    'nav-item-active': '#E0D5BE',
    'status-bar': '#EDE6D6',
  },
}

const night: ThemeDefinition = {
  dark: true,
  colors: {
    background: '#0D1117',
    surface: '#161B22',
    'surface-variant': '#1C2128',
    primary: '#58A6FF',
    secondary: '#8B949E',
    accent: '#58A6FF',
    error: '#F85149',
    success: '#3FB950',
    warning: '#D29922',
    info: '#58A6FF',
    'on-background': '#C9D1D9',
    'on-surface': '#C9D1D9',
    'table-row-hover': '#1C2128',
    'table-row-selected': '#1F2937',
    'nav-item-active': '#1F2937',
    'status-bar': '#161B22',
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
  defaults: {
    global: {
      style: "font-family: 'Source Sans 3', sans-serif",
    },
  },
})
