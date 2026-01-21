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
          Get started
        </h1>
        <h2 class="text-xl text-slate-300">
          Create your account
        </h2>
      </div>

      <Card class="bg-slate-800/60 border-slate-700/50">
        <!-- Success State -->
        <div v-if="registrationSuccess" class="text-center space-y-6">
          <div class="w-16 h-16 bg-blue-500/20 rounded-full flex items-center justify-center mx-auto">
            <svg class="w-8 h-8 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
            </svg>
          </div>
          <div>
            <h3 class="text-2xl font-semibold text-white mb-2">Check Your Email!</h3>
            <p class="text-slate-300 mb-4">
              We've sent a verification link to <strong class="text-white">{{ email }}</strong>
            </p>
            <p class="text-sm text-slate-400 mb-6">
              Click the link in the email to verify your account and get started.
            </p>
            <div class="space-y-3">
              <router-link
                to="/verify-email"
                class="block w-full px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-500 text-white rounded-lg text-base font-semibold hover:from-blue-600 hover:to-indigo-600 transition-all shadow-lg shadow-blue-500/20 text-center"
              >
                Verify Email
              </router-link>
              <button
                @click="resendVerification"
                :disabled="resending"
                class="w-full px-6 py-3 bg-slate-700/50 text-white rounded-lg text-base font-semibold hover:bg-slate-700/70 transition-all disabled:opacity-50"
              >
                {{ resending ? 'Sending...' : "Didn't receive email? Resend" }}
              </button>
            </div>
          </div>
        </div>

        <!-- Registration Form -->
        <form v-else @submit.prevent="handleRegister" class="space-y-6">
          <Input
            v-model="name"
            label="Full name"
            type="text"
            placeholder="John Doe"
            :error="nameError"
            @blur="validateName"
          />

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
            hint="At least 8 characters with uppercase, lowercase, and number"
          />

          <Input
            v-model="confirmPassword"
            label="Confirm password"
            type="password"
            placeholder="••••••••"
            :error="confirmPasswordError"
            required
            @blur="validateConfirmPassword"
          />

          <div class="flex items-center">
            <input
              v-model="agreeToTerms"
              type="checkbox"
              class="rounded border-slate-600 bg-slate-700 text-blue-500 focus:ring-blue-500"
              required
            />
            <label class="ml-2 text-sm text-slate-300">
              I agree to the
              <router-link to="/terms" class="text-blue-400 hover:text-blue-300 transition-colors">Terms of Service</router-link>
              and
              <router-link to="/privacy" class="text-blue-400 hover:text-blue-300 transition-colors">Privacy Policy</router-link>
            </label>
          </div>

          <Button
            type="submit"
            :loading="loading"
            :disabled="!isValid || !agreeToTerms"
            full-width
            size="lg"
          >
            Create account
          </Button>

          <div v-if="error || storeError" class="p-3 bg-red-900/20 border border-red-800/50 rounded-lg">
            <p class="text-sm text-red-400">{{ error || storeError }}</p>
          </div>
        </form>

        <!-- OAuth Buttons -->
        <div v-if="!registrationSuccess" class="mt-6">
          <div class="relative">
            <div class="absolute inset-0 flex items-center">
              <div class="w-full border-t border-slate-700"></div>
            </div>
            <div class="relative flex justify-center text-sm">
              <span class="px-2 bg-slate-800/60 text-slate-400">Or continue with</span>
            </div>
          </div>

          <div class="mt-6 grid grid-cols-2 gap-3">
            <button
              @click="handleGoogleRegister"
              :disabled="oauthLoading"
              class="flex items-center justify-center px-4 py-2 border border-slate-700 rounded-lg bg-slate-800/60 text-white hover:bg-slate-700/60 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="w-5 h-5 mr-2" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
              </svg>
           
            </button>
            <button
              @click="handleXRegister"
              :disabled="oauthLoading"
              class="flex items-center justify-center px-4 py-2 border border-slate-700 rounded-lg bg-slate-800/60 text-white hover:bg-slate-700/60 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z"/>
              </svg>
        
            </button>
          </div>
        </div>

        <div class="mt-6 text-center">
          <p class="text-sm text-slate-300">
            Already have an account?
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
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useValidation, validationSchemas } from '@/composables/useValidation'
import { useToast } from '@/composables/useToast'
import * as yup from 'yup'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { ErrorHandler } from '@/utils/errorHandler'
import { authApi } from '@/services/api/auth'

const router = useRouter()
const authStore = useAuthStore()
const { showToast } = useToast()

const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const agreeToTerms = ref(false)

const nameError = ref('')
const emailError = ref('')
const passwordError = ref('')
const confirmPasswordError = ref('')

const loading = computed(() => authStore.loading)
const storeError = computed(() => authStore.error)
const error = ref('')
const registrationSuccess = ref(false)
const resending = ref(false)
const oauthLoading = ref(false)

