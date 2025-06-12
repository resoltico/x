/// <reference lib="webworker" />
/// <reference path="../types/worker-globals.d.ts" />

import type { WorkerMessage, WorkerResponse, ProcessingType, ProcessingParameters } from '../types'

// Import processing modules
import { BinarizationProcessor } from '../modules/processing/binarization'
import { MorphologyProcessor } from '../modules/processing/morphology'
import { NoiseReductionProcessor } from '../modules/processing/noiseReduction'
import { ScalingProcessor } from '../modules/processing/scaling'

/**
 * Web Worker for image processing tasks
 * Runs processing operations in a separate thread to avoid blocking the UI
 */

// Processing task tracker
const activeTasks = new Map<string, boolean>()

// Ensure we're in worker context
declare const self: DedicatedWorkerGlobalScope & typeof globalThis

// Message handler
self.onmessage = async (event: MessageEvent<WorkerMessage>) => {
  const { id, type, payload } = event.data

  console.log(`Worker received message: ${type} for task ${id}`)

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
    console.error('Worker error processing message:', error)
    sendError(id, error instanceof Error ? error.message : 'Unknown error')
  }
}

/**
 * Handle process request
 */
async function handleProcessRequest(taskId: string, payload: any) {
  try {
    console.log(`Worker processing task ${taskId}`)
    
    // Mark task as active
    activeTasks.set(taskId, true)
    
    sendProgress(taskId, 10, 'Initializing processing...')

    const { imageData, type, parameters } = payload
    
    if (!imageData) {
      throw new Error('No image data provided')
    }

    console.log(`Processing type: ${type}`, parameters)
    sendProgress(taskId, 30, 'Starting image processing...')

    // Process the image using the appropriate processor
    const result = await processImage(imageData, type, parameters)

    // Check if task was cancelled
    if (!activeTasks.get(taskId)) {
      console.log(`Task ${taskId} was cancelled`)
      return // Task was cancelled
    }

    sendProgress(taskId, 90, 'Finalizing result...')

    // Send result back to main thread
    sendResult(taskId, result)
    
  } catch (error) {
    console.error(`Worker task ${taskId} failed:`, error)
    sendError(taskId, error instanceof Error ? error.message : 'Processing failed')
  } finally {
    activeTasks.delete(taskId)
  }
}

/**
 * Handle cancel request
 */
function handleCancelRequest(taskId: string) {
  console.log(`Worker cancelling task ${taskId}`)
  activeTasks.delete(taskId)
  sendError(taskId, 'Task cancelled by user')
}

/**
 * Main image processing function
 */
async function processImage(
  imageData: any,
  type: ProcessingType,
  parameters: ProcessingParameters
): Promise<ArrayBuffer> {
  console.log(`Worker processing image: ${imageData.width}x${imageData.height}, type: ${type}`)
  
  // Process based on type using the appropriate processor
  switch (type) {
    case 'binarization':
      if (!parameters.binarization) {
        throw new Error('Binarization parameters required')
      }
      return await BinarizationProcessor.process(imageData, parameters.binarization)
      
    case 'morphology':
      if (!parameters.morphology) {
        throw new Error('Morphology parameters required')
      }
      return await MorphologyProcessor.process(imageData, parameters.morphology)
      
    case 'noise-reduction':
      if (!parameters.noise) {
        throw new Error('Noise reduction parameters required')
      }
      return await NoiseReductionProcessor.process(imageData, parameters.noise)
      
    case 'scaling':
      if (!parameters.scaling) {
        throw new Error('Scaling parameters required')
      }
      return await ScalingProcessor.process(imageData, parameters.scaling)
      
    default:
      throw new Error(`Unsupported processing type: ${type}`)
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
  console.log(`Worker progress for ${taskId}: ${progress}%`)
  self.postMessage(response)
}

/**
 * Send result to main thread
 */
function sendResult(taskId: string, result: ArrayBuffer) {
  console.log(`Worker sending result for ${taskId}, size: ${result.byteLength} bytes`)
  
  const response: WorkerResponse = {
    id: taskId,
    type: 'result',
    payload: { result }
  }
  
  try {
    // Use transferable objects for zero-copy transfer
    self.postMessage(response, [result])
  } catch (error) {
    console.error('Failed to transfer result:', error)
    // Fallback: try without transfer
    const response2: WorkerResponse = {
      id: taskId,
      type: 'result',
      payload: { result: result.slice(0) } // Create a copy
    }
    self.postMessage(response2)
  }
}

/**
 * Send error to main thread
 */
function sendError(taskId: string, error: string) {
  console.error(`Worker error for ${taskId}:`, error)
  
  const response: WorkerResponse = {
    id: taskId,
    type: 'error',
    payload: { error }
  }
  self.postMessage(response)
}

// Handle worker errors with proper type
self.onerror = (event: ErrorEvent) => {
  console.error('Worker script error:', event.message, event.filename, event.lineno)
}

// Handle unhandled promise rejections
self.onunhandledrejection = (event: PromiseRejectionEvent) => {
  console.error('Worker unhandled rejection:', event.reason)
}

console.log('Image processing worker initialized')

// Export empty object to make this a module
export {}