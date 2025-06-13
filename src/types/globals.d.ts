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

  interface ErrorEventInit {
    message?: string
    filename?: string
    lineno?: number
    colno?: number
    error?: any
    bubbles?: boolean
    cancelable?: boolean
    composed?: boolean
  }

  interface PromiseRejectionEvent extends Event {
    readonly promise: Promise<any>
    readonly reason: any
    preventDefault(): void
  }

  interface PromiseRejectionEventInit {
    promise: Promise<any>
    reason?: any
    bubbles?: boolean
    cancelable?: boolean
    composed?: boolean
  }

  // Enhanced MessageEvent types
  interface MessageEvent<T = any> extends Event {
    readonly data: T
    readonly lastEventId: string
    readonly origin: string
    readonly ports: readonly MessagePort[]
    readonly source: MessageEventSource | null
  }

  interface MessageEventInit<T = any> {
    data?: T
    lastEventId?: string
    origin?: string
    ports?: MessagePort[]
    source?: MessageEventSource | null
    bubbles?: boolean
    cancelable?: boolean
    composed?: boolean
  }

  // Enhanced DragEvent types
  interface DragEvent extends MouseEvent {
    readonly dataTransfer: DataTransfer | null
  }

  interface DragEventInit {
    dataTransfer?: DataTransfer | null
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

  // Enhanced DataTransfer types
  interface DataTransfer {
    dropEffect: 'none' | 'copy' | 'link' | 'move'
    effectAllowed: string
    readonly files: FileList
    readonly items: DataTransferItemList
    readonly types: readonly string[]
    clearData(format?: string): void
    getData(format: string): string
    setData(format: string, data: string): void
    setDragImage(element: HTMLElement, x: number, y: number): void
  }

  interface DataTransferItemList {
    readonly length: number
    add(data: string, type: string): DataTransferItem | null
    add(data: File): DataTransferItem | null
    clear(): void
    remove(index: number): void
    [index: number]: DataTransferItem
  }

  interface DataTransferItem {
    readonly kind: string
    readonly type: string
    getAsFile(): File | null
    getAsString(callback: FunctionStringCallback | null): void
    webkitGetAsEntry?(): FileSystemEntry | null
  }

  interface FunctionStringCallback {
    (data: string): void
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

  // FileSystem API types for drag and drop
  interface FileSystemEntry {
    readonly isDirectory: boolean
    readonly isFile: boolean
    readonly name: string
    readonly fullPath: string
    readonly filesystem: FileSystem
  }

  interface FileSystem {
    readonly name: string
    readonly root: FileSystemDirectoryEntry
  }

  interface FileSystemDirectoryEntry extends FileSystemEntry {
    createReader(): FileSystemDirectoryReader
    getDirectory(path?: string, options?: FileSystemFlags, successCallback?: FileSystemEntryCallback, errorCallback?: ErrorCallback): void
    getFile(path?: string, options?: FileSystemFlags, successCallback?: FileSystemEntryCallback, errorCallback?: ErrorCallback): void
  }

  interface FileSystemDirectoryReader {
    readEntries(successCallback: FileSystemEntriesCallback, errorCallback?: ErrorCallback): void
  }

  interface FileSystemFlags {
    create?: boolean
    exclusive?: boolean
  }

  interface FileSystemEntryCallback {
    (entry: FileSystemEntry): void
  }

  interface FileSystemEntriesCallback {
    (entries: FileSystemEntry[]): void
  }

  interface ErrorCallback {
    (error: DOMException): void
  }
}

export {}