import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createVuetify } from 'vuetify'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import LoginView from '../LoginView.vue'

vi.mock('@/api/auth', () => ({
  login: vi.fn(),
  register: vi.fn(),
  refresh: vi.fn(),
  logout: vi.fn(),
}))

const vuetify = createVuetify()
const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', component: { template: '<div />' } },
    { path: '/login', component: { template: '<div />' } },
    { path: '/books', component: { template: '<div />' } },
  ],
})

function mountLogin() {
  return mount(LoginView, {
    global: {
      plugins: [vuetify, createPinia(), router],
    },
  })
}

describe('LoginView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders login form by default', () => {
    const wrapper = mountLogin()
    expect(wrapper.text()).toContain('Вход')
    expect(wrapper.text()).toContain('Регистрация')
  })

  it('shows email and password fields', () => {
    const wrapper = mountLogin()
    const inputs = wrapper.findAll('input')
    expect(inputs.length).toBeGreaterThanOrEqual(2)
  })

  it('has login button', () => {
    const wrapper = mountLogin()
    const btn = wrapper.findAll('button').find(b => b.text().includes('Войти'))
    expect(btn).toBeTruthy()
  })

  it('has register tab', () => {
    const wrapper = mountLogin()
    expect(wrapper.text()).toContain('Регистрация')
  })

  it('renders login and register tabs', () => {
    const wrapper = mountLogin()
    const tabs = wrapper.findAll('.v-tab')
    expect(tabs.length).toBe(2)
  })
})
