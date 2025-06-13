// src/modules/worker/WorkerFactory.ts
// Factory for creating and managing individual workers

import type { WorkerCapabilities, WorkerEnvironmentInfo } from '@/types/worker-status'

export class WorkerFactory {
  private static instance: WorkerFactory
  private environmentInfo: WorkerEnvironmentInfo
  private capabilities: WorkerCapabilities

  constructor() {
    this.environmentInfo = this.detectEnvironment()
    this.capabilities = this.detectCapabilities()
  }

  static getInstance(): WorkerFactory {
    if (!WorkerFactory.instance) {
      WorkerFactory.instance = new WorkerFactory()
    }
    return WorkerFactory.instance
  }

  /**
   * Create a new worker with multiple URL attempts
   */
  async createWorker(index: number): Promise<Worker | null> {
    const urls = this.getWorkerUrls()
    
    for (const url of urls) {
      try {
        console.log(`🔧 Attempting to create worker ${index} with URL: ${url}`)
        
        if (url.includes('*')) {
          // Handle glob patterns by checking for actual files
          const baseUrl = url.replace('*', '')
          const actualUrls = await this.resolveGlobPattern(baseUrl)
          for (const actualUrl of actualUrls) {
            const worker = await this.tryCreateSingleWorker(actualUrl, index)
            if (worker) return worker
          }
          continue
        }
        
        const worker = await this.tryCreateSingleWorker(url, index)
        if (worker) return worker
        
      } catch (error) {
        console.warn(`❌ Worker ${index} failed with URL ${url}:`, error)
        continue
      }
    }
    
    console.error(`❌ All URLs failed for worker ${index}`)
    return null
  }

  /**
   * Try to create a single worker instance
   */
  private async tryCreateSingleWorker(url: string, index: number): Promise<Worker | null> {
    try {
      const options: WorkerOptions = { 
        type: url.endsWith('.ts') ? 'module' : 'classic',
        name: `image-worker-${index}`
      }
      
      const worker = new Worker(url, options)
      
      // Test the worker with a timeout
      const testResult = await this.testWorker(worker, 3000)
      if (testResult) {
        console.log(`✅ Worker ${index} created successfully with URL: ${url}`)
        return worker
      } else {
        worker.terminate()
        console.warn(`❌ Worker ${index} failed test with URL: ${url}`)
        return null
      }
    } catch (error) {
      console.warn(`❌ Worker ${index} creation failed with URL ${url}:`, error)
      return null
    }
  }

  /**
   * Resolve glob patterns to actual URLs
   */
  private async resolveGlobPattern(baseUrl: string): Promise<string[]> {
    const urls: string[] = []
    
    // Try common hash patterns for Vite builds
    const commonHashes = [
      'a1b2c3d4', 'e5f6g7h8', '12345678', 'abcdef12',
      '9876543210ab', 'fedcba0987'
    ]
    
    for (const hash of commonHashes) {
      urls.push(`${baseUrl}${hash}.js`)
    }
    
    return urls
  }

  /**
   * Test if a worker is responsive
   */
  private testWorker(worker: Worker, timeout: number = 3000): Promise<boolean> {
    return new Promise((resolve) => {
      let resolved = false
      const testId = `test-${Date.now()}-${Math.random()}`
      
      const timer = setTimeout(() => {
        if (!resolved) {
          resolved = true
          resolve(false)
        }
      }, timeout)
      
      const onMessage = (event: MessageEvent) => {
        if (event.data?.id === testId && !resolved) {
          resolved = true
          clearTimeout(timer)
          worker.removeEventListener('message', onMessage)
          worker.removeEventListener('error', onError)
          resolve(true)
        }
      }
      
      const onError = () => {
        if (!resolved) {
          resolved = true
          clearTimeout(timer)
          worker.removeEventListener('message', onMessage)
          worker.removeEventListener('error', onError)
          resolve(false)
        }
      }
      
      worker.addEventListener('message', onMessage)
      worker.addEventListener('error', onError)
      
      try {
        worker.postMessage({ id: testId, type: 'test' })
      } catch (error) {
        console.warn('Failed to post test message to worker:', error)
        onError()
      }
    })
  }

  /**
   * Get possible worker URLs based on environment
   */
  private getWorkerUrls(): string[] {
    const urls: string[] = []
    
    // For development (Vite dev server)
    if (this.environmentInfo.isDevelopment) {
      urls.push('/src/workers/imageProcessingWorker.ts')
    }
    
    // For production build - try multiple common paths with better detection
    const workerPaths = [
      // Standard Vite build paths
      '/assets/imageProcessingWorker-*.js',
      '/workers/imageProcessingWorker-*.js',
      '/assets/imageProcessingWorker.js',
      '/workers/imageProcessingWorker.js',
      
      // Alternative build paths
      './assets/imageProcessingWorker-*.js',
      './workers/imageProcessingWorker-*.js',
      './imageProcessingWorker-*.js',
      
      // Relative paths from current location
      'assets/imageProcessingWorker-*.js',
      'workers/imageProcessingWorker-*.js',
      'imageProcessingWorker-*.js'
    ]
    
    urls.push(...workerPaths)
    
    // Fallback inline worker (always works)
    urls.push('data:text/javascript;base64,' + this.btoa(this.getInlineWorkerCode()))
    
    return urls
  }

