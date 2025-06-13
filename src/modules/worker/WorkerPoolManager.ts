// src/modules/worker/WorkerPoolManager.ts
// Manages the pool of workers and their lifecycle

import { WorkerFactory } from './WorkerFactory'
import type { WorkerStatus, WorkerTaskMapping, WorkerPerformanceMetrics } from '@/types/worker-status'

export class WorkerPoolManager {
  private workers: Worker[] = []
  private availableWorkers: Worker[] = []
  private activeTasksMap = new Map<string, WorkerTaskMapping>()
  private workerCount: number
  private workerFactory: WorkerFactory
  private isInitialized = false
  private initializationError: string | null = null
  private fallbackWorkerCreated = false

  constructor(workerCount?: number) {
    this.workerFactory = WorkerFactory.getInstance()
    this.workerCount = workerCount || this.workerFactory.getRecommendedWorkerCount()
  }

  /**
   * Initialize the worker pool with enhanced fallback handling
   */
  async initialize(): Promise<void> {
    if (this.isInitialized) return

    console.group('🚀 Initializing Worker Pool')
    console.log(`Target worker count: ${this.workerCount}`)

    try {
      const env = this.workerFactory.getEnvironmentInfo()
      const caps = this.workerFactory.getCapabilities()
      
      console.log('Environment:', env)
      console.log('Capabilities:', caps)

      let successfulWorkers = 0

      // Try to create regular workers first
      for (let i = 0; i < this.workerCount; i++) {
        try {
          const worker = await this.workerFactory.createWorker(i)
          if (worker) {
            this.setupWorkerEventHandlers(worker, i)
            this.workers.push(worker)
            this.availableWorkers.push(worker)
            successfulWorkers++
            console.log(`✅ Worker ${i} initialized successfully`)
          }
        } catch (workerError) {
          console.error(`❌ Failed to create worker ${i}:`, workerError)
        }
      }
      
      // If no workers were created successfully, ensure we have at least a fallback
      if (successfulWorkers === 0 && !this.fallbackWorkerCreated) {
        console.warn('⚠️ No regular workers initialized, creating fallback worker...')
        await this.createFallbackWorker()
      }
      
      if (this.workers.length === 0) {
        const errorMsg = 'No workers could be initialized. Image processing will not be available.'
        this.initializationError = errorMsg
        console.error('❌', errorMsg)
        console.groupEnd()
        throw new Error(errorMsg)
      }
      
      this.isInitialized = true
      console.log(`✅ Successfully initialized ${this.workers.length} out of ${this.workerCount} requested workers`)
      
      if (successfulWorkers < this.workerCount) {
        console.warn(`⚠️ Only ${successfulWorkers}/${this.workerCount} workers initialized successfully`)
      }
      
      console.groupEnd()
    } catch (error) {
      this.initializationError = error instanceof Error ? error.message : 'Unknown initialization error'
      console.error('❌ Failed to initialize worker pool:', error)
      console.groupEnd()
      throw error
    }
  }

  /**
   * Create a fallback inline worker as last resort
   */
  private async createFallbackWorker(): Promise<void> {
    try {
      console.log('🔧 Creating fallback inline worker...')
      
      // Get the inline worker code from WorkerFactory
      const factory = WorkerFactory.getInstance()
      const inlineCode = (factory as any).getInlineWorkerCode()
      const blob = new Blob([inlineCode], { type: 'application/javascript' })
      const workerUrl = URL.createObjectURL(blob)
      
      const fallbackWorker = new Worker(workerUrl, { name: 'fallback-worker' })
      
      // Test the fallback worker
      const testResult = await this.testFallbackWorker(fallbackWorker)
      if (testResult) {
        this.setupWorkerEventHandlers(fallbackWorker, 999) // Use 999 as fallback worker index
        this.workers.push(fallbackWorker)
        this.availableWorkers.push(fallbackWorker)
        this.fallbackWorkerCreated = true
        console.log('✅ Fallback worker created successfully')
      } else {
        fallbackWorker.terminate()
        throw new Error('Fallback worker test failed')
      }
      
      // Clean up blob URL
      URL.revokeObjectURL(workerUrl)
      
    } catch (error) {
      console.error('❌ Failed to create fallback worker:', error)
      throw error
    }
  }

