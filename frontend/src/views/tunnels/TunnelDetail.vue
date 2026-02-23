<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <button
          @click="$router.push(backUrl || '/dashboard/tunnels')"
          class="text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 mb-2 flex items-center space-x-2"
          aria-label="Back to tunnels list"
        >
          <ArrowLeft class="w-4 h-4" />
          <span>{{ isAdminView ? 'Back to Tunnel Management' : 'Back to Tunnels' }}</span>
        </button>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white" id="page-title">Tunnel Details</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          View detailed information and statistics for this tunnel
        </p>
      </div>
    </div>

    <!-- Loading State -->
    <Card v-if="loading">
      <div class="text-center py-8" role="status" aria-live="polite">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600" aria-hidden="true"></div>
        <p class="text-gray-500 dark:text-gray-400 mt-2">Loading tunnel details...</p>
      </div>
    </Card>

    <!-- Tunnel Info -->
    <div v-else-if="tunnel" class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Basic Info -->
      <Card>
        <div class="p-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Tunnel Information</h2>
          <div class="space-y-4">
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Subdomain</label>
              <p class="text-lg font-semibold text-gray-900 dark:text-white mt-1">{{ tunnel.subdomain }}</p>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Public URL</label>
              <div class="flex items-center space-x-2 mt-1 flex-wrap gap-2">
                <a
                  :href="tunnel.publicUrl"
                  target="_blank"
                  class="text-blue-600 dark:text-blue-400 hover:underline break-all min-w-0 flex-1"
                  :aria-label="`Open ${tunnel.publicUrl} in new tab`"
                >
                  {{ tunnel.publicUrl }}
                </a>
                <button
                  @click="copyToClipboard(tunnel.publicUrl)"
                  class="p-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors flex-shrink-0"
                  aria-label="Copy public URL"
                  title="Copy public URL"
                >
                  <Copy class="w-4 h-4" />
                </button>
              </div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Local URL</label>
              <div class="flex items-center space-x-2 mt-1 flex-wrap gap-2">
                <code class="px-3 py-2 bg-gray-100 dark:bg-gray-700 text-gray-900 dark:text-white rounded text-sm font-mono break-all min-w-0 flex-1 border border-gray-200 dark:border-gray-600">
                  {{ tunnel.localUrl || 'Not available' }}
                </code>
                <button
                  v-if="tunnel.localUrl"
                  @click="copyToClipboard(tunnel.localUrl)"
                  class="p-2 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700 rounded transition-colors flex-shrink-0"
                  aria-label="Copy local URL"
                  title="Copy local URL"
                >
                  <Copy class="w-4 h-4" />
                </button>
              </div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Status</label>
              <div class="mt-1">
                <span
                  :class="[
                    'px-3 py-1 text-sm font-medium rounded-full inline-flex items-center space-x-2',
                    tunnel.status === 'active'
                      ? 'bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400'
                      : 'bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-400'
                  ]"
                >
                  <div
                    v-if="tunnel.status === 'active'"
                    class="w-2 h-2 bg-green-500 rounded-full animate-pulse"
                    aria-hidden="true"
                  ></div>
                  <span>{{ tunnel.status === 'active' ? 'Active' : 'Inactive' }}</span>
                </span>
              </div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Created</label>
              <p class="text-gray-900 dark:text-white mt-1">{{ formatDate(tunnel.createdAt) }}</p>
            </div>
            <div v-if="tunnel.lastActive">
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Last Active</label>
              <p class="text-gray-900 dark:text-white mt-1">{{ formatDate(tunnel.lastActive) }}</p>
            </div>
            <div v-if="isAdminView && (tunnel.userDisplay || tunnel.userId)">
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">User</label>
              <p class="text-gray-900 dark:text-white mt-1">{{ tunnel.userDisplay || tunnel.userId }}</p>
            </div>
            <div v-if="tunnel.protocol">
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Protocol</label>
              <p class="text-gray-900 dark:text-white mt-1">{{ tunnel.protocol }}</p>
            </div>
            <div v-if="tunnel.customDomain">
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Custom domain</label>
              <div class="flex items-center gap-2 mt-1 flex-wrap">
                <p class="text-gray-900 dark:text-white">{{ tunnel.customDomain }}</p>
                <button
                  v-if="!isAdminView"
                  @click="openUnassignDomainDialog"
                  class="px-3 py-1.5 text-sm bg-amber-100 dark:bg-amber-900/40 text-amber-800 dark:text-amber-300 hover:bg-amber-200 dark:hover:bg-amber-900/60 rounded-lg transition-colors"
                  :disabled="unassigningDomain"
                  aria-label="Unassign custom domain from this tunnel"
                >
                  Unassign domain
                </button>
              </div>
            </div>
          </div>
        </div>
      </Card>

      <!-- Statistics -->
      <Card>
        <div class="p-6">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Statistics</h2>
          <div class="space-y-4">
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Total Requests</label>
              <p class="text-2xl font-bold text-gray-900 dark:text-white mt-1">
                {{ tunnel.requestCount.toLocaleString() }}
              </p>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">Uptime</label>
              <p class="text-lg font-semibold text-gray-900 dark:text-white mt-1">
                {{ uptimeDisplay }}
              </p>
            </div>
          </div>
        </div>
      </Card>
    </div>

    <!-- Actions (only for own tunnel, not in admin view; show Disconnect only when tunnel is active) -->
    <Card v-if="tunnel && !isAdminView">
      <div class="p-6">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Actions</h2>
        <div class="flex space-x-3 items-center">
          <button
            v-if="tunnel.status === 'active'"
            @click="openDisconnectDialog"
            class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors"
            :disabled="disconnecting"
            aria-label="Disconnect tunnel"
          >
            Disconnect Tunnel
          </button>
          <p v-else class="text-sm text-gray-500 dark:text-gray-400">
            Tunnel is disconnected. Start the tunnel again from the CLI to reconnect.
          </p>
        </div>
      </div>
    </Card>

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

    <!-- Unassign domain Confirmation Dialog -->
    <ConfirmationDialog
      :show="showUnassignDomainDialog"
      title="Unassign custom domain"
      :message="`Remove ${tunnel?.customDomain || ''} from this tunnel? The domain will stay in your account and can be assigned to another tunnel later.`"
      variant="warning"
      confirm-text="Unassign"
      cancel-text="Cancel"
      :loading="unassigningDomain"
      @confirm="unassignDomain"
      @cancel="showUnassignDomainDialog = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Card from '@/components/ui/Card.vue'
