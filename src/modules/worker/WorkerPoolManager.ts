// src/modules/worker/WorkerPoolManager.ts
// Enhanced pool manager with comprehensive debugging

import { WorkerFactory } from './WorkerFactory'
import { debugLogger } from '@/utils/debugLogger'
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
  private messageHandler?: (event: MessageEvent, worker: Worker) => void

  constructor(workerCount?: number) {
    this.workerFactory = WorkerFactory.getInstance()
    this.workerCount = workerCount || this.workerFactory.getRecommendedWorkerCount()
    
    debugLogger.log('info', 'worker-pool', 'WorkerPoolManager created', {
      requestedWorkers: workerCount,
      recommendedWorkers: this.workerCount
    })
  }

  /**
   * Initialize the worker pool with enhanced fallback handling
   */
  async initialize(): Promise<void> {
    if (this.isInitialized) {
      debugLogger.log('info', 'worker-pool', 'Already initialized, skipping')
      return
    }

    const startTime = Date.now()
    debugLogger.log('info', 'worker-pool', `🚀 Initializing Worker Pool with ${this.workerCount} workers`)

    try {
      const env = this.workerFactory.getEnvironmentInfo()
      const caps = this.workerFactory.getCapabilities()
      
      debugLogger.log('info', 'worker-pool', 'Environment and capabilities', { env, caps })

      let successfulWorkers = 0
      const workerCreationPromises: Promise<void>[] = []

      // Try to create workers concurrently for better performance
      for (let i = 0; i < this.workerCount; i++) {
        const promise = this.createWorkerWithRetry(i).then((worker) => {
          if (worker) {
            this.workers.push(worker)
            this.availableWorkers.push(worker)
            successfulWorkers++
            debugLogger.log('info', 'worker-pool', `✅ Worker ${i} initialized successfully`)
          } else {
            debugLogger.log('warn', 'worker-pool', `❌ Failed to create worker ${i}`)
          }
        })
        workerCreationPromises.push(promise)
      }
      
      // Wait for all worker creation attempts to complete
      await Promise.all(workerCreationPromises)
      
      // If no workers were created successfully, ensure we have at least a fallback
      if (successfulWorkers === 0 && !this.fallbackWorkerCreated) {
        debugLogger.log('warn', 'worker-pool', '⚠️ No regular workers initialized, creating fallback worker...')
        await this.createFallbackWorker()
        if (this.workers.length > 0) {
          successfulWorkers = 1
        }
      }
      
      if (this.workers.length === 0) {
        const errorMsg = 'No workers could be initialized. Image processing will not be available.'
        this.initializationError = errorMsg
        debugLogger.log('error', 'worker-pool', errorMsg)
        throw new Error(errorMsg)
      }
      
      this.isInitialized = true
      const duration = Date.now() - startTime
      
      debugLogger.log('info', 'worker-pool', `✅ Worker pool initialized`, {
        successful: successfulWorkers,
        requested: this.workerCount,
        hasFallback: this.fallbackWorkerCreated,
        duration: `${duration}ms`
      })
      
      if (successfulWorkers < this.workerCount) {
        debugLogger.log('warn', 'worker-pool', `⚠️ Only ${successfulWorkers}/${this.workerCount} workers initialized successfully`)
      }
      
    } catch (error) {
      this.initializationError = error instanceof Error ? error.message : 'Unknown initialization error'
      const duration = Date.now() - startTime
      debugLogger.log('error', 'worker-pool', `❌ Failed to initialize worker pool after ${duration}ms`, error)
      throw error
    }
  }

  /**
   * Create a worker with retry logic
   */
  private async createWorkerWithRetry(index: number, maxRetries: number = 2): Promise<Worker | null> {
    for (let attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        debugLogger.log('debug', 'worker-pool', `Creating worker ${index}, attempt ${attempt + 1}/${maxRetries + 1}`)
        
        const worker = await this.workerFactory.createWorker(index)
        if (worker) {
          this.setupWorkerEventHandlers(worker, index)
          return worker
        }
      } catch (workerError) {
        debugLogger.log('warn', 'worker-pool', `Worker ${index} creation attempt ${attempt + 1} failed`, workerError)
        
        if (attempt < maxRetries) {
          // Wait briefly before retrying
          await new Promise(resolve => setTimeout(resolve, 100 * (attempt + 1)))
        }
      }
    }
    
    debugLogger.log('error', 'worker-pool', `Failed to create worker ${index} after ${maxRetries + 1} attempts`)
    return null
  }

  /**
   * Create a fallback inline worker as last resort
   */
  private async createFallbackWorker(): Promise<void> {
    try {
      debugLogger.log('info', 'worker-pool', '🔧 Creating fallback inline worker...')
      
      // Use the WorkerFactory to create an inline worker
      const fallbackWorker = await this.workerFactory.createWorker(999)
      
      if (fallbackWorker) {
        this.setupWorkerEventHandlers(fallbackWorker, 999)
        this.workers.push(fallbackWorker)
        this.availableWorkers.push(fallbackWorker)
        this.fallbackWorkerCreated = true
        debugLogger.log('info', 'worker-pool', '✅ Fallback worker created successfully')
      } else {
        throw new Error('Failed to create fallback worker')
      }
      
    } catch (error) {
      debugLogger.log('error', 'worker-pool', '❌ Failed to create fallback worker', error)
      throw error
    }
  }

  /**
   * Setup event handlers for a worker with enhanced logging
   */
  private setupWorkerEventHandlers(worker: Worker, index: number) {
    worker.onmessage = (event) => {
      try {
        debugLogger.log('debug', 'worker-pool', `Worker ${index} message`, event.data)
        if (this.messageHandler) {
          this.messageHandler(event, worker)
        }
      } catch (error) {
        debugLogger.log('error', 'worker-pool', `Worker ${index} message handling error`, error)
      }
    }
    
    worker.onerror = (event) => {
      debugLogger.log('error', 'worker-pool', `Worker ${index} error`, event)
      this.handleWorkerError(event, index)
    }
    
    worker.onmessageerror = (event) => {
      debugLogger.log('error', 'worker-pool', `Worker ${index} message error`, event)
    }
  }

  /**
   * Handle worker errors with better recovery
   */
  private handleWorkerError(event: ErrorEvent, index?: number) {
    debugLogger.log('error', 'worker-pool', `💥 Worker ${index ?? 'unknown'} error`, event)
    
    // Try to replace the failed worker
    this.replaceFailedWorker(event.target as Worker, index)
  }

  /**
   * Replace a failed worker with enhanced fallback
   */
  private async replaceFailedWorker(failedWorker: Worker, index?: number) {
    try {
      debugLogger.log('info', 'worker-pool', `🔄 Replacing failed worker ${index ?? 'unknown'}`)
      
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
          const newWorker = await this.createWorkerWithRetry(Date.now())
          if (newWorker) {
            this.workers.push(newWorker)
            this.availableWorkers.push(newWorker)
            debugLogger.log('info', 'worker-pool', `✅ Successfully replaced failed worker`)
          } else {
            throw new Error('Failed to create replacement worker')
          }
        } catch (replacementError) {
          debugLogger.log('error', 'worker-pool', `❌ Failed to create replacement worker`, replacementError)
          
          // If we have no workers left and haven't created a fallback, create one
          if (this.workers.length === 0 && !this.fallbackWorkerCreated) {
            debugLogger.log('warn', 'worker-pool', '⚠️ No workers left, creating emergency fallback...')
            await this.createFallbackWorker()
          }
        }
      }
      
      // Terminate the failed worker
      try {
        failedWorker.terminate()
      } catch (terminateError) {
        debugLogger.log('warn', 'worker-pool', '⚠️ Error terminating failed worker', terminateError)
      }
    } catch (error) {
      debugLogger.log('error', 'worker-pool', '❌ Failed to replace worker', error)
    }
  }

  /**
   * Get an available worker from the pool
   */
  getAvailableWorker(): Worker | null {
    const worker = this.availableWorkers.shift() || null
    if (worker) {
      debugLogger.log('debug', 'worker-pool', 'Worker assigned from pool', {
        remaining: this.availableWorkers.length
      })
    } else {
      debugLogger.log('warn', 'worker-pool', 'No available workers in pool')
    }
    return worker
  }

  /**
   * Return a worker to the available pool
   */
  returnWorker(worker: Worker) {
    if (this.workers.includes(worker) && !this.availableWorkers.includes(worker)) {
      this.availableWorkers.push(worker)
      debugLogger.log('debug', 'worker-pool', 'Worker returned to pool', {
        available: this.availableWorkers.length
      })
    }
  }

  /**
   * Add a task mapping
   */
  addTaskMapping(taskId: string, worker: Worker, task: any) {
    this.activeTasksMap.set(taskId, { worker, task })
    debugLogger.log('debug', 'worker-pool', `Task mapping added: ${taskId}`, {
      activeTasks: this.activeTasksMap.size
    })
  }

  /**
   * Remove a task mapping
   */
  removeTaskMapping(taskId: string): WorkerTaskMapping | undefined {
    const mapping = this.activeTasksMap.get(taskId)
    if (mapping) {
      this.activeTasksMap.delete(taskId)
      debugLogger.log('debug', 'worker-pool', `Task mapping removed: ${taskId}`, {
        activeTasks: this.activeTasksMap.size
      })
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
    debugLogger.log('info', 'worker-pool', '⏸️ Pausing worker pool')
    this.availableWorkers = []
  }

  /**
   * Resume worker operations
   */
  resume() {
    debugLogger.log('info', 'worker-pool', '▶️ Resuming worker pool')
    this.availableWorkers = this.workers.filter(worker => 
      !Array.from(this.activeTasksMap.values()).some(({ worker: activeWorker }) => activeWorker === worker)
    )
  }

  /**
   * Destroy all workers
   */
  destroy() {
    const startTime = Date.now()
    debugLogger.log('info', 'worker-pool', '🗑️ Destroying worker pool...')
    
    // Terminate all workers
    this.workers.forEach((worker, index) => {
      try {
        worker.terminate()
        debugLogger.log('debug', 'worker-pool', `🗑️ Terminated worker ${index}`)
      } catch (error) {
        debugLogger.log('warn', 'worker-pool', `⚠️ Error terminating worker ${index}`, error)
      }
    })
    
    // Clear arrays and maps
    this.workers = []
    this.availableWorkers = []
    this.activeTasksMap.clear()
    this.isInitialized = false
    this.initializationError = null
    this.fallbackWorkerCreated = false
    
    const duration = Date.now() - startTime
    debugLogger.log('info', 'worker-pool', `✅ Worker pool destroyed in ${duration}ms`)
  }

  /**
   * Check if pool is ready
   */
  isReady(): boolean {
    const ready = this.isInitialized && this.workers.length > 0
    if (!ready) {
      debugLogger.log('debug', 'worker-pool', 'Pool not ready', {
        initialized: this.isInitialized,
        workerCount: this.workers.length
      })
    }
    return ready
  }

  /**
   * Wait for pool to be ready
   */
  async waitForReady(timeout: number = 10000): Promise<boolean> {
    const startTime = Date.now()
    debugLogger.log('info', 'worker-pool', `⏳ Waiting for pool to be ready (timeout: ${timeout}ms)`)
    
    while (!this.isReady() && (Date.now() - startTime) < timeout) {
      await new Promise(resolve => setTimeout(resolve, 100))
    }
    
    const duration = Date.now() - startTime
    const ready = this.isReady()
    
    if (ready) {
      debugLogger.log('info', 'worker-pool', `✅ Pool ready after ${duration}ms`)
    } else {
      debugLogger.log('error', 'worker-pool', `❌ Pool not ready after ${duration}ms timeout`)
    }
    
    return ready
  }

  /**
   * Get detailed status for debugging
   */
  getDetailedStatus() {
    const basicStatus = this.getStatus()
    const detailed = {
      ...basicStatus,
      fallbackWorkerCreated: this.fallbackWorkerCreated,
      workerTypes: this.workers.map((_, index) => 
        index === 999 ? 'fallback' : 'regular'
      ),
      activeTasksCount: this.activeTasksMap.size,
      capabilities: this.workerFactory.getCapabilities(),
      environment: this.workerFactory.getEnvironmentInfo()
    }
    
    debugLogger.log('debug', 'worker-pool', 'Detailed status', detailed)
    return detailed
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
    this.messageHandler = handler
  }

  /**
   * Test all workers
   */
  async testAllWorkers(): Promise<{ [index: number]: boolean }> {
    debugLogger.log('info', 'worker-pool', '🧪 Testing all workers...')
    
    const results: { [index: number]: boolean } = {}
    
    const testPromises = this.workers.map(async (worker, index) => {
      try {
        const testResult = await this.testSingleWorker(worker, index)
        results[index] = testResult
        debugLogger.log('debug', 'worker-pool', `Worker ${index} test: ${testResult ? 'PASS' : 'FAIL'}`)
      } catch (error) {
        results[index] = false
        debugLogger.log('error', 'worker-pool', `Worker ${index} test error`, error)
      }
    })
    
    await Promise.all(testPromises)
    
    const passed = Object.values(results).filter(result => result).length
    const total = Object.keys(results).length
    
    debugLogger.log('info', 'worker-pool', `✅ Worker tests completed: ${passed}/${total} passed`, results)
    return results
  }

  /**
   * Test a single worker
   */
  private testSingleWorker(worker: Worker, index: number): Promise<boolean> {
    return new Promise((resolve) => {
      const testId = `test-worker-${index}-${Date.now()}`
      const timeout = setTimeout(() => {
        worker.removeEventListener('message', onMessage)
        resolve(false)
      }, 3000)

      const onMessage = (event: MessageEvent) => {
        if (event.data?.id === testId) {
          clearTimeout(timeout)
          worker.removeEventListener('message', onMessage)
          resolve(true)
        }
      }

      worker.addEventListener('message', onMessage)
      worker.postMessage({ id: testId, type: 'test' })
    })
  }

  /**
   * Get worker diagnostic information
   */
  async getDiagnostics(): Promise<any> {
    debugLogger.log('info', 'worker-pool', '🔬 Running worker pool diagnostics...')
    
    const workerTests = await this.testAllWorkers()
    const workerUrls = await this.workerFactory.diagnoseWorkerUrls()
    
    const diagnostics = {
      timestamp: new Date().toISOString(),
      pool: this.getDetailedStatus(),
      factory: {
        environment: this.workerFactory.getEnvironmentInfo(),
        capabilities: this.workerFactory.getCapabilities(),
        recommendedWorkerCount: this.workerFactory.getRecommendedWorkerCount()
      },
      workerTests,
      workerUrls,
      performance: this.getPerformanceMetrics()
    }
    
    debugLogger.log('info', 'worker-pool', '✅ Worker pool diagnostics completed', diagnostics)
    return diagnostics
  }
}