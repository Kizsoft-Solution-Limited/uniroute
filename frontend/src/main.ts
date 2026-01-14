import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import { logException, logMessage } from './utils/errorLogger'
import './assets/styles/main.css'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(router)

// Global error handler for unhandled errors
window.addEventListener('error', (event) => {
  logException(new Error(event.message), {
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    error: event.error?.toString(),
  })
})

// Global handler for unhandled promise rejections
window.addEventListener('unhandledrejection', (event) => {
  const error = event.reason instanceof Error 
    ? event.reason 
    : new Error(String(event.reason))
  logException(error, {
    type: 'unhandledrejection',
    reason: String(event.reason),
  })
})

// Vue error handler
app.config.errorHandler = (err, instance, info) => {
  const error = err instanceof Error ? err : new Error(String(err))
  logException(error, {
    component: instance?.$options?.name || 'Unknown',
    info,
    type: 'vue-error',
  })
}

// Check authentication on app start
const authStore = useAuthStore()
authStore.checkAuth().then(() => {
  app.mount('#app')
}).catch(() => {
  app.mount('#app')
})

