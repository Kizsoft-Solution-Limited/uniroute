<template>
  <div class="flex h-[calc(100vh-8rem)] sm:h-[calc(100vh-12rem)] max-h-[800px] gap-4">
    <!-- Conversations Sidebar -->
    <div class="w-64 flex-shrink-0 hidden lg:block">
      <Card class="h-full flex flex-col">
        <div class="p-4 border-b border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between mb-2">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Conversations</h2>
            <button
              @click="createNewConversation"
              class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-400"
              title="New conversation"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
        </div>
        <div class="flex-1 overflow-y-auto p-2">
          <div v-if="conversationsLoading" class="text-center py-4 text-gray-500 dark:text-gray-400">
            Loading...
          </div>
          <div v-else-if="conversations.length === 0" class="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
            No conversations yet
          </div>
          <div v-else class="space-y-1">
            <button
              v-for="conv in conversations"
              :key="conv.id"
              @click="loadConversation(conv.id)"
              :class="[
                'w-full text-left px-3 py-2 rounded-lg text-sm transition-colors group',
                currentConversationId === conv.id
                  ? 'bg-blue-100 dark:bg-blue-900 text-blue-900 dark:text-blue-100'
                  : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300'
              ]"
            >
              <div class="flex items-center justify-between">
                <div class="flex-1 min-w-0">
                  <div class="font-medium truncate">
                    {{ conv.title || 'New Conversation' }}
                  </div>
                  <div class="text-xs opacity-75 truncate">
                    {{ formatDate(conv.updated_at ?? conv.UpdatedAt) }}
                  </div>
                </div>
                <button
                  @click.stop="deleteConversation(conv.id)"
                  class="ml-2 p-1 rounded hover:bg-red-100 dark:hover:bg-red-900 opacity-0 group-hover:opacity-100 transition-opacity"
                  title="Delete conversation"
                >
                  <X class="w-4 h-4" />
                </button>
              </div>
            </button>
          </div>
        </div>
      </Card>
    </div>

    <!-- Main Chat Area -->
    <div class="flex-1 flex flex-col min-w-0">
    <!-- Mobile Header with Conversations Button -->
    <div class="lg:hidden mb-4">
      <div class="flex items-center justify-between gap-2">
        <button
          @click.stop="showMobileConversations = true"
          type="button"
          class="p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          aria-label="Open conversations"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <h1 class="text-xl font-bold text-gray-900 dark:text-white flex-1 text-center">AI Chat</h1>
        <button
          @click="createNewConversation"
          class="p-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          aria-label="New conversation"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-4 gap-4">
      <div class="flex-1 hidden lg:block">
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
            <template v-for="group in modelGroups" :key="group.label">
              <optgroup :label="group.label">
                <option v-for="opt in group.options" :key="opt.value" :value="opt.value">
                  {{ opt.label }}
                </option>
              </optgroup>
            </template>
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
                <!-- Text content -->
                <p v-if="typeof message.content === 'string'" class="whitespace-pre-wrap break-words">{{ message.content }}</p>
                
                <!-- Multimodal content (images + audio + text) -->
                <div v-else-if="Array.isArray(message.content)" class="space-y-2">
                  <template v-for="(part, partIndex) in message.content" :key="partIndex">
                    <p v-if="part.type === 'text'" class="whitespace-pre-wrap break-words">{{ part.text }}</p>
                    <div v-else-if="part.type === 'image_url' && part.image_url" class="relative">
                      <img 
                        :src="getImageUrl(part.image_url.url)" 
                        :data-part-index="partIndex"
                        alt="Attached image"
                        class="max-w-full h-auto rounded-lg border border-gray-300 dark:border-gray-600 max-h-96 object-contain"
                        @error="handleImageError($event)"
                        @load="handleImageLoad($event)"
                      />
                      <div v-if="imageErrors.has(partIndex)" class="mt-2 p-2 bg-red-50 dark:bg-red-900/20 rounded text-sm text-red-700 dark:text-red-300">
                        Failed to load image
                      </div>
                    </div>
                    <div 
                      v-else-if="part.type === 'audio_url' && part.audio_url"
                      class="bg-gray-100 dark:bg-gray-800 rounded-lg p-3 border border-gray-300 dark:border-gray-600"
                    >
                      <div class="flex items-center gap-2">
                        <svg class="w-5 h-5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
                        </svg>
                        <span class="text-sm text-gray-700 dark:text-gray-300">Audio file</span>
                        <audio :src="part.audio_url.url" controls class="flex-1 h-8" />
                      </div>
                    </div>
                  </template>
                </div>
                
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
    <Card class="sticky bottom-0 z-10">
      <div class="p-2 sm:p-4">
        <!-- Attached Images Preview -->
        <div v-if="attachedImages.length > 0" class="mb-2 sm:mb-3 flex flex-wrap gap-2">
          <div
            v-for="(image, index) in attachedImages"
            :key="`img-${index}`"
            class="relative group"
          >
            <img
              :src="image.preview"
              alt="Preview"
              class="w-16 h-16 sm:w-20 sm:h-20 object-cover rounded-lg border border-gray-300 dark:border-gray-600"
              @error="handlePreviewImageError($event, index)"
            />
            <button
              @click="removeImage(index)"
              class="absolute -top-1 -right-1 sm:-top-2 sm:-right-2 bg-red-500 text-white rounded-full p-1 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity"
              aria-label="Remove image"
            >
              <X class="w-3 h-3" />
            </button>
          </div>
        </div>

        <!-- Attached Audio/Voice Preview -->
        <div v-if="attachedAudios.length > 0" class="mb-2 sm:mb-3 flex flex-wrap gap-2">
          <div
            v-for="(audio, index) in attachedAudios"
            :key="`audio-${index}`"
            class="relative group bg-gray-100 dark:bg-gray-800 rounded-lg p-2 border border-gray-300 dark:border-gray-600"
          >
            <div class="flex items-center gap-2">
              <svg class="w-4 h-4 sm:w-5 sm:h-5 text-gray-600 dark:text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
              </svg>
          <div class="flex-1 min-w-0">
                <p class="text-xs text-gray-700 dark:text-gray-300 truncate">{{ audio.file.name }}</p>
                <p v-if="audio.duration" class="text-xs text-gray-500 dark:text-gray-400">{{ formatDuration(audio.duration) }}</p>
              </div>
              <audio :src="audio.preview" controls class="h-6 sm:h-8 flex-1 max-w-[120px] sm:max-w-none" />
            </div>
            <button
              @click="removeAudio(index)"
              class="absolute -top-1 -right-1 sm:-top-2 sm:-right-2 bg-red-500 text-white rounded-full p-1 opacity-100 sm:opacity-0 sm:group-hover:opacity-100 transition-opacity"
              aria-label="Remove audio"
            >
              <X class="w-3 h-3" />
            </button>
          </div>
        </div>

        <!-- Action Buttons Row (Mobile) -->
        <div class="flex items-center justify-between gap-2 mb-2 sm:hidden">
          <div class="flex items-center gap-1">
            <button
              type="button"
              @click="imageInputRef?.click()"
              class="p-2 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
              title="Attach image(s)"
              aria-label="Attach image"
            >
              <ImageIcon class="w-5 h-5" />
            </button>
            <input
              ref="imageInputRef"
              type="file"
              accept="image/*"
              multiple
              @change="handleImageSelect"
              class="hidden"
              aria-label="Image input"
            />
            <button
              type="button"
              @click="audioInputRef?.click()"
              class="p-2 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors"
              title="Attach audio file(s)"
              aria-label="Attach audio"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
              </svg>
            </button>
            <input
              ref="audioInputRef"
              type="file"
              accept="audio/*"
              multiple
              @change="handleAudioSelect"
              class="hidden"
              aria-label="Audio input"
            />
            <button
              type="button"
              @click="toggleVoiceRecording"
              :class="[
                'p-2 rounded-lg transition-colors',
                isRecording
                  ? 'bg-red-500 text-white hover:bg-red-600'
                  : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
              ]"
              :title="isRecording ? 'Stop recording' : 'Start voice recording'"
              aria-label="Voice recording"
            >
              <svg v-if="!isRecording" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
              </svg>
              <div v-else class="w-5 h-5 bg-white rounded-full animate-pulse"></div>
            </button>
          </div>
        </div>

        <form @submit.prevent="handleSend" class="flex flex-col sm:flex-row items-end gap-2 sm:gap-2">
          <!-- Desktop: Buttons on left, textarea in middle -->
          <div class="hidden sm:flex items-end gap-2 flex-1 min-w-0 w-full">
            <button
              type="button"
              @click="imageInputRef?.click()"
              class="p-2 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors flex-shrink-0"
              title="Attach image(s)"
              aria-label="Attach image"
            >
              <ImageIcon class="w-5 h-5" />
            </button>
            <input
              ref="imageInputRef"
              type="file"
              accept="image/*"
              multiple
              @change="handleImageSelect"
              class="hidden"
              aria-label="Image input"
            />
            <button
              type="button"
              @click="audioInputRef?.click()"
              class="p-2 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors flex-shrink-0"
              title="Attach audio/voice file(s)"
              aria-label="Attach audio"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
              </svg>
            </button>
            <input
              ref="audioInputRef"
              type="file"
              accept="audio/*"
              multiple
              @change="handleAudioSelect"
              class="hidden"
              aria-label="Audio input"
            />
            <button
              type="button"
              @click="toggleVoiceRecording"
              :class="[
                'p-2 rounded-lg transition-colors flex-shrink-0',
                isRecording
                  ? 'bg-red-500 text-white hover:bg-red-600'
                  : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
              ]"
              :title="isRecording ? 'Stop recording' : 'Start voice recording'"
              aria-label="Voice recording"
            >
              <svg v-if="!isRecording" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11a7 7 0 01-7 7m0 0a7 7 0 01-7-7m7 7v4m0 0H8m4 0h4m-4-8a3 3 0 01-3-3V5a3 3 0 116 0v6a3 3 0 01-3 3z" />
              </svg>
              <div v-else class="w-5 h-5 bg-white rounded-full animate-pulse"></div>
            </button>
            <textarea
              v-model="inputMessage"
              @keydown.enter.exact.prevent="handleSend"
              @keydown.shift.enter.exact="inputMessage += '\n'"
              rows="3"
              placeholder="Type your message... (Shift+Enter for new line)"
              class="flex-1 px-4 py-3 text-base rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-[60px]"
              :disabled="loading"
              aria-label="Message input"
            ></textarea>
          </div>

          <!-- Mobile: Textarea full width with proper spacing -->
          <div class="flex-1 min-w-0 w-full sm:hidden">
            <textarea
              v-model="inputMessage"
              @keydown.enter.exact.prevent="handleSend"
              @keydown.shift.enter.exact="inputMessage += '\n'"
              rows="4"
              placeholder="Type your message..."
              class="w-full px-4 py-3 text-base rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
              :disabled="loading"
              aria-label="Message input"
            ></textarea>
          </div>

          <!-- Send Button -->
          <button
            type="submit"
            :disabled="loading || (!inputMessage.trim() && attachedImages.length === 0 && attachedAudios.length === 0)"
            class="w-full sm:w-auto px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors flex items-center justify-center gap-2 flex-shrink-0 sm:h-auto min-h-[48px] sm:min-h-0"
            aria-label="Send message"
          >
            <Send v-if="!loading" class="w-5 h-5" />
            <div v-else class="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
            <span class="sm:hidden">Send</span>
            <span class="hidden sm:inline">Send</span>
          </button>
        </form>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2 hidden sm:block">
          Press Enter to send, Shift+Enter for new line. Click icons to attach images, audio files, or use voice recording.
        </p>
        <div v-if="isRecording" class="mt-2 p-2 bg-red-50 dark:bg-red-900/20 rounded-lg flex items-center gap-2">
          <div class="w-3 h-3 bg-red-500 rounded-full animate-pulse"></div>
          <span class="text-xs text-red-700 dark:text-red-300">Recording... Click again to stop</span>
        </div>
        <div v-if="transcribedText" class="mt-2 p-2 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
          <p class="text-xs text-blue-700 dark:text-blue-300 mb-1">Transcribed:</p>
          <p class="text-sm text-blue-900 dark:text-blue-100">{{ transcribedText }}</p>
        </div>
      </div>
    </Card>
    </div>

    <!-- Mobile Conversations Drawer -->
    <Teleport to="body">
      <div
        v-if="showMobileConversations"
        class="fixed inset-0 z-[9999] lg:hidden dark"
        @click.self="showMobileConversations = false"
        @touchmove.prevent
        style="touch-action: none;"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black bg-opacity-50" @click="showMobileConversations = false"></div>
        
        <!-- Drawer -->
        <div class="absolute left-0 top-0 bottom-0 w-80 max-w-[85vw] bg-gray-900 shadow-xl flex flex-col">
          <div class="p-4 border-b border-gray-700 flex items-center justify-between">
            <h2 class="text-lg font-semibold text-white">Conversations</h2>
            <button
              @click="showMobileConversations = false"
              class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400"
              aria-label="Close"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="flex-1 overflow-y-auto p-2">
            <div v-if="conversationsLoading" class="text-center py-4 text-gray-400">
              Loading...
            </div>
            <div v-else-if="conversations.length === 0" class="text-center py-4 text-gray-400 text-sm">
              No conversations yet
            </div>
            <div v-else class="space-y-1">
              <button
                v-for="conv in conversations"
                :key="conv.id"
                @click="loadConversation(conv.id)"
                :class="[
                  'w-full text-left px-3 py-2 rounded-lg text-sm transition-colors group',
                  currentConversationId === conv.id
                    ? 'bg-blue-900 text-blue-100'
                    : 'hover:bg-gray-700 text-gray-300'
                ]"
              >
                <div class="flex items-center justify-between">
                  <div class="flex-1 min-w-0">
                    <div class="font-medium truncate">
                      {{ conv.title || 'New Conversation' }}
                    </div>
                    <div class="text-xs opacity-75 truncate">
                      {{ formatDate(conv.updated_at ?? conv.UpdatedAt) }}
                    </div>
                  </div>
                  <button
                    @click.stop="deleteConversation(conv.id)"
                    class="ml-2 p-1 rounded hover:bg-red-900 opacity-0 group-hover:opacity-100 transition-opacity"
                    title="Delete conversation"
                  >
                    <X class="w-4 h-4" />
                  </button>
                </div>
              </button>
            </div>
          </div>
          <div class="p-4 border-t border-gray-700">
            <button
              @click="createNewConversation"
              class="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors flex items-center justify-center gap-2"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              <span>New Conversation</span>
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import Card from '@/components/ui/Card.vue'
import { MessageSquare, Send, Settings, Image as ImageIcon, X } from 'lucide-vue-next'
import { chatApi, type Message, type ChatResponse, type ContentPart, type StreamChunk } from '@/services/api/chat'
import { conversationsApi, type Conversation } from '@/services/api/conversations'
import { providersApi, type ProviderInfo } from '@/services/api/providers'
import { useToast } from '@/composables/useToast'

