// src/workers/imageProcessingWorker.ts
/// <reference lib="webworker" />

import type { WorkerMessage, WorkerResponse, ProcessingType, ProcessingParameters } from '../types'

// Import processing modules
import { BinarizationProcessor } from '../modules/processing/binarization'
import { MorphologyProcessor } from '../modules/processing/morphology'
import { NoiseReductionProcessor } from '../modules/processing/noiseReduction'
import { ScalingProcessor } from '../modules/processing/scaling'

/**
 * Enhanced Web Worker for image processing tasks
 * Runs processing operations in a separate thread to avoid blocking the UI
 */

// Processing task tracker
const activeTasks = new Map<string, boolean>()

// Ensure we're in worker context
declare const self: DedicatedWorkerGlobalScope & typeof globalThis

// Enhanced console logging for worker
const workerLog = (level: 'log' | 'warn' | 'error', ...args: any[]) => {
  const timestamp = new Date().toISOString().substr(11, 12)
  const prefix = `[${timestamp}] 🔧 Worker:`
  console[level](prefix, ...args)
}

// Message handler with enhanced error handling
self.onmessage = async (event: MessageEvent<WorkerMessage>) => {
  const { id, type, payload } = event.data

  workerLog('log', `Received message: ${type} for task ${id}`)

  try {
    switch (type) {
      case 'test':
        // Handle test messages for worker validation
        sendTestResponse(id)
        break
        
      case 'process':
        await handleProcessRequest(id, payload)
        break
        
      case 'cancel':
        handleCancelRequest(id)
        break
        
      default:
        sendError(id, `Unknown message type: ${type}`)
    }
  } catch (error) {
    workerLog('error', 'Error processing message:', error)
    sendError(id, error instanceof Error ? error.message : 'Unknown error')
  }
}

/**
 * Send test response for worker validation
 */
function sendTestResponse(taskId: string) {
  const response = {
    id: taskId,
    type: 'test-response' as const
  }
  self.postMessage(response)
}

/**
 * Handle process request with enhanced error handling
 */
async function handleProcessRequest(taskId: string, payload: any) {
  try {
    workerLog('log', `Processing task ${taskId}`)
    
    // Mark task as active
    activeTasks.set(taskId, true)
    
    sendProgress(taskId, 5, 'Initializing worker processing...')

    if (!payload) {
      throw new Error('No payload provided')
    }

    const { imageData, type, parameters } = payload
    
    if (!imageData) {
      throw new Error('No image data provided')
    }

    if (!type) {
      throw new Error('No processing type specified')
    }

    if (!parameters) {
      throw new Error('No parameters provided')
    }

    workerLog('log', `Processing type: ${type}`, {
      imageSize: imageData.width ? `${imageData.width}x${imageData.height}` : 'unknown',
      parameters: Object.keys(parameters)
    })
    
    sendProgress(taskId, 15, 'Validating image data...')

    // Validate image data
    if (!imageData.data || !imageData.width || !imageData.height) {
      throw new Error('Invalid image data structure')
    }

    sendProgress(taskId, 25, 'Starting image processing...')

    // Process the image using the appropriate processor
    const result = await processImage(taskId, imageData, type, parameters)

    // Check if task was cancelled during processing
    if (!activeTasks.get(taskId)) {
      workerLog('log', `Task ${taskId} was cancelled during processing`)
      return // Task was cancelled
    }

    sendProgress(taskId, 95, 'Finalizing result...')

    // Validate result
    if (!result || !(result instanceof ArrayBuffer)) {
      throw new Error('Invalid processing result')
    }

    // Send result back to main thread
    sendResult(taskId, result)
    
  } catch (error) {
    workerLog('error', `Task ${taskId} failed:`, error)
    sendError(taskId, error instanceof Error ? error.message : 'Processing failed')
  } finally {
    activeTasks.delete(taskId)
  }
}

/**
 * Handle cancel request
 */
function handleCancelRequest(taskId: string) {
  workerLog('log', `Cancelling task ${taskId}`)
  activeTasks.delete(taskId)
  sendError(taskId, 'Task cancelled by user')
}

/**
 * Main image processing function with progress tracking
 */
