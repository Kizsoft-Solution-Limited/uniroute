<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-slate-950 via-blue-950 to-indigo-950">
    <div class="text-center">
      <div v-if="loading" class="space-y-4">
        <div class="inline-block animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        <p class="text-white text-lg">Completing authentication...</p>
      </div>
      <div v-else-if="error" class="space-y-4">
        <div class="w-16 h-16 bg-red-500/20 rounded-full flex items-center justify-center mx-auto">
          <svg class="w-8 h-8 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </div>
        <h2 class="text-2xl font-semibold text-white mb-2">Authentication Failed</h2>
        <p class="text-slate-300 mb-6">{{ error }}</p>
        <router-link
          to="/login"
          class="inline-block px-6 py-3 bg-blue-500 text-white rounded-lg font-semibold hover:bg-blue-600 transition-colors"
        >
          Back to Login
        </router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const loading = ref(true)
const error = ref<string | null>(null)

onMounted(async () => {
  const token = route.query.token as string
  const provider = route.query.provider as string
  const errorParam = route.query.error as string

  if (errorParam) {
    loading.value = false
    error.value = 'Authentication failed. Please try again.'
    return
  }

  if (!token) {
    loading.value = false
    error.value = 'No authentication token received.'
    return
  }

  try {
    // Store token
    authStore.setToken(token)
    sessionStorage.setItem('auth_token', token)

    // Verify auth and get user info
    await authStore.checkAuth()

    // Redirect to dashboard
    router.push('/dashboard')
  } catch (err: any) {
    loading.value = false
    error.value = 'Failed to complete authentication. Please try again.'
    console.error('OAuth callback error:', err)
  }
})
</script>
