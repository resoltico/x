// src/utils/debugLogger.ts
// Enhanced debug logger with worker diagnostics

export interface DebugEvent {
  timestamp: string
  level: 'info' | 'warn' | 'error' | 'debug'
  category: string
  message: string
  data?: any
  stack?: string
}

class DebugLogger {
  private events: DebugEvent[] = []
  private maxEvents = 1000
  private listeners: Array<(event: DebugEvent) => void> = []

  log(level: DebugEvent['level'], category: string, message: string, data?: any, error?: Error) {
    const event: DebugEvent = {
      timestamp: new Date().toISOString(),
      level,
      category,
      message,
      data: data ? this.sanitizeData(data) : undefined,
      stack: error?.stack
    }

    this.events.push(event)
    if (this.events.length > this.maxEvents) {
      this.events.shift()
    }

    // Console output with formatting
    const prefix = `[${event.timestamp.slice(11, 23)}] [${category.toUpperCase()}]`
    const fullMessage = `${prefix} ${message}`

    switch (level) {
      case 'error':
        console.error(fullMessage, data, error)
        break
      case 'warn':
        console.warn(fullMessage, data)
        break
      case 'debug':
        console.debug(fullMessage, data)
        break
      default:
        console.log(fullMessage, data)
    }

    // Notify listeners
    this.listeners.forEach(listener => {
      try {
        listener(event)
      } catch (err) {
        console.error('Debug listener error:', err)
      }
    })
  }

  private sanitizeData(data: any): any {
    try {
      // Handle circular references and large objects
      return JSON.parse(JSON.stringify(data, (key, value) => {
        if (value instanceof ArrayBuffer) {
          return `[ArrayBuffer ${value.byteLength} bytes]`
        }
        if (value instanceof Worker) {
          return '[Worker instance]'
        }
        if (typeof value === 'function') {
          return '[Function]'
        }
        return value
      }))
    } catch {
      return String(data)
    }
  }

  getEvents(category?: string, level?: DebugEvent['level']): DebugEvent[] {
    return this.events.filter(event => {
      if (category && event.category !== category) return false
      if (level && event.level !== level) return false
      return true
    })
  }

  addListener(listener: (event: DebugEvent) => void) {
    this.listeners.push(listener)
  }

  removeListener(listener: (event: DebugEvent) => void) {
    const index = this.listeners.indexOf(listener)
    if (index > -1) {
      this.listeners.splice(index, 1)
    }
  }

  clear() {
    this.events = []
  }

  export(): string {
    return JSON.stringify(this.events, null, 2)
  }

  // Worker diagnostics
  async diagnoseWorkerSupport(): Promise<any> {
    const diagnosis = {
      timestamp: new Date().toISOString(),
      environment: this.detectEnvironment(),
      webWorkerSupport: typeof Worker !== 'undefined',
      sharedArrayBufferSupport: typeof SharedArrayBuffer !== 'undefined',
      offscreenCanvasSupport: typeof OffscreenCanvas !== 'undefined',
      imageBitmapSupport: typeof ImageBitmap !== 'undefined',
      createImageBitmapSupport: typeof createImageBitmap !== 'undefined',
      wasmSupport: await this.checkWasmSupport(),
      hardwareConcurrency: navigator.hardwareConcurrency || 'unknown',
      memoryInfo: this.getMemoryInfo(),
      userAgent: navigator.userAgent,
      location: {
        protocol: window.location.protocol,
        hostname: window.location.hostname,
        port: window.location.port,
        pathname: window.location.pathname
      },
      headers: await this.checkCrossOriginHeaders(),
      workerUrls: await this.testWorkerUrls()
    }

    this.log('info', 'diagnostics', 'Worker support diagnosis completed', diagnosis)
    return diagnosis
  }

  private detectEnvironment() {
    const isFileProtocol = window.location.protocol === 'file:'
    const isLocalhost = ['localhost', '127.0.0.1', '0.0.0.0'].includes(window.location.hostname)
    const isDevelopment = isLocalhost || window.location.port === '3000' || window.location.port === '5173'
    const isProduction = !isDevelopment && !isFileProtocol

    return {
      isFileProtocol,
      isLocalhost,
      isDevelopment,
      isProduction,
      port: window.location.port
    }
  }

