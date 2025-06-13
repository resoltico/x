// src/modules/WorkerOrchestratorModule.ts
// Enhanced orchestrator with comprehensive debugging

import { WorkerPoolManager } from './worker/WorkerPoolManager'
import { TaskQueueManager } from './worker/TaskQueueManager'
import { debugLogger } from '@/utils/debugLogger'
import type { 
  ProcessingTask, 
  ProcessingType,
  ProcessingParameters,
  ImageData
} from '@/types'
import type { WorkerResponse } from '@/types/worker-messages'
import type { WorkerStatus, WorkerPerformanceMetrics } from '@/types/worker-status'

/**
 * Enhanced WorkerOrchestratorModule with comprehensive debugging
 */
export class WorkerOrchestratorModule {
  private static instance: WorkerOrchestratorModule
  private workerPool: WorkerPoolManager
  private taskQueue: TaskQueueManager
  private onTaskUpdate?: (task: ProcessingTask) => void
  private initializationStartTime: number = 0

  constructor(workerCount?: number) {
    debugLogger.log('info', 'orchestrator', 'Creating WorkerOrchestrator', { workerCount })
    
    this.workerPool = new WorkerPoolManager(workerCount)
    this.taskQueue = new TaskQueueManager()
    
    // Set up cross-communication
    this.setupCommunication()
  }

  static getInstance(workerCount?: number): WorkerOrchestratorModule {
    if (!WorkerOrchestratorModule.instance) {
      WorkerOrchestratorModule.instance = new WorkerOrchestratorModule(workerCount)
    }
    return WorkerOrchestratorModule.instance
  }

  /**
   * Setup communication between components
   */
  private setupCommunication(): void {
    debugLogger.log('debug', 'orchestrator', 'Setting up component communication')
    
    // Set task update callback for queue manager
    this.taskQueue.setTaskUpdateCallback((task) => {
      debugLogger.log('debug', 'orchestrator', `Task update: ${task.id} -> ${task.status}`, {
        progress: task.progress,
        type: task.type
      })
      
      if (this.onTaskUpdate) {
        this.onTaskUpdate(task)
      }
    })

    // Set message handler for worker pool
    this.workerPool.setMessageHandler((event, worker) => {
      this.handleWorkerMessage(event, worker)
    })
  }

  /**
   * Initialize the orchestrator with comprehensive logging
   */
  async initialize(): Promise<void> {
    this.initializationStartTime = Date.now()
    debugLogger.log('info', 'orchestrator', '🎼 Initializing Worker Orchestrator...')
    
    try {
      await this.workerPool.initialize()
      
      const duration = Date.now() - this.initializationStartTime
      const status = this.workerPool.getStatus()
      
      debugLogger.log('info', 'orchestrator', '✅ Worker Orchestrator initialized successfully', {
        duration: `${duration}ms`,
        workers: `${status.availableWorkers}/${status.totalWorkers}`,
        environment: status.environment
      })
      
      // Test the orchestrator with a diagnostic task if in development
      if (import.meta.env.DEV) {
        setTimeout(() => this.runSelfDiagnostic(), 1000)
      }
      
    } catch (error) {
      const duration = Date.now() - this.initializationStartTime
      debugLogger.log('error', 'orchestrator', `❌ Failed to initialize Worker Orchestrator after ${duration}ms`, error)
      throw error
    }
  }

  /**
   * Run self-diagnostic to test orchestrator functionality
   */
  private async runSelfDiagnostic(): Promise<void> {
    try {
      debugLogger.log('info', 'orchestrator', '🔬 Running self-diagnostic...')
      
      const status = this.getWorkerStatus()
      debugLogger.log('debug', 'orchestrator', 'Current orchestrator status', status)
      
      if (status.availableWorkers === 0) {
        debugLogger.log('warn', 'orchestrator', 'No workers available for self-diagnostic')
        return
      }
      
      // Create a minimal test image
      const testImageData: ImageData = {
        data: new ArrayBuffer(100 * 100 * 4), // 100x100 RGBA
        width: 100,
        height: 100,
        channels: 4,
        format: 'PNG',
        filename: 'test.png',
        size: 40000
      }
      
      const testParams: ProcessingParameters = {
        binarization: {
          method: 'otsu',
          threshold: 128
        }
      }
      
      debugLogger.log('debug', 'orchestrator', 'Submitting diagnostic task...')
      const taskId = await this.submitTask(testImageData, 'binarization', testParams)
      debugLogger.log('info', 'orchestrator', `✅ Self-diagnostic task submitted: ${taskId}`)
      
    } catch (error) {
      debugLogger.log('error', 'orchestrator', '❌ Self-diagnostic failed', error)
    }
  }

