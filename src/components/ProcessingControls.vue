<!-- src/components/ProcessingControls.vue -->
<template>
  <div class="card">
    <h2 class="text-xl font-semibold text-slate-800 mb-4">Processing Controls</h2>

    <!-- System Status Alert -->
    <SystemStatusAlert 
      :status="systemStatus"
      @retry="retryInitialization"
    />

    <!-- Worker Status Display -->
    <WorkerStatusDisplay 
      v-if="!systemStatus.error"
      :status="systemStatus"
    />

    <!-- Algorithm Selection -->
    <div class="space-y-6">
      <!-- Processing Type Selection -->
      <ProcessingTypeSelector
        v-model="selectedType"
        :disabled="!hasImage || isProcessing || !systemStatus.initialized"
      />

      <!-- Parameter Controls -->
      <ParameterControls
        v-if="selectedType"
        :processing-type="selectedType"
        :parameters="currentParameters"
        @update:parameters="updateParameters"
      />

      <!-- Action Buttons -->
      <ProcessingActions
        :can-process="canProcess"
        :is-processing="isProcessing"
        @preview="processPreview"
        @process="processFullSize"
      />

      <!-- Quick Presets -->
      <QuickPresets
        :has-image="hasImage"
        :system-initialized="systemStatus.initialized"
        @apply-preset="applyPreset"
      />

      <!-- Algorithm Info -->
      <AlgorithmInfo 
        v-if="selectedType"
        :processing-type="selectedType"
      />

      <!-- Debug Information -->
      <DebugPanel
        v-if="showDebugInfo"
        :system-status="systemStatus"
        :has-image="hasImage"
        :can-process="canProcess"
        :selected-type="selectedType"
        :active-tasks="activeTasks"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { WorkerOrchestratorModule } from '@/modules/WorkerOrchestratorModule'
import { ProcessingModule } from '@/modules/ProcessingModule'
import { SystemStatusManager } from '@/modules/processing/SystemStatusManager'
import { ParameterValidator } from '@/modules/processing/ParameterValidator'
import { PROCESSING_PRESETS } from '@/modules/processing/presets'
import type { 
  ProcessingType, 
  ProcessingParameters 
} from '@/types'

// Import child components
import SystemStatusAlert from './processing/SystemStatusAlert.vue'
import WorkerStatusDisplay from './processing/WorkerStatusDisplay.vue'
import ProcessingTypeSelector from './processing/ProcessingTypeSelector.vue'
import ParameterControls from './processing/ParameterControls.vue'
import ProcessingActions from './processing/ProcessingActions.vue'
import QuickPresets from './processing/QuickPresets.vue'
import AlgorithmInfo from './processing/AlgorithmInfo.vue'
import DebugPanel from './processing/DebugPanel.vue'

// Store
const appStore = useAppStore()

// Modules
let workerOrchestrator: WorkerOrchestratorModule | null = null
let processingModule: ProcessingModule | null = null
const systemStatusManager = SystemStatusManager.getInstance()

// System status
const systemStatus = ref(systemStatusManager.getStatus())

// State
const selectedType = ref<ProcessingType | ''>('')
const currentParameters = ref<ProcessingParameters>({})

// Debug mode
const showDebugInfo = ref(process.env.NODE_ENV === 'development')

// Update system status when it changes
let statusUpdateInterval: number | null = null

// Computed
const hasImage = computed(() => appStore.hasImage)
const isProcessing = computed(() => appStore.isProcessing)
const activeTasks = computed(() => appStore.activeTasks)
const canProcess = computed(() => 
  hasImage.value && 
  selectedType.value !== '' && 
  systemStatus.value.initialized && 
  !systemStatus.value.error
)

// Methods
const updateParameters = (newParams: ProcessingParameters) => {
  currentParameters.value = newParams
}

const processPreview = async () => {
  if (!canProcess.value || !appStore.currentImage || !workerOrchestrator) {
    console.warn('❌ Cannot process: missing requirements')
    return
  }

  try {
    console.group('🔍 Starting Preview Processing')
    
    // Validate parameters
    const validation = ParameterValidator.validate(selectedType.value as ProcessingType, currentParameters.value)
    if (!validation.isValid) {
      console.error('❌ Parameter validation failed:', validation.errors)
      return
    }

    if (validation.warnings.length > 0) {
      console.warn('⚠️ Parameter warnings:', validation.warnings)
    }
    
    const taskId = await workerOrchestrator.submitTask(
      appStore.currentImage,
      selectedType.value as ProcessingType,
      currentParameters.value
    )
    
    console.log('✅ Preview processing task submitted:', taskId)
    
    // Add task to store
    const task = appStore.addTask(selectedType.value as ProcessingType, currentParameters.value)
    console.log('✅ Task added to store:', task.id)
    console.groupEnd()
    
  } catch (error) {
    console.error('❌ Failed to start preview processing:', error)
    systemStatusManager.setInitializationError(error instanceof Error ? error.message : 'Unknown error')
  }
}