const { showToast } = useToast()

interface AttachedImage {
  file: File
  dataUrl: string
  preview: string
}

interface AttachedAudio {
  file: File
  dataUrl: string
  preview: string
  duration?: number
}

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
const attachedImages = ref<AttachedImage[]>([])
const attachedAudios = ref<AttachedAudio[]>([])
const imageInputRef = ref<HTMLInputElement | null>(null)
const audioInputRef = ref<HTMLInputElement | null>(null)
const loading = ref(false)
const selectedModel = ref('gpt-4')
const temperature = ref(0.7)
const maxTokens = ref(1000)
const showSettings = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)

// Conversation management
const conversations = ref<Conversation[]>([])
const conversationsLoading = ref(false)
const currentConversationId = ref<string | null>(null)

// Voice recording
const isRecording = ref(false)
const recognition: any = ref(null)
const transcribedText = ref('')
const speechSynthesis = ref<SpeechSynthesis | null>(null)

// Image error tracking
const imageErrors = ref(new Set<number>())

// Mobile conversations drawer
const showMobileConversations = ref(false)

const suggestions = [
  'Explain quantum computing',
  'Write a Python function',
  'What is machine learning?',
  'Help me debug this code'
]

const PROVIDER_DISPLAY_NAMES: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic (Claude)',
  google: 'Google (Gemini)',
  vllm: 'vLLM',
  ollama: 'Ollama',
  local: 'Ollama'
}

