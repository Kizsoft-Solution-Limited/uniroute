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

    <!-- Routing Analytics Section -->
    <div v-if="!loading" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Cost Estimator -->
      <Card>
        <div class="p-6">
          <div class="flex items-center justify-between mb-4">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white flex items-center">
              <Calculator class="w-5 h-5 mr-2" />
              Cost Estimator
            </h2>
            <button
              @click="showCostEstimator = !showCostEstimator"
              class="text-sm text-blue-600 dark:text-blue-400 hover:underline"
            >
              {{ showCostEstimator ? 'Hide' : 'Show' }}
            </button>
          </div>

          <div v-if="showCostEstimator" class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Model
              </label>
              <input
                v-model="costEstimateModel"
                type="text"
                placeholder="e.g., gpt-4, claude-3-opus"
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Messages (one per line)
              </label>
              <textarea
                v-model="costEstimateMessages"
                rows="4"
                placeholder="Enter messages, one per line..."
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
              ></textarea>
            </div>

            <button
              @click="estimateCost"
              :disabled="costEstimateLoading || !costEstimateModel || !costEstimateMessages.trim()"
              class="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {{ costEstimateLoading ? 'Estimating...' : 'Estimate Cost' }}
            </button>

            <div v-if="costEstimateResult" class="mt-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-2">
                Cost Estimates for {{ costEstimateResult.model }}
              </h3>
              <div class="space-y-2">
                <div
                  v-for="[provider, cost] in Object.entries(costEstimateResult.estimates)"
                  :key="provider"
                  class="flex items-center justify-between text-sm"
                >
                  <span class="text-gray-600 dark:text-gray-400 capitalize">{{ provider }}</span>
                  <span class="font-medium text-gray-900 dark:text-white">${{ cost.toFixed(6) }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </Card>

      <!-- Latency Stats -->
      <Card>
        <div class="p-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center">
            <Gauge class="w-5 h-5 mr-2" />
            Provider Latency Stats
          </h2>
          <div v-if="latencyStats && Object.keys(latencyStats.latency_stats).length > 0" class="space-y-3">
            <div
              v-for="[provider, stats] in Object.entries(latencyStats.latency_stats)"
              :key="provider"
              class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg"
            >
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium text-gray-900 dark:text-white capitalize">{{ provider }}</span>
                <span class="text-xs text-gray-500 dark:text-gray-400">{{ stats.samples }} samples</span>
              </div>
              <div class="grid grid-cols-3 gap-2 text-xs">
                <div>
                  <span class="text-gray-500 dark:text-gray-400">Avg:</span>
                  <span class="ml-1 font-medium text-gray-900 dark:text-white">{{ stats.average_ms.toFixed(0) }}ms</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">Min:</span>
                  <span class="ml-1 font-medium text-green-600 dark:text-green-400">{{ stats.min_ms.toFixed(0) }}ms</span>
                </div>
                <div>
                  <span class="text-gray-500 dark:text-gray-400">Max:</span>
                  <span class="ml-1 font-medium text-red-600 dark:text-red-400">{{ stats.max_ms.toFixed(0) }}ms</span>
                </div>
              </div>
            </div>
          </div>
          <div v-else class="text-center py-4">
            <p class="text-sm text-gray-500 dark:text-gray-400 mb-2">
              No latency data available yet.
            </p>
            <p class="text-xs text-gray-400 dark:text-gray-500">
              Latency statistics will appear after you make AI chat requests through UniRoute.
            </p>
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
import { BarChart3, Calculator, Gauge } from 'lucide-vue-next'
import { analyticsApi, type UsageStats, type Request, type CostEstimateResponse, type LatencyStats } from '@/services/api/analytics'
import { useToast } from '@/composables/useToast'

const { showToast } = useToast()
const loading = ref(false)
const stats = ref<UsageStats | null>(null)
const recentRequests = ref<Request[]>([])
const timeRange = ref('30')
const latencyStats = ref<LatencyStats | null>(null)

// Cost estimation
const showCostEstimator = ref(false)
const costEstimateModel = ref('gpt-4')
const costEstimateMessages = ref('')
const costEstimateLoading = ref(false)
const costEstimateResult = ref<CostEstimateResponse | null>(null)

onMounted(() => {
  loadStats()
  loadRecentRequests()
  // Latency stats will be loaded after stats are loaded (if user has requests)
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
    
    // Load latency stats if user has made requests
    if (response.total_requests > 0) {
      loadLatencyStats()
    }
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

const loadLatencyStats = async () => {
  try {
    const response = await analyticsApi.getLatencyStats()
    latencyStats.value = response
  } catch (error: any) {
    console.error('Failed to load latency stats:', error)
  }
}

const estimateCost = async () => {
  if (!costEstimateModel.value || !costEstimateMessages.value.trim()) {
    return
  }

  costEstimateLoading.value = true
  costEstimateResult.value = null

  try {
    // Parse messages (simple format: assume user message if not structured)
    const messages = costEstimateMessages.value.trim().split('\n').filter(m => m.trim()).map(content => ({
      role: 'user' as const,
      content: content.trim()
    }))

    // If empty, add a default message
    if (messages.length === 0) {
      messages.push({ role: 'user', content: costEstimateMessages.value.trim() || 'Hello' })
    }

    const response = await analyticsApi.estimateCost({
      model: costEstimateModel.value,
      messages
    })
    costEstimateResult.value = response
  } catch (error: any) {
    console.error('Failed to estimate cost:', error)
    showToast(error.response?.data?.message || 'Failed to estimate cost', 'error')
  } finally {
    costEstimateLoading.value = false
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
