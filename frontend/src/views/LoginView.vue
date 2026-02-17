<template>
  <v-container class="fill-height" fluid>
    <v-row justify="center">
      <v-col cols="12" sm="8" md="4">
        <v-card elevation="8">
          <v-card-title class="text-h5 text-center pa-6">HomeLib</v-card-title>
          <v-tabs v-model="tab" grow>
            <v-tab value="login">Вход</v-tab>
            <v-tab value="register">Регистрация</v-tab>
          </v-tabs>
          <v-card-text>
            <v-alert v-if="error" type="error" class="mb-4" closable @click:close="error = ''">
              {{ error }}
            </v-alert>

            <v-window v-model="tab">
              <v-window-item value="login">
                <v-form ref="loginFormRef" @submit.prevent="handleLogin">
                  <v-text-field
                    v-model="loginForm.email"
                    label="Email"
                    type="email"
                    required
                    :rules="[rules.required, rules.email]"
                    prepend-inner-icon="mdi-email"
                  />
                  <v-text-field
                    v-model="loginForm.password"
                    label="Пароль"
                    type="password"
                    required
                    :rules="[rules.required]"
                    prepend-inner-icon="mdi-lock"
                  />
                  <v-btn type="submit" color="primary" block size="large" :loading="loading" class="mt-4">
                    Войти
                  </v-btn>
                </v-form>
              </v-window-item>

              <v-window-item value="register">
                <v-form ref="registerFormRef" @submit.prevent="handleRegister">
                  <v-text-field
                    v-model="registerForm.email"
                    label="Email"
                    type="email"
                    required
                    :rules="[rules.required, rules.email]"
                    prepend-inner-icon="mdi-email"
                  />
                  <v-text-field
                    v-model="registerForm.username"
                    label="Имя пользователя"
                    required
                    :rules="[rules.required, rules.minLength(3)]"
                    prepend-inner-icon="mdi-account"
                  />
                  <v-text-field
                    v-model="registerForm.display_name"
                    label="Отображаемое имя"
                    required
                    :rules="[rules.required]"
                    prepend-inner-icon="mdi-badge-account"
                  />
                  <v-text-field
                    v-model="registerForm.password"
                    label="Пароль"
                    type="password"
                    required
                    :rules="[rules.required, rules.minLength(8)]"
                    prepend-inner-icon="mdi-lock"
                  />
                  <v-btn type="submit" color="primary" block size="large" :loading="loading" class="mt-4">
                    Зарегистрироваться
                  </v-btn>
                </v-form>
              </v-window-item>
            </v-window>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { ref, reactive, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const tab = ref('login')
const loading = ref(false)
const error = ref('')

const loginFormRef: Ref<{ validate: () => Promise<{ valid: boolean }> } | null> = ref(null)
const registerFormRef: Ref<{ validate: () => Promise<{ valid: boolean }> } | null> = ref(null)

const loginForm = reactive({ email: '', password: '' })
const registerForm = reactive({ email: '', username: '', display_name: '', password: '' })

const rules = {
  required: (v: string) => !!v || 'Обязательное поле',
  email: (v: string) => /.+@.+\..+/.test(v) || 'Некорректный email',
  minLength: (min: number) => (v: string) => v.length >= min || `Минимум ${min} символов`,
}

async function handleLogin() {
  const { valid } = await loginFormRef.value!.validate()
  if (!valid) return
  loading.value = true
  error.value = ''
  try {
    await auth.login(loginForm)
    router.push('/books')
  } catch (e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Ошибка входа'
    } else {
      error.value = 'Ошибка входа'
    }
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  const { valid } = await registerFormRef.value!.validate()
  if (!valid) return
  loading.value = true
  error.value = ''
  try {
    await auth.register(registerForm)
    router.push('/books')
  } catch (e: unknown) {
    if (e && typeof e === 'object' && 'response' in e) {
      const axiosError = e as { response?: { data?: { error?: string } } }
      error.value = axiosError.response?.data?.error || 'Ошибка регистрации'
    } else {
      error.value = 'Ошибка регистрации'
    }
  } finally {
    loading.value = false
  }
}
</script>
