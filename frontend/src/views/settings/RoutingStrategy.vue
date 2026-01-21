<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold text-white">Routing Strategy</h2>
      <p class="text-slate-400 mt-1">
        Configure how your requests are routed to AI providers
      </p>
    </div>

    <!-- Lock Warning Banner -->
    <div
      v-if="strategy?.is_locked"
      class="bg-amber-500/10 border-2 border-amber-500/30 rounded-lg p-4 shadow-lg"
    >
      <div class="flex items-start gap-3">
        <svg
          class="w-6 h-6 text-amber-400 mt-0.5 flex-shrink-0"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
          />
        </svg>
        <div class="flex-1">
          <h3 class="font-bold text-amber-400 mb-1 text-lg">⚠️ Strategy Locked by Administrator</h3>
          <p class="text-sm text-amber-300/90 mb-2">
            The routing strategy is currently locked. You <strong class="text-amber-200">cannot change</strong> your routing preference and must use the default strategy set by the administrator.
          </p>
          <p class="text-xs text-amber-300/70">
            Current default strategy: <span class="font-semibold text-amber-200">{{ getStrategyName(strategy?.default_strategy || 'model') }}</span>
          </p>
        </div>
      </div>
    </div>

    <!-- Current Strategy -->
    <Card>
      <div class="space-y-4">
        <h3 class="text-lg font-semibold text-white">Current Strategy</h3>
        
        <div v-if="loading" class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="text-slate-400 mt-2">Loading strategy...</p>
        </div>

        <div v-else-if="strategy" class="space-y-4">
          <!-- Effective Strategy -->
          <div class="bg-slate-800/50 rounded-lg p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-slate-400">Active Strategy</p>
                <p class="text-2xl font-bold text-white mt-1">{{ getStrategyName(strategy.strategy) }}</p>
                <p v-if="strategy.user_strategy" class="text-xs text-blue-400 mt-1">
                  Using your custom preference
                </p>
                <p v-else class="text-xs text-slate-500 mt-1">
                  Using default strategy
                </p>
              </div>
              <div
                class="px-3 py-1 rounded-full text-sm font-medium bg-blue-500/20 text-blue-400"
              >
                Active
              </div>
            </div>
          </div>

          <!-- Default Strategy Info -->
          <div v-if="strategy.default_strategy" class="bg-slate-800/30 rounded-lg p-3 border border-slate-700">
            <p class="text-xs text-slate-400 mb-1">Default Strategy</p>
            <p class="text-sm font-medium text-slate-300">{{ getStrategyName(strategy.default_strategy) }}</p>
          </div>

          <!-- Strategy Selection (only if not locked) -->
          <div v-if="!strategy.is_locked" class="space-y-3">
            <label class="block text-sm font-medium text-slate-300">
              Select Your Routing Strategy
            </label>
            <div class="grid gap-3 md:grid-cols-2">
              <div
                v-for="availableStrategy in strategy.available_strategies"
                :key="availableStrategy"
                @click="selectedStrategy = availableStrategy"
                class="p-4 rounded-lg border-2 cursor-pointer transition-all"
                :class="
                  selectedStrategy === availableStrategy
                    ? 'border-blue-500 bg-blue-500/10'
                    : 'border-slate-700 bg-slate-800/50 hover:border-slate-600'
                "
              >
                <div class="flex items-start justify-between">
                  <div class="flex-1">
                    <h4 class="font-semibold text-white">{{ getStrategyName(availableStrategy) }}</h4>
                    <p class="text-sm text-slate-400 mt-1">{{ getStrategyDescription(availableStrategy) }}</p>
                  </div>
                  <div
                    v-if="availableStrategy === strategy.strategy"
                    class="ml-3 px-2 py-1 rounded-full text-xs font-medium bg-green-500/20 text-green-400"
                  >
                    Active
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Locked Message (when strategy is locked) -->
          <div v-else class="bg-slate-800/30 rounded-lg p-4 border border-amber-500/20">
            <div class="flex items-start gap-3">
              <svg
                class="w-5 h-5 text-amber-400 mt-0.5 flex-shrink-0"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"
                />
              </svg>
              <div class="flex-1">
                <p class="text-sm font-medium text-amber-400 mb-1">
                  Strategy Selection Disabled
                </p>
                <p class="text-xs text-slate-400">
                  You cannot change your routing strategy because it is locked by an administrator. 
                  You are currently using the default strategy: <span class="font-semibold text-slate-300">{{ getStrategyName(strategy.default_strategy) }}</span>
                </p>
              </div>
            </div>
          </div>

          <!-- Use Default Option -->
          <div v-if="!strategy.is_locked && strategy.user_strategy" class="mt-4">
            <Button
              @click="handleUseDefault"
              variant="outline"
              :disabled="saving"
              class="w-full"
            >
              <span>Use Default Strategy</span>
            </Button>
          </div>

          <Button
            v-if="!strategy.is_locked"
            @click="handleUpdateStrategy"
            :disabled="saving || selectedStrategy === strategy.strategy || !selectedStrategy"
            class="w-full"
          >
            <span v-if="saving">Updating...</span>
            <span v-else>Update Strategy</span>
          </Button>

          <div v-if="updateResult" class="mt-4 p-4 rounded-lg" :class="updateResult.success ? 'bg-green-500/10 border border-green-500/20' : 'bg-red-500/10 border border-red-500/20'">
            <p :class="updateResult.success ? 'text-green-400' : 'text-red-400'">
              {{ updateResult.message }}
            </p>
          </div>
        </div>
      </div>
    </Card>

    <!-- Custom Rules Info Banner -->
    <div
      v-if="strategy?.strategy === 'custom' || selectedStrategy === 'custom'"
      class="bg-blue-500/10 border-2 border-blue-500/30 rounded-lg p-4 shadow-lg"
    >
      <div class="flex items-start gap-3">
        <svg
          class="w-6 h-6 text-blue-400 mt-0.5 flex-shrink-0"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
          />
        </svg>
        <div class="flex-1">
          <h3 class="font-bold text-blue-400 mb-1 text-lg">ℹ️ Custom Routing Rules</h3>
          <p class="text-sm text-blue-300/90 mb-2">
            When using the <strong class="text-blue-200">Custom</strong> strategy, you can define your own routing rules to control how requests are routed to providers.
          </p>
          <p class="text-xs text-blue-300/70">
            <strong>Note:</strong> If you haven't defined any custom rules yet, the system will fall back to the model-based strategy. 
            <router-link to="/dashboard/settings/custom-rules" class="text-blue-300 underline hover:text-blue-200">
              Configure custom rules here
            </router-link>.
          </p>
        </div>
      </div>
    </div>

    <!-- Strategy Information -->
    <Card>
      <div class="space-y-4">
        <h3 class="text-lg font-semibold text-white">Strategy Details</h3>
        <div class="space-y-3 text-sm text-slate-400">
          <div>
            <h4 class="font-medium text-white mb-1">Model-Based</h4>
            <p>Selects provider based on the requested model. If a provider supports the model, it's used.</p>
          </div>
          <div>
            <h4 class="font-medium text-white mb-1">Cost-Based</h4>
            <p>Selects the cheapest provider for the requested model based on pricing calculations.</p>
          </div>
          <div>
            <h4 class="font-medium text-white mb-1">Latency-Based</h4>
            <p>Selects the fastest provider based on historical response time data.</p>
          </div>
          <div>
            <h4 class="font-medium text-white mb-1">Load-Balanced</h4>
            <p>Distributes requests evenly across all available providers using round-robin.</p>
          </div>
          <div>
            <h4 class="font-medium text-white mb-1">Custom</h4>
            <p>Uses custom routing rules. You can define your own custom routing rules in the Custom Rules section.</p>
          </div>
        </div>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { routingApi, type UserRoutingStrategy } from '@/services/api/routing'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import { ErrorHandler } from '@/utils/errorHandler'
