<template>
  <div class="space-y-6">
    <!-- Loading State -->
    <div v-if="loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
      <p class="text-slate-400 mt-2">Loading dashboard...</p>
    </div>

    <!-- Stats Cards -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
      <Card class="hover:shadow-lg transition-all transform hover:-translate-y-1 animate-fade-in">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-400">Total Requests</p>
            <p class="text-3xl font-bold text-white mt-2">{{ stats.totalRequests.toLocaleString() }}</p>
            <p v-if="stats.requestGrowth !== 0" class="text-sm mt-2" :class="stats.requestGrowth > 0 ? 'text-green-400' : 'text-red-400'">
              <span class="inline-flex items-center">
                <TrendingUp class="w-4 h-4 mr-1" />
                {{ stats.requestGrowth > 0 ? '+' : '' }}{{ stats.requestGrowth }}% this month
              </span>
            </p>
          </div>
          <div class="w-16 h-16 bg-blue-500/20 rounded-full flex items-center justify-center">
            <BarChart3 class="w-8 h-8 text-blue-400" />
          </div>
        </div>
      </Card>

      <Card class="hover:shadow-lg transition-all transform hover:-translate-y-1 animate-slide-up" style="animation-delay: 0.1s">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-400">Active Tunnels</p>
            <p class="text-3xl font-bold text-white mt-2">{{ stats.activeTunnels }}</p>
            <p class="text-sm text-slate-400 mt-2">
              {{ stats.totalTunnels }} total
            </p>
          </div>
          <div class="w-16 h-16 bg-purple-500/20 rounded-full flex items-center justify-center">
            <Network class="w-8 h-8 text-purple-400" />
          </div>
        </div>
      </Card>

      <Card class="hover:shadow-lg transition-all transform hover:-translate-y-1 animate-slide-up" style="animation-delay: 0.2s">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-400">API Keys</p>
            <p class="text-3xl font-bold text-white mt-2">{{ stats.apiKeys }}</p>
            <p class="text-sm text-slate-400 mt-2">
              {{ stats.activeApiKeys }} active
            </p>
          </div>
          <div class="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center">
            <Key class="w-8 h-8 text-green-400" />
          </div>
        </div>
      </Card>

      <Card class="hover:shadow-lg transition-all transform hover:-translate-y-1 animate-slide-up" style="animation-delay: 0.3s">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm font-medium text-slate-400">Total Cost</p>
            <p class="text-3xl font-bold text-white mt-2">${{ stats.totalCost.toFixed(2) }}</p>
            <p class="text-sm text-slate-400 mt-2">
              This month
            </p>
          </div>
          <div class="w-16 h-16 bg-yellow-500/20 rounded-full flex items-center justify-center">
            <DollarSign class="w-8 h-8 text-yellow-400" />
          </div>
        </div>
      </Card>
    </div>

    <!-- Quick Actions -->
    <Card>
      <h2 class="text-xl font-semibold text-white mb-4">Quick Actions</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <router-link
          to="/dashboard/api-keys"
          class="flex items-center space-x-3 p-4 border-2 border-dashed border-slate-700/50 rounded-lg hover:border-blue-500 transition-colors group"
        >
          <div class="w-10 h-10 bg-blue-500/20 rounded-lg flex items-center justify-center group-hover:bg-blue-500/30 transition-colors">
            <Plus class="w-5 h-5 text-blue-400" />
          </div>
          <div>
            <p class="font-medium text-white">Create API Key</p>
            <p class="text-sm text-slate-400">Generate a new API key</p>
          </div>
        </router-link>

        <router-link
          to="/dashboard/tunnels"
          class="flex items-center space-x-3 p-4 border-2 border-dashed border-slate-700/50 rounded-lg hover:border-purple-500 transition-colors group"
        >
          <div class="w-10 h-10 bg-purple-500/20 rounded-lg flex items-center justify-center group-hover:bg-purple-500/30 transition-colors">
            <Network class="w-5 h-5 text-purple-400" />
          </div>
          <div>
            <p class="font-medium text-white">Create Tunnel</p>
            <p class="text-sm text-slate-400">Expose your local server</p>
          </div>
        </router-link>

        <router-link
          to="/dashboard/domains"
          class="flex items-center space-x-3 p-4 border-2 border-dashed border-slate-700/50 rounded-lg hover:border-cyan-500 transition-colors group"
        >
          <div class="w-10 h-10 bg-cyan-500/20 rounded-lg flex items-center justify-center group-hover:bg-cyan-500/30 transition-colors">
            <Globe class="w-5 h-5 text-cyan-400" />
          </div>
          <div>
            <p class="font-medium text-white">Manage Domains</p>
            <p class="text-sm text-slate-400">Add custom domains</p>
          </div>
        </router-link>

        <router-link
          to="/dashboard/settings/provider-keys"
          class="flex items-center space-x-3 p-4 border-2 border-dashed border-slate-700/50 rounded-lg hover:border-green-500 transition-colors group"
        >
          <div class="w-10 h-10 bg-green-500/20 rounded-lg flex items-center justify-center group-hover:bg-green-500/30 transition-colors">
            <Key class="w-5 h-5 text-green-400" />
          </div>
          <div>
            <p class="font-medium text-white">Add Provider Key</p>
            <p class="text-sm text-slate-400">Configure AI provider keys</p>
          </div>
        </router-link>

        <a
          href="https://polar.sh/uniroute/donate"
          target="_blank"
          rel="noopener noreferrer"
          class="flex items-center space-x-3 p-4 border-2 border-dashed border-slate-700/50 rounded-lg hover:border-pink-500 transition-colors group"
        >
          <div class="w-10 h-10 bg-pink-500/20 rounded-lg flex items-center justify-center group-hover:bg-pink-500/30 transition-colors">
            <Heart class="w-5 h-5 text-pink-400" />
          </div>
          <div>
            <p class="font-medium text-white">Support Us</p>
            <p class="text-sm text-slate-400">Donate via Polar.sh</p>
          </div>
        </a>
      </div>
    </Card>

    <!-- Tunnel Chart (Admin Only) -->
    <Card v-if="isAdmin">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-semibold text-white">Active Tunnels Over Time</h2>
        <div class="flex items-center space-x-4 text-sm text-slate-400">
          <button
            @click="chartHours = 6"
            :class="chartHours === 6 ? 'text-purple-400 font-semibold' : 'hover:text-white'"
            class="transition-colors"
          >
            6h
          </button>
          <button
            @click="chartHours = 24"
            :class="chartHours === 24 ? 'text-purple-400 font-semibold' : 'hover:text-white'"
            class="transition-colors"
          >
            24h
          </button>
          <button
            @click="chartHours = 168"
            :class="chartHours === 168 ? 'text-purple-400 font-semibold' : 'hover:text-white'"
            class="transition-colors"
          >
            7d
          </button>
        </div>
      </div>
      <div v-if="tunnelChartLoading" class="flex items-center justify-center h-80">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-purple-500"></div>
      </div>
      <div v-else-if="tunnelChartData.length === 0" class="flex items-center justify-center h-80 text-slate-400">
        <p>No tunnel data available</p>
      </div>
      <div v-else class="w-full">
        <TunnelChart 
          :data="tunnelChartData" 
          :width="Math.max(800, Math.min(1200, tunnelChartData.length * 60))" 
          :height="400"
          :hours="chartHours"
        />
      </div>
    </Card>

    <!-- Recent Activity & Provider Distribution -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Recent Activity -->
      <Card>
        <h2 class="text-xl font-semibold text-white mb-4">Recent Activity</h2>
        <div class="space-y-4">
          <div
            v-for="(activity, index) in recentActivity"
            :key="index"
            class="flex items-start space-x-3 p-3 rounded-lg hover:bg-slate-800/60 transition-colors"
          >
            <div class="w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0" :class="activity.iconBg">
              <span class="text-lg">{{ activity.icon }}</span>
            </div>
            <div class="flex-1 min-w-0">
              <p class="text-sm font-medium text-white">{{ activity.title }}</p>
              <p class="text-xs text-slate-400">{{ activity.time }}</p>
            </div>
          </div>
          <div v-if="recentActivity.length === 0" class="text-center py-8">
            <p class="text-slate-400">No recent activity</p>
          </div>
        </div>
      </Card>

      <!-- Provider Distribution -->
      <Card>
        <h2 class="text-xl font-semibold text-white mb-4">Provider Usage</h2>
        <div class="space-y-4">
          <div
            v-for="provider in providerUsage"
            :key="provider.name"
            class="space-y-2"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-2">
                <span class="text-xl">{{ provider.icon }}</span>
                <span class="font-medium text-white">{{ provider.name }}</span>
              </div>
              <span class="text-sm font-semibold text-white">{{ provider.percentage }}%</span>
            </div>
            <div class="w-full bg-slate-700/50 rounded-full h-2">
              <div
                class="h-2 rounded-full transition-all duration-500"
                :class="provider.color"
                :style="{ width: `${provider.percentage}%` }"
              ></div>
            </div>
            <div class="flex items-center justify-between text-xs text-slate-400">
              <span>{{ provider.requests }} requests</span>
              <span>${{ provider.cost.toFixed(2) }}</span>
            </div>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import Card from '@/components/ui/Card.vue'