  /**
   * Browser-compatible btoa implementation
   */
  private btoa(str: string): string {
    if (typeof globalThis.btoa !== 'undefined') {
      return globalThis.btoa(str)
    }
    // Fallback for environments without btoa
    return Buffer.from(str, 'binary').toString('base64')
  }

  /**
   * Detect current environment
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
    const isLocalhost = window.location.hostname === 'localhost' || 
                       window.location.hostname === '127.0.0.1' || 
                       window.location.hostname === '0.0.0.0'
    const isDevelopment = isLocalhost || 
                         window.location.hostname.includes('dev') ||
                         window.location.port === '3000' || // Vite dev server
                         window.location.port === '5173'    // Alternative Vite port
    const isProduction = !isDevelopment && !isFileProtocol

    return {
      isProduction,
      isDevelopment,
      isFileProtocol,
      hostname: window.location.hostname,
      port: window.location.port || undefined
    }
  }

  /**
   * Detect worker capabilities
   */
  private detectCapabilities(): WorkerCapabilities {
    return {
      hasOffscreenCanvas: typeof OffscreenCanvas !== 'undefined',
      hasImageBitmap: typeof ImageBitmap !== 'undefined',
      hasCreateImageBitmap: typeof createImageBitmap !== 'undefined',
      hasArrayBuffer: typeof ArrayBuffer !== 'undefined',
      hasUint8ClampedArray: typeof Uint8ClampedArray !== 'undefined',
      hardwareConcurrency: navigator.hardwareConcurrency || 4
    }
  }

  /**
   * Get inline worker code as fallback
   */
  private getInlineWorkerCode(): string {
    return `
      console.log('🔧 Enhanced fallback worker starting...');
      
      function simpleBinarization(imageData, threshold = 128) {
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
        return canvas.convertToBlob().then(blob => blob.arrayBuffer());
      }
      
      function simpleScale(imageData, factor) {
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
        
        return scaledCanvas.convertToBlob().then(blob => blob.arrayBuffer());
      }
      
      function simpleMorphology(imageData, operation) {
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
        return canvas.convertToBlob().then(blob => blob.arrayBuffer());
      }
      
      function simpleNoiseReduction(imageData) {
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
        return canvas.convertToBlob().then(blob => blob.arrayBuffer());
      }
      
      self.onmessage = async function(event) {
        const { id, type, payload } = event.data;
        
        console.log('🔧 Enhanced fallback worker received:', type, 'for task:', id);
        
        if (type === 'test') {
          self.postMessage({ id, type: 'test-response' });
          return;
        }
        
        if (type === 'process') {
          try {
            const { imageData, type: processType, parameters } = payload;
            
            self.postMessage({
              id, type: 'progress',
              payload: { progress: 25, message: 'Processing with enhanced fallback worker...' }
            });
            
            let result;
            
            switch (processType) {
              case 'binarization':
                const threshold = parameters.binarization?.threshold || 128;
                result = await simpleBinarization(imageData, threshold);
                break;
                
              case 'scaling':
                const factor = parameters.scaling?.factor || 2;
                result = await simpleScale(imageData, factor);
                break;
                
              case 'morphology':
                const operation = parameters.morphology?.operation || 'opening';
                result = await simpleMorphology(imageData, operation);
                break;
                
              case 'noise-reduction':
                result = await simpleNoiseReduction(imageData);
                break;
                
              default:
                result = imageData.data.slice(0);
            }
            
            self.postMessage({
              id, type: 'progress',
              payload: { progress: 75, message: 'Finalizing...' }
            });
            
            self.postMessage({
              id, type: 'result',
              payload: { result }
            }, result instanceof ArrayBuffer ? [result] : []);
            
          } catch (error) {
            console.error('🔧 Enhanced fallback worker error:', error);
            self.postMessage({
              id, type: 'error',
              payload: { error: 'Enhanced fallback processing failed: ' + error.message }
            });
          }
        }
      };
      
      console.log('🔧 Enhanced fallback worker initialized and ready');
    `
  }

  getEnvironmentInfo(): WorkerEnvironmentInfo {
    return { ...this.environmentInfo }
  }

  getCapabilities(): WorkerCapabilities {
    return { ...this.capabilities }
  }

  getRecommendedWorkerCount(): number {
    return Math.min(this.capabilities.hardwareConcurrency, 8)
  }
}