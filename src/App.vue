<template>
  <div id="app" class="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
    <!-- Header -->
    <header class="bg-white shadow-sm border-b border-slate-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
        <div class="flex items-center justify-between">
          <div class="flex items-center space-x-4">
            <h1 class="text-2xl font-bold text-slate-900">
              Engraving Processor Pro
            </h1>
            <span class="px-2 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full">
              AI-Powered
            </span>
            <!-- System Status -->
            <div class="flex items-center space-x-2 text-xs">
              <div 
                class="w-2 h-2 rounded-full"
                :class="systemStatus.color"
                :title="systemStatus.text"
              ></div>
              <span class="text-slate-500">{{ systemStatus.text }}</span>
            </div>
          </div>
          <div class="flex items-center space-x-4">
            <span class="text-sm text-slate-500">v{{ version }}</span>
            <div class="flex items-center space-x-2 text-xs text-slate-400">
              <span>{{ environmentInfo.label }}</span>
              <span v-if="environmentInfo.warning" class="text-yellow-600" :title="environmentInfo.warning">⚠️</span>
            </div>
          </div>
        </div>
      </div>
    </header>

    <!-- Error Banner -->
    <div v-if="error" class="bg-red-600 text-white px-4 py-2 text-sm">
      <div class="max-w-7xl mx-auto flex items-center justify-between">
        <div class="flex items-center">
          <svg class="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
          <span>{{ error }}</span>
        </div>
        <button class="hover:bg-red-700 px-2 py-1 rounded text-xs" @click="error = null">
          Dismiss
        </button>
      </div>
    </div>

    <!-- Browser Compatibility Warning -->
    <div v-if="compatibilityWarning" class="bg-yellow-600 text-white px-4 py-2 text-sm">
      <div class="max-w-7xl mx-auto flex items-center justify-between">
        <div class="flex items-center">
          <svg class="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
          <span>{{ compatibilityWarning }}</span>
        </div>
        <button class="hover:bg-yellow-700 px-2 py-1 rounded text-xs" @click="compatibilityWarning = null">
          Dismiss
        </button>
      </div>
    </div>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div class="grid grid-cols-1 lg:grid-cols-3 gap-8">
        <!-- Left Column: Image Input & Preview -->
        <div class="lg:col-span-2 space-y-6">
          <ImageInput />
          <PreviewRenderer />
        </div>

        <!-- Right Column: Controls & Processing -->
        <div class="space-y-6">
          <ProcessingControls />
          <ProgressDisplay />
        </div>
      </div>
    </main>

    <!-- Footer -->
    <footer class="bg-white border-t border-slate-200 mt-16">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div class="text-center text-sm text-slate-500">
          <p>&copy; 2025 Ervins Strauhmanis. Built with Vue 3, TypeScript & WebAssembly.</p>
          <div class="mt-2 flex items-center justify-center space-x-4 text-xs">
            <span>Environment: {{ environmentInfo.label }}</span>
            <span>Browser: {{ browserInfo }}</span>
            <span v-if="performanceInfo">Memory: {{ performanceInfo }}</span>
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import ImageInput from './components/ImageInput.vue'
import PreviewRenderer from './components/PreviewRenderer.vue'
import ProcessingControls from './components/ProcessingControls.vue'
import ProgressDisplay from './components/ProgressDisplay.vue'

const version = ref('1.0.0')
const error = ref<string | null>(null)
const compatibilityWarning = ref<string | null>(null)
const store = useAppStore()

// Environment detection
const environmentInfo = computed(() => {
  const isFileProtocol = window.location.protocol === 'file:'
  const isLocalhost = window.location.hostname === 'localhost' || 
                     window.location.hostname === '127.0.0.1' ||
                     window.location.hostname === '0.0.0.0'
  const isDevelopment = isLocalhost && (window.location.port === '3000' || window.location.port === '5173')
  const isPreview = isLocalhost && window.location.port === '4173'
  const isProduction = !isDevelopment && !isPreview && !isFileProtocol

  if (isFileProtocol) {
    return {
      label: 'File Protocol',
      warning: 'Some features may not work properly when served from file:// protocol'
    }
  }
  
  if (isDevelopment) {
    return {
      label: `Development${window.location.port ? ` (${window.location.port})` : ''}`,
      warning: null
    }
  }
  
  if (isPreview) {
    return {
      label: `Preview${window.location.port ? ` (${window.location.port})` : ''}`,
      warning: null
    }
  }
  
  if (isProduction) {
    return {
      label: 'Production',
      warning: null
    }
  }
  
  return {
    label: 'Unknown',
    warning: 'Unable to detect environment'
  }
})

