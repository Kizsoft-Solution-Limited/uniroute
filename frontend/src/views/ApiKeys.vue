<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 dark:text-white">API Keys</h1>
        <p class="text-gray-600 dark:text-gray-400 mt-1">
          Manage your API keys and access tokens
        </p>
      </div>
      <Button @click="showCreateModal = true" :icon="Plus">
        Create API Key
      </Button>
    </div>

    <!-- API Keys List -->
    <Card v-if="loading">
      <div class="text-center py-8">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        <p class="text-gray-500 dark:text-gray-400 mt-2">Loading API keys...</p>
      </div>
    </Card>

    <div v-else-if="apiKeys.length === 0" class="text-center py-12">
      <Key class="w-16 h-16 text-gray-400 mx-auto mb-4" />
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">No API keys yet</h3>
      <p class="text-gray-600 dark:text-gray-400 mb-4">
        Create your first API key to start using UniRoute
      </p>
      <Button @click="showCreateModal = true" :icon="Plus">
        Create API Key
      </Button>
    </div>

    <div v-else class="grid gap-4">
      <Card
        v-for="key in apiKeys"
        :key="key.id"
        class="hover:shadow-lg transition-all"
      >
        <div class="flex items-start justify-between">
          <div class="flex-1">
            <div class="flex items-center space-x-3 mb-2">
              <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
                {{ key.name }}
              </h3>
              <span
                v-if="key.isActive"
                class="px-2 py-1 text-xs font-medium bg-green-100 dark:bg-green-900/30 text-green-800 dark:text-green-400 rounded-full"
              >
                Active
              </span>
              <span
                v-else
                class="px-2 py-1 text-xs font-medium bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-400 rounded-full"
              >
                Inactive
              </span>
            </div>
            <div class="space-y-1 text-sm text-gray-600 dark:text-gray-400">
              <p>
                <span class="font-medium">Created:</span>
                {{ formatDate(key.createdAt) }}
              </p>
              <p v-if="key.expiresAt">
                <span class="font-medium">Expires:</span>
                {{ formatDate(key.expiresAt) }}
              </p>
              <p>
                <span class="font-medium">Rate Limit:</span>
                {{ key.rateLimitPerMinute }}/min, {{ key.rateLimitPerDay }}/day
              </p>
            </div>
            <div v-if="key.keyPreview" class="mt-3">
              <div class="flex items-center space-x-2 bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 rounded-lg p-2.5">
                <code class="flex-1 text-sm font-semibold text-gray-900 dark:text-white font-mono select-all">
                  {{ key.keyPreview }}
                </code>
                <button
                  @click="copyToClipboard(key.keyPreview)"
                  class="p-1.5 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 rounded transition-colors flex-shrink-0"
                  title="Copy to clipboard"
                >
                  <Copy class="w-4 h-4" />
                </button>
              </div>
            </div>
          </div>
          <div class="flex items-center space-x-2 ml-4">
            <button
              v-if="key.isActive"
              @click="openRevokeDialog(key.id)"
              class="p-2 text-amber-600 dark:text-amber-400 hover:bg-amber-50 dark:hover:bg-amber-900/20 rounded-lg transition-colors"
              title="Revoke (disable)"
            >
              <ShieldOff class="w-5 h-5" />
            </button>
            <button
              @click="openDeleteDialog(key.id)"
              class="p-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
              :title="key.isActive ? 'Delete (remove)' : 'Delete (remove)'"
            >
              <Trash2 class="w-5 h-5" />
            </button>
          </div>
        </div>
      </Card>
    </div>

    <!-- Create API Key Modal -->
    <div
      v-if="showCreateModal"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      @click.self="showCreateModal = false"
    >
      <Card class="w-full max-w-md">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-4">
          Create API Key
        </h2>
        <form @submit.prevent="createApiKey" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Name
            </label>
            <Input
              v-model="newKey.name"
              placeholder="My API Key"
              required
            />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Rate Limit (per minute)
              </label>
              <Input
                v-model.number="newKey.rate_limit_per_minute"
                type="number"
                min="1"
                placeholder="60"
                required
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Rate Limit (per day)
              </label>
              <Input
                v-model.number="newKey.rate_limit_per_day"
                type="number"
                min="1"
                placeholder="10000"
                required
              />
            </div>
          </div>
          <p class="text-xs text-gray-500 dark:text-gray-400 -mt-1">
            Defaults: 60/min, 10,000/day. Use higher values (e.g. 300/min, 100,000/day) for more traffic.
          </p>
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Expiration
            </label>
            <div class="space-y-3">
              <label class="flex items-center gap-2 cursor-pointer">
                <input
                  v-model="newKey.expirationMode"
                  type="radio"
                  value="never"
                  class="rounded-full border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                />
                <span class="text-gray-700 dark:text-gray-300">Never expire</span>
              </label>
              <label class="flex items-center gap-2 cursor-pointer">
                <input
                  v-model="newKey.expirationMode"
                  type="radio"
                  value="custom"
                  class="rounded-full border-gray-300 dark:border-gray-600 text-blue-600 focus:ring-blue-500"
                />
                <span class="text-gray-700 dark:text-gray-300">Expire on date</span>
              </label>
              <Input
                v-if="newKey.expirationMode === 'custom'"
                v-model="newKey.expiresAt"
                type="date"
                class="mt-2 ml-6 max-w-xs"
                :min="minExpiryDate"
              />
            </div>
          </div>
          <div class="flex items-center space-x-4 pt-4">
            <Button type="submit" :loading="creating">
              Create Key
            </Button>
            <Button variant="outline" @click="showCreateModal = false">
              Cancel
            </Button>
          </div>
        </form>
      </Card>
    </div>

    <!-- Show API Key Modal (only shown once after creation) -->
    <div
      v-if="showKeyModal"
      class="fixed inset-0 bg-black/50 backdrop-blur-sm z-50 flex items-center justify-center p-4"
      @click.self="showKeyModal = false"
    >
      <Card class="w-full max-w-md">
        <h2 class="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          API Key Created
        </h2>
        <p class="text-sm text-yellow-600 dark:text-yellow-400 mb-4">
          ⚠️ Save this key now - it will not be shown again!
        </p>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Your API Key
            </label>
            <div class="relative">
              <div class="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 border-2 border-blue-200 dark:border-blue-700 rounded-lg p-4">
                <code 
                  class="block text-base font-bold text-gray-900 dark:text-white font-mono break-all select-all cursor-text"
                  id="api-key-display"
                >
                  {{ newlyCreatedKey }}
                </code>
              </div>
              <div class="flex items-center justify-end mt-3 space-x-2">
                <button
                  @click="selectAllKey"
                  class="px-3 py-1.5 text-sm font-medium text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300 transition-colors"
                  title="Select all"
                >
                  Select All
                </button>
                <button
                  @click="copyToClipboard(newlyCreatedKey)"
                  class="px-4 py-1.5 bg-blue-600 hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 text-white text-sm font-medium rounded-lg transition-colors flex items-center space-x-2"
                  title="Copy to clipboard"
                >
                  <Copy class="w-4 h-4" />
                  <span>Copy</span>
                </button>
              </div>
            </div>
          </div>
          <Button @click="showKeyModal = false" class="w-full">
            I've Saved It
          </Button>
        </div>
      </Card>
    </div>

    <!-- Revoke (disable) Confirmation Dialog -->
    <ConfirmationDialog
      :show="showRevokeDialog"
      title="Revoke (disable) API Key"
      message="The key will stop working immediately but will stay in the list as inactive. You can remove it later with Delete (remove)."
      variant="danger"
      confirm-text="Revoke (disable)"
      cancel-text="Cancel"
      :loading="revoking"
      @confirm="revokeKey"
      @cancel="cancelRevoke"
    />

    <!-- Delete (remove) Confirmation Dialog -->
    <ConfirmationDialog
      :show="showDeleteDialog"
      title="Delete (remove) API Key"
      message="Permanently remove this key from the list. This cannot be undone. The key will stop working if it is still active."
      variant="danger"
      confirm-text="Delete (remove)"
      cancel-text="Cancel"
      :loading="deleting"
      @confirm="deleteKeyPermanently"
      @cancel="cancelDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { Key, Plus, Copy, Trash2, ShieldOff } from 'lucide-vue-next'
