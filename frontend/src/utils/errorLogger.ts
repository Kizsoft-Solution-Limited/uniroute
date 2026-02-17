import { apiClient } from '@/services/api/client'

export interface ErrorLogData {
  error_type: 'exception' | 'message' | 'network' | 'server'
  message: string
  stack_trace?: string
  url?: string
  context?: Record<string, any>
  severity?: 'error' | 'warning' | 'info'
}

export async function logError(data: ErrorLogData): Promise<void> {
  try {
    const url = data.url || window.location.href

    let stackTrace = data.stack_trace
    if (!stackTrace && data.error_type === 'exception') {
      stackTrace = new Error().stack
    }

    await apiClient.post('/api/errors/log', {
      error_type: data.error_type,
      message: data.message,
      stack_trace: stackTrace,
      url,
      context: {
        ...data.context,
        user_agent: navigator.userAgent,
        timestamp: new Date().toISOString(),
      },
      severity: data.severity || 'error',
    })
  } catch (err) {
    if (import.meta.env.DEV) {
      console.warn('Failed to log error to backend:', err)
    }
  }
}

export function logException(error: Error, context?: Record<string, any>): void {
  logError({
    error_type: 'exception',
    message: error.message,
    stack_trace: error.stack,
    context: {
      ...context,
      name: error.name,
    },
    severity: 'error',
  })
}

export function logMessage(message: string, severity: 'error' | 'warning' | 'info' = 'error', context?: Record<string, any>): void {
  logError({
    error_type: 'message',
    message,
    context,
    severity,
  })
}

export function logNetworkError(message: string, context?: Record<string, any>): void {
  logError({
    error_type: 'network',
    message,
    context,
    severity: 'error',
  })
}

export function logServerError(message: string, statusCode?: number, context?: Record<string, any>): void {
  logError({
    error_type: 'server',
    message,
    context: {
      ...context,
      status_code: statusCode,
    },
    severity: 'error',
  })
}


