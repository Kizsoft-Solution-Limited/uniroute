<template>
  <div class="flex flex-col h-[calc(100vh-8rem)] sm:h-[calc(100vh-12rem)] max-h-[800px]">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-4 gap-4">
      <div class="flex-1">
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 dark:text-white" id="page-title">AI Chat</h1>
        <p class="text-sm sm:text-base text-gray-600 dark:text-gray-400 mt-1">
          Chat with any AI model through UniRoute
        </p>
      </div>
      <div class="flex flex-wrap items-center gap-2 sm:gap-4">
        <!-- Model Selector -->
        <div class="flex items-center space-x-2 flex-1 sm:flex-initial min-w-[140px]">
          <label class="text-xs sm:text-sm text-gray-600 dark:text-gray-400 whitespace-nowrap">Model:</label>
          <select
            v-model="selectedModel"
            class="flex-1 sm:flex-initial px-2 sm:px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white text-xs sm:text-sm"
            aria-label="Select AI model"
          >
            <!-- OpenAI Models -->
            <optgroup label="OpenAI">
              <option value="gpt-4o">GPT-4o (Latest)</option>
              <option value="gpt-4o-mini">GPT-4o Mini</option>
              <option value="gpt-4-turbo">GPT-4 Turbo</option>
              <option value="gpt-4">GPT-4</option>
              <option value="gpt-3.5-turbo">GPT-3.5 Turbo</option>
            </optgroup>
            <!-- Anthropic Models -->
            <optgroup label="Anthropic (Claude)">
              <option value="claude-3-5-sonnet-20241022">Claude 3.5 Sonnet (Latest)</option>
              <option value="claude-3-5-haiku-20241022">Claude 3.5 Haiku</option>
              <option value="claude-3-opus-20240229">Claude 3 Opus</option>
              <option value="claude-3-sonnet-20240229">Claude 3 Sonnet</option>
              <option value="claude-3-haiku-20240307">Claude 3 Haiku</option>
            </optgroup>
            <!-- Google Models -->
            <optgroup label="Google (Gemini)">
              <option value="gemini-2.0-flash-exp">Gemini 2.0 Flash (Experimental)</option>
              <option value="gemini-1.5-pro-latest">Gemini 1.5 Pro (Latest)</option>
              <option value="gemini-1.5-pro">Gemini 1.5 Pro</option>
              <option value="gemini-1.5-flash">Gemini 1.5 Flash</option>
              <option value="gemini-pro">Gemini Pro</option>
            </optgroup>
            <!-- Local Models -->
            <optgroup label="Local">
              <option value="llama2">Llama 2</option>
              <option value="mistral">Mistral</option>
            </optgroup>
          </select>
        </div>
        <!-- Settings -->
        <button
          @click="showSettings = !showSettings"
          class="p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          aria-label="Chat settings"
        >
          <Settings class="w-5 h-5" />
        </button>
        <!-- Clear Chat -->
        <button
          @click="clearChat"
          class="px-3 sm:px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors text-xs sm:text-sm whitespace-nowrap"
        >
          Clear
        </button>
      </div>
    </div>

    <!-- Settings Panel -->
    <Card v-if="showSettings" class="mb-4">
      <div class="p-4 space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Temperature: {{ temperature.toFixed(1) }}
          </label>
          <input
            v-model.number="temperature"
            type="range"
            min="0"
            max="2"
            step="0.1"
            class="w-full"
          />
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            Controls randomness (0.0 = deterministic, 2.0 = very creative)
          </p>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Max Tokens: {{ maxTokens }}
          </label>
          <input
            v-model.number="maxTokens"
            type="number"
            min="1"
            max="4000"
            class="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          />
        </div>
      </div>
    </Card>

    <!-- Messages Container -->
    <Card class="flex-1 flex flex-col overflow-hidden mb-4 min-h-0">
      <div
        ref="messagesContainer"
        class="flex-1 overflow-y-auto p-3 sm:p-6 space-y-3 sm:space-y-4"
        role="log"
        aria-live="polite"
        aria-label="Chat messages"
      >
        <!-- Welcome Message -->
        <div v-if="messages.length === 0" class="text-center py-12">
          <MessageSquare class="w-16 h-16 text-gray-400 mx-auto mb-4" />
          <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
            Start a conversation
          </h3>
          <p class="text-gray-600 dark:text-gray-400 mb-4">
            Ask me anything! I can help with questions, coding, writing, and more.
          </p>
          <div class="flex flex-wrap gap-2 justify-center">
            <button
              v-for="suggestion in suggestions"
              :key="suggestion"
              @click="sendMessage(suggestion)"
              class="px-4 py-2 text-sm rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
            >
              {{ suggestion }}
            </button>
          </div>
        </div>

        <!-- Messages -->
        <div
          v-for="(message, index) in messages"
          :key="index"
          :class="[
            'flex',
            message.role === 'user' ? 'justify-end' : 'justify-start'
          ]"
        >
          <div
            :class="[
              'max-w-[85%] sm:max-w-[80%] rounded-lg p-3 sm:p-4 text-sm sm:text-base',
              message.role === 'user'
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-white'
            ]"
          >
            <div class="flex items-start space-x-2">
              <div class="flex-1">
                <p class="whitespace-pre-wrap break-words">{{ message.content }}</p>
                <div
                  v-if="message.metadata"
                  class="mt-2 text-xs opacity-75 flex items-center space-x-3"
                >
                  <span v-if="message.metadata.tokens">
                    {{ message.metadata.tokens }} tokens
                  </span>
                  <span v-if="message.metadata.cost">
                    ${{ message.metadata.cost.toFixed(6) }}
                  </span>
                  <span v-if="message.metadata.provider" class="capitalize">
                    {{ message.metadata.provider }}
                  </span>
                  <span v-if="message.metadata.latency">
                    {{ message.metadata.latency }}ms
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Loading Indicator -->
        <div v-if="loading" class="flex justify-start">
          <div class="bg-gray-100 dark:bg-gray-800 rounded-lg p-4">
            <div class="flex items-center space-x-2">
              <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
              <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0.2s"></div>
              <div class="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style="animation-delay: 0.4s"></div>
            </div>
          </div>
        </div>
      </div>
    </Card>

    <!-- Input Area -->
    <Card>
      <div class="p-3 sm:p-4">
        <form @submit.prevent="handleSend" class="flex items-end gap-2 sm:gap-2">
          <div class="flex-1 min-w-0">
            <textarea
              v-model="inputMessage"
              @keydown.enter.exact.prevent="handleSend"
              @keydown.shift.enter.exact="inputMessage += '\n'"
              rows="3"
              placeholder="Type your message... (Shift+Enter for new line)"
              class="w-full px-3 sm:px-4 py-3 sm:py-3 text-base sm:text-base rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-[80px] sm:min-h-[60px]"
              :disabled="loading"
              aria-label="Message input"
            ></textarea>
          </div>
          <button
            type="submit"
            :disabled="loading || !inputMessage.trim()"
            class="px-4 sm:px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center space-x-1 sm:space-x-2 flex-shrink-0 h-[80px] sm:h-auto"
            aria-label="Send message"
          >
            <Send v-if="!loading" class="w-5 h-5" />
            <div v-else class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            <span class="hidden sm:inline">Send</span>
          </button>
        </form>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2 hidden sm:block">
          Press Enter to send, Shift+Enter for new line
        </p>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import Card from '@/components/ui/Card.vue'
