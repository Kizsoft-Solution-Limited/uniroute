<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white" id="page-title">Analytics</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          Track usage, costs, and performance metrics
        </p>
      </div>
      <div class="flex space-x-2">
        <select
          v-model="timeRange"
          @change="loadStats"
          class="px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          aria-label="Select time range"
        >
          <option value="7">Last 7 days</option>
          <option value="30">Last 30 days</option>
          <option value="90">Last 90 days</option>
        </select>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="text-center py-12" role="status" aria-live="polite">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" aria-hidden="true"></div>
      <p class="text-gray-500 dark:text-gray-400 mt-2">Loading analytics...</p>
    </div>

    <!-- Stats Cards -->
    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <Card>
        <div class="p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-600 dark:text-gray-400">Total Requests</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {{ stats?.total_requests?.toLocaleString() || '0' }}
              </p>
            </div>
            <div class="w-12 h-12 bg-blue-100 dark:bg-blue-900/30 rounded-lg flex items-center justify-center">
              <BarChart3 class="w-6 h-6 text-blue-600 dark:text-blue-400" />
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <div class="p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-600 dark:text-gray-400">Total Tokens</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {{ stats?.total_tokens?.toLocaleString() || '0' }}
              </p>
            </div>
            <div class="w-12 h-12 bg-green-100 dark:bg-green-900/30 rounded-lg flex items-center justify-center">
              <BarChart3 class="w-6 h-6 text-green-600 dark:text-green-400" />
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <div class="p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-600 dark:text-gray-400">Total Cost</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                ${{ stats?.total_cost?.toFixed(4) || '0.0000' }}
              </p>
            </div>
            <div class="w-12 h-12 bg-purple-100 dark:bg-purple-900/30 rounded-lg flex items-center justify-center">
              <BarChart3 class="w-6 h-6 text-purple-600 dark:text-purple-400" />
            </div>
          </div>
        </div>
      </Card>

      <Card>
        <div class="p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-gray-600 dark:text-gray-400">Avg Latency</p>
              <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {{ stats?.average_latency_ms?.toFixed(0) || '0' }}ms
              </p>
            </div>
            <div class="w-12 h-12 bg-yellow-100 dark:bg-yellow-900/30 rounded-lg flex items-center justify-center">
              <BarChart3 class="w-6 h-6 text-yellow-600 dark:text-yellow-400" />
            </div>
          </div>
        </div>
      </Card>
    </div>

    <!-- Charts Section -->
    <div v-if="!loading && stats" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Requests by Provider -->
      <Card>
        <div class="p-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Requests by Provider</h2>
          <div class="space-y-3">
            <div
              v-for="[provider, count] in Object.entries(stats.requests_by_provider || {})"
              :key="provider"
              class="flex items-center justify-between"
            >
              <span class="text-sm text-gray-600 dark:text-gray-400">{{ provider }}</span>
              <div class="flex items-center space-x-2">
                <div class="w-32 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                  <div
                    class="h-full bg-blue-600 dark:bg-blue-500 rounded-full transition-all"
                    :style="{ width: `${(count / stats.total_requests) * 100}%` }"
                  ></div>
                </div>
                <span class="text-sm font-medium text-gray-900 dark:text-white w-12 text-right">
                  {{ count }}
                </span>
              </div>
            </div>
            <p v-if="Object.keys(stats.requests_by_provider || {}).length === 0" class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
              No provider data available
            </p>
          </div>
        </div>
      </Card>

      <!-- Cost by Provider -->
      <Card>
        <div class="p-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Cost by Provider</h2>
          <div class="space-y-3">
            <div
              v-for="[provider, cost] in Object.entries(stats.cost_by_provider || {})"
              :key="provider"
              class="flex items-center justify-between"
            >
              <span class="text-sm text-gray-600 dark:text-gray-400">{{ provider }}</span>
              <div class="flex items-center space-x-2">
                <div class="w-32 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
                  <div
                    class="h-full bg-purple-600 dark:bg-purple-500 rounded-full transition-all"
                    :style="{ width: `${(cost / stats.total_cost) * 100}%` }"
                  ></div>
                </div>
                <span class="text-sm font-medium text-gray-900 dark:text-white w-16 text-right">
                  ${{ cost.toFixed(4) }}
                </span>
              </div>
            </div>
            <p v-if="Object.keys(stats.cost_by_provider || {}).length === 0" class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
              No cost data available
            </p>
          </div>
        </div>
      </Card>
    </div>

    <!-- Recent Requests -->
    <Card v-if="!loading">
      <div class="p-6">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Recent Requests</h2>
        <div class="overflow-x-auto">
          <table class="w-full" role="table" aria-label="Recent requests table">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Provider</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Model</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Tokens</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Cost</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Latency</th>
                <th class="px-4 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">Time</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200 dark:divide-gray-700">
              <tr v-for="request in recentRequests" :key="request.id" class="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                <td class="px-4 py-3 text-sm text-gray-900 dark:text-white">{{ request.provider }}</td>
                <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{{ request.model }}</td>
                <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{{ request.total_tokens.toLocaleString() }}</td>
                <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">${{ request.cost.toFixed(4) }}</td>
                <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{{ request.latency_ms }}ms</td>
                <td class="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">{{ formatDate(request.created_at) }}</td>
              </tr>
            </tbody>
          </table>
          <p v-if="recentRequests.length === 0" class="text-sm text-gray-500 dark:text-gray-400 text-center py-4">
            No requests found
          </p>
        </div>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Card from '@/components/ui/Card.vue'
import { BarChart3 } from 'lucide-vue-next'
import { analyticsApi, type UsageStats, type Request } from '@/services/api/analytics'

const loading = ref(false)
const stats = ref<UsageStats | null>(null)
const recentRequests = ref<Request[]>([])
const timeRange = ref('30')

onMounted(() => {
  loadStats()
  loadRecentRequests()
})

const loadStats = async () => {
  loading.value = true
  try {
    const days = parseInt(timeRange.value)
    const endTime = new Date()
    const startTime = new Date()
    startTime.setDate(startTime.getDate() - days)

    const response = await analyticsApi.getUsageStats(
      startTime.toISOString(),
      endTime.toISOString()
    )
    stats.value = response
  } catch (error: any) {
    console.error('Failed to load analytics:', error)
  } finally {
    loading.value = false
  }
}

const loadRecentRequests = async () => {
  try {
    const response = await analyticsApi.getRequests(10, 0)
    recentRequests.value = response.requests
  } catch (error: any) {
    console.error('Failed to load recent requests:', error)
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>
