// Worker-specific type definitions

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