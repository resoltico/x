<template>
  <div class="card">
    <h2 class="text-xl font-semibold text-gray-800 mb-4">Processing Controls</h2>

    <!-- Algorithm Selection -->
    <div class="space-y-6">
      <!-- Processing Type -->
      <div>
        <label class="label">Processing Type</label>
        <select
          v-model="selectedType"
          :disabled="!hasImage || isProcessing"
          class="input"
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
            K Factor: {{ binarizationParams.k.toFixed(2) }}
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
      <div class="flex space-x-3 pt-4 border-t border-gray-200">
        <button
          @click="processPreview"
          :disabled="!canProcess || isProcessing"
          class="btn btn-secondary flex-1"
        >
          <svg v-if="isProcessing" class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          {{ isProcessing ? 'Processing...' : 'Preview' }}
        </button>
        
        <button
          @click="processFullSize"
          :disabled="!canProcess || isProcessing"
          class="btn btn-primary flex-1"
        >
          Process Full Size
        </button>
      </div>

      <!-- Quick Presets -->
      <div class="pt-4 border-t border-gray-200">
        <h3 class="text-sm font-medium text-gray-700 mb-3">Quick Presets</h3>
        <div class="grid grid-cols-2 gap-2">
          <button
            @click="applyPreset('document')"
            :disabled="!hasImage"
            class="btn btn-secondary text-sm"
          >
            📄 Document
          </button>
          <button
            @click="applyPreset('engraving')"
            :disabled="!hasImage"
            class="btn btn-secondary text-sm"
          >
            🖼️ Engraving
          </button>
          <button
            @click="applyPreset('pixel-art')"
            :disabled="!hasImage"
            class="btn btn-secondary text-sm"
          >
            🎮 Pixel Art
          </button>
          <button
            @click="applyPreset('noise-clean')"
            :disabled="!hasImage"
            class="btn btn-secondary text-sm"
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
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, watch } from 'vue'
  import { useAppStore } from '@/stores/app'
  import { WorkerOrchestratorModule } from '@/modules/WorkerOrchestratorModule'
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

  // Worker orchestrator
  const workerOrchestrator = WorkerOrchestratorModule.getInstance()

  // Refs
  const selectedType = ref<ProcessingType | ''>('')

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
  const canProcess = computed(() => hasImage.value && selectedType.value !== '')

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
    if (!canProcess.value || !appStore.currentImage) return

    try {
      const parameters = buildProcessingParameters()
      const taskId = await workerOrchestrator.submitTask(
        appStore.currentImage,
        selectedType.value as ProcessingType,
        parameters
      )
      
      console.log('Preview processing started:', taskId)
    } catch (error) {
      console.error('Failed to start preview processing:', error)
    }
  }

  const processFullSize = async () => {
    if (!canProcess.value || !appStore.currentImage) return

    try {
      const parameters = buildProcessingParameters()
      const taskId = await workerOrchestrator.submitTask(
        appStore.currentImage,
        selectedType.value as ProcessingType,
        parameters
      )
      
      console.log('Full-size processing started:', taskId)
    } catch (error) {
      console.error('Failed to start full-size processing:', error)
    }
  }

  const applyPreset = (presetName: string) => {
    switch (presetName) {
      case 'document':
        selectedType.value = 'binarization'
        binarizationParams.value = {
          method: 'sauvola',
          windowSize: 15,
          k: 0.2,
          threshold: 128
        }
        break

      case 'engraving':
        selectedType.value = 'binarization'
        binarizationParams.value = {
          method: 'otsu',
          windowSize: 15,
          k: 0.2,
          threshold: 128
        }
        break

      case 'pixel-art':
        selectedType.value = 'scaling'
        scalingParams.value = {
          method: 'scale2x',
          factor: 2
        }
        break

      case 'noise-clean':
        selectedType.value = 'noise-reduction'
        noiseParams.value = {
          method: 'median',
          kernelSize: 3,
          threshold: 50
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

  // Initialize worker orchestrator and set up callbacks
  const initializeWorkers = async () => {
    try {
      await workerOrchestrator.initialize()
      
      // Set up task update callback
      workerOrchestrator.setTaskUpdateCallback((task) => {
        appStore.updateTask(task.id, task)
        
        // If task completed successfully, update processed image
        if (task.status === 'completed' && task.result) {
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
    } catch (error) {
      console.error('Failed to initialize workers:', error)
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
  initializeWorkers()
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
</style>