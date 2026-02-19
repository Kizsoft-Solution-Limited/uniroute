<template>
  <div class="provider-keys-container">
    <!-- Header -->
    <div class="header">
      <h1 class="text-3xl font-bold text-gray-900 dark:text-white">
        Provider API Keys
      </h1>
      <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
        Manage your AI provider API keys. Keys are encrypted and stored securely.
      </p>
    </div>

    <!-- Security Notice -->
    <div class="security-notice">
      <div class="flex items-start space-x-3 p-4 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg">
        <svg class="w-5 h-5 text-blue-600 dark:text-blue-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <div class="flex-1">
          <h3 class="text-sm font-medium text-blue-900 dark:text-blue-200">
            Security Information
          </h3>
          <p class="mt-1 text-sm text-blue-700 dark:text-blue-300">
            Your API keys are encrypted using AES-256-GCM and never stored in plaintext. 
            Keys are only used to make requests on your behalf to the respective providers.
          </p>
        </div>
      </div>
    </div>

    <!-- Provider Keys List -->
    <div class="provider-list mt-6">
      <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <ProviderKeyCard
          v-for="provider in providers"
          :key="provider.name"
          :provider="provider"
          :existing-key="getExistingKey(provider.name)"
          @add="handleAddKey"
          @update="handleUpdateKey"
          @delete="openDeleteDialog"
          @test="handleTestKey"
        />
      </div>
    </div>

    <!-- Loading Overlay -->
    <div v-if="loading" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white dark:bg-gray-800 rounded-lg p-6">
        <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
        <p class="mt-4 text-sm text-gray-600 dark:text-gray-400">Processing...</p>
      </div>
    </div>

    <!-- Toast Notifications -->
    <Toast
      v-if="toast.show"
      :message="toast.message"
      :type="toast.type"
      @close="toast.show = false"
    />

    <!-- Delete Confirmation Dialog -->
    <ConfirmationDialog
      :show="showDeleteDialog"
      title="Delete Provider Key"
      :message="`Are you sure you want to delete your ${providerToDelete} API key? This action cannot be undone.`"
      variant="danger"
      confirm-text="Delete Key"
      cancel-text="Cancel"
      :loading="deleting"
      @confirm="handleDeleteKey"
      @cancel="cancelDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useProviderKeys } from '@/composables/useProviderKeys'
import ProviderKeyCard from '@/components/provider-keys/ProviderKeyCard.vue'
import Toast from '@/components/ui/Toast.vue'

interface Provider {
  name: string
  displayName: string
  icon: string
  description: string
  placeholder: string
}

const providers: Provider[] = [
  {
    name: 'openai',
    displayName: 'OpenAI',
    icon: '🤖',
    description: 'GPT-4, GPT-3.5, and other OpenAI models',
    placeholder: 'sk-...'
  },
  {
    name: 'anthropic',
    displayName: 'Anthropic',
    icon: '🧠',
    description: 'Claude models from Anthropic',
    placeholder: 'sk-ant-...'
  },
  {
    name: 'google',
    displayName: 'Google',
    icon: '🔍',
    description: 'Gemini and other Google AI models',
    placeholder: 'AIza...'
  }
]

const { providerKeys, loading, fetchKeys, addKey, updateKey, deleteKey, testKey } = useProviderKeys()

const toast = ref({
  show: false,
  message: '',
  type: 'success' as 'success' | 'error' | 'warning' | 'info'
})

const existingKeys = ref<Map<string, any>>(new Map())

onMounted(async () => {
  await loadKeys()
})

const loadKeys = async () => {
  try {
    await fetchKeys()
    // Map keys by provider name
    existingKeys.value = new Map()
    providerKeys.value.forEach(key => {
      existingKeys.value.set(key.provider, key)
    })
  } catch (error: any) {
    showToast(error.message || 'Failed to load provider keys', 'error')
  }
}

const getExistingKey = (providerName: string) => {
  return existingKeys.value.get(providerName) || null
}

const handleAddKey = async (provider: string, apiKey: string) => {
  try {
    await addKey(provider, apiKey)
    await loadKeys()
    showToast(`${provider} key added successfully`, 'success')
  } catch (error: any) {
    showToast(error.message || 'Failed to add provider key', 'error')
  }
}

const handleUpdateKey = async (provider: string, apiKey: string) => {
  try {
    await updateKey(provider, apiKey)
    await loadKeys()
    showToast(`${provider} key updated successfully`, 'success')
  } catch (error: any) {
    showToast(error.message || 'Failed to update provider key', 'error')
  }
}

const handleDeleteKey = async (provider: string) => {
  if (!confirm(`Are you sure you want to delete your ${provider} API key?`)) {
    return
  }

  try {
    await deleteKey(provider)
    await loadKeys()
    showToast(`${provider} key deleted successfully`, 'success')
  } catch (error: any) {
    showToast(error.message || 'Failed to delete provider key', 'error')
  }
}

const handleTestKey = async (provider: string) => {
  try {
    const result = await testKey(provider)
    if (result.status === 'valid' || result.status === 'connected') {
      showToast(result.message || `${provider} key is valid`, 'success')
    } else {
      showToast(result.message || `${provider} key test failed`, 'error')
    }
  } catch (error: any) {
    showToast(error.message || 'Failed to test provider key', 'error')
  }
}

const showToast = (message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info') => {
  toast.value = {
    show: true,
    message,
    type
  }
  setTimeout(() => {
    toast.value.show = false
  }, 5000)
}
</script>

<style scoped>
.provider-keys-container {
  @apply max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8;
}

.header {
  @apply mb-6;
}

.security-notice {
  @apply mb-6;
}
</style>

