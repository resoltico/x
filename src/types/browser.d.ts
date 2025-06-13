// src/types/browser.d.ts
// Browser API type definitions that may be missing

/// <reference lib="dom" />
/// <reference lib="webworker" />

declare global {
  // Enhanced Mouse Event types
  interface MouseEventInit {
    screenX?: number
    screenY?: number
    clientX?: number
    clientY?: number
    ctrlKey?: boolean
    shiftKey?: boolean
    altKey?: boolean
    metaKey?: boolean
    button?: number
    buttons?: number
    relatedTarget?: EventTarget | null
    bubbles?: boolean
    cancelable?: boolean
    composed?: boolean
  }

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

  // Enhanced Timer types
  type TimerHandler = string | ((...args: any[]) => void)

  // Enhanced Structured Clone types
  interface StructuredSerializeOptions {
    transfer?: Transferable[]
  }

  // Network Information API
  interface NetworkInformation extends EventTarget {
    readonly downlink: number
    readonly effectiveType: string
    readonly rtt: number
    readonly saveData: boolean
  }

  // Lock Manager API
  interface LockManager {
    request(name: string, callback: LockGrantedCallback): Promise<any>
    request(name: string, options: LockOptions, callback: LockGrantedCallback): Promise<any>
    query(): Promise<LockManagerSnapshot>
  }

  interface LockGrantedCallback {
    (lock: Lock | null): any
  }

  interface LockOptions {
    mode?: LockMode
    ifAvailable?: boolean
    steal?: boolean
    signal?: AbortSignal
  }

  type LockMode = 'exclusive' | 'shared'

  interface Lock {
    readonly name: string
    readonly mode: LockMode
  }

  interface LockManagerSnapshot {
    held: LockInfo[]
    pending: LockInfo[]
  }

  interface LockInfo {
    name: string
    mode: LockMode
    clientId: string
  }

  // Permissions API
  interface Permissions {
    query(permissionDesc: PermissionDescriptor): Promise<PermissionStatus>
  }

  interface PermissionDescriptor {
    name: string
  }

  interface PermissionStatus extends EventTarget {
    readonly state: PermissionState
    onchange: ((this: PermissionStatus, ev: Event) => any) | null
  }

  type PermissionState = 'granted' | 'denied' | 'prompt'

  // Serial API
  interface Serial extends EventTarget {
    onconnect: ((this: Serial, ev: Event) => any) | null
    ondisconnect: ((this: Serial, ev: Event) => any) | null
    getPorts(): Promise<SerialPort[]>
    requestPort(options?: SerialPortRequestOptions): Promise<SerialPort>
  }

  interface SerialPort extends EventTarget {
    readonly readable: ReadableStream | null
    readonly writable: WritableStream | null
    onconnect: ((this: SerialPort, ev: Event) => any) | null
    ondisconnect: ((this: SerialPort, ev: Event) => any) | null
    close(): Promise<void>
    getInfo(): SerialPortInfo
    getSignals(): Promise<SerialInputSignals>
    open(options: SerialOptions): Promise<void>
    setSignals(signals?: SerialOutputSignals): Promise<void>
  }

  interface SerialPortRequestOptions {
    filters?: SerialPortFilter[]
  }

  interface SerialPortFilter {
    usbVendorId?: number
    usbProductId?: number
  }

  interface SerialPortInfo {
    usbVendorId?: number
    usbProductId?: number
  }

  interface SerialOptions {
    baudRate: number
    dataBits?: number
    stopBits?: number
    parity?: ParityType
    bufferSize?: number
    flowControl?: FlowControlType
  }

  type ParityType = 'none' | 'even' | 'odd'
  type FlowControlType = 'none' | 'hardware'

  interface SerialInputSignals {
    dataCarrierDetect: boolean
    clearToSend: boolean
    ringIndicator: boolean
    dataSetReady: boolean
  }

  interface SerialOutputSignals {
    dataTerminalReady?: boolean
    requestToSend?: boolean
    break?: boolean
  }

  // Storage Manager API
  interface StorageManager {
    estimate(): Promise<StorageEstimate>
    persist(): Promise<boolean>
    persisted(): Promise<boolean>
  }

  interface StorageEstimate {
    quota?: number
    usage?: number
    usageDetails?: Record<string, number>
  }

  // Navigator UA Data API
  interface NavigatorUAData {
    readonly brands: NavigatorUABrandVersion[]
    readonly mobile: boolean
    readonly platform: string
    getHighEntropyValues(hints: string[]): Promise<UADataValues>
    toJSON(): UALowEntropyJSON
  }

  interface NavigatorUABrandVersion {
    readonly brand: string
    readonly version: string
  }

  interface UADataValues {
    readonly brands?: NavigatorUABrandVersion[]
    readonly mobile?: boolean
    readonly platform?: string
    readonly architecture?: string
    readonly bitness?: string
    readonly model?: string
    readonly platformVersion?: string
    readonly uaFullVersion?: string
  }

  interface UALowEntropyJSON {
    readonly brands: NavigatorUABrandVersion[]
    readonly mobile: boolean
    readonly platform: string
  }

  // Deprecated Storage Quota API
  interface DeprecatedStorageQuota {
    queryUsageAndQuota(successCallback: StorageUsageCallback, errorCallback?: StorageErrorCallback): void
    requestQuota(newQuotaInBytes: number, successCallback?: StorageQuotaCallback, errorCallback?: StorageErrorCallback): void
  }

  interface StorageUsageCallback {
    (currentUsageInBytes: number, currentQuotaInBytes: number): void
  }

  interface StorageQuotaCallback {
    (grantedQuotaInBytes: number): void
  }

  interface StorageErrorCallback {
    (error: DOMException): void
  }

  // Writeable Stream API
  interface WritableStream<W = any> {
    readonly locked: boolean
    abort(reason?: any): Promise<void>
    close(): Promise<void>
    getWriter(): WritableStreamDefaultWriter<W>
  }

  interface WritableStreamDefaultWriter<W = any> {
    readonly closed: Promise<undefined>
    readonly desiredSize: number | null
    readonly ready: Promise<undefined>
    abort(reason?: any): Promise<void>
    close(): Promise<void>
    releaseLock(): void
    write(chunk?: W): Promise<void>
  }
}

export {}