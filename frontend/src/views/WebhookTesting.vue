<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-white">Webhook Testing</h1>
        <p class="text-slate-400 mt-1">
          Inspect, replay, and test webhook requests
        </p>
      </div>
      <div class="flex items-center space-x-4">
        <select
          v-model="selectedTunnel"
          @change="onTunnelChange"
          class="px-4 py-2 border border-slate-700 rounded-lg bg-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="">Select tunnel...</option>
          <option v-for="tunnel in tunnels" :key="tunnel.id" :value="tunnel.id">
            {{ tunnel.subdomain }} - {{ tunnel.public_url }}
          </option>
        </select>
      </div>
    </div>

    <!-- Filters & Search -->
    <Card v-if="selectedTunnel">
      <div class="space-y-4">
        <!-- Search Bar -->
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-2">
            Search
          </label>
          <Input
            v-model="filters.search"
            @input="debouncedSearch"
            placeholder="Search in path, headers, or body..."
            class="w-full"
          />
        </div>
        
        <!-- Filter Grid -->
        <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">
              Method
            </label>
            <select
              v-model="filters.method"
              @change="loadRequests"
              class="w-full px-3 py-2 border border-slate-700 rounded-lg bg-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All</option>
              <option value="GET">GET</option>
              <option value="POST">POST</option>
              <option value="PUT">PUT</option>
              <option value="PATCH">PATCH</option>
              <option value="DELETE">DELETE</option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">
              Path Filter
            </label>
            <Input
              v-model="filters.path"
              @input="debouncedSearch"
              placeholder="/webhook"
              class="w-full"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">
              Status Code
            </label>
            <select
              v-model="filters.statusCode"
              @change="loadRequests"
              class="w-full px-3 py-2 border border-slate-700 rounded-lg bg-slate-800 text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              <option value="">All</option>
              <option value="200">200 OK</option>
              <option value="400">400 Bad Request</option>
              <option value="404">404 Not Found</option>
              <option value="500">500 Server Error</option>
            </select>
          </div>
          <div class="flex items-end">
            <Button @click="loadRequests" :loading="loading" class="w-full">
              <Search class="w-4 h-4 mr-2" />
              Refresh
            </Button>
          </div>
        </div>
      </div>
    </Card>

    <!-- Requests List -->
    <Card v-if="loading && requests.length === 0">
      <div class="text-center py-8">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <p class="text-slate-400 mt-2">Loading requests...</p>
      </div>
    </Card>

    <div v-else-if="!selectedTunnel" class="text-center py-12">
      <Webhook class="w-16 h-16 text-slate-500 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-white mb-2">No tunnel selected</h3>
      <p class="text-slate-400 mb-4">
        Select a tunnel to view webhook requests
      </p>
    </div>

    <div v-else-if="filteredRequests.length === 0" class="text-center py-12">
      <Webhook class="w-16 h-16 text-slate-500 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-white mb-2">No requests found</h3>
      <p class="text-slate-400">
        <span v-if="filters.search || filters.method || filters.path || filters.statusCode">
          No requests match your filters. Try adjusting your search criteria.
        </span>
        <span v-else>
          Requests will appear here once your tunnel receives traffic
        </span>
      </p>
    </div>

    <div v-else class="space-y-4">
      <div
        v-for="request in filteredRequests"
        :key="request.id"
        class="bg-slate-800/60 rounded-lg border border-slate-700/50 p-4 hover:border-blue-500/50 hover:shadow-lg transition-all cursor-pointer"
        @click="selectRequest(request)"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center space-x-3 mb-2">
              <span
                class="px-2 py-1 text-xs font-semibold rounded"
                :class="{
                  'bg-blue-900/30 text-blue-400': request.method === 'GET',
                  'bg-green-900/30 text-green-400': request.method === 'POST',
                  'bg-yellow-900/30 text-yellow-400': request.method === 'PUT',
                  'bg-purple-900/30 text-purple-400': request.method === 'PATCH',
                  'bg-red-900/30 text-red-400': request.method === 'DELETE',
                }"
              >
                {{ request.method }}
              </span>
              <code class="text-sm font-mono text-white">{{ request.path }}</code>
              <span
                class="px-2 py-1 text-xs font-medium rounded"
                :class="{
                  'bg-green-900/30 text-green-400': request.status_code >= 200 && request.status_code < 300,
                  'bg-yellow-900/30 text-yellow-400': request.status_code >= 300 && request.status_code < 400,
                  'bg-red-900/30 text-red-400': request.status_code >= 400,
                }"
              >
                {{ request.status_code }}
              </span>
            </div>
            <div class="flex items-center space-x-4 text-xs text-slate-400">
              <span>{{ formatDate(request.created_at) }}</span>
              <span>{{ request.latency_ms }}ms</span>
              <span>{{ formatBytes(request.request_size) }} / {{ formatBytes(request.response_size) }}</span>
              <span v-if="request.remote_addr">{{ request.remote_addr }}</span>
            </div>
          </div>
          <div class="flex items-center space-x-2 ml-4">
            <button
              @click.stop="replayRequest(request)"
              class="p-2 text-blue-400 hover:bg-blue-900/20 rounded-lg transition-colors"
              title="Replay request"
            >
              <RotateCcw class="w-5 h-5" />
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Request Detail Modal -->
    <div
      v-if="selectedRequestDetail"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      @click.self="selectedRequestDetail = null"
    >
      <Card class="w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        <div class="flex items-center justify-between mb-6">
          <h2 class="text-2xl font-bold text-white">Request Details</h2>
          <button
            @click="selectedRequestDetail = null"
            class="p-2 text-slate-400 hover:text-white transition-colors"
          >
            <X class="w-5 h-5" />
          </button>
        </div>

        <div class="space-y-6">
          <!-- Request Info -->
          <div>
            <h3 class="text-lg font-semibold text-white mb-4">Request</h3>
            <div class="bg-slate-800/60 rounded-lg p-4 space-y-2 border border-slate-700/50">
              <div class="flex items-center space-x-2">
                <span class="font-medium text-slate-300">Method:</span>
                <code class="text-sm text-white">{{ selectedRequestDetail.method }}</code>
              </div>
              <div class="flex items-center space-x-2">
                <span class="font-medium text-slate-300">Path:</span>
                <code class="text-sm text-white">{{ selectedRequestDetail.path }}</code>
              </div>
              <div v-if="selectedRequestDetail.query_string" class="flex items-center space-x-2">
                <span class="font-medium text-slate-300">Query:</span>
                <code class="text-sm text-white">{{ selectedRequestDetail.query_string }}</code>
              </div>
            </div>
          </div>

          <!-- Request Headers -->
          <div>
            <h3 class="text-lg font-semibold text-white mb-4">Request Headers</h3>
            <div class="bg-slate-800/60 rounded-lg p-4 border border-slate-700/50">
              <pre class="text-sm text-white overflow-x-auto">{{ formatHeaders(selectedRequestDetail.request_headers) }}</pre>
            </div>
          </div>

          <!-- Request Body -->
          <div v-if="selectedRequestDetail.request_body">
            <h3 class="text-lg font-semibold text-white mb-4">Request Body</h3>
            <div class="bg-slate-800/60 rounded-lg p-4 border border-slate-700/50">
              <pre class="text-sm text-white overflow-x-auto">{{ formatBody(selectedRequestDetail.request_body) }}</pre>
            </div>
          </div>

          <!-- Response Info -->
          <div>
            <h3 class="text-lg font-semibold text-white mb-4">Response</h3>
            <div class="bg-slate-800/60 rounded-lg p-4 space-y-2 border border-slate-700/50">
              <div class="flex items-center space-x-2">
                <span class="font-medium text-slate-300">Status:</span>
                <code class="text-sm text-white">{{ selectedRequestDetail.status_code }}</code>
              </div>
              <div class="flex items-center space-x-2">
                <span class="font-medium text-slate-300">Latency:</span>
                <code class="text-sm text-white">{{ selectedRequestDetail.latency_ms }}ms</code>
              </div>
            </div>
          </div>

          <!-- Response Headers -->
          <div v-if="selectedRequestDetail.response_headers">
            <h3 class="text-lg font-semibold text-white mb-4">Response Headers</h3>
            <div class="bg-slate-800/60 rounded-lg p-4 border border-slate-700/50">
              <pre class="text-sm text-white overflow-x-auto">{{ formatHeaders(selectedRequestDetail.response_headers) }}</pre>
            </div>
          </div>

          <!-- Response Body -->
          <div v-if="selectedRequestDetail.response_body">
            <h3 class="text-lg font-semibold text-white mb-4">Response Body</h3>
            <div class="bg-slate-800/60 rounded-lg p-4 border border-slate-700/50">
              <pre class="text-sm text-white overflow-x-auto">{{ formatBody(selectedRequestDetail.response_body) }}</pre>
            </div>
          </div>

          <!-- Actions -->
          <div class="flex items-center space-x-4 pt-4 border-t border-slate-700">
            <Button @click="replayRequest(selectedRequestDetail)" :loading="replaying">
              <RotateCcw class="w-4 h-4 mr-2" />
              Replay Request
            </Button>
            <Button variant="outline" @click="selectedRequestDetail = null">
              Close
            </Button>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { Webhook, Search, RotateCcw, X } from 'lucide-vue-next'
