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
  private isInitialized = false

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
    if (this.isInitialized) return

    try {
      console.log(`Initializing ${this.workerCount} workers...`)
      
      for (let i = 0; i < this.workerCount; i++) {
        try {
          // Create worker using the worker file path
          const worker = new Worker('/src/workers/imageProcessingWorker.ts', { 
            type: 'module',
            name: `image-worker-${i}`
          })
          
          worker.onmessage = this.handleWorkerMessage.bind(this)
          worker.onerror = this.handleWorkerError.bind(this)
          
          this.workers.push(worker)
          this.availableWorkers.push(worker)
          
          console.log(`Worker ${i} initialized successfully`)
        } catch (workerError) {
          console.error(`Failed to create worker ${i}:`, workerError)
        }
      }
      
      if (this.workers.length === 0) {
        throw new Error('No workers could be initialized')
      }
      
      this.isInitialized = true
      console.log(`Successfully initialized ${this.workers.length} out of ${this.workerCount} requested workers`)
    } catch (error) {
      console.error('Failed to initialize worker pool:', error)
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
    if (!this.isInitialized) {
      throw new Error('Worker orchestrator not initialized')
    }

    const task: ProcessingTask = {
      id: this.generateTaskId(),
      type,
      parameters,
      status: 'pending',
      progress: 0,
      createdAt: new Date()
    }

    console.log('Submitting task:', task.id, 'Type:', type)
    this.taskQueue.push(task)
    this.notifyTaskUpdate(task)
    
    // Process the queue with the image data
    await this.processQueue(imageData)
    
    return task.id
  }

  /**
   * Cancel a task
   */
  cancelTask(taskId: string): boolean {
    console.log('Cancelling task:', taskId)
    
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
  private async processQueue(imageData?: ImageData) {
    console.log(`Processing queue: ${this.taskQueue.length} queued, ${this.availableWorkers.length} available workers`)
    
    while (this.taskQueue.length > 0 && this.availableWorkers.length > 0) {
      const task = this.taskQueue.shift()!
      const worker = this.availableWorkers.shift()!
      
      console.log(`Assigning task ${task.id} to worker`)
      await this.executeTask(worker, task, imageData)
    }
  }

  /**
   * Execute a task on a worker
   */
  private async executeTask(worker: Worker, task: ProcessingTask, imageData?: ImageData) {
    try {
      console.log(`Executing task ${task.id} of type ${task.type}`)
      
      task.status = 'processing'
      this.activeTasksMap.set(task.id, { worker, task })
      this.notifyTaskUpdate(task)

      // Send task to worker with the image data
      const message: WorkerMessage = {
        id: task.id,
        type: 'process',
        payload: {
          type: task.type,
          parameters: task.parameters,
          imageData: imageData || null
        }
      }

      console.log('Sending message to worker:', message)
      worker.postMessage(message)
    } catch (error) {
      console.error('Failed to execute task:', error)
      task.status = 'failed'
      task.error = error instanceof Error ? error.message : 'Unknown error'
      task.completedAt = new Date()
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
    console.log('Received worker message:', { id, type, payload: payload ? 'present' : 'empty' })
    
    const activeTask = this.activeTasksMap.get(id)
    
    if (!activeTask) {
      console.warn('Received message for unknown task:', id)
      return
    }

    const { worker, task } = activeTask

    switch (type) {
      case 'progress':
        if (payload && typeof payload.progress === 'number') {
          console.log(`Task ${id} progress: ${payload.progress}%`)
          task.progress = payload.progress
          this.notifyTaskUpdate(task)
        }
        break

      case 'result':
        if (payload && payload.result) {
          console.log(`Task ${id} completed successfully`)
          task.status = 'completed'
          task.progress = 100
          task.result = payload.result
          task.completedAt = new Date()
          this.activeTasksMap.delete(id)
          this.availableWorkers.push(worker)
          this.notifyTaskUpdate(task)
          this.processQueue()
        }
        break

      case 'error':
        if (payload && payload.error) {
          console.error(`Task ${id} failed:`, payload.error)
          task.status = 'failed'
          task.error = payload.error
          task.completedAt = new Date()
          this.activeTasksMap.delete(id)
          this.availableWorkers.push(worker)
          this.notifyTaskUpdate(task)
          this.processQueue()
        }
        break
    }
  }

  /**
   * Handle worker errors
   */
  private handleWorkerError(event: ErrorEvent) {
    console.error('Worker error:', event.message, event.filename, event.lineno)
    
    // Find and handle any tasks using the failed worker
    for (const [taskId, { worker, task }] of this.activeTasksMap.entries()) {
      if (worker === event.target) {
        console.log(`Failing task ${taskId} due to worker error`)
        task.status = 'failed'
        task.error = 'Worker error: ' + (event.message || 'Unknown worker error')
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
      console.log('Replacing failed worker')
      
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
        try {
          const newWorker = new Worker('/src/workers/imageProcessingWorker.ts', { 
            type: 'module',
            name: `image-worker-replacement-${Date.now()}`
          })
          
          newWorker.onmessage = this.handleWorkerMessage.bind(this)
          newWorker.onerror = this.handleWorkerError.bind(this)
          
          this.workers.push(newWorker)
          this.availableWorkers.push(newWorker)
          
          console.log('Successfully replaced failed worker')
        } catch (createError) {
          console.error('Failed to create replacement worker:', createError)
        }
      }
      
      // Terminate the failed worker
      try {
        failedWorker.terminate()
      } catch (terminateError) {
        console.warn('Error terminating failed worker:', terminateError)
      }
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
      activeTasks: this.activeTasksMap.size,
      initialized: this.isInitialized
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
   * Get performance metrics
   */
  getPerformanceMetrics() {
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

    return {
      averageProcessingTime: totalTime / completedTasks.length,
      totalTasksCompleted: completedTasks.length,
      successRate: 100 // All completed tasks are successful by definition
    }
  }

  /**
   * Check if orchestrator is ready
   */
  isReady(): boolean {
    return this.isInitialized && this.workers.length > 0
  }

  /**
   * Wait for orchestrator to be ready
   */
  async waitForReady(timeout: number = 5000): Promise<boolean> {
    const startTime = Date.now()
    
    while (!this.isReady() && (Date.now() - startTime) < timeout) {
      await new Promise(resolve => setTimeout(resolve, 100))
    }
    
    return this.isReady()
  }

  /**
   * Pause processing (don't assign new tasks to workers)
   */
  pause() {
    // Move all available workers to a paused state
    // This prevents new tasks from being assigned
    this.availableWorkers = []
  }

  /**
   * Resume processing
   */
  resume() {
    // Restore available workers and process queue
    this.availableWorkers = this.workers.filter(worker => 
      !Array.from(this.activeTasksMap.values()).some(({ worker: activeWorker }) => activeWorker === worker)
    )
    this.processQueue()
  }

  /**
   * Destroy all workers and clean up
   */
  destroy() {
    console.log('Destroying worker orchestrator')
    
    // Cancel all active tasks
    for (const [taskId] of this.activeTasksMap) {
      this.cancelTask(taskId)
    }

    // Terminate all workers
    this.workers.forEach((worker, index) => {
      try {
        worker.terminate()
        console.log(`Terminated worker ${index}`)
      } catch (error) {
        console.warn(`Error terminating worker ${index}:`, error)
      }
    })
    
    // Clear arrays and maps
    this.workers = []
    this.availableWorkers = []
    this.taskQueue = []
    this.activeTasksMap.clear()
    this.isInitialized = false
    
    console.log('Worker orchestrator destroyed')
  }
}