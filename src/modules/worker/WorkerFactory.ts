// src/modules/worker/WorkerFactory.ts
// Enhanced factory with better URL detection and debugging

import type { WorkerCapabilities, WorkerEnvironmentInfo } from '@/types/worker-status'
import { debugLogger } from '@/utils/debugLogger'

export class WorkerFactory {
  private static instance: WorkerFactory
  private environmentInfo: WorkerEnvironmentInfo
  private capabilities: WorkerCapabilities

  constructor() {
    this.environmentInfo = this.detectEnvironment()
    this.capabilities = this.detectCapabilities()
    debugLogger.log('info', 'worker', 'WorkerFactory initialized', {
      environment: this.environmentInfo,
      capabilities: this.capabilities
    })
  }

  static getInstance(): WorkerFactory {
    if (!WorkerFactory.instance) {
      WorkerFactory.instance = new WorkerFactory()
    }
    return WorkerFactory.instance
  }

  /**
   * Create a new worker with comprehensive URL detection
   */
  async createWorker(index: number): Promise<Worker | null> {
    debugLogger.log('info', 'worker', `Creating worker ${index}`, { environment: this.environmentInfo })
    
    const urls = await this.getWorkerUrls()
    debugLogger.log('debug', 'worker', `Testing ${urls.length} URLs for worker ${index}`, urls)
    
    for (const url of urls) {
      try {
        debugLogger.log('debug', 'worker', `Attempting worker ${index} with URL: ${url}`)
        
        const worker = await this.tryCreateSingleWorker(url, index)
        if (worker) {
          debugLogger.log('info', 'worker', `Worker ${index} created successfully with: ${url}`)
          return worker
        }
        
      } catch (error) {
        debugLogger.log('warn', 'worker', `Worker ${index} failed with URL ${url}`, error)
        continue
      }
    }
    
    debugLogger.log('error', 'worker', `All URLs failed for worker ${index}`)
    return null
  }

  /**
   * Try to create a single worker instance with detailed logging
   */
  private async tryCreateSingleWorker(url: string, index: number): Promise<Worker | null> {
    try {
      // First test if the URL is accessible
      const accessible = await this.testUrlAccessibility(url)
      if (!accessible) {
        debugLogger.log('debug', 'worker', `URL not accessible: ${url}`)
        return null
      }

      const options: WorkerOptions = { 
        type: this.getWorkerType(url),
        name: `image-worker-${index}`
      }
      
      debugLogger.log('debug', 'worker', `Creating worker with options`, { url, options })
      const worker = new Worker(url, options)
      
      // Test the worker with a timeout
      const testResult = await this.testWorker(worker, 5000)
      if (testResult) {
        debugLogger.log('info', 'worker', `Worker ${index} test passed for: ${url}`)
        return worker
      } else {
        debugLogger.log('warn', 'worker', `Worker ${index} test failed for: ${url}`)
        worker.terminate()
        return null
      }
    } catch (error) {
      debugLogger.log('error', 'worker', `Worker ${index} creation error with ${url}`, error)
      return null
    }
  }

  /**
   * Test URL accessibility
   */
  private async testUrlAccessibility(url: string): Promise<boolean> {
    // Skip accessibility test for data URLs and blob URLs
    if (url.startsWith('data:') || url.startsWith('blob:')) {
      return true
    }

    try {
      // Check if fetch is available (it might not be in worker context)
      if (typeof globalThis.fetch === 'undefined') {
        return true // Assume accessible if we can't test
      }
      
      const response = await globalThis.fetch(url, { 
        method: 'HEAD',
        mode: 'no-cors' // Allow cross-origin requests
      })
      return response.ok || response.type === 'opaque'
    } catch (error) {
      debugLogger.log('debug', 'worker', `URL accessibility test failed for ${url}`, error)
      return false
    }
  }

  /**
   * Determine worker type based on URL
   */
  private getWorkerType(url: string): 'classic' | 'module' {
    if (url.endsWith('.ts') || url.includes('type=module') || this.environmentInfo.isDevelopment) {
      return 'module'
    }
    return 'classic'
  }

