<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950 py-12 px-4 sm:px-6 lg:px-8">
    <!-- Navigation -->
    <nav class="fixed top-0 left-0 right-0 z-50 bg-slate-950/95 backdrop-blur-xl border-b border-slate-800/50">
      <div class="container mx-auto px-6 py-4">
        <div class="flex items-center justify-between">
          <router-link to="/" class="flex items-center space-x-3">
            <div class="w-10 h-10 bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500 rounded-lg flex items-center justify-center shadow-lg shadow-blue-500/20">
              <span class="text-white font-bold text-lg">U</span>
            </div>
            <span class="text-xl font-bold text-white tracking-tight">UniRoute</span>
          </router-link>
          <router-link
            to="/login"
            class="px-4 py-2 bg-slate-800/60 text-white rounded-lg text-sm font-semibold hover:bg-slate-700/60 transition-all"
          >
            Sign in
          </router-link>
        </div>
      </div>
    </nav>

    <div class="max-w-md w-full space-y-8 pt-20">
      <div class="text-center">
        <h1 class="text-4xl font-bold text-white mb-2">
          Reset password
        </h1>
        <h2 class="text-xl text-slate-300">
          Enter your email to receive a reset link
        </h2>
      </div>

      <Card class="bg-slate-800/60 border-slate-700/50">
        <form @submit.prevent="handleRequestReset" class="space-y-6">
          <Input
            v-model="email"
            label="Email address"
            type="email"
            placeholder="you@example.com"
            :error="emailError"
            required
            @blur="validateEmail"
          />

          <Button
            type="submit"
            :loading="loading"
            :disabled="!email || emailError !== ''"
            full-width
            size="lg"
          >
            Send reset link
          </Button>

          <div v-if="error" class="p-3 bg-red-900/20 border border-red-800/50 rounded-lg">
            <p class="text-sm text-red-400">{{ error }}</p>
          </div>

          <div v-if="success" class="p-3 bg-green-900/20 border border-green-800/50 rounded-lg">
            <p class="text-sm text-green-400">{{ success }}</p>
          </div>
        </form>

        <div class="mt-6 text-center">
          <p class="text-sm text-slate-300">
            Remember your password?
            <router-link
              to="/login"
              class="font-medium text-blue-400 hover:text-blue-300 transition-colors"
            >
              Sign in
            </router-link>
          </p>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useValidation, validationSchemas } from '@/composables/useValidation'
import * as yup from 'yup'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { authApi } from '@/services/api/auth'
import { ErrorHandler } from '@/utils/errorHandler'

const router = useRouter()
const route = useRoute()

const email = ref('')
const emailError = ref('')
const loading = ref(false)
const error = ref('')
const success = ref('')

const resetSchema = yup.object({
  email: validationSchemas.email
})

const { validate } = useValidation(resetSchema)

const validateEmail = async () => {
  try {
    await resetSchema.validateAt('email', { email: email.value })
    emailError.value = ''
  } catch (err: any) {
    emailError.value = err.message
  }
}

const handleRequestReset = async () => {
  emailError.value = ''
  error.value = ''
  success.value = ''

  const formValid = await validate()
  if (!formValid) {
    return
  }

  loading.value = true
  try {
    await authApi.requestPasswordReset(email.value)
    success.value = 'If the email exists, a password reset link has been sent. Please check your inbox.'
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message || 'Failed to send reset link'
  } finally {
    loading.value = false
  }
}
</script>

