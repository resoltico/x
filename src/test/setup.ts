import { beforeAll } from 'vitest'

// Mock Web APIs that might not be available in test environment
beforeAll(() => {
  // Mock File API
  global.File = class MockFile {
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
  global.FileReader = class MockFileReader {
    result: any = null
    onload: ((event: any) => void) | null = null
    onerror: ((event: any) => void) | null = null
    
    readAsArrayBuffer(file: File) {
      setTimeout(() => {
        this.result = new ArrayBuffer(8)
        if (this.onload) {
          this.onload({ target: { result: this.result } })
        }
      }, 0)
    }
  } as any

  // Mock URL API
  global.URL = {
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

  global.HTMLCanvasElement = class MockCanvas {
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
  global.Image = class MockImage {
    onload: (() => void) | null = null
    onerror: (() => void) | null = null
    src = ''
    width = 100
    height = 100
    
    set src(value: string) {
      setTimeout(() => {
        if (this.onload) this.onload()
      }, 0)
    }
  } as any

  // Mock Worker
  global.Worker = class MockWorker {
    onmessage: ((event: MessageEvent) => void) | null = null
    onerror: ((event: ErrorEvent) => void) | null = null
    
    constructor(url: string | URL, options?: WorkerOptions) {
      // Mock worker constructor
    }
    
    postMessage(message: any, transfer?: Transferable[]) {
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
  if (!global.performance) {
    global.performance = {} as Performance
  }
  
  global.performance.memory = {
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
  global.Blob = class MockBlob {
    size: number
    type: string
    
    constructor(chunks: any[] = [], options: any = {}) {
      this.size = chunks.reduce((size, chunk) => size + chunk.length, 0)
      this.type = options.type || ''
    }
  } as any

  // Mock ResizeObserver
  global.ResizeObserver = class MockResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as any

  // Mock IntersectionObserver
  global.IntersectionObserver = class MockIntersectionObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as any
})