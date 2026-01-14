<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Tunnels</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          View and manage your active tunnels
        </p>
      </div>
    </div>

    <!-- Tunnels List -->
    <Card v-if="loading">
      <div class="text-center py-8">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="text-gray-500 dark:text-gray-400 mt-2">Loading tunnels...</p>
      </div>
    </Card>

    <div v-else-if="tunnels.length === 0" class="text-center py-12">
      <Network class="w-16 h-16 text-gray-400 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">No tunnels yet</h3>
      <p class="text-gray-600 dark:text-gray-400 mb-4">
        Create a tunnel using the CLI: <code class="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded text-sm">uniroute tunnel</code>
      </p>
    </div>

    <div v-else class="grid gap-4">
      <Card
        v-for="tunnel in tunnels"
        :key="tunnel.id"
        class="hover:shadow-lg transition-all"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center space-x-3 mb-2">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ tunnel.subdomain }}
              </h3>
              <span
                v-if="tunnel.status === 'active'"
                class="px-2 py-1 text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400 rounded-full flex items-center space-x-1"
              >
                <div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <span>Active</span>
              </span>
              <span
                v-else
                class="px-2 py-1 text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-400 rounded-full"
              >
                Inactive
              </span>
            </div>
            <div class="space-y-2 text-sm">
              <div class="flex items-center space-x-2 text-gray-600 dark:text-gray-400">
                <Globe class="w-4 h-4" />
                <span class="font-medium">Public URL:</span>
                <a
                  :href="tunnel.publicUrl"
                  target="_blank"
                  class="text-blue-600 dark:text-blue-400 hover:underline"
                >
                  {{ tunnel.publicUrl }}
                </a>
                <button
                  @click="copyToClipboard(tunnel.publicUrl)"
                  class="p-1 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
                  title="Copy URL"
                >
                  <Copy class="w-4 h-4" />
                </button>
              </div>
              <div class="flex items-center space-x-2 text-gray-600 dark:text-gray-400">
                <Server class="w-4 h-4" />
                <span class="font-medium">Local URL:</span>
                <code class="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded text-xs">
                  {{ tunnel.localUrl }}
                </code>
              </div>
              <div class="flex items-center space-x-4 text-xs text-gray-500 dark:text-gray-400 mt-3">
                <span>
                  <span class="font-medium">Requests:</span>
                  {{ tunnel.requestCount.toLocaleString() }}
                </span>
                <span>
                  <span class="font-medium">Created:</span>
                  {{ formatDate(tunnel.createdAt) }}
                </span>
                <span v-if="tunnel.lastActive">
                  <span class="font-medium">Last Active:</span>
                  {{ formatDate(tunnel.lastActive) }}
                </span>
              </div>
            </div>
          </div>
          <div class="flex items-center space-x-2 ml-4">
            <button
              @click="viewTunnelStats(tunnel.id)"
              class="p-2 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
              title="View stats"
            >
              <BarChart3 class="w-5 h-5" />
            </button>
            <button
              @click="openDisconnectDialog(tunnel.id)"
              class="p-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
              title="Disconnect"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
        </div>
      </Card>
    </div>

    <!-- Disconnect Confirmation Dialog -->
    <ConfirmationDialog
      :show="showDisconnectDialog"
      title="Disconnect Tunnel"
      message="Are you sure you want to disconnect this tunnel? The tunnel will stop forwarding requests immediately."
      variant="warning"
      confirm-text="Disconnect"
      cancel-text="Cancel"
      :loading="disconnecting"
      @confirm="disconnectTunnel"
      @cancel="cancelDisconnect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Card from '@/components/ui/Card.vue'
import { Network, Globe, Server, Copy, BarChart3, X } from 'lucide-vue-next'
import { useToast } from '@/composables/useToast'
import { tunnelsApi } from '@/services/api/tunnels'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'

interface Tunnel {
  id: string
  subdomain: string
  publicUrl: string
  localUrl: string
  status: 'active' | 'inactive'
  requestCount: number
  createdAt: string
  lastActive?: string
}

const { showToast } = useToast()
const router = useRouter()

const loading = ref(false)
const tunnels = ref<Tunnel[]>([])
const showDisconnectDialog = ref(false)
const tunnelToDisconnect = ref<string | null>(null)
const disconnecting = ref(false)

onMounted(() => {
  loadTunnels()
})

const loadTunnels = async () => {
  loading.value = true
  try {
    const response = await tunnelsApi.list()
    tunnels.value = response.tunnels.map(t => ({
      id: t.id,
      subdomain: t.subdomain,
      publicUrl: t.public_url,
      localUrl: t.local_url,
      status: t.status as 'active' | 'inactive',
      requestCount: t.request_count,
      createdAt: t.created_at,
      lastActive: t.last_active || undefined
    }))
  } catch (error: any) {
    showToast(error.response?.data?.error || 'Failed to load tunnels', 'error')
  } finally {
    loading.value = false
  }
}

const viewTunnelStats = (id: string) => {
  // Navigate to tunnel detail page
  router.push({ name: 'tunnel-detail', params: { id } })
}

const openDisconnectDialog = (id: string) => {
  tunnelToDisconnect.value = id
  showDisconnectDialog.value = true
}

const disconnectTunnel = async () => {
  if (!tunnelToDisconnect.value) return
  
  disconnecting.value = true
  try {
    await tunnelsApi.disconnect(tunnelToDisconnect.value)
    showToast('Tunnel disconnected successfully', 'success')
    showDisconnectDialog.value = false
    tunnelToDisconnect.value = null
    await loadTunnels()
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to disconnect tunnel', 'error')
  } finally {
    disconnecting.value = false
  }
}

const cancelDisconnect = () => {
  showDisconnectDialog.value = false
  tunnelToDisconnect.value = null
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    showToast('Copied to clipboard', 'success')
  } catch (error) {
    showToast('Failed to copy to clipboard', 'error')
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>
