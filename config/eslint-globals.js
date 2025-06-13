// src/config/eslint-globals.js
// Comprehensive browser globals for ESLint configuration

export const browserGlobals = {
  // Standard browser globals
  window: 'readonly',
  document: 'readonly',
  navigator: 'readonly',
  performance: 'readonly',
  console: 'readonly',
  setTimeout: 'readonly',
  setInterval: 'readonly',
  clearTimeout: 'readonly',
  clearInterval: 'readonly',
  
  // URL and Blob APIs
  URL: 'readonly',
  URLSearchParams: 'readonly',
  Blob: 'readonly',
  
  // File APIs
  File: 'readonly',
  FileReader: 'readonly',
  FileList: 'readonly',
  
  // Image and Media APIs
  Image: 'readonly',
  ImageBitmap: 'readonly',
  createImageBitmap: 'readonly',
  
  // HTML Elements
  HTMLElement: 'readonly',
  HTMLCanvasElement: 'readonly',
  HTMLInputElement: 'readonly',
  HTMLImageElement: 'readonly',
  SVGImageElement: 'readonly',
  HTMLVideoElement: 'readonly',
  
  // Canvas APIs
  CanvasRenderingContext2D: 'readonly',
  OffscreenCanvas: 'readonly',
  OffscreenCanvasRenderingContext2D: 'readonly',
  ImageData: 'readonly',
  
  // Worker APIs
  Worker: 'readonly',
  WorkerOptions: 'readonly',
  MessageEvent: 'readonly',
  ErrorEvent: 'readonly',
  PromiseRejectionEvent: 'readonly',
  
  // Event APIs
  Event: 'readonly',
  DragEvent: 'readonly',
  WheelEvent: 'readonly',
  MouseEvent: 'readonly',
  TouchEvent: 'readonly',
  KeyboardEvent: 'readonly',
  
  // DOM APIs
  Node: 'readonly',
  Element: 'readonly',
  EventTarget: 'readonly',
  DataTransfer: 'readonly',
  DataTransferItemList: 'readonly',
  DataTransferItem: 'readonly',
  AbortSignal: 'readonly',
  DOMException: 'readonly',
  
  // Data structures
  Map: 'readonly',
  Set: 'readonly',
  Array: 'readonly',
  Object: 'readonly',
  Promise: 'readonly',
  Error: 'readonly',
  Math: 'readonly',
  Number: 'readonly',
  String: 'readonly',
  Date: 'readonly',
  
  // Typed arrays
  Uint8ClampedArray: 'readonly',
  ArrayBuffer: 'readonly',
  Transferable: 'readonly',
  
  // Web APIs
  fetch: 'readonly',
  WebAssembly: 'readonly',
  btoa: 'readonly',
  
  // MessagePort and related
  MessagePort: 'readonly',
  MessageEventSource: 'readonly',
  
  // Storage and quota
  MediaSource: 'readonly',
  
  // Standard globals
  globalThis: 'readonly'
}

export const workerGlobals = {
  // Worker globals
  self: 'readonly',
  importScripts: 'readonly',
  WorkerGlobalScope: 'readonly',
  DedicatedWorkerGlobalScope: 'readonly',
  
  // Worker-safe APIs
  MessageEvent: 'readonly',
  ErrorEvent: 'readonly',
  PromiseRejectionEvent: 'readonly',
  OffscreenCanvas: 'readonly',
  OffscreenCanvasRenderingContext2D: 'readonly',
  ImageData: 'readonly',
  ImageBitmap: 'readonly',
  createImageBitmap: 'readonly',
  
  // Standard APIs available in workers
  console: 'readonly',
  setTimeout: 'readonly',
  setInterval: 'readonly',
  clearTimeout: 'readonly',
  clearInterval: 'readonly',
  URL: 'readonly',
  Blob: 'readonly',
  Image: 'readonly',
  
  // Data structures
  Map: 'readonly',
  Set: 'readonly',
  Array: 'readonly',
  Object: 'readonly',
  Promise: 'readonly',
  Error: 'readonly',
  Math: 'readonly',
  Number: 'readonly',
  String: 'readonly',
  Date: 'readonly',
  
  // Typed arrays
  Uint8ClampedArray: 'readonly',
  ArrayBuffer: 'readonly',
  Event: 'readonly',
  
  // Web APIs
  fetch: 'readonly',
  WebAssembly: 'readonly'
}

export const testGlobals = {
  // Vitest globals
  vi: 'readonly',
  describe: 'readonly',
  it: 'readonly',
  test: 'readonly',
  expect: 'readonly',
  beforeAll: 'readonly',
  afterAll: 'readonly',
  beforeEach: 'readonly',
  afterEach: 'readonly',
  
  // All browser globals for tests
  ...browserGlobals,
  
  // Additional test-specific globals
  Performance: 'readonly',
  EventInit: 'readonly',
  MessageEventInit: 'readonly',
  ErrorEventInit: 'readonly',
  PromiseRejectionEventInit: 'readonly',
  DragEventInit: 'readonly',
  MouseEventInit: 'readonly',
  
  // Global this is writable in tests
  globalThis: 'writable'
}