<template>
  <div class="card">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-xl font-semibold text-slate-800">Processing Status</h2>
      
      <button
        v-if="completedTasks.length > 0"
        @click="clearCompleted"
        class="btn btn-secondary text-sm"
      >
        Clear Completed
      </button>
    </div>

    <!-- Current Task -->
    <div v-if="currentTask" class="mb-6">
      <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center">
            <div class="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
            <span class="font-medium text-blue-800">{{ formatTaskType(currentTask.type) }}</span>
          </div>
          <button
            @click="cancelTask(currentTask.id)"
            class="text-blue-600 hover:text-blue-800 text-sm"
          >
            Cancel
          </button>
        </div>
        
        <!-- Progress Bar -->
        <div class="w-full bg-blue-200 rounded-full h-2 mb-2">
          <div
            class="bg-blue-600 h-2 rounded-full transition-all duration-300"
            :style="{ width: `${currentTask.progress}%` }"
          ></div>
        </div>
        
        <div class="flex items-center justify-between text-sm text-blue-700">
          <span>{{ currentTask.progress }}% complete</span>
          <span>{{ formatDuration(Date.now() - currentTask.createdAt.getTime()) }}</span>
        </div>
      </div>
    </div>

    <!-- Task Queue -->
    <div v-if="queuedTasks.length > 0" class="mb-6">
      <h3 class="text-sm font-medium text-slate-700 mb-3">Queued Tasks ({{ queuedTasks.length }})</h3>
      <div class="space-y-2">
        <div
          v-for="task in queuedTasks"
          :key="task.id"
          class="bg-slate-50 border border-slate-200 rounded-lg p-3"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center">
              <div class="w-2 h-2 bg-slate-400 rounded-full mr-2"></div>
              <span class="text-sm font-medium text-slate-700">{{ formatTaskType(task.type) }}</span>
            </div>
            <button
              @click="cancelTask(task.id)"
              class="text-slate-500 hover:text-slate-700 text-sm"
            >
              Cancel
            </button>
          </div>
          <div class="text-xs text-slate-500 mt-1">
            Queued {{ formatDuration(Date.now() - task.createdAt.getTime()) }} ago
          </div>
        </div>
      </div>
    </div>

    <!-- Completed Tasks -->
    <div v-if="completedTasks.length > 0" class="mb-6">
      <h3 class="text-sm font-medium text-slate-700 mb-3">Recent Completed ({{ completedTasks.length }})</h3>
      <div class="space-y-2 max-h-64 overflow-y-auto">
        <div
          v-for="task in completedTasks.slice(-5)"
          :key="task.id"
          class="border rounded-lg p-3"
          :class="getTaskStatusClass(task.status)"
        >
          <div class="flex items-center justify-between">
            <div class="flex items-center">
              <div
                class="w-2 h-2 rounded-full mr-2"
                :class="getTaskStatusIndicator(task.status)"
              ></div>
              <span class="text-sm font-medium">{{ formatTaskType(task.type) }}</span>
            </div>
            <div class="flex items-center space-x-2">
              <span class="text-xs text-slate-500">
                {{ formatDuration(task.completedAt!.getTime() - task.createdAt.getTime()) }}
              </span>
              <button
                v-if="task.status === 'completed'"
                @click="downloadResult(task)"
                class="text-green-600 hover:text-green-800 text-sm"
                title="Download result"
              >
                ⬇️
              </button>
            </div>
          </div>
          
          <!-- Task Details -->
          <div class="mt-2 text-xs text-slate-600">
            <div v-if="task.status === 'completed'" class="text-green-700">
              ✅ Completed successfully
            </div>
            <div v-else-if="task.status === 'failed'" class="text-red-700">
              ❌ {{ task.error || 'Processing failed' }}
            </div>
            <div v-else-if="task.status === 'cancelled'" class="text-slate-600">
              🚫 Cancelled by user
            </div>
          </div>
          
          <!-- Parameters Summary -->
          <div class="mt-1 text-xs text-slate-500">
            {{ getParametersSummary(task) }}
          </div>
        </div>
      </div>
    </div>

    <!-- Worker Status -->
    <div class="bg-slate-50 border border-slate-200 rounded-lg p-4">
      <h3 class="text-sm font-medium text-slate-700 mb-3">System Status</h3>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-slate-500">Active Workers:</span>
          <span class="ml-2 font-medium">{{ workerStatus.activeWorkers }}</span>
        </div>
        <div>
          <span class="text-slate-500">Available:</span>
          <span class="ml-2 font-medium">{{ workerStatus.availableWorkers }}</span>
        </div>
        <div>
          <span class="text-slate-500">Queue:</span>
          <span class="ml-2 font-medium">{{ workerStatus.queuedTasks }}</span>
        </div>
        <div>
          <span class="text-slate-500">Memory:</span>
          <span class="ml-2 font-medium">{{ formatMemory() }}</span>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-if="!hasAnyTasks" class="text-center py-8">
      <svg class="mx-auto h-12 w-12 text-slate-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-3 7h3m-3 4h3m-6-4h.01M9 16h.01" />
      </svg>
      <p class="text-slate-500">No processing tasks</p>
      <p class="text-slate-400 text-sm mt-1">Upload an image and start processing to see progress here</p>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { computed, ref, onMounted, onUnmounted } from 'vue'
  import { useAppStore } from '@/stores/app'
  import { WorkerOrchestratorModule } from '@/modules/WorkerOrchestratorModule'
  import { downloadImage } from '@/utils/imageHelpers'
  import type { ProcessingTask, ProcessingType } from '@/types'

  // Store
  const appStore = useAppStore()

  // Worker orchestrator
  const workerOrchestrator = WorkerOrchestratorModule.getInstance()

  // Refs
  const workerStatus = ref({
    totalWorkers: 0,
    availableWorkers: 0,
    activeWorkers: 0,
    queuedTasks: 0,
    activeTasks: 0
  })

  // Computed
  const activeTasks = computed(() => appStore.activeTasks)
  const currentTask = computed(() => appStore.currentTask)
  const completedTasks = computed(() => appStore.completedTasks)
  
  const queuedTasks = computed(() => 
    activeTasks.value.filter(task => task.status === 'pending')
  )

  const hasAnyTasks = computed(() => activeTasks.value.length > 0)

  // Methods
  const cancelTask = (taskId: string) => {
    workerOrchestrator.cancelTask(taskId)
    appStore.cancelTask(taskId)
  }

  const clearCompleted = () => {
    appStore.clearCompletedTasks()
  }

  const downloadResult = (task: ProcessingTask) => {
    if (!task.result) return

    const filename = `${task.type}_${Date.now()}`
    downloadImage(task.result, filename, 'PNG')
  }

  const formatTaskType = (type: ProcessingType): string => {
    const typeMap: Record<ProcessingType, string> = {
      'binarization': 'Binarization',
      'morphology': 'Morphological Ops',
      'noise-reduction': 'Noise Reduction',
      'scaling': 'Image Scaling'
    }
    return typeMap[type] || type
  }

  const formatDuration = (ms: number): string => {
    const seconds = Math.floor(ms / 1000)
    const minutes = Math.floor(seconds / 60)
    
    if (minutes > 0) {
      return `${minutes}m ${seconds % 60}s`
    }
    return `${seconds}s`
  }

  const getTaskStatusClass = (status: string): string => {
    switch (status) {
      case 'completed':
        return 'bg-green-50 border-green-200'
      case 'failed':
        return 'bg-red-50 border-red-200'
      case 'cancelled':
        return 'bg-slate-50 border-slate-200'
      default:
        return 'bg-blue-50 border-blue-200'
    }
  }

  const getTaskStatusIndicator = (status: string): string => {
    switch (status) {
      case 'completed':
        return 'bg-green-500'
      case 'failed':
        return 'bg-red-500'
      case 'cancelled':
        return 'bg-slate-500'
      default:
        return 'bg-blue-500'
    }
  }

  const getParametersSummary = (task: ProcessingTask): string => {
    const params = task.parameters
    
    if (params.binarization) {
      return `Method: ${params.binarization.method}, Window: ${params.binarization.windowSize || 'N/A'}`
    }
    
    if (params.morphology) {
      return `Operation: ${params.morphology.operation}, Kernel: ${params.morphology.kernelSize}x${params.morphology.kernelSize}`
    }
    
    if (params.noise) {
      return `Method: ${params.noise.method}, Kernel: ${params.noise.kernelSize || 'N/A'}`
    }
    
    if (params.scaling) {
      return `Method: ${params.scaling.method}, Factor: ${params.scaling.factor}x`
    }
    
    return 'Default parameters'
  }

  const formatMemory = (): string => {
    // Use optional chaining and type assertion for performance.memory
    const memory = (performance as any).memory
    if (!memory) return 'N/A'
    
    const used = Math.round(memory.usedJSHeapSize / 1024 / 1024)
    const total = Math.round(memory.totalJSHeapSize / 1024 / 1024)
    
    return `${used}/${total} MB`
  }

  const updateWorkerStatus = () => {
    workerStatus.value = workerOrchestrator.getWorkerStatus()
  }

  // Update worker status periodically
  let statusInterval: number | null = null

  onMounted(() => {
    updateWorkerStatus()
    statusInterval = window.setInterval(updateWorkerStatus, 1000)
  })

  onUnmounted(() => {
    if (statusInterval) {
      clearInterval(statusInterval)
    }
  })
</script>

<style scoped>
  .task-item {
    transition: all 0.2s ease-in-out;
  }

  .task-item:hover {
    transform: translateY(-1px);
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  /* Custom scrollbar for completed tasks */
  .max-h-64::-webkit-scrollbar {
    width: 6px;
  }

  .max-h-64::-webkit-scrollbar-track {
    background: #f1f5f9;
    border-radius: 3px;
  }

  .max-h-64::-webkit-scrollbar-thumb {
    background: #cbd5e1;
    border-radius: 3px;
  }

  .max-h-64::-webkit-scrollbar-thumb:hover {
    background: #94a3b8;
  }
</style>