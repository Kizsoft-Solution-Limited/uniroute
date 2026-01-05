import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'
import './assets/styles/main.css'

const app = createApp(App)

const pinia = createPinia()
app.use(pinia)
app.use(router)

// Check authentication on app start
const authStore = useAuthStore()
authStore.checkAuth().then(() => {
  app.mount('#app')
}).catch(() => {
  app.mount('#app')
})

