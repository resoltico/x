import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'

// Initialize Pinia store
const pinia = createPinia()

// Create and mount Vue app
const app = createApp(App)
app.use(pinia)
app.mount('#app')

// Global error handling
window.addEventListener('error', (event) => {
  console.error('Global error:', event.error)
})

window.addEventListener('unhandledrejection', (event) => {
  console.error('Unhandled promise rejection:', event.reason)
})