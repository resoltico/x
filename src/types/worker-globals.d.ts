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
    // Use the standard ErrorEvent type for better compatibility
    onerror: ((this: DedicatedWorkerGlobalScope, ev: ErrorEvent) => any) | null
    onunhandledrejection: ((this: DedicatedWorkerGlobalScope, ev: PromiseRejectionEvent) => any) | null
  }
  
  // Additional worker-specific globals
  interface WorkerNavigator {
    hardwareConcurrency: number
  }
  
  const navigator: WorkerNavigator
}

export {}