  private async checkWasmSupport(): Promise<boolean> {
    try {
      // Check if WebAssembly is available
      if (typeof globalThis.WebAssembly === 'undefined') {
        return false
      }

      const wasmCode = new Uint8Array([
        0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00
      ])
      await globalThis.WebAssembly.instantiate(wasmCode)
      return true
    } catch {
      return false
    }
  }

  private getMemoryInfo() {
    const performance = globalThis.performance as any
    if (performance?.memory) {
      return {
        used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024),
        total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024),
        limit: Math.round(performance.memory.jsHeapSizeLimit / 1024 / 1024)
      }
    }
    return null
  }

  private async checkCrossOriginHeaders(): Promise<any> {
    try {
      // Check if fetch is available
      if (typeof globalThis.fetch === 'undefined') {
        return { error: 'fetch not available' }
      }

      const response = await globalThis.fetch(window.location.href, { method: 'HEAD' })
      return {
        coep: response.headers.get('cross-origin-embedder-policy'),
        coop: response.headers.get('cross-origin-opener-policy'),
        status: response.status
      }
    } catch (error) {
      return { error: error instanceof Error ? error.message : 'Unknown error' }
    }
  }

  private async testWorkerUrls(): Promise<any> {
    const workerUrls = [
      '/src/workers/imageProcessingWorker.ts',
      '/workers/imageProcessingWorker.js',
      '/assets/imageProcessingWorker.js',
      '/workers/vips-es6-CoQywvDx.js', // From your logs
      '/public/workers/imageProcessingWorker.js'
    ]

    const results: any = {}

    for (const url of workerUrls) {
      try {
        // Check if fetch is available
        if (typeof globalThis.fetch !== 'undefined') {
          const response = await globalThis.fetch(url, { method: 'HEAD' })
          results[url] = {
            status: response.status,
            contentType: response.headers.get('content-type'),
            accessible: response.ok
          }
        } else {
          results[url] = {
            error: 'fetch not available',
            accessible: false
          }
        }
      } catch (error) {
        results[url] = {
          error: error instanceof Error ? error.message : 'Unknown error',
          accessible: false
        }
      }
    }

    return results
  }

  // Test worker creation with detailed logging
  async testWorkerCreation(url: string): Promise<any> {
    this.log('debug', 'worker-test', `Testing worker creation with URL: ${url}`)
    
    return new Promise((resolve) => {
      const startTime = Date.now()
      let worker: Worker | null = null
      let testCompleted = false

      const complete = (result: any) => {
        if (testCompleted) return
        testCompleted = true
        
        if (worker) {
          try {
            worker.terminate()
          } catch (err) {
            this.log('warn', 'worker-test', 'Error terminating test worker', err)
          }
        }
        
        const duration = Date.now() - startTime
        this.log('debug', 'worker-test', `Worker test completed in ${duration}ms`, result)
        resolve(result)
      }

      try {
        worker = new Worker(url, { type: 'module' })
        
        const timeout = setTimeout(() => {
          complete({
            url,
            success: false,
            error: 'Timeout after 5 seconds',
            duration: Date.now() - startTime
          })
        }, 5000)

        worker.onmessage = (event) => {
          clearTimeout(timeout)
          complete({
            url,
            success: true,
            message: event.data,
            duration: Date.now() - startTime
          })
        }

        worker.onerror = (error) => {
          clearTimeout(timeout)
          complete({
            url,
            success: false,
            error: error.message || 'Worker error',
            duration: Date.now() - startTime
          })
        }

        // Send test message
        worker.postMessage({ id: 'test', type: 'test' })

      } catch (error) {
        complete({
          url,
          success: false,
          error: error instanceof Error ? error.message : 'Unknown error',
          duration: Date.now() - startTime
        })
      }
    })
  }
}

export const debugLogger = new DebugLogger()

// Global access for debugging
if (typeof window !== 'undefined') {
  (window as any).debugLogger = debugLogger
}