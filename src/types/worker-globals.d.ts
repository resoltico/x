// Worker global scope type definitions

/// <reference lib="webworker" />

declare const self: DedicatedWorkerGlobalScope & typeof globalThis

// Enhanced worker event types with proper compatibility
interface WorkerErrorEvent extends ErrorEvent {
  message: string
  filename?: string
  lineno?: number
  colno?: number
  error?: any
}

interface WorkerPromiseRejectionEvent extends Event {
  promise: Promise<any>
  reason: any
}

// Extend the global scope for workers
declare global {
  interface DedicatedWorkerGlobalScope {
    // Use proper error event handler typing that matches the spec
    onerror: ((this: DedicatedWorkerGlobalScope, ev: string | Event) => any) | null
    onunhandledrejection: ((this: DedicatedWorkerGlobalScope, ev: PromiseRejectionEvent) => any) | null
    
    // Add createImageBitmap for worker contexts
    createImageBitmap(image: ImageBitmapSource): Promise<ImageBitmap>
    createImageBitmap(image: ImageBitmapSource, options: ImageBitmapOptions): Promise<ImageBitmap>
    createImageBitmap(image: ImageBitmapSource, sx: number, sy: number, sw: number, sh: number): Promise<ImageBitmap>
    createImageBitmap(image: ImageBitmapSource, sx: number, sy: number, sw: number, sh: number, options: ImageBitmapOptions): Promise<ImageBitmap>
  }
  
  // Additional worker-specific globals
  interface WorkerNavigator {
    hardwareConcurrency: number
  }
  
  const navigator: WorkerNavigator
  
  // Add ImageBitmap support for workers
  interface ImageBitmap {
    readonly width: number
    readonly height: number
    close(): void
  }
  
  interface ImageBitmapOptions {
    imageOrientation?: 'none' | 'flipY'
    premultiplyAlpha?: 'none' | 'premultiply' | 'default'
    colorSpaceConversion?: 'none' | 'default'
    resizeWidth?: number
    resizeHeight?: number
    resizeQuality?: 'pixelated' | 'low' | 'medium' | 'high'
  }
  
  type ImageBitmapSource = HTMLImageElement | SVGImageElement | HTMLVideoElement | HTMLCanvasElement | ImageBitmap | OffscreenCanvas | Blob
}

export {}