import { useToast } from '@/composables/useToast'

const { showToast } = useToast()

const loading = ref(false)
const saving = ref(false)
const strategy = ref<UserRoutingStrategy | null>(null)
const selectedStrategy = ref<string>('')
const updateResult = ref<{ success: boolean; message: string } | null>(null)

const getStrategyName = (strategy: string): string => {
  const names: Record<string, string> = {
    model: 'Model-Based',
    cost: 'Cost-Based',
    latency: 'Latency-Based',
    balanced: 'Load-Balanced',
    custom: 'Custom'
  }
  return names[strategy] || strategy
}

const getStrategyDescription = (strategy: string): string => {
  const descriptions: Record<string, string> = {
    model: 'Route by model compatibility',
    cost: 'Route by lowest cost',
    latency: 'Route by fastest response',
    balanced: 'Route by round-robin',
    custom: 'Route by custom rules'
  }
  return descriptions[strategy] || ''
}

const loadStrategy = async () => {
  loading.value = true
  try {
    strategy.value = await routingApi.getUserStrategy()
    // Set selected strategy to current effective strategy
    selectedStrategy.value = strategy.value.strategy
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    showToast(`Failed to load strategy: ${appError.message}`, 'error')
    ErrorHandler.logError(err, 'UserRoutingStrategy')
  } finally {
    loading.value = false
  }
}