import { useToast } from '@/composables/useToast'
import { getTunnelServerURL } from '@/services/api/client'
import { webhookTestingApi } from '@/services/api/webhookTesting'

interface Tunnel {
  id: string
  subdomain: string
  public_url: string
  local_url: string
}

interface Request {
  id: string
  request_id: string
  method: string
  path: string
  query_string: string
  status_code: number
  latency_ms: number
  request_size: number
  response_size: number
  remote_addr: string
  user_agent: string
  created_at: string
}

interface RequestDetail extends Request {
  request_headers: Record<string, string>
  request_body: string
  response_headers: Record<string, string>
  response_body: string
}

const { showToast } = useToast()

const loading = ref(false)
const replaying = ref(false)
const tunnels = ref<Tunnel[]>([])
const selectedTunnel = ref('')
const requests = ref<Request[]>([])
const selectedRequestDetail = ref<RequestDetail | null>(null)
const allRequestDetails = ref<Map<string, RequestDetail>>(new Map())

const filters = ref({
  method: '',
  path: '',
  statusCode: '',
  search: ''
})

let searchTimeout: ReturnType<typeof setTimeout> | null = null

onMounted(() => {
  loadTunnels()
})

// Auto-load requests when tunnel is selected
const onTunnelChange = () => {
  if (selectedTunnel.value) {
    loadRequests()
  } else {
    requests.value = []
  }
}

