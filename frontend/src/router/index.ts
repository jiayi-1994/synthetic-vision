import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('../views/Login.vue'),
    meta: { public: true },
  },
  {
    path: '/',
    name: 'dashboard',
    component: () => import('../views/Dashboard.vue'),
  },
  {
    path: '/gallery',
    name: 'gallery',
    component: () => import('../views/Gallery.vue'),
  },
  {
    path: '/marketplace',
    name: 'marketplace',
    component: () => import('../views/Marketplace.vue'),
  },
  {
    path: '/analytics',
    name: 'analytics',
    component: () => import('../views/Analytics.vue'),
  },
  {
    path: '/admin',
    name: 'admin',
    component: () => import('../views/Admin.vue'),
    meta: { admin: true },
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  const isPublic = to.meta.public === true

  // Already-authenticated users shouldn't see the login page.
  if (isPublic) {
    if (auth.isAuthenticated) {
      return { path: '/' }
    }
    return true
  }

  // Protected route but no token -> login.
  if (!auth.isAuthenticated) {
    return { path: '/login' }
  }

  // We have a token but no user object yet (e.g. fresh page load) -> hydrate.
  if (!auth.user) {
    try {
      await auth.fetchMe()
    } catch (e) {
      auth.logout()
      return { path: '/login' }
    }
  }

  // Admin-gated route.
  if (to.meta.admin === true && !auth.isAdmin) {
    return { path: '/' }
  }

  return true
})

export default router
