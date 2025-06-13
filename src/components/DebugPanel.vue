<!-- src/components/DebugPanel.vue -->
<template>
  <div v-if="showDebug" class="fixed bottom-4 right-4 w-96 max-h-96 bg-black text-green-400 font-mono text-xs rounded-lg shadow-xl z-50 overflow-hidden">
    <div class="bg-gray-800 p-2 flex items-center justify-between">
      <span class="text-white font-bold">🔧 Debug Console</span>
      <div class="flex space-x-2">
        <button @click="runDiagnostics" class="px-2 py-1 bg-blue-600 text-white rounded text-xs hover:bg-blue-700">
          Diagnose
        </button>
        <button @click="exportLogs" class="px-2 py-1 bg-green-600 text-white rounded text-xs hover:bg-green-700">
          Export
        </button>
        <button @click="clearLogs" class="px-2 py-1 bg-red-600 text-white rounded text-xs hover:bg-red-700">
          Clear
        </button>
        <button @click="showDebug = false" class="px-2 py-1 bg-gray-600 text-white rounded text-xs hover:bg-gray-700">
          ×
        </button>
      </div>
    </div>
    
    <div class="p-2 overflow-y-auto max-h-80">
      <!-- Live stats -->
      <div class="mb-2 p-2 bg-gray-900 rounded">
        <div class="text-yellow-400 mb-1">System Status:</div>
        <div>Workers: {{ workerStatus.availableWorkers }}/{{ workerStatus.totalWorkers }}</div>
        <div>Tasks: {{ workerStatus.activeTasks }} active, {{ workerStatus.queuedTasks }} queued</div>
        <div>Memory: {{ memoryUsage }}</div>
        <div>Status: <span :class="getStatusColor(systemStatus)">{{ systemStatus }}</span></div>
      </div>

      <!-- Log entries -->
      <div class="space-y-1">
        <div
          v-for="event in recentEvents"
          :key="event.timestamp"
          :class="getLogClass(event.level)"
          class="p-1 rounded text-xs"
        >
          <div class="flex items-start space-x-2">
            <span class="text-gray-400 whitespace-nowrap">
              {{ formatTime(event.timestamp) }}
            </span>
            <span class="uppercase text-xs px-1 rounded" :class="getCategoryClass(event.category)">
              {{ event.category }}
            </span>
            <span class="flex-1">{{ event.message }}</span>
          </div>
          <div v-if="event.data" class="ml-20 text-gray-500 text-xs truncate" :title="JSON.stringify(event.data)">
            {{ formatData(event.data) }}
          </div>
        </div>
      </div>

      <!-- Empty state -->
      <div v-if="recentEvents.length === 0" class="text-gray-500 text-center py-4">
        No debug events yet...
      </div>
    </div>
  </div>

  <!-- Debug toggle button -->
  <button
    v-if="!showDebug"
    @click="showDebug = true"
    class="fixed bottom-4 right-4 w-12 h-12 bg-black text-green-400 rounded-full shadow-xl z-50 flex items-center justify-center hover:bg-gray-800 transition-colors"
    title="Open Debug Console"
  >
    🔧
  </button>

  <!-- Diagnostic overlay -->
  <div v-if="showDiagnostics" class="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50">
    <div class="bg-white rounded-lg p-6 max-w-4xl max-h-96 overflow-auto">
      <div class="flex items-center justify-between mb-4">
        <h2 class="text-xl font-bold">System Diagnostics</h2>
        <button @click="showDiagnostics = false" class="text-gray-500 hover:text-gray-700">×</button>
      </div>
      
      <div v-if="diagnosticData" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <h3 class="font-semibold mb-2">Environment</h3>
            <pre class="text-xs bg-gray-100 p-2 rounded">{{ JSON.stringify(diagnosticData.environment, null, 2) }}</pre>
          </div>
          <div>
            <h3 class="font-semibold mb-2">Browser Support</h3>
            <div class="space-y-1 text-sm">
              <div :class="diagnosticData.webWorkerSupport ? 'text-green-600' : 'text-red-600'">
                ✓ Web Workers: {{ diagnosticData.webWorkerSupport ? 'Supported' : 'Not Supported' }}
              </div>
              <div :class="diagnosticData.offscreenCanvasSupport ? 'text-green-600' : 'text-red-600'">
                ✓ OffscreenCanvas: {{ diagnosticData.offscreenCanvasSupport ? 'Supported' : 'Not Supported' }}
              </div>
              <div :class="diagnosticData.wasmSupport ? 'text-green-600' : 'text-red-600'">
                ✓ WebAssembly: {{ diagnosticData.wasmSupport ? 'Supported' : 'Not Supported' }}
              </div>
            </div>
          </div>
        </div>

        <div>
          <h3 class="font-semibold mb-2">Worker URLs Test Results</h3>
          <div class="space-y-1 text-xs">
            <div v-for="(result, url) in diagnosticData.workerUrls" :key="url" class="flex items-center space-x-2">
              <span :class="result.accessible ? 'text-green-600' : 'text-red-600'">
                {{ result.accessible ? '✓' : '✗' }}
              </span>
              <span class="font-mono">{{ url }}</span>
              <span class="text-gray-500">
                {{ result.status || result.error }}
              </span>
            </div>
          </div>
        </div>

        <div v-if="diagnosticData.headers">
          <h3 class="font-semibold mb-2">Cross-Origin Headers</h3>
          <pre class="text-xs bg-gray-100 p-2 rounded">{{ JSON.stringify(diagnosticData.headers, null, 2) }}</pre>
        </div>
      </div>

      <div v-else class="text-center py-8">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
        <p>Running diagnostics...</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { debugLogger, type DebugEvent } from '@/utils/debugLogger'
