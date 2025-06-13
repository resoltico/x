// src/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'
import { debugLogger } from '@/utils/debugLogger'

debugLogger.log('info', 'main', '🚀 Starting Engraving Processor Pro...')

// Enhanced browser compatibility check with detailed logging
const performCompatibilityCheck = () => {
  const issues: string[] = []
  const warnings: string[] = []
  
  // Critical requirements
  if (typeof Worker === 'undefined') {
    issues.push('Web Workers not supported')
  }
  if (typeof ArrayBuffer === 'undefined') {
    issues.push('ArrayBuffer not supported')
  }
  if (typeof Uint8ClampedArray === 'undefined') {
    issues.push('Uint8ClampedArray not supported')
  }
  if (typeof HTMLCanvasElement === 'undefined') {
    issues.push('Canvas not supported')
  }
  
  // Nice-to-have features
  if (typeof OffscreenCanvas === 'undefined') {
    warnings.push('OffscreenCanvas not supported (will use fallback)')
  }
  if (typeof createImageBitmap === 'undefined') {
    warnings.push('ImageBitmap not supported (will use fallback)')
  }
  if (typeof SharedArrayBuffer === 'undefined') {
    warnings.push('SharedArrayBuffer not supported')
  }
  
  // Check WebAssembly support
  if (typeof WebAssembly === 'undefined') {
    warnings.push('WebAssembly not supported (will use JavaScript fallback)')
  }
  
  // Log results
  if (issues.length > 0) {
    debugLogger.log('error', 'main', 'Critical browser compatibility issues', issues)
  }
  if (warnings.length > 0) {
    debugLogger.log('warn', 'main', 'Browser compatibility warnings', warnings)
  }
  if (issues.length === 0 && warnings.length === 0) {
    debugLogger.log('info', 'main', 'Browser compatibility check passed ✓')
  }
  
  return { issues, warnings }
}

const { issues, warnings } = performCompatibilityCheck()

if (issues.length > 0) {
  const errorMessage = `Browser not supported. Missing: ${issues.join(', ')}`
  debugLogger.log('error', 'main', errorMessage)
  
  document.body.innerHTML = `
    <div style="position:fixed;top:0;left:0;right:0;bottom:0;background:#fee2e2;display:flex;align-items:center;justify-content:center;font-family:system-ui">
      <div style="background:white;padding:2rem;border-radius:1rem;box-shadow:0 25px 50px -12px rgba(0,0,0,0.25);max-width:600px;text-align:center">
        <div style="color:#dc2626;font-size:3rem;margin-bottom:1rem">⚠️</div>
        <h1 style="color:#991b1b;margin:0 0 1rem 0">Browser Not Supported</h1>
        <p style="color:#7f1d1d;margin:0 0 1rem 0">${errorMessage}</p>
        <p style="color:#7f1d1d;margin:0 0 1.5rem 0">Please use a modern browser like Chrome 90+, Firefox 88+, Safari 14+, or Edge 90+.</p>
        <div style="margin-bottom:1rem">
          <details style="text-align:left;background:#fef2f2;padding:1rem;border-radius:0.5rem">
            <summary style="cursor:pointer;font-weight:bold">Technical Details</summary>
            <ul style="margin:0.5rem 0;padding-left:1.5rem">
              ${issues.map(issue => `<li>${issue}</li>`).join('')}
            </ul>
          </details>
        </div>
        <button onclick="window.location.reload()" style="background:#dc2626;color:white;border:none;padding:0.5rem 1rem;border-radius:0.5rem;cursor:pointer">Reload</button>
      </div>
    </div>
  `
  throw new Error('Browser not supported')
}

// Show warnings if any
if (warnings.length > 0) {
  debugLogger.log('warn', 'main', 'Browser warnings', warnings)
}

// Initialize app with enhanced error handling
debugLogger.log('info', 'main', 'Creating Vue app...')

const pinia = createPinia()
const app = createApp(App)

app.use(pinia)

// Enhanced error handler with debug logging
app.config.errorHandler = (error, vm, info) => {
  debugLogger.log('error', 'vue', 'Vue error occurred', { 
    error: error instanceof Error ? error.message : String(error),
    info,
    stack: error instanceof Error ? error.stack : undefined
  })
  
  // Also log to console for immediate visibility
  console.error('❌ Vue Error:', error, info)
}

// Global error handlers with debug integration
window.addEventListener('error', (event) => {
  debugLogger.log('error', 'window', 'Global error', {
    message: event.message,
    filename: event.filename,
    lineno: event.lineno,
    colno: event.colno,
    error: event.error
  })
})

window.addEventListener('unhandledrejection', (event) => {
  debugLogger.log('error', 'window', 'Unhandled promise rejection', {
    reason: event.reason,
    promise: event.promise
  })
  event.preventDefault()
})

// Performance monitoring setup
if ((performance as any).memory) {
  debugLogger.log('info', 'main', 'Memory info available', {
    used: Math.round((performance as any).memory.usedJSHeapSize / 1024 / 1024) + 'MB',
    total: Math.round((performance as any).memory.totalJSHeapSize / 1024 / 1024) + 'MB',
    limit: Math.round((performance as any).memory.jsHeapSizeLimit / 1024 / 1024) + 'MB'
  })
}

// Mount app with error handling
try {
  debugLogger.log('info', 'main', 'Mounting Vue app...')
  app.mount('#app')
  debugLogger.log('info', 'main', '✅ Application started successfully')
} catch (mountError) {
  debugLogger.log('error', 'main', 'Failed to mount Vue app', mountError)
  throw mountError
}

// Development helpers with debug integration
if (import.meta.env.DEV) {
  ;(window as any).__APP__ = app
  ;(window as any).__DEBUG__ = debugLogger
  debugLogger.log('info', 'main', '🔧 Development mode - enhanced debugging available')
  debugLogger.log('info', 'main', 'Available globals: window.__APP__, window.__DEBUG__, window.debugLogger')
  
  // Auto-run diagnostics in development after a delay
  setTimeout(async () => {
    try {
      debugLogger.log('info', 'main', 'Running automatic diagnostics...')
      await debugLogger.diagnoseWorkerSupport()
    } catch (error) {
      debugLogger.log('error', 'main', 'Auto-diagnostics failed', error)
    }
  }, 3000)
} else {
  debugLogger.log('info', 'main', '🏭 Production mode')
}