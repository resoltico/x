<template>
  <div class="card">
    <h2 class="text-xl font-semibold text-slate-800 mb-4">Image Input</h2>
    
    <!-- Drop Zone -->
    <div
      ref="dropZone"
      class="drop-zone"
      :class="{ 'drag-over': isDragOver }"
      @drop="handleDrop"
      @dragover="handleDragOver"
      @dragenter="handleDragEnter"
      @dragleave="handleDragLeave"
      @click="triggerFileInput"
    >
      <div class="text-center">
        <svg
          class="mx-auto h-12 w-12 text-slate-400 mb-4"
          stroke="currentColor"
          fill="none"
          viewBox="0 0 48 48"
        >
          <path
            d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
        
        <div v-if="!isLoading">
          <p class="text-lg text-slate-600 mb-2">
            Drop your image here, or 
            <span class="text-blue-600 font-medium cursor-pointer hover:text-blue-700">
              click to browse
            </span>
          </p>
          <p class="text-sm text-slate-500">
            Supports: {{ supportedFormats.join(', ') }} (max {{ maxFileSize }})
          </p>
        </div>
        
        <div v-else class="flex items-center justify-center">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
          <span class="ml-3 text-slate-600">Processing image...</span>
        </div>
      </div>
    </div>

    <!-- Hidden File Input -->
    <input
      ref="fileInput"
      type="file"
      :accept="acceptedTypes"
      @change="handleFileSelect"
      class="hidden"
    />

    <!-- Current Image Info -->
    <div v-if="currentImage" class="mt-6 p-4 bg-slate-50 rounded-lg">
      <h3 class="text-sm font-medium text-slate-800 mb-2">Current Image</h3>
      <div class="grid grid-cols-2 gap-4 text-sm">
        <div>
          <span class="text-slate-500">Filename:</span>
          <span class="ml-2 font-medium">{{ currentImage.filename || 'Unknown' }}</span>
        </div>
        <div>
          <span class="text-slate-500">Size:</span>
          <span class="ml-2 font-medium">{{ formatFileSize(currentImage.size) }}</span>
        </div>
        <div>
          <span class="text-slate-500">Dimensions:</span>
          <span class="ml-2 font-medium">{{ currentImage.width }}×{{ currentImage.height }}</span>
        </div>
        <div>
          <span class="text-slate-500">Format:</span>
          <span class="ml-2 font-medium">{{ currentImage.format }}</span>
        </div>
      </div>
      
      <button
        @click="clearImage"
        class="mt-3 btn btn-secondary text-sm"
      >
        Clear Image
      </button>
    </div>

    <!-- Warnings -->
    <div v-if="warnings.length > 0" class="mt-4">
      <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
        <div class="flex">
          <svg class="h-5 w-5 text-yellow-400 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
          <div class="ml-3">
            <h3 class="text-sm font-medium text-yellow-800">Warnings</h3>
            <ul class="mt-2 text-sm text-yellow-700 list-disc list-inside">
              <li v-for="warning in warnings" :key="warning">{{ warning }}</li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <!-- Error Display -->
    <div v-if="error" class="mt-4">
      <div class="bg-red-50 border border-red-200 rounded-lg p-4">
        <div class="flex">
          <svg class="h-5 w-5 text-red-400 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
          </svg>
          <div class="ml-3">
            <h3 class="text-sm font-medium text-red-800">Error</h3>
            <p class="mt-1 text-sm text-red-700">{{ error }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, onUnmounted } from 'vue'
  import { useAppStore } from '@/stores/app'
  import { ImageInputModule } from '@/modules/ImageInputModule'
  import { formatFileSize } from '@/utils/fileValidation'
  import type { ImageData } from '@/types'

  // Store
  const appStore = useAppStore()

  // Refs
  const dropZone = ref<HTMLElement>()
  const fileInput = ref<HTMLInputElement>()
  const isDragOver = ref(false)
  const isLoading = ref(false)
  const error = ref<string>('')
  const warnings = ref<string[]>([])

  // Image input module
  const imageInputModule = ImageInputModule.getInstance()

  // Computed
  const currentImage = computed(() => appStore.currentImage)
  const supportedFormats = computed(() => imageInputModule.getSupportedFormats())
  const acceptedTypes = computed(() => imageInputModule.getSupportedExtensions())
  const maxFileSize = computed(() => '10MB')

  // Event handlers
  const handleDrop = async (event: DragEvent) => {
    isDragOver.value = false
    
    const dropHandler = imageInputModule.createDropHandler(
      handleImageSuccess,
      handleImageError
    )
    
    isLoading.value = true
    await dropHandler.handleDrop(event)
    isLoading.value = false
  }

  const handleDragOver = (event: DragEvent) => {
    event.preventDefault()
    event.dataTransfer!.dropEffect = 'copy'
  }

  const handleDragEnter = (event: DragEvent) => {
    event.preventDefault()
    if (imageInputModule.validateDropData(event)) {
      isDragOver.value = true
    }
  }

  const handleDragLeave = (event: DragEvent) => {
    event.preventDefault()
    // Only remove drag-over state if leaving the drop zone
    if (!dropZone.value?.contains(event.relatedTarget as Node)) {
      isDragOver.value = false
    }
  }

  const handleFileSelect = async (event: Event) => {
    const target = event.target as HTMLInputElement
    const files = target.files
    
    if (!files || files.length === 0) return

    isLoading.value = true
    error.value = ''
    warnings.value = []

    try {
      const result = await imageInputModule.processFile(files[0])
      
      if (result.success && result.imageData) {
        handleImageSuccess(result.imageData, result.validation.warnings)
      } else {
        handleImageError(result.error || 'Failed to process file')
      }
    } catch (err) {
      handleImageError(err instanceof Error ? err.message : 'Unknown error')
    } finally {
      isLoading.value = false
      // Reset file input
      target.value = ''
    }
  }

  const triggerFileInput = () => {
    if (!isLoading.value) {
      fileInput.value?.click()
    }
  }

  const clearImage = () => {
    appStore.setCurrentImage(null as any)
    error.value = ''
    warnings.value = []
  }

  const handleImageSuccess = (imageData: ImageData, imageWarnings?: string[]) => {
    appStore.setCurrentImage(imageData)
    error.value = ''
    warnings.value = imageWarnings || []
  }

  const handleImageError = (errorMessage: string) => {
    error.value = errorMessage
    warnings.value = []
  }

  // Prevent default drag behaviors on the entire document
  const preventDefaults = (e: Event) => {
    e.preventDefault()
    e.stopPropagation()
  }

  const handleDocumentDrop = (e: Event) => {
    preventDefaults(e)
    // Only prevent if not dropping on our drop zone
    if (!dropZone.value?.contains(e.target as Node)) {
      isDragOver.value = false
    }
  }

  // Lifecycle
  onMounted(() => {
    // Prevent default drag/drop behavior on the entire document
    ;['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
      document.addEventListener(eventName, preventDefaults, false)
    })
    document.addEventListener('drop', handleDocumentDrop, false)
  })

  onUnmounted(() => {
    ;['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
      document.removeEventListener(eventName, preventDefaults, false)
    })
    document.removeEventListener('drop', handleDocumentDrop, false)
  })
</script>

<style scoped>
  .drop-zone {
    cursor: pointer;
    transition: all 0.2s ease-in-out;
  }

  .drop-zone:hover {
    @apply border-blue-400 bg-blue-50;
  }

  .drop-zone.drag-over {
    @apply border-blue-500 bg-blue-50;
    transform: scale(1.02);
  }
</style>