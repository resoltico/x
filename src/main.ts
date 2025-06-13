// src/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'

console.log('🚀 Starting Engraving Processor Pro...')

// Quick browser compatibility check
const isCompatible = 
  typeof Worker !== 'undefined' &&
  typeof ArrayBuffer !== 'undefined' &&
  typeof Uint8ClampedArray !== 'undefined' &&
  typeof HTMLCanvasElement !== 'undefined'

if (!isCompatible) {
  document.body.innerHTML = `
    <div style="position:fixed;top:0;left:0;right:0;bottom:0;background:#fee2e2;display:flex;align-items:center;justify-content:center;font-family:system-ui">
      <div style="background:white;padding:2rem;border-radius:1rem;box-shadow:0 25px 50px -12px rgba(0,0,0,0.25);max-width:500px;text-align:center">
        <div style="color:#dc2626;font-size:3rem;margin-bottom:1rem">⚠️</div>
        <h1 style="color:#991b1b;margin:0 0 1rem 0">Browser Not Supported</h1>
        <p style="color:#7f1d1d;margin:0 0 1.5rem 0">Please use a modern browser like Chrome, Firefox, Safari, or Edge.</p>
        <button onclick="window.location.reload()" style="background:#dc2626;color:white;border:none;padding:0.5rem 1rem;border-radius:0.5rem;cursor:pointer">Reload</button>
      </div>
    </div>
  `
  throw new Error('Browser not supported')
}

// Initialize app
const pinia = createPinia()
const app = createApp(App)

app.use(pinia)

// Simple error handler
app.config.errorHandler = (error) => {
  console.error('❌ Vue Error:', error)
}

// Global error handlers
window.addEventListener('error', (event) => {
  console.error('❌ Global Error:', event.error)
})

window.addEventListener('unhandledrejection', (event) => {
  console.error('❌ Unhandled Rejection:', event.reason)
  event.preventDefault()
})

// Mount app
app.mount('#app')

console.log('✅ Application started successfully')

// Development helpers
if (import.meta.env.DEV) {
  ;(window as any).__APP__ = app
  console.log('🔧 Development mode - app available as window.__APP__')
}