  /**
   * Submit a processing task with enhanced logging
   */
  async submitTask(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): Promise<string> {
    const startTime = Date.now()
    debugLogger.log('info', 'orchestrator', `📋 Submitting task: ${type}`, {
      imageSize: `${imageData.width}x${imageData.height}`,
      dataSize: `${Math.round(imageData.data.byteLength / 1024)}KB`,
      parameters: Object.keys(parameters)
    })
    
    if (!this.workerPool.isReady()) {
      const error = 'Worker orchestrator not ready'
      debugLogger.log('error', 'orchestrator', error)
      throw new Error(error)
    }

    try {
      // Create and queue task
      const task = this.taskQueue.createTask(imageData, type, parameters)
      debugLogger.log('debug', 'orchestrator', `Created task: ${task.id}`)
      
      this.taskQueue.queueTask(task)
      debugLogger.log('debug', 'orchestrator', `Task queued: ${task.id}`)
      
      const queueStatus = this.taskQueue.getQueueStatus()
      const poolStatus = this.workerPool.getStatus()
      
      debugLogger.log('debug', 'orchestrator', 'Queue and pool status', {
        queueLength: queueStatus.length,
        availableWorkers: poolStatus.availableWorkers,
        activeWorkers: poolStatus.activeWorkers
      })
      
      // Process the queue with the imageData
      await this.processQueue(imageData)
      
      const duration = Date.now() - startTime
      debugLogger.log('info', 'orchestrator', `✅ Task submitted successfully: ${task.id}`, {
        duration: `${duration}ms`
      })
      
      return task.id
    } catch (error) {
      const duration = Date.now() - startTime
      debugLogger.log('error', 'orchestrator', `❌ Failed to submit task after ${duration}ms`, error)
      throw error
    }
  }

  /**
   * Cancel a task with logging
   */
  cancelTask(taskId: string): boolean {
    debugLogger.log('info', 'orchestrator', `🚫 Cancelling task: ${taskId}`)
    
    // Try to remove from queue first
    if (this.taskQueue.removeFromQueue(taskId)) {
      debugLogger.log('info', 'orchestrator', `✅ Task ${taskId} removed from queue`)
      return true
    }

    // Cancel active task
    const activeTask = this.workerPool.getTaskMapping(taskId)
    if (activeTask) {
      try {
        activeTask.worker.postMessage({
          id: taskId,
          type: 'cancel'
        })
        
        activeTask.task.status = 'cancelled'
        const mapping = this.workerPool.removeTaskMapping(taskId)
        if (mapping) {
          this.workerPool.returnWorker(mapping.worker)
        }
        
        this.taskQueue.updateTask(taskId, { status: 'cancelled' })
        this.processQueue()
        
        debugLogger.log('info', 'orchestrator', `✅ Active task ${taskId} cancelled`)
        return true
      } catch (error) {
        debugLogger.log('error', 'orchestrator', `❌ Failed to cancel task ${taskId}`, error)
      }
    }

    debugLogger.log('warn', 'orchestrator', `⚠️ Task ${taskId} not found`)
    return false
  }

  /**
   * Process the task queue with enhanced logging
   */
  private async processQueue(imageData?: ImageData): Promise<void> {
    const queueStatus = this.taskQueue.getQueueStatus()
    const poolStatus = this.workerPool.getStatus()
    
    debugLogger.log('debug', 'orchestrator', `🔄 Processing queue`, {
      queued: queueStatus.length,
      availableWorkers: poolStatus.availableWorkers,
      activeWorkers: poolStatus.activeWorkers
    })
    
    let tasksProcessed = 0
    
    while (this.taskQueue.hasPendingTasks() && poolStatus.availableWorkers > 0) {
      const task = this.taskQueue.getNextTask()
      const worker = this.workerPool.getAvailableWorker()
      
      if (!task || !worker) {
        debugLogger.log('warn', 'orchestrator', '⚠️ No task or worker available, breaking queue processing')
        break
      }
      
      debugLogger.log('debug', 'orchestrator', `▶️ Assigning task ${task.id} to worker`)
      await this.executeTask(worker, task, imageData)
      tasksProcessed++
      
      // Update pool status for next iteration
      const updatedPoolStatus = this.workerPool.getStatus()
      if (updatedPoolStatus.availableWorkers === 0) {
        debugLogger.log('info', 'orchestrator', '⚠️ No more available workers, stopping queue processing')
        break
      }
    }
    
    const finalQueueStatus = this.taskQueue.getQueueStatus()
    const finalPoolStatus = this.workerPool.getStatus()
    
    debugLogger.log('info', 'orchestrator', `📊 Queue processing completed`, {
      tasksProcessed,
      remainingQueued: finalQueueStatus.length,
      availableWorkers: finalPoolStatus.availableWorkers
    })
    
    if (finalQueueStatus.length > 0 && finalPoolStatus.availableWorkers === 0) {
      debugLogger.log('warn', 'orchestrator', `⚠️ ${finalQueueStatus.length} tasks waiting for available workers`)
    }
  }

