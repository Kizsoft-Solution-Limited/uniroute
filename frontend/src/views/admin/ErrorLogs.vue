<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-3xl font-bold text-white">Error Logs</h1>
      <p class="text-slate-400 mt-1">Monitor and manage application errors</p>
    </div>

    <!-- Filters -->
    <Card>
      <div class="space-y-4">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">Error Type</label>
          <select
            v-model="filters.error_type"
            class="w-full px-3 py-2 border border-slate-700 rounded-lg bg-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="loadErrorLogs"
          >
            <option value="">All Types</option>
            <option value="exception">Exception</option>
            <option value="message">Message</option>
            <option value="network">Network</option>
            <option value="server">Server</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">Severity</label>
          <select
            v-model="filters.severity"
            class="w-full px-3 py-2 border border-slate-700 rounded-lg bg-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="loadErrorLogs"
          >
            <option value="">All Severities</option>
            <option value="error">Error</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">Status</label>
          <select
            v-model="filters.resolved"
            class="w-full px-3 py-2 border border-slate-700 rounded-lg bg-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="loadErrorLogs"
          >
            <option :value="undefined">All</option>
            <option :value="false">Unresolved</option>
            <option :value="true">Resolved</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">Limit</label>
          <select
            v-model="filters.limit"
            class="w-full px-3 py-2 border border-slate-700 rounded-lg bg-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="loadErrorLogs"
          >
            <option :value="50">50</option>
            <option :value="100">100</option>
            <option :value="200">200</option>
          </select>
        </div>
      </div>

        <div class="flex justify-end pt-4">
          <Button variant="outline" @click="resetFilters">
            Reset Filters
          </Button>
        </div>
      </div>
    </Card>

    <!-- Loading State -->
    <Card v-if="loading">
      <div class="text-center py-12">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="mt-2 text-slate-400">Loading error logs...</p>
      </div>
    </Card>

    <!-- Error State -->
    <Card v-else-if="error">
      <div class="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
        <p class="text-red-400">{{ error }}</p>
      </div>
    </Card>

    <!-- Error Logs Table -->
    <Card v-else>
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-slate-700">
          <thead class="bg-slate-800/50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">Time</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">Type</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">Severity</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">Message</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">User</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">Status</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-slate-300 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="bg-slate-800/30 divide-y divide-slate-700">
            <tr v-for="errorLog in (errorLogs || [])" :key="errorLog.id" class="hover:bg-slate-800/50">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-white">
                {{ formatDate(errorLog.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  class="px-2 py-1 text-xs font-semibold rounded-full"
                  :class="getErrorTypeClass(errorLog.error_type || 'unknown')"
                >
                  {{ errorLog.error_type || 'N/A' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  class="px-2 py-1 text-xs font-semibold rounded-full"
                  :class="getSeverityClass(errorLog.severity || 'error')"
                >
                  {{ errorLog.severity || 'N/A' }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-white">
                <div class="max-w-md truncate" :title="errorLog.message || 'No message'">
                  {{ errorLog.message || 'N/A' }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-slate-400">
                {{ errorLog.user_id ? truncateUUID(errorLog.user_id) : 'N/A' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  class="px-2 py-1 text-xs font-semibold rounded-full"
                  :class="errorLog.resolved ? 'bg-green-500/20 text-green-400' : 'bg-yellow-500/20 text-yellow-400'"
                >
                  {{ errorLog.resolved ? 'Resolved' : 'Unresolved' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <button
                  v-if="!errorLog.resolved"
                  @click="markResolved(errorLog.id)"
                  class="text-blue-400 hover:text-blue-300 mr-4 transition-colors"
                >
                  Mark Resolved
                </button>
                <button
                  @click="viewDetails(errorLog)"
                  class="text-slate-400 hover:text-slate-300 transition-colors"
                >
                  View Details
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Empty State -->
      <div v-if="!errorLogs || errorLogs.length === 0" class="text-center py-12">
        <p class="text-slate-400">No error logs found</p>
      </div>

      <!-- Summary -->
      <div v-if="errorLogs && errorLogs.length > 0" class="px-6 py-4 bg-slate-800/50 border-t border-slate-700">
        <p class="text-sm text-slate-400">
          Showing {{ errorLogs.length }} of {{ totalCount }} error logs
        </p>
      </div>
    </Card>

    <!-- Error Details Modal -->
    <div
      v-if="selectedError"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm flex items-center justify-center z-50 p-4"
      @click.self="selectedError = null"
    >
      <Card class="max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        <div class="p-6">
          <div class="flex justify-between items-center mb-4">
            <h2 class="text-xl font-bold text-white">Error Details</h2>
            <button
              @click="selectedError = null"
              class="text-slate-400 hover:text-white transition-colors"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Message</label>
              <p class="text-sm text-white bg-slate-800/50 p-3 rounded-lg border border-slate-700">{{ selectedError.message }}</p>
            </div>

            <div v-if="selectedError.stack_trace">
              <label class="block text-sm font-medium text-slate-300 mb-2">Stack Trace</label>
              <pre class="text-xs text-white bg-slate-800/50 p-3 rounded-lg border border-slate-700 overflow-x-auto">{{ selectedError.stack_trace }}</pre>
            </div>

            <div v-if="selectedError.url">
              <label class="block text-sm font-medium text-slate-300 mb-2">URL</label>
              <p class="text-sm text-white bg-slate-800/50 p-3 rounded-lg border border-slate-700 break-all">{{ selectedError.url }}</p>
            </div>

            <div v-if="selectedError.user_agent">
              <label class="block text-sm font-medium text-slate-300 mb-2">User Agent</label>
              <p class="text-sm text-white bg-slate-800/50 p-3 rounded-lg border border-slate-700">{{ selectedError.user_agent }}</p>
            </div>

            <div v-if="selectedError.ip_address">
              <label class="block text-sm font-medium text-slate-300 mb-2">IP Address</label>
              <p class="text-sm text-white bg-slate-800/50 p-3 rounded-lg border border-slate-700">{{ selectedError.ip_address }}</p>
            </div>

            <div v-if="selectedError.context && Object.keys(selectedError.context).length > 0">
              <label class="block text-sm font-medium text-slate-300 mb-2">Context</label>
              <pre class="text-xs text-white bg-slate-800/50 p-3 rounded-lg border border-slate-700 overflow-x-auto">{{ JSON.stringify(selectedError.context, null, 2) }}</pre>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-slate-300 mb-2">Error Type</label>
                <p class="text-sm text-white">{{ selectedError.error_type }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-slate-300 mb-2">Severity</label>
                <p class="text-sm text-white">{{ selectedError.severity }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-slate-300 mb-2">Status</label>
                <p class="text-sm text-white">{{ selectedError.resolved ? 'Resolved' : 'Unresolved' }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-slate-300 mb-2">Created At</label>
                <p class="text-sm text-white">{{ formatDate(selectedError.created_at) }}</p>
              </div>
            </div>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { errorsApi, type ErrorLog, type ErrorLogFilters } from '@/services/api/errors'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import { ErrorHandler } from '@/utils/errorHandler'

const errorLogs = ref<ErrorLog[]>([])
const loading = ref(false)
const error = ref<string | null>(null)
const totalCount = ref(0)
const selectedError = ref<ErrorLog | null>(null)

const filters = ref<ErrorLogFilters>({
  error_type: '',
  severity: '',
  resolved: undefined,
  limit: 50,
})

const loadErrorLogs = async () => {
  loading.value = true
  error.value = null

  try {
    const response = await errorsApi.getErrorLogs(filters.value)
    errorLogs.value = response.errors || []
    totalCount.value = response.count || 0
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message
    errorLogs.value = [] // Ensure it's always an array
    totalCount.value = 0
    ErrorHandler.logError(err, 'ErrorLogs')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.value = {
    error_type: '',
    severity: '',
    resolved: undefined,
    limit: 50,
  }
  loadErrorLogs()
}

const markResolved = async (errorId: string) => {
  try {
    await errorsApi.markResolved(errorId)
    // Reload error logs
    await loadErrorLogs()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message
    ErrorHandler.logError(err, 'ErrorLogs')
  }
}

const viewDetails = (errorLog: ErrorLog) => {
  selectedError.value = errorLog
}

const formatDate = (dateString: string | null | undefined) => {
  if (!dateString) {
    return 'N/A'
  }
  try {
    const date = new Date(dateString)
    if (isNaN(date.getTime())) {
      return 'Invalid Date'
    }
    return date.toLocaleString()
  } catch (e) {
    return 'Invalid Date'
  }
}

const truncateUUID = (uuid: string) => {
  return uuid.substring(0, 8) + '...'
}

const getErrorTypeClass = (type: string) => {
  const classes: Record<string, string> = {
    exception: 'bg-red-500/20 text-red-400',
    message: 'bg-blue-500/20 text-blue-400',
    network: 'bg-yellow-500/20 text-yellow-400',
    server: 'bg-purple-500/20 text-purple-400',
  }
  return classes[type] || 'bg-slate-500/20 text-slate-400'
}

const getSeverityClass = (severity: string) => {
  const classes: Record<string, string> = {
    error: 'bg-red-500/20 text-red-400',
    warning: 'bg-yellow-500/20 text-yellow-400',
    info: 'bg-blue-500/20 text-blue-400',
  }
  return classes[severity] || 'bg-slate-500/20 text-slate-400'
}

onMounted(() => {
  loadErrorLogs()
})
</script>


