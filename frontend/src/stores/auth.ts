import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/services/api/auth'
import type { User, LoginCredentials, RegisterData } from '@/types/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  const isAuthenticated = computed(() => !!token.value && !!user.value)

  const hasPermission = (permission: string): boolean => {
    if (!user.value) return false
    
    // Admin has all permissions
    if (user.value.roles?.includes('admin')) {
      return true
    }
    
    // Check user-specific permissions
    // TODO: Fetch from API or include in user object
    const userPermissions = user.value.permissions || []
    
    // Support wildcard permissions (e.g., 'api-keys:*' matches 'api-keys:read')
    return userPermissions.some((p: string) => {
      if (p === permission) return true
      if (p.endsWith(':*')) {
        const prefix = p.slice(0, -2)
        return permission.startsWith(prefix + ':')
      }
      return false
    })
  }

  const hasRole = (role: string): boolean => {
    if (!user.value) return false
    return user.value.roles?.includes(role) || false
  }

  const login = async (credentials: LoginCredentials) => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.login(credentials)
      token.value = response.token
      user.value = response.user
      
      // Token storage is handled by the component based on rememberMe
      // This allows for sessionStorage vs localStorage choice
      
      return response
    } catch (err: any) {
      error.value = err.message || 'Login failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  const register = async (data: RegisterData) => {
    loading.value = true
    error.value = null
    try {
      const response = await authApi.register(data)
      // Registration now returns message instead of token (email verification required)
      // If token is present, store it; otherwise, user needs to verify email
      if (response.token) {
        token.value = response.token
        user.value = response.user
        localStorage.setItem('auth_token', response.token)
      } else if (response.user) {
        // Store user info even without token (for verification flow)
        user.value = response.user
      }
      
      return response
    } catch (err: any) {
      error.value = err.message || 'Registration failed'
      throw err
    } finally {
      loading.value = false
    }
  }

  const logout = async () => {
    loading.value = true
    try {
      await authApi.logout()
    } catch (err) {
      console.error('Logout error:', err)
    } finally {
      token.value = null
      user.value = null
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_token_expires')
      localStorage.removeItem('remembered_email')
      sessionStorage.removeItem('auth_token')
      
      loading.value = false
    }
  }

  const checkAuth = async () => {
    // Check localStorage first (remember me), then sessionStorage
    let storedToken = localStorage.getItem('auth_token')
    
    // Check if token has expired (if expiration is set)
    if (storedToken) {
      const expires = localStorage.getItem('auth_token_expires')
      if (expires && Date.now() > parseInt(expires)) {
        // Token expired, clear it
        localStorage.removeItem('auth_token')
        localStorage.removeItem('auth_token_expires')
        storedToken = null
      }
    }
    
    // Fall back to sessionStorage if no localStorage token
    if (!storedToken) {
      storedToken = sessionStorage.getItem('auth_token')
    }
    
    if (!storedToken) {
      return false
    }

    token.value = storedToken
    try {
      const profile = await authApi.getProfile()
      user.value = profile
      
      return true
    } catch (err) {
      // Token invalid, clear it
      token.value = null
      localStorage.removeItem('auth_token')
      localStorage.removeItem('auth_token_expires')
      sessionStorage.removeItem('auth_token')
      
      return false
    }
  }

  const refreshToken = async () => {
    try {
      const response = await authApi.refreshToken()
      token.value = response.token
      if (response.token) {
        localStorage.setItem('auth_token', response.token)
      }
      return response
    } catch (err) {
      // Refresh failed, logout
      await logout()
      throw err
    }
  }

  const resendVerificationEmail = async (email: string) => {
    loading.value = true
    error.value = null
    try {
      await authApi.resendVerification(email)
      return { message: 'Verification email sent successfully' }
    } catch (err: any) {
      // Extract error message from response
      const errorMessage = err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to resend verification email'
      error.value = errorMessage
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    hasPermission,
    hasRole,
    login,
    register,
    logout,
    checkAuth,
    refreshToken,
    resendVerificationEmail
  }
})

