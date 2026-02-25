<template>
  <div class="flex h-[calc(100vh-8rem)] sm:h-[calc(100vh-12rem)] max-h-[800px] gap-4">
    <!-- Conversations Sidebar: flex + overflow so list scrolls and all items are clickable -->
    <div class="w-64 flex-shrink-0 hidden lg:flex lg:flex-col lg:min-h-0 self-stretch">
      <div class="rounded-lg border border-slate-700/50 bg-slate-800/60 flex flex-col flex-1 min-h-0 overflow-hidden">
        <div class="flex-shrink-0 p-4 border-b border-gray-200 dark:border-gray-700">
          <div class="flex items-center justify-between mb-2">
            <h2 class="text-lg font-semibold text-gray-900 dark:text-white">Conversations</h2>
            <button
              @click="createNewConversation"
              type="button"
              class="p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-600 dark:text-gray-400"
              title="New conversation"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
            </button>
          </div>
        </div>
        <div class="flex-1 min-h-0 overflow-y-auto overflow-x-hidden p-2 overscroll-contain">
          <div v-if="conversationsLoading" class="text-center py-4 text-gray-500 dark:text-gray-400">
            Loading...
          </div>
          <div v-else-if="(conversations || []).length === 0" class="text-center py-4 text-gray-500 dark:text-gray-400 text-sm">
            No conversations yet
          </div>
          <div v-else class="space-y-1">
            <button
              v-for="conv in (conversations || [])"
              :key="conv.id"
              type="button"
              @click="loadConversation(conv.id)"
              :class="[
                'w-full text-left px-3 py-2 rounded-lg text-sm transition-colors group',
                String(currentConversationId) === String(conv.id)
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
                  type="button"
                  class="ml-2 flex-shrink-0 p-1 rounded hover:bg-red-100 dark:hover:bg-red-900 opacity-0 group-hover:opacity-100 transition-opacity"
                  title="Delete conversation"
                >
                  <X class="w-4 h-4" />
                </button>
              </div>
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Main Chat Area: min-h-0 so messages area can shrink and scroll -->
    <div class="flex-1 flex flex-col min-w-0 min-h-0">
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
          <span v-if="currentConversationId && messages.length > 0" class="ml-1 text-gray-500 dark:text-gray-500">
            · {{ messages.length }} message{{ messages.length === 1 ? '' : 's' }}
          </span>
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
        <div v-if="isSearchCapableModel" class="flex items-center gap-2">
          <input
            id="web-search-toggle"
            v-model="webSearch"
            type="checkbox"
            class="rounded border-gray-300 dark:border-gray-600"
          />
          <label for="web-search-toggle" class="text-sm text-gray-700 dark:text-gray-300">
            Search the web – use real-time info when helpful
          </label>
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
        <div class="flex items-center gap-2">
          <input
            id="speak-responses"
            v-model="speakResponsesEnabled"
            type="checkbox"
            class="rounded border-gray-300 dark:border-gray-600"
          />
          <label for="speak-responses" class="text-sm text-gray-700 dark:text-gray-300">
            Speak responses (read assistant reply aloud)
          </label>
        </div>
      </div>
    </Card>

    <div class="flex-1 flex flex-col min-h-0 overflow-hidden mb-4 rounded-lg border border-slate-700/50 bg-slate-800/60">
      <div
        ref="messagesContainer"
        class="flex-1 min-h-0 overflow-y-auto overflow-x-hidden p-3 sm:p-6 space-y-3 sm:space-y-4 overscroll-contain"
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
          :ref="el => { if (el && index === messages.length - 1) lastMessageEl = el as HTMLElement }"
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
                  <span v-if="message.metadata.tokens" :title="'Input + output tokens for this reply'">
                    {{ message.metadata.tokens }} total tokens
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

                <!-- Accept / Reject suggested edit (IDE or dashboard) -->
                <div
                  v-if="message.role === 'assistant' && message.suggestedEdit"
                  class="mt-3 pt-3 border-t border-gray-200 dark:border-gray-600 flex items-center gap-2"
                >
                  <span class="text-xs text-gray-600 dark:text-gray-400">Suggested edit: {{ message.suggestedEdit.file }}</span>
                  <button
                    type="button"
                    @click="handleAcceptEdit(message.suggestedEdit!, index)"
                    class="px-2 py-1 text-xs font-medium rounded bg-green-600 text-white hover:bg-green-700"
                  >
                    Accept
                  </button>
                  <button
                    type="button"
                    @click="handleCopyEditAsJson(message.suggestedEdit!)"
                    class="px-2 py-1 text-xs font-medium rounded bg-gray-400 text-gray-100 hover:bg-gray-500 dark:bg-gray-600 dark:hover:bg-gray-500"
                  >
                    Copy as JSON
                  </button>
                  <button
                    type="button"
                    @click="handleRejectEdit(index)"
                    class="px-2 py-1 text-xs font-medium rounded bg-gray-500 text-gray-200 hover:bg-gray-600"
                  >
                    Reject
                  </button>
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
    </div>

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
            :disabled="loading || !hasContentToSend"
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
      </div>
    </Card>
    </div>

    <!-- Mobile Conversations Drawer -->
    <Teleport to="body">
      <div
        v-if="showMobileConversations"
        class="fixed inset-0 z-[9999] lg:hidden dark"
        @click.self="showMobileConversations = false"
      >
        <div class="absolute inset-0 bg-black bg-opacity-50" @click="showMobileConversations = false" @touchmove.prevent></div>

        <div class="absolute left-0 top-0 bottom-0 w-80 max-w-[85vw] bg-gray-900 shadow-xl flex flex-col overflow-hidden">
          <div class="p-4 border-b border-gray-700 flex items-center justify-between flex-shrink-0">
            <h2 class="text-lg font-semibold text-white">Conversations</h2>
            <button
              @click="showMobileConversations = false"
              class="p-1.5 rounded-lg hover:bg-gray-700 text-gray-400"
              aria-label="Close"
            >
              <X class="w-5 h-5" />
            </button>
          </div>
          <div class="flex-1 min-h-0 overflow-y-auto overflow-x-hidden p-2 overscroll-contain touch-pan-y">
            <div v-if="conversationsLoading" class="text-center py-4 text-gray-400">
              Loading...
            </div>
            <div v-else-if="(conversations || []).length === 0" class="text-center py-4 text-gray-400 text-sm">
              No conversations yet
            </div>
            <div v-else class="space-y-1">
              <button
                v-for="conv in (conversations || [])"
                :key="conv.id"
                type="button"
                @click="loadConversation(conv.id)"
                :class="[
                  'w-full text-left px-3 py-2 rounded-lg text-sm transition-colors group',
                  String(currentConversationId) === String(conv.id)
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
import { chatApi, type Message, type ChatResponse, type ContentPart, type StreamChunk, type SuggestedEdit } from '@/services/api/chat'
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
  suggestedEdit?: SuggestedEdit
}

