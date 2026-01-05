<template>
  <div class="error-logs-page">
    <div class="header">
      <h1 class="text-2xl font-bold text-gray-900">Error Logs</h1>
      <p class="text-sm text-gray-600 mt-1">Monitor and manage application errors</p>
    </div>

    <!-- Filters -->
    <div class="filters bg-white rounded-lg shadow p-4 mt-6">
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Error Type</label>
          <select
            v-model="filters.error_type"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
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
          <label class="block text-sm font-medium text-gray-700 mb-1">Severity</label>
          <select
            v-model="filters.severity"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="loadErrorLogs"
          >
            <option value="">All Severities</option>
            <option value="error">Error</option>
            <option value="warning">Warning</option>
            <option value="info">Info</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Status</label>
          <select
            v-model="filters.resolved"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="loadErrorLogs"
          >
            <option :value="undefined">All</option>
            <option :value="false">Unresolved</option>
            <option :value="true">Resolved</option>
          </select>
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Limit</label>
          <select
            v-model="filters.limit"
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            @change="loadErrorLogs"
          >
            <option :value="50">50</option>
            <option :value="100">100</option>
            <option :value="200">200</option>
          </select>
        </div>
      </div>

      <div class="mt-4 flex justify-end">
        <button
          @click="resetFilters"
          class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
        >
          Reset Filters
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="mt-6 text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <p class="mt-2 text-gray-600">Loading error logs...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="mt-6 bg-red-50 border border-red-200 rounded-lg p-4">
      <p class="text-red-800">{{ error }}</p>
    </div>

    <!-- Error Logs Table -->
    <div v-else class="mt-6 bg-white rounded-lg shadow overflow-hidden">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Time</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Type</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Severity</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Message</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">User</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="errorLog in errorLogs" :key="errorLog.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ formatDate(errorLog.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  class="px-2 py-1 text-xs font-semibold rounded-full"
                  :class="getErrorTypeClass(errorLog.error_type)"
                >
                  {{ errorLog.error_type }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  class="px-2 py-1 text-xs font-semibold rounded-full"
                  :class="getSeverityClass(errorLog.severity)"
                >
                  {{ errorLog.severity }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-gray-900">
                <div class="max-w-md truncate" :title="errorLog.message">
                  {{ errorLog.message }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ errorLog.user_id ? truncateUUID(errorLog.user_id) : 'N/A' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  class="px-2 py-1 text-xs font-semibold rounded-full"
                  :class="errorLog.resolved ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'"
                >
                  {{ errorLog.resolved ? 'Resolved' : 'Unresolved' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
                <button
                  v-if="!errorLog.resolved"
                  @click="markResolved(errorLog.id)"
                  class="text-blue-600 hover:text-blue-900 mr-4"
                >
                  Mark Resolved
                </button>
                <button
                  @click="viewDetails(errorLog)"
                  class="text-gray-600 hover:text-gray-900"
                >
                  View Details
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Empty State -->
      <div v-if="errorLogs.length === 0" class="text-center py-12">
        <p class="text-gray-500">No error logs found</p>
      </div>

      <!-- Summary -->
      <div v-if="errorLogs.length > 0" class="px-6 py-4 bg-gray-50 border-t border-gray-200">
        <p class="text-sm text-gray-600">
          Showing {{ errorLogs.length }} of {{ totalCount }} error logs
        </p>
      </div>
    </div>

    <!-- Error Details Modal -->
    <div
      v-if="selectedError"
      class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50"
      @click.self="selectedError = null"
    >
      <div class="bg-white rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-y-auto">
        <div class="p-6">
          <div class="flex justify-between items-center mb-4">
            <h2 class="text-xl font-bold text-gray-900">Error Details</h2>
            <button
              @click="selectedError = null"
              class="text-gray-400 hover:text-gray-600"
            >
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700">Message</label>
              <p class="mt-1 text-sm text-gray-900 bg-gray-50 p-3 rounded">{{ selectedError.message }}</p>
            </div>

            <div v-if="selectedError.stack_trace">
              <label class="block text-sm font-medium text-gray-700">Stack Trace</label>
              <pre class="mt-1 text-xs text-gray-900 bg-gray-50 p-3 rounded overflow-x-auto">{{ selectedError.stack_trace }}</pre>
            </div>

            <div v-if="selectedError.url">
              <label class="block text-sm font-medium text-gray-700">URL</label>
              <p class="mt-1 text-sm text-gray-900 bg-gray-50 p-3 rounded break-all">{{ selectedError.url }}</p>
            </div>

            <div v-if="selectedError.user_agent">
              <label class="block text-sm font-medium text-gray-700">User Agent</label>
              <p class="mt-1 text-sm text-gray-900 bg-gray-50 p-3 rounded">{{ selectedError.user_agent }}</p>
            </div>

            <div v-if="selectedError.ip_address">
              <label class="block text-sm font-medium text-gray-700">IP Address</label>
              <p class="mt-1 text-sm text-gray-900 bg-gray-50 p-3 rounded">{{ selectedError.ip_address }}</p>
            </div>

            <div v-if="selectedError.context && Object.keys(selectedError.context).length > 0">
              <label class="block text-sm font-medium text-gray-700">Context</label>
              <pre class="mt-1 text-xs text-gray-900 bg-gray-50 p-3 rounded overflow-x-auto">{{ JSON.stringify(selectedError.context, null, 2) }}</pre>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700">Error Type</label>
                <p class="mt-1 text-sm text-gray-900">{{ selectedError.error_type }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">Severity</label>
                <p class="mt-1 text-sm text-gray-900">{{ selectedError.severity }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">Status</label>
                <p class="mt-1 text-sm text-gray-900">{{ selectedError.resolved ? 'Resolved' : 'Unresolved' }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700">Created At</label>
                <p class="mt-1 text-sm text-gray-900">{{ formatDate(selectedError.created_at) }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { errorsApi, type ErrorLog, type ErrorLogFilters } from '@/services/api/errors'
import ErrorHandler from '@/utils/errorHandler'

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
    errorLogs.value = response.errors
    totalCount.value = response.count
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message
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

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleString()
}

const truncateUUID = (uuid: string) => {
  return uuid.substring(0, 8) + '...'
}

const getErrorTypeClass = (type: string) => {
  const classes: Record<string, string> = {
    exception: 'bg-red-100 text-red-800',
    message: 'bg-blue-100 text-blue-800',
    network: 'bg-yellow-100 text-yellow-800',
    server: 'bg-purple-100 text-purple-800',
  }
  return classes[type] || 'bg-gray-100 text-gray-800'
}

const getSeverityClass = (severity: string) => {
  const classes: Record<string, string> = {
    error: 'bg-red-100 text-red-800',
    warning: 'bg-yellow-100 text-yellow-800',
    info: 'bg-blue-100 text-blue-800',
  }
  return classes[severity] || 'bg-gray-100 text-gray-800'
}

onMounted(() => {
  loadErrorLogs()
})
</script>

<style scoped>
.error-logs-page {
  padding: 1.5rem;
}
</style>


