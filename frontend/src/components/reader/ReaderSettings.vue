<template>
  <Teleport to="body">
    <div v-if="store.settingsVisible" class="settings-overlay" @click.self="store.toggleSettings()">
      <div class="settings-panel">
        <div class="settings-header">
          <h3>Настройки</h3>
          <button class="close-btn" @click="store.toggleSettings()">✕</button>
        </div>

        <div class="settings-body">
          <!-- Font size -->
          <div class="setting-row">
            <label>Размер шрифта</label>
            <div class="controls">
              <button @click="changeSetting('fontSize', -1)">−</button>
              <span class="value">{{ store.settings.fontSize }}px</span>
              <button @click="changeSetting('fontSize', 1)">+</button>
            </div>
          </div>

          <!-- Font family -->
          <div class="setting-row">
            <label>Шрифт</label>
            <ReaderFontPicker
              :model-value="store.settings.fontFamily"
              @update:model-value="store.updateSettings({ fontFamily: $event })"
            />
          </div>

          <!-- Font weight -->
          <div class="setting-row">
            <label>Насыщенность</label>
            <div class="toggle-group">
              <button
                :class="{ active: store.settings.fontWeight === 400 }"
                @click="store.updateSettings({ fontWeight: 400 })"
              >Нормальный</button>
              <button
                :class="{ active: store.settings.fontWeight === 500 }"
                @click="store.updateSettings({ fontWeight: 500 })"
              >Полужирный</button>
            </div>
          </div>

          <!-- Letter spacing -->
          <div class="setting-row">
            <label>Межбуквенный ({{ store.settings.letterSpacing.toFixed(2) }}em)</label>
            <input
              type="range"
              min="-0.05"
              max="0.1"
              step="0.01"
              :value="store.settings.letterSpacing"
              @input="store.updateSettings({ letterSpacing: parseFloat(($event.target as HTMLInputElement).value) })"
            />
          </div>

          <!-- Line height -->
          <div class="setting-row">
            <label>Интервал ({{ store.settings.lineHeight.toFixed(1) }})</label>
            <input
              type="range"
              min="1.0"
              max="2.5"
              step="0.1"
              :value="store.settings.lineHeight"
              @input="store.updateSettings({ lineHeight: parseFloat(($event.target as HTMLInputElement).value) })"
            />
          </div>

          <!-- Paragraph spacing -->
          <div class="setting-row">
            <label>Абзацный отступ ({{ store.settings.paragraphSpacing.toFixed(1) }}em)</label>
            <input
              type="range"
              min="0"
              max="2"
              step="0.1"
              :value="store.settings.paragraphSpacing"
              @input="store.updateSettings({ paragraphSpacing: parseFloat(($event.target as HTMLInputElement).value) })"
            />
          </div>

          <!-- Margins horizontal -->
          <div class="setting-row">
            <label>Поля по горизонтали ({{ store.settings.marginHorizontal }}%)</label>
            <input
              type="range"
              min="0"
              max="20"
              step="1"
              :value="store.settings.marginHorizontal"
              @input="store.updateSettings({ marginHorizontal: parseInt(($event.target as HTMLInputElement).value) })"
            />
          </div>

          <!-- Margins vertical -->
          <div class="setting-row">
            <label>Поля по вертикали ({{ store.settings.marginVertical }}%)</label>
            <input
              type="range"
              min="0"
              max="10"
              step="1"
              :value="store.settings.marginVertical"
              @input="store.updateSettings({ marginVertical: parseInt(($event.target as HTMLInputElement).value) })"
            />
          </div>

          <!-- First line indent -->
          <div class="setting-row">
            <label>Красная строка ({{ store.settings.firstLineIndent.toFixed(1) }}em)</label>
            <input
              type="range"
              min="0"
              max="3"
              step="0.5"
              :value="store.settings.firstLineIndent"
              @input="store.updateSettings({ firstLineIndent: parseFloat(($event.target as HTMLInputElement).value) })"
            />
          </div>

          <!-- Text align -->
          <div class="setting-row">
            <label>Выравнивание</label>
            <div class="toggle-group">
              <button
                :class="{ active: store.settings.textAlign === 'left' }"
                @click="store.updateSettings({ textAlign: 'left' })"
              >По левому</button>
              <button
                :class="{ active: store.settings.textAlign === 'justify' }"
                @click="store.updateSettings({ textAlign: 'justify' })"
              >По ширине</button>
            </div>
          </div>

          <!-- Hyphenation -->
          <div class="setting-row">
            <label>Авто-переносы</label>
            <button
              class="toggle-btn"
              :class="{ active: store.settings.hyphenation }"
              @click="store.updateSettings({ hyphenation: !store.settings.hyphenation })"
            >{{ store.settings.hyphenation ? 'Вкл' : 'Выкл' }}</button>
          </div>

          <!-- Themes -->
          <div class="setting-row">
            <label>Тема</label>
            <div class="theme-buttons">
              <button
                v-for="theme in themes"
                :key="theme.id"
                class="theme-btn"
                :class="{ active: store.settings.theme === theme.id }"
                :style="{ background: theme.bg, color: theme.text, border: `2px solid ${theme.border}` }"
                @click="store.updateSettings({ theme: theme.id })"
              >Aa</button>
            </div>
          </div>

          <!-- View mode -->
          <div class="setting-row">
            <label>Режим</label>
            <div class="toggle-group">
              <button
                :class="{ active: store.settings.viewMode === 'paginated' }"
                @click="store.updateSettings({ viewMode: 'paginated' })"
              >Страницы</button>
              <button
                :class="{ active: store.settings.viewMode === 'scroll' }"
                @click="store.updateSettings({ viewMode: 'scroll' })"
              >Прокрутка</button>
            </div>
          </div>

          <!-- Page animation -->
          <div class="setting-row">
            <label>Анимация</label>
            <div class="toggle-group">
              <button
                v-for="anim in ['slide', 'fade', 'none'] as const"
                :key="anim"
                :class="{ active: store.settings.pageAnimation === anim }"
                @click="store.updateSettings({ pageAnimation: anim })"
              >{{ animLabels[anim] }}</button>
            </div>
          </div>

          <!-- Show progress -->
          <div class="setting-row">
            <label>Прогресс</label>
            <button
              class="toggle-btn"
              :class="{ active: store.settings.showProgress }"
              @click="store.updateSettings({ showProgress: !store.settings.showProgress })"
            >{{ store.settings.showProgress ? 'Вкл' : 'Выкл' }}</button>
          </div>

          <!-- Show clock -->
          <div class="setting-row">
            <label>Часы</label>
            <button
              class="toggle-btn"
              :class="{ active: store.settings.showClock }"
              @click="store.updateSettings({ showClock: !store.settings.showClock })"
            >{{ store.settings.showClock ? 'Вкл' : 'Выкл' }}</button>
          </div>

          <!-- Tap zones -->
          <div class="setting-row">
            <label>Зоны тапа</label>
            <div class="toggle-group">
              <button
                :class="{ active: store.settings.tapZones === 'lrc' }"
                @click="store.updateSettings({ tapZones: 'lrc' })"
              >Л/Ц/П</button>
              <button
                :class="{ active: store.settings.tapZones === 'lr' }"
                @click="store.updateSettings({ tapZones: 'lr' })"
              >Л/П</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { useReaderStore } from '@/stores/reader'