const messages = ref<ChatMessage[]>([])
const inputMessage = ref('')
const attachedImages = ref<AttachedImage[]>([])
const attachedAudios = ref<AttachedAudio[]>([])
const imageInputRef = ref<HTMLInputElement | null>(null)
const audioInputRef = ref<HTMLInputElement | null>(null)
const loading = ref(false)
const streamAbortController = ref<AbortController | null>(null)
const CHAT_MODEL_STORAGE_KEY = 'uniroute-chat-selected-model'
const getStoredModel = (): string => {
  if (typeof window === 'undefined') return 'gpt-4'
  const saved = localStorage.getItem(CHAT_MODEL_STORAGE_KEY)
  return (saved && saved.trim()) ? saved.trim() : 'gpt-4'
}
const selectedModel = ref(getStoredModel())
watch(selectedModel, (model) => {
  if (typeof window !== 'undefined' && model) {
    localStorage.setItem(CHAT_MODEL_STORAGE_KEY, model)
  }
}, { immediate: false })
const temperature = ref(0.7)
const maxTokens = ref(1000)
const webSearch = ref(true)
const showSettings = ref(false)
const messagesContainer = ref<HTMLElement | null>(null)
const lastMessageEl = ref<HTMLElement | null>(null)

const conversations = ref<Conversation[]>([])
const conversationsLoading = ref(false)
const currentConversationId = ref<string | null>(null)