import { useToast } from '@/composables/useToast'
import { apiKeysApi, type ApiKey, type CreateApiKeyResponse } from '@/services/api/apikeys'
import ConfirmationDialog from '@/components/ui/ConfirmationDialog.vue'

const { showToast } = useToast()

const loading = ref(false)
const creating = ref(false)
const apiKeys = ref<ApiKey[]>([])
const showCreateModal = ref(false)
const showKeyModal = ref(false)
const newlyCreatedKey = ref<string>('')
const showRevokeDialog = ref(false)
const keyToRevoke = ref<string | null>(null)
const revoking = ref(false)
const showDeleteDialog = ref(false)
const keyToDelete = ref<string | null>(null)
const deleting = ref(false)
const newKey = ref({
  name: '',
  rate_limit_per_minute: 60,
  rate_limit_per_day: 10000,
  expirationMode: 'never' as 'never' | 'custom',
  expiresAt: ''
})

const minExpiryDate = new Date().toISOString().slice(0, 10)

onMounted(() => {
  loadApiKeys()
})

const loadApiKeys = async () => {
  loading.value = true
  try {
    const response = await apiKeysApi.list()
    apiKeys.value = response.keys.map(key => ({
      ...key,
      isActive: key.is_active,
      createdAt: key.created_at,
      expiresAt: key.expires_at || undefined,
      rateLimitPerMinute: key.rate_limit_per_minute,
      rateLimitPerDay: key.rate_limit_per_day,
      keyPreview: key.key_preview || `${key.id.substring(0, 8)}...` // Show partial ID as preview
    }))
  } catch (error: any) {
    showToast(error.response?.data?.error || 'Failed to load API keys', 'error')
  } finally {
    loading.value = false
  }
}

