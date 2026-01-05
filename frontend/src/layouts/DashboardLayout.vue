<template>
  <div class="dark min-h-screen bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950">
    <!-- Mobile Menu Button -->
    <button
      @click="mobileMenuOpen = !mobileMenuOpen"
      class="lg:hidden fixed top-4 left-4 z-50 p-2 rounded-lg bg-slate-800/80 backdrop-blur-sm border border-slate-700/50 text-white"
    >
      <Menu v-if="!mobileMenuOpen" class="w-6 h-6" />
      <X v-else class="w-6 h-6" />
    </button>

    <!-- Sidebar Navigation -->
    <aside
      :class="[
        'fixed inset-y-0 left-0 w-64 bg-slate-900/80 backdrop-blur-xl border-r border-slate-800/50 z-40 transform transition-transform duration-300 ease-in-out',
        mobileMenuOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'
      ]"
    >
      <div class="flex flex-col h-full">
        <!-- Logo -->
        <div class="p-6 border-b border-slate-800/50">
          <div class="flex items-center space-x-2">
            <div class="w-10 h-10 bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500 rounded-lg flex items-center justify-center shadow-lg shadow-blue-500/20">
              <span class="text-white font-bold text-xl">U</span>
            </div>
            <span class="text-xl font-bold text-white">UniRoute</span>
          </div>
        </div>

        <!-- Navigation -->
        <nav class="flex-1 p-4 space-y-1 overflow-y-auto scrollbar-thin">
          <router-link
            v-for="item in navItems"
            :key="item.path"
            :to="item.path"
            class="flex items-center space-x-3 px-4 py-3 rounded-lg text-slate-300 hover:bg-slate-800/60 transition-colors group"
            :class="{ 'bg-blue-500/20 text-blue-400 font-semibold': isActive(item.path) }"
          >
            <component :is="item.icon" class="w-5 h-5" />
            <span>{{ item.label }}</span>
          </router-link>
        </nav>

        <!-- User Section -->
        <div class="p-4 border-t border-slate-800/50">
          <div class="flex items-center space-x-3 mb-4">
            <div class="w-10 h-10 bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500 rounded-full flex items-center justify-center shadow-lg shadow-blue-500/20">
              <span class="text-white font-semibold text-sm">
                {{ userInitials }}
              </span>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-white truncate">
                {{ userEmail }}
              </p>
              <p class="text-xs text-slate-400 truncate">
                {{ userRole }}
              </p>
            </div>
          </div>
          <button
            @click="handleLogout"
            class="w-full flex items-center space-x-3 px-4 py-2 rounded-lg text-slate-300 hover:bg-slate-800/60 transition-colors"
          >
            <LogOut class="w-5 h-5" />
            <span>Logout</span>
          </button>
        </div>
      </div>
    </aside>

    <!-- Mobile Overlay -->
    <div
      v-if="mobileMenuOpen"
      @click="mobileMenuOpen = false"
      class="lg:hidden fixed inset-0 bg-black/50 z-30"
    ></div>

    <!-- Main Content -->
    <main class="lg:ml-64 min-h-screen">
      <!-- Top Bar -->
      <header class="sticky top-0 z-40 bg-slate-900/80 backdrop-blur-xl border-b border-slate-800/50">
        <div class="px-4 sm:px-6 py-4 flex items-center justify-between">
          <div>
            <h1 class="text-2xl font-bold text-white">
              {{ pageTitle }}
            </h1>
            <p v-if="pageDescription" class="text-sm text-slate-400 mt-1">
              {{ pageDescription }}
            </p>
          </div>
          <div class="flex items-center space-x-4">
            <!-- Notifications -->
            <button class="p-2 rounded-lg text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors relative">
              <Bell class="w-5 h-5" />
              <span class="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
            </button>
          </div>
        </div>
      </header>

      <!-- Page Content -->
      <div class="p-4 sm:p-6">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import {
  LayoutDashboard,
  Key,
  Network,
  BarChart3,
  Settings,
  LogOut,
  Bell,
  Menu,
  X,
  Webhook,
  AlertTriangle,
  Mail,
  Route,
  Users
} from 'lucide-vue-next'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const mobileMenuOpen = ref(false)