import { ArrowLeft, Copy } from 'lucide-vue-next'
import { useToast } from '@/composables/useToast'
import { tunnelsApi } from '@/services/api/tunnels'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'

const route = useRoute()
const router = useRouter()
const { showToast } = useToast()

const isAdminView = computed(() => route.name === 'admin-tunnel-detail')
const backUrl = computed(() => (isAdminView.value ? '/dashboard/admin/tunnels' : ''))

const loading = ref(false)
const disconnecting = ref(false)
const showDisconnectDialog = ref(false)
const unassigningDomain = ref(false)
const showUnassignDomainDialog = ref(false)
const tunnel = ref<{
  id: string
  subdomain: string
  publicUrl: string
  localUrl: string
  status: 'active' | 'inactive'
  requestCount: number
  createdAt: string
  lastActive?: string
  activeSince?: string
  protocol?: string
  customDomain?: string
  userId?: string
  userDisplay?: string
} | null>(null)

onMounted(() => {
  loadTunnel()
})

const loadTunnel = async () => {
  loading.value = true
  try {
    const tunnelId = route.params.id as string
    if (!tunnelId) {
      showToast('Invalid tunnel ID', 'error')
      router.push(isAdminView.value ? '/dashboard/admin/tunnels' : '/dashboard/tunnels')
      return
    }
    const response = isAdminView.value
      ? await tunnelsApi.getAdmin(tunnelId)
      : await tunnelsApi.get(tunnelId)
    const t = response.tunnel
    tunnel.value = {
      id: t.id,
      subdomain: t.subdomain,
      publicUrl: t.public_url,
      localUrl: t.local_url,
      status: t.status as 'active' | 'inactive',
      requestCount: t.request_count || 0,
      createdAt: t.created_at,
      lastActive: t.last_active || undefined,
      activeSince: t.active_since || undefined,
      protocol: t.protocol || undefined,
      customDomain: t.custom_domain || undefined,
      userId: t.user_id || undefined,
      userDisplay: t.user_display || undefined
    }
  } catch (error: any) {
    console.error('Failed to load tunnel:', error)
    showToast(error.response?.data?.error || error.message || 'Failed to load tunnel details', 'error')
    router.push(isAdminView.value ? '/dashboard/admin/tunnels' : '/dashboard/tunnels')
  } finally {
    loading.value = false
  }
}

const openDisconnectDialog = () => {
  if (!tunnel.value) return
  showDisconnectDialog.value = true
}

const disconnectTunnel = async () => {
  if (!tunnel.value) return

  disconnecting.value = true
  try {
    await tunnelsApi.disconnect(tunnel.value.id)
    showToast('Tunnel disconnected successfully', 'success')
    showDisconnectDialog.value = false
    router.push('/dashboard/tunnels')
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to disconnect tunnel', 'error')
  } finally {
    disconnecting.value = false
  }
}

const cancelDisconnect = () => {
  showDisconnectDialog.value = false
}

const openUnassignDomainDialog = () => {
  if (!tunnel.value?.customDomain) return
  showUnassignDomainDialog.value = true
}

const unassignDomain = async () => {
  if (!tunnel.value) return

  unassigningDomain.value = true
  try {
    await tunnelsApi.setCustomDomain(tunnel.value.id, '')
    showToast('Custom domain unassigned from this tunnel', 'success')
    showUnassignDomainDialog.value = false
    await loadTunnel()
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to unassign domain', 'error')
  } finally {
    unassigningDomain.value = false
  }
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
  return new Date(date).toLocaleString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatDuration = (since: string) => {
  const now = new Date()
  const start = new Date(since)
  const diff = now.getTime() - start.getTime()
  const days = Math.floor(diff / (1000 * 60 * 60 * 24))
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
  if (days > 0) return `${days}d ${hours}h ${minutes}m`
  if (hours > 0) return `${hours}h ${minutes}m`
  return `${minutes}m`
}

const uptimeDisplay = computed(() => {
  const t = tunnel.value
  if (!t) return '—'
  if (t.status === 'active' && t.activeSince) return formatDuration(t.activeSince)
  if (t.status === 'inactive') return '—'
  return formatDuration(t.createdAt)
})
</script>

