import type { 
  ProcessingTask, 
  WorkerMessage, 
  WorkerResponse,
  ProcessingType,
  ProcessingParameters,
  ImageData
} from '@/types'

/**
 * WorkerOrchestratorModule manages Web Workers for non-blocking image processing
 */
export class WorkerOrchestratorModule {
  private static instance: WorkerOrchestratorModule
  private workers: Worker[] = []
  private availableWorkers: Worker[] = []
  private taskQueue: ProcessingTask[] = []
  private activeTasksMap = new Map<string, { worker: Worker; task: ProcessingTask }>()
  private workerCount: number
  private onTaskUpdate?: (task: ProcessingTask) => void

  constructor(workerCount: number = navigator.hardwareConcurrency || 4) {
    this.workerCount = Math.min(workerCount, 8) // Cap at 8 workers
  }

  static getInstance(workerCount?: number): WorkerOrchestratorModule {
    if (!WorkerOrchestratorModule.instance) {
      WorkerOrchestratorModule.instance = new WorkerOrchestratorModule(workerCount)
    }
    return WorkerOrchestratorModule.instance
  }

  /**
   * Initialize workers
   */
  async initialize(): Promise<void> {
    if (this.workers.length > 0) return

    try {
      for (let i = 0; i < this.workerCount; i++) {
        const worker = new Worker(
          new URL('../workers/imageProcessingWorker.ts', import.meta.url),
          { type: 'module' }
        )
        
        worker.onmessage = this.handleWorkerMessage.bind(this)
        worker.onerror = this.handleWorkerError.bind(this)
        
        this.workers.push(worker)
        this.availableWorkers.push(worker)
      }
      
      console.log(`Initialized ${this.workerCount} image processing workers`)
    } catch (error) {
      console.error('Failed to initialize workers:', error)
      throw new Error('Failed to initialize worker pool')
    }
  }

  /**
   * Set callback for task updates
   */
  setTaskUpdateCallback(callback: (task: ProcessingTask) => void) {
    this.onTaskUpdate = callback
  }

  /**
   * Submit a processing task
   */
  async submitTask(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): Promise<string> {
    const task: ProcessingTask = {
      id: this.generateTaskId(),
      type,
      parameters,
      status: 'pending',
      progress: 0,
      createdAt: new Date()
    }

    this.taskQueue.push(task)
    this.processQueue()
    
    return task.id
  }

  /**
   * Cancel a task
   */
  cancelTask(taskId: string): boolean {
    // Remove from queue if pending
    const queueIndex = this.taskQueue.findIndex(task => task.id === taskId)
    if (queueIndex !== -1) {
      const task = this.taskQueue[queueIndex]
      task.status = 'cancelled'
      this.taskQueue.splice(queueIndex, 1)
      this.notifyTaskUpdate(task)
      return true
    }

    // Cancel active task
    const activeTask = this.activeTasksMap.get(taskId)
    if (activeTask) {
      activeTask.worker.postMessage({
        id: taskId,
        type: 'cancel'
      } as WorkerMessage)
      
      activeTask.task.status = 'cancelled'
      this.activeTasksMap.delete(taskId)
      this.availableWorkers.push(activeTask.worker)
      this.notifyTaskUpdate(activeTask.task)
      this.processQueue()
      return true
    }

    return false
  }

  /**
   * Get task status
   */
  getTaskStatus(taskId: string): ProcessingTask | null {
    // Check queue
    const queuedTask = this.taskQueue.find(task => task.id === taskId)
    if (queuedTask) return queuedTask

    // Check active tasks
    const activeTask = this.activeTasksMap.get(taskId)
    if (activeTask) return activeTask.task

    return null
  }

  /**
   * Get all active tasks
   */
  getActiveTasks(): ProcessingTask[] {
    const queuedTasks = this.taskQueue.map(task => ({ ...task }))
    const activeTasks = Array.from(this.activeTasksMap.values()).map(({ task }) => ({ ...task }))
    return [...queuedTasks, ...activeTasks]
  }

  /**
   * Process the task queue
   */
  private processQueue() {
    while (this.taskQueue.length > 0 && this.availableWorkers.length > 0) {
      const task = this.taskQueue.shift()!
      const worker = this.availableWorkers.shift()!
      
      this.executeTask(worker, task)
    }
  }

