<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">Custom Domains</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          Manage your custom domains and assign them to tunnels
        </p>
      </div>
      <button
        @click="showAddModal = true"
        class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-semibold transition-colors flex items-center gap-2"
      >
        <Plus class="w-5 h-5" />
        Add Domain
      </button>
    </div>

    <!-- Domains List -->
    <Card v-if="loading">
      <div class="text-center py-8">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="text-gray-500 dark:text-gray-400 mt-2">Loading domains...</p>
      </div>
    </Card>

    <div v-else-if="domains.length === 0" class="text-center py-12">
      <Globe class="w-16 h-16 text-gray-400 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">No domains yet</h3>
      <p class="text-gray-600 dark:text-gray-400 mb-4">
        Add your first custom domain to use with your tunnels
      </p>
      <button
        @click="showAddModal = true"
        class="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-semibold transition-colors flex items-center gap-2 mx-auto"
      >
        <Plus class="w-5 h-5" />
        Add Domain
      </button>
    </div>

    <div v-else class="grid gap-4">
      <Card
        v-for="domain in domains"
        :key="domain.id"
        class="hover:shadow-lg transition-all"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center space-x-3 mb-3">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white font-mono">
                {{ domain.domain }}
              </h3>
              <span
                v-if="domain.dns_configured"
                class="px-2 py-1 text-xs font-medium bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded-full"
              >
                DNS Configured
              </span>
            </div>
            <div class="space-y-2 text-sm text-gray-600 dark:text-gray-400">
              <p>
                <span class="font-medium">Created:</span>
                {{ formatDate(domain.created_at) }}
              </p>
              <div v-if="!domain.dns_configured" class="mt-3 p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
                <p class="text-xs font-medium text-yellow-800 dark:text-yellow-300 mb-2">DNS Setup Required:</p>
                <div class="space-y-2">
                  <code class="block text-xs text-yellow-900 dark:text-yellow-200 font-mono bg-yellow-100 dark:bg-yellow-900/50 px-2 py-1 rounded">
                    CNAME {{ domain.domain }} → tunnel.uniroute.co
                  </code>
                  <p class="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
                    <strong>Root domain (no www)?</strong> Many providers don't allow CNAME on apex. Use ALIAS/ANAME record <code class="bg-yellow-200/50 dark:bg-yellow-800/50 px-1 rounded">@</code> → <code class="bg-yellow-200/50 dark:bg-yellow-800/50 px-1 rounded">tunnel.uniroute.co</code>, or an A record with the tunnel server IP.
                  </p>
                  <p class="text-xs text-yellow-700 dark:text-yellow-300">
                    After configuring DNS, click "Verify DNS" to check if it's set up correctly.
                  </p>
                </div>
              </div>
            </div>
          </div>
          <div class="flex items-center space-x-2 ml-4">
            <button
              v-if="!domain.dns_configured"
              @click="verifyDomain(domain.id)"
              :disabled="verifyingDomain[domain.id]"
              class="p-2 text-green-600 dark:text-green-400 hover:bg-green-50 dark:hover:bg-green-900/20 rounded-lg transition-colors disabled:opacity-50"
              title="Verify DNS configuration"
            >
              <svg v-if="!verifyingDomain[domain.id]" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span v-else class="inline-block animate-spin">⏳</span>
            </button>
            <button
              @click="copyToClipboard(domain.domain)"
              class="p-2 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded-lg transition-colors"
              title="Copy domain"
            >
              <Copy class="w-5 h-5" />
            </button>
            <button
              @click="openDeleteDialog(domain.id, domain.domain)"
              class="p-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
              title="Delete domain"
            >
              <Trash2 class="w-5 h-5" />
            </button>
          </div>
        </div>
      </Card>
    </div>

    <!-- Add Domain Modal -->
    <div
      v-if="showAddModal"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      @click.self="closeAddModal"
    >
      <div class="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md w-full shadow-xl">
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Add Custom Domain</h3>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Domain
            </label>
            <input
              v-model="newDomainInput"
              type="text"
              placeholder="example.com"
              class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              @keyup.enter="addDomain"
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Add this domain to your account. You can then assign it to any tunnel.
            </p>
          </div>
          <div class="flex gap-2 justify-end">
            <button
              @click="closeAddModal"
              class="px-4 py-2 text-sm font-medium bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded-lg transition-colors"
            >
              Cancel
            </button>
            <button
              @click="addDomain"
              :disabled="addingDomain || !newDomainInput.trim()"
              class="px-4 py-2 text-sm font-medium bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <span v-if="!addingDomain">Add Domain</span>
              <span v-else class="inline-block animate-spin">⏳</span>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Delete Confirmation Dialog -->
    <ConfirmationDialog
      :show="showDeleteDialog"
      title="Delete Domain"
      :message="`Are you sure you want to delete ${domainToDelete?.domain}? This will remove it from all tunnels using it.`"
      variant="warning"
      confirm-text="Delete"
      cancel-text="Cancel"
      :loading="deletingDomain"
      @confirm="deleteDomain"
      @cancel="cancelDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Card from '@/components/ui/Card.vue'