  /**
   * Get possible worker URLs with intelligent detection
   */
  private async getWorkerUrls(): Promise<string[]> {
    const urls: string[] = []
    
    debugLogger.log('debug', 'worker', 'Detecting worker URLs', this.environmentInfo)

    // Development environment (Vite dev server)
    if (this.environmentInfo.isDevelopment) {
      urls.push('/src/workers/imageProcessingWorker.ts')
      debugLogger.log('debug', 'worker', 'Added development URL')
    }

    // Production environment - detect actual built files
    if (this.environmentInfo.isProduction || !this.environmentInfo.isDevelopment) {
      const productionUrls = await this.detectProductionWorkerUrls()
      urls.push(...productionUrls)
      debugLogger.log('debug', 'worker', `Added ${productionUrls.length} production URLs`, productionUrls)
    }

    // Fallback URLs
    const fallbackUrls = [
      '/public/workers/imageProcessingWorker.js',
      '/workers/imageProcessingWorker.js',
      '/assets/imageProcessingWorker.js'
    ]
    urls.push(...fallbackUrls)

    // Create inline worker as ultimate fallback
    try {
      const inlineWorkerUrl = this.createInlineWorker()
      urls.push(inlineWorkerUrl)
      debugLogger.log('debug', 'worker', 'Added inline worker fallback')
    } catch (error) {
      debugLogger.log('error', 'worker', 'Failed to create inline worker', error)
    }

    debugLogger.log('info', 'worker', `Generated ${urls.length} worker URLs to test`, urls)
    return urls
  }

  /**
   * Detect production worker URLs by scanning for actual files
   */
  private async detectProductionWorkerUrls(): Promise<string[]> {
    const urls: string[] = []

    // Common production paths
    const basePaths = ['/workers/', '/assets/', '/']

    // Try to find actual files by checking common hashed names
    for (const basePath of basePaths) {
      // Try to check for Vite manifest
      try {
        // Check if fetch is available
        if (typeof globalThis.fetch !== 'undefined') {
          const manifestResponse = await globalThis.fetch('/.vite/manifest.json', { mode: 'no-cors' })
          if (manifestResponse.ok) {
            const manifest = await manifestResponse.json()
            for (const [key, value] of Object.entries(manifest)) {
              if (key.includes('worker') || key.includes('Worker')) {
                const manifestUrl = `/${(value as any).file}`
                urls.push(manifestUrl)
                debugLogger.log('debug', 'worker', `Found worker in manifest: ${manifestUrl}`)
              }
            }
          }
        }
      } catch {
        // Manifest not available, continue with other detection methods
      }

      // Try common hash patterns for production builds
      const hashPatterns = [
        // Vite-style hashes
        'imageProcessingWorker-[a-zA-Z0-9]{8}.js',
        'worker-[a-zA-Z0-9]{8}.js',
        // Webpack-style hashes
        'imageProcessingWorker.[a-zA-Z0-9]{8}.js',
        'worker.[a-zA-Z0-9]{8}.js'
      ]

      for (const _pattern of hashPatterns) {
        // Generate some likely hash values to test
        const testHashes = [
          'CoQywvDx', 'V7JHClvx', 'C8Ss_lB_', 'CPGz7D-4', // From your actual build
          'abcdef12', '12345678', 'a1b2c3d4', 'xyz98765'
        ]

        for (const hash of testHashes) {
          const testUrl = basePath + _pattern.replace('[a-zA-Z0-9]{8}', hash)
          urls.push(testUrl)
        }
      }
    }

    return urls
  }

  /**
   * Create inline worker as fallback
   */
  private createInlineWorker(): string {
    const workerCode = this.getInlineWorkerCode()
    const blob = new Blob([workerCode], { type: 'application/javascript' })
    const url = URL.createObjectURL(blob)
    debugLogger.log('debug', 'worker', 'Created inline worker blob URL', { url })
    return url
  }

