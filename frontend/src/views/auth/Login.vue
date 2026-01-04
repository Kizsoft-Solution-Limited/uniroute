<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-purple-50 dark:from-gray-900 dark:to-gray-800 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div class="text-center">
        <h1 class="text-4xl font-bold text-gray-900 dark:text-white mb-2">
          UniRoute
        </h1>
        <h2 class="text-2xl font-semibold text-gray-700 dark:text-gray-300">
          Sign in to your account
        </h2>
      </div>

      <Card class="glass">
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
                class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <span class="ml-2 text-sm text-gray-600 dark:text-gray-400">
                Remember me
              </span>
            </label>
            <router-link
              to="/forgot-password"
              class="text-sm text-blue-600 hover:text-blue-500 dark:text-blue-400"
            >
              Forgot password?
            </router-link>
          </div>

          <Button
            type="submit"
            :loading="loading"
            :disabled="!isValid"
            full-width
            size="lg"
          >
            Sign in
          </Button>

          <div v-if="error" class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <p class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
          </div>
        </form>

        <div class="mt-6 text-center">
          <p class="text-sm text-gray-600 dark:text-gray-400">
            Don't have an account?
            <router-link
              to="/register"
              class="font-medium text-blue-600 hover:text-blue-500 dark:text-blue-400"
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
import { ref, computed } from 'vue'
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

const loading = computed(() => authStore.loading)
const error = computed(() => authStore.error)

const loginSchema = yup.object({
  email: validationSchemas.email,
  password: validationSchemas.required('Password is required')
})

const { isValid, validate } = useValidation(loginSchema)

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

  // Validate form
  const formValid = await validate()
  if (!formValid) {
    return
  }

  try {
    await authStore.login({
      email: email.value,
      password: password.value
    })

    // Redirect to dashboard or intended route
    const redirect = route.query.redirect as string
    router.push(redirect || '/dashboard')
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    // Error is already set in store, but we can add additional handling here
    console.error('Login error:', appError)
  }
}
</script>