import TunnelChart from '@/components/charts/TunnelChart.vue'
import {
  BarChart3,
  Network,
  Key,
  DollarSign,
  Plus,
  TrendingUp,
  Heart,
  Globe
} from 'lucide-vue-next'
import { dashboardApi } from '@/services/api/dashboard'
import { tunnelsApi } from '@/services/api/tunnels'
import { useToast } from '@/composables/useToast'
import { useAuthStore } from '@/stores/auth'

const { showToast } = useToast()
const authStore = useAuthStore()
const isAdmin = computed(() => authStore.hasRole('admin'))

interface Activity {
  icon: string
  iconBg: string
  title: string
  time: string
}

interface ProviderUsage {
  name: string
  icon: string
  percentage: number
  requests: number
  cost: number
  color: string
}

const loading = ref(false)
const stats = ref({
  totalRequests: 0,
  requestGrowth: 0,
  activeTunnels: 0,
  totalTunnels: 0,
  apiKeys: 0,
  activeApiKeys: 0,
  totalCost: 0
})

const recentActivity = ref<Activity[]>([])
const providerUsage = ref<ProviderUsage[]>([])
const tunnelChartData = ref<Array<{ time: string; active_tunnels: number; total_tunnels: number }>>([])
const tunnelChartLoading = ref(false)