const isRecording = ref(false)
const mediaRecorderRef = ref<MediaRecorder | null>(null)
const recordingChunksRef = ref<Blob[]>([])
const recordingStreamRef = ref<MediaStream | null>(null)
const speechSynthesis = ref<SpeechSynthesis | null>(null)
const speakResponsesEnabled = ref(false)
const imageErrors = ref(new Set<number>())
const showMobileConversations = ref(false)

const isSearchCapableModel = computed(() => {
  const m = selectedModel.value
  return /^gemini-/i.test(m) || /^gpt-|^o\d/i.test(m) || /^claude-/i.test(m)
})
const hasContentToSend = computed(() => {
  return !!(
    (inputMessage.value && inputMessage.value.trim()) ||
    attachedImages.value.length > 0 ||
    attachedAudios.value.length > 0
  )
})

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

const DEFAULT_OLLAMA_MODELS = [
  { value: 'llama3.2:3b', label: 'Llama 3.2 3B' },
  { value: 'llama3.2:latest', label: 'Llama 3.2 (latest)' },
  { value: 'llava:latest', label: 'Llava (latest) — Vision' },
  { value: 'mistral:latest', label: 'Mistral (latest)' }
]

const OLLAMA_VISION_MODELS: Record<string, string> = {
  'llava:latest': 'Llava (latest) — Vision'
}

const DEFAULT_VLLM_MODELS = [
  { value: 'TinyLlama/TinyLlama-1.1B-Chat-v1.0', label: 'TinyLlama 1.1B Chat' }
]

