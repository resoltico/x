// src/modules/WorkerOrchestratorModule.ts
// Refactored orchestrator using modular components

import { WorkerPoolManager } from './worker/WorkerPoolManager'
import { TaskQueueManager } from './worker/TaskQueueManager'
import type { 
  ProcessingTask, 
  ProcessingType,
  ProcessingParameters,
  ImageData
} from '@/types'
import type { WorkerResponse } from '@/types/worker-messages'
import type { WorkerStatus, WorkerPerformanceMetrics } from '@/types/worker-status'

/**
 * Refactored WorkerOrchestratorModule using composition pattern
 * Delegates specific responsibilities to specialized managers
 */
export class WorkerOrchestratorModule {
  private static instance: WorkerOrchestratorModule
  private workerPool: WorkerPoolManager
  private taskQueue: TaskQueueManager
  private onTaskUpdate?: (task: ProcessingTask) => void

  constructor(workerCount?: number) {
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
    // Set task update callback for queue manager
    this.taskQueue.setTaskUpdateCallback((task) => {
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
   * Initialize the orchestrator
   */
  async initialize(): Promise<void> {
    console.group('🎼 Initializing Worker Orchestrator')
    
    try {
      await this.workerPool.initialize()
      console.log('✅ Worker Orchestrator initialized successfully')
      console.groupEnd()
    } catch (error) {
      console.error('❌ Failed to initialize Worker Orchestrator:', error)
      console.groupEnd()
      throw error
    }
  }

  /**
   * Submit a processing task
   */
  async submitTask(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): Promise<string> {
    console.group(`📋 Submitting Task: ${type}`)
    
    if (!this.workerPool.isReady()) {
      const error = 'Worker orchestrator not ready'
      console.error('❌', error)
      console.groupEnd()
      throw new Error(error)
    }

    try {
      // Create and queue task
      const task = this.taskQueue.createTask(imageData, type, parameters)
      this.taskQueue.queueTask(task)
      
      console.log(`Task ID: ${task.id}`)
      console.log(`Queue status:`, this.taskQueue.getQueueStatus())
      
      // Process the queue
      await this.processQueue(imageData)
      
      console.groupEnd()
      return task.id
    } catch (error) {
      console.error('❌ Failed to submit task:', error)
      console.groupEnd()
      throw error
    }
  }

  /**
   * Cancel a task
   */
  cancelTask(taskId: string): boolean {
    console.log(`🚫 Cancelling task: ${taskId}`)
    
    // Try to remove from queue first
    if (this.taskQueue.removeFromQueue(taskId)) {
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
        
        console.log(`✅ Active task ${taskId} cancelled`)
        return true
      } catch (error) {
        console.error(`❌ Failed to cancel task ${taskId}:`, error)
      }
    }

    console.warn(`⚠️ Task ${taskId} not found`)
    return false
  }

  /**
   * Process the task queue
   */
  private async processQueue(imageData?: ImageData): Promise<void> {
    const queueStatus = this.taskQueue.getQueueStatus()
    const poolStatus = this.workerPool.getStatus()
    
    console.log(`🔄 Processing queue: ${queueStatus.length} queued, ${poolStatus.availableWorkers} available workers`)
    
    while (this.taskQueue.hasPendingTasks() && poolStatus.availableWorkers > 0) {
      const task = this.taskQueue.getNextTask()
      const worker = this.workerPool.getAvailableWorker()
      
      if (!task || !worker) break
      
      console.log(`▶️ Assigning task ${task.id} to worker`)
      await this.executeTask(worker, task, imageData)
    }
    
    const finalQueueStatus = this.taskQueue.getQueueStatus()
    const finalPoolStatus = this.workerPool.getStatus()
    
    if (finalQueueStatus.length > 0 && finalPoolStatus.availableWorkers === 0) {
      console.warn(`⚠️ ${finalQueueStatus.length} tasks waiting for available workers`)
    }
  }

  /**
   * Execute a task on a worker
   */
  private async executeTask(worker: Worker, task: ProcessingTask, imageData?: ImageData): Promise<void> {
    try {
      console.log(`🎯 Executing task ${task.id} of type ${task.type}`)
      
      task.status = 'processing'
      this.workerPool.addTaskMapping(task.id, worker, task)
      this.taskQueue.updateTask(task.id, { status: 'processing' })

      // Send task to worker
      const message = this.taskQueue.createWorkerMessage(task, imageData)
      console.log(`📤 Sending message to worker for task ${task.id}`)
      worker.postMessage(message)
    } catch (error) {
      console.error(`❌ Failed to execute task ${task.id}:`, error)
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
   * Handle messages from workers
   */
  private handleWorkerMessage(event: MessageEvent<WorkerResponse>, worker: Worker): void {
    const { id, type, payload } = event.data
    
    // Don't log test messages
    if (id.startsWith('test-')) {
      return
    }
    
    console.log(`📨 Worker message: ${type} for task ${id}`)
    
    const activeTask = this.workerPool.getTaskMapping(id)
    
    if (!activeTask) {
      console.warn(`⚠️ Received message for unknown task: ${id}`)
      return
    }

    const { task } = activeTask

    switch (type) {
      case 'progress':
        if (payload && typeof payload.progress === 'number') {
          console.log(`📊 Task ${id} progress: ${payload.progress}%`)
          this.taskQueue.updateTask(id, { progress: payload.progress })
        }
        break

      case 'result':
        if (payload && payload.result) {
          console.log(`✅ Task ${id} completed successfully`)
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
          this.processQueue()
        }
        break

      case 'error':
        if (payload && payload.error) {
          console.error(`❌ Task ${id} failed: ${payload.error}`)
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
          this.processQueue()
        }
        break
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
    const activeTasks = Array.from(this.workerPool.getStatus().activeTasks)
    
    // This is a simplified version - in reality we'd need to properly merge
    return [...queuedTasks] as ProcessingTask[]
  }

  /**
   * Get worker pool status
   */
  getWorkerStatus(): WorkerStatus {
    const poolStatus = this.workerPool.getStatus()
    const queueStatus = this.taskQueue.getQueueStatus()
    
    return {
      ...poolStatus,
      queuedTasks: queueStatus.length
    }
  }

  /**
   * Get performance metrics
   */
  getPerformanceMetrics(): WorkerPerformanceMetrics {
    return this.workerPool.getPerformanceMetrics()
  }

  /**
   * Clear completed tasks
   */
  clearCompletedTasks(): void {
    this.taskQueue.cleanup()
  }

  /**
   * Check if orchestrator is ready
   */
  isReady(): boolean {
    return this.workerPool.isReady()
  }

  /**
   * Wait for orchestrator to be ready
   */
  async waitForReady(timeout: number = 5000): Promise<boolean> {
    return this.workerPool.waitForReady(timeout)
  }

  /**
   * Pause processing
   */
  pause(): void {
    console.log('⏸️ Pausing worker orchestrator')
    this.workerPool.pause()
  }

  /**
   * Resume processing
   */
  resume(): void {
    console.log('▶️ Resuming worker orchestrator')
    this.workerPool.resume()
    this.processQueue()
  }

  /**
   * Destroy orchestrator
   */
  destroy(): void {
    console.log('🗑️ Destroying worker orchestrator')
    
    // Cancel all queued tasks
    this.taskQueue.clearQueue()
    
    // Destroy worker pool
    this.workerPool.destroy()
    
    console.log('✅ Worker orchestrator destroyed')
  }

  /**
   * Get queue statistics
   */
  getQueueStatistics() {
    return this.taskQueue.getStatistics()
  }

  /**
   * Clean up old tasks
   */
  cleanupOldTasks(maxAge: number = 60000): number {
    return this.taskQueue.cleanup(maxAge)
  }
}