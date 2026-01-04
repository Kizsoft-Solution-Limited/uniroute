/**
 * Centralized error handling utility
 */

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
      return {
        message: error.response.data?.message || 'An error occurred',
        code: error.response.data?.code,
        status: error.response.status,
        details: error.response.data?.details
      }
    } else if (error.request) {
      // Request made but no response
      return {
        message: 'Network error. Please check your connection.',
        code: 'NETWORK_ERROR'
      }
    } else {
      // Something else happened
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
   * Log error (in production, send to error tracking service)
   */
  static logError(error: Error | AppError, context?: string) {
    if (import.meta.env.DEV) {
      console.error(`[${context || 'Error'}]`, error)
    }
    // TODO: Send to error tracking service (Sentry, etc.)
  }
}

