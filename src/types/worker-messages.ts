// src/types/worker-messages.ts
// Enhanced worker message type definitions

export interface BaseWorkerMessage {
  id: string
  type: string
}

export interface TestWorkerMessage extends BaseWorkerMessage {
  type: 'test'
}

export interface ProcessWorkerMessage extends BaseWorkerMessage {
  type: 'process' 
  payload: {
    imageData: any
    type: string
    parameters: any
  }
}

export interface CancelWorkerMessage extends BaseWorkerMessage {
  type: 'cancel'
}

export interface ProgressWorkerMessage extends BaseWorkerMessage {
  type: 'progress'
  payload?: {
    progress?: number
    message?: string
  }
}

export type WorkerMessage = 
  | TestWorkerMessage
  | ProcessWorkerMessage 
  | CancelWorkerMessage
  | ProgressWorkerMessage

export interface BaseWorkerResponse {
  id: string
  type: string
}

export interface TestWorkerResponse extends BaseWorkerResponse {
  type: 'test-response'
}

export interface ResultWorkerResponse extends BaseWorkerResponse {
  type: 'result'
  payload: {
    result: ArrayBuffer
  }
}

export interface ProgressWorkerResponse extends BaseWorkerResponse {
  type: 'progress'
  payload: {
    progress: number
    message?: string
  }
}

export interface ErrorWorkerResponse extends BaseWorkerResponse {
  type: 'error'
  payload: {
    error: string
  }
}

export interface ReadyWorkerResponse extends BaseWorkerResponse {
  type: 'ready'
  payload: {
    message: string
  }
}

export type WorkerResponse = 
  | TestWorkerResponse
  | ResultWorkerResponse
  | ProgressWorkerResponse
  | ErrorWorkerResponse
  | ReadyWorkerResponse