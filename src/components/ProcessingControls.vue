<!-- src/components/ProcessingControls.vue -->
<template>
  <div class="card">
    <h2 class="text-xl font-semibold text-slate-800 mb-4">Processing Controls</h2>

    <!-- System Status Alert -->
    <div v-if="systemStatus.error" class="mb-4 bg-red-50 border border-red-200 rounded-lg p-4">
      <div class="flex">
        <svg class="h-5 w-5 text-red-400 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
        </svg>
        <div class="ml-3">
          <h3 class="text-sm font-medium text-red-800">System Error</h3>
          <p class="mt-1 text-sm text-red-700">{{ systemStatus.error }}</p>
          <button 
            class="mt-2 text-sm text-red-600 hover:text-red-800 underline"
            @click="retryInitialization"
          >
            Retry Initialization
          </button>
        </div>
      </div>
    </div>

    <!-- Worker Status -->
    <div v-if="!systemStatus.error" class="mb-4 bg-slate-50 border border-slate-200 rounded-lg p-3">
      <h3 class="text-sm font-medium text-slate-700 mb-2">System Status</h3>
      <div class="grid grid-cols-2 gap-3 text-xs">
        <div class="flex items-center">
          <div 
            class="w-2 h-2 rounded-full mr-2"
            :class="systemStatus.initialized ? 'bg-green-500' : 'bg-red-500'"
          ></div>
          <span>{{ systemStatus.initialized ? 'Ready' : 'Initializing...' }}</span>
        </div>
        <div>
          <span class="text-slate-500">Workers:</span>
          <span class="ml-1 font-medium">{{ systemStatus.availableWorkers }}/{{ systemStatus.totalWorkers }}</span>
        </div>
        <div>
          <span class="text-slate-500">Queue:</span>
          <span class="ml-1 font-medium">{{ systemStatus.queuedTasks }}</span>
        </div>
        <div>
          <span class="text-slate-500">Environment:</span>
          <span class="ml-1 font-medium text-xs">{{ systemStatus.environment }}</span>
        </div>
      </div>
    </div>

    <!-- Algorithm Selection -->
    <div class="space-y-6">
      <!-- Processing Type -->
      <div>
        <label class="label">Processing Type</label>
        <select
          v-model="selectedType"
          class="input"
          :disabled="!hasImage || isProcessing || !systemStatus.initialized"
        >
          <option value="">Select algorithm...</option>
          <option value="binarization">Binarization</option>
          <option value="morphology">Morphological Operations</option>
          <option value="noise-reduction">Noise Reduction</option>
          <option value="scaling">Pixel Art Scaling</option>
        </select>
      </div>

      <!-- Binarization Controls -->
      <div v-if="selectedType === 'binarization'" class="space-y-4">
        <div>
          <label class="label">Method</label>
          <select v-model="binarizationParams.method" class="input">
            <option value="otsu">Otsu (Global)</option>
            <option value="sauvola">Sauvola (Adaptive)</option>
            <option value="niblack">Niblack (Adaptive)</option>
          </select>
        </div>

        <div v-if="binarizationParams.method === 'sauvola' || binarizationParams.method === 'niblack'">
          <label class="label">
            Window Size: {{ binarizationParams.windowSize }}
          </label>
          <input
            v-model.number="binarizationParams.windowSize"
            type="range"
            min="3"
            max="51"
            step="2"
            class="slider"
          />
        </div>

        <div v-if="binarizationParams.method === 'sauvola' || binarizationParams.method === 'niblack'">
          <label class="label">
            K Factor: {{ (binarizationParams.k ?? 0.2).toFixed(2) }}
          </label>
          <input
            v-model.number="binarizationParams.k"
            type="range"
            min="-1"
            max="1"
            step="0.01"
            class="slider"
          />
        </div>

        <div v-if="binarizationParams.method === 'otsu'">
          <label class="label">
            Threshold: {{ binarizationParams.threshold }}
          </label>
          <input
            v-model.number="binarizationParams.threshold"
            type="range"
            min="0"
            max="255"
            step="1"
            class="slider"
          />
        </div>
      </div>

      <!-- Morphology Controls -->
      <div v-if="selectedType === 'morphology'" class="space-y-4">
        <div>
          <label class="label">Operation</label>
          <select v-model="morphologyParams.operation" class="input">
            <option value="opening">Opening</option>
            <option value="closing">Closing</option>
            <option value="erosion">Erosion</option>
            <option value="dilation">Dilation</option>
          </select>
        </div>

        <div>
          <label class="label">
            Kernel Size: {{ morphologyParams.kernelSize }}
          </label>
          <input
            v-model.number="morphologyParams.kernelSize"
            type="range"
            min="3"
            max="15"
            step="2"
            class="slider"
          />
        </div>

        <div>
          <label class="label">
            Iterations: {{ morphologyParams.iterations }}
          </label>
          <input
            v-model.number="morphologyParams.iterations"
            type="range"
            min="1"
            max="5"
            step="1"
            class="slider"
          />
        </div>
      </div>

      <!-- Noise Reduction Controls -->
      <div v-if="selectedType === 'noise-reduction'" class="space-y-4">
        <div>
          <label class="label">Method</label>
          <select v-model="noiseParams.method" class="input">
            <option value="median">Median Filter</option>
            <option value="binary-noise-removal">Binary Noise Removal</option>
          </select>
        </div>

        <div v-if="noiseParams.method === 'median'">
          <label class="label">
            Kernel Size: {{ noiseParams.kernelSize }}
          </label>
          <input
            v-model.number="noiseParams.kernelSize"
            type="range"
            min="3"
            max="9"
            step="2"
            class="slider"
          />
        </div>

        <div v-if="noiseParams.method === 'binary-noise-removal'">
          <label class="label">
            Minimum Component Size: {{ noiseParams.threshold }}
          </label>
          <input
            v-model.number="noiseParams.threshold"
            type="range"
            min="10"
            max="500"
            step="10"
            class="slider"
          />
        </div>
      </div>

      <!-- Scaling Controls -->
      <div v-if="selectedType === 'scaling'" class="space-y-4">
        <div>
          <label class="label">Method</label>
          <select v-model="scalingParams.method" class="input">
            <option value="scale2x">Scale2x (Pixel Art)</option>
            <option value="scale3x">Scale3x (Pixel Art)</option>
            <option value="scale4x">Scale4x (Pixel Art)</option>
            <option value="nearest">Nearest Neighbor</option>
            <option value="bilinear">Bilinear</option>
          </select>
        </div>

        <div v-if="!scalingParams.method.startsWith('scale')">
          <label class="label">
            Scale Factor: {{ scalingParams.factor.toFixed(1) }}x
          </label>
          <input
            v-model.number="scalingParams.factor"
            type="range"
            min="0.5"
            max="4"
            step="0.1"
            class="slider"
          />
        </div>
      </div>

      <!-- Action Buttons -->
      <div class="flex space-x-3 pt-4 border-t border-slate-200">
        <button
          class="btn btn-secondary flex-1"
          :disabled="!canProcess || isProcessing"
          @click="processPreview"
        >
          <svg v-if="isProcessing" class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ isProcessing ? 'Processing...' : 'Preview' }}
        </button>
        
        <button
          class="btn btn-primary flex-1"
          :disabled="!canProcess || isProcessing"
          @click="processFullSize"
        >
          Process Full Size
        </button>
      </div>

      <!-- Quick Presets -->
      <div class="pt-4 border-t border-slate-200">
        <h3 class="text-sm font-medium text-slate-700 mb-3">Quick Presets</h3>
        <div class="grid grid-cols-2 gap-2">
          <button
            class="btn btn-secondary text-sm"
            :disabled="!hasImage || !systemStatus.initialized"
            @click="applyPreset('document')"
          >
            📄 Document
          </button>
          <button
            class="btn btn-secondary text-sm"
            :disabled="!hasImage || !systemStatus.initialized"
            @click="applyPreset('engraving')"
          >
            🖼️ Engraving
          </button>
          <button
            class="btn btn-secondary text-sm"
            :disabled="!hasImage || !systemStatus.initialized"
            @click="applyPreset('pixel-art')"
          >
            🎮 Pixel Art
          </button>
          <button
            class="btn btn-secondary text-sm"
            :disabled="!hasImage || !systemStatus.initialized"
            @click="applyPreset('noise-clean')"
          >
            ✨ Noise Clean
          </button>
        </div>
      </div>

      <!-- Parameter Info -->
      <div v-if="selectedType" class="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-blue-800 mb-2">Algorithm Info</h3>
        <p class="text-sm text-blue-700">{{ getAlgorithmDescription() }}</p>
      </div>

      <!-- Debug Information (only in development) -->
      <div v-if="showDebugInfo" class="bg-gray-50 border border-gray-200 rounded-lg p-4">
        <h3 class="text-sm font-medium text-gray-800 mb-2">Debug Information</h3>
        <div class="text-xs text-gray-600 space-y-1">
          <div>Worker Status: {{ JSON.stringify(systemStatus) }}</div>
          <div>Has Image: {{ hasImage }}</div>
          <div>Can Process: {{ canProcess }}</div>
          <div>Selected Type: {{ selectedType }}</div>
          <div>Active Tasks: {{ activeTasks.length }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
  import { useAppStore } from '@/stores/app'
  import { WorkerOrchestratorModule } from '@/modules/WorkerOrchestratorModule'
  import { ProcessingModule } from '@/modules/ProcessingModule'
  import { PROCESSING_PRESETS } from '@/modules/processing/presets'
  import type { 
    ProcessingType, 
    BinarizationParams, 
    MorphologyParams, 
    NoiseReductionParams, 
    ScalingParams,
    ProcessingParameters 
  } from '@/types'

  // Store
  const appStore = useAppStore()

  // Worker orchestrator and processing module
  let workerOrchestrator: WorkerOrchestratorModule | null = null
  let processingModule: ProcessingModule | null = null

  // System status
  const systemStatus = ref({
    initialized: false,
    totalWorkers: 0,
    availableWorkers: 0,
    queuedTasks: 0,
    environment: 'Unknown',
    error: null as string | null
  })

  // Refs
  const selectedType = ref<ProcessingType | ''>('')
  const initializationAttempts = ref(0)
  const maxInitializationAttempts = 3

  // Debug mode (enable in development)
  const showDebugInfo = ref(process.env.NODE_ENV === 'development')

  // Parameter objects
  const binarizationParams = ref<BinarizationParams>({
    method: 'otsu',
    windowSize: 15,
    k: 0.2,
    threshold: 128
  })

  const morphologyParams = ref<MorphologyParams>({
    operation: 'opening',
    kernelSize: 3,
    iterations: 1
  })

  const noiseParams = ref<NoiseReductionParams>({
    method: 'median',
    kernelSize: 3,
    threshold: 50
  })

  const scalingParams = ref<ScalingParams>({
    method: 'scale2x',
    factor: 2
  })

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

  // Update system status periodically
  let statusUpdateInterval: number | null = null

  const updateSystemStatus = () => {
    if (workerOrchestrator) {
      const status = workerOrchestrator.getWorkerStatus()
      systemStatus.value = {
        initialized: status.initialized,
        totalWorkers: status.totalWorkers,
        availableWorkers: status.availableWorkers,
        queuedTasks: status.queuedTasks,
        environment: status.environment,
        error: status.initializationError
      }
    }
  }

  // Methods
  const buildProcessingParameters = (): ProcessingParameters => {
    const params: ProcessingParameters = {}
    
    switch (selectedType.value) {
      case 'binarization':
        params.binarization = { ...binarizationParams.value }
        break
      case 'morphology':
        params.morphology = { ...morphologyParams.value }
        break
      case 'noise-reduction':
        params.noise = { ...noiseParams.value }
        break
      case 'scaling':
        params.scaling = { ...scalingParams.value }
        // Set factor based on method for pixel art scaling
        if (scalingParams.value.method.startsWith('scale')) {
          const factor = parseInt(scalingParams.value.method.replace('scale', '').replace('x', ''))
          params.scaling!.factor = factor
        }
        break
    }
    
    return params
  }

  const processPreview = async () => {
    if (!canProcess.value || !appStore.currentImage || !workerOrchestrator) {
      console.warn('❌ Cannot process: missing requirements', {
        canProcess: canProcess.value,
        hasImage: !!appStore.currentImage,
        hasOrchestrator: !!workerOrchestrator
      })
      return
    }

    try {
      console.group('🔍 Starting Preview Processing')
      const parameters = buildProcessingParameters()
      console.log('Parameters:', parameters)
      
      const taskId = await workerOrchestrator.submitTask(
        appStore.currentImage,
        selectedType.value as ProcessingType,
        parameters
      )
      
      console.log('✅ Preview processing task submitted:', taskId)
      
      // Add task to store
      const task = appStore.addTask(selectedType.value as ProcessingType, parameters)
      console.log('✅ Task added to store:', task.id)
      console.groupEnd()
      
    } catch (error) {
      console.error('❌ Failed to start preview processing:', error)
      systemStatus.value.error = error instanceof Error ? error.message : 'Unknown error'
    }
  }

  const processFullSize = async () => {
    if (!canProcess.value || !appStore.currentImage || !workerOrchestrator) {
      console.warn('❌ Cannot process: missing requirements', {
        canProcess: canProcess.value,
        hasImage: !!appStore.currentImage,
        hasOrchestrator: !!workerOrchestrator
      })
      return
    }

    try {
      console.group('🔧 Starting Full-Size Processing')
      const parameters = buildProcessingParameters()
      console.log('Parameters:', parameters)
      
      const taskId = await workerOrchestrator.submitTask(
        appStore.currentImage,
        selectedType.value as ProcessingType,
        parameters
      )
      
      console.log('✅ Full-size processing task submitted:', taskId)
      
      // Add task to store
      const task = appStore.addTask(selectedType.value as ProcessingType, parameters)
      console.log('✅ Task added to store:', task.id)
      console.groupEnd()
      
    } catch (error) {
      console.error('❌ Failed to start full-size processing:', error)
      systemStatus.value.error = error instanceof Error ? error.message : 'Unknown error'
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
    
    switch (preset.type) {
      case 'binarization':
        if (preset.parameters.binarization) {
          binarizationParams.value = { ...preset.parameters.binarization }
        }
        break
      case 'morphology':
        if (preset.parameters.morphology) {
          morphologyParams.value = { ...preset.parameters.morphology }
        }
        break
      case 'noise-reduction':
        if (preset.parameters.noise) {
          noiseParams.value = { ...preset.parameters.noise }
        }
        break
      case 'scaling':
        if (preset.parameters.scaling) {
          scalingParams.value = { ...preset.parameters.scaling }
        }
        break
    }
  }

  const getAlgorithmDescription = (): string => {
    switch (selectedType.value) {
      case 'binarization':
        return 'Converts grayscale images to black and white using various thresholding techniques. Otsu is best for images with clear bimodal histograms, while Sauvola and Niblack work better for documents with varying lighting.'

      case 'morphology':
        return 'Applies morphological operations to binary images. Opening removes noise, closing fills gaps, erosion shrinks objects, and dilation expands them. Useful for cleaning up binary images.'

      case 'noise-reduction':
        return 'Removes noise from images. Median filtering works well for salt-and-pepper noise, while binary noise removal eliminates small unwanted components from binary images.'

      case 'scaling':
        return 'Scales images using different algorithms. Scale2x/3x/4x are specialized for pixel art and maintain sharp edges, while nearest neighbor and bilinear offer traditional scaling approaches.'

      default:
        return 'Select an algorithm to see its description.'
    }
  }

  const retryInitialization = async () => {
    if (initializationAttempts.value >= maxInitializationAttempts) {
      console.error('❌ Maximum initialization attempts reached')
      return
    }

    initializationAttempts.value++
    systemStatus.value.error = null
    console.log(`🔄 Retrying initialization (attempt ${initializationAttempts.value}/${maxInitializationAttempts})`)
    
    await initializeWorkers()
  }

  // Initialize worker orchestrator and set up callbacks
  const initializeWorkers = async () => {
    try {
      console.group('🚀 Initializing Processing System')
      console.log('Attempt:', initializationAttempts.value + 1)
      
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
          // Convert ArrayBuffer back to ImageData format
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
      
      systemStatus.value.error = null
      console.log('✅ Processing system fully initialized')
      console.groupEnd()
      
    } catch (error) {
      console.error('❌ Failed to initialize processing system:', error)
      systemStatus.value.error = error instanceof Error ? error.message : 'Initialization failed'
      console.groupEnd()
    }
  }

  // Watch for scaling method changes to update factor
  watch(() => scalingParams.value.method, (newMethod) => {
    if (newMethod.startsWith('scale')) {
      const factor = parseInt(newMethod.replace('scale', '').replace('x', ''))
      scalingParams.value.factor = factor
    }
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
    
    if (workerOrchestrator) {
      // Don't destroy the orchestrator as it might be used by other components
      // Just clean up our references
      workerOrchestrator = null
    }
    
    processingModule = null
  })
</script>

<style scoped>
  .slider::-webkit-slider-thumb {
    -webkit-appearance: none;
    appearance: none;
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #3b82f6;
    cursor: pointer;
    border: 2px solid #ffffff;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.1);
  }

  .slider::-moz-range-thumb {
    width: 20px;
    height: 20px;
    border-radius: 50%;
    background: #3b82f6;
    cursor: pointer;
    border: 2px solid #ffffff;
    box-shadow: 0 0 0 1px rgba(0, 0, 0, 0.1);
  }

  .slider:focus {
    outline: none;
  }

  .slider:focus::-webkit-slider-thumb {
    box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.3);
  }

  /* Debug panel styling */
  .debug-panel {
    font-family: 'Courier New', monospace;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>