<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-50 to-purple-50 dark:from-gray-900 dark:to-gray-800 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
      <div class="text-center">
        <h1 class="text-4xl font-bold text-gray-900 dark:text-white mb-2">
          UniRoute
        </h1>
        <h2 class="text-2xl font-semibold text-gray-700 dark:text-gray-300">
          Create your account
        </h2>
      </div>

      <Card class="glass">
        <form @submit.prevent="handleRegister" class="space-y-6">
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
              class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              required
            />
            <label class="ml-2 text-sm text-gray-600 dark:text-gray-400">
              I agree to the
              <a href="/terms" class="text-blue-600 hover:text-blue-500 dark:text-blue-400">Terms of Service</a>
              and
              <a href="/privacy" class="text-blue-600 hover:text-blue-500 dark:text-blue-400">Privacy Policy</a>
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

          <div v-if="error" class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg">
            <p class="text-sm text-red-600 dark:text-red-400">{{ error }}</p>
          </div>
        </form>

        <div class="mt-6 text-center">
          <p class="text-sm text-gray-600 dark:text-gray-400">
            Already have an account?
            <router-link
              to="/login"
              class="font-medium text-blue-600 hover:text-blue-500 dark:text-blue-400"
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
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useValidation, validationSchemas } from '@/composables/useValidation'
import * as yup from 'yup'
import Card from '@/components/ui/Card.vue'
import Input from '@/components/ui/Input.vue'
import Button from '@/components/ui/Button.vue'
import { ErrorHandler } from '@/utils/errorHandler'

const router = useRouter()
const authStore = useAuthStore()

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
const error = computed(() => authStore.error)

const registerSchema = yup.object({
  name: yup.string().optional(),
  email: validationSchemas.email,
  password: validationSchemas.password,
  confirmPassword: yup
    .string()
    .oneOf([yup.ref('password')], 'Passwords must match')
    .required('Please confirm your password')
})

const { isValid, validate } = useValidation(registerSchema)

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

  // Validate form
  const formValid = await validate()
  if (!formValid) {
    return
  }

  try {
    await authStore.register({
      name: name.value || undefined,
      email: email.value,
      password: password.value
    })

    // Redirect to dashboard
    router.push('/dashboard')
  } catch (err: any) {
    const appError = ErrorHandler.handleApiError(err)
    console.error('Registration error:', appError)
  }
}
</script>