import { Globe, Plus, Copy, Trash2 } from 'lucide-vue-next'
import { useToast } from '@/composables/useToast'
import { domainsApi, type CustomDomain } from '@/services/api/domains'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'

const { showToast } = useToast()

const loading = ref(false)
const domains = ref<CustomDomain[]>([])
const showAddModal = ref(false)
const newDomainInput = ref('')
const addingDomain = ref(false)
const showDeleteDialog = ref(false)
const domainToDelete = ref<{ id: string; domain: string } | null>(null)
const deletingDomain = ref(false)
const verifyingDomain = ref<Record<string, boolean>>({})

onMounted(() => {
  loadDomains()
})

const loadDomains = async () => {
  loading.value = true
  try {
    const response = await domainsApi.list()
    domains.value = response.domains
  } catch (error: any) {
    showToast(error.response?.data?.error || 'Failed to load domains', 'error')
  } finally {
    loading.value = false
  }
}

const addDomain = async () => {
  const domain = newDomainInput.value.trim()
  if (!domain) {
    showToast('Please enter a domain', 'error')
    return
  }
  
  addingDomain.value = true
  try {
    await domainsApi.create(domain)
    showToast('Domain added successfully', 'success')
    closeAddModal()
    await loadDomains()
  } catch (error: any) {
    showToast(
      error.response?.data?.error || error.message || 'Failed to add domain',
      'error'
    )
  } finally {
    addingDomain.value = false
  }
}

const closeAddModal = () => {
  showAddModal.value = false
  newDomainInput.value = ''
}

const openDeleteDialog = (id: string, domain: string) => {
  domainToDelete.value = { id, domain }
  showDeleteDialog.value = true
}

const deleteDomain = async () => {
  if (!domainToDelete.value) return
  
  deletingDomain.value = true
  try {
    await domainsApi.delete(domainToDelete.value.id)
    showToast('Domain deleted successfully', 'success')
    showDeleteDialog.value = false
    domainToDelete.value = null
    await loadDomains()
  } catch (error: any) {
    showToast(
      error.response?.data?.error || error.message || 'Failed to delete domain',
      'error'
    )
  } finally {
    deletingDomain.value = false
  }
}

const cancelDelete = () => {
  showDeleteDialog.value = false
  domainToDelete.value = null
}

const verifyDomain = async (domainId: string) => {
  verifyingDomain.value[domainId] = true
  try {
    const result = await domainsApi.verify(domainId)
    if (result.dns_configured) {
      showToast('DNS is properly configured!', 'success')
      await loadDomains()
    } else {
      showToast(
        result.dns_error || 'DNS not configured yet. Please add the CNAME record and try again.',
        'warning'
      )
    }
  } catch (error: any) {
    showToast(
      error.response?.data?.error || error.message || 'Failed to verify DNS',
      'error'
    )
  } finally {
    verifyingDomain.value[domainId] = false
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
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>