const staticModelGroups = [
  { label: 'OpenAI', options: [
    { value: 'gpt-5.2', label: 'GPT-5.2 (Latest)' },
    { value: 'gpt-5.2-pro', label: 'GPT-5.2 Pro' },
    { value: 'gpt-5.1', label: 'GPT-5.1' },
    { value: 'gpt-5', label: 'GPT-5' },
    { value: 'gpt-5-mini', label: 'GPT-5 Mini' },
    { value: 'gpt-5-nano', label: 'GPT-5 Nano' },
    { value: 'gpt-4.1', label: 'GPT-4.1' },
    { value: 'gpt-4.1-mini', label: 'GPT-4.1 Mini' },
    { value: 'gpt-4.1-nano', label: 'GPT-4.1 Nano' },
    { value: 'o3', label: 'o3 (Reasoning)' },
    { value: 'o3-mini', label: 'o3 Mini' },
    { value: 'o4-mini', label: 'o4 Mini' },
    { value: 'o1', label: 'o1 (Reasoning)' },
    { value: 'gpt-4o', label: 'GPT-4o' },
    { value: 'gpt-4o-mini', label: 'GPT-4o Mini' },
    { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
    { value: 'gpt-4', label: 'GPT-4' },
    { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' }
  ]},
  { label: 'Anthropic (Claude)', options: [
    { value: 'claude-opus-4-6', label: 'Claude Opus 4.6 (Latest)' },
    { value: 'claude-sonnet-4-6', label: 'Claude Sonnet 4.6 — Extended thinking' },
    { value: 'claude-haiku-4-5-20251001', label: 'Claude Haiku 4.5' },
    { value: 'claude-sonnet-4-5-20250929', label: 'Claude Sonnet 4.5' },
    { value: 'claude-opus-4-5-20251101', label: 'Claude Opus 4.5' },
    { value: 'claude-opus-4-1-20250805', label: 'Claude Opus 4.1' },
    { value: 'claude-sonnet-4-20250514', label: 'Claude Sonnet 4' },
    { value: 'claude-opus-4-20250514', label: 'Claude Opus 4' },
    { value: 'claude-3-5-sonnet-20241022', label: 'Claude 3.5 Sonnet' },
    { value: 'claude-3-5-haiku-20241022', label: 'Claude 3.5 Haiku' },
    { value: 'claude-3-opus-20240229', label: 'Claude 3 Opus' },
    { value: 'claude-3-sonnet-20240229', label: 'Claude 3 Sonnet' },
    { value: 'claude-3-haiku-20240307', label: 'Claude 3 Haiku' }
  ]},
  { label: 'Google (Gemini)', options: [
    { value: 'gemini-3-pro-preview', label: 'Gemini 3 Pro Preview' },
    { value: 'gemini-3-flash-preview', label: 'Gemini 3 Flash Preview' },
    { value: 'gemini-2.5-pro', label: 'Gemini 2.5 Pro' },
    { value: 'gemini-2.5-flash', label: 'Gemini 2.5 Flash' },
    { value: 'gemini-2.5-flash-lite', label: 'Gemini 2.5 Flash Lite' },
    { value: 'gemini-2.0-flash-exp', label: 'Gemini 2.0 Flash (Experimental)' },
    { value: 'gemini-1.5-pro-latest', label: 'Gemini 1.5 Pro (Latest)' },
    { value: 'gemini-1.5-pro', label: 'Gemini 1.5 Pro' },
    { value: 'gemini-1.5-flash-8b', label: 'Gemini 1.5 Flash 8B' },
    { value: 'gemini-1.5-flash', label: 'Gemini 1.5 Flash' },
    { value: 'gemini-pro', label: 'Gemini Pro' },
    { value: 'gemini-pro-vision', label: 'Gemini Pro Vision' }
  ]},
  { label: 'Ollama', options: DEFAULT_OLLAMA_MODELS },
  { label: 'vLLM', options: DEFAULT_VLLM_MODELS }
]

const apiProviders = ref<ProviderInfo[]>([])
const modelGroups = computed(() => {
  if (apiProviders.value.length === 0) return staticModelGroups
  return apiProviders.value.map((p) => {
    const isOllama = p.name === 'local' || p.name === 'ollama'
    let options = p.models.map((m) => ({
      value: m,
      label: isOllama && OLLAMA_VISION_MODELS[m] ? OLLAMA_VISION_MODELS[m] : m
    }))
    if (options.length === 0 && isOllama) {
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

watch(messages, () => {
  scrollToBottomAfterUpdate()
}, { deep: true })

watch(showMobileConversations, (isOpen) => {
  if (isOpen) {
    document.body.style.overflow = 'hidden'
  } else {
    document.body.style.overflow = ''
  }
})

const scrollToBottom = () => {
  const el = lastMessageEl.value
  if (el) {
    el.scrollIntoView({ behavior: 'auto', block: 'end', inline: 'nearest' })
  } else if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const scrollToBottomAfterUpdate = () => {
  nextTick(() => {
    requestAnimationFrame(() => scrollToBottom())
  })
}

function handleAcceptEdit(edit: SuggestedEdit, messageIndex: number) {
  const payload = { file: edit.file, range: edit.range, newText: edit.new_text }
  try {
    if (typeof window !== 'undefined' && window.parent !== window) {
      window.parent.postMessage({ type: 'applyEdit', edit: payload }, '*')
      showToast('Edit sent to IDE. If you\'re in the UniRoute extension, it will be applied.', 'success')
    } else {
      const json = editJsonForClipboard(edit)
      navigator.clipboard.writeText(json).then(() => {
        showToast('Edit JSON copied. In JetBrains: UniRoute → Apply suggested edit from clipboard.', 'success')
      }).catch(() => {
        showToast('Edit: ' + edit.file + ' – open in IDE or use Copy as JSON.', 'info')
      })
    }
  } catch (_) {
    showToast('Could not send edit to IDE. Copy from the message or use the extension.', 'info')
  }
  messages.value[messageIndex] = { ...messages.value[messageIndex], suggestedEdit: undefined }
}

function handleRejectEdit(messageIndex: number) {
  messages.value[messageIndex] = { ...messages.value[messageIndex], suggestedEdit: undefined }
  showToast('Edit rejected', 'info')
}

function editJsonForClipboard(edit: SuggestedEdit): string {
  return JSON.stringify({
    file: edit.file,
    range: edit.range,
    new_text: edit.new_text
  })
}

function handleCopyEditAsJson(edit: SuggestedEdit) {
  const json = editJsonForClipboard(edit)
  navigator.clipboard.writeText(json).then(() => {
    showToast('Edit JSON copied. In JetBrains: UniRoute → Apply suggested edit from clipboard.', 'success')
  }).catch(() => {
    showToast('Failed to copy', 'error')
  })
}

const handleSend = async () => {
  const textToSend = inputMessage.value.trim()
  if ((!textToSend && attachedImages.value.length === 0 && attachedAudios.value.length === 0) || loading.value) return

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
      showToast('Failed to create conversation. Try again.', 'error')
      return
    }
  }

  let messageContent: string | ContentPart[] = textToSend

  if (attachedImages.value.length > 0 || attachedAudios.value.length > 0) {
    const parts: ContentPart[] = []
    for (const image of attachedImages.value) {
      parts.push({
        type: 'image_url',
        image_url: { url: image.dataUrl }
      })
    }
    for (const audio of attachedAudios.value) {
      parts.push({
        type: 'audio_url',
        audio_url: { url: audio.dataUrl }
      })
    }
    if (textToSend) {
      parts.push({ type: 'text', text: textToSend })
    }
    messageContent = parts
  }

  const userMessage: ChatMessage = {
    role: 'user',
    content: messageContent
  }

  messages.value.push(userMessage)
  inputMessage.value = ''
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

  if (streamAbortController.value) {
    streamAbortController.value.abort()
  }
  const controller = new AbortController()
  streamAbortController.value = controller
  let streamAborted = false

  try {
    const chatMessages: Message[] = messages.value
      .filter(m => m.role !== 'system' || messages.value.indexOf(m) === 0)
      .map(m => ({ 
        role: m.role, 
        content: m.content
      }))

    const chatRequestData: any = {
      model: selectedModel.value,
      messages: chatMessages,
      temperature: temperature.value,
      max_tokens: maxTokens.value
    }
    if (webSearch.value) {
      chatRequestData.web_search = true
    }
    if (currentConversationId.value) {
      chatRequestData.conversation_id = currentConversationId.value
    }

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
          if (chunk.content) {
            assistantMessage.content += chunk.content
            messages.value[assistantMessageIndex] = { ...assistantMessage }
            scrollToBottomAfterUpdate()
          }

          if (chunk.suggested_edit) {
            assistantMessage.suggestedEdit = chunk.suggested_edit
            messages.value[assistantMessageIndex] = { ...assistantMessage }
            scrollToBottomAfterUpdate()
          }

          if (chunk.done && chunk.usage) {
            assistantMessage.metadata = {
              tokens: chunk.usage.total_tokens,
              cost: 0,
              provider: chunk.provider || 'unknown',
              latency: 0
            }
            messages.value[assistantMessageIndex] = { ...assistantMessage }
            scrollToBottomAfterUpdate()
          }

          if (chunk.done && typeof assistantMessage.content === 'string' && assistantMessage.content.trim() === '') {
            const err = chunk.error || 'No content from model. Try again or use a shorter conversation.'
            throw new Error(err)
          }

          if (chunk.error) {
            throw new Error(chunk.error)
          }
        },
        (error) => {
          messages.value.splice(assistantMessageIndex, 1)
          showToast('Failed to get response: ' + error.message, 'error')
        },
        (usage) => {
          if (usage) {
            assistantMessage.metadata = {
              ...assistantMessage.metadata,
              tokens: usage.total_tokens
            }
            messages.value[assistantMessageIndex] = { ...assistantMessage }
          }
        },
        true,
        controller.signal
      )
    } catch (error: any) {
      if (error?.name === 'AbortError') {
        messages.value.splice(assistantMessageIndex, 1)
        streamAborted = true
        return
      }
      if (messages.value[assistantMessageIndex]?.content === '') {
        messages.value.splice(assistantMessageIndex, 1)
      }
      showToast('Failed to get response: ' + (error.message || 'Unknown error'), 'error')
    }

    if (streamAborted) return

    const contentStr = typeof assistantMessage.content === 'string' ? assistantMessage.content : ''
    if (contentStr.trim() === '') {
      assistantMessage.content = 'No response from the model. Check your API key and model name, or try again.'
      messages.value[assistantMessageIndex] = { ...assistantMessage }
      scrollToBottomAfterUpdate()
      showToast('No response from the model', 'error')
    }

    if (speakResponsesEnabled.value && speechSynthesis.value && assistantMessage.content) {
      const textToSpeak = typeof assistantMessage.content === 'string'
        ? assistantMessage.content
        : assistantMessage.content.filter((p): p is ContentPart & { text: string } => p.type === 'text' && !!p.text).map(p => p.text).join(' ')
      if (textToSpeak) speakText(textToSpeak)
    }
  } catch (error: any) {
    console.error('Chat error:', error)
    const errorMessage = error.response?.data?.error || error.message || 'Failed to get response'
    showToast(errorMessage, 'error')
    const errorChatMessage: ChatMessage = {
      role: 'assistant',
      content: `Error: ${errorMessage}`,
      metadata: {
        provider: 'error'
      }
    }
    messages.value.push(errorChatMessage)
  } finally {
    if (!streamAborted) {
      loading.value = false
    }
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

    if (file.size > 10 * 1024 * 1024) {
      showToast('Image size must be less than 10MB', 'error')
      return
    }

    const reader = new FileReader()
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string
      attachedImages.value.push({ file, dataUrl, preview: dataUrl })
    }
    reader.readAsDataURL(file)
  })
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

    if (file.size > 25 * 1024 * 1024) {
      showToast('Audio file size must be less than 25MB', 'error')
      return
    }

    const reader = new FileReader()
    reader.onload = (e) => {
      const dataUrl = e.target?.result as string
      const audio = new Audio()
      audio.src = dataUrl
      audio.addEventListener('loadedmetadata', () => {
        attachedAudios.value.push({
          file,
          dataUrl,
          preview: dataUrl,
          duration: audio.duration
        })
      }, { once: true })
      setTimeout(() => {
        if (!attachedAudios.value.find(a => a.file === file)) {
          attachedAudios.value.push({ file, dataUrl, preview: dataUrl })
        }
      }, 1000)
    }
    reader.readAsDataURL(file)
  })
  target.value = ''
}