const processFullSize = async () => {
  if (!canProcess.value || !appStore.currentImage || !workerOrchestrator) {
    console.warn('❌ Cannot process: missing requirements')
    return
  }

  try {
    console.group('🔧 Starting Full-Size Processing')
    
    // Validate parameters
    const validation = ParameterValidator.validate(selectedType.value as ProcessingType, currentParameters.value)
    if (!validation.isValid) {
      console.error('❌ Parameter validation failed:', validation.errors)
      return
    }

    if (validation.warnings.length > 0) {
      console.warn('⚠️ Parameter warnings:', validation.warnings)
    }
    
    const taskId = await workerOrchestrator.submitTask(
      appStore.currentImage,
      selectedType.value as ProcessingType,
      currentParameters.value
    )
    
    console.log('✅ Full-size processing task submitted:', taskId)
    
    // Add task to store
    const task = appStore.addTask(selectedType.value as ProcessingType, currentParameters.value)
    console.log('✅ Task added to store:', task.id)
    console.groupEnd()
    
  } catch (error) {
    console.error('❌ Failed to start full-size processing:', error)
    systemStatusManager.setInitializationError(error instanceof Error ? error.message : 'Unknown error')
  }
}

const applyPreset = (presetName: string) => {
  const preset = PROCESSING_PRESETS[presetName]
  if (!preset) {
    console.warn('⚠️ Unknown preset:', presetName)
    return
  }

  console.log('🎯 Applying preset:', presetName)
  selectedType.value = preset.type
  currentParameters.value = { ...preset.parameters }
}

const retryInitialization = async () => {
  if (systemStatusManager.isMaxAttemptsReached()) {
    console.error('❌ Maximum initialization attempts reached')
    return
  }

  const attempt = systemStatusManager.incrementInitializationAttempts()
  systemStatusManager.clearInitializationError()
  console.log(`🔄 Retrying initialization (attempt ${attempt}/${systemStatusManager.getMaxAttempts()})`)
  
  await initializeWorkers()
}

const initializeWorkers = async () => {
  try {
    console.group('🚀 Initializing Processing System')
    
    workerOrchestrator = WorkerOrchestratorModule.getInstance()
    await workerOrchestrator.initialize()
    console.log('✅ Workers initialized successfully')
    
    // Initialize processing module
    processingModule = ProcessingModule.getInstance()
    await processingModule.initialize()
    console.log('✅ Processing module initialized')
    
    // Set up task update callback
    workerOrchestrator.setTaskUpdateCallback((task) => {
      console.log('📝 Task update received:', task.id, task.status, `${task.progress}%`)
      appStore.updateTask(task.id, task)
      
      // If task completed successfully, update processed image
      if (task.status === 'completed' && task.result) {
        console.log('✅ Task completed, updating processed image')
        const processedImageData = {
          data: task.result,
          width: appStore.currentImage?.width || 0,
          height: appStore.currentImage?.height || 0,
          channels: 4,
          format: appStore.currentImage?.format || 'PNG',
          filename: `processed_${appStore.currentImage?.filename || 'image'}`,
          size: task.result.byteLength
        }
        appStore.setProcessedImage(processedImageData)
      }
    })
    
    // Start status updates
    statusUpdateInterval = window.setInterval(updateSystemStatus, 1000)
    updateSystemStatus()
    
    systemStatusManager.markInitialized()
    console.log('✅ Processing system fully initialized')
    console.groupEnd()
    
  } catch (error) {
    console.error('❌ Failed to initialize processing system:', error)
    systemStatusManager.setInitializationError(error instanceof Error ? error.message : 'Initialization failed')
    console.groupEnd()
  }
}

const updateSystemStatus = () => {
  if (workerOrchestrator) {
    const status = workerOrchestrator.getWorkerStatus()
    systemStatusManager.updateStatus(status)
    systemStatus.value = systemStatusManager.getStatus()
  }
}

// Set up status update callback
systemStatusManager.addStatusUpdateCallback((status) => {
  systemStatus.value = status
})

// Initialize on component mount
onMounted(() => {
  console.log('🎛️ ProcessingControls component mounted, initializing...')
  initializeWorkers()
})

// Cleanup on component unmount
onUnmounted(() => {
  console.log('🎛️ ProcessingControls component unmounting, cleaning up...')
  
  if (statusUpdateInterval) {
    clearInterval(statusUpdateInterval)
  }
  
  // Remove status update callback
  systemStatusManager.removeStatusUpdateCallback((status) => {
    systemStatus.value = status
  })
  
  // Don't destroy the orchestrator as it might be used by other components
  workerOrchestrator = null
  processingModule = null
})
</script>