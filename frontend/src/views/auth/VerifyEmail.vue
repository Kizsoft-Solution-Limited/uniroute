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
          Verify Your Email
        </h1>
        <h2 class="text-xl text-slate-300">
          {{ verifying ? 'Verifying your email...' : 'Please verify your email address' }}
        </h2>
      </div>

      <Card class="bg-slate-800/60 border-slate-700/50">
        <!-- Success State -->
        <div v-if="verified" class="text-center space-y-6">
          <div class="w-16 h-16 bg-green-500/20 rounded-full flex items-center justify-center mx-auto">
            <svg class="w-8 h-8 text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <div>
            <h3 class="text-2xl font-semibold text-white mb-2">Email Verified!</h3>
            <p class="text-slate-300 mb-6">
              Your email has been successfully verified. You can now sign in to your account.
            </p>
            <router-link
              to="/login"
              class="inline-block px-8 py-3 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-base font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20"
            >
              Sign in
            </router-link>
          </div>
        </div>

        <!-- Verifying State -->
        <div v-else-if="verifying" class="text-center space-y-6">
          <div class="w-16 h-16 bg-blue-500/20 rounded-full flex items-center justify-center mx-auto">
            <svg class="w-8 h-8 text-blue-400 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
          <div>
            <h3 class="text-2xl font-semibold text-white mb-2">Verifying Email...</h3>
            <p class="text-slate-300">
              Please wait while we verify your email address.
            </p>
          </div>
        </div>

        <!-- Resend Form -->
        <div v-else-if="showResendForm && !verified" class="space-y-6">
          <p class="text-slate-300 text-center">
            Enter your email address to receive a new verification link.
          </p>
          <form @submit.prevent="handleResend" class="space-y-6">
            <Input
              v-model="resendEmail"
              label="Email address"
              type="email"
              placeholder="you@example.com"
              :error="emailError"
              required
              @blur="validateEmail"
            />

            <Button
              type="submit"
              :loading="resending"
              :disabled="!isResendValid || resending"
              full-width
              size="lg"
            >
              {{ resending ? 'Sending...' : 'Resend Verification Email' }}
            </Button>

            <div v-if="resendSuccess" class="p-3 bg-green-900/20 border border-green-800/50 rounded-lg">
              <p class="text-sm text-green-400">{{ resendSuccess }}</p>
            </div>
            <div v-if="resendError" class="p-3 bg-red-900/20 border border-red-800/50 rounded-lg">
              <p class="text-sm text-red-400">{{ resendError }}</p>
            </div>
          </form>

          <div class="text-center">
            <button
              @click="showResendForm = false"
              class="text-sm text-slate-400 hover:text-slate-300 transition-colors"
            >
              Back to verification
            </button>
          </div>
        </div>

        <!-- Error State -->
        <div v-else-if="error && !showResendForm" class="text-center space-y-6">
          <div class="w-16 h-16 bg-red-500/20 rounded-full flex items-center justify-center mx-auto">
            <svg class="w-8 h-8 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
          <div>
            <h3 class="text-2xl font-semibold text-white mb-2">Verification Failed</h3>
            <p class="text-slate-300 mb-6">
              {{ error }}
            </p>
            <div class="space-y-3">
              <button
                @click="showResendForm = true"
                class="w-full px-6 py-3 bg-blue-500/20 text-blue-400 rounded-lg text-base font-semibold hover:bg-blue-500/30 transition-all border border-blue-500/30"
              >
                Resend Verification Email
              </button>
              <router-link
                to="/login"
                class="block w-full px-6 py-3 bg-slate-700/50 text-white rounded-lg text-base font-semibold hover:bg-slate-700/70 transition-all text-center"
              >
                Back to Login
              </router-link>
            </div>
          </div>
        </div>

        <!-- Verification Form (default) -->
        <div v-else class="space-y-6">
          <p class="text-slate-300 text-center">
            Enter the verification token sent to your email address.
          </p>
          <form @submit.prevent="handleVerify" class="space-y-6">
            <Input
              v-model="verificationToken"
              label="Verification Token"
              type="text"
              placeholder="Enter verification token"
              :error="tokenError"
              required
              @blur="validateToken"
            />

            <Button
              type="submit"
              :loading="verifying"
              :disabled="verifying"
              full-width
              size="lg"
              @click="handleVerify"
            >
              {{ verifying ? 'Verifying...' : 'Verify Email' }}
            </Button>

            <div v-if="error" class="p-3 bg-red-900/20 border border-red-800/50 rounded-lg">
              <p class="text-sm text-red-400">{{ error }}</p>
            </div>
          </form>

          <div class="text-center">
            <p class="text-sm text-slate-400 mb-2">Didn't receive the email?</p>
            <button
              @click="showResendForm = true"
              class="text-sm text-blue-400 hover:text-blue-300 transition-colors font-medium"
            >
              Resend verification email
            </button>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useValidation, validationSchemas } from '@/composables/useValidation'
