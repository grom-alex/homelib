import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('@/pages/LoginPage.vue'),
      meta: { guest: true },
    },
    {
      path: '/',
      redirect: '/books',
    },
    {
      path: '/books',
      name: 'catalog',
      component: () => import('@/pages/CatalogPage.vue'),
    },
    {
      path: '/books/:id',
      name: 'book',
      component: () => import('@/pages/BookPage.vue'),
    },
    {
      path: '/authors',
      name: 'authors',
      component: () => import('@/pages/AuthorsPage.vue'),
    },
    {
      path: '/authors/:id',
      name: 'author',
      component: () => import('@/pages/AuthorPage.vue'),
    },
    {
      path: '/genres',
      name: 'genres',
      component: () => import('@/pages/GenresPage.vue'),
    },
    {
      path: '/series',
      name: 'series',
      component: () => import('@/pages/SeriesPage.vue'),
    },
    {
      path: '/admin/import',
      name: 'admin-import',
      component: () => import('@/pages/AdminImportPage.vue'),
      meta: { admin: true },
    },
  ],
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.guest) return true
  if (!auth.isAuthenticated) return { name: 'login' }
  if (to.meta.admin && !auth.isAdmin) return { name: 'catalog' }
  return true
})

export default router
