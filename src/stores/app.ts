import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { 
  ImageData, 
  ProcessingTask, 
  ProcessingParameters,
  ProcessingType,
  Plugin,
  CanvasState 
} from '@/types'

export const useAppStore = defineStore('app', () => {
  // State
  const currentImage = ref<ImageData | null>(null)
  const processedImage = ref<ImageData | null>(null)
  const activeTasks = ref<ProcessingTask[]>([])
  const plugins = ref<Plugin[]>([])
  const isInitialized = ref(false)
  const canvasState = ref<CanvasState>({
    zoom: 1,
    offsetX: 0,
    offsetY: 0,
    showOriginal: true
  })

  // Computed
  const isProcessing = computed(() => 
    activeTasks.value.some(task => task.status === 'processing')
  )

  const hasImage = computed(() => currentImage.value !== null)

  const hasProcessedImage = computed(() => processedImage.value !== null)

  const currentTask = computed(() => 
    activeTasks.value.find(task => task.status === 'processing')
  )

  const completedTasks = computed(() =>
    activeTasks.value.filter(task => task.status === 'completed')
  )

  // Actions
  const setCurrentImage = (image: ImageData) => {
    currentImage.value = image
    // Reset processed image when new image is loaded
    processedImage.value = null
    // Clear canvas offset
    canvasState.value.offsetX = 0
    canvasState.value.offsetY = 0
    canvasState.value.zoom = 1
  }

  const setProcessedImage = (image: ImageData) => {
    processedImage.value = image
  }

  const addTask = (
    type: ProcessingType, 
    parameters: ProcessingParameters
  ): ProcessingTask => {
    const task: ProcessingTask = {
      id: generateTaskId(),
      type,
      parameters,
      status: 'pending',
      progress: 0,
      createdAt: new Date()
    }
    
    activeTasks.value.push(task)
    return task
  }

  const updateTask = (taskId: string, updates: Partial<ProcessingTask>) => {
    const taskIndex = activeTasks.value.findIndex(task => task.id === taskId)
    if (taskIndex !== -1) {
      activeTasks.value[taskIndex] = {
        ...activeTasks.value[taskIndex],
        ...updates
      }
      
      // If task completed, set completion time
      if (updates.status === 'completed' || updates.status === 'failed') {
        activeTasks.value[taskIndex].completedAt = new Date()
      }
    }
  }

  const removeTask = (taskId: string) => {
    const taskIndex = activeTasks.value.findIndex(task => task.id === taskId)
    if (taskIndex !== -1) {
      activeTasks.value.splice(taskIndex, 1)
    }
  }

  const cancelTask = (taskId: string) => {
    updateTask(taskId, { status: 'cancelled' })
  }

  const clearCompletedTasks = () => {
    activeTasks.value = activeTasks.value.filter(
      task => !['completed', 'failed', 'cancelled'].includes(task.status)
    )
  }

  const updateCanvasState = (updates: Partial<CanvasState>) => {
    canvasState.value = {
      ...canvasState.value,
      ...updates
    }
  }

  const resetCanvas = () => {
    canvasState.value = {
      zoom: 1,
      offsetX: 0,
      offsetY: 0,
      showOriginal: true
    }
  }

  const addPlugin = (plugin: Plugin) => {
    const existingIndex = plugins.value.findIndex(p => p.name === plugin.name)
    if (existingIndex !== -1) {
      plugins.value[existingIndex] = plugin
    } else {
      plugins.value.push(plugin)
    }
  }

  const removePlugin = (pluginName: string) => {
    const pluginIndex = plugins.value.findIndex(p => p.name === pluginName)
    if (pluginIndex !== -1) {
      plugins.value.splice(pluginIndex, 1)
    }
  }

  const initialize = async () => {
    if (isInitialized.value) return

    try {
      // Initialize WASM modules, plugins, etc.
      console.log('Initializing Engraving Processor Pro...')
      
      // Load any saved state from localStorage if needed
      // Note: Not using localStorage as per restrictions, but this is where it would go
      
      isInitialized.value = true
      console.log('Initialization complete')
    } catch (error) {
      console.error('Failed to initialize:', error)
      throw error
    }
  }

  const reset = () => {
    currentImage.value = null
    processedImage.value = null
    activeTasks.value = []
    resetCanvas()
  }

  // Utility functions
  const generateTaskId = (): string => {
    return `task_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  return {
    // State
    currentImage,
    processedImage,
    activeTasks,
    plugins,
    isInitialized,
    canvasState,
    
    // Computed
    isProcessing,
    hasImage,
    hasProcessedImage,
    currentTask,
    completedTasks,
    
    // Actions
    setCurrentImage,
    setProcessedImage,
    addTask,
    updateTask,
    removeTask,
    cancelTask,
    clearCompletedTasks,
    updateCanvasState,
    resetCanvas,
    addPlugin,
    removePlugin,
    initialize,
    reset
  }