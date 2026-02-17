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
      path: '/authors',
      name: 'authors',
      component: () => import('@/views/AuthorsView.vue'),
    },
    {
      path: '/authors/:id',
      name: 'author',
      component: () => import('@/views/AuthorView.vue'),
    },
    {
      path: '/genres',
      name: 'genres',
      component: () => import('@/views/GenresView.vue'),
    },
    {
      path: '/series',
      name: 'series',
      component: () => import('@/views/SeriesView.vue'),
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
  if (!auth.initialized) {
    await auth.init()
  }
  if (to.meta.guest) {
    if (to.name === 'login' && auth.isAuthenticated) return { name: 'catalog' }
    return true
  }
  if (!auth.isAuthenticated) return { name: 'login' }
  if (to.meta.admin && !auth.isAdmin) return { name: 'catalog' }
  return true
})

export default router
