<template>
  <div class="font-picker">
    <div
      v-for="font in fontFamilies"
      :key="font"
      class="font-option"
      :class="{ active: modelValue === font }"
      :style="{ fontFamily: font === 'System' ? 'system-ui, sans-serif' : `'${font}', serif` }"
      @click="$emit('update:modelValue', font)"
    >
      {{ font === 'System' ? 'Системный' : font }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { fontFamilies } from '@/types/reader'

defineProps<{ modelValue: string }>()
defineEmits<{ 'update:modelValue': [value: string] }>()
</script>

<style scoped>
.font-picker {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.font-option {
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 16px;
  transition: background 0.2s;
}

.font-option:hover {
  background: rgba(128, 128, 128, 0.1);
}

.font-option.active {
  background: rgba(var(--v-theme-primary), 0.12);
  font-weight: 500;
}
</style>
