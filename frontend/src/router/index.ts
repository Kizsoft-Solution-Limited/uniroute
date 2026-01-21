import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'landing',
    component: () => import('@/views/LandingPage.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/auth/Register.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/forgot-password',
    name: 'forgot-password',
    component: () => import('@/views/auth/ForgotPassword.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/verify-email',
    name: 'verify-email',
    component: () => import('@/views/auth/VerifyEmail.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/download',
    name: 'download',
    component: () => import('@/views/Download.vue'),
    meta: { requiresGuest: true }
  },
  {
    path: '/security',
    name: 'security',
    component: () => import('@/views/Security.vue')
  },
  {
    path: '/terms',
    name: 'terms',
    component: () => import('@/views/Terms.vue')
  },
  {
    path: '/privacy',
    name: 'privacy',
    component: () => import('@/views/Privacy.vue')
  },
  {
    path: '/dashboard',
    component: () => import('@/layouts/DashboardLayout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/views/Dashboard.vue')
      },
      {
        path: 'api-keys',
        name: 'api-keys',
        component: () => import('@/views/ApiKeys.vue'),
        meta: { permission: 'api-keys:read' }
      },
      {
        path: 'tunnels',
        name: 'tunnels',
        component: () => import('@/views/Tunnels.vue'),
        meta: { permission: 'tunnels:read' }
      },
      {
        path: 'tunnels/:id',
        name: 'tunnel-detail',
        component: () => import('@/views/tunnels/TunnelDetail.vue'),
        meta: { permission: 'tunnels:read' }
      },
      {
        path: 'chat',
        name: 'chat',
        component: () => import('@/views/Chat.vue'),
        meta: { permission: 'chat:use' }
      },
      {
        path: 'analytics',
        name: 'analytics',
        component: () => import('@/views/Analytics.vue'),
        meta: { permission: 'analytics:read' }
      },
      {
        path: 'webhook-testing',
        name: 'webhook-testing',
        component: () => import('@/views/WebhookTesting.vue'),
        meta: { permission: 'tunnels:read' }
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/settings/Settings.vue'),
        children: [
          {
            path: 'profile',
            name: 'settings-profile',
            component: () => import('@/views/settings/Profile.vue')
          },
          {
            path: 'provider-keys',
            name: 'settings-provider-keys',
            component: () => import('@/views/settings/ProviderKeys.vue'),
            meta: { permission: 'provider-keys:manage' }
          },
          {
            path: 'routing-strategy',
            name: 'settings-routing-strategy',
            component: () => import('@/views/settings/RoutingStrategy.vue')
          },
          {
            path: 'custom-rules',
            name: 'settings-custom-rules',
            component: () => import('@/views/settings/CustomRules.vue')
          }
        ]
      },
      {
        path: 'admin',
        children: [
          {
            path: 'users',
            name: 'admin-users',
            component: () => import('@/views/admin/UserManagement.vue'),
            meta: { permission: 'admin:users' }
          },
          {
            path: 'email',
            name: 'admin-email',
            component: () => import('@/views/admin/EmailConfig.vue'),
            meta: { permission: 'admin:email' }
          },
          {
            path: 'provider-keys',
            name: 'admin-provider-keys',
            component: () => import('@/views/admin/AdminProviderKeys.vue'),
            meta: { permission: 'admin:provider-keys' }
          },
          {
            path: 'routing',
            name: 'admin-routing',
            component: () => import('@/views/admin/RoutingStrategy.vue'),
            meta: { permission: 'admin:routing' }
          },
          {
            path: 'custom-rules',
            name: 'admin-custom-rules',
            component: () => import('@/views/admin/CustomRules.vue'),
            meta: { permission: 'admin:routing' }
          },
          {
            path: 'errors',
            name: 'admin-errors',
            component: () => import('@/views/admin/ErrorLogs.vue'),
            meta: { permission: 'admin:errors' }
          }
        ]
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/NotFound.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to, from, savedPosition) {
    // If there's a saved position (e.g., browser back/forward), use it
    if (savedPosition) {
      return savedPosition
    }
    // If there's a hash in the URL, scroll to that element
    if (to.hash) {
      return {
        el: to.hash,
        behavior: 'smooth',
        top: 80 // Offset for fixed header
      }
    }
    // Otherwise, scroll to top
    return { top: 0, behavior: 'instant' }
  }
})

// Navigation guards
router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()
  
  // Skip auth check for login/register pages to avoid redirect loops
  if (to.name === 'login' || to.name === 'register' || to.name === 'forgot-password' || to.name === 'verify-email' || to.name === 'oauth-callback') {
    next()
    return
  }
  
  // Check if we have a token but no user - need to verify auth
  const hasToken = authStore.token || localStorage.getItem('auth_token') || sessionStorage.getItem('auth_token')
  const hasUser = !!authStore.user
  
  // If we have a token but no user, try to check auth
  if (hasToken && !hasUser) {
    try {
      const authResult = await authStore.checkAuth()
      // If checkAuth failed, it will clear the token
      if (!authResult) {
        // Auth check failed, redirect to login if route requires auth
        if (to.meta.requiresAuth) {
          next({ name: 'login', query: { redirect: to.fullPath } })
          return
        }
      }
    } catch (err) {
      // Auth check failed, redirect to login if route requires auth
      if (to.meta.requiresAuth) {
        next({ name: 'login', query: { redirect: to.fullPath } })
        return
      }
    }
  }
  
  const isAuthenticated = authStore.isAuthenticated

  // Check if route requires authentication
  if (to.meta.requiresAuth && !isAuthenticated) {
    next({ name: 'login', query: { redirect: to.fullPath } })
    return
  }

  // Check if route requires guest (not authenticated)
  if (to.meta.requiresGuest && isAuthenticated) {
    next({ name: 'dashboard' })
    return
  }

  // Check permissions
  if (to.meta.permission && !authStore.hasPermission(to.meta.permission as string)) {
    next({ name: 'dashboard' })
    return
  }

  next()
})

export default router