const userEmail = computed(() => authStore.user?.email || 'user@example.com')
const userRoles = computed(() => authStore.user?.roles || ['user'])
const userRole = computed(() => {
  // For display purposes, show primary role (admin if present, otherwise user)
  if (userRoles.value.includes('admin')) return 'admin'
  return 'user'
})
const isAdmin = computed(() => userRoles.value.includes('admin'))
const userInitials = computed(() => {
  const email = userEmail.value
  const parts = email.split('@')[0].split('.')
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase()
  }
  return email.substring(0, 2).toUpperCase()
})

const pageTitle = computed(() => {
  const titles: Record<string, string> = {
    dashboard: 'Dashboard',
    'api-keys': 'API Keys',
    tunnels: 'Tunnels',
    analytics: 'Analytics',
    'webhook-testing': 'Webhook Testing',
    'settings-profile': 'Profile Settings',
    'settings-provider-keys': 'Provider Keys',
    'admin-users': 'User Management',
    'admin-email': 'Email Configuration',
    'admin-provider-keys': 'Provider Keys Management',
    'admin-routing': 'Routing Strategy',
    'admin-errors': 'Error Logs'
  }
  return titles[route.name as string] || 'Dashboard'
})

const pageDescription = computed(() => {
  const descriptions: Record<string, string> = {
    dashboard: 'Overview of your UniRoute usage and activity',
    'api-keys': 'Manage your API keys and access tokens',
    tunnels: 'View and manage your active tunnels',
    analytics: 'Track usage, costs, and performance metrics',
    'webhook-testing': 'Inspect, replay, and test webhook requests',
    'admin-users': 'Manage users and their roles',
    'admin-email': 'View SMTP configuration and test email delivery',
    'admin-provider-keys': 'Manage system-wide provider API keys',
    'admin-routing': 'Configure how UniRoute selects AI providers',
    'admin-errors': 'Monitor and manage application errors'
  }
  return descriptions[route.name as string] || ''
})

const navItems = computed(() => {
  const items = [
    {
      path: '/dashboard',
      label: 'Dashboard',
      icon: LayoutDashboard
    },
    {
      path: '/dashboard/api-keys',
      label: 'API Keys',
      icon: Key
    },
    {
      path: '/dashboard/tunnels',
      label: 'Tunnels',
      icon: Network
    },
    {
      path: '/dashboard/analytics',
      label: 'Analytics',
      icon: BarChart3
    },
    {
      path: '/dashboard/webhook-testing',
      label: 'Webhook Testing',
      icon: Webhook
    },
    {
      path: '/dashboard/settings/profile',
      label: 'Settings',
      icon: Settings
    }
  ]

  // Add admin routes only if user is admin
  if (isAdmin.value) {
    items.push(
      {
        path: '/dashboard/admin/users',
        label: 'User Management',
        icon: Users
      },
      {
        path: '/dashboard/admin/email',
        label: 'Email Config',
        icon: Mail
      },
      {
        path: '/dashboard/admin/routing',
        label: 'Routing Strategy',
        icon: Route
      },
      {
        path: '/dashboard/errors',
        label: 'Error Logs',
        icon: AlertTriangle
      }
    )
  }

  return items
})

const isActive = (path: string) => {
  // Exact match
  if (route.path === path) return true
  
  // For dashboard root, only match exact path (not child routes)
  if (path === '/dashboard') {
    return route.path === '/dashboard' || route.path === '/dashboard/'
  }
  
  // For other paths, match if route starts with the path
  return route.path.startsWith(path + '/')
}

const handleLogout = async () => {
  await authStore.logout()
  router.push('/login')
}
</script>