  /**
   * Execute a task on a worker
   */
  private async executeTask(worker: Worker, task: ProcessingTask) {
    try {
      task.status = 'processing'
      this.activeTasksMap.set(task.id, { worker, task })
      this.notifyTaskUpdate(task)

      // Send task to worker
      const message: WorkerMessage = {
        id: task.id,
        type: 'process',
        payload: {
          type: task.type,
          parameters: task.parameters
        }
      }

      worker.postMessage(message)
    } catch (error) {
      console.error('Failed to execute task:', error)
      task.status = 'failed'
      task.error = error instanceof Error ? error.message : 'Unknown error'
      this.activeTasksMap.delete(task.id)
      this.availableWorkers.push(worker)
      this.notifyTaskUpdate(task)
      this.processQueue()
    }
  }

  /**
   * Handle messages from workers
   */
  private handleWorkerMessage(event: MessageEvent<WorkerResponse>) {
    const { id, type, payload } = event.data
    const activeTask = this.activeTasksMap.get(id)
    
    if (!activeTask) {
      console.warn('Received message for unknown task:', id)
      return
    }

    const { worker, task } = activeTask

    switch (type) {
      case 'progress':
        task.progress = payload.progress
        this.notifyTaskUpdate(task)
        break

      case 'result':
        task.status = 'completed'
        task.progress = 100
        task.result = payload.result
        task.completedAt = new Date()
        this.activeTasksMap.delete(id)
        this.availableWorkers.push(worker)
        this.notifyTaskUpdate(task)
        this.processQueue()
        break

      case 'error':
        task.status = 'failed'
        task.error = payload.error
        task.completedAt = new Date()
        this.activeTasksMap.delete(id)
        this.availableWorkers.push(worker)
        this.notifyTaskUpdate(task)
        this.processQueue()
        break
    }
  }

  /**
   * Handle worker errors
   */
  private handleWorkerError(event: ErrorEvent) {
    console.error('Worker error:', event.error)
    
    // Find and handle any tasks using the failed worker
    for (const [taskId, { worker, task }] of this.activeTasksMap.entries()) {
      if (worker === event.target) {
        task.status = 'failed'
        task.error = 'Worker error: ' + event.error?.message || 'Unknown worker error'
        task.completedAt = new Date()
        this.activeTasksMap.delete(taskId)
        this.notifyTaskUpdate(task)
        break
      }
    }

    // Try to replace the failed worker
    this.replaceFailedWorker(event.target as Worker)
  }

  /**
   * Replace a failed worker
   */
  private replaceFailedWorker(failedWorker: Worker) {
    try {
      // Remove from available workers
      const availableIndex = this.availableWorkers.indexOf(failedWorker)
      if (availableIndex !== -1) {
        this.availableWorkers.splice(availableIndex, 1)
      }

      // Remove from workers list
      const workerIndex = this.workers.indexOf(failedWorker)
      if (workerIndex !== -1) {
        this.workers.splice(workerIndex, 1)
        
        // Create replacement worker
        const newWorker = new Worker(
          new URL('../workers/imageProcessingWorker.ts', import.meta.url),
          { type: 'module' }
        )
        
        newWorker.onmessage = this.handleWorkerMessage.bind(this)
        newWorker.onerror = this.handleWorkerError.bind(this)
        
        this.workers.push(newWorker)
        this.availableWorkers.push(newWorker)
        
        console.log('Replaced failed worker')
      }
      
      // Terminate the failed worker
      failedWorker.terminate()
    } catch (error) {
      console.error('Failed to replace worker:', error)
    }
  }

  /**
   * Notify about task updates
   */
  private notifyTaskUpdate(task: ProcessingTask) {
    if (this.onTaskUpdate) {
      this.onTaskUpdate({ ...task })
    }
  }

  /**
   * Generate unique task ID
   */
  private generateTaskId(): string {
    return `task_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  /**
   * Get worker pool status
   */
  getWorkerStatus() {
    return {
      totalWorkers: this.workers.length,
      availableWorkers: this.availableWorkers.length,
      activeWorkers: this.workers.length - this.availableWorkers.length,
      queuedTasks: this.taskQueue.length,
      activeTasks: this.activeTasksMap.size
    }
  }

  /**
   * Clear completed tasks from memory
   */
  clearCompletedTasks() {
    // This would be called by the store to clean up completed tasks
    // The actual task storage is handled by the store
  }

  /**
   * Destroy all workers and clean up
   */
  destroy() {
    // Cancel all active tasks
    for (const [taskId] of this.activeTasksMap) {
      this.cancelTask(taskId)
    }

    // Terminate all workers
    this.workers.forEach(worker => worker.terminate())
    
    // Clear arrays
    this.workers = []
    this.availableWorkers = []
    this.taskQueue = []
    this.activeTasksMap.clear()
    
    console.log('Worker orchestrator destroyed')
  }
}