// Must match models on the Ollama host (ollama list). Fallback when API cannot reach Ollama.
const DEFAULT_OLLAMA_MODELS = [
  { value: 'llama3.2:3b', label: 'Llama 3.2 3B' },
  { value: 'llama3.2:latest', label: 'Llama 3.2 (latest)' },
  { value: 'mistral:latest', label: 'Mistral (latest)' }
]

// Fallback when backend cannot reach vLLM. Real list comes from the host (GET /v1/models).
// On the host, run e.g.: vllm serve <model> (e.g. mistralai/Mistral-7B-Instruct-v0.2)
const DEFAULT_VLLM_MODELS = [
  { value: 'mistralai/Mistral-7B-Instruct-v0.2', label: 'Mistral 7B Instruct' },
  { value: 'meta-llama/Llama-2-7b-chat-hf', label: 'Llama 2 7B Chat' }
]

const staticModelGroups = [
  { label: 'OpenAI', options: [
    { value: 'gpt-4o', label: 'GPT-4o (Latest)' },
    { value: 'gpt-4o-mini', label: 'GPT-4o Mini' },
    { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
    { value: 'gpt-4', label: 'GPT-4' },
    { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' }
  ]},
  { label: 'Anthropic (Claude)', options: [
    { value: 'claude-3-5-sonnet-20241022', label: 'Claude 3.5 Sonnet (Latest)' },
    { value: 'claude-3-5-haiku-20241022', label: 'Claude 3.5 Haiku' },
    { value: 'claude-3-opus-20240229', label: 'Claude 3 Opus' },
    { value: 'claude-3-sonnet-20240229', label: 'Claude 3 Sonnet' },
    { value: 'claude-3-haiku-20240307', label: 'Claude 3 Haiku' }
  ]},
  { label: 'Google (Gemini)', options: [
    { value: 'gemini-2.0-flash-exp', label: 'Gemini 2.0 Flash (Experimental)' },
    { value: 'gemini-1.5-pro-latest', label: 'Gemini 1.5 Pro (Latest)' },
    { value: 'gemini-1.5-pro', label: 'Gemini 1.5 Pro' },
    { value: 'gemini-1.5-flash', label: 'Gemini 1.5 Flash' },
    { value: 'gemini-pro', label: 'Gemini Pro' }
  ]},
  { label: 'Ollama', options: DEFAULT_OLLAMA_MODELS },
  { label: 'vLLM', options: DEFAULT_VLLM_MODELS }
]

const apiProviders = ref<ProviderInfo[]>([])
const modelGroups = computed(() => {
  if (apiProviders.value.length === 0) return staticModelGroups
  return apiProviders.value.map((p) => {
    let options = p.models.map((m) => ({ value: m, label: m }))
    if (options.length === 0 && (p.name === 'local' || p.name === 'ollama')) {
      options = DEFAULT_OLLAMA_MODELS
    }
    if (options.length === 0 && p.name === 'vllm') {
      options = DEFAULT_VLLM_MODELS
    }
    return {
      label: PROVIDER_DISPLAY_NAMES[p.name] || p.name,
      options
    }
  })
})

// Auto-scroll to bottom when new messages arrive
watch(messages, () => {
  nextTick(() => {
    scrollToBottom()
  })
}, { deep: true })

// Lock body scroll when mobile drawer is open
watch(showMobileConversations, (isOpen) => {
  if (isOpen) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const handleSend = async () => {
  // Use transcribed text if available, otherwise use input
  const textToSend = transcribedText.value || inputMessage.value.trim()
  
  if ((!textToSend && attachedImages.value.length === 0 && attachedAudios.value.length === 0) || loading.value) return

  // Create conversation if none exists
  if (!currentConversationId.value) {
    try {
      const newConv = await conversationsApi.createConversation({
        title: textToSend.substring(0, 50) || 'New Conversation',
        model: selectedModel.value
      })
      currentConversationId.value = newConv.id
      await loadConversations()
    } catch (error) {
      console.error('Failed to create conversation:', error)
    }
  }

  // Build message content - support multimodal (images + audio + text)
  let messageContent: string | ContentPart[] = textToSend
  
  // If images or audio are attached, create multimodal content
  if (attachedImages.value.length > 0 || attachedAudios.value.length > 0) {
    const parts: ContentPart[] = []
    
    // Add text first if present
    if (textToSend) {
      parts.push({
        type: 'text',
        text: textToSend
      })
    }
    
    // Add all images
    for (const image of attachedImages.value) {
      parts.push({
        type: 'image_url',
        image_url: {
          url: image.dataUrl // base64 data URL
        }
      })
    }
    
    // Add all audio files
    for (const audio of attachedAudios.value) {
      parts.push({
        type: 'audio_url',
        audio_url: {
          url: audio.dataUrl // base64 data URL
        }
      })
    }
    
    messageContent = parts
  }

  const userMessage: ChatMessage = {
    role: 'user',
    content: messageContent
  }

  messages.value.push(userMessage)
  inputMessage.value = ''
  transcribedText.value = '' // Clear transcribed text
  // Clear attached files (only revoke blob URLs)
  attachedImages.value.forEach(img => {
    if (img.preview.startsWith('blob:')) {
      URL.revokeObjectURL(img.preview)
    }
  })
  attachedAudios.value.forEach(audio => {
    if (audio.preview.startsWith('blob:')) {
      URL.revokeObjectURL(audio.preview)
    }
  })
  attachedImages.value = []
  attachedAudios.value = []
  loading.value = true

  try {
    // Build messages array (include conversation history)
    const chatMessages: Message[] = messages.value
      .filter(m => m.role !== 'system' || messages.value.indexOf(m) === 0)
      .map(m => ({ 
        role: m.role, 
        content: m.content // Can be string or ContentPart[]
      }))

    const chatRequestData: any = {
      model: selectedModel.value,
      messages: chatMessages,
      temperature: temperature.value,
      max_tokens: maxTokens.value
    }

    // Add conversation_id if we have one
    if (currentConversationId.value) {
      chatRequestData.conversation_id = currentConversationId.value
    }

    // Use streaming for better UX (real-time response)
    let assistantMessage: ChatMessage = {
      role: 'assistant',
      content: '',
      metadata: {}
    }
    messages.value.push(assistantMessage)
    const assistantMessageIndex = messages.value.length - 1

    try {
      await chatApi.chatStream(
        chatRequestData,
        (chunk) => {
          // Append chunk content to assistant message
          if (chunk.content) {
            assistantMessage.content += chunk.content
            messages.value[assistantMessageIndex] = { ...assistantMessage }
            scrollToBottom()
          }

          // Update metadata when stream completes
          if (chunk.done && chunk.usage) {
            assistantMessage.metadata = {
              tokens: chunk.usage.total_tokens,
              cost: 0, // Cost calculation can be added later
              provider: chunk.provider || 'unknown',
              latency: 0 // Latency tracking can be added later
            }
            messages.value[assistantMessageIndex] = { ...assistantMessage }
          }

          // Handle errors
          if (chunk.error) {
            throw new Error(chunk.error)
          }
        },
        (error) => {
          // Remove incomplete message on error
          messages.value.splice(assistantMessageIndex, 1)
          showToast('Failed to get response: ' + error.message, 'error')
        },
        (usage) => {
          // Stream completed successfully
          if (usage) {
            assistantMessage.metadata = {
              ...assistantMessage.metadata,
              tokens: usage.total_tokens
            }
            messages.value[assistantMessageIndex] = { ...assistantMessage }
          }
        },
        true // Use JWT auth for frontend
      )
    } catch (error: any) {
      // Remove incomplete message on error
      if (messages.value[assistantMessageIndex]?.content === '') {
        messages.value.splice(assistantMessageIndex, 1)
      }
      showToast('Failed to get response: ' + (error.message || 'Unknown error'), 'error')
    }

    // Speak the response if text-to-speech is available (after streaming completes)
    if (speechSynthesis.value && assistantMessage.content) {
      speakText(assistantMessage.content)
    }

    // Note: Success toast is shown in the streaming callback
    await loadConversations() // Refresh conversation list
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

const handleImageSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return

  Array.from(target.files).forEach((file) => {
    if (!file.type.startsWith('image/')) {
      showToast('Please select an image file', 'error')
      return
    }

    // Check file size (max 10MB)
    if (file.size > 10 * 1024 * 1024) {
      showToast('Image size must be less than 10MB', 'error')
      return
    }

    const reader = new FileReader()
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string
      // Use data URL for preview to avoid CSP blob: issues
      attachedImages.value.push({
        file,
        dataUrl,
        preview: dataUrl // Use data URL instead of blob URL
      })
    }
    reader.readAsDataURL(file)
  })

  // Reset input
  target.value = ''
}

const handleAudioSelect = (event: Event) => {
  const target = event.target as HTMLInputElement
  if (!target.files || target.files.length === 0) return

  Array.from(target.files).forEach((file) => {
    if (!file.type.startsWith('audio/')) {
      showToast('Please select an audio file', 'error')
      return
    }

    // Check file size (max 25MB for audio)
    if (file.size > 25 * 1024 * 1024) {
      showToast('Audio file size must be less than 25MB', 'error')
      return
    }

    const reader = new FileReader()
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string
      // Use data URL for audio preview to avoid CSP blob: issues
      const audio = new Audio()
      audio.src = dataUrl // Use data URL instead of blob URL
      
      audio.addEventListener('loadedmetadata', () => {
        attachedAudios.value.push({
          file,
          dataUrl,
          preview: dataUrl, // Use data URL instead of blob URL to avoid CSP issues
          duration: audio.duration
        })
      }, { once: true })
      
      // Fallback if metadata doesn't load
      setTimeout(() => {
        if (!attachedAudios.value.find(a => a.file === file)) {
          attachedAudios.value.push({
            file,
            dataUrl,
            preview: dataUrl // Use data URL instead of blob URL
          })
        }
      }, 1000)
    }
    reader.readAsDataURL(file)
  })

  // Reset input
  target.value = ''
}

const formatDuration = (seconds: number): string => {
  if (!seconds || isNaN(seconds)) return ''
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const removeImage = (index: number) => {
  // Only revoke if it's a blob URL (not a data URL)
  const preview = attachedImages.value[index].preview
  if (preview.startsWith('blob:')) {
    URL.revokeObjectURL(preview)
  }
  attachedImages.value.splice(index, 1)
}

const removeAudio = (index: number) => {
  // Revoke object URL to free memory
  URL.revokeObjectURL(attachedAudios.value[index].preview)
  attachedAudios.value.splice(index, 1)
}

const sendMessage = (message: string) => {
  inputMessage.value = message
  handleSend()
}

const clearChat = () => {
  if (confirm('Are you sure you want to clear the chat history?')) {
    // Clean up preview URLs (only revoke blob URLs)
    attachedImages.value.forEach(img => {
      if (img.preview.startsWith('blob:')) {
        URL.revokeObjectURL(img.preview)
      }
    })
    attachedAudios.value.forEach(audio => {
      if (audio.preview.startsWith('blob:')) {
        URL.revokeObjectURL(audio.preview)
      }
    })
    attachedImages.value = []
    attachedAudios.value = []
    messages.value = []
    currentConversationId.value = null
    showToast('Chat history cleared', 'success')
  }
}

// Conversation management functions
const loadConversations = async () => {
  try {
    conversationsLoading.value = true
    conversations.value = await conversationsApi.listConversations(50, 0)
  } catch (error) {
    console.error('Failed to load conversations:', error)
    showToast('Failed to load conversations', 'error')
  } finally {
    conversationsLoading.value = false
  }
}

const loadConversation = async (id: string) => {
  try {
    const data = await conversationsApi.getConversation(id)
    currentConversationId.value = id
    selectedModel.value = data.conversation.model || 'gpt-4'
    
    // Convert stored messages to ChatMessage format
    messages.value = data.messages.map(msg => ({
      role: msg.role,
      content: msg.content,
      metadata: msg.metadata || undefined
    }))
    
    showToast('Conversation loaded', 'success')
    showMobileConversations.value = false // Close mobile drawer
  } catch (error) {
    console.error('Failed to load conversation:', error)
    showToast('Failed to load conversation', 'error')
  }
}

const createNewConversation = async () => {
  currentConversationId.value = null
  messages.value = []
  inputMessage.value = ''
  transcribedText.value = ''
  showMobileConversations.value = false // Close mobile drawer
  await loadConversations()
}

const deleteConversation = async (id: string) => {
  if (!confirm('Are you sure you want to delete this conversation?')) return
  
  try {
    await conversationsApi.deleteConversation(id)
    if (currentConversationId.value === id) {
      await createNewConversation()
    } else {
      await loadConversations()
    }
    showToast('Conversation deleted', 'success')
  } catch (error) {
    console.error('Failed to delete conversation:', error)
    showToast('Failed to delete conversation', 'error')
  }
}

const formatDate = (dateString: string | null | undefined): string => {
  if (dateString == null || dateString === '') return 'Just now'
  const date = new Date(dateString)
  if (Number.isNaN(date.getTime())) return 'Just now'
  const now = new Date()
  const diffMs = now.getTime() - date.getTime()
  const diffMins = Math.floor(diffMs / 60000)
  const diffHours = Math.floor(diffMs / 3600000)
  const diffDays = Math.floor(diffMs / 86400000)

  if (diffMins < 1) return 'Just now'
  if (diffMins < 60) return `${diffMins}m ago`
  if (diffHours < 24) return `${diffHours}h ago`
  if (diffDays < 7) return `${diffDays}d ago`
  return date.toLocaleDateString()
}

// Voice recording functions
const initSpeechRecognition = () => {
  if (typeof window === 'undefined') return
  
  const SpeechRecognition = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition
  if (!SpeechRecognition) {
    console.warn('Speech recognition not supported')
    return null
  }

  const recognition = new SpeechRecognition()
  recognition.continuous = true
  recognition.interimResults = true
  recognition.lang = 'en-US'

  recognition.onresult = (event: any) => {
    let interimTranscript = ''
    let finalTranscript = ''

    for (let i = event.resultIndex; i < event.results.length; i++) {
      const transcript = event.results[i][0].transcript
      if (event.results[i].isFinal) {
        finalTranscript += transcript + ' '
      } else {
        interimTranscript += transcript
      }
    }

    if (finalTranscript) {
      transcribedText.value += finalTranscript
    } else {
      // Show interim results in input
      inputMessage.value = transcribedText.value + interimTranscript
    }
  }

  recognition.onerror = (event: any) => {
    console.error('Speech recognition error:', event.error)
    if (event.error === 'no-speech') {
      showToast('No speech detected', 'warning')
    } else {
      showToast('Speech recognition error', 'error')
    }
    isRecording.value = false
  }

  recognition.onend = () => {
    isRecording.value = false
  }

  return recognition
}

const toggleVoiceRecording = async () => {
  if (!recognition.value) {
    recognition.value = initSpeechRecognition()
    if (!recognition.value) {
      showToast('Speech recognition not supported in your browser', 'error')
      return
    }
  }

  if (isRecording.value) {
    recognition.value.stop()
    isRecording.value = false
    // Move transcribed text to input
    if (transcribedText.value) {
      inputMessage.value = transcribedText.value
    }
  } else {
    // Check if Permissions-Policy allows microphone
    if (typeof navigator.permissions !== 'undefined') {
      try {
        const permissionStatus = await navigator.permissions.query({ name: 'microphone' as PermissionName })
        if (permissionStatus.state === 'denied') {
          showToast('Microphone access is blocked by browser permissions policy. Please check your browser settings.', 'error')
          return
        }
      } catch (e) {
        // Permissions API might not be fully supported, continue anyway
        console.warn('Permissions API not fully supported:', e)
      }
    }

    // Request microphone permission before starting
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      // Stop the stream immediately - we just needed permission
      stream.getTracks().forEach(track => track.stop())
      
      transcribedText.value = ''
      recognition.value.start()
      isRecording.value = true
    } catch (error: any) {
      console.error('Microphone permission error:', error)
      isRecording.value = false
      
      if (error.name === 'NotAllowedError' || error.name === 'PermissionDeniedError') {
        showToast('Microphone permission denied. Please allow microphone access in your browser settings and try again.', 'error')
      } else if (error.name === 'NotFoundError' || error.name === 'DevicesNotFoundError') {
        showToast('No microphone found. Please connect a microphone and try again.', 'error')
      } else if (error.name === 'NotReadableError' || error.name === 'TrackStartError') {
        showToast('Microphone is already in use by another application.', 'error')
      } else if (error.name === 'OverconstrainedError') {
        showToast('Microphone constraints could not be satisfied.', 'error')
      } else {
        showToast(`Failed to access microphone: ${error.message || 'Unknown error'}. Please check your browser settings.`, 'error')
      }
    }
  }
}

const speakText = (text: string) => {
  if (!speechSynthesis.value) {
    speechSynthesis.value = window.speechSynthesis
  }

  if (!speechSynthesis.value) {
    console.warn('Text-to-speech not supported')
    return
  }

  // Cancel any ongoing speech
  speechSynthesis.value.cancel()

  const utterance = new SpeechSynthesisUtterance(text)
  utterance.rate = 1.0
  utterance.pitch = 1.0
  utterance.volume = 1.0

  speechSynthesis.value.speak(utterance)
}

// Image URL helper - ensures proper data URL format
const getImageUrl = (url: string | undefined): string => {
  if (!url) return ''
  
  // If it's already a data URL or HTTP URL, return as-is
  if (url.startsWith('data:') || url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  
  // If it's base64 without data URL prefix, add it
  // Try to detect image type from the base64 data
  if (url.length > 0 && !url.includes('://')) {
    // Check if it looks like base64
    const base64Pattern = /^[A-Za-z0-9+/=]+$/
    if (base64Pattern.test(url.substring(0, 100))) {
      // Try to detect image type from first few characters
      let mimeType = 'image/png' // default
      if (url.startsWith('/9j/') || url.startsWith('/9j')) {
        mimeType = 'image/jpeg'
      } else if (url.startsWith('iVBORw0KGgo')) {
        mimeType = 'image/png'
      } else if (url.startsWith('R0lGOD')) {
        mimeType = 'image/gif'
      } else if (url.startsWith('UklGR')) {
        mimeType = 'image/webp'
      }
      return `data:${mimeType};base64,${url}`
    }
  }
  
  return url
}

// Image error handlers
const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  const partIndex = parseInt(img.getAttribute('data-part-index') || '0')
  imageErrors.value.add(partIndex)
  console.error('Failed to load image:', {
    src: img.src.substring(0, 100),
    partIndex,
    error: 'Image load failed'
  })
}

const handleImageLoad = (event: Event) => {
  const img = event.target as HTMLImageElement
  const partIndex = parseInt(img.getAttribute('data-part-index') || '0')
  imageErrors.value.delete(partIndex)
}

const handlePreviewImageError = (_event: Event, index: number) => {
  console.error('Failed to load preview image:', index)
  showToast('Failed to load image preview', 'error')
}

// Load conversations on mount
onMounted(async () => {
  scrollToBottom()
  try {
    const res = await providersApi.list()
    if (res.providers?.length) apiProviders.value = res.providers
  } catch {
    // use static model list
  }
  await loadConversations()
  
  // Initialize speech synthesis
  if (typeof window !== 'undefined' && 'speechSynthesis' in window) {
    speechSynthesis.value = window.speechSynthesis
  }
})

// Cleanup on unmount
onUnmounted(() => {
  // Restore body scroll
  document.body.style.overflow = ''
  
  // Clean up preview URLs (only revoke blob URLs)
  attachedImages.value.forEach(img => {
    if (img.preview.startsWith('blob:')) {
      URL.revokeObjectURL(img.preview)
    }
  })
  attachedAudios.value.forEach(audio => {
    if (audio.preview.startsWith('blob:')) {
      URL.revokeObjectURL(audio.preview)
    }
  })
  
  // Stop speech recognition if active
  if (recognition.value && isRecording.value) {
    recognition.value.stop()
  }
  
  // Stop any ongoing speech
  if (speechSynthesis.value) {
    speechSynthesis.value.cancel()
  }
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