  /**
   * Execute a task on a worker with comprehensive logging
   */
  private async executeTask(worker: Worker, task: ProcessingTask, imageData?: ImageData): Promise<void> {
    const startTime = Date.now()
    
    try {
      debugLogger.log('info', 'orchestrator', `🎯 Executing task ${task.id}`, {
        type: task.type,
        imageSize: imageData ? `${imageData.width}x${imageData.height}` : 'unknown'
      })
      
      task.status = 'processing'
      this.workerPool.addTaskMapping(task.id, worker, task)
      this.taskQueue.updateTask(task.id, { status: 'processing' })

      // Send task to worker
      const message = this.taskQueue.createWorkerMessage(task, imageData)
      debugLogger.log('debug', 'orchestrator', `📤 Sending message to worker for task ${task.id}`)
      
      // Use transferable objects for imageData if it's an ArrayBuffer
      const transferables: Transferable[] = []
      if (imageData?.data && imageData.data instanceof ArrayBuffer) {
        transferables.push(imageData.data)
      }
      
      worker.postMessage(message, transferables)
      
      debugLogger.log('debug', 'orchestrator', `✅ Task ${task.id} sent to worker`, {
        duration: `${Date.now() - startTime}ms`
      })
      
    } catch (error) {
      const duration = Date.now() - startTime
      debugLogger.log('error', 'orchestrator', `❌ Failed to execute task ${task.id} after ${duration}ms`, error)
      
      task.status = 'failed'
      task.error = error instanceof Error ? error.message : 'Unknown error'
      task.completedAt = new Date()
      
      this.workerPool.removeTaskMapping(task.id)
      this.workerPool.returnWorker(worker)
      this.taskQueue.updateTask(task.id, { 
        status: 'failed', 
        error: task.error,
        completedAt: task.completedAt
      })
      this.processQueue()
    }
  }

  /**
   * Handle messages from workers with enhanced logging
   */
  private handleWorkerMessage(event: MessageEvent<WorkerResponse>, worker: Worker): void {
    const message = event.data
    const { id, type } = message
    const payload = message.payload
    
    // Don't log test messages unless in debug mode
    if (id.startsWith('test-') && !import.meta.env.DEV) {
      return
    }
    
    debugLogger.log('debug', 'orchestrator', `📨 Worker message: ${type} for task ${id}`, payload)
    
    const activeTask = this.workerPool.getTaskMapping(id)
    
    if (!activeTask) {
      debugLogger.log('warn', 'orchestrator', `⚠️ Received message for unknown task: ${id}`)
      return
    }

    const { task } = activeTask

    switch (type) {
      case 'progress':
        if (payload && typeof payload.progress === 'number') {
          debugLogger.log('debug', 'orchestrator', `📊 Task ${id} progress: ${payload.progress}%`, {
            message: payload.message
          })
          this.taskQueue.updateTask(id, { progress: payload.progress })
        }
        break

      case 'result':
        if (payload && payload.result) {
          const duration = Date.now() - task.createdAt.getTime()
          debugLogger.log('info', 'orchestrator', `✅ Task ${id} completed successfully`, {
            duration: `${duration}ms`,
            resultSize: `${Math.round(payload.result.byteLength / 1024)}KB`
          })
          
          task.status = 'completed'
          task.progress = 100
          task.result = payload.result
          task.completedAt = new Date()
          
          this.workerPool.removeTaskMapping(id)
          this.workerPool.returnWorker(worker)
          this.taskQueue.updateTask(id, {
            status: 'completed',
            progress: 100,
            result: payload.result,
            completedAt: task.completedAt
          })
          
          // Continue processing queue
          this.processQueue()
        }
        break

      case 'error':
        if (payload && payload.error) {
          const duration = Date.now() - task.createdAt.getTime()
          debugLogger.log('error', 'orchestrator', `❌ Task ${id} failed after ${duration}ms`, {
            error: payload.error
          })
          
          task.status = 'failed'
          task.error = payload.error
          task.completedAt = new Date()
          
          this.workerPool.removeTaskMapping(id)
          this.workerPool.returnWorker(worker)
          this.taskQueue.updateTask(id, {
            status: 'failed',
            error: payload.error,
            completedAt: task.completedAt
          })
          
          // Continue processing queue
          this.processQueue()
        }
        break

      case 'ready':
        debugLogger.log('info', 'orchestrator', `🟢 Worker ready: ${payload?.message || 'Worker initialized'}`)
        break

      default:
        debugLogger.log('warn', 'orchestrator', `Unknown message type: ${type}`, message)
    }
  }

