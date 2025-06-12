// Worker global scope type definitions

/// <reference lib="webworker" />

declare const self: DedicatedWorkerGlobalScope & typeof globalThis

// Enhanced worker event types
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
    onerror: ((this: DedicatedWorkerGlobalScope, ev: WorkerErrorEvent) => any) | null
    onunhandledrejection: ((this: DedicatedWorkerGlobalScope, ev: WorkerPromiseRejectionEvent) => any) | null
  }
  
  // Additional worker-specific globals
  interface WorkerNavigator {
    hardwareConcurrency: number
  }
  
  const navigator: WorkerNavigator
}

export {}