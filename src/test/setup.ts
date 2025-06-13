// src/test/setup.ts
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
      readonly EMPTY = 0
      readonly LOADING = 1
      readonly DONE = 2
      readonly readyState = 0
      
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
    fill: vi.fn(),
    createImageData: vi.fn(() => ({
      data: new Uint8ClampedArray(4),
      width: 1,
      height: 1
    })),
    imageSmoothingEnabled: true,
    imageSmoothingQuality: 'high' as const
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

  // Mock OffscreenCanvas
  if (typeof globalThis.OffscreenCanvas === 'undefined') {
    globalThis.OffscreenCanvas = class MockOffscreenCanvas {
      width: number
      height: number
      
      constructor(width: number, height: number) {
        this.width = width
        this.height = height
      }
      
      getContext() {
        return mockContext
      }
      
      convertToBlob() {
        return Promise.resolve(new Blob(['mock-data'], { type: 'image/png' }))
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
      naturalWidth = 100
      naturalHeight = 100
      
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

  // Mock Worker with enhanced error handling
  if (typeof globalThis.Worker === 'undefined') {
    globalThis.Worker = class MockWorker {
      onmessage: ((event: MessageEvent) => void) | null = null
      onerror: ((event: ErrorEvent) => void) | null = null
      onmessageerror: ((event: MessageEvent) => void) | null = null
      
      constructor(_url: string | URL, _options?: WorkerOptions) {
        // Mock worker constructor
      }
      
      postMessage(message: any, _transfer?: Transferable[]) {
        // Mock postMessage with enhanced response
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

  // Mock MessageEvent
  if (typeof globalThis.MessageEvent === 'undefined') {
    globalThis.MessageEvent = class MockMessageEvent {
      data: any
      type: string
      lastEventId = ''
      origin = 'http://localhost'
      ports: readonly MessagePort[] = []
      source: MessageEventSource | null = null
      bubbles = false
      cancelable = false
      composed = false
      currentTarget = null
      defaultPrevented = false
      eventPhase = 0
      isTrusted = true
      target = null
      timeStamp = Date.now()
      
      constructor(type: string, eventInitDict?: MessageEventInit) {
        this.type = type
        this.data = eventInitDict?.data
        this.lastEventId = eventInitDict?.lastEventId || ''
        this.origin = eventInitDict?.origin || 'http://localhost'
        this.ports = eventInitDict?.ports || []
        this.source = eventInitDict?.source || null
      }
      
      preventDefault() {}
      stopPropagation() {}
      stopImmediatePropagation() {}
      initEvent() {}
    } as any
  }

  // Mock ErrorEvent
  if (typeof globalThis.ErrorEvent === 'undefined') {
    globalThis.ErrorEvent = class MockErrorEvent {
      message: string
      filename?: string
      lineno?: number
      colno?: number
      error?: any
      type: string
      bubbles = false
      cancelable = false
      composed = false
      currentTarget = null
      defaultPrevented = false
      eventPhase = 0
      isTrusted = true
      target = null
      timeStamp = Date.now()
      
      constructor(type: string, eventInitDict?: ErrorEventInit) {
        this.type = type
        this.message = eventInitDict?.message || ''
        this.filename = eventInitDict?.filename
        this.lineno = eventInitDict?.lineno
        this.colno = eventInitDict?.colno
        this.error = eventInitDict?.error
      }
      
      preventDefault() {}
      stopPropagation() {}
      stopImmediatePropagation() {}
      initEvent() {}
    } as any
  }

  // Mock PromiseRejectionEvent
  if (typeof globalThis.PromiseRejectionEvent === 'undefined') {
    globalThis.PromiseRejectionEvent = class MockPromiseRejectionEvent {
      promise: Promise<any>
      reason: any
      type: string
      bubbles = false
      cancelable = false
      composed = false
      currentTarget = null
      defaultPrevented = false
      eventPhase = 0
      isTrusted = true
      target = null
      timeStamp = Date.now()
      
      constructor(type: string, eventInitDict: PromiseRejectionEventInit) {
        this.type = type
        this.promise = eventInitDict.promise
        this.reason = eventInitDict.reason
      }
      
      preventDefault() {}
      stopPropagation() {}
      stopImmediatePropagation() {}
      initEvent() {}
    } as any
  }

  // Mock DragEvent (for testing drag and drop)
  if (typeof globalThis.DragEvent === 'undefined') {
    globalThis.DragEvent = class MockDragEvent {
      dataTransfer: DataTransfer | null = null
      type: string
      bubbles = false
      cancelable = false
      composed = false
      currentTarget = null
      defaultPrevented = false
      eventPhase = 0
      isTrusted = true
      target = null
      timeStamp = Date.now()
      
      constructor(type: string, eventInitDict?: DragEventInit) {
        this.type = type
        this.dataTransfer = eventInitDict?.dataTransfer || null
      }
      
      preventDefault() {}
      stopPropagation() {}
      stopImmediatePropagation() {}
      initEvent() {}
    } as any
  }

  // Mock DataTransfer
  if (typeof globalThis.DataTransfer === 'undefined') {
    globalThis.DataTransfer = class MockDataTransfer {
      dropEffect: 'none' | 'copy' | 'link' | 'move' = 'none'
      effectAllowed: string = 'uninitialized'
      files: FileList = [] as any
      items: DataTransferItemList = [] as any
      types: readonly string[] = []
      
      clearData() {}
      getData() { return '' }
      setData() {}
      setDragImage() {}
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
      
      arrayBuffer() {
        return Promise.resolve(new ArrayBuffer(this.size))
      }
      
      text() {
        return Promise.resolve('mock-text')
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

  // Mock ArrayBuffer methods
  if (typeof ArrayBuffer.prototype.slice === 'undefined') {
    ArrayBuffer.prototype.slice = function(start?: number, end?: number) {
      const length = this.byteLength
      const relativeStart = start === undefined ? 0 : start
      const relativeEnd = end === undefined ? length : end
      
      const actualStart = relativeStart < 0 ? Math.max(length + relativeStart, 0) : Math.min(relativeStart, length)
      const actualEnd = relativeEnd < 0 ? Math.max(length + relativeEnd, 0) : Math.min(relativeEnd, length)
      
      const size = Math.max(actualEnd - actualStart, 0)
      const copy = new ArrayBuffer(size)
      
      return copy
    }
  }

  // Mock Uint8ClampedArray for ImageData
  if (typeof globalThis.Uint8ClampedArray === 'undefined') {
    globalThis.Uint8ClampedArray = class MockUint8ClampedArray extends Uint8Array {
      constructor(length: number | ArrayBufferLike | ArrayLike<number>) {
        super(length as any)
      }
    } as any
  }

  // Mock ImageData constructor
  if (typeof globalThis.ImageData === 'undefined') {
    globalThis.ImageData = class MockImageData {
      data: Uint8ClampedArray
      width: number
      height: number
      
      constructor(data: Uint8ClampedArray, width: number, height?: number)
      constructor(width: number, height: number)
      constructor(dataOrWidth: Uint8ClampedArray | number, widthOrHeight: number, height?: number) {
        if (typeof dataOrWidth === 'number') {
          this.width = dataOrWidth
          this.height = widthOrHeight
          this.data = new Uint8ClampedArray(dataOrWidth * widthOrHeight * 4)
        } else {
          this.data = dataOrWidth
          this.width = widthOrHeight
          this.height = height || (dataOrWidth.length / widthOrHeight / 4)
        }
      }
    } as any
  }

  // Mock createImageBitmap
  if (typeof globalThis.createImageBitmap === 'undefined') {
    globalThis.createImageBitmap = async function(_source: any, _options?: any): Promise<ImageBitmap> {
      return {
        width: 100,
        height: 100,
        close: () => {}
      } as ImageBitmap
    }
  }

  // Mock ImageBitmap
  if (typeof globalThis.ImageBitmap === 'undefined') {
    globalThis.ImageBitmap = class MockImageBitmap {
      width = 100
      height = 100
      close() {}
    } as any
  }
})