import * as yup from 'yup'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { authApi } from '@/services/api/auth'
import { ErrorHandler } from '@/utils/errorHandler'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const verificationToken = ref('')
const tokenError = ref('')
const verifying = ref(false)
const verified = ref(false)
const error = ref<string | null>(null)
const showResendForm = ref(false)

// Resend form
const resendEmail = ref('')
const emailError = ref('')
const resending = ref(false)
const resendSuccess = ref<string | null>(null)
const resendError = ref<string | null>(null)

// Get token from URL query if present
onMounted(() => {
  const token = route.query.token as string
  if (token) {
    verificationToken.value = token
    handleVerify()
  }
})

const tokenSchema = yup.object({
  token: yup.string().required('Verification token is required').min(10, 'Token is too short')
})

const emailSchema = yup.object({
  email: validationSchemas.email
})

const { validate: validateTokenForm } = useValidation(tokenSchema)
const { validate: validateEmailForm } = useValidation(emailSchema)

const token = computed(() => verificationToken.value)

const isValid = computed(() => {
  return verificationToken.value.trim() !== '' && tokenError.value === ''
})

const isResendValid = computed(() => {
  return resendEmail.value.trim() !== '' && emailError.value === ''
})

const validateToken = async () => {
  try {
    await tokenSchema.validateAt('token', { token: verificationToken.value })
    tokenError.value = ''
  } catch (err: any) {
    tokenError.value = err.message
  }
}

const validateEmail = async () => {
  try {
    await emailSchema.validateAt('email', { email: resendEmail.value })
    emailError.value = ''
  } catch (err: any) {
    emailError.value = err.message
  }
}

const handleVerify = async () => {
  // Clear previous errors
  tokenError.value = ''
  error.value = null

  // Validate token
  if (!verificationToken.value || verificationToken.value.trim() === '') {
    tokenError.value = 'Verification token is required'
    return
  }

  // Additional validation
  try {
    await tokenSchema.validateAt('token', { token: verificationToken.value })
  } catch (err: any) {
    tokenError.value = err.message
    return
  }

  verifying.value = true

  try {
    const response = await authApi.verifyEmail(verificationToken.value)
    
    // Store auth data (coerce undefined to null for store types)
    authStore.token = response.token ?? null
    authStore.user = response.user ?? null
    
    verified.value = true
    
    // Redirect to login after 2 seconds so user can sign in
    setTimeout(() => {
      router.push('/login')
    }, 2000)
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message || 'Verification failed. Please check your token and try again.'
  } finally {
    verifying.value = false
  }
}

const handleResend = async () => {
  emailError.value = ''
  resendError.value = null
  resendSuccess.value = null

  const formValid = await validateEmailForm()
  if (!formValid) {
    return
  }

  resending.value = true

  try {
    await authApi.resendVerification(resendEmail.value)
    resendSuccess.value = 'Verification email sent! Please check your inbox.'
    setTimeout(() => {
      showResendForm.value = false
      resendEmail.value = ''
    }, 3000)
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    resendError.value = appError.message || 'Failed to resend verification email. Please try again.'
  } finally {
    resending.value = false
  }
}
</script>

