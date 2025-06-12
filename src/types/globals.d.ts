// Global type extensions for browser APIs

declare global {
  interface Performance {
    memory?: {
      usedJSHeapSize: number
      totalJSHeapSize: number
      jsHeapSizeLimit: number
    }
  }
  
  interface Navigator {
    hardwareConcurrency?: number
  }

  // Enhanced Canvas API types
  interface CanvasRenderingContext2D {
    imageSmoothingEnabled: boolean
    imageSmoothingQuality?: 'low' | 'medium' | 'high'
  }

  interface OffscreenCanvasRenderingContext2D {
    imageSmoothingEnabled: boolean
    imageSmoothingQuality?: 'low' | 'medium' | 'high'
  }

  // Enhanced Worker types
  interface Worker {
    onmessageerror?: ((this: Worker, ev: MessageEvent) => any) | null
  }

  interface WorkerOptions {
    type?: 'classic' | 'module'
    credentials?: 'omit' | 'same-origin' | 'include'
    name?: string
  }

  // Enhanced File API types
  interface FileReader {
    readonly EMPTY: number
    readonly LOADING: number
    readonly DONE: number
    readonly readyState: number
  }

  // Enhanced Blob types
  interface Blob {
    stream?(): ReadableStream<Uint8Array>
    arrayBuffer(): Promise<ArrayBuffer>
    text(): Promise<string>
  }

  // Enhanced URL types
  interface URL {
    createObjectURL(object: File | Blob | MediaSource): string
    revokeObjectURL(url: string): void
  }

  // Enhanced ImageData types
  interface ImageData {
    readonly data: Uint8ClampedArray
    readonly height: number
    readonly width: number
  }

  // Enhanced Array Buffer types
  interface ArrayBuffer {
    readonly byteLength: number
    slice(begin?: number, end?: number): ArrayBuffer
  }

  // Enhanced Event types for better error handling
  interface ErrorEvent extends Event {
    readonly message: string
    readonly filename?: string
    readonly lineno?: number
    readonly colno?: number
    readonly error?: any
  }

  interface PromiseRejectionEvent extends Event {
    readonly promise: Promise<any>
    readonly reason: any
    preventDefault(): void
  }

  // Enhanced MessageEvent types
  interface MessageEvent<T = any> extends Event {
    readonly data: T
    readonly lastEventId: string
    readonly origin: string
    readonly ports: readonly MessagePort[]
    readonly source: MessageEventSource | null
  }

  // Constants for development/production builds
  const __DEV__: boolean
  const __VERSION__: string
}

export {}