const registerSchema = yup.object({
  name: yup.string().optional(),
  email: validationSchemas.email,
  password: validationSchemas.password,
  confirmPassword: yup
    .string()
    .oneOf([yup.ref('password')], 'Passwords must match')
    .required('Please confirm your password')
})

const { validate } = useValidation(registerSchema)

// Manual validation state tracking - check if form is valid
// Simplified: just check if fields are filled and passwords match
// Full validation happens on submit
const isValid = computed(() => {
  // Check if all required fields are filled
  const fieldsFilled = email.value.trim() !== '' && 
                       password.value.trim() !== '' && 
                       confirmPassword.value.trim() !== ''
  
  // Check if passwords match (allow empty confirm password initially)
  const passwordsMatch = confirmPassword.value === '' || password.value === confirmPassword.value
  
  // Only check for errors if they exist (don't block button if no errors yet)
  const hasErrors = emailError.value !== '' || 
                   passwordError.value !== '' || 
                   confirmPasswordError.value !== ''
  
  return fieldsFilled && passwordsMatch && !hasErrors
})

// Watch for changes and validate fields
watch([email], () => {
  if (email.value) {
    validateEmail()
  }
})

watch([password], () => {
  if (password.value) {
    validatePassword()
  }
})

watch([confirmPassword], () => {
  if (confirmPassword.value) {
    validateConfirmPassword()
  }
})

const validateName = async () => {
  try {
    await registerSchema.validateAt('name', { name: name.value })
    nameError.value = ''
  } catch (err: any) {
    nameError.value = err.message
  }
}

const validateEmail = async () => {
  try {
    await registerSchema.validateAt('email', { email: email.value })
    emailError.value = ''
  } catch (err: any) {
    emailError.value = err.message
  }
}

const validatePassword = async () => {
  try {
    await registerSchema.validateAt('password', { password: password.value })
    passwordError.value = ''
    // Also validate confirm password if it has a value
    if (confirmPassword.value) {
      await validateConfirmPassword()
    }
  } catch (err: any) {
    passwordError.value = err.message
  }
}

const validateConfirmPassword = async () => {
  try {
    await registerSchema.validateAt('confirmPassword', {
      password: password.value,
      confirmPassword: confirmPassword.value
    })
    confirmPasswordError.value = ''
  } catch (err: any) {
    confirmPasswordError.value = err.message
  }
}

const handleRegister = async () => {
  // Clear previous errors
  nameError.value = ''
  emailError.value = ''
  passwordError.value = ''
  confirmPasswordError.value = ''
  error.value = ''

  // Validate all fields first
  await Promise.all([
    validateEmail(),
    validatePassword(),
    validateConfirmPassword()
  ])

  // Check if form is valid
  if (!isValid.value || !agreeToTerms.value) {
    error.value = 'Please fill in all required fields correctly and agree to the terms'
    return
  }

  // Additional validation using schema
  try {
    await registerSchema.validate({
      name: name.value,
      email: email.value,
      password: password.value,
      confirmPassword: confirmPassword.value
    }, { abortEarly: false })
  } catch (validationErr: any) {
    // Set field-specific errors
    if (validationErr.inner) {
      validationErr.inner.forEach((err: any) => {
        if (err.path === 'email') emailError.value = err.message
        if (err.path === 'password') passwordError.value = err.message
        if (err.path === 'confirmPassword') confirmPasswordError.value = err.message
      })
    }
    error.value = 'Please fix the errors above'
    return
  }

  try {
    const response = await authStore.register({
      name: name.value || undefined,
      email: email.value,
      password: password.value
    })

    // Registration successful - show verification message
    registrationSuccess.value = true
  } catch (err: any) {
    // Error is already set in the store, but we can add additional handling
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message || 'Registration failed. Please try again.'
    console.error('Registration error:', appError)
  }
}

const resendVerification = async () => {
  resending.value = true
  error.value = ''
  try {
    await authApi.resendVerification(email.value)
    // Show success toast
    showToast('Verification email sent! Please check your inbox.', 'success', 5000)
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    const errorMessage = appError.message || 'Failed to resend verification email.'
    error.value = errorMessage
    // Show error toast
    showToast(errorMessage, 'error', 5000)
  } finally {
    resending.value = false
  }
}

const handleGoogleRegister = async () => {
  oauthLoading.value = true
  error.value = ''

  try {
    const response = await oauthApi.getGoogleAuthURL()
    // Redirect to Google OAuth
    window.location.href = response.auth_url
  } catch (err: any) {
    oauthLoading.value = false
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message || 'Failed to initiate Google registration'
  }
}

const handleXRegister = async () => {
  oauthLoading.value = true
  error.value = ''

  try {
    const response = await oauthApi.getXAuthURL()
    // Redirect to X OAuth
    window.location.href = response.auth_url
  } catch (err: any) {
    oauthLoading.value = false
    const appError = ErrorHandler.handleApiError(err)
    error.value = appError.message || 'Failed to initiate X registration'
  }
}
</script>