import { useAppStore } from '@/stores/app'
import { WorkerOrchestratorModule } from '@/modules/WorkerOrchestratorModule'

const showDebug = ref(false)
const showDiagnostics = ref(false)
const diagnosticData = ref<any>(null)
const events = ref<DebugEvent[]>([])

const appStore = useAppStore()

// Get recent events (last 50)
const recentEvents = computed(() => events.value.slice(-50).reverse())

// System status
const systemStatus = computed(() => {
  if (!appStore.isInitialized) return 'initializing'
  if (appStore.isProcessing) return 'processing'
  return 'ready'
})

// Worker status
const workerStatus = computed(() => {
  try {
    const orchestrator = WorkerOrchestratorModule.getInstance()
    return orchestrator.getWorkerStatus()
  } catch {
    return {
      totalWorkers: 0,
      availableWorkers: 0,
      activeWorkers: 0,
      queuedTasks: 0,
      activeTasks: 0
    }
  }
})

// Memory usage
const memoryUsage = computed(() => {
  const performance = globalThis.performance as any
  if (performance?.memory) {
    const used = Math.round(performance.memory.usedJSHeapSize / 1024 / 1024)
    const total = Math.round(performance.memory.totalJSHeapSize / 1024 / 1024)
    return `${used}/${total}MB`
  }
  return 'N/A'
})

// Event listener for debug events
const onDebugEvent = (event: DebugEvent) => {
  events.value.push(event)
  if (events.value.length > 200) {
    events.value.shift()
  }
}

// Methods
const runDiagnostics = async () => {
  showDiagnostics.value = true
  diagnosticData.value = null
  
  try {
    diagnosticData.value = await debugLogger.diagnoseWorkerSupport()
  } catch (error) {
    console.error('Diagnostics failed:', error)
    diagnosticData.value = { error: 'Diagnostics failed' }
  }
}

const exportLogs = () => {
  const logs = debugLogger.export()
  const blob = new Blob([logs], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  
  const link = document.createElement('a')
  link.href = url
  link.download = `debug-logs-${new Date().toISOString().slice(0, 19)}.json`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  URL.revokeObjectURL(url)
}

const clearLogs = () => {
  debugLogger.clear()
  events.value = []
}

const formatTime = (timestamp: string) => {
  return new Date(timestamp).toLocaleTimeString().slice(-8)
}

const formatData = (data: any) => {
  if (typeof data === 'object') {
    return JSON.stringify(data).slice(0, 100) + (JSON.stringify(data).length > 100 ? '...' : '')
  }
  return String(data).slice(0, 100)
}

const getLogClass = (level: string) => {
  switch (level) {
    case 'error': return 'bg-red-900 text-red-300'
    case 'warn': return 'bg-yellow-900 text-yellow-300'
    case 'debug': return 'bg-blue-900 text-blue-300'
    default: return 'bg-gray-800 text-gray-300'
  }
}

const getCategoryClass = (category: string) => {
  switch (category) {
    case 'worker': return 'bg-purple-600 text-white'
    case 'processing': return 'bg-blue-600 text-white'
    case 'error': return 'bg-red-600 text-white'
    case 'diagnostics': return 'bg-green-600 text-white'
    default: return 'bg-gray-600 text-white'
  }
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'ready': return 'text-green-400'
    case 'processing': return 'text-blue-400'
    case 'initializing': return 'text-yellow-400'
    default: return 'text-red-400'
  }
}

// Lifecycle
onMounted(() => {
  debugLogger.addListener(onDebugEvent)
  events.value = debugLogger.getEvents()
  
  // Show debug panel in development
  if (import.meta.env.DEV) {
    setTimeout(() => {
      showDebug.value = true
    }, 1000)
  }
})

onUnmounted(() => {
  debugLogger.removeListener(onDebugEvent)
})

// Global hotkey to toggle debug (Ctrl+Shift+D)
onMounted(() => {
  const handleKeydown = (event: KeyboardEvent) => {
    if (event.ctrlKey && event.shiftKey && event.key === 'D') {
      event.preventDefault()
      showDebug.value = !showDebug.value
    }
  }
  
  window.addEventListener('keydown', handleKeydown)
  
  onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown)
  })
})
</script>