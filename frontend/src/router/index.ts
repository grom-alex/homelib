import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useParentalStore } from '@/stores/parental'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guest: true },
    },
    {
      path: '/',
      redirect: '/books',
    },
    {
      path: '/books',
      name: 'catalog',
      component: () => import('@/views/CatalogView.vue'),
    },
    {
      path: '/authors',
      redirect: '/books',
    },
    {
      path: '/genres',
      redirect: '/books',
    },
    {
      path: '/series',
      redirect: '/books',
    },
    {
      path: '/books/:id',
      name: 'book',
      component: () => import('@/views/BookView.vue'),
    },
    {
      path: '/books/:id/read',
      name: 'reader',
      component: () => import('@/views/ReaderView.vue'),
    },
    {
      path: '/admin/import',
      name: 'admin-import',
      component: () => import('@/views/AdminImportView.vue'),
      meta: { admin: true },
    },
    {
      path: '/admin/parental',
      name: 'admin-parental',
      component: () => import('@/views/AdminParentalView.vue'),
      meta: { admin: true },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
      meta: { guest: true },
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  const parental = useParentalStore()
  if (!auth.initialized) {
    await auth.init()
  }
  // Load parental status once after authentication
  if (auth.isAuthenticated && !parental.loaded) {
    parental.loadStatus()
  }
  if (to.meta.guest) {
    if (to.name === 'login' && auth.isAuthenticated) return { name: 'catalog' }
    return true
  }
  if (!auth.isAuthenticated) {
    // Safety net: if init failed but cookie might still be valid, try one more refresh
    // before giving up. Handles transient 503s during backend restart.
    try {
      await auth.refreshToken()
      return true
    } catch {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }
  if (to.meta.admin && !auth.isAdmin) return { name: 'catalog' }
  return true
})

export default router