import type { ReaderSettings } from '@/types/reader'
import ReaderFontPicker from './ReaderFontPicker.vue'

const store = useReaderStore()

const themes: { id: ReaderSettings['theme']; bg: string; text: string; border: string }[] = [
  { id: 'light', bg: '#ffffff', text: '#1a1a1a', border: '#e2e8f0' },
  { id: 'sepia', bg: '#f5e6d3', text: '#5c4b37', border: '#d4c4b0' },
  { id: 'dark', bg: '#1e1e1e', text: '#d4d4d4', border: '#404040' },
  { id: 'night', bg: '#000000', text: '#666666', border: '#1a1a1a' },
  { id: 'custom', bg: 'linear-gradient(135deg, #ff6b6b, #4ecdc4)', text: '#ffffff', border: '#999' },
]

const animLabels: Record<string, string> = {
  slide: 'Сдвиг',
  fade: 'Плавно',
  none: 'Без',
}

function changeSetting(key: 'fontSize', delta: number) {
  const current = store.settings[key]
  const limits: Record<string, [number, number]> = {
    fontSize: [12, 36],
  }
  const [min, max] = limits[key]
  const newVal = Math.max(min, Math.min(max, current + delta))
  store.updateSettings({ [key]: newVal })
}
</script>

<style scoped>
.settings-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 200;
  display: flex;
  justify-content: flex-end;
}

.settings-panel {
  width: 360px;
  max-width: 90vw;
  height: 100%;
  background: var(--reader-bg, #fff);
  color: var(--reader-text, #1a1a1a);
  overflow-y: auto;
  box-shadow: -4px 0 20px rgba(0, 0, 0, 0.2);
}

.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--reader-border, #e2e8f0);
}

.settings-header h3 {
  margin: 0;
  font-size: 18px;
}

.close-btn {
  background: none;
  border: none;
  font-size: 20px;
  cursor: pointer;
  color: inherit;
  padding: 4px 8px;
}

.settings-body {
  padding: 12px 20px;
}

.setting-row {
  margin-bottom: 16px;
}

.setting-row label {
  display: block;
  font-size: 13px;
  margin-bottom: 6px;
  opacity: 0.7;
}

.controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.controls button {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid var(--reader-border, #e2e8f0);
  background: transparent;
  color: inherit;
  font-size: 18px;
  cursor: pointer;
}

.controls .value {
  min-width: 50px;
  text-align: center;
  font-size: 16px;
}

.toggle-group {
  display: flex;
  gap: 4px;
}

.toggle-group button {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid var(--reader-border, #e2e8f0);
  border-radius: 6px;
  background: transparent;
  color: inherit;
  font-size: 13px;
  cursor: pointer;
}

.toggle-group button.active {
  background: rgba(128, 128, 128, 0.15);
  font-weight: 500;
}

.toggle-btn {
  padding: 6px 16px;
  border: 1px solid var(--reader-border, #e2e8f0);
  border-radius: 6px;
  background: transparent;
  color: inherit;
  font-size: 13px;
  cursor: pointer;
}

.toggle-btn.active {
  background: rgba(128, 128, 128, 0.15);
}

.theme-buttons {
  display: flex;
  gap: 8px;
}

.theme-btn {
  width: 44px;
  height: 44px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}

.theme-btn.active {
  outline: 2px solid currentColor;
  outline-offset: 2px;
}

input[type="range"] {
  width: 100%;
  accent-color: var(--reader-link, #2563eb);
}
</style>
