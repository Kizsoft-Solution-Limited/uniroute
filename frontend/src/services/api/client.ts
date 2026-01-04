import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import DOMPurify from 'dompurify'

// Create axios instance with base configuration
export const apiClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || 'https://api.uniroute.dev',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: true // For httpOnly cookies
})

// Request interceptor - Add auth token and sanitize
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Get token from localStorage (httpOnly cookies preferred in production)
    const token = localStorage.getItem('auth_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }

    // Sanitize request data if it's an object
    if (config.data && typeof config.data === 'object') {
      config.data = sanitizeObject(config.data)
    }

    // Add CSRF token if enabled
    if (import.meta.env.VITE_CSRF_TOKEN_ENABLED === 'true') {
      const csrfToken = getCookie('csrf_token')
      if (csrfToken && config.headers) {
        config.headers['X-CSRF-Token'] = csrfToken
      }
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor - Handle errors and sanitize
apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    // Sanitize response data if it contains HTML
    if (response.data && typeof response.data === 'object') {
      response.data = sanitizeResponse(response.data)
    }
    return response
  },
  async (error) => {
    if (error.response) {
      const status = error.response.status
      
      // Handle specific error codes
      switch (status) {
        case 401:
          // Unauthorized - clear token and redirect to login
          localStorage.removeItem('auth_token')
          if (window.location.pathname !== '/login') {
            window.location.href = '/login?redirect=' + encodeURIComponent(window.location.pathname)
          }
          break
        case 403:
          // Forbidden
          console.error('Access forbidden')
          break
        case 404:
          // Not found
          console.error('Resource not found')
          break
        case 429:
          // Rate limited
          console.error('Rate limit exceeded')
          break
        case 500:
        case 502:
        case 503:
          // Server errors
          console.error('Server error')
          break
      }

      // Sanitize error messages to prevent XSS
      if (error.response.data?.message) {
        error.response.data.message = DOMPurify.sanitize(error.response.data.message, { ALLOWED_TAGS: [] })
      }
    }
    
    return Promise.reject(error)
  }
)

// Sanitize object recursively
function sanitizeObject(obj: any): any {
  if (typeof obj !== 'object' || obj === null) {
    return typeof obj === 'string' ? DOMPurify.sanitize(obj, { ALLOWED_TAGS: [] }) : obj
  }

  if (Array.isArray(obj)) {
    return obj.map(sanitizeObject)
  }

  const sanitized: any = {}
  for (const key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      sanitized[key] = sanitizeObject(obj[key])
    }
  }
  return sanitized
}

// Sanitize response data
function sanitizeResponse(data: any): any {
  // Only sanitize if it's a string or contains HTML
  if (typeof data === 'string') {
    return DOMPurify.sanitize(data)
  }
  return data
}

// Helper function to get cookie value
function getCookie(name: string): string | null {
  if (typeof document === 'undefined') return null
  const value = `; ${document.cookie}`
  const parts = value.split(`; ${name}=`)
  if (parts.length === 2) {
    return parts.pop()?.split(';').shift() || null
  }
  return null
}

