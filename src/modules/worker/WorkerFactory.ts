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
          continue // Skip wildcard URLs for now
        }
        
        const options: WorkerOptions = { 
          type: url.endsWith('.ts') ? 'module' : 'classic',
          name: `image-worker-${index}`
        }
        
        const worker = new Worker(url, options)
        
        // Test the worker with a timeout
        const testResult = await this.testWorker(worker, 3000)
        if (testResult) {
          console.log(`✅ Worker ${index} created successfully`)
          return worker
        } else {
          worker.terminate()
        }
      } catch (_error) {
        console.warn(`❌ Worker ${index} failed with URL ${url}:`, _error)
        continue
      }
    }
    
    console.error(`❌ All URLs failed for worker ${index}`)
    return null
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
    
    // For production build - try multiple common paths
    urls.push('/workers/imageProcessingWorker.js')
    urls.push('/assets/imageProcessingWorker.js')
    urls.push('/assets/imageProcessingWorker-*.js')
    urls.push('/workers/imageProcessingWorker-*.js')
    
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
    const isLocalhost = window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1'
    const isDevelopment = isLocalhost || window.location.hostname.includes('dev')
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
      console.log('🔧 Inline fallback worker starting...');
      
      function simpleBinarization(imageData, threshold = 128) {
        const data = new Uint8ClampedArray(imageData.data);
        
        for (let i = 0; i < data.length; i += 4) {
          const gray = data[i] * 0.299 + data[i + 1] * 0.587 + data[i + 2] * 0.114;
          const binary = gray > threshold ? 255 : 0;
          data[i] = binary;
          data[i + 1] = binary;
          data[i + 2] = binary;
        }
        
        return data.buffer;
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
      
      self.onmessage = async function(event) {
        const { id, type, payload } = event.data;
        
        console.log('🔧 Fallback worker received:', type, 'for task:', id);
        
        if (type === 'test') {
          self.postMessage({ id, type: 'test-response' });
          return;
        }
        
        if (type === 'process') {
          try {
            const { imageData, type: processType, parameters } = payload;
            
            self.postMessage({
              id, type: 'progress',
              payload: { progress: 25, message: 'Processing with fallback worker...' }
            });
            
            let result;
            
            switch (processType) {
              case 'binarization':
                const threshold = parameters.binarization?.threshold || 128;
                result = simpleBinarization(imageData, threshold);
                break;
                
              case 'scaling':
                const factor = parameters.scaling?.factor || 2;
                result = await simpleScale(imageData, factor);
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
            console.error('🔧 Fallback worker error:', error);
            self.postMessage({
              id, type: 'error',
              payload: { error: 'Fallback processing failed: ' + error.message }
            });
          }
        }
      };
      
      console.log('🔧 Inline fallback worker initialized and ready');
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