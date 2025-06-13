// src/types/globals.d.ts
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

  // Enhanced HTML element types
  interface HTMLImageElement {
    readonly naturalWidth: number
    readonly naturalHeight: number
  }

  interface HTMLVideoElement {
    readonly videoWidth: number
    readonly videoHeight: number
  }

  // ReadableStream type
  interface ReadableStream<R = any> {
    readonly locked: boolean
    cancel(reason?: any): Promise<void>
    getReader(): ReadableStreamDefaultReader<R>
  }

  interface ReadableStreamDefaultReader<R = any> {
    readonly closed: Promise<undefined>
    cancel(reason?: any): Promise<void>
    read(): Promise<ReadableStreamDefaultReadResult<R>>
    releaseLock(): void
  }

  interface ReadableStreamDefaultReadResult<T> {
    done: boolean
    value: T
  }

  // MessagePort type
  interface MessagePort extends EventTarget {
    onmessage: ((this: MessagePort, ev: MessageEvent) => any) | null
    onmessageerror: ((this: MessagePort, ev: MessageEvent) => any) | null
    close(): void
    postMessage(message: any, transfer?: Transferable[]): void
    start(): void
  }
}

export {}