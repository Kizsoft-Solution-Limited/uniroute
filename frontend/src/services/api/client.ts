import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import DOMPurify from 'dompurify'

// Create axios instance with base configuration
// For local development, use http://localhost:8084 (gateway port)
// For production, use https://api.uniroute.dev
const getBaseURL = () => {
  // Check environment variable first
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }
  
  // For local development, default to localhost
  if (import.meta.env.DEV) {
    return 'http://localhost:8084'
  }
  
  // Production default
  return 'https://api.uniroute.dev'
}

export const apiClient: AxiosInstance = axios.create({
  baseURL: getBaseURL(),
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: true // For httpOnly cookies
})

// Request interceptor - Add auth token and sanitize
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    // Get token from localStorage (remember me) or sessionStorage (session only)
    let token = localStorage.getItem('auth_token')
    if (!token) {
      token = sessionStorage.getItem('auth_token')
    }
    
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
    // Handle network errors (EOF, connection refused, etc.)
    if (error.code === 'ECONNABORTED' || error.message === 'Network Error' || error.message?.includes('EOF')) {
      const networkError = {
        ...error,
        response: {
          ...error.response,
          data: {
            message: 'Unable to connect to server. Please check if the backend is running.',
            code: 'CONNECTION_ERROR',
            details: 'The server may be down or unreachable. Please try again later.'
          },
          status: 0
        }
      }
      return Promise.reject(networkError)
    }

    // Handle connection refused errors
    if (error.code === 'ERR_NETWORK' || error.code === 'ECONNREFUSED') {
      const connectionError = {
        ...error,
        response: {
          ...error.response,
          data: {
            message: 'Connection refused. Please ensure the backend server is running on port 8084.',
            code: 'CONNECTION_REFUSED',
            details: `Cannot connect to ${getBaseURL()}. Is the server running?`
          },
          status: 0
        }
      }
      return Promise.reject(connectionError)
    }

    if (error.response) {
      const status = error.response.status
      
      // Handle specific error codes
      switch (status) {
        case 401:
          // Unauthorized - clear token
          localStorage.removeItem('auth_token')
          localStorage.removeItem('auth_token_expires')
          sessionStorage.removeItem('auth_token')
          
          // Only redirect if not already on login page and not during auth check
          // The router guard will handle the redirect, so we don't need to do it here
          // This prevents double redirects and redirect loops
          const isAuthCheck = error.config?.url?.includes('/auth/profile') || error.config?.url?.includes('/auth/refresh')
          if (window.location.pathname !== '/login' && !isAuthCheck) {
            // Let the router guard handle the redirect for better UX
            // window.location.href = '/login?redirect=' + encodeURIComponent(window.location.pathname)
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
          // Server errors - log to backend
          import('@/utils/errorLogger').then(({ logServerError }) => {
            logServerError(
              `Server error ${status}: ${error.response.data?.error || error.response.data?.message || 'Unknown error'}`,
              status,
              {
                url: error.config?.url,
                method: error.config?.method,
                responseData: error.response.data,
              }
            )
          }).catch(() => {
            // Error logger not available - silently fail
          })
          console.error('Server error')
          break
      }

      // Sanitize error messages to prevent XSS
      if (error.response.data?.message) {
        error.response.data.message = DOMPurify.sanitize(error.response.data.message, { ALLOWED_TAGS: [] })
      }
    } else if (error.request && !error.response) {
      // Network errors - log to backend
      import('@/utils/errorLogger').then(({ logNetworkError }) => {
        logNetworkError(`Network error: ${error.message || 'Unknown network error'}`, {
          code: error.code,
          message: error.message,
        })
      }).catch(() => {
        // Error logger not available - silently fail
      })
      
      // Request was made but no response received (network error)
      const networkError = {
        ...error,
        response: {
          data: {
            message: 'Network error. Please check your connection and ensure the backend server is running.',
            code: 'NETWORK_ERROR',
            details: `Unable to reach ${getBaseURL()}`
          },
          status: 0
        }
      }
      return Promise.reject(networkError)
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

