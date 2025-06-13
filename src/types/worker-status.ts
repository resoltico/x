// src/types/worker-status.ts
// Worker status and management types

export interface WorkerStatus {
  totalWorkers: number
  availableWorkers: number
  activeWorkers: number
  queuedTasks: number
  activeTasks: number
  initialized: boolean
  initializationError: string | null
  environment: string
}

export interface WorkerTaskMapping {
  worker: Worker
  task: ProcessingTask
}

export interface WorkerEnvironmentInfo {
  isProduction: boolean
  isDevelopment: boolean
  isFileProtocol: boolean
  hostname: string
  port?: string
}

export interface WorkerCapabilities {
  hasOffscreenCanvas: boolean
  hasImageBitmap: boolean
  hasCreateImageBitmap: boolean
  hasArrayBuffer: boolean
  hasUint8ClampedArray: boolean
  hardwareConcurrency: number
}

export interface WorkerPerformanceMetrics {
  averageProcessingTime: number
  totalTasksCompleted: number
  successRate: number
  memoryUsage?: {
    used: number
    total: number
    limit: number
  }
}

// Re-export ProcessingTask from main types
import type { ProcessingTask } from './index'
export type { ProcessingTask }