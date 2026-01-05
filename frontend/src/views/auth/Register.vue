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
</script>