  /**
   * Set callback for task updates
   */
  setTaskUpdateCallback(callback: (task: ProcessingTask) => void): void {
    this.onTaskUpdate = callback
  }

  /**
   * Get task status
   */
  getTaskStatus(taskId: string): ProcessingTask | null {
    // Check queue first
    const queuedTask = this.taskQueue.getTask(taskId)
    if (queuedTask) return queuedTask

    // Check active tasks
    const activeTask = this.workerPool.getTaskMapping(taskId)
    if (activeTask) return activeTask.task

    return null
  }

  /**
   * Get all active tasks
   */
  getActiveTasks(): ProcessingTask[] {
    const queuedTasks = this.taskQueue.getQueueStatus().tasks
    const allTasks: ProcessingTask[] = []
    
    // Add queued tasks (they're already ProcessingTask objects)
    queuedTasks.forEach(taskInfo => {
      const fullTask = this.taskQueue.getTask(taskInfo.id)
      if (fullTask) {
        allTasks.push(fullTask)
      }
    })
    
    // Add active tasks from worker mappings
    this.workerPool.getActiveTasks().forEach(mapping => {
      if (mapping.task) {
        allTasks.push(mapping.task)
      }
    })
    
    return allTasks
  }

  /**
   * Get worker pool status with enhanced logging
   */
  getWorkerStatus(): WorkerStatus {
    const poolStatus = this.workerPool.getStatus()
    const queueStatus = this.taskQueue.getQueueStatus()
    
    const status: WorkerStatus = {
      ...poolStatus,
      queuedTasks: queueStatus.length
    }
    
    // Log status periodically in debug mode
    if (import.meta.env.DEV && Math.random() < 0.1) { // 10% chance
      debugLogger.log('debug', 'orchestrator', 'Current worker status', status)
    }
    
    return status
  }

  /**
   * Get performance metrics
   */
  getPerformanceMetrics(): WorkerPerformanceMetrics {
    const metrics = this.workerPool.getPerformanceMetrics()
    debugLogger.log('debug', 'orchestrator', 'Performance metrics', metrics)
    return metrics
  }

  /**
   * Clear completed tasks
   */
  clearCompletedTasks(): void {
    const beforeCount = this.taskQueue.getAllTasks().length
    this.taskQueue.cleanup()
    const afterCount = this.taskQueue.getAllTasks().length
    const removed = beforeCount - afterCount
    
    debugLogger.log('info', 'orchestrator', `🧹 Cleared ${removed} completed tasks`)
  }

  /**
   * Check if orchestrator is ready
   */
  isReady(): boolean {
    const ready = this.workerPool.isReady()
    if (!ready) {
      debugLogger.log('warn', 'orchestrator', 'Orchestrator not ready', {
        poolReady: this.workerPool.isReady(),
        poolStatus: this.workerPool.getStatus()
      })
    }
    return ready
  }

  /**
   * Wait for orchestrator to be ready
   */
  async waitForReady(timeout: number = 10000): Promise<boolean> {
    debugLogger.log('info', 'orchestrator', `⏳ Waiting for orchestrator to be ready (timeout: ${timeout}ms)`)
    
    const startTime = Date.now()
    const ready = await this.workerPool.waitForReady(timeout)
    const duration = Date.now() - startTime
    
    if (ready) {
      debugLogger.log('info', 'orchestrator', `✅ Orchestrator ready after ${duration}ms`)
    } else {
      debugLogger.log('error', 'orchestrator', `❌ Orchestrator not ready after ${duration}ms timeout`)
    }
    
    return ready
  }

