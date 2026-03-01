<template>
  <v-dialog v-model="dialog" max-width="360" persistent>
    <v-card>
      <v-card-title class="text-h6">Разблокировка контента</v-card-title>
      <v-card-text>
        <p class="mb-4 text-body-2">Введите родительский PIN для доступа к контенту 18+</p>
        <v-text-field
          v-model="pin"
          label="PIN-код"
          type="password"
          maxlength="6"
          counter
          :error-messages="errorMessage"
          :disabled="loading"
          autofocus
          @keyup.enter="onSubmit"
        />
      </v-card-text>
      <v-card-actions>
        <v-spacer />
        <v-btn variant="text" :disabled="loading" @click="onCancel">Отмена</v-btn>
        <v-btn color="primary" :loading="loading" :disabled="pin.length < 4" @click="onSubmit">
          Подтвердить
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useParentalStore } from '@/stores/parental'

const dialog = defineModel<boolean>({ default: false })

const emit = defineEmits<{
  unlocked: []
}>()

const parentalStore = useParentalStore()
const pin = ref('')
const loading = ref(false)
const errorMessage = ref('')

watch(dialog, (open) => {
  if (open) {
    pin.value = ''
    errorMessage.value = ''
  }
})

async function onSubmit() {
  if (pin.value.length < 4) return
  loading.value = true
  errorMessage.value = ''
  try {
    await parentalStore.unlock(pin.value)
    dialog.value = false
    emit('unlocked')
  } catch {
    errorMessage.value = 'Неверный PIN-код'
  } finally {
    loading.value = false
  }
}

function onCancel() {
  dialog.value = false
}
</script>