  /**
   * Test fallback worker
   */
  private testFallbackWorker(worker: Worker, timeout: number = 2000): Promise<boolean> {
    return new Promise((resolve) => {
      let resolved = false
      const testId = `fallback-test-${Date.now()}`
      
      const timer = setTimeout(() => {
        if (!resolved) {
          resolved = true
          resolve(false)
        }
      }, timeout)
      
      const onMessage = (event: MessageEvent) => {
        if (event.data?.id === testId && !resolved) {
          resolved = true
          clearTimeout(timer)
          worker.removeEventListener('message', onMessage)
          worker.removeEventListener('error', onError)
          resolve(true)
        }
      }
      
      const onError = () => {
        if (!resolved) {
          resolved = true
          clearTimeout(timer)
          worker.removeEventListener('message', onMessage)
          worker.removeEventListener('error', onError)
          resolve(false)
        }
      }
      
      worker.addEventListener('message', onMessage)
      worker.addEventListener('error', onError)
      
      try {
        worker.postMessage({ id: testId, type: 'test' })
      } catch (error) {
        console.warn('Failed to test fallback worker:', error)
        onError()
      }
    })
  }

  /**
   * Setup event handlers for a worker
   */
  private setupWorkerEventHandlers(worker: Worker, index: number) {
    worker.onmessage = (event) => {
      try {
        this.handleWorkerMessage(event, worker)
      } catch (error) {
        console.error(`Worker ${index} message handling error:`, error)
      }
    }
    
    worker.onerror = (event) => {
      console.error(`Worker ${index} error:`, event)
      this.handleWorkerError(event, index)
    }
    
    worker.onmessageerror = (event) => {
      console.error(`Worker ${index} message error:`, event)
    }
  }

  /**
   * Handle messages from workers (to be implemented by orchestrator)
   */
  private handleWorkerMessage(_event: MessageEvent, _worker: Worker) {
    // This will be implemented by the orchestrator
    // The pool manager focuses on worker lifecycle, not task logic
  }

  /**
   * Handle worker errors with better recovery
   */
  private handleWorkerError(event: ErrorEvent, index?: number) {
    console.error(`💥 Worker ${index ?? 'unknown'} error:`, event)
    
    // Try to replace the failed worker
    this.replaceFailedWorker(event.target as Worker, index)
  }

  /**
   * Replace a failed worker with enhanced fallback
   */
  private async replaceFailedWorker(failedWorker: Worker, index?: number) {
    try {
      console.log(`🔄 Replacing failed worker ${index ?? 'unknown'}`)
      
      // Remove from available workers
      const availableIndex = this.availableWorkers.indexOf(failedWorker)
      if (availableIndex !== -1) {
        this.availableWorkers.splice(availableIndex, 1)
      }

      // Remove from workers list
      const workerIndex = this.workers.indexOf(failedWorker)
      if (workerIndex !== -1) {
        this.workers.splice(workerIndex, 1)
        
        // Try to create replacement worker
        try {
          const newWorker = await this.workerFactory.createWorker(Date.now())
          if (newWorker) {
            this.setupWorkerEventHandlers(newWorker, Date.now())
            this.workers.push(newWorker)
            this.availableWorkers.push(newWorker)
            console.log(`✅ Successfully replaced failed worker`)
          } else {
            throw new Error('Failed to create replacement worker')
          }
        } catch (replacementError) {
          console.error(`❌ Failed to create replacement worker:`, replacementError)
          
          // If we have no workers left and haven't created a fallback, create one
          if (this.workers.length === 0 && !this.fallbackWorkerCreated) {
            console.warn('⚠️ No workers left, creating emergency fallback...')
            await this.createFallbackWorker()
          }
        }
      }
      
      // Terminate the failed worker
      try {
        failedWorker.terminate()
      } catch (terminateError) {
        console.warn('⚠️ Error terminating failed worker:', terminateError)
      }
    } catch (error) {
      console.error('❌ Failed to replace worker:', error)
    }
  }

  /**
   * Get an available worker from the pool
   */
  getAvailableWorker(): Worker | null {
    return this.availableWorkers.shift() || null
  }

  /**
   * Return a worker to the available pool
   */
  returnWorker(worker: Worker) {
    if (this.workers.includes(worker) && !this.availableWorkers.includes(worker)) {
      this.availableWorkers.push(worker)
    }
  }

  /**
   * Add a task mapping
   */
  addTaskMapping(taskId: string, worker: Worker, task: any) {
    this.activeTasksMap.set(taskId, { worker, task })
  }

  /**
   * Remove a task mapping
   */
  removeTaskMapping(taskId: string): WorkerTaskMapping | undefined {
    const mapping = this.activeTasksMap.get(taskId)
    if (mapping) {
      this.activeTasksMap.delete(taskId)
    }
    return mapping
  }