// Debounced search function
const debouncedSearch = () => {
  if (searchTimeout) {
    clearTimeout(searchTimeout)
  }
  searchTimeout = setTimeout(() => {
    // Client-side filtering is handled by computed property
    // If we need server-side search, we'd call loadRequests here
  }, 300)
}

// Client-side filtering for search
const filteredRequests = computed(() => {
  if (!filters.value.search) {
    return requests.value
  }
  
  const searchTerm = filters.value.search.toLowerCase()
  return requests.value.filter(request => {
    // Search in path
    if (request.path.toLowerCase().includes(searchTerm)) {
      return true
    }
    
    // Search in request details if loaded
    const detail = allRequestDetails.value.get(request.id)
    if (detail) {
      // Search in headers
      const headersStr = formatHeaders(detail.request_headers).toLowerCase()
      if (headersStr.includes(searchTerm)) {
        return true
      }
      
      // Search in body
      if (detail.request_body && detail.request_body.toLowerCase().includes(searchTerm)) {
        return true
      }
    }
    
    return false
  })
})

const loadTunnels = async () => {
  try {
    const tunnelServerUrl = getTunnelServerURL()
    const data = await webhookTestingApi.listTunnels(tunnelServerUrl)
    tunnels.value = data
  } catch (error: any) {
    showToast(error.message || 'Failed to load tunnels', 'error')
  }
}

const loadRequests = async () => {
  if (!selectedTunnel.value) return

  loading.value = true
  try {
    const tunnelServerUrl = getTunnelServerURL()
    const data = await webhookTestingApi.listRequests(tunnelServerUrl, selectedTunnel.value, {
      method: filters.value.method || undefined,
      path: filters.value.path || undefined,
      limit: 50,
      offset: 0
    })
    requests.value = data.requests
  } catch (error: any) {
    showToast('Failed to load requests', 'error')
  } finally {
    loading.value = false
  }
}

const selectRequest = async (request: Request) => {
  if (!selectedTunnel.value) return

  // Check if we already have the details cached
  if (allRequestDetails.value.has(request.id)) {
    selectedRequestDetail.value = allRequestDetails.value.get(request.id)!
    return
  }

  try {
    const tunnelServerUrl = getTunnelServerURL()
    const data = await webhookTestingApi.getRequest(tunnelServerUrl, selectedTunnel.value, request.request_id)
    allRequestDetails.value.set(request.id, data)
    selectedRequestDetail.value = data
  } catch (error: any) {
    showToast('Failed to load request details', 'error')
  }
}

const replayRequest = async (request: Request | RequestDetail) => {
  if (!selectedTunnel.value) return

  replaying.value = true
  try {
    const tunnelServerUrl = getTunnelServerURL()
    const data = await webhookTestingApi.replayRequest(tunnelServerUrl, selectedTunnel.value, request.request_id)
    showToast(`Request replayed successfully (Status: ${data.status_code})`, 'success')
    await loadRequests()
  } catch (error: any) {
    showToast('Failed to replay request', 'error')
  } finally {
    replaying.value = false
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
}

const formatHeaders = (headers: Record<string, string>) => {
  return Object.entries(headers || {})
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n')
}

const formatBody = (body: string) => {
  try {
    const parsed = JSON.parse(body)
    return JSON.stringify(parsed, null, 2)
  } catch {
    return body
  }
}
</script>

