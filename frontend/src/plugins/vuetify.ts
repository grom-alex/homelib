import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createVuetify } from 'vuetify'
import { aliases, mdi } from 'vuetify/iconsets/mdi'

export default createVuetify({
  icons: {
    defaultSet: 'mdi',
    aliases,
    sets: { mdi },
  },
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary: '#1565C0',
          secondary: '#424242',
          accent: '#FF6F00',
          error: '#D32F2F',
          success: '#388E3C',
          warning: '#F57C00',
          info: '#0288D1',
        },
      },
      dark: {
        colors: {
          primary: '#42A5F5',
          secondary: '#BDBDBD',
          accent: '#FFB74D',
          error: '#EF5350',
          success: '#66BB6A',
          warning: '#FFA726',
          info: '#29B6F6',
        },
      },
    },
  },
})