// Load chart hours from localStorage, default to 24 if not found
const getStoredChartHours = (): number => {
  const stored = localStorage.getItem('uniroute_chart_hours')
  if (stored) {
    const hours = parseInt(stored, 10)
    // Validate that it's one of the allowed values (6, 24, or 168)
    if (hours === 6 || hours === 24 || hours === 168) {
      return hours
    }
  }
  return 24 // Default to 24h
}

const chartHours = ref(getStoredChartHours())
let chartUpdateInterval: number | null = null

// Save chart hours to localStorage whenever it changes
watch(chartHours, (newHours) => {
  localStorage.setItem('uniroute_chart_hours', newHours.toString())
  if (isAdmin.value) {
    loadTunnelChart()
  }
})

onMounted(async () => {
  await loadDashboardData()
  if (isAdmin.value) {
    await loadTunnelChart()
    // Update chart every 30 seconds (only for admins)
    chartUpdateInterval = window.setInterval(() => {
      if (isAdmin.value) {
        loadTunnelChart()
      }
    }, 30000)
  }
})

onUnmounted(() => {
  if (chartUpdateInterval !== null) {
    clearInterval(chartUpdateInterval)
  }
})

watch(isAdmin, (newVal) => {
  if (newVal) {
    loadTunnelChart()
  } else {
    tunnelChartData.value = []
  }
})

