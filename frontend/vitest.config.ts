import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import vuetify from 'vite-plugin-vuetify'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [
    vue(),
    vuetify({ autoImport: true }),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test-setup.ts'],
    server: {
      deps: {
        inline: ['vuetify'],
      },
    },
    coverage: {
      provider: 'v8',
      reporter: ['text', 'lcov'],
      include: ['src/**/*.{ts,vue}'],
      thresholds: {
        perFile: true,
        lines: 80,
        functions: 0,
        branches: 70,
        statements: 80,
      },
      exclude: [
        'node_modules/',
        'dist/',
        'src/main.ts',
        'src/plugins/',
        'src/test-setup.ts',
        '**/*.d.ts',
        'src/App.vue',
        'src/**/__tests__/**',
        'src/router/**',
        'src/views/**',
        'src/components/AppHeader.vue',
        'src/components/common/PaginationBar.vue',
        'src/components/reader/**',
        'src/types/**',
        'src/composables/usePagination.ts',
        'src/composables/useReaderGestures.ts',
        'src/composables/useReaderKeyboard.ts',
        'src/composables/useReadingProgress.ts',
        'src/composables/useReaderSettings.ts',
      ],
    },
  },
})
