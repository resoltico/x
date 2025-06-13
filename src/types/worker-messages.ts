// src/types/worker-messages.ts
// Enhanced worker message type definitions

export interface BaseWorkerMessage {
  id: string
  type: string
  payload?: any // Add optional payload to base interface
}

export interface TestWorkerMessage extends BaseWorkerMessage {
  type: 'test'
  payload?: undefined // Explicitly mark as no payload for test messages
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
  payload?: undefined // Explicitly mark as no payload for cancel messages
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
  payload?: any // Make payload optional and flexible for all response types
}

export interface TestWorkerResponse extends BaseWorkerResponse {
  type: 'test-response'
  payload?: any
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
  payload?: {
    message: string
  }
}

export type WorkerResponse = 
  | TestWorkerResponse
  | ResultWorkerResponse
  | ProgressWorkerResponse
  | ErrorWorkerResponse
  | ReadyWorkerResponse