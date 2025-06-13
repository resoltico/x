// src/types/browser.d.ts
// Browser API type definitions that may be missing

/// <reference lib="dom" />
/// <reference lib="webworker" />

declare global {
  // ImageBitmap related types
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

  type ImageBitmapSource = 
    | HTMLImageElement 
    | SVGImageElement 
    | HTMLVideoElement 
    | HTMLCanvasElement 
    | ImageBitmap 
    | OffscreenCanvas 
    | Blob

  function createImageBitmap(image: ImageBitmapSource): Promise<ImageBitmap>
  function createImageBitmap(image: ImageBitmapSource, options: ImageBitmapOptions): Promise<ImageBitmap>
  function createImageBitmap(image: ImageBitmapSource, sx: number, sy: number, sw: number, sh: number): Promise<ImageBitmap>
  function createImageBitmap(image: ImageBitmapSource, sx: number, sy: number, sw: number, sh: number, options: ImageBitmapOptions): Promise<ImageBitmap>

  // SVG related types
  interface SVGAnimatedLength {
    readonly baseVal: SVGLength
    readonly animVal: SVGLength
  }

  interface SVGLength {
    readonly unitType: number
    value: number
    valueInSpecifiedUnits: number
    valueAsString: string
    newValueSpecifiedUnits(unitType: number, valueInSpecifiedUnits: number): void
    convertToSpecifiedUnits(unitType: number): void
  }

  // Service Worker types
  interface ServiceWorker extends EventTarget {
    readonly scriptURL: string
    readonly state: ServiceWorkerState
    onstatechange: ((this: ServiceWorker, ev: Event) => any) | null
    postMessage(message: any, transfer?: Transferable[]): void
  }

  type ServiceWorkerState = 'installing' | 'installed' | 'activating' | 'activated' | 'redundant'

  // Message event source types
  type MessageEventSource = WindowProxy | MessagePort | ServiceWorker

  // Enhanced Worker types
  interface DedicatedWorkerGlobalScope extends WorkerGlobalScope {
    readonly name: string
    onmessage: ((this: DedicatedWorkerGlobalScope, ev: MessageEvent) => any) | null
    onmessageerror: ((this: DedicatedWorkerGlobalScope, ev: MessageEvent) => any) | null
    postMessage(message: any, transfer?: Transferable[]): void
  }

  interface WorkerGlobalScope extends EventTarget {
    readonly caches: CacheStorage
    readonly crypto: Crypto
    readonly indexedDB: IDBFactory
    readonly isSecureContext: boolean
    readonly location: WorkerLocation
    readonly navigator: WorkerNavigator
    readonly origin: string
    readonly performance: Performance
    readonly self: WorkerGlobalScope & typeof globalThis
    onerror: ((this: WorkerGlobalScope, ev: ErrorEvent) => any) | null
    onlanguagechange: ((this: WorkerGlobalScope, ev: Event) => any) | null
    onoffline: ((this: WorkerGlobalScope, ev: Event) => any) | null
    ononline: ((this: WorkerGlobalScope, ev: Event) => any) | null
    onrejectionhandled: ((this: WorkerGlobalScope, ev: PromiseRejectionEvent) => any) | null
    onunhandledrejection: ((this: WorkerGlobalScope, ev: PromiseRejectionEvent) => any) | null
    atob(data: string): string
    btoa(data: string): string
    clearInterval(id?: number): void
    clearTimeout(id?: number): void
    createImageBitmap(image: ImageBitmapSource, options?: ImageBitmapOptions): Promise<ImageBitmap>
    createImageBitmap(image: ImageBitmapSource, sx: number, sy: number, sw: number, sh: number, options?: ImageBitmapOptions): Promise<ImageBitmap>
    fetch(input: RequestInfo | URL, init?: RequestInit): Promise<Response>
    importScripts(...urls: string[]): void
    queueMicrotask(callback: VoidFunction): void
    reportError(e: any): void
    setInterval(handler: TimerHandler, timeout?: number, ...arguments: any[]): number
    setTimeout(handler: TimerHandler, timeout?: number, ...arguments: any[]): number
    structuredClone(value: any, options?: StructuredSerializeOptions): any
  }

  interface WorkerLocation {
    readonly hash: string
    readonly host: string
    readonly hostname: string
    readonly href: string
    readonly origin: string
    readonly pathname: string
    readonly port: string
    readonly protocol: string
    readonly search: string
  }

  interface WorkerNavigator {
    readonly appCodeName: string
    readonly appName: string
    readonly appVersion: string
    readonly connection: NetworkInformation
    readonly cookieEnabled: boolean
    readonly hardwareConcurrency: number
    readonly language: string
    readonly languages: readonly string[]
    readonly locks: LockManager
    readonly onLine: boolean
    readonly permissions: Permissions
    readonly platform: string
    readonly product: string
    readonly productSub: string
    readonly serial: Serial
    readonly serviceWorker: ServiceWorkerContainer
    readonly storage: StorageManager
    readonly userAgent: string
    readonly userAgentData: NavigatorUAData
    readonly vendor: string
    readonly vendorSub: string
    readonly webkitPersistentStorage: DeprecatedStorageQuota
    readonly webkitTemporaryStorage: DeprecatedStorageQuota
  }
}

export {}