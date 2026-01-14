<template>
  <div class="space-y-6">
    <!-- Header -->
    <div>
      <h1 class="text-3xl font-bold text-white">Email Configuration</h1>
      <p class="text-slate-400 mt-1">
        View SMTP configuration status and test email delivery
      </p>
    </div>

    <!-- Configuration Status -->
    <Card>
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h2 class="text-xl font-semibold text-white">SMTP Status</h2>
          <div
            class="px-3 py-1 rounded-full text-sm font-medium"
            :class="
              config?.configured
                ? 'bg-green-500/20 text-green-400'
                : 'bg-red-500/20 text-red-400'
            "
          >
            {{ config?.status || 'Loading...' }}
          </div>
        </div>

        <div v-if="loading" class="text-center py-8">
          <div class="inline-block animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <p class="text-slate-400 mt-2">Loading configuration...</p>
        </div>

        <div v-else-if="config">
          <!-- Configuration Details -->
          <div v-if="config.configured" class="space-y-3">
            <div class="bg-slate-800/50 rounded-lg p-4">
              <h3 class="text-sm font-medium text-slate-300 mb-3">SMTP Details</h3>
              <div class="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <span class="text-slate-400">Host:</span>
                  <span class="text-white ml-2">{{ config.smtp?.host || 'N/A' }}</span>
                </div>
                <div>
                  <span class="text-slate-400">Port:</span>
                  <span class="text-white ml-2">{{ config.smtp?.port || 'N/A' }}</span>
                </div>
              </div>
            </div>

            <div v-if="config.note" class="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4">
              <p class="text-sm text-blue-400">{{ config.note }}</p>
            </div>
          </div>

          <!-- Not Configured -->
          <div v-else class="space-y-3">
            <div class="bg-red-500/10 border border-red-500/20 rounded-lg p-4">
              <p class="text-sm text-red-400 mb-3">{{ config.note }}</p>
              <div v-if="config.required_env_vars" class="mt-3">
                <p class="text-sm font-medium text-red-300 mb-2">Required Environment Variables:</p>
                <ul class="list-disc list-inside space-y-1 text-sm text-red-400">
                  <li v-for="envVar in config.required_env_vars" :key="envVar">{{ envVar }}</li>
                </ul>
              </div>
            </div>
          </div>

          <!-- Troubleshooting -->
          <div v-if="config.troubleshooting" class="bg-slate-800/50 rounded-lg p-4">
            <h3 class="text-sm font-medium text-slate-300 mb-3">Troubleshooting</h3>
            <ul class="space-y-2 text-sm text-slate-400">
              <li v-for="(tip, key) in config.troubleshooting" :key="key">
                • {{ tip }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </Card>

    <!-- Test Email -->
    <Card>
      <div class="space-y-4">
        <h2 class="text-xl font-semibold text-white">Test Email</h2>
        <p class="text-sm text-slate-400">
          Send a test email to verify SMTP configuration is working correctly.
        </p>

        <form @submit.prevent="handleTestEmail" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">
              Recipient Email
            </label>
            <Input
              v-model="testEmail.to"
              type="email"
              placeholder="test@example.com"
              required
              :disabled="testing || !config?.configured"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">
              Subject (optional)
            </label>
            <Input
              v-model="testEmail.subject"
              type="text"
              placeholder="UniRoute SMTP Test Email"
              :disabled="testing || !config?.configured"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-slate-300 mb-2">
              Message (optional)
            </label>
            <textarea
              v-model="testEmail.message"
              rows="4"
              class="w-full px-4 py-2 bg-slate-800 border border-slate-700 rounded-lg text-white placeholder-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
              placeholder="Custom message (HTML supported)"
              :disabled="testing || !config?.configured"
            ></textarea>
          </div>

          <Button
            type="submit"
            :disabled="testing || !config?.configured"
            class="w-full"
          >
            <span v-if="testing">Sending...</span>
            <span v-else>Send Test Email</span>
          </Button>
        </form>

        <div v-if="testResult" class="mt-4 p-4 rounded-lg" :class="testResult.success ? 'bg-green-500/10 border border-green-500/20' : 'bg-red-500/10 border border-red-500/20'">
          <p :class="testResult.success ? 'text-green-400' : 'text-red-400'">
            {{ testResult.message }}
          </p>
          <p v-if="testResult.note" class="text-sm text-slate-400 mt-2">
            {{ testResult.note }}
          </p>
        </div>
      </div>
    </Card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { emailApi, type EmailConfig } from '@/services/api/email'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { ErrorHandler } from '@/utils/errorHandler'

const loading = ref(false)
const testing = ref(false)
const config = ref<EmailConfig | null>(null)
const testEmail = ref({
  to: '',
  subject: '',
  message: ''
})
const testResult = ref<{ success: boolean; message: string; note?: string } | null>(null)

const loadConfig = async () => {
  loading.value = true
  try {
    config.value = await emailApi.getConfig()
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    testResult.value = {
      success: false,
      message: `Failed to load configuration: ${appError.message}`
    }
    ErrorHandler.logError(err, 'EmailConfig')
  } finally {
    loading.value = false
  }
}

const handleTestEmail = async () => {
  if (!testEmail.value.to) return

  testing.value = true
  testResult.value = null

  try {
    const response = await emailApi.testEmail({
      to: testEmail.value.to,
      subject: testEmail.value.subject || undefined,
      message: testEmail.value.message || undefined
    })

    testResult.value = {
      success: true,
      message: response.message,
      note: response.note
    }

    // Clear form
    testEmail.value = {
      to: '',
      subject: '',
      message: ''
    }
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    testResult.value = {
      success: false,
      message: `Failed to send test email: ${appError.message}`
    }
    ErrorHandler.logError(err, 'EmailConfig')
  } finally {
    testing.value = false
  }
}

onMounted(() => {
  loadConfig()
})
</script>

