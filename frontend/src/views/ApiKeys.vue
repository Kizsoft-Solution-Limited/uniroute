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
            <div v-if="key.keyPreview" class="mt-3 flex items-center space-x-2">
              <code class="px-3 py-1.5 bg-gray-100 dark:bg-gray-800 rounded text-sm font-mono">
                {{ key.keyPreview }}
              </code>
              <button
                @click="copyToClipboard(key.keyPreview)"
                class="p-1.5 text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
                title="Copy to clipboard"
              >
                <Copy class="w-4 h-4" />
              </button>
            </div>
          </div>
          <div class="flex items-center space-x-2 ml-4">
            <button
              @click="revokeKey(key.id)"
              class="p-2 text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 rounded-lg transition-colors"
              title="Revoke key"
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
                v-model.number="newKey.rateLimitPerMinute"
                type="number"
                min="1"
                required
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                Rate Limit (per day)
              </label>
              <Input
                v-model.number="newKey.rateLimitPerDay"
                type="number"
                min="1"
                required
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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import Input from '@/components/ui/Input.vue'
import { Key, Plus, Copy, Trash2 } from 'lucide-vue-next'
import { useToast } from '@/composables/useToast'

interface ApiKey {
  id: string
  name: string
  isActive: boolean
  createdAt: string
  expiresAt?: string
  rateLimitPerMinute: number
  rateLimitPerDay: number
  keyPreview?: string
}

const { showToast } = useToast()

const loading = ref(false)
const creating = ref(false)
const apiKeys = ref<ApiKey[]>([])
const showCreateModal = ref(false)
const newKey = ref({
  name: '',
  rateLimitPerMinute: 60,
  rateLimitPerDay: 10000
})

onMounted(() => {
  loadApiKeys()
})

const loadApiKeys = async () => {
  loading.value = true
  try {
    // TODO: Fetch from API
    // const response = await apiKeysApi.list()
    // apiKeys.value = response.data
    
    // Mock data for now
    apiKeys.value = []
  } catch (error: any) {
    showToast('Failed to load API keys', 'error')
  } finally {
    loading.value = false
  }
}

const createApiKey = async () => {
  creating.value = true
  try {
    // TODO: Create via API
    // const response = await apiKeysApi.create(newKey.value)
    // apiKeys.value.push(response.data)
    
    showToast('API key created successfully', 'success')
    showCreateModal.value = false
    newKey.value = {
      name: '',
      rateLimitPerMinute: 60,
      rateLimitPerDay: 10000
    }
    await loadApiKeys()
  } catch (error: any) {
    showToast(error.message || 'Failed to create API key', 'error')
  } finally {
    creating.value = false
  }
}

const revokeKey = async (id: string) => {
  if (!confirm('Are you sure you want to revoke this API key? This action cannot be undone.')) {
    return
  }
  
  try {
    // TODO: Revoke via API
    // await apiKeysApi.revoke(id)
    showToast('API key revoked successfully', 'success')
    await loadApiKeys()
  } catch (error: any) {
    showToast(error.message || 'Failed to revoke API key', 'error')
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
    day: 'numeric'
  })
}
</script>
