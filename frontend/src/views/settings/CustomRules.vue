<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h2 class="text-2xl font-bold text-white">Custom Routing Rules</h2>
      <p class="text-slate-400 mt-1">
        Define custom rules to control how your requests are routed to AI providers
      </p>
    </div>

    <!-- Info Banner -->
    <div class="bg-blue-500/10 border-2 border-blue-500/30 rounded-lg p-4 shadow-lg">
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
          <h3 class="font-bold text-blue-400 mb-1 text-lg">ℹ️ About Custom Routing Rules</h3>
          <p class="text-sm text-blue-300/90 mb-2">
            Custom routing rules allow you to specify conditions for routing requests to specific providers. Rules are evaluated in priority order (higher priority first).
          </p>
          <p class="text-xs text-blue-300/70">
            <strong>Note:</strong> These rules only apply when your routing strategy is set to "Custom". If no rules match, the system falls back to model-based routing.
          </p>
        </div>
      </div>
    </div>

    <!-- Rules List -->
    <Card>
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold text-white">Your Custom Rules</h3>
          <Button @click="addRule" variant="primary">
            <span>+ Add Rule</span>
          </Button>
        </div>

        <div v-if="loading" class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="text-slate-400 mt-2">Loading rules...</p>
        </div>

        <div v-else-if="rules.length === 0" class="text-center py-8 text-slate-400">
          <p>No custom rules defined yet.</p>
          <p class="text-sm mt-2">Click "Add Rule" to create your first custom routing rule.</p>
        </div>

        <div v-else class="space-y-3">
          <div
            v-for="(rule, index) in rules"
            :key="index"
            class="bg-slate-800/50 rounded-lg p-4 border border-slate-700"
          >
            <div class="flex items-start justify-between">
              <div class="flex-1 space-y-2">
                <div class="flex items-center gap-3">
                  <h4 class="font-semibold text-white">{{ rule.name || `Rule ${index + 1}` }}</h4>
                  <span
                    class="px-2 py-1 rounded-full text-xs font-medium"
                    :class="rule.enabled ? 'bg-green-500/20 text-green-400' : 'bg-slate-700 text-slate-400'"
                  >
                    {{ rule.enabled ? 'Enabled' : 'Disabled' }}
                  </span>
                  <span class="px-2 py-1 rounded-full text-xs font-medium bg-blue-500/20 text-blue-400">
                    Priority: {{ rule.priority }}
                  </span>
                </div>
                <p v-if="rule.description" class="text-sm text-slate-400">{{ rule.description }}</p>
                <div class="text-sm text-slate-300 space-y-1">
                  <p>
                    <span class="text-slate-500">Condition:</span>
                    <span class="ml-2 font-medium">{{ getConditionText(rule) }}</span>
                  </p>
                  <p>
                    <span class="text-slate-500">Route to:</span>
                    <span class="ml-2 font-medium text-blue-400">{{ rule.provider_name }}</span>
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-2 ml-4">
                <button
                  @click="editRule(index)"
                  class="p-2 text-slate-400 hover:text-blue-400 transition-colors"
                  title="Edit rule"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                  </svg>
                </button>
                <button
                  @click="deleteRule(index)"
                  class="p-2 text-slate-400 hover:text-red-400 transition-colors"
                  title="Delete rule"
                >
                  <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                  </svg>
                </button>
              </div>
            </div>
          </div>
        </div>

        <div v-if="updateResult" class="mt-4 p-4 rounded-lg" :class="updateResult.success ? 'bg-green-500/10 border border-green-500/20' : 'bg-red-500/10 border border-red-500/20'">
          <p :class="updateResult.success ? 'text-green-400' : 'text-red-400'">
            {{ updateResult.message }}
          </p>
        </div>

        <div v-if="rules.length > 0" class="flex justify-end gap-3 pt-4 border-t border-slate-700">
          <Button @click="loadRules" variant="outline" :disabled="saving">
            Cancel
          </Button>
          <Button @click="saveRules" variant="primary" :disabled="saving">
            <span v-if="saving">Saving...</span>
            <span v-else>Save Rules</span>
          </Button>
        </div>
      </div>
    </Card>

    <!-- Rule Editor Modal -->
    <div
      v-if="editingRule !== null"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      @click.self="cancelEdit"
    >
      <Card class="w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        <div class="space-y-4">
          <h3 class="text-xl font-bold text-white">
            {{ editingRule === -1 ? 'Add Rule' : 'Edit Rule' }}
          </h3>

          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Rule Name *</label>
              <input
                v-model="currentRule.name"
                type="text"
                class="w-full px-3 py-2 rounded-lg border border-slate-600 bg-slate-800 text-white"
                placeholder="e.g., Route GPT-4 to OpenAI"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Description</label>
              <textarea
                v-model="currentRule.description"
                rows="2"
                class="w-full px-3 py-2 rounded-lg border border-slate-600 bg-slate-800 text-white"
                placeholder="Optional description"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Condition Type *</label>
              <select
                v-model="currentRule.condition_type"
                class="w-full px-3 py-2 rounded-lg border border-slate-600 bg-slate-800 text-white"
                @change="resetConditionValue"
              >
                <option value="model">Model Name</option>
                <option value="cost_threshold">Cost Threshold</option>
                <option value="latency_threshold">Latency Threshold</option>
              </select>
            </div>

            <div v-if="currentRule.condition_type === 'model'">
              <label class="block text-sm font-medium text-slate-300 mb-2">Model Name *</label>
              <input
                v-model="currentRule.condition_value.model"
                type="text"
                class="w-full px-3 py-2 rounded-lg border border-slate-600 bg-slate-800 text-white"
                placeholder="e.g., gpt-4, claude-3-opus"
              />
            </div>

            <div v-if="currentRule.condition_type === 'cost_threshold'">
              <label class="block text-sm font-medium text-slate-300 mb-2">Maximum Cost (USD) *</label>
              <input
                v-model.number="currentRule.condition_value.max_cost"
                type="number"
                step="0.001"
                min="0"
                class="w-full px-3 py-2 rounded-lg border border-slate-600 bg-slate-800 text-white"
                placeholder="e.g., 0.01"
              />
            </div>

            <div v-if="currentRule.condition_type === 'latency_threshold'">
              <label class="block text-sm font-medium text-slate-300 mb-2">Maximum Latency (ms) *</label>
              <input
                v-model.number="currentRule.condition_value.max_latency_ms"
                type="number"
                min="0"
                class="w-full px-3 py-2 rounded-lg border border-slate-600 bg-slate-800 text-white"
                placeholder="e.g., 1000"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Provider *</label>
              <select
                v-model="currentRule.provider_name"
                class="w-full px-3 py-2 rounded-lg border border-slate-600 bg-slate-800 text-white"
              >
                <option value="">Select provider</option>
                <option value="openai">OpenAI</option>
                <option value="anthropic">Anthropic</option>
                <option value="google">Google</option>
                <option value="local">Local (Ollama)</option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium text-slate-300 mb-2">Priority *</label>
              <input
                v-model.number="currentRule.priority"
                type="number"
                min="0"
                class="w-full px-3 py-2 rounded-lg border border-slate-600 bg-slate-800 text-white"
                placeholder="Higher priority = checked first"
              />
              <p class="text-xs text-slate-400 mt-1">Higher priority rules are evaluated first (e.g., 100 > 50)</p>
            </div>

            <div class="flex items-center gap-2">
              <input
                v-model="currentRule.enabled"
                type="checkbox"
                id="rule-enabled"
                class="w-4 h-4 rounded border-slate-600 bg-slate-800 text-blue-600"
              />
              <label for="rule-enabled" class="text-sm text-slate-300">Enable this rule</label>
            </div>
          </div>

          <div class="flex justify-end gap-3 pt-4 border-t border-slate-700">
            <Button @click="cancelEdit" variant="outline">Cancel</Button>
            <Button @click="saveCurrentRule" variant="primary">Save Rule</Button>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { routingApi, type CustomRoutingRule } from '@/services/api/routing'