async function processImage(
  taskId: string,
  imageData: any,
  type: ProcessingType,
  parameters: ProcessingParameters
): Promise<ArrayBuffer> {
  workerLog('log', `Processing image: ${imageData.width}x${imageData.height}, type: ${type}`)
  
  // Check if task was cancelled
  if (!activeTasks.get(taskId)) {
    throw new Error('Task was cancelled')
  }
  
  sendProgress(taskId, 35, `Applying ${type} processing...`)
  
  try {
    // Process based on type using the appropriate processor
    let result: ArrayBuffer
    
    switch (type) {
      case 'binarization':
        if (!parameters.binarization) {
          throw new Error('Binarization parameters required')
        }
        sendProgress(taskId, 45, 'Applying binarization...')
        result = await BinarizationProcessor.process(imageData, parameters.binarization)
        break
        
      case 'morphology':
        if (!parameters.morphology) {
          throw new Error('Morphology parameters required')
        }
        sendProgress(taskId, 45, 'Applying morphological operations...')
        result = await MorphologyProcessor.process(imageData, parameters.morphology)
        break
        
      case 'noise-reduction':
        if (!parameters.noise) {
          throw new Error('Noise reduction parameters required')
        }
        sendProgress(taskId, 45, 'Reducing noise...')
        result = await NoiseReductionProcessor.process(imageData, parameters.noise)
        break
        
      case 'scaling':
        if (!parameters.scaling) {
          throw new Error('Scaling parameters required')
        }
        sendProgress(taskId, 45, 'Scaling image...')
        result = await ScalingProcessor.process(imageData, parameters.scaling)
        break
        
      default:
        throw new Error(`Unsupported processing type: ${type}`)
    }
    
    // Check if task was cancelled during processing
    if (!activeTasks.get(taskId)) {
      throw new Error('Task was cancelled during processing')
    }
    
    sendProgress(taskId, 85, 'Processing completed, preparing result...')
    
    if (!result || result.byteLength === 0) {
      throw new Error('Processing produced empty result')
    }
    
    workerLog('log', `Processing completed successfully, result size: ${result.byteLength} bytes`)
    return result
    
  } catch (error) {
    workerLog('error', 'Processing error:', error)
    throw error
  }
}

/**
 * Send progress update to main thread
 */
function sendProgress(taskId: string, progress: number, message?: string) {
  const response: WorkerResponse = {
    id: taskId,
    type: 'progress',
    payload: { progress, message }
  }
  workerLog('log', `Progress for ${taskId}: ${progress}%${message ? ' - ' + message : ''}`)
  self.postMessage(response)
}

/**
 * Send result to main thread with enhanced error handling
 */
function sendResult(taskId: string, result: ArrayBuffer) {
  workerLog('log', `Sending result for ${taskId}, size: ${result.byteLength} bytes`)
  
  const response: WorkerResponse = {
    id: taskId,
    type: 'result',
    payload: { result }
  }
  
  try {
    // Use transferable objects for zero-copy transfer
    self.postMessage(response, [result])
    workerLog('log', `Result sent successfully for ${taskId}`)
  } catch (transferError) {
    workerLog('warn', 'Failed to transfer result, trying fallback:', transferError)
    // Fallback: try without transfer (creates a copy)
    try {
      const response2: WorkerResponse = {
        id: taskId,
        type: 'result',
        payload: { result: result.slice(0) } // Create a copy
      }
      self.postMessage(response2)
      workerLog('log', `Result sent with fallback method for ${taskId}`)
    } catch (fallbackError) {
      workerLog('error', 'Both transfer methods failed:', fallbackError)
      sendError(taskId, 'Failed to send processing result')
    }
  }
}

/**
 * Send error to main thread
 */
function sendError(taskId: string, error: string) {
  workerLog('error', `Error for ${taskId}:`, error)
  
  const response: WorkerResponse = {
    id: taskId,
    type: 'error',
    payload: { error }
  }
  
  try {
    self.postMessage(response)
  } catch (postError) {
    workerLog('error', 'Failed to send error message:', postError)
  }
}

// Enhanced error handling with proper typing
self.onerror = (event: string | Event) => {
  if (typeof event === 'string') {
    workerLog('error', 'Worker script error:', event)
  } else if (event instanceof ErrorEvent) {
    workerLog('error', 'Worker script error:', {
      message: event.message,
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      error: event.error
    })
  } else {
    workerLog('error', 'Worker script error:', event)
  }
  
  // Try to notify main thread about critical worker error
  try {
    self.postMessage({
      id: 'worker-error',
      type: 'error',
      payload: { error: 'Worker encountered a critical error' }
    })
  } catch (postError) {
    // If we can't even send an error message, log it
    workerLog('error', 'Cannot send error to main thread:', postError)
  }
}

// Handle unhandled promise rejections
self.onunhandledrejection = (event: PromiseRejectionEvent) => {
  workerLog('error', 'Worker unhandled rejection:', event.reason)
  
  // Try to prevent the default behavior
  event.preventDefault()
  
  // Try to notify main thread
  try {
    self.postMessage({
      id: 'worker-rejection',
      type: 'error',
      payload: { error: `Unhandled promise rejection: ${event.reason}` }
    })
  } catch (postError) {
    workerLog('error', 'Cannot send rejection error to main thread:', postError)
  }
}

// Enhanced worker initialization logging
workerLog('log', 'Image processing worker initialized successfully')
workerLog('log', 'Worker capabilities:', {
  hasOffscreenCanvas: typeof OffscreenCanvas !== 'undefined',
  hasImageBitmap: typeof ImageBitmap !== 'undefined',  
  hasCreateImageBitmap: typeof createImageBitmap !== 'undefined',
  hasArrayBuffer: typeof ArrayBuffer !== 'undefined',
  hasUint8ClampedArray: typeof Uint8ClampedArray !== 'undefined'
})

// Send initialization complete message
try {
  self.postMessage({
    id: 'worker-init',
    type: 'ready',
    payload: { message: 'Worker initialized and ready' }
  })
} catch (initError) {
  workerLog('error', 'Cannot send initialization message:', initError)
}

// Export empty object to make this a module
export {}