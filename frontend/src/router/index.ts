import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

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
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
      meta: { guest: true },
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  console.warn('[GUARD] to:', to.fullPath, 'initialized:', auth.initialized, 'authenticated:', auth.isAuthenticated)
  if (!auth.initialized) {
    await auth.init()
    console.warn('[GUARD] after init: authenticated:', auth.isAuthenticated, 'user:', auth.user?.email)
  }
  if (to.meta.guest) {
    if (to.name === 'login' && auth.isAuthenticated) {
      console.warn('[GUARD] guest route + authenticated → redirect to catalog')
      return { name: 'catalog' }
    }
    return true
  }
  if (!auth.isAuthenticated) {
    // Safety net: if init failed but cookie might still be valid, try one more refresh
    // before giving up. Handles race conditions and transient failures during init.
    console.warn('[GUARD] NOT authenticated, trying safety-net refresh')
    try {
      await auth.refreshToken()
      console.warn('[GUARD] safety-net refresh succeeded')
      return true
    } catch {
      console.warn('[GUARD] safety-net refresh also failed → redirect to login')
      return { name: 'login', query: { redirect: to.fullPath } }
    }
  }
  if (to.meta.admin && !auth.isAdmin) return { name: 'catalog' }
  return true
})

export default router