  /**
   * Get task mapping
   */
  getTaskMapping(taskId: string): WorkerTaskMapping | undefined {
    return this.activeTasksMap.get(taskId)
  }

  /**
   * Get all active task mappings
   */
  getActiveTasks(): WorkerTaskMapping[] {
    return Array.from(this.activeTasksMap.values())
  }

  /**
   * Get worker pool status
   */
  getStatus(): WorkerStatus {
    return {
      totalWorkers: this.workers.length,
      availableWorkers: this.availableWorkers.length,
      activeWorkers: this.workers.length - this.availableWorkers.length,
      queuedTasks: 0, // This will be managed by orchestrator
      activeTasks: this.activeTasksMap.size,
      initialized: this.isInitialized,
      initializationError: this.initializationError,
      environment: this.getEnvironmentString()
    }
  }

  /**
   * Get performance metrics
   */
  getPerformanceMetrics(): WorkerPerformanceMetrics {
    // Basic implementation - can be enhanced
    const completedTasks = Array.from(this.activeTasksMap.values())
      .map(({ task }) => task)
      .filter(task => task.status === 'completed' && task.completedAt)

    if (completedTasks.length === 0) {
      return {
        averageProcessingTime: 0,
        totalTasksCompleted: 0,
        successRate: 0
      }
    }

    const totalTime = completedTasks.reduce((sum, task) => {
      const processingTime = task.completedAt!.getTime() - task.createdAt.getTime()
      return sum + processingTime
    }, 0)

    const metrics: WorkerPerformanceMetrics = {
      averageProcessingTime: totalTime / completedTasks.length,
      totalTasksCompleted: completedTasks.length,
      successRate: 100 // All completed tasks are successful by definition
    }

    // Add memory usage if available
    const performance = globalThis.performance as any
    if (performance?.memory) {
      metrics.memoryUsage = {
        used: Math.round(performance.memory.usedJSHeapSize / 1024 / 1024),
        total: Math.round(performance.memory.totalJSHeapSize / 1024 / 1024),
        limit: Math.round(performance.memory.jsHeapSizeLimit / 1024 / 1024)
      }
    }

    return metrics
  }

  /**
   * Pause all workers (prevent new task assignments)
   */
  pause() {
    console.log('⏸️ Pausing worker pool')
    this.availableWorkers = []
  }

  /**
   * Resume worker operations
   */
  resume() {
    console.log('▶️ Resuming worker pool')
    this.availableWorkers = this.workers.filter(worker => 
      !Array.from(this.activeTasksMap.values()).some(({ worker: activeWorker }) => activeWorker === worker)
    )
  }

  /**
   * Destroy all workers
   */
  destroy() {
    console.log('🗑️ Destroying worker pool')
    
    // Terminate all workers
    this.workers.forEach((worker, index) => {
      try {
        worker.terminate()
        console.log(`🗑️ Terminated worker ${index}`)
      } catch (error) {
        console.warn(`⚠️ Error terminating worker ${index}:`, error)
      }
    })
    
    // Clear arrays and maps
    this.workers = []
    this.availableWorkers = []
    this.activeTasksMap.clear()
    this.isInitialized = false
    this.initializationError = null
    this.fallbackWorkerCreated = false
    
    console.log('✅ Worker pool destroyed')
  }

  /**
   * Check if pool is ready
   */
  isReady(): boolean {
    return this.isInitialized && this.workers.length > 0
  }

  /**
   * Wait for pool to be ready
   */
  async waitForReady(timeout: number = 5000): Promise<boolean> {
    const startTime = Date.now()
    
    while (!this.isReady() && (Date.now() - startTime) < timeout) {
      await new Promise(resolve => setTimeout(resolve, 100))
    }
    
    return this.isReady()
  }

  /**
   * Get detailed status for debugging
   */
  getDetailedStatus() {
    return {
      ...this.getStatus(),
      fallbackWorkerCreated: this.fallbackWorkerCreated,
      workerTypes: this.workers.map((_, index) => 
        index === 999 ? 'fallback' : 'regular'
      )
    }
  }

  private getEnvironmentString(): string {
    const env = this.workerFactory.getEnvironmentInfo()
    if (env.isFileProtocol) return 'File Protocol'
    if (env.isDevelopment) return `Development${env.port ? ` (port ${env.port})` : ''}`
    if (env.isProduction) return 'Production'
    return 'Unknown'
  }

  // Allow orchestrator to set message handler
  setMessageHandler(handler: (event: MessageEvent, worker: Worker) => void) {
    this.handleWorkerMessage = handler
  }
}