<template>
  <div class="provider-card">
    <div class="card-header">
      <div class="flex items-center space-x-3">
        <span class="text-3xl">{{ provider.icon }}</span>
        <div>
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ provider.displayName }}
          </h3>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ provider.description }}
          </p>
        </div>
      </div>
    </div>

    <div class="card-body">
      <!-- Existing Key Display -->
      <div v-if="existingKey" class="existing-key">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center space-x-2">
            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
              <svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
              </svg>
              Configured
            </span>
            <span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
              <svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                <path fill-rule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clip-rule="evenodd" />
              </svg>
              Encrypted
            </span>
          </div>
        </div>
        <div class="text-xs text-gray-500 dark:text-gray-400 mb-4">
          Last updated: {{ formatDate(existingKey.updated_at) }}
        </div>
        <div class="flex space-x-2">
          <button
            @click="showUpdateForm = true"
            class="flex-1 px-4 py-2 text-sm font-medium text-blue-700 bg-blue-50 border border-blue-200 rounded-lg hover:bg-blue-100 dark:bg-blue-900/20 dark:text-blue-300 dark:border-blue-800 dark:hover:bg-blue-900/30 transition-colors"
          >
            Update
          </button>
          <button
            @click="handleTest"
            class="flex-1 px-4 py-2 text-sm font-medium text-gray-700 bg-gray-50 border border-gray-200 rounded-lg hover:bg-gray-100 dark:bg-gray-800 dark:text-gray-300 dark:border-gray-700 dark:hover:bg-gray-700 transition-colors"
          >
            Test
          </button>
          <button
            @click="handleDelete"
            class="px-4 py-2 text-sm font-medium text-red-700 bg-red-50 border border-red-200 rounded-lg hover:bg-red-100 dark:bg-red-900/20 dark:text-red-300 dark:border-red-800 dark:hover:bg-red-900/30 transition-colors"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
          </button>
        </div>
      </div>

      <!-- Add/Update Form -->
      <div v-else-if="showUpdateForm || !existingKey" class="key-form">
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              API Key
            </label>
            <input
              v-model="apiKeyInput"
              type="password"
              :placeholder="provider.placeholder"
              class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-800 dark:border-gray-700 dark:text-white"
              required
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              Your key will be encrypted before storage
            </p>
          </div>
          <div class="flex space-x-2">
            <button
              type="submit"
              class="flex-1 px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
            >
              {{ existingKey ? 'Update' : 'Add' }} Key
            </button>
            <button
              v-if="existingKey"
              type="button"
              @click="showUpdateForm = false"
              class="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:hover:bg-gray-700 transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface Props {
  provider: {
    name: string
    displayName: string
    icon: string
    description: string
    placeholder: string
  }
  existingKey: any | null
}

const props = defineProps<Props>()

const emit = defineEmits<{
  add: [provider: string, apiKey: string]
  update: [provider: string, apiKey: string]
  delete: [provider: string]
  test: [provider: string]
}>()

const apiKeyInput = ref('')
const showUpdateForm = ref(false)

const handleSubmit = () => {
  if (!apiKeyInput.value.trim()) {
    return
  }

  if (props.existingKey) {
    emit('update', props.provider.name, apiKeyInput.value)
  } else {
    emit('add', props.provider.name, apiKeyInput.value)
  }

  apiKeyInput.value = ''
  showUpdateForm.value = false
}

const handleDelete = () => {
  emit('delete', props.provider.name)
}

const handleTest = () => {
  emit('test', props.provider.name)
}

const formatDate = (dateString: string) => {
  const date = new Date(dateString)
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}
</script>

<style scoped>
.provider-card {
  @apply bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6 shadow-sm hover:shadow-md transition-shadow;
}

.card-header {
  @apply mb-4 pb-4 border-b border-gray-200 dark:border-gray-700;
}

.card-body {
  @apply space-y-4;
}

.existing-key {
  @apply space-y-3;
}

.key-form {
  @apply space-y-4;
}
</style>

