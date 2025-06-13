// src/types/worker-globals.d.ts
// Worker global scope type definitions

/// <reference lib="webworker" />

declare global {
  interface DedicatedWorkerGlobalScope {
    // Use proper error event handler typing that matches the spec
    onerror: ((this: DedicatedWorkerGlobalScope, ev: string | Event) => any) | null
    onunhandledrejection: ((this: DedicatedWorkerGlobalScope, ev: PromiseRejectionEvent) => any) | null
  }
  
  // Additional worker-specific globals
  interface WorkerNavigator {
    hardwareConcurrency: number
  }

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
}

export {}