const createApiKey = async () => {
  if (!newKey.value.name.trim()) {
    showToast('Please enter a name for the API key', 'error')
    return
  }

  const payload: Parameters<typeof apiKeysApi.create>[0] = {
    name: newKey.value.name.trim(),
    rate_limit_per_minute: newKey.value.rate_limit_per_minute || 60,
    rate_limit_per_day: newKey.value.rate_limit_per_day || 10000
  }
  if (newKey.value.expirationMode === 'custom' && newKey.value.expiresAt) {
    payload.expires_at = new Date(newKey.value.expiresAt + 'T23:59:59Z').toISOString()
  }

  creating.value = true
  try {
    const response: CreateApiKeyResponse = await apiKeysApi.create(payload)
    
    // Show the key to user (only time it's shown!)
    newlyCreatedKey.value = response.key
    showCreateModal.value = false
    showKeyModal.value = true
    
    // Reset form
    newKey.value = {
      name: '',
      rate_limit_per_minute: 60,
      rate_limit_per_day: 10000,
      expirationMode: 'never',
      expiresAt: ''
    }
    
    // Reload list
    await loadApiKeys()
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to create API key', 'error')
  } finally {
    creating.value = false
  }
}

const openRevokeDialog = (id: string) => {
  keyToRevoke.value = id
  showRevokeDialog.value = true
}

const openDeleteDialog = (id: string) => {
  keyToDelete.value = id
  showDeleteDialog.value = true
}

const revokeKey = async () => {
  if (!keyToRevoke.value) return
  revoking.value = true
  try {
    await apiKeysApi.revoke(keyToRevoke.value)
    showToast('API key revoked (disabled)', 'success')
    showRevokeDialog.value = false
    keyToRevoke.value = null
    await loadApiKeys()
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to revoke API key', 'error')
  } finally {
    revoking.value = false
  }
}

const deleteKeyPermanently = async () => {
  if (!keyToDelete.value) return
  deleting.value = true
  try {
    await apiKeysApi.deletePermanently(keyToDelete.value)
    showToast('API key deleted permanently', 'success')
    showDeleteDialog.value = false
    keyToDelete.value = null
    await loadApiKeys()
  } catch (error: any) {
    showToast(error.response?.data?.error || error.message || 'Failed to delete API key', 'error')
  } finally {
    deleting.value = false
  }
}

const cancelRevoke = () => {
  showRevokeDialog.value = false
  keyToRevoke.value = null
}

const cancelDelete = () => {
  showDeleteDialog.value = false
  keyToDelete.value = null
}

const copyToClipboard = async (text: string) => {
  try {
    await navigator.clipboard.writeText(text)
    showToast('Copied to clipboard', 'success')
  } catch (error) {
    showToast('Failed to copy to clipboard', 'error')
  }
}

const selectAllKey = () => {
  const element = document.getElementById('api-key-display')
  if (element) {
    const range = document.createRange()
    range.selectNodeContents(element)
    const selection = window.getSelection()
    if (selection) {
      selection.removeAllRanges()
      selection.addRange(range)
    }
  }
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric'
  })
}
</script>
