<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-3xl font-bold text-white">Routing Strategy</h1>
      <p class="text-slate-400 mt-1">
        Configure how UniRoute selects AI providers for requests
      </p>
    </div>

    <!-- Current Strategy -->
    <Card>
      <div class="space-y-4">
        <h2 class="text-xl font-semibold text-white">Current Strategy</h2>
        
        <div v-if="loading" class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="text-slate-400 mt-2">Loading strategy...</p>
        </div>

        <div v-else-if="strategy" class="space-y-4">
          <div class="bg-slate-800/50 rounded-lg p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-slate-400">Active Strategy</p>
                <p class="text-2xl font-bold text-white mt-1">{{ getStrategyName(strategy.strategy) }}</p>
              </div>
              <div
                class="px-3 py-1 rounded-full text-sm font-medium bg-blue-500/20 text-blue-400"
              >
                Active
              </div>
            </div>
          </div>

          <!-- Lock Status -->
          <div 
            class="rounded-lg p-4 border-2"
            :class="strategy.is_locked 
              ? 'bg-amber-500/10 border-amber-500/30' 
              : 'bg-green-500/10 border-green-500/30'"
          >
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-slate-400 mb-1">Strategy Lock Status</p>
                <p class="text-lg font-semibold" :class="strategy.is_locked ? 'text-amber-400' : 'text-green-400'">
                  {{ strategy.is_locked ? '🔒 Locked' : '🔓 Unlocked' }}
                </p>
                <p class="text-xs mt-1" :class="strategy.is_locked ? 'text-amber-300/80' : 'text-green-300/80'">
                  {{ strategy.is_locked 
                    ? 'Users cannot override the default strategy' 
                    : 'Users can set their own routing strategy' }}
                </p>
              </div>
              <Button
                @click="handleToggleLock"
                :disabled="locking"
                variant="primary"
                :class="[
                  'ml-4 min-w-[140px]',
                  strategy.is_locked 
                    ? 'bg-amber-600 hover:bg-amber-700 border-amber-500' 
                    : 'bg-blue-600 hover:bg-blue-700 border-blue-500'
                ]"
              >
                <span v-if="locking">
                  {{ strategy.is_locked ? 'Unlocking...' : 'Locking...' }}
                </span>
                <span v-else>
                  {{ strategy.is_locked ? '🔓 Unlock Strategy' : '🔒 Lock Strategy' }}
                </span>
              </Button>
            </div>
          </div>

          <!-- Strategy Selection -->
          <div class="space-y-3">
            <label class="block text-sm font-medium text-slate-300">
              Select Routing Strategy
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
                    <h3 class="font-semibold text-white">{{ getStrategyName(availableStrategy) }}</h3>
                    <p class="text-sm text-slate-400 mt-1">{{ getStrategyDescription(availableStrategy) }}</p>
                  </div>
                  <div
                    v-if="availableStrategy === strategy.strategy"
                    class="ml-3 px-2 py-1 rounded-full text-xs font-medium bg-green-500/20 text-green-400"
                  >
                    Current
                  </div>
                </div>
              </div>
            </div>
          </div>

          <Button
            @click="handleUpdateStrategy"
            :disabled="saving || selectedStrategy === strategy.strategy"
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

    <!-- Strategy Information -->
    <Card>
      <div class="space-y-4">
        <h2 class="text-xl font-semibold text-white">Strategy Details</h2>
        <div class="space-y-3 text-sm text-slate-400">
          <div>
            <h3 class="font-medium text-white mb-1">Model-Based (default)</h3>
            <p>Selects provider based on the requested model. If a provider supports the model, it's used.</p>
          </div>
          <div>
            <h3 class="font-medium text-white mb-1">Cost-Based</h3>
            <p>Selects the cheapest provider for the requested model based on pricing calculations.</p>
          </div>
          <div>
            <h3 class="font-medium text-white mb-1">Latency-Based</h3>
            <p>Selects the fastest provider based on historical response time data.</p>
          </div>
          <div>
            <h3 class="font-medium text-white mb-1">Load-Balanced</h3>
            <p>Distributes requests evenly across all available providers using round-robin.</p>
          </div>
          <div>
            <h3 class="font-medium text-white mb-1">Custom</h3>
            <p>Uses custom routing rules defined by administrators.</p>
          </div>
        </div>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { routingApi, type RoutingStrategy } from '@/services/api/routing'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import { ErrorHandler } from '@/utils/errorHandler'

const loading = ref(false)
const saving = ref(false)
const locking = ref(false)
const strategy = ref<RoutingStrategy | null>(null)
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
    strategy.value = await routingApi.getStrategy()
    selectedStrategy.value = strategy.value.strategy
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    updateResult.value = {
      success: false,
      message: `Failed to load strategy: ${appError.message}`
    }
    ErrorHandler.logError(err, 'RoutingStrategy')
  } finally {
    loading.value = false
  }
}

const handleUpdateStrategy = async () => {
  if (!selectedStrategy.value) return

  saving.value = true
  updateResult.value = null

  try {
    const response = await routingApi.setStrategy({
      strategy: selectedStrategy.value
    })

    updateResult.value = {
      success: true,
      message: response.message || 'Routing strategy updated successfully'
    }

    // Reload to get updated strategy
    await loadStrategy()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    updateResult.value = {
      success: false,
      message: `Failed to update strategy: ${appError.message}`
    }
    ErrorHandler.logError(err, 'RoutingStrategy')
  } finally {
    saving.value = false
  }
}

const handleToggleLock = async () => {
  if (!strategy.value) return

  locking.value = true
  updateResult.value = null

  try {
    const currentLockStatus = strategy.value.is_locked
    const newLockStatus = !currentLockStatus
    
    const response = await routingApi.setStrategyLock(newLockStatus)

    updateResult.value = {
      success: true,
      message: response.message || `Strategy ${newLockStatus ? 'locked' : 'unlocked'} successfully`
    }

    // Immediately update local state for better UX
    if (strategy.value) {
      strategy.value.is_locked = newLockStatus
    }

    // Reload to get updated lock status from server
    await loadStrategy()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    const action = strategy.value?.is_locked ? 'unlock' : 'lock'
    updateResult.value = {
      success: false,
      message: `Failed to ${action} strategy: ${appError.message}`
    }
    ErrorHandler.logError(err, 'RoutingStrategy')
  } finally {
    locking.value = false
  }
}

onMounted(() => {
  loadStrategy()
})
</script>