import Card from '@/components/ui/Card.vue'
import Button from '@/components/ui/Button.vue'
import { ErrorHandler } from '@/utils/errorHandler'
import { useToast } from '@/composables/useToast'

const { showToast } = useToast()

const loading = ref(false)
const saving = ref(false)
const rules = ref<CustomRoutingRule[]>([])
const editingRule = ref<number | null>(null)
const updateResult = ref<{ success: boolean; message: string } | null>(null)

const currentRule = ref<CustomRoutingRule>({
  name: '',
  condition_type: 'model',
  condition_value: {},
  provider_name: '',
  priority: 0,
  enabled: true,
  description: ''
})

const getConditionText = (rule: CustomRoutingRule): string => {
  switch (rule.condition_type) {
    case 'model':
      return `Model is "${rule.condition_value.model || 'N/A'}"`
    case 'cost_threshold':
      return `Cost ≤ $${rule.condition_value.max_cost || 0}`
    case 'latency_threshold':
      return `Latency ≤ ${rule.condition_value.max_latency_ms || 0}ms`
    default:
      return 'Unknown condition'
  }
}

const resetConditionValue = () => {
  currentRule.value.condition_value = {}
}

const loadRules = async () => {
  loading.value = true
  updateResult.value = null
  try {
    const response = await routingApi.getUserCustomRules()
    rules.value = response.rules || []
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    showToast(`Failed to load rules: ${appError.message}`, 'error')
    ErrorHandler.logError(err, 'CustomRules')
  } finally {
    loading.value = false
  }
}