const formatDuration = (seconds: number): string => {
  if (!seconds || isNaN(seconds)) return ''
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}

const removeImage = (index: number) => {
  const preview = attachedImages.value[index].preview
  if (preview.startsWith('blob:')) {
    URL.revokeObjectURL(preview)
  }
  attachedImages.value.splice(index, 1)
}

const removeAudio = (index: number) => {
  const preview = attachedAudios.value[index].preview
  if (preview.startsWith('blob:')) {
    URL.revokeObjectURL(preview)
  }
  attachedAudios.value.splice(index, 1)
}

const sendMessage = (message: string) => {
  inputMessage.value = message
  handleSend()
}

const clearChat = () => {
  if (confirm('Are you sure you want to clear the chat history?')) {
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

const loadConversations = async () => {
  try {
    conversationsLoading.value = true
    const list = await conversationsApi.listConversations(50, 0)
    conversations.value = Array.isArray(list) ? list : []
  } catch (error) {
    console.error('Failed to load conversations:', error)
    showToast('Failed to load conversations', 'error')
    conversations.value = []
  } finally {
    conversationsLoading.value = false
  }
}

const loadConversation = async (id: string) => {
  const conversationId = id != null ? (typeof id === 'string' ? id : String(id)) : ''
  if (!conversationId) {
    showToast('Cannot load: conversation ID missing', 'error')
    return
  }
  try {
    const data = await conversationsApi.getConversation(conversationId)
    if (!data?.conversation) {
      showToast('Invalid response: conversation missing', 'error')
      currentConversationId.value = null
      messages.value = []
      return
    }
    currentConversationId.value = conversationId
    const convModel = data.conversation?.model ?? (data.conversation as { Model?: string })?.Model
    selectedModel.value = (typeof convModel === 'string' && convModel.trim()) ? convModel.trim() : 'gpt-4'
    const rawMessages = data.messages != null && Array.isArray(data.messages) ? data.messages : []
    messages.value = rawMessages.map((msg: any) => ({
      role: msg.role ?? 'user',
      content: msg.content ?? '',
      metadata: msg.metadata || undefined
    }))
    showToast('Conversation loaded', 'success')
    showMobileConversations.value = false
    scrollToBottomAfterUpdate()
  } catch (error: any) {
    console.error('Failed to load conversation:', error)
    const msg = error.response?.data?.error ?? error.response?.data?.message ?? error.message ?? 'Failed to load conversation'
    showToast(typeof msg === 'string' ? msg : 'Failed to load conversation', 'error')
    currentConversationId.value = null
    messages.value = []
  }
}

const createNewConversation = async () => {
  currentConversationId.value = null
  messages.value = []
  inputMessage.value = ''
  showMobileConversations.value = false
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

const toggleVoiceRecording = async () => {
  if (isRecording.value) {
    if (mediaRecorderRef.value && mediaRecorderRef.value.state !== 'inactive') {
      mediaRecorderRef.value.stop()
    }
    recordingStreamRef.value?.getTracks().forEach(track => track.stop())
    recordingStreamRef.value = null
    mediaRecorderRef.value = null
    isRecording.value = false
    return
  }

  if (typeof navigator.permissions !== 'undefined') {
    try {
      const status = await navigator.permissions.query({ name: 'microphone' as PermissionName })
      if (status.state === 'denied') {
        showToast('Microphone access is blocked. Check browser settings.', 'error')
        return
      }
    } catch (_) {}
  }

  try {
    const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
    recordingStreamRef.value = stream
    const mime = MediaRecorder.isTypeSupported('audio/webm;codecs=opus') ? 'audio/webm;codecs=opus' : 'audio/webm'
    const recorder = new MediaRecorder(stream)
    recordingChunksRef.value = []

    recorder.ondataavailable = (e) => {
      if (e.data.size > 0) recordingChunksRef.value.push(e.data)
    }
    recorder.onstop = () => {
      const blob = new Blob(recordingChunksRef.value, { type: mime })
      const reader = new FileReader()
      reader.onload = () => {
        const dataUrl = reader.result as string
        const preview = URL.createObjectURL(blob)
        const file = new File([blob], 'recording.webm', { type: blob.type })
        attachedAudios.value.push({
          file,
          dataUrl: file.type.startsWith('audio/') ? dataUrl : dataUrl,
          preview,
          duration: undefined
        })
      }
      reader.readAsDataURL(blob)
    }
    recorder.start(200)
    mediaRecorderRef.value = recorder
    isRecording.value = true
  } catch (err: any) {
    isRecording.value = false
    if (err.name === 'NotAllowedError' || err.name === 'PermissionDeniedError') {
      showToast('Microphone permission denied.', 'error')
    } else if (err.name === 'NotFoundError') {
      showToast('No microphone found.', 'error')
    } else {
      showToast(err.message || 'Failed to access microphone', 'error')
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

  speechSynthesis.value.cancel()

  const utterance = new SpeechSynthesisUtterance(text)
  utterance.rate = 1.0
  utterance.pitch = 1.0
  utterance.volume = 1.0

  speechSynthesis.value.speak(utterance)
}

const getImageUrl = (url: string | undefined): string => {
  if (!url) return ''
  if (url.startsWith('data:') || url.startsWith('http://') || url.startsWith('https://')) {
    return url
  }
  if (url.length > 0 && !url.includes('://')) {
    const base64Pattern = /^[A-Za-z0-9+/=]+$/
    if (base64Pattern.test(url.substring(0, 100))) {
      let mimeType = 'image/png'
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

onMounted(async () => {
  scrollToBottom()
  try {
    const res = await providersApi.list()
    if (res.providers?.length) apiProviders.value = res.providers
  } catch {}
  await loadConversations()
  if (typeof window !== 'undefined' && 'speechSynthesis' in window) {
    speechSynthesis.value = window.speechSynthesis
  }
})

onUnmounted(() => {
  document.body.style.overflow = ''
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
  
  if (mediaRecorderRef.value && mediaRecorderRef.value.state !== 'inactive') {
    mediaRecorderRef.value.stop()
  }
  recordingStreamRef.value?.getTracks().forEach(track => track.stop())
  recordingStreamRef.value = null
  if (speechSynthesis.value) {
    speechSynthesis.value.cancel()
  }
})
</script>

<style scoped>
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

