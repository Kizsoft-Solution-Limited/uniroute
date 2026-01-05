/**
 * Centralized error handling utility
 */

import { logException, logMessage } from './errorLogger'

export interface AppError {
  message: string
  code?: string
  status?: number
  details?: any
}

export class ErrorHandler {
  /**
   * Handle API errors
   */
  static handleApiError(error: any): AppError {
    if (error.response) {
      // Server responded with error
      const status = error.response.status
      const data = error.response.data || {}
      
      // Handle connection errors (status 0 usually means no connection)
      if (status === 0 || data.code === 'CONNECTION_ERROR' || data.code === 'CONNECTION_REFUSED') {
        return {
          message: data.message || 'Unable to connect to server. Please ensure the backend is running on port 8084.',
          code: data.code || 'CONNECTION_ERROR',
          status: 0,
          details: data.details || 'The server may be down or unreachable.'
        }
      }
      
      // Backend may return 'error' or 'message' field
      const errorMessage = data.message || data.error || 'An error occurred'
      
      return {
        message: errorMessage,
        code: data.code,
        status: status,
        details: data.details
      }
    } else if (error.request) {
      // Request made but no response (network error, EOF, etc.)
      const errorMessage = error.message || ''
      if (errorMessage.includes('EOF') || errorMessage.includes('ECONNREFUSED') || errorMessage.includes('Network Error')) {
        return {
          message: 'Connection failed. Please ensure the backend server is running on port 8084.',
          code: 'CONNECTION_ERROR',
          details: 'The server may be down or unreachable. Check if the backend is running.'
        }
      }
      
      return {
        message: 'Network error. Please check your connection.',
        code: 'NETWORK_ERROR',
        details: 'Unable to reach the server. Please try again later.'
      }
    } else {
      // Something else happened (including EOF errors)
      if (error.message?.includes('EOF') || error.message === 'EOF') {
        return {
          message: 'Connection closed unexpectedly. Please ensure the backend server is running.',
          code: 'EOF_ERROR',
          details: 'The server connection was closed. This usually means the backend is not running.'
        }
      }
      
      return {
        message: error.message || 'An unexpected error occurred',
        code: 'UNKNOWN_ERROR'
      }
    }
  }

  /**
   * Handle validation errors
   */
  static handleValidationError(errors: any): string[] {
    if (Array.isArray(errors)) {
      return errors
    }
    if (typeof errors === 'object') {
      return Object.values(errors).flat() as string[]
    }
    return [errors?.toString() || 'Validation error']
  }

  /**
   * Format error message for display
   */
  static formatError(error: AppError | string): string {
    if (typeof error === 'string') {
      return error
    }
    return error.message || 'An error occurred'
  }

  /**
   * Log error (sends to backend error logging endpoint)
   */
  static logError(error: Error | AppError, context?: string) {
    if (import.meta.env.DEV) {
      console.error(`[${context || 'Error'}]`, error)
    }
    
    // Send to backend error logging
    try {
      if (error instanceof Error) {
        logException(error, {
          context: context || 'ErrorHandler',
          errorType: 'AppError',
        })
      } else {
        // AppError object
        logMessage(error.message || 'Unknown error', 'error', {
          context: context || 'ErrorHandler',
          code: error.code,
          status: error.status,
          details: error.details,
        })
      }
    } catch (err) {
      // Silently fail - don't break the app
      if (import.meta.env.DEV) {
        console.warn('Failed to log error to backend:', err)
      }
    }
  }
}