import { MessageSquare, Send, Settings } from 'lucide-vue-next'
import { chatApi, type Message, type ChatResponse } from '@/services/api/chat'
import { useToast } from '@/composables/useToast'

const { showToast } = useToast()

interface ChatMessage extends Message {
  metadata?: {
    tokens?: number
    cost?: number
    provider?: string
    latency?: number
  }
}

const messages = ref<ChatMessage[]>([])
const inputMessage = ref('')
const loading = ref(false)
const selectedModel = ref('gpt-4')
const temperature = ref(0.7)
const maxTokens = ref(1000)
const showSettings = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)

const suggestions = [
  'Explain quantum computing',
  'Write a Python function',
  'What is machine learning?',
  'Help me debug this code'
]

// Auto-scroll to bottom when new messages arrive
watch(messages, () => {
  nextTick(() => {
    scrollToBottom()
  })
}, { deep: true })

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const handleSend = async () => {
  if (!inputMessage.value.trim() || loading.value) return

  const userMessage: ChatMessage = {
    role: 'user',
    content: inputMessage.value.trim()
  }

  messages.value.push(userMessage)
  inputMessage.value = ''
  loading.value = true

  try {
    // Build messages array (include conversation history)
    const chatMessages: Message[] = messages.value
      .filter(m => m.role !== 'system' || messages.value.indexOf(m) === 0)
      .map(m => ({ role: m.role, content: m.content }))

    const response: ChatResponse = await chatApi.chat({
      model: selectedModel.value,
      messages: chatMessages,
      temperature: temperature.value,
      max_tokens: maxTokens.value
    }, true) // Use JWT auth for frontend (default is true)

    const assistantMessage: ChatMessage = {
      role: 'assistant',
      content: response.choices[0].message.content,
      metadata: {
        tokens: response.usage.total_tokens,
        cost: response.cost,
        provider: response.provider,
        latency: response.latency_ms
      }
    }

    messages.value.push(assistantMessage)
  } catch (error: any) {
    console.error('Chat error:', error)
    const errorMessage = error.response?.data?.error || error.message || 'Failed to get response'
    showToast(errorMessage, 'error')

    // Add error message to chat
    const errorChatMessage: ChatMessage = {
      role: 'assistant',
      content: `Error: ${errorMessage}`,
      metadata: {
        provider: 'error'
      }
    }
    messages.value.push(errorChatMessage)
  } finally {
    loading.value = false
  }
}

const sendMessage = (message: string) => {
  inputMessage.value = message
  handleSend()
}

const clearChat = () => {
  if (confirm('Are you sure you want to clear the chat history?')) {
    messages.value = []
    showToast('Chat history cleared', 'success')
  }
}

onMounted(() => {
  scrollToBottom()
})
</script>

<style scoped>
/* Custom scrollbar */
.overflow-y-auto::-webkit-scrollbar {
  width: 8px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background: rgba(156, 163, 175, 0.5);
  border-radius: 4px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background: rgba(156, 163, 175, 0.7);
}
</style>

