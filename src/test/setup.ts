import { beforeAll, vi } from 'vitest'

// Mock Web APIs that might not be available in test environment
beforeAll(() => {
  // Mock File API
  if (typeof globalThis.File === 'undefined') {
    globalThis.File = class MockFile {
      name: string
      size: number
      type: string
      
      constructor(chunks: any[], filename: string, options: any = {}) {
        this.name = filename
        this.size = chunks.reduce((size, chunk) => size + chunk.length, 0)
        this.type = options.type || ''
      }
    } as any
  }

  // Mock FileReader
  if (typeof globalThis.FileReader === 'undefined') {
    globalThis.FileReader = class MockFileReader {
      result: any = null
      onload: ((event: any) => void) | null = null
      onerror: ((event: any) => void) | null = null
      
      readAsArrayBuffer(_file: File) {
        setTimeout(() => {
          this.result = new ArrayBuffer(8)
          if (this.onload) {
            this.onload({ target: { result: this.result } })
          }
        }, 0)
      }
    } as any
  }

  // Mock URL API
  if (typeof globalThis.URL === 'undefined') {
    globalThis.URL = {
      createObjectURL: () => 'mock-url',
      revokeObjectURL: () => {}
    } as any
  }

  // Mock Canvas API
  const mockContext = {
    clearRect: vi.fn(),
    drawImage: vi.fn(),
    getImageData: vi.fn(() => ({
      data: new Uint8ClampedArray(4),
      width: 1,
      height: 1
    })),
    putImageData: vi.fn(),
    fillRect: vi.fn(),
    strokeRect: vi.fn(),
    beginPath: vi.fn(),
    closePath: vi.fn(),
    stroke: vi.fn(),
    fill: vi.fn()
  }

  if (typeof globalThis.HTMLCanvasElement === 'undefined') {
    globalThis.HTMLCanvasElement = class MockCanvas {
      width = 300
      height = 150
      
      getContext() {
        return mockContext
      }
      
      toDataURL() {
        return 'data:image/png;base64,mock-data'
      }
    } as any
  }

  // Mock Image
  if (typeof globalThis.Image === 'undefined') {
    globalThis.Image = class MockImage {
      onload: (() => void) | null = null
      onerror: (() => void) | null = null
      width = 100
      height = 100
      
      set src(_value: string) {
        setTimeout(() => {
          if (this.onload) this.onload()
        }, 0)
      }
      
      get src() {
        return ''
      }
    } as any
  }

  // Mock Worker
  if (typeof globalThis.Worker === 'undefined') {
    globalThis.Worker = class MockWorker {
      onmessage: ((event: MessageEvent) => void) | null = null
      onerror: ((event: ErrorEvent) => void) | null = null
      
      constructor(_url: string | URL, _options?: WorkerOptions) {
        // Mock worker constructor
      }
      
      postMessage(message: any, _transfer?: Transferable[]) {
        // Mock postMessage
        setTimeout(() => {
          if (this.onmessage) {
            this.onmessage(new MessageEvent('message', { data: { 
              id: message.id, 
              type: 'result', 
              payload: { result: new ArrayBuffer(8) } 
            }}))
          }
        }, 0)
      }
      
      terminate() {
        // Mock terminate
      }
    } as any
  }

  // Mock performance.memory
  if (!globalThis.performance) {
    globalThis.performance = {} as Performance
  }
  
  if (!(globalThis.performance as any).memory) {
    Object.defineProperty(globalThis.performance, 'memory', {
      value: {
        usedJSHeapSize: 10 * 1024 * 1024,
        totalJSHeapSize: 50 * 1024 * 1024,
        jsHeapSizeLimit: 100 * 1024 * 1024
      },
      writable: true
    })
  }

  // Mock navigator.hardwareConcurrency
  if (!navigator.hardwareConcurrency) {
    Object.defineProperty(navigator, 'hardwareConcurrency', {
      writable: true,
      value: 4
    })
  }

  // Mock Blob
  if (typeof globalThis.Blob === 'undefined') {
    globalThis.Blob = class MockBlob {
      size: number
      type: string
      
      constructor(chunks: any[] = [], options: any = {}) {
        this.size = chunks.reduce((size, chunk) => size + chunk.length, 0)
        this.type = options.type || ''
      }
    } as any
  }

  // Mock ResizeObserver
  if (typeof globalThis.ResizeObserver === 'undefined') {
    globalThis.ResizeObserver = class MockResizeObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    } as any
  }

  // Mock IntersectionObserver
  if (typeof globalThis.IntersectionObserver === 'undefined') {
    globalThis.IntersectionObserver = class MockIntersectionObserver {
      observe() {}
      unobserve() {}
      disconnect() {}
    } as any
  }
})