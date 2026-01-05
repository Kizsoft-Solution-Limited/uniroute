<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-3xl font-bold text-white">Provider Keys Management</h1>
      <p class="text-slate-400 mt-1">
        Manage system-wide provider API keys for all users
      </p>
    </div>

    <!-- Security Notice -->
    <Card>
      <div class="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
        <div class="flex items-start space-x-3">
          <svg class="w-5 h-5 text-blue-400 mt-0.5 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <div>
            <h3 class="text-sm font-medium text-blue-300 mb-1">Admin Provider Keys</h3>
            <p class="text-sm text-blue-400">
              These are system-wide provider keys that will be used for all users when they don't have their own keys configured.
              Keys are encrypted using AES-256-GCM and never stored in plaintext.
            </p>
          </div>
        </div>
      </div>
    </Card>

    <!-- Provider Keys List -->
    <div v-if="loading" class="text-center py-12">
      <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <p class="text-slate-400 mt-2">Loading provider keys...</p>
    </div>

    <div v-else class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <Card
        v-for="provider in providers"
        :key="provider.name"
        class="relative"
      >
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-3">
              <span class="text-2xl">{{ provider.icon }}</span>
              <div>
                <h3 class="font-semibold text-white">{{ provider.displayName }}</h3>
                <p class="text-sm text-slate-400">{{ provider.description }}</p>
              </div>
            </div>
            <div
              v-if="getExistingKey(provider.name)"
              class="px-2 py-1 rounded-full text-xs font-medium bg-green-500/20 text-green-400"
            >
              Configured
            </div>
          </div>

          <div v-if="editingProvider === provider.name" class="space-y-3">
            <Input
              v-model="editForm.apiKey"
              type="password"
              :placeholder="provider.placeholder"
              label="API Key"
            />
            <div class="flex space-x-2">
              <Button
                @click="handleSave(provider.name)"
                :disabled="saving || !editForm.apiKey"
                class="flex-1"
              >
                {{ getExistingKey(provider.name) ? 'Update' : 'Add' }}
              </Button>
              <Button
                @click="editingProvider = null"
                variant="secondary"
                :disabled="saving"
              >
                Cancel
              </Button>
            </div>
          </div>

          <div v-else class="space-y-2">
            <div v-if="getExistingKey(provider.name)" class="text-sm text-slate-400">
              Key configured (hidden for security)
            </div>
            <div class="flex space-x-2">
              <Button
                @click="startEditing(provider.name)"
                variant="secondary"
                class="flex-1"
              >
                {{ getExistingKey(provider.name) ? 'Update' : 'Add Key' }}
              </Button>
              <Button
                v-if="getExistingKey(provider.name)"
                @click="handleTest(provider.name)"
                :disabled="testing === provider.name"
                variant="secondary"
              >
                Test
              </Button>
              <Button
                v-if="getExistingKey(provider.name)"
                @click="handleDelete(provider.name)"
                variant="danger"
                :disabled="deleting === provider.name"
              >
                Delete
              </Button>
            </div>
          </div>
        </div>
      </Card>
    </div>

    <!-- Toast Notifications -->
    <div
      v-if="toast.show"
      class="fixed bottom-4 right-4 z-50"
    >
      <div
        class="px-4 py-3 rounded-lg shadow-lg"
        :class="
          toast.type === 'success'
            ? 'bg-green-500 text-white'
            : toast.type === 'error'
            ? 'bg-red-500 text-white'
            : 'bg-blue-500 text-white'
        "
      >
        <p>{{ toast.message }}</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { providerKeysApi, type ProviderKey } from '@/services/api/providerKeys'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import ErrorHandler from '@/utils/errorHandler'

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

const loading = ref(false)
const saving = ref(false)
const testing = ref<string | null>(null)
const deleting = ref<string | null>(null)
const providerKeys = ref<ProviderKey[]>([])
const editingProvider = ref<string | null>(null)
const editForm = ref({
  apiKey: ''
})

const toast = ref({
  show: false,
  message: '',
  type: 'success' as 'success' | 'error' | 'info'
})

const getExistingKey = (provider: string): ProviderKey | undefined => {
  return providerKeys.value.find(k => k.provider === provider)
}

const startEditing = (provider: string) => {
  editingProvider.value = provider
  editForm.value.apiKey = ''
}

const loadKeys = async () => {
  loading.value = true
  try {
    const response = await providerKeysApi.list()
    providerKeys.value = response.keys
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    showToast(`Failed to load keys: ${appError.message}`, 'error')
    ErrorHandler.logError(err, 'AdminProviderKeys')
  } finally {
    loading.value = false
  }
}

const handleSave = async (provider: string) => {
  if (!editForm.value.apiKey) return

  saving.value = true
  try {
    const existing = getExistingKey(provider)
    if (existing) {
      await providerKeysApi.update(provider, editForm.value.apiKey)
      showToast(`${provider} key updated successfully`, 'success')
    } else {
      await providerKeysApi.add(provider, editForm.value.apiKey)
      showToast(`${provider} key added successfully`, 'success')
    }
    editingProvider.value = null
    editForm.value.apiKey = ''
    await loadKeys()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    showToast(`Failed to save key: ${appError.message}`, 'error')
    ErrorHandler.logError(err, 'AdminProviderKeys')
  } finally {
    saving.value = false
  }
}

const handleTest = async (provider: string) => {
  testing.value = provider
  try {
    const response = await providerKeysApi.test(provider)
    showToast(response.message || `Test successful for ${provider}`, 'success')
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    showToast(`Test failed: ${appError.message}`, 'error')
    ErrorHandler.logError(err, 'AdminProviderKeys')
  } finally {
    testing.value = null
  }
}

const handleDelete = async (provider: string) => {
  if (!confirm(`Are you sure you want to delete the ${provider} key?`)) return

  deleting.value = provider
  try {
    await providerKeysApi.delete(provider)
    showToast(`${provider} key deleted successfully`, 'success')
    await loadKeys()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    showToast(`Failed to delete key: ${appError.message}`, 'error')
    ErrorHandler.logError(err, 'AdminProviderKeys')
  } finally {
    deleting.value = null
  }
}

const showToast = (message: string, type: 'success' | 'error' | 'info' = 'success') => {
  toast.value = { show: true, message, type }
  setTimeout(() => {
    toast.value.show = false
  }, 3000)
}

onMounted(() => {
  loadKeys()
})
</script>

