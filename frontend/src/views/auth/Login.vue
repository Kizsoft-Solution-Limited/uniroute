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
            to="/register"
            class="px-4 py-2 bg-slate-800/60 text-white rounded-lg text-sm font-semibold hover:bg-slate-700/60 transition-all"
          >
            Sign up
          </router-link>
        </div>
      </div>
    </nav>

    <div class="max-w-md w-full space-y-8 pt-20">
      <div class="text-center">
        <h1 class="text-4xl font-bold text-white mb-2">
          Welcome back
        </h1>
        <h2 class="text-xl text-slate-300">
          Sign in to your account
        </h2>
      </div>

      <Card class="bg-slate-800/60 border-slate-700/50">
        <form @submit.prevent="handleLogin" class="space-y-6">
          <Input
            v-model="email"
            label="Email address"
            type="email"
            placeholder="you@example.com"
            :error="emailError"
            required
            @blur="validateEmail"
          />

          <Input
            v-model="password"
            label="Password"
            type="password"
            placeholder="••••••••"
            :error="passwordError"
            required
            @blur="validatePassword"
          />

          <div class="flex items-center justify-between">
            <label class="flex items-center">
              <input
                v-model="rememberMe"
                type="checkbox"
                class="rounded border-slate-600 bg-slate-700 text-blue-500 focus:ring-blue-500"
              />
              <span class="ml-2 text-sm text-slate-300">
                Remember me
              </span>
            </label>
            <router-link
              to="/forgot-password"
              class="text-sm text-blue-400 hover:text-blue-300 transition-colors"
            >
              Forgot password?
            </router-link>
          </div>

          <Button
            type="submit"
            :loading="loading"
            :disabled="!isValid || loading"
            full-width
            size="lg"
          >
            {{ loading ? 'Signing in...' : 'Sign in' }}
          </Button>

          <div v-if="error" :class="error.includes('verify') || error.includes('verification') || error.includes('sent') ? 'p-4 bg-yellow-900/20 border border-yellow-800/50 rounded-lg' : 'p-3 bg-red-900/20 border border-red-800/50 rounded-lg'">
            <p :class="error.includes('verify') || error.includes('verification') || error.includes('sent') ? 'text-sm text-yellow-400 mb-2' : 'text-sm text-red-400'">{{ error }}</p>
            <div v-if="(error.includes('verify') || error.includes('verification')) && !error.includes('sent')" class="mt-3">
              <p class="text-xs text-slate-400 mb-2">Didn't receive the email?</p>
              <button
                type="button"
                @click="resendVerification"
                class="text-sm text-blue-400 hover:text-blue-300 underline"
                :disabled="resendingVerification"
              >
                {{ resendingVerification ? 'Sending...' : 'Resend verification email' }}
              </button>
            </div>
          </div>
        </form>

        <div class="mt-6 text-center">
          <p class="text-sm text-slate-300">
            Don't have an account?
            <router-link
              to="/register"
              class="font-medium text-blue-400 hover:text-blue-300 transition-colors"
            >
              Sign up
            </router-link>
          </p>
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
import { ErrorHandler } from '@/utils/errorHandler'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const rememberMe = ref(false)
const emailError = ref('')
const passwordError = ref('')
const error = ref<string | null>(null)
const resendingVerification = ref(false)

// Use auth store loading state
const loading = computed(() => authStore.loading)

// Load remembered email if exists
onMounted(() => {
  const rememberedEmail = localStorage.getItem('remembered_email')
  if (rememberedEmail) {
    email.value = rememberedEmail
    rememberMe.value = true
  }
})

const loginSchema = yup.object({
  email: validationSchemas.email,
  password: validationSchemas.required('Password is required')
})

// Custom validate function that uses actual form values
const validate = async (): Promise<boolean> => {
  try {
    await loginSchema.validate({
      email: email.value,
      password: password.value
    }, { abortEarly: false })
    return true
  } catch (error: any) {
    // Set field errors
    if (error.inner) {
      error.inner.forEach((err: any) => {
        if (err.path === 'email') {
          emailError.value = err.message
        } else if (err.path === 'password') {
          passwordError.value = err.message
        }
      })
    }
    return false
  }
}

// Check if form is valid (both fields filled and no errors)
const isValid = computed(() => {
  const fieldsFilled = email.value.trim() !== '' && password.value.trim() !== ''
  const noErrors = emailError.value === '' && passwordError.value === ''
  return fieldsFilled && noErrors && !loading.value
})

const validateEmail = async () => {
  try {
    await loginSchema.validateAt('email', { email: email.value })
    emailError.value = ''
  } catch (err: any) {
    emailError.value = err.message
  }
}

const validatePassword = async () => {
  try {
    await loginSchema.validateAt('password', { password: password.value })
    passwordError.value = ''
  } catch (err: any) {
    passwordError.value = err.message
  }
}

const handleLogin = async () => {
  // Clear previous errors
  emailError.value = ''
  passwordError.value = ''
  error.value = null

  // Validate form
  const formValid = await validate()
  if (!formValid) {
    // Trigger field validation to show errors
    await validateEmail()
    await validatePassword()
    return
  }

  try {
    await authStore.login({
      email: email.value,
      password: password.value
    })

    // Handle remember me
    if (rememberMe.value) {
      localStorage.setItem('remembered_email', email.value)
      // Store token with longer expiration (30 days)
      const token = authStore.token
      if (token) {
        localStorage.setItem('auth_token', token)
        localStorage.setItem('auth_token_expires', String(Date.now() + 30 * 24 * 60 * 60 * 1000))
      }
    } else {
      // Remove remembered email
      localStorage.removeItem('remembered_email')
      // Use sessionStorage for token (cleared on browser close)
      const token = authStore.token
      if (token) {
        sessionStorage.setItem('auth_token', token)
        localStorage.removeItem('auth_token')
        localStorage.removeItem('auth_token_expires')
      }
    }

    // Redirect to dashboard or intended route
    const redirect = route.query.redirect as string
    // Validate redirect is an internal route (security: prevent open redirect)
    if (redirect && (redirect.startsWith('/dashboard') || redirect.startsWith('/'))) {
      router.push(redirect)
    } else {
      router.push('/dashboard')
    }
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    
    // Check if email is not verified
    if (err.response?.data?.code === 'EMAIL_NOT_VERIFIED' || err.response?.status === 403) {
      // Show email verification message with resend option
      // Backend automatically sends verification email, so message reflects that
      error.value = err.response?.data?.message || 'Please verify your email address before logging in. A verification link has been sent to your email.'
    } else {
      // Use error message from backend (could be in 'error' or 'message' field)
      const backendError = err.response?.data?.error || err.response?.data?.message
      error.value = backendError || appError.message || 'Login failed. Please check your credentials.'
    }
  }
}

const resendVerification = async () => {
  if (!email.value) {
    error.value = 'Please enter your email address first'
    return
  }

  resendingVerification.value = true
  error.value = null

  try {
    await authStore.resendVerificationEmail(email.value)
    // Show success message
    error.value = 'Verification email sent! Please check your inbox.'
    // Clear success message after 5 seconds
    setTimeout(() => {
      if (error.value?.includes('sent')) {
        error.value = null
      }
    }, 5000)
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message || 'Failed to resend verification email'
  } finally {
    resendingVerification.value = false
  }
}
</script>

