// Core image processing types
export interface ImageData {
  data: ArrayBuffer
  width: number
  height: number
  channels: number
  format: ImageFormat
  filename?: string
  size: number
}

export type ImageFormat = 'PNG' | 'JPEG' | 'TIFF' | 'WebP'

export interface ProcessingTask {
  id: string
  type: ProcessingType
  parameters: ProcessingParameters
  status: TaskStatus
  progress: number
  result?: ArrayBuffer
  error?: string
  createdAt: Date
  completedAt?: Date
}

export type ProcessingType = 
  | 'binarization'
  | 'morphology'
  | 'noise-reduction'
  | 'scaling'

export type TaskStatus = 
  | 'pending'
  | 'processing'
  | 'completed'
  | 'failed'
  | 'cancelled'

// Processing parameters for different algorithms
export interface ProcessingParameters {
  binarization?: BinarizationParams
  morphology?: MorphologyParams
  noise?: NoiseReductionParams
  scaling?: ScalingParams
}

export interface BinarizationParams {
  method: 'sauvola' | 'niblack' | 'otsu'
  windowSize?: number
  k?: number
  threshold?: number
}

export interface MorphologyParams {
  operation: 'opening' | 'closing' | 'dilation' | 'erosion'
  kernelSize: number
  iterations: number
}

export interface NoiseReductionParams {
  method: 'median' | 'binary-noise-removal'
  kernelSize?: number
  threshold?: number
}

export interface ScalingParams {
  method: 'scale2x' | 'scale3x' | 'scale4x' | 'nearest' | 'bilinear'
  factor: number
}

// UI Control types
export interface ControlParameter {
  name: string
  type: 'slider' | 'select' | 'toggle' | 'number'
  value: number | string | boolean
  min?: number
  max?: number
  step?: number
  options?: string[]
  label: string
  description?: string
}

// Worker communication types
export interface WorkerMessage {
  id: string
  type: 'process' | 'cancel' | 'progress'
  payload?: {
    imageData?: any
    type?: ProcessingType
    parameters?: ProcessingParameters
    progress?: number
    message?: string
  }
}

export interface WorkerResponse {
  id: string
  type: 'result' | 'progress' | 'error'
  payload?: {
    result?: ArrayBuffer
    progress?: number
    message?: string
    error?: string
  }
}

// Enhanced worker payload types
export interface ProcessingTaskPayload {
  imageData: any
  type: ProcessingType
  parameters: ProcessingParameters
}

export interface ProcessingResultPayload {
  result: ArrayBuffer
}

export interface ProcessingProgressPayload {
  progress: number
  message?: string
}

export interface ProcessingErrorPayload {
  error: string
}

// Plugin system types
export interface Plugin {
  name: string
  version: string
  description: string
  parameters: ControlParameter[]
  process: (data: ArrayBuffer, params: any) => Promise<ArrayBuffer>
}

// Store state types
export interface AppState {
  currentImage: ImageData | null
  processedImage: ImageData | null
  activeTasks: ProcessingTask[]
  isProcessing: boolean
  plugins: Plugin[]
}

// Canvas and preview types
export interface CanvasState {
  zoom: number
  offsetX: number
  offsetY: number
  showOriginal: boolean
}

// File validation types
export interface FileValidation {
  isValid: boolean
  error?: string
  warnings?: string[]
}

// Performance monitoring types
export interface PerformanceMetrics {
  processingTime: number
  memoryUsage: number
  taskId: string
  algorithm: string
}

// Error handling types
export interface ProcessingError {
  code: string
  message: string
  details?: any
  timestamp: Date
}

// Worker status types
export interface WorkerStatus {
  totalWorkers: number
  availableWorkers: number
  activeWorkers: number
  queuedTasks: number
  activeTasks: number
  initialized: boolean
}