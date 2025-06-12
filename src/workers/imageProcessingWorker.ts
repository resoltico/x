import type { WorkerMessage, WorkerResponse, ProcessingType, ProcessingParameters } from '@/types'
import { ProcessingModule } from '@/modules/ProcessingModule'

/**
 * Web Worker for image processing tasks
 * Runs processing operations in a separate thread to avoid blocking the UI
 */

// Global processing module instance
let processingModule: ProcessingModule | null = null
let isInitialized = false

// Message handler
self.onmessage = async (event: MessageEvent<WorkerMessage>) => {
  const { id, type, payload } = event.data

  try {
    switch (type) {
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
    sendError(id, error instanceof Error ? error.message : 'Unknown error')
  }
}

/**
 * Handle process request
 */
async function handleProcessRequest(taskId: string, payload: any) {
  try {
    // Initialize processing module if needed
    if (!isInitialized) {
      sendProgress(taskId, 10, 'Initializing processing engine...')
      processingModule = ProcessingModule.getInstance()
      await processingModule.initialize()
      isInitialized = true
      sendProgress(taskId, 20, 'Processing engine ready')
    }

    if (!processingModule) {
      throw new Error('Processing module not available')
    }

    const { imageData, type, parameters } = payload
    
    sendProgress(taskId, 30, 'Starting image processing...')

    // Process the image
    const result = await processingModule.processImage(
      imageData,
      type as ProcessingType,
      parameters as ProcessingParameters
    )

    sendProgress(taskId, 90, 'Finalizing result...')

    // Send result back to main thread
    sendResult(taskId, result)
    
  } catch (error) {
    sendError(taskId, error instanceof Error ? error.message : 'Processing failed')
  }
}

/**
 * Handle cancel request
 */
function handleCancelRequest(taskId: string) {
  // For now, we'll just acknowledge the cancellation
  // In a more advanced implementation, we could interrupt ongoing operations
  sendError(taskId, 'Task cancelled by user')
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
  self.postMessage(response)
}

/**
 * Send result to main thread
 */
function sendResult(taskId: string, result: ArrayBuffer) {
  const response: WorkerResponse = {
    id: taskId,
    type: 'result',
    payload: { result }
  }
  
  // Use transferable objects for zero-copy transfer
  self.postMessage(response, [result])
}

/**
 * Send error to main thread
 */
function sendError(taskId: string, error: string) {
  const response: WorkerResponse = {
    id: taskId,
    type: 'error',
    payload: { error }
  }
  self.postMessage(response)
}

// Handle worker errors
self.onerror = (event) => {
  console.error('Worker error:', event.error)
}

self.onunhandledrejection = (event) => {
  console.error('Worker unhandled rejection:', event.reason)
}

// Export empty object to make this a module
export {}