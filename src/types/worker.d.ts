// Worker-specific type definitions

declare global {
  // Enhanced ErrorEvent interface for workers
  interface ErrorEvent extends Event {
    message: string
    filename?: string
    lineno?: number
    colno?: number
    error?: any
  }

  // Enhanced PromiseRejectionEvent interface
  interface PromiseRejectionEvent extends Event {
    promise: Promise<any>
    reason: any
  }

  // Worker global scope enhancements
  interface WorkerGlobalScope {
    onerror: ((this: WorkerGlobalScope, ev: ErrorEvent) => any) | null
    onunhandledrejection: ((this: WorkerGlobalScope, ev: PromiseRejectionEvent) => any) | null
  }

  // Dedicated worker global scope
  interface DedicatedWorkerGlobalScope extends WorkerGlobalScope {
    onmessage: ((this: DedicatedWorkerGlobalScope, ev: MessageEvent) => any) | null
    postMessage(message: any, transfer?: Transferable[]): void
  }

  // Service worker registration (for future use)
  interface ServiceWorkerRegistration {
    active: ServiceWorker | null
    installing: ServiceWorker | null
    waiting: ServiceWorker | null
  }
}

// Worker message types
export interface WorkerMessage {
  id: string
  type: 'process' | 'cancel' | 'progress'
  payload?: any
}

export interface WorkerResponse {
  id: string
  type: 'result' | 'progress' | 'error'
  payload?: any
}

// Processing task types
export interface ProcessingTaskPayload {
  imageData: any
  type: string
  parameters: any
}

export interface ProcessingResultPayload {
  result: ArrayBuffer
}

export interface ProcessingProgressPayload {
  progress: number
  message?: string
}

export interface ProcessingErrorPayload {
  error: string
}

export {}