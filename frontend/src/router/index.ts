import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { updateDocumentHead } from '@/utils/head'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'landing',
    component: () => import('@/views/LandingPage.vue'),
    meta: {
      requiresGuest: true,
      title: 'AI Gateway & Secure Tunneling',
      description: 'One unified API for all AI models. Plus secure tunneling for any service. Route, secure, and manage traffic to any LLM—cloud or local.'
    }
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { requiresGuest: true, title: 'Sign in', description: 'Sign in to your UniRoute account.' }
  },
  {
    path: '/register',
    name: 'register',
    component: () => import('@/views/auth/Register.vue'),
    meta: { requiresGuest: true, title: 'Get started', description: 'Create your UniRoute account and start using the AI gateway and secure tunneling.' }
  },
  {
    path: '/forgot-password',
    name: 'forgot-password',
    component: () => import('@/views/auth/ForgotPassword.vue'),
    meta: { requiresGuest: true, title: 'Forgot password', description: 'Reset your UniRoute account password.' }
  },
  {
    path: '/verify-email',
    name: 'verify-email',
    component: () => import('@/views/auth/VerifyEmail.vue'),
    meta: { requiresGuest: true, title: 'Verify email', description: 'Verify your email address.' }
  },
  {
    path: '/download',
    name: 'download',
    component: () => import('@/views/Download.vue'),
    meta: { requiresGuest: true, title: 'Download CLI', description: 'Download the UniRoute CLI for Windows, macOS, and Linux. Secure tunneling and AI gateway from the terminal.' }
  },
  {
    path: '/docs',
    name: 'docs',
    redirect: '/docs/introduction',
    meta: { requiresGuest: true, title: 'Documentation', description: 'UniRoute documentation: installation, authentication, tunnels, API reference, and deployment.' }
  },
  {
    path: '/docs/:path(.*)',
    name: 'docs-page',
    component: () => import('@/views/Docs.vue'),
    meta: { requiresGuest: true, title: 'Documentation', description: 'UniRoute documentation and guides.' }
  },
  {
    path: '/security',
    name: 'security',
    component: () => import('@/views/Security.vue'),
    meta: { title: 'Security', description: 'UniRoute security overview and best practices.' }
  },
  {
    path: '/terms',
    name: 'terms',
    component: () => import('@/views/Terms.vue'),
    meta: { title: 'Terms of Service', description: 'UniRoute terms of service.' }
  },
  {
    path: '/privacy',
    name: 'privacy',
    component: () => import('@/views/Privacy.vue'),
    meta: { title: 'Privacy Policy', description: 'UniRoute privacy policy.' }
  },
  {
    path: '/pricing',
    name: 'pricing',
    component: () => import('@/views/Pricing.vue'),
    meta: { requiresGuest: true, title: 'Pricing', description: 'UniRoute pricing plans and features.' }
  },
  {
    path: '/dashboard',
    component: () => import('@/layouts/DashboardLayout.vue'),
    meta: { requiresAuth: true, title: 'Dashboard' },
    children: [
      {
        path: '',
        name: 'dashboard',
        component: () => import('@/views/Dashboard.vue'),
        meta: { title: 'Dashboard', description: 'UniRoute dashboard and overview.' }
      },
      {
        path: 'api-keys',
        name: 'api-keys',
        component: () => import('@/views/ApiKeys.vue'),
        meta: { permission: 'api-keys:read', title: 'API Keys', description: 'Manage your UniRoute API keys.' }
      },
      {
        path: 'tunnels',
        name: 'tunnels',
        component: () => import('@/views/Tunnels.vue'),
        meta: { permission: 'tunnels:read', title: 'Tunnels', description: 'Manage your secure tunnels.' }
      },
      {
        path: 'domains',
        name: 'domains',
        component: () => import('@/views/Domains.vue'),
        meta: { permission: 'tunnels:read', title: 'Domains', description: 'Custom domains for tunnels.' }
      },
      {
        path: 'tunnels/:id',
        name: 'tunnel-detail',
        component: () => import('@/views/tunnels/TunnelDetail.vue'),
        meta: { permission: 'tunnels:read', title: 'Tunnel details', description: 'Tunnel details and stats.' }
      },
      {
        path: 'chat',
        name: 'chat',
        component: () => import('@/views/Chat.vue'),
        meta: { permission: 'chat:use', title: 'AI Chat', description: 'Chat with AI models via UniRoute.' }
      },
      {
        path: 'analytics',
        name: 'analytics',
        component: () => import('@/views/Analytics.vue'),
        meta: { permission: 'analytics:read', title: 'Analytics', description: 'Request analytics and usage.' }
      },
      {
        path: 'webhook-testing',
        name: 'webhook-testing',
        component: () => import('@/views/WebhookTesting.vue'),
        meta: { permission: 'tunnels:read', title: 'Webhook testing', description: 'Test webhook deliveries.' }
      },
      {
        path: 'settings',
        name: 'settings',
        component: () => import('@/views/settings/Settings.vue'),
        meta: { title: 'Settings' },
        children: [
          {
            path: 'profile',
            name: 'settings-profile',
            component: () => import('@/views/settings/Profile.vue'),
            meta: { title: 'Profile', description: 'Update your UniRoute profile.' }
          },
          {
            path: 'provider-keys',
            name: 'settings-provider-keys',
            component: () => import('@/views/settings/ProviderKeys.vue'),
            meta: { permission: 'provider-keys:manage', title: 'Provider keys', description: 'Manage provider API keys.' }
          },
          {
            path: 'routing-strategy',
            name: 'settings-routing-strategy',
            component: () => import('@/views/settings/RoutingStrategy.vue'),
            meta: { title: 'Routing strategy', description: 'Configure routing strategy.' }
          },
          {
            path: 'custom-rules',
            name: 'settings-custom-rules',
            component: () => import('@/views/settings/CustomRules.vue'),
            meta: { title: 'Custom rules', description: 'Custom routing rules.' }
          }
        ]
      },
      {
        path: 'admin',
        meta: { title: 'Admin' },
        children: [
          {
            path: 'users',
            name: 'admin-users',
            component: () => import('@/views/admin/UserManagement.vue'),
            meta: { permission: 'admin:users', title: 'User management', description: 'Manage users and roles.' }
          },
          {
            path: 'email',
            name: 'admin-email',
            component: () => import('@/views/admin/EmailConfig.vue'),
            meta: { permission: 'admin:email', title: 'Email config', description: 'Email configuration.' }
          },
          {
            path: 'provider-keys',
            name: 'admin-provider-keys',
            component: () => import('@/views/admin/AdminProviderKeys.vue'),
            meta: { permission: 'admin:provider-keys', title: 'Provider keys', description: 'Admin provider keys.' }
          },
          {
            path: 'routing',
            name: 'admin-routing',
            component: () => import('@/views/admin/RoutingStrategy.vue'),
            meta: { permission: 'admin:routing', title: 'Routing', description: 'Admin routing strategy.' }
          },
          {
            path: 'custom-rules',
            name: 'admin-custom-rules',
            component: () => import('@/views/admin/CustomRules.vue'),
            meta: { permission: 'admin:routing', title: 'Custom rules', description: 'Admin custom rules.' }
          },
          {
            path: 'errors',
            name: 'admin-errors',
            component: () => import('@/views/admin/ErrorLogs.vue'),
            meta: { permission: 'admin:errors', title: 'Error logs', description: 'Application error logs.' }
          }
        ]
      }
    ]
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/NotFound.vue'),
    meta: { title: 'Page not found', description: 'The page you are looking for could not be found.' }
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

// SEO: update document title, meta, canonical, and robots per route. Dashboard/admin = noindex, nofollow.
router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  const description = to.meta.description as string | undefined
  const isAppRoute = to.path.startsWith('/dashboard') || to.path.startsWith('/admin') || to.meta.requiresAuth === true
  updateDocumentHead(title, description, to.fullPath, isAppRoute)
})

export default router

