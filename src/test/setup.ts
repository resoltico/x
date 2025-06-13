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
        this.size = chunks.reduce((size, chunk) => size + (typeof chunk === 'string' ? chunk.length : chunk.byteLength || 0), 0)
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
      
      readAsArrayBuffer(file: File) {
        setTimeout(() => {
          this.result = new ArrayBuffer(file.size || 8)
          if (this.onload) {
            this.onload({ target: { result: this.result } })
          }
        }, 0)
      }
      
      readAsDataURL(file: File) {
        setTimeout(() => {
          this.result = `data:${file.type};base64,dGVzdA==`
          if (this.onload) {
            this.onload({ target: { result: this.result } })
          }
        }, 0)
      }
    } as any
  }

  // Mock URL API with proper implementation
  if (typeof globalThis.URL === 'undefined' || !globalThis.URL.createObjectURL) {
    const mockURLs = new Set<string>()
    
    globalThis.URL = {
      createObjectURL: (_object: File | Blob | MediaSource) => {
        const mockUrl = `blob:mock-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
        mockURLs.add(mockUrl)
        return mockUrl
      },
      revokeObjectURL: (url: string) => {
        mockURLs.delete(url)
      }
    } as any
  }

  // Mock Canvas API with comprehensive implementation
  const mockContext = {
    canvas: null as any,
    clearRect: vi.fn(),
    drawImage: vi.fn(),
    getImageData: vi.fn(() => ({
      data: new Uint8ClampedArray(400), // 10x10 RGBA
      width: 10,
      height: 10
    })),
    putImageData: vi.fn(),
    createImageData: vi.fn((width: number, height: number) => ({
      data: new Uint8ClampedArray(width * height * 4),
      width,
      height
    })),
    fillRect: vi.fn(),
    strokeRect: vi.fn(),
    beginPath: vi.fn(),
    closePath: vi.fn(),
    stroke: vi.fn(),
    fill: vi.fn(),
    arc: vi.fn(),
    moveTo: vi.fn(),
    lineTo: vi.fn(),
    save: vi.fn(),
    restore: vi.fn(),
    translate: vi.fn(),
    rotate: vi.fn(),
    scale: vi.fn(),
    transform: vi.fn(),
    setTransform: vi.fn(),
    resetTransform: vi.fn(),
    imageSmoothingEnabled: true,
    imageSmoothingQuality: 'high' as const,
    fillStyle: '#000000',
    strokeStyle: '#000000',
    lineWidth: 1,
    lineCap: 'butt' as const,
    lineJoin: 'miter' as const,
    miterLimit: 10,
    shadowBlur: 0,
    shadowColor: 'rgba(0, 0, 0, 0)',
    shadowOffsetX: 0,
    shadowOffsetY: 0,
    font: '10px sans-serif',
    textAlign: 'start' as const,
    textBaseline: 'alphabetic' as const,
    globalAlpha: 1,
    globalCompositeOperation: 'source-over' as const,
    measureText: vi.fn(() => ({ width: 50 })),
    fillText: vi.fn(),
    strokeText: vi.fn()
  }

  // Mock HTMLCanvasElement
  if (typeof globalThis.HTMLCanvasElement === 'undefined') {
    globalThis.HTMLCanvasElement = class MockCanvas {
      width = 300
      height = 150
      
      getContext(contextType: string) {
        if (contextType === '2d') {
          const ctx = { ...mockContext }
          ctx.canvas = this
          return ctx
        }
        return null
      }
      
      toDataURL(type = 'image/png') {
        return `data:${type};base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGAWjR9awAAAABJRU5ErkJggg==`
      }
      
      toBlob(callback: (blob: Blob) => void, type = 'image/png') {
        setTimeout(() => {
          callback(new Blob(['mock-canvas-data'], { type }))
        }, 0)
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
      
      getContext(contextType: string) {
        if (contextType === '2d') {
          const ctx = { ...mockContext }
          ctx.canvas = this
          return ctx
        }
        return null
      }
      
      convertToBlob(_options: any = {}) {
        return Promise.resolve(new Blob(['mock-offscreen-canvas-data'], { 
          type: _options.type || 'image/png' 
        }))
      }
      
      transferToImageBitmap() {
        return {
          width: this.width,
          height: this.height,
          close: vi.fn()
        } as ImageBitmap
      }
    } as any
  }

  // Mock Image
  if (typeof globalThis.Image === 'undefined') {
    globalThis.Image = class MockImage {
      onload: (() => void) | null = null
      onerror: (() => void) | null = null
      onabort: (() => void) | null = null
      width = 100
      height = 100
      naturalWidth = 100
      naturalHeight = 100
      complete = false
      crossOrigin: string | null = null
      loading: 'eager' | 'lazy' = 'eager'
      referrerPolicy = ''
      sizes = ''
      srcset = ''
      useMap = ''
      
      private _src = ''
      
      set src(value: string) {
        this._src = value
        // Simulate async loading
        setTimeout(() => {
          this.complete = true
          if (this.onload) this.onload()
        }, 0)
      }
      
      get src() {
        return this._src
      }
      
      decode() {
        return Promise.resolve()
      }
    } as any
  }

  // Mock Worker with enhanced functionality
  if (typeof globalThis.Worker === 'undefined') {
    globalThis.Worker = class MockWorker {
      onmessage: ((event: MessageEvent) => void) | null = null
      onerror: ((event: ErrorEvent) => void) | null = null
      onmessageerror: ((event: MessageEvent) => void) | null = null
      
      constructor(private url: string | URL, private _options?: WorkerOptions) {
        // Mock worker constructor
        console.log(`Mock Worker created for: ${url}`)
      }
      
      postMessage(message: any, _transfer?: Transferable[]) {
        // Mock postMessage with realistic delay and response
        setTimeout(() => {
          if (this.onmessage) {
            const responseData = {
              id: message.id,
              type: 'result',
              payload: { 
                result: new ArrayBuffer(1024) // Mock processed data
              }
            }
            this.onmessage(new MessageEvent('message', { data: responseData }))
          }
        }, 10) // Small delay to simulate processing
      }
      
      terminate() {
        // Mock terminate
        this.onmessage = null
        this.onerror = null
        this.onmessageerror = null
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
        if (eventInitDict) {
          this.data = eventInitDict.data
          this.lastEventId = eventInitDict.lastEventId || ''
          this.origin = eventInitDict.origin || 'http://localhost'
          this.ports = eventInitDict.ports || []
          this.source = eventInitDict.source || null
        }
      }
      
      preventDefault() {}
      stopPropagation() {}
      stopImmediatePropagation() {}
      initEvent() {}
    } as any
  }

  // Fixed performance.memory mock with proper type assertion
  if (!globalThis.performance) {
    globalThis.performance = {
      now: vi.fn(() => Date.now()),
      mark: vi.fn(),
      measure: vi.fn(),
      clearMarks: vi.fn(),
      clearMeasures: vi.fn(),
      getEntries: vi.fn(() => []),
      getEntriesByName: vi.fn(() => []),
      getEntriesByType: vi.fn(() => []),
      timeOrigin: Date.now(),
      // Fix: Properly define toJSON to avoid type errors
      toJSON: vi.fn(() => ({}))
    } as unknown as Performance
  }
  
  // Mock performance.memory with proper typing
  if (!(globalThis.performance as any).memory) {
    Object.defineProperty(globalThis.performance, 'memory', {
      value: {
        usedJSHeapSize: 10 * 1024 * 1024,
        totalJSHeapSize: 50 * 1024 * 1024,
        jsHeapSizeLimit: 100 * 1024 * 1024
      },
      writable: true,
      configurable: true
    })
  }

  // Mock navigator.hardwareConcurrency
  if (!navigator.hardwareConcurrency) {
    Object.defineProperty(navigator, 'hardwareConcurrency', {
      writable: true,
      value: 4,
      configurable: true
    })
  }

  // Mock Blob with enhanced functionality
  if (typeof globalThis.Blob === 'undefined') {
    globalThis.Blob = class MockBlob {
      size: number
      type: string
      
      constructor(chunks: any[] = [], _options: any = {}) {
        this.size = chunks.reduce((size, chunk) => {
          if (typeof chunk === 'string') return size + chunk.length
          if (chunk instanceof ArrayBuffer) return size + chunk.byteLength
          if (ArrayBuffer.isView(chunk)) return size + chunk.byteLength
          return size + String(chunk).length
        }, 0)
        this.type = _options.type || ''
      }
      
      arrayBuffer() {
        return Promise.resolve(new ArrayBuffer(this.size))
      }
      
      text() {
        return Promise.resolve('mock-blob-text')
      }
      
      stream() {
        // Mock ReadableStream
        return {
          getReader: () => ({
            read: () => Promise.resolve({ done: true, value: undefined }),
            releaseLock: () => {},
            closed: Promise.resolve(undefined)
          }),
          cancel: () => Promise.resolve(),
          locked: false
        }
      }
      
      slice(start?: number, end?: number, contentType?: string) {
        return new MockBlob(['sliced'], { type: contentType || this.type })
      }
    } as any
  }

  // Mock createImageBitmap
  if (typeof globalThis.createImageBitmap === 'undefined') {
    globalThis.createImageBitmap = async function(_source: any, _options?: any): Promise<ImageBitmap> {
      // Simulate async operation
      await new Promise(resolve => setTimeout(resolve, 1))
      
      return {
        width: 100,
        height: 100,
        close: vi.fn()
      } as ImageBitmap
    }
  }

  // Mock ImageBitmap
  if (typeof globalThis.ImageBitmap === 'undefined') {
    globalThis.ImageBitmap = class MockImageBitmap {
      width = 100
      height = 100
      close = vi.fn()
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

  // Mock DragEvent and related APIs
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
        if (eventInitDict) {
          this.dataTransfer = eventInitDict.dataTransfer || null
        }
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
      items: DataTransferItemList = {
        length: 0,
        add: vi.fn(),
        clear: vi.fn(),
        remove: vi.fn()
      } as any
      types: readonly string[] = []
      
      clearData() {}
      getData() { return '' }
      setData() {}
      setDragImage() {}
    } as any
  }

  // Mock ResizeObserver
  if (typeof globalThis.ResizeObserver === 'undefined') {
    globalThis.ResizeObserver = class MockResizeObserver {
      observe = vi.fn()
      unobserve = vi.fn()
      disconnect = vi.fn()
    } as any
  }

  // Mock IntersectionObserver
  if (typeof globalThis.IntersectionObserver === 'undefined') {
    globalThis.IntersectionObserver = class MockIntersectionObserver {
      observe = vi.fn()
      unobserve = vi.fn()
      disconnect = vi.fn()
    } as any
  }

  // Mock document.createElement for canvas
  const originalCreateElement = document.createElement.bind(document)
  document.createElement = vi.fn((tagName: string): HTMLElement => {
    if (tagName.toLowerCase() === 'canvas') {
      // Create a mock canvas element that properly extends HTMLElement
      const mockCanvas = originalCreateElement('div') as any
      
      // Use Object.defineProperty to properly set canvas properties
      Object.defineProperty(mockCanvas, 'width', {
        value: 300,
        writable: true,
        configurable: true
      })
      
      Object.defineProperty(mockCanvas, 'height', {
        value: 150,
        writable: true,
        configurable: true
      })
      
      Object.defineProperty(mockCanvas, 'getContext', {
        value: vi.fn((contextType: string) => {
          if (contextType === '2d') {
            const ctx = { ...mockContext }
            ctx.canvas = mockCanvas
            return ctx
          }
          return null
        }),
        writable: true,
        configurable: true
      })
      
      Object.defineProperty(mockCanvas, 'toDataURL', {
        value: vi.fn((type = 'image/png') => {
          return `data:${type};base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGAWjR9awAAAABJRU5ErkJggg==`
        }),
        writable: true,
        configurable: true
      })
      
      Object.defineProperty(mockCanvas, 'toBlob', {
        value: vi.fn((callback: (blob: Blob) => void, type = 'image/png') => {
          setTimeout(() => {
            callback(new Blob(['mock-canvas-data'], { type }))
          }, 0)
        }),
        writable: true,
        configurable: true
      })
      
      Object.defineProperty(mockCanvas, 'getBoundingClientRect', {
        value: vi.fn(() => ({
          left: 0,
          top: 0,
          right: 300,
          bottom: 150,
          width: 300,
          height: 150,
          x: 0,
          y: 0,
          toJSON: () => ({})
        })),
        writable: true,
        configurable: true
      })
      
      return mockCanvas
    }
    return originalCreateElement(tagName)
  })

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
      
      // Copy data (simplified)
      const srcView = new Uint8Array(this, actualStart, size)
      const dstView = new Uint8Array(copy)
      dstView.set(srcView)
      
      return copy
    }
  }

  // Suppress console errors for expected test scenarios
  const originalConsoleError = console.error
  console.error = (...args: any[]) => {
    const message = args[0]
    if (typeof message === 'string') {
      // Suppress specific expected errors during tests
      if (message.includes('Not implemented: HTMLCanvasElement.prototype.getContext') ||
          message.includes('ENOENT: no such file or directory') ||
          message.includes('wasm-vips') ||
          message.includes('Failed to initialize ProcessingModule')) {
        return
      }
    }
    originalConsoleError(...args)
  }

  // Mock window.fs for file operations (if needed)
  if (typeof (globalThis as any).window === 'undefined') {
    (globalThis as any).window = globalThis
  }
  
  if (!(globalThis as any).window.fs) {
    (globalThis as any).window.fs = {
      readFile: vi.fn().mockResolvedValue(new ArrayBuffer(1024))
    }
  }
})