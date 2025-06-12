<template>
  <div class="card">
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-xl font-semibold text-slate-800">Preview</h2>
      
      <!-- View Controls -->
      <div class="flex items-center space-x-2">
        <button
          v-if="hasProcessedImage"
          class="btn btn-secondary text-sm"
          @click="toggleView"
        >
          Show {{ canvasState.showOriginal ? 'Processed' : 'Original' }}
        </button>
        
        <div class="flex items-center space-x-1 border border-slate-300 rounded-lg p-1">
          <button
            class="p-1 hover:bg-slate-100 rounded"
            title="Fit to view"
            @click="fitToCanvas"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 8V4m0 0h4M4 4l5 5m11-1V4m0 0h-4m4 0l-5 5M4 16v4m0 0h4m-4 0l5-5m11 5l-5-5m5 5v-4m0 4h-4" />
            </svg>
          </button>
          
          <button
            class="p-1 hover:bg-slate-100 rounded"
            title="Actual size (100%)"
            @click="zoomToActualSize"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
            </svg>
          </button>
          
          <button
            class="p-1 hover:bg-slate-100 rounded"
            title="Reset view"
            @click="resetView"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </button>
        </div>
      </div>
    </div>

    <!-- Canvas Container -->
    <div class="relative bg-slate-100 rounded-lg overflow-hidden" style="height: 500px;">
      <canvas
        ref="canvas"
        class="w-full h-full cursor-grab active:cursor-grabbing"
        @wheel="handleWheel"
        @mousedown="handleMouseDown"
        @mousemove="handleMouseMove"
        @mouseup="handleMouseUp"
        @mouseleave="handleMouseUp"
        @touchstart="handleTouchStart"
        @touchmove="handleTouchMove"
        @touchend="handleTouchEnd"
      />
      
      <!-- Empty State -->
      <div
        v-if="!hasImage"
        class="absolute inset-0 flex items-center justify-center"
      >
        <div class="text-center">
          <svg class="mx-auto h-16 w-16 text-slate-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
          </svg>
          <p class="text-slate-500 text-lg">No image loaded</p>
          <p class="text-slate-400 text-sm mt-1">Upload an image to see preview</p>
        </div>
      </div>

      <!-- Processing Overlay -->
      <div
        v-if="isProcessing"
        class="processing-overlay"
      >
        <div class="text-center text-white">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-white mx-auto mb-4"></div>
          <p class="text-lg font-medium">Processing Image...</p>
          <p class="text-sm opacity-75 mt-1">{{ processingMessage }}</p>
        </div>
      </div>
    </div>

    <!-- Zoom Info -->
    <div v-if="hasImage" class="mt-4 flex items-center justify-between text-sm text-slate-600">
      <div class="flex items-center space-x-4">
        <span>Zoom: {{ Math.round(canvasState.zoom * 100) }}%</span>
        <span>{{ currentImage?.width }}×{{ currentImage?.height }}px</span>
        <span v-if="currentImage">{{ formatFileSize(currentImage.size) }}</span>
      </div>
      
      <div class="flex items-center space-x-2">
        <span class="text-xs">{{ canvasState.showOriginal ? 'Original' : 'Processed' }}</span>
        <div
          class="w-3 h-3 rounded-full"
          :class="canvasState.showOriginal ? 'bg-blue-500' : 'bg-green-500'"
        ></div>
      </div>
    </div>

    <!-- Export Options -->
    <div v-if="hasProcessedImage" class="mt-4 pt-4 border-t border-slate-200">
      <div class="flex items-center justify-between">
        <span class="text-sm font-medium text-slate-700">Export Options</span>
        <div class="flex space-x-2">
          <button
            class="btn btn-primary text-sm"
            @click="downloadProcessed"
          >
            Download Processed
          </button>
          <button
            class="btn btn-secondary text-sm"
            @click="exportView"
          >
            Export View
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
  import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
  import { useAppStore } from '@/stores/app'
  import { PreviewRendererModule } from '@/modules/PreviewRendererModule'
  import { formatFileSize } from '@/utils/fileValidation'
  import { downloadImage } from '@/utils/imageHelpers'

  // Store
  const appStore = useAppStore()

  // Refs
  const canvas = ref<HTMLCanvasElement>()
  const processingMessage = ref('Initializing...')

  // Preview renderer module
  let previewRenderer: PreviewRendererModule | null = null

  // Computed
  const currentImage = computed(() => appStore.currentImage)
  const processedImage = computed(() => appStore.processedImage)
  const hasImage = computed(() => appStore.hasImage)
  const hasProcessedImage = computed(() => appStore.hasProcessedImage)
  const isProcessing = computed(() => appStore.isProcessing)
  const canvasState = computed(() => appStore.canvasState)
  const currentTask = computed(() => appStore.currentTask)

  // Methods
  const initializeRenderer = () => {
    if (!canvas.value) return

    previewRenderer = new PreviewRendererModule()
    previewRenderer.initialize(canvas.value)
    
    // Sync with store state
    previewRenderer.updateState(canvasState.value)
  }

  const toggleView = () => {
    if (!previewRenderer) return
    
    previewRenderer.toggleImageView()
    appStore.updateCanvasState(previewRenderer.getState())
  }

  const fitToCanvas = () => {
    if (!previewRenderer) return
    
    previewRenderer.fitToCanvas()
    appStore.updateCanvasState(previewRenderer.getState())
  }

  const zoomToActualSize = () => {
    if (!previewRenderer) return
    
    previewRenderer.zoomToActualSize()
    appStore.updateCanvasState(previewRenderer.getState())
  }

  const resetView = () => {
    if (!previewRenderer) return
    
    previewRenderer.resetView()
    appStore.updateCanvasState(previewRenderer.getState())
  }

  const downloadProcessed = () => {
    if (!processedImage.value) return
    
    downloadImage(
      processedImage.value.data,
      processedImage.value.filename || 'processed_image',
      processedImage.value.format
    )
  }

  const exportView = () => {
    if (!previewRenderer) return
    
    const dataUrl = previewRenderer.exportView()
    const link = document.createElement('a')
    link.href = dataUrl
    link.download = 'view_export.png'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  // Event handlers (delegated to renderer)
  const handleWheel = (_event: WheelEvent) => {
    // Renderer handles wheel events internally
  }

  const handleMouseDown = (_event: MouseEvent) => {
    // Renderer handles mouse events internally
  }

  const handleMouseMove = (_event: MouseEvent) => {
    // Renderer handles mouse events internally
  }

  const handleMouseUp = (_event: MouseEvent) => {
    // Renderer handles mouse events internally
  }

  const handleTouchStart = (_event: TouchEvent) => {
    // Renderer handles touch events internally
  }

  const handleTouchMove = (_event: TouchEvent) => {
    // Renderer handles touch events internally
  }

  const handleTouchEnd = (_event: TouchEvent) => {
    // Renderer handles touch events internally
  }

  // Watchers
  watch(currentImage, (newImage) => {
    if (previewRenderer && newImage) {
      previewRenderer.setCurrentImage(newImage)
      appStore.updateCanvasState(previewRenderer.getState())
    }
  })

  watch(processedImage, (newProcessedImage) => {
    if (previewRenderer && newProcessedImage) {
      previewRenderer.setProcessedImage(newProcessedImage)
    }
  })

  watch(currentTask, (task) => {
    if (task) {
      processingMessage.value = `${task.type} - ${task.progress}%`
    } else {
      processingMessage.value = 'Initializing...'
    }
  })

  // Keyboard shortcuts
  const handleKeydown = (event: KeyboardEvent) => {
    if (!previewRenderer || !hasImage.value) return

    switch (event.key) {
      case ' ':
        event.preventDefault()
        if (hasProcessedImage.value) {
          toggleView()
        }
        break
      case 'f':
        event.preventDefault()
        fitToCanvas()
        break
      case '1':
        event.preventDefault()
        zoomToActualSize()
        break
      case 'r':
        event.preventDefault()
        resetView()
        break
    }
  }

  // Lifecycle
  onMounted(async () => {
    await nextTick()
    initializeRenderer()
    
    // Add keyboard shortcuts
    window.addEventListener('keydown', handleKeydown)
    
    // Handle window resize
    const handleResize = () => {
      if (previewRenderer) {
        previewRenderer.resize()
      }
    }
    window.addEventListener('resize', handleResize)
  })

  onUnmounted(() => {
    if (previewRenderer) {
      previewRenderer.destroy()
    }
    window.removeEventListener('keydown', handleKeydown)
    window.removeEventListener('resize', () => {})
  })
</script>

<style scoped>
  .processing-overlay {
    backdrop-filter: blur(4px);
  }

  canvas {
    image-rendering: -webkit-optimize-contrast;
    image-rendering: crisp-edges;
  }
</style>