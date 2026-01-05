/**
 * Sentry error tracking configuration
 */

import * as Sentry from '@sentry/vue'
import { BrowserTracing } from '@sentry/tracing'
import type { App } from 'vue'
import { router } from '@/router'

/**
 * Initialize Sentry error tracking
 * Only initializes in production or if VITE_SENTRY_DSN is set
 */
export function initSentry(app: App) {
  const dsn = import.meta.env.VITE_SENTRY_DSN
  const environment = import.meta.env.MODE || 'development'
  
  // Only initialize if DSN is provided
  if (!dsn) {
    console.log('Sentry DSN not configured, error tracking disabled')
    return
  }

  Sentry.init({
    app,
    dsn,
    environment,
    integrations: [
      new BrowserTracing({
        routingInstrumentation: Sentry.vueRouterInstrumentation(router),
        tracePropagationTargets: ['localhost', /^https:\/\/api\.uniroute\.dev/],
      }),
    ],
    // Performance Monitoring
    tracesSampleRate: environment === 'production' ? 0.1 : 1.0, // 10% in prod, 100% in dev
    
    // Session Replay (optional, can be enabled later)
    // replaysSessionSampleRate: 0.1,
    // replaysOnErrorSampleRate: 1.0,
    
    // Release tracking
    release: import.meta.env.VITE_APP_VERSION || 'unknown',
    
    // Filter out sensitive data
    beforeSend(event, hint) {
      // Don't send errors in development (unless explicitly enabled)
      if (environment === 'development' && import.meta.env.VITE_SENTRY_ENABLE_DEV !== 'true') {
        return null
      }
      
      // Filter out sensitive information
      if (event.request) {
        // Remove sensitive headers
        if (event.request.headers) {
          delete event.request.headers['Authorization']
          delete event.request.headers['X-CSRF-Token']
        }
        
        // Remove sensitive query params
        if (event.request.query_string) {
          const params = new URLSearchParams(event.request.query_string)
          params.delete('token')
          params.delete('password')
          event.request.query_string = params.toString()
        }
      }
      
      // Remove sensitive data from user context
      if (event.user) {
        delete event.user.email // Only send user ID, not email
      }
      
      return event
    },
    
    // Ignore certain errors
    ignoreErrors: [
      // Browser extensions
      'top.GLOBALS',
      'originalCreateNotification',
      'canvas.contentDocument',
      'MyApp_RemoveAllHighlights',
      'atomicFindClose',
      // Network errors that are expected
      'Network Error',
      'Failed to fetch',
      'NetworkError',
      // Chrome extensions
      'chrome-extension://',
      'moz-extension://',
    ],
  })

  console.log('Sentry initialized for error tracking')
}

/**
 * Set user context for Sentry
 */
export function setSentryUser(user: { id: string; email?: string; name?: string }) {
  Sentry.setUser({
    id: user.id,
    email: user.email,
    username: user.name || user.email,
  })
}

/**
 * Clear user context (on logout)
 */
export function clearSentryUser() {
  Sentry.setUser(null)
}

/**
 * Capture exception manually
 */
export function captureException(error: Error, context?: Record<string, any>) {
  Sentry.captureException(error, {
    contexts: {
      custom: context || {},
    },
  })
}

/**
 * Capture message manually
 */
export function captureMessage(message: string, level: Sentry.SeverityLevel = 'error', context?: Record<string, any>) {
  Sentry.captureMessage(message, {
    level,
    contexts: {
      custom: context || {},
    },
  })
}