const handleUpdateStrategy = async () => {
  if (!selectedStrategy.value) return

  // Check if locked before attempting update
  if (strategy.value?.is_locked) {
    showToast('Cannot update strategy: The routing strategy is locked by an administrator. You cannot override the default strategy.', 'error')
    updateResult.value = {
      success: false,
      message: 'The routing strategy is locked by an administrator. You cannot override the default strategy.'
    }
    return
  }

  saving.value = true
  updateResult.value = null

  try {
    const response = await routingApi.setUserStrategy({
      strategy: selectedStrategy.value
    })

    updateResult.value = {
      success: true,
      message: response.message || 'Routing strategy updated successfully'
    }

    showToast('Routing strategy updated successfully', 'success')

    // Reload to get updated strategy
    await loadStrategy()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    
    // Check if error is due to locked strategy
    const errorMessage = appError.message || 'Failed to update strategy'
    const isLockedError = errorMessage.toLowerCase().includes('locked') || 
                         errorMessage.toLowerCase().includes('cannot override')
    
    updateResult.value = {
      success: false,
      message: errorMessage
    }
    
    // Show prominent error for locked strategy
    if (isLockedError) {
      showToast('⚠️ Strategy is locked: ' + errorMessage, 'error')
    } else {
      showToast(errorMessage, 'error')
    }
    
    ErrorHandler.logError(err, 'UserRoutingStrategy')
  } finally {
    saving.value = false
  }
}

const handleUseDefault = async () => {
  if (!strategy.value) return

  // Check if locked before attempting to clear
  if (strategy.value.is_locked) {
    showToast('Cannot reset strategy: The routing strategy is locked by an administrator. You cannot override the default strategy.', 'error')
    updateResult.value = {
      success: false,
      message: 'The routing strategy is locked by an administrator. You cannot override the default strategy.'
    }
    return
  }

  saving.value = true
  updateResult.value = null

  try {
    // Clear user preference to use default
    const response = await routingApi.clearUserStrategy()

    updateResult.value = {
      success: true,
      message: response.message || 'Now using default strategy'
    }

    showToast('Now using default strategy', 'success')

    // Reload to get updated strategy
    await loadStrategy()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    
    // Check if error is due to locked strategy
    const errorMessage = appError.message || 'Failed to reset to default'
    const isLockedError = errorMessage.toLowerCase().includes('locked') || 
                         errorMessage.toLowerCase().includes('cannot override')
    
    updateResult.value = {
      success: false,
      message: errorMessage
    }
    
    // Show prominent error for locked strategy
    if (isLockedError) {
      showToast('⚠️ Strategy is locked: ' + errorMessage, 'error')
    } else {
      showToast(errorMessage, 'error')
    }
    
    ErrorHandler.logError(err, 'UserRoutingStrategy')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadStrategy()
})
</script>