const saveRules = async () => {
  saving.value = true
  updateResult.value = null

  // Validate rules
  for (const rule of rules.value) {
    if (!rule.name || !rule.provider_name) {
      updateResult.value = {
        success: false,
        message: 'All rules must have a name and provider'
      }
      saving.value = false
      return
    }
    if (rule.condition_type === 'model' && !rule.condition_value.model) {
      updateResult.value = {
        success: false,
        message: 'Model condition requires a model name'
      }
      saving.value = false
      return
    }
    if (rule.condition_type === 'cost_threshold' && !rule.condition_value.max_cost) {
      updateResult.value = {
        success: false,
        message: 'Cost threshold condition requires a maximum cost'
      }
      saving.value = false
      return
    }
    if (rule.condition_type === 'latency_threshold' && !rule.condition_value.max_latency_ms) {
      updateResult.value = {
        success: false,
        message: 'Latency threshold condition requires a maximum latency'
      }
      saving.value = false
      return
    }
  }

  try {
    const response = await routingApi.setUserCustomRules({ rules: rules.value })
    updateResult.value = {
      success: true,
      message: response.message || 'Custom rules saved successfully'
    }
    showToast('Custom rules saved successfully', 'success')
    await loadRules()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    updateResult.value = {
      success: false,
      message: appError.message || 'Failed to save custom rules'
    }
    showToast(`Failed to save rules: ${appError.message}`, 'error')
    ErrorHandler.logError(err, 'CustomRules')
  } finally {
    saving.value = false
  }
}

const addRule = () => {
  currentRule.value = {
    name: '',
    condition_type: 'model',
    condition_value: {},
    provider_name: '',
    priority: rules.value.length > 0 ? Math.max(...rules.value.map(r => r.priority)) + 1 : 10,
    enabled: true,
    description: ''
  }
  editingRule.value = -1
}

const editRule = (index: number) => {
  currentRule.value = { ...rules.value[index] }
  editingRule.value = index
}

const deleteRule = (index: number) => {
  if (confirm('Are you sure you want to delete this rule?')) {
    rules.value.splice(index, 1)
  }
}

const saveCurrentRule = () => {
  // Validate
  if (!currentRule.value.name || !currentRule.value.provider_name) {
    showToast('Name and provider are required', 'error')
    return
  }

  if (editingRule.value === -1) {
    // Add new rule
    rules.value.push({ ...currentRule.value })
  } else if (editingRule.value !== null) {
    // Update existing rule
    rules.value[editingRule.value] = { ...currentRule.value }
  }

  editingRule.value = null
  showToast('Rule saved', 'success')
}

const cancelEdit = () => {
  editingRule.value = null
}

onMounted(() => {
  loadRules()
})
</script>
