import type { WorkerMessage, WorkerResponse, ProcessingType, ProcessingParameters } from '../types'

// Note: We'll implement a simpler processing approach in the worker to avoid WASM complexity
// In a production environment, you might want to use a different approach for WASM in workers

/**
 * Web Worker for image processing tasks
 * Runs processing operations in a separate thread to avoid blocking the UI
 */

// Processing task tracker
const activeTasks = new Map<string, boolean>()

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
    // Mark task as active
    activeTasks.set(taskId, true)
    
    sendProgress(taskId, 10, 'Initializing processing...')

    const { imageData, type, parameters } = payload
    
    if (!imageData) {
      throw new Error('No image data provided')
    }

    sendProgress(taskId, 30, 'Starting image processing...')

    // Process the image using canvas-based methods
    const result = await processImageWithCanvas(imageData, type, parameters)

    // Check if task was cancelled
    if (!activeTasks.get(taskId)) {
      return // Task was cancelled
    }

    sendProgress(taskId, 90, 'Finalizing result...')

    // Send result back to main thread
    sendResult(taskId, result)
    
  } catch (error) {
    sendError(taskId, error instanceof Error ? error.message : 'Processing failed')
  } finally {
    activeTasks.delete(taskId)
  }
}

/**
 * Handle cancel request
 */
function handleCancelRequest(taskId: string) {
  activeTasks.delete(taskId)
  sendError(taskId, 'Task cancelled by user')
}

/**
 * Process image using Canvas API (fallback when WASM is not available)
 */
async function processImageWithCanvas(
  imageData: any,
  type: ProcessingType,
  parameters: ProcessingParameters
): Promise<ArrayBuffer> {
  // Create an OffscreenCanvas for processing
  const canvas = new OffscreenCanvas(imageData.width, imageData.height)
  const ctx = canvas.getContext('2d')!
  
  // Create ImageData from the buffer
  const uint8Array = new Uint8Array(imageData.data)
  const canvasImageData = new ImageData(
    new Uint8ClampedArray(uint8Array),
    imageData.width,
    imageData.height
  )
  
  ctx.putImageData(canvasImageData, 0, 0)
  
  // Apply processing based on type
  switch (type) {
    case 'binarization':
      await processBinarization(ctx, imageData, parameters.binarization!)
      break
    case 'scaling':
      return await processScaling(canvas, imageData, parameters.scaling!)
    case 'morphology':
      await processMorphology(ctx, imageData, parameters.morphology!)
      break
    case 'noise-reduction':
      await processNoiseReduction(ctx, imageData, parameters.noise!)
      break
    default:
      console.warn(`Processing type ${type} not implemented in worker fallback`)
  }
  
  // Convert canvas to blob and then to ArrayBuffer
  const blob = await canvas.convertToBlob({ type: 'image/png' })
  return await blob.arrayBuffer()
}

/**
 * Canvas-based binarization
 */
async function processBinarization(
  ctx: OffscreenCanvasRenderingContext2D,
  imageData: any,
  params: any
) {
  const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
  const pixels = data.data
  const threshold = params.threshold || 128
  
  for (let i = 0; i < pixels.length; i += 4) {
    const gray = pixels[i] * 0.299 + pixels[i + 1] * 0.587 + pixels[i + 2] * 0.114
    const binary = gray > threshold ? 255 : 0
    pixels[i] = binary
    pixels[i + 1] = binary
    pixels[i + 2] = binary
    // Alpha channel remains unchanged
  }
  
  ctx.putImageData(data, 0, 0)
}

/**
 * Canvas-based scaling
 */
async function processScaling(
  canvas: OffscreenCanvas,
  imageData: any,
  params: any
): Promise<ArrayBuffer> {
  const factor = params.factor || 2
  const scaledCanvas = new OffscreenCanvas(
    imageData.width * factor,
    imageData.height * factor
  )
  const scaledCtx = scaledCanvas.getContext('2d')!
  
  // Set image smoothing based on method
  scaledCtx.imageSmoothingEnabled = params.method !== 'nearest'
  if (params.method === 'nearest') {
    scaledCtx.imageSmoothingQuality = 'low'
  }
  
  scaledCtx.drawImage(
    canvas,
    0, 0, imageData.width, imageData.height,
    0, 0, scaledCanvas.width, scaledCanvas.height
  )
  
  const blob = await scaledCanvas.convertToBlob({ type: 'image/png' })
  return await blob.arrayBuffer()
}

/**
 * Canvas-based morphology (simplified)
 */
async function processMorphology(
  ctx: OffscreenCanvasRenderingContext2D,
  imageData: any,
  params: any
) {
  // Simplified morphology - just apply a basic filter
  const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
  const pixels = data.data
  const kernelSize = params.kernelSize || 3
  
  // Apply a simple smoothing filter as a basic morphological operation
  for (let y = kernelSize; y < imageData.height - kernelSize; y++) {
    for (let x = kernelSize; x < imageData.width - kernelSize; x++) {
      const idx = (y * imageData.width + x) * 4
      
      let r = 0, g = 0, b = 0
      let count = 0
      
      for (let ky = -kernelSize; ky <= kernelSize; ky++) {
        for (let kx = -kernelSize; kx <= kernelSize; kx++) {
          const kidx = ((y + ky) * imageData.width + (x + kx)) * 4
          r += pixels[kidx]
          g += pixels[kidx + 1]
          b += pixels[kidx + 2]
          count++
        }
      }
      
      pixels[idx] = r / count
      pixels[idx + 1] = g / count
      pixels[idx + 2] = b / count
    }
  }
  
  ctx.putImageData(data, 0, 0)
}

/**
 * Canvas-based noise reduction
 */
async function processNoiseReduction(
  ctx: OffscreenCanvasRenderingContext2D,
  imageData: any,
  params: any
) {
  const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
  const pixels = data.data
  const kernelSize = params.kernelSize || 3
  
  // Apply median filter (simplified)
  const newPixels = new Uint8ClampedArray(pixels)
  
  for (let y = kernelSize; y < imageData.height - kernelSize; y++) {
    for (let x = kernelSize; x < imageData.width - kernelSize; x++) {
      const idx = (y * imageData.width + x) * 4
      
      const neighborhood: number[] = []
      
      for (let ky = -kernelSize; ky <= kernelSize; ky++) {
        for (let kx = -kernelSize; kx <= kernelSize; kx++) {
          const kidx = ((y + ky) * imageData.width + (x + kx)) * 4
          const gray = pixels[kidx] * 0.299 + pixels[kidx + 1] * 0.587 + pixels[kidx + 2] * 0.114
          neighborhood.push(gray)
        }
      }
      
      neighborhood.sort((a, b) => a - b)
      const median = neighborhood[Math.floor(neighborhood.length / 2)]
      
      newPixels[idx] = median
      newPixels[idx + 1] = median
      newPixels[idx + 2] = median
    }
  }
  
  const newData = new ImageData(newPixels, imageData.width, imageData.height)
  ctx.putImageData(newData, 0, 0)
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