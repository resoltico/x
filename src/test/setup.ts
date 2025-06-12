import { beforeAll, vi } from 'vitest'

// Mock Web APIs that might not be available in test environment
beforeAll(() => {
  // Mock File API
  (global as any).File = class MockFile {
    name: string
    size: number
    type: string
    
    constructor(chunks: any[], filename: string, options: any = {}) {
      this.name = filename
      this.size = chunks.reduce((size, chunk) => size + chunk.length, 0)
      this.type = options.type || ''
    }
  } as any

  // Mock FileReader
  (global as any).FileReader = class MockFileReader {
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

  // Mock URL API
  (global as any).URL = {
    createObjectURL: () => 'mock-url',
    revokeObjectURL: () => {}
  } as any

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

  (global as any).HTMLCanvasElement = class MockCanvas {
    width = 300
    height = 150
    
    getContext() {
      return mockContext
    }
    
    toDataURL() {
      return 'data:image/png;base64,mock-data'
    }
  } as any

  // Mock Image
  (global as any).Image = class MockImage {
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

  // Mock Worker
  (global as any).Worker = class MockWorker {
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

  // Mock performance.memory
  if (!(global as any).performance) {
    (global as any).performance = {} as Performance
  }
  
  (global as any).performance.memory = {
    usedJSHeapSize: 10 * 1024 * 1024,
    totalJSHeapSize: 50 * 1024 * 1024,
    jsHeapSizeLimit: 100 * 1024 * 1024
  }

  // Mock navigator.hardwareConcurrency
  Object.defineProperty(navigator, 'hardwareConcurrency', {
    writable: true,
    value: 4
  })

  // Mock Blob
  (global as any).Blob = class MockBlob {
    size: number
    type: string
    
    constructor(chunks: any[] = [], options: any = {}) {
      this.size = chunks.reduce((size, chunk) => size + chunk.length, 0)
      this.type = options.type || ''
    }
  } as any

  // Mock ResizeObserver
  (global as any).ResizeObserver = class MockResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as any

  // Mock IntersectionObserver
  (global as any).IntersectionObserver = class MockIntersectionObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as any
})