  /**
   * Test if a worker is responsive
   */
  private testWorker(worker: Worker, timeout: number = 5000): Promise<boolean> {
    return new Promise((resolve) => {
      let resolved = false
      const testId = `test-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
      
      const timer = setTimeout(() => {
        if (!resolved) {
          resolved = true
          debugLogger.log('warn', 'worker', `Worker test timeout after ${timeout}ms`, { testId })
          resolve(false)
        }
      }, timeout)
      
      const onMessage = (event: MessageEvent) => {
        debugLogger.log('debug', 'worker', 'Received test response', { testId, data: event.data })
        if (event.data?.id === testId && !resolved) {
          resolved = true
          clearTimeout(timer)
          worker.removeEventListener('message', onMessage)
          worker.removeEventListener('error', onError)
          debugLogger.log('info', 'worker', 'Worker test successful', { testId })
          resolve(true)
        }
      }
      
      const onError = (error: ErrorEvent) => {
        if (!resolved) {
          resolved = true
          clearTimeout(timer)
          worker.removeEventListener('message', onMessage)
          worker.removeEventListener('error', onError)
          debugLogger.log('error', 'worker', 'Worker test error', { testId, error })
          resolve(false)
        }
      }
      
      worker.addEventListener('message', onMessage)
      worker.addEventListener('error', onError)
      
      try {
        debugLogger.log('debug', 'worker', 'Sending test message', { testId })
        worker.postMessage({ id: testId, type: 'test' })
      } catch (error) {
        debugLogger.log('error', 'worker', 'Failed to send test message', { testId, error })
        onError(error as ErrorEvent)
      }
    })
  }

  /**
   * Detect current environment with detailed logging
   */
  private detectEnvironment(): WorkerEnvironmentInfo {
    if (typeof window === 'undefined') {
      return {
        isProduction: false,
        isDevelopment: false,
        isFileProtocol: false,
        hostname: 'server-side'
      }
    }

    const isFileProtocol = window.location.protocol === 'file:'
    const isLocalhost = ['localhost', '127.0.0.1', '0.0.0.0'].includes(window.location.hostname)
    const isDevelopment = isLocalhost || 
                         window.location.hostname.includes('dev') ||
                         window.location.port === '3000' || // Vite dev server
                         window.location.port === '5173'    // Alternative Vite port
    const isProduction = !isDevelopment && !isFileProtocol

    const info: WorkerEnvironmentInfo = {
      isProduction,
      isDevelopment,
      isFileProtocol,
      hostname: window.location.hostname,
      port: window.location.port || undefined
    }

    debugLogger.log('info', 'worker', 'Environment detected', info)
    return info
  }

  /**
   * Detect worker capabilities
   */
  private detectCapabilities(): WorkerCapabilities {
    const capabilities: WorkerCapabilities = {
      hasOffscreenCanvas: typeof OffscreenCanvas !== 'undefined',
      hasImageBitmap: typeof ImageBitmap !== 'undefined',
      hasCreateImageBitmap: typeof createImageBitmap !== 'undefined',
      hasArrayBuffer: typeof ArrayBuffer !== 'undefined',
      hasUint8ClampedArray: typeof Uint8ClampedArray !== 'undefined',
      hardwareConcurrency: navigator.hardwareConcurrency || 4
    }

    debugLogger.log('info', 'worker', 'Capabilities detected', capabilities)
    return capabilities
  }

  /**
   * Get inline worker code with enhanced functionality
   */
  private getInlineWorkerCode(): string {
    return `
// Enhanced inline worker with detailed logging
console.log('🔧 Enhanced inline worker starting...');

// Import statement for module workers
// Note: This may not work in all environments, so we have fallbacks
let BinarizationProcessor, MorphologyProcessor, NoiseReductionProcessor, ScalingProcessor;

// Try to import processing modules if available
try {
  // This will work in development with Vite
  if (typeof importScripts !== 'undefined') {
    // Classic worker - can't use ES6 imports
    console.log('🔧 Classic worker environment detected');
  }
} catch (error) {
  console.log('🔧 Module imports not available, using fallback processing');
}

// Enhanced logging for worker
const workerLog = (level, ...args) => {
  const timestamp = new Date().toISOString().substr(11, 12);
  const prefix = '[' + timestamp + '] 🔧 InlineWorker:';
  console[level](prefix, ...args);
};

// Simple processing functions
function simpleBinarization(imageData, threshold = 128) {
  return new Promise((resolve) => {
    try {
      const canvas = new OffscreenCanvas(imageData.width, imageData.height);
      const ctx = canvas.getContext('2d');
      const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
      ctx.putImageData(canvasImageData, 0, 0);
      
      const data = ctx.getImageData(0, 0, imageData.width, imageData.height);
      const pixels = data.data;
      
      for (let i = 0; i < pixels.length; i += 4) {
        const gray = pixels[i] * 0.299 + pixels[i + 1] * 0.587 + pixels[i + 2] * 0.114;
        const binary = gray > threshold ? 255 : 0;
        pixels[i] = binary;
        pixels[i + 1] = binary;
        pixels[i + 2] = binary;
      }
      
      ctx.putImageData(data, 0, 0);
      canvas.convertToBlob().then(blob => blob.arrayBuffer()).then(resolve);
    } catch (error) {
      workerLog('error', 'Binarization failed:', error);
      resolve(imageData.data.slice(0));
    }
  });
}

function simpleScale(imageData, factor) {
  return new Promise((resolve) => {
    try {
      const canvas = new OffscreenCanvas(imageData.width, imageData.height);
      const ctx = canvas.getContext('2d');
      const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
      ctx.putImageData(canvasImageData, 0, 0);
      
      const scaledCanvas = new OffscreenCanvas(
        Math.round(imageData.width * factor),
        Math.round(imageData.height * factor)
      );
      const scaledCtx = scaledCanvas.getContext('2d');
      scaledCtx.imageSmoothingEnabled = false;
      scaledCtx.drawImage(canvas, 0, 0, scaledCanvas.width, scaledCanvas.height);
      
      scaledCanvas.convertToBlob().then(blob => blob.arrayBuffer()).then(resolve);
    } catch (error) {
      workerLog('error', 'Scaling failed:', error);
      resolve(imageData.data.slice(0));
    }
  });
}

function simpleMorphology(imageData, operation) {
  return new Promise((resolve) => {
    try {
      const canvas = new OffscreenCanvas(imageData.width, imageData.height);
      const ctx = canvas.getContext('2d');
      const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
      ctx.putImageData(canvasImageData, 0, 0);
      
      // Simple erosion/dilation simulation
      const data = ctx.getImageData(0, 0, imageData.width, imageData.height);
      const pixels = data.data;
      const newPixels = new Uint8ClampedArray(pixels);
      
      for (let y = 1; y < imageData.height - 1; y++) {
        for (let x = 1; x < imageData.width - 1; x++) {
          const idx = (y * imageData.width + x) * 4;
          let value = pixels[idx];
          
          // Simple 3x3 kernel operation
          if (operation === 'erosion') {
            for (let dy = -1; dy <= 1; dy++) {
              for (let dx = -1; dx <= 1; dx++) {
                const nIdx = ((y + dy) * imageData.width + (x + dx)) * 4;
                value = Math.min(value, pixels[nIdx]);
              }
            }
          } else if (operation === 'dilation') {
            for (let dy = -1; dy <= 1; dy++) {
              for (let dx = -1; dx <= 1; dx++) {
                const nIdx = ((y + dy) * imageData.width + (x + dx)) * 4;
                value = Math.max(value, pixels[nIdx]);
              }
            }
          }
          
          newPixels[idx] = value;
          newPixels[idx + 1] = value;
          newPixels[idx + 2] = value;
        }
      }
      
      const newData = new ImageData(newPixels, imageData.width, imageData.height);
      ctx.putImageData(newData, 0, 0);
      canvas.convertToBlob().then(blob => blob.arrayBuffer()).then(resolve);
    } catch (error) {
      workerLog('error', 'Morphology failed:', error);
      resolve(imageData.data.slice(0));
    }
  });
}

function simpleNoiseReduction(imageData) {
  return new Promise((resolve) => {
    try {
      const canvas = new OffscreenCanvas(imageData.width, imageData.height);
      const ctx = canvas.getContext('2d');
      const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
      ctx.putImageData(canvasImageData, 0, 0);
      
      // Simple median filter
      const data = ctx.getImageData(0, 0, imageData.width, imageData.height);
      const pixels = data.data;
      const newPixels = new Uint8ClampedArray(pixels);
      
      for (let y = 1; y < imageData.height - 1; y++) {
        for (let x = 1; x < imageData.width - 1; x++) {
          const idx = (y * imageData.width + x) * 4;
          const values = [];
          
          for (let dy = -1; dy <= 1; dy++) {
            for (let dx = -1; dx <= 1; dx++) {
              const nIdx = ((y + dy) * imageData.width + (x + dx)) * 4;
              values.push(pixels[nIdx]);
            }
          }
          
          values.sort((a, b) => a - b);
          const median = values[Math.floor(values.length / 2)];
          
          newPixels[idx] = median;
          newPixels[idx + 1] = median;
          newPixels[idx + 2] = median;
        }
      }
      
      const newData = new ImageData(newPixels, imageData.width, imageData.height);
      ctx.putImageData(newData, 0, 0);
      canvas.convertToBlob().then(blob => blob.arrayBuffer()).then(resolve);
    } catch (error) {
      workerLog('error', 'Noise reduction failed:', error);
      resolve(imageData.data.slice(0));
    }
  });
}

// Enhanced message handler
self.onmessage = async function(event) {
  const { id, type, payload } = event.data;
  
  workerLog('log', 'Received message:', type, 'for task:', id);
  
  try {
    if (type === 'test') {
      workerLog('log', 'Responding to test message');
      self.postMessage({ id, type: 'test-response', payload: { message: 'Enhanced inline worker ready' } });
      return;
    }
    
    if (type === 'process') {
      const { imageData, type: processType, parameters } = payload;
      
      workerLog('log', 'Processing:', processType, 'Image size:', imageData.width + 'x' + imageData.height);
      
      self.postMessage({
        id, type: 'progress',
        payload: { progress: 25, message: 'Processing with enhanced inline worker...' }
      });
      
      let result;
      
      switch (processType) {
        case 'binarization':
          const threshold = parameters.binarization?.threshold || 128;
          workerLog('log', 'Applying binarization with threshold:', threshold);
          result = await simpleBinarization(imageData, threshold);
          break;
          
        case 'scaling':
          const factor = parameters.scaling?.factor || 2;
          workerLog('log', 'Applying scaling with factor:', factor);
          result = await simpleScale(imageData, factor);
          break;
          
        case 'morphology':
          const operation = parameters.morphology?.operation || 'opening';
          workerLog('log', 'Applying morphology:', operation);
          result = await simpleMorphology(imageData, operation);
          break;
          
        case 'noise-reduction':
          workerLog('log', 'Applying noise reduction');
          result = await simpleNoiseReduction(imageData);
          break;
          
        default:
          workerLog('warn', 'Unknown processing type:', processType);
          result = imageData.data.slice(0);
      }
      
      self.postMessage({
        id, type: 'progress',
        payload: { progress: 75, message: 'Finalizing...' }
      });
      
      workerLog('log', 'Processing completed, result size:', result.byteLength);
      
      self.postMessage({
        id, type: 'result',
        payload: { result }
      }, result instanceof ArrayBuffer ? [result] : []);
      
    }
  } catch (error) {
    workerLog('error', 'Processing error:', error);
    self.postMessage({
      id, type: 'error',
      payload: { error: 'Enhanced inline worker processing failed: ' + error.message }
    });
  }
};

// Enhanced error handling
self.onerror = function(error) {
  workerLog('error', 'Worker script error:', error);
};

self.onunhandledrejection = function(event) {
  workerLog('error', 'Worker unhandled rejection:', event.reason);
  event.preventDefault();
};

// Log worker capabilities
workerLog('log', 'Worker capabilities:', {
  hasOffscreenCanvas: typeof OffscreenCanvas !== 'undefined',
  hasImageBitmap: typeof ImageBitmap !== 'undefined',
  hasCreateImageBitmap: typeof createImageBitmap !== 'undefined',
  hasArrayBuffer: typeof ArrayBuffer !== 'undefined',
  hardwareConcurrency: navigator.hardwareConcurrency || 'unknown'
});

workerLog('log', 'Enhanced inline worker initialized and ready');

// Send ready message
try {
  self.postMessage({
    id: 'worker-init',
    type: 'ready',
    payload: { message: 'Enhanced inline worker initialized' }
  });
} catch (error) {
  workerLog('error', 'Could not send ready message:', error);
}
`
  }

  getEnvironmentInfo(): WorkerEnvironmentInfo {
    return { ...this.environmentInfo }
  }

  getCapabilities(): WorkerCapabilities {
    return { ...this.capabilities }
  }

  getRecommendedWorkerCount(): number {
    const base = Math.min(this.capabilities.hardwareConcurrency, 8)
    // In production or when debugging is disabled, use fewer workers
    if (this.environmentInfo.isProduction) {
      return Math.max(1, Math.floor(base / 2))
    }
    return Math.max(1, base)
  }

  /**
   * Test all possible worker URLs and return results
   */
  async diagnoseWorkerUrls(): Promise<{ [url: string]: any }> {
    const urls = await this.getWorkerUrls()
    const results: { [url: string]: any } = {}

    debugLogger.log('info', 'worker', `Diagnosing ${urls.length} worker URLs`)

    for (const url of urls) {
      try {
        const startTime = Date.now()
        const accessible = await this.testUrlAccessibility(url)
        const duration = Date.now() - startTime

        results[url] = {
          accessible,
          duration,
          type: this.getWorkerType(url)
        }

        if (accessible) {
          // Try to actually create a worker
          try {
            const worker = await this.tryCreateSingleWorker(url, 999)
            results[url].workerCreated = worker !== null
            if (worker) {
              worker.terminate()
            }
          } catch (error) {
            results[url].workerCreated = false
            results[url].workerError = error instanceof Error ? error.message : 'Unknown error'
          }
        }

        debugLogger.log('debug', 'worker', `URL diagnosis: ${url}`, results[url])
      } catch (error) {
        results[url] = {
          accessible: false,
          error: error instanceof Error ? error.message : 'Unknown error'
        }
      }
    }

    debugLogger.log('info', 'worker', 'Worker URL diagnosis completed', results)
    return results
  }
}