  /**
   * Pause processing
   */
  pause(): void {
    debugLogger.log('info', 'orchestrator', '⏸️ Pausing worker orchestrator')
    this.workerPool.pause()
  }

  /**
   * Resume processing
   */
  resume(): void {
    debugLogger.log('info', 'orchestrator', '▶️ Resuming worker orchestrator')
    this.workerPool.resume()
    this.processQueue()
  }

  /**
   * Destroy orchestrator with comprehensive cleanup
   */
  destroy(): void {
    debugLogger.log('info', 'orchestrator', '🗑️ Destroying worker orchestrator...')
    
    const startTime = Date.now()
    
    // Cancel all queued tasks
    const queuedCount = this.taskQueue.getQueueStatus().length
    this.taskQueue.clearQueue()
    
    // Destroy worker pool
    this.workerPool.destroy()
    
    const duration = Date.now() - startTime
    debugLogger.log('info', 'orchestrator', `✅ Worker orchestrator destroyed`, {
      duration: `${duration}ms`,
      cancelledTasks: queuedCount
    })
  }

  /**
   * Get queue statistics
   */
  getQueueStatistics() {
    const stats = this.taskQueue.getStatistics()
    debugLogger.log('debug', 'orchestrator', 'Queue statistics', stats)
    return stats
  }

  /**
   * Clean up old tasks
   */
  cleanupOldTasks(maxAge: number = 60000): number {
    const removed = this.taskQueue.cleanup(maxAge)
    if (removed > 0) {
      debugLogger.log('info', 'orchestrator', `🧹 Cleaned up ${removed} old tasks`)
    }
    return removed
  }

  /**
   * Get detailed status for debugging
   */
  getDetailedStatus() {
    const poolStatus = this.workerPool.getDetailedStatus()
    const queueStats = this.taskQueue.getStatistics()
    const activeTasks = this.getActiveTasks()
    
    const detailedStatus = {
      pool: poolStatus,
      queue: queueStats,
      activeTasks: activeTasks.length,
      performance: this.getPerformanceMetrics(),
      uptime: Date.now() - this.initializationStartTime
    }
    
    debugLogger.log('debug', 'orchestrator', 'Detailed status', detailedStatus)
    return detailedStatus
  }

  /**
   * Run comprehensive diagnostics
   */
  async runDiagnostics(): Promise<any> {
    debugLogger.log('info', 'orchestrator', '🔬 Running orchestrator diagnostics...')
    
    try {
      const diagnostics = {
        timestamp: new Date().toISOString(),
        orchestrator: {
          ready: this.isReady(),
          uptime: Date.now() - this.initializationStartTime,
          status: this.getDetailedStatus()
        },
        workerPool: await this.workerPool.getDetailedStatus(),
        taskQueue: this.taskQueue.getStatistics(),
        performance: this.getPerformanceMetrics()
      }
      
      debugLogger.log('info', 'orchestrator', '✅ Diagnostics completed', diagnostics)
      return diagnostics
      
    } catch (error) {
      debugLogger.log('error', 'orchestrator', '❌ Diagnostics failed', error)
      throw error
    }
  }

  /**
   * Test orchestrator functionality
   */
  async testFunctionality(): Promise<boolean> {
    debugLogger.log('info', 'orchestrator', '🧪 Testing orchestrator functionality...')
    
    try {
      if (!this.isReady()) {
        debugLogger.log('error', 'orchestrator', 'Cannot test - orchestrator not ready')
        return false
      }
      
      // Create a minimal test
      const testImageData: ImageData = {
        data: new ArrayBuffer(10 * 10 * 4),
        width: 10,
        height: 10,
        channels: 4,
        format: 'PNG',
        filename: 'test.png',
        size: 400
      }
      
      const testParams: ProcessingParameters = {
        binarization: { method: 'otsu', threshold: 128 }
      }
      
      const taskId = await this.submitTask(testImageData, 'binarization', testParams)
      debugLogger.log('info', 'orchestrator', `✅ Test task submitted: ${taskId}`)
      return true
      
    } catch (error) {
      debugLogger.log('error', 'orchestrator', '❌ Functionality test failed', error)
      return false
    }
  }
}