const loadTunnelChart = async () => {
  tunnelChartLoading.value = true
  try {
    // For time-series views with trend/slope:
    // - 6h: 30-minute intervals (0.5 hours) for trend visualization
    // - 24h: 1-hour intervals for trend visualization
    // - 7d: Daily aggregation (interval=24)
    let interval: number | undefined
    if (chartHours.value >= 168) {
      interval = 24 // Daily intervals for 7d view
    } else if (chartHours.value === 6) {
      interval = 0.5 // 30-minute intervals for 6h view
    } else {
      interval = 1.0 // 1-hour intervals for 24h view
    }
    console.log(`Loading tunnel chart: hours=${chartHours.value}, interval=${interval}`)
    const stats = await tunnelsApi.getStats(chartHours.value, interval, true)
    console.log('Tunnel stats received:', stats) // Debug log
    console.log(`Data points: ${stats.data?.length || 0}`)
    tunnelChartData.value = stats.data || []
    if (stats.data && stats.data.length === 0) {
      console.warn('Tunnel stats API returned empty data array')
    }
  } catch (error: any) {
    console.error('Failed to load tunnel chart data:', error)
    console.error('Error details:', error.response?.data || error.message)
    tunnelChartData.value = []
    // Don't show toast for chart errors to avoid spam
  } finally {
    tunnelChartLoading.value = false
  }
}

const loadDashboardData = async () => {
  loading.value = true
  try {
    const data = await dashboardApi.getOverview()
    
    stats.value = {
      totalRequests: data.total_requests,
      requestGrowth: data.request_growth,
      activeTunnels: data.active_tunnels,
      totalTunnels: data.total_tunnels,
      apiKeys: data.api_keys,
      activeApiKeys: data.active_api_keys,
      totalCost: data.total_cost
    }

    // Map recent activity
    recentActivity.value = data.recent_activity.map(activity => ({
      icon: activity.icon,
      iconBg: getIconBg(activity.type),
      title: activity.title,
      time: activity.time
    }))

    // Map provider usage with colors
    const colorMap: Record<string, string> = {
      'Openai': 'bg-blue-600',
      'Anthropic': 'bg-purple-600',
      'Google': 'bg-green-600',
      'Google-ai': 'bg-green-600',
      'Gemini': 'bg-yellow-600',
      'Meta-llama': 'bg-orange-600',
      'Llama': 'bg-orange-600',
      'Mistral': 'bg-indigo-600'
    }

    providerUsage.value = data.provider_usage.map(provider => ({
      name: provider.name,
      icon: provider.icon,
      percentage: provider.percentage,
      requests: provider.requests,
      cost: provider.cost,
      color: colorMap[provider.name] || 'bg-gray-600'
    }))

    // Animate numbers
    animateNumbers()
  } catch (error: any) {
    console.error('Failed to load dashboard data:', error)
    showToast('Failed to load dashboard data', 'error')
  } finally {
    loading.value = false
  }
}

function getIconBg(type: string): string {
  const bgMap: Record<string, string> = {
    api_key: 'bg-blue-100 dark:bg-blue-900/30',
    tunnel: 'bg-purple-100 dark:bg-purple-900/30',
    request: 'bg-green-100 dark:bg-green-900/30'
  }
  return bgMap[type] || 'bg-gray-100 dark:bg-gray-900/30'
}

const animateNumbers = () => {
  // Animate stats counters
  const duration = 1000
  const steps = 60
  const stepDuration = duration / steps

  Object.keys(stats.value).forEach(key => {
    const target = stats.value[key as keyof typeof stats.value]
    if (typeof target === 'number' && target > 0) {
      let current = 0
      const increment = target / steps
      const timer = setInterval(() => {
        current += increment
        if (current >= target) {
          current = target
          clearInterval(timer)
        }
        stats.value[key as keyof typeof stats.value] = Math.floor(current) as any
      }, stepDuration)
    }
  })
}
</script>
