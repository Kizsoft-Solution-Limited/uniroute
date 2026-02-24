import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import DOMPurify from 'dompurify'

const getBaseURL = () => {
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }
  if (import.meta.env.DEV) {
    return 'http://localhost:8084'
  }
  if (typeof window !== 'undefined') {
    const host = window.location.hostname
    if (host === 'uniroute.co' || host === 'www.uniroute.co') {
      return 'https://app.uniroute.co'
    }
    return window.location.origin
  }
  return 'https://app.uniroute.co'
}

export function getTunnelServerURL(): string {
  if (import.meta.env.VITE_TUNNEL_SERVER_URL) {
    return import.meta.env.VITE_TUNNEL_SERVER_URL.replace(/\/$/, '')
  }
  if (import.meta.env.DEV) {
    return 'http://localhost:8080'
  }
  if (typeof window !== 'undefined') {
    const host = window.location.hostname
    if (host === 'uniroute.co' || host === 'www.uniroute.co' || host === 'app.uniroute.co') {
      return 'https://tunnel.uniroute.co'
    }
    return window.location.origin
  }
  return 'https://tunnel.uniroute.co'
}

export const apiClient: AxiosInstance = axios.create({
  baseURL: getBaseURL(),
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  },
  withCredentials: true
})

apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    let token = localStorage.getItem('auth_token')
    if (!token) {
      token = sessionStorage.getItem('auth_token')
    }
    
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }

    if (config.data && typeof config.data === 'object') {
      config.data = sanitizeObject(config.data)
    }

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

apiClient.interceptors.response.use(
  (response: AxiosResponse) => {
    if (response.data && typeof response.data === 'object') {
      response.data = sanitizeResponse(response.data)
    }
    return response
  },
  async (error) => {
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

      switch (status) {
        case 401: {
          const url = error.config?.url ?? ''
          const isAuthCheck = url.includes('/auth/profile') || url.includes('/auth/refresh')
          if (isAuthCheck) {
            localStorage.removeItem('auth_token')
            localStorage.removeItem('auth_token_expires')
            sessionStorage.removeItem('auth_token')
          }
          break
        }
        case 403:
          console.error('Access forbidden')
          break
        case 404:
          console.error('Resource not found')
          break
        case 429:
          console.error('Rate limit exceeded')
          break
        case 500:
        case 502:
        case 503:
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
          }).catch(() => {})
          console.error('Server error')
          break
      }

      if (error.response.data?.message) {
        error.response.data.message = DOMPurify.sanitize(error.response.data.message, { ALLOWED_TAGS: [] })
      }
    } else if (error.request && !error.response) {
      import('@/utils/errorLogger').then(({ logNetworkError }) => {
        logNetworkError(`Network error: ${error.message || 'Unknown network error'}`, {
          code: error.code,
          message: error.message,
        })
      }).catch(() => {})

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

function sanitizeResponse(data: any): any {
  if (typeof data === 'string') {
    return DOMPurify.sanitize(data)
  }
  return data
}

function getCookie(name: string): string | null {
  if (typeof document === 'undefined') return null
  const value = `; ${document.cookie}`
  const parts = value.split(`; ${name}=`)
  if (parts.length === 2) {
    return parts.pop()?.split(';').shift() || null
  }
  return null
}

