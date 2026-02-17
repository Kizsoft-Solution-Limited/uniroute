import * as Sentry from '@sentry/vue'
import { BrowserTracing } from '@sentry/tracing'
import type { App } from 'vue'
import { router } from '@/router'

export function initSentry(app: App) {
  const dsn = import.meta.env.VITE_SENTRY_DSN
  const environment = import.meta.env.MODE || 'development'
  
  if (!dsn) {
    if (import.meta.env.DEV) {
      console.debug('Sentry DSN not configured, error tracking disabled')
    }
    return
  }

  Sentry.init({
    app,
    dsn,
    environment,
    integrations: [
      new BrowserTracing({
        routingInstrumentation: Sentry.vueRouterInstrumentation(router),
        tracePropagationTargets: ['localhost', /^https:\/\/api\.uniroute\.co/],
      }),
    ],
    tracesSampleRate: environment === 'production' ? 0.1 : 1.0,
    release: import.meta.env.VITE_APP_VERSION || 'unknown',
    beforeSend(event, hint) {
      if (environment === 'development' && import.meta.env.VITE_SENTRY_ENABLE_DEV !== 'true') {
        return null
      }
      if (event.request) {
        if (event.request.headers) {
          delete event.request.headers['Authorization']
          delete event.request.headers['X-CSRF-Token']
        }
        if (event.request.query_string) {
          const params = new URLSearchParams(event.request.query_string)
          params.delete('token')
          params.delete('password')
          event.request.query_string = params.toString()
        }
      }
      if (event.user) {
        delete event.user.email
      }
      return event
    },
    ignoreErrors: [
      'top.GLOBALS',
      'originalCreateNotification',
      'canvas.contentDocument',
      'MyApp_RemoveAllHighlights',
      'atomicFindClose',
      'Network Error',
      'Failed to fetch',
      'NetworkError',
      'chrome-extension://',
      'moz-extension://',
    ],
  })

  if (import.meta.env.DEV) {
    console.debug('Sentry initialized for error tracking')
  }
}

export function setSentryUser(user: { id: string; email?: string; name?: string }) {
  Sentry.setUser({
    id: user.id,
    email: user.email,
    username: user.name || user.email,
  })
}

export function clearSentryUser() {
  Sentry.setUser(null)
}

export function captureException(error: Error, context?: Record<string, any>) {
  Sentry.captureException(error, {
    contexts: {
      custom: context || {},
    },
  })
}

export function captureMessage(message: string, level: Sentry.SeverityLevel = 'error', context?: Record<string, any>) {
  Sentry.captureMessage(message, {
    level,
    contexts: {
      custom: context || {},
    },
  })
}


