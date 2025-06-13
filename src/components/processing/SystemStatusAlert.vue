<!-- src/components/processing/SystemStatusAlert.vue -->
<template>
  <div v-if="status.error" class="mb-4 bg-red-50 border border-red-200 rounded-lg p-4">
    <div class="flex">
      <svg class="h-5 w-5 text-red-400 mt-0.5 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
        <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
      </svg>
      <div class="ml-3 flex-1">
        <h3 class="text-sm font-medium text-red-800">System Initialization Error</h3>
        <div class="mt-2 text-sm text-red-700">
          <p class="mb-2">{{ status.error }}</p>
          
          <!-- Troubleshooting suggestions -->
          <div class="bg-red-100 border border-red-200 rounded p-3 mt-3">
            <h4 class="text-xs font-medium text-red-800 mb-2">💡 Troubleshooting Tips:</h4>
            <ul class="text-xs text-red-700 space-y-1 list-disc list-inside">
              <li>Check your browser's console for detailed error messages</li>
              <li>Ensure you're using a modern browser (Chrome, Firefox, Safari, Edge)</li>
              <li>Try refreshing the page or restarting your browser</li>
              <li>Disable browser extensions that might interfere with web workers</li>
              <li v-if="isFileProtocol">You're using file:// protocol - try serving over HTTP instead</li>
              <li v-if="!hasWebWorkerSupport">Your browser doesn't support Web Workers</li>
              <li v-if="!hasOffscreenCanvasSupport">Limited browser support detected - processing will use fallback mode</li>
            </ul>
          </div>
        </div>
        
        <div class="mt-4 flex items-center space-x-3">
          <button 
            class="text-sm bg-red-600 text-white px-3 py-1 rounded hover:bg-red-700 transition-colors"
            :disabled="isRetrying"
            @click="handleRetry"
          >
            <span v-if="isRetrying" class="flex items-center">
              <svg class="animate-spin -ml-1 mr-1 h-3 w-3 text-white" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Retrying...
            </span>
            <span v-else>Retry Initialization</span>
          </button>
          
          <button 
            class="text-sm text-red-600 hover:text-red-800 underline"
            @click="forceEnableFallback"
          >
            Use Fallback Mode
          </button>
          
          <button 
            class="text-sm text-red-600 hover:text-red-800 underline"
            @click="showDetailedInfo = !showDetailedInfo"
          >
            {{ showDetailedInfo ? 'Hide' : 'Show' }} Details
          </button>
        </div>
        
        <!-- Detailed error information -->
        <div v-if="showDetailedInfo" class="mt-4 p-3 bg-red-100 border border-red-200 rounded">
          <h4 class="text-xs font-medium text-red-800 mb-2">🔧 Technical Details:</h4>
          <div class="text-xs text-red-700 font-mono space-y-1">
            <div><strong>Environment:</strong> {{ status.environment }}</div>
            <div><strong>Browser:</strong> {{ browserInfo }}</div>
            <div><strong>Workers:</strong> {{ status.totalWorkers }}/{{ status.availableWorkers }} available</div>
            <div><strong>Web Worker Support:</strong> {{ hasWebWorkerSupport ? 'Yes' : 'No' }}</div>
            <div><strong>OffscreenCanvas Support:</strong> {{ hasOffscreenCanvasSupport ? 'Yes' : 'No' }}</div>
            <div><strong>Hardware Concurrency:</strong> {{ hardwareConcurrency }}</div>
            <div><strong>Protocol:</strong> {{ location.protocol }}</div>
            <div><strong>URL:</strong> {{ location.href }}</div>
          </div>
        </div>
      </div>
    </div>
  </div>
  
  <!-- Warning for limited functionality -->
  <div v-else-if="!status.initialized && !status.error" class="mb-4 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
    <div class="flex">
      <svg class="h-5 w-5 text-yellow-400 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
        <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
      </svg>
      <div class="ml-3">
        <h3 class="text-sm font-medium text-yellow-800">System Initializing...</h3>
        <p class="mt-1 text-sm text-yellow-700">
          Setting up image processing workers. This may take a moment.
        </p>
        <div class="mt-2 flex items-center">
          <svg class="animate-spin h-4 w-4 text-yellow-600 mr-2" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          <span class="text-sm text-yellow-700">Please wait...</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { SystemStatus } from '@/modules/processing/SystemStatusManager'

const props = defineProps<{
  status: SystemStatus
}>()

const emit = defineEmits<{
  retry: []
}>()

// Local state
const isRetrying = ref(false)
const showDetailedInfo = ref(false)

// Computed properties for environment detection
const hasWebWorkerSupport = computed(() => typeof Worker !== 'undefined')
const hasOffscreenCanvasSupport = computed(() => typeof OffscreenCanvas !== 'undefined')
const hardwareConcurrency = computed(() => navigator.hardwareConcurrency || 'Unknown')
const isFileProtocol = computed(() => location.protocol === 'file:')
const location = computed(() => ({
  protocol: window.location.protocol,
  href: window.location.href
}))

const browserInfo = computed(() => {
  const ua = navigator.userAgent
  if (ua.includes('Chrome')) return 'Chrome'
  if (ua.includes('Firefox')) return 'Firefox'
  if (ua.includes('Safari') && !ua.includes('Chrome')) return 'Safari'
  if (ua.includes('Edge')) return 'Edge'
  return 'Unknown'
})

// Methods
const handleRetry = async () => {
  isRetrying.value = true
  try {
    emit('retry')
    // Wait a bit before allowing another retry
    await new Promise(resolve => setTimeout(resolve, 2000))
  } finally {
    isRetrying.value = false
  }
}

const forceEnableFallback = () => {
  console.warn('🔧 User requested fallback mode')
  // Force the system to use fallback mode
  if (typeof window !== 'undefined' && (window as any).__FORCE_FALLBACK) {
    (window as any).__FORCE_FALLBACK()
  } else {
    // Reload the page with a flag to force fallback
    const url = new URL(window.location.href)
    url.searchParams.set('fallback', 'true')
    window.location.href = url.toString()
  }
}
</script>