// Browser detection
const browserInfo = computed(() => {
  const ua = navigator.userAgent
  if (ua.includes('Chrome')) return 'Chrome'
  if (ua.includes('Firefox')) return 'Firefox'
  if (ua.includes('Safari') && !ua.includes('Chrome')) return 'Safari'
  if (ua.includes('Edge')) return 'Edge'
  return 'Unknown'
})

// Performance monitoring
const performanceInfo = computed(() => {
  const perf = (performance as any).memory
  if (perf) {
    const used = Math.round(perf.usedJSHeapSize / 1024 / 1024)
    const total = Math.round(perf.totalJSHeapSize / 1024 / 1024)
    return `${used}/${total} MB`
  }
  return null
})

// System status indicator
const systemStatus = computed(() => {
  if (error.value) return { color: 'bg-red-500', text: 'Error' }
  if (store.isProcessing) return { color: 'bg-blue-500', text: 'Processing' }
  if (!store.isInitialized) return { color: 'bg-yellow-500', text: 'Initializing' }
  return { color: 'bg-green-500', text: 'Ready' }
})

// Browser compatibility check
const checkBrowserCompatibility = () => {
  const issues: string[] = []
  
  if (typeof Worker === 'undefined') {
    issues.push('Web Workers not supported')
  }
  
  if (typeof OffscreenCanvas === 'undefined') {
    issues.push('OffscreenCanvas not supported (will use fallback)')
  }
  
  if (typeof createImageBitmap === 'undefined') {
    issues.push('ImageBitmap not supported (will use fallback)')
  }
  
  if (window.location.protocol === 'file:') {
    issues.push('File protocol detected - some features may not work')
  }
  
  if (issues.length > 0) {
    compatibilityWarning.value = `Browser compatibility issues: ${issues.join(', ')}`
  }
}

// URL parameter handling for fallback mode
const checkUrlParameters = () => {
  if (typeof window !== 'undefined' && window.URLSearchParams) {
    const urlParams = new window.URLSearchParams(window.location.search)
    if (urlParams.get('fallback') === 'true') {
      console.warn('🔧 Fallback mode requested via URL parameter')
      // Set up global fallback flag
      ;(window as any).__FORCE_FALLBACK = () => {
        console.log('🔧 Forcing fallback mode...')
        // This would be handled by the system status manager
      }
    }
  }
}

// Global error handlers with better error categorization
const setupErrorHandlers = () => {
  window.addEventListener('error', (event) => {
    const message = event.message || 'Unknown error'
    const filename = event.filename || 'Unknown file'
    const lineno = event.lineno || 0
    
    console.error('🔴 Global Error:', { message, filename, lineno, error: event.error })
    
    // Categorize errors
    if (message.includes('Loading') || message.includes('fetch')) {
      error.value = `Resource loading error: ${message}`
    } else if (message.includes('Worker') || message.includes('worker')) {
      error.value = `Worker error: ${message}`
    } else if (message.includes('WebAssembly') || message.includes('wasm')) {
      error.value = `WebAssembly error: ${message} (falling back to JavaScript processing)`
    } else {
      error.value = `Application error: ${message}`
    }
  })

  window.addEventListener('unhandledrejection', (event) => {
    const reason = event.reason
    console.error('🔴 Unhandled Promise Rejection:', reason)
    
    if (reason instanceof Error) {
      if (reason.message.includes('Loading') || reason.message.includes('import')) {
        error.value = `Module loading error: ${reason.message}`
      } else if (reason.message.includes('Worker')) {
        error.value = `Worker initialization error: ${reason.message}`
      } else {
        error.value = `Promise rejection: ${reason.message}`
      }
    } else {
      error.value = `Promise rejected: ${String(reason)}`
    }
    
    // Prevent the default unhandled rejection behavior
    event.preventDefault()
  })
}

// Performance monitoring
let performanceInterval: number | null = null

const startPerformanceMonitoring = () => {
  performanceInterval = window.setInterval(() => {
    const perf = (performance as any).memory
    if (perf) {
      const used = perf.usedJSHeapSize / 1024 / 1024
      const limit = perf.jsHeapSizeLimit / 1024 / 1024
      
      // Warn if memory usage is getting high
      if (used > limit * 0.8) {
        console.warn(`⚠️ High memory usage: ${Math.round(used)}MB / ${Math.round(limit)}MB`)
      }
    }
  }, 30000) // Check every 30 seconds
}

// Lifecycle hooks
onMounted(() => {
  console.log('🎛️ App component loaded')
  
  checkBrowserCompatibility()
  checkUrlParameters()
  setupErrorHandlers()
  startPerformanceMonitoring()
  
  // Initialize the store
  store.initialize().catch((initError) => {
    console.error('Failed to initialize store:', initError)
    error.value = `Initialization failed: ${initError.message}`
  })
})

onUnmounted(() => {
  if (performanceInterval) {
    clearInterval(performanceInterval)
  }
})
</script>