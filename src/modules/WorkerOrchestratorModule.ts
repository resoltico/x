// src/modules/WorkerOrchestratorModule.ts
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
 * Enhanced with better error handling and logging
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
  private initializationError: string | null = null

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
   * Initialize workers with enhanced error handling
   */
  async initialize(): Promise<void> {
    if (this.isInitialized) return

    console.group('🚀 Initializing Worker Orchestrator')
    console.log(`Target worker count: ${this.workerCount}`)
    console.log(`Available CPU cores: ${navigator.hardwareConcurrency || 'unknown'}`)
    console.log(`Environment: ${this.detectEnvironment()}`)

    try {
      const workerUrls = this.getWorkerUrls()
      console.log('Worker URLs to try:', workerUrls)

      for (let i = 0; i < this.workerCount; i++) {
        try {
          const worker = await this.createWorker(i, workerUrls)
          if (worker) {
            this.workers.push(worker)
            this.availableWorkers.push(worker)
            console.log(`✅ Worker ${i} initialized successfully`)
          }
        } catch (workerError) {
          console.error(`❌ Failed to create worker ${i}:`, workerError)
        }
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
      console.groupEnd()
    } catch (error) {
      this.initializationError = error instanceof Error ? error.message : 'Unknown initialization error'
      console.error('❌ Failed to initialize worker pool:', error)
      console.groupEnd()
      throw error
    }
  }

  /**
   * Detect current environment
   */
  private detectEnvironment(): string {
    if (typeof window === 'undefined') return 'Node.js/SSR'
    if (window.location.protocol === 'file:') return 'File Protocol'
    if (window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1') {
      return `Development (${window.location.port ? 'port ' + window.location.port : 'no port'})`
    }
    return 'Production'
  }

  /**
   * Get possible worker URLs based on environment
   */
  private getWorkerUrls(): string[] {
    const urls: string[] = []
    
    // For development (Vite dev server)
    if (this.detectEnvironment().includes('Development')) {
      urls.push('/src/workers/imageProcessingWorker.ts')
    }
    
    // For production build - try multiple common paths
    urls.push('/workers/imageProcessingWorker.js')
    urls.push('/assets/imageProcessingWorker.js')
    
    // Try to find any worker file in assets
    urls.push('/assets/imageProcessingWorker-*.js')
    urls.push('/workers/imageProcessingWorker-*.js')
    
    // Fallback inline worker (always works)
    urls.push('data:text/javascript;base64,' + btoa(this.getInlineWorkerCode()))
    
    return urls
  }

  /**
   * Create a worker with multiple URL attempts
   */
  private async createWorker(index: number, urls: string[]): Promise<Worker | null> {
    for (const url of urls) {
      try {
        console.log(`Attempting to create worker ${index} with URL: ${url}`)
        
        // Handle wildcard URLs by trying to resolve them
        if (url.includes('*')) {
          // Skip wildcard URLs for now - they need special handling
          continue
        }
        
        const options: WorkerOptions = { 
          type: url.endsWith('.ts') ? 'module' : 'classic',
          name: `image-worker-${index}`
        }
        
        const worker = new Worker(url, options)
        
        // Test the worker with a timeout
        const testResult = await this.testWorker(worker, 3000)
        if (testResult) {
          this.setupWorkerEventHandlers(worker, index)
          return worker
        } else {
          worker.terminate()
        }
      } catch (error) {
        console.warn(`Worker ${index} failed with URL ${url}:`, error)
        continue
      }
    }
    
    console.error(`❌ All URLs failed for worker ${index}`)
    return null
  }

  /**
   * Test if a worker is responsive
   */
  private testWorker(worker: Worker, timeout: number = 3000): Promise<boolean> {
    return new Promise((resolve) => {
      let resolved = false
      const testId = `test-${Date.now()}-${Math.random()}`
      
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
      
      // Send test message
      try {
        worker.postMessage({ id: testId, type: 'test' })
      } catch (error) {
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
        this.handleWorkerMessage(event)
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
   * Get inline worker code as fallback
   */
  private getInlineWorkerCode(): string {
    return `
      console.log('🔧 Inline fallback worker starting...');
      
      // Simple binarization implementation
      function simpleBinarization(imageData, threshold = 128) {
        const data = new Uint8ClampedArray(imageData.data);
        
        for (let i = 0; i < data.length; i += 4) {
          const gray = data[i] * 0.299 + data[i + 1] * 0.587 + data[i + 2] * 0.114;
          const binary = gray > threshold ? 255 : 0;
          data[i] = binary;
          data[i + 1] = binary;
          data[i + 2] = binary;
        }
        
        return data.buffer;
      }
      
      // Simple scaling implementation
      function simpleScale(imageData, factor) {
        const canvas = new OffscreenCanvas(imageData.width, imageData.height);
        const ctx = canvas.getContext('2d');
        const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
        ctx.putImageData(canvasImageData, 0, 0);
        
        const scaledCanvas = new OffscreenCanvas(
          Math.round(imageData.width * factor),
          Math.round(imageData.height * factor)
        );
        const scaledCtx = scaledCanvas.getContext('2d');
        scaledCtx.imageSmoothingEnabled = false;
        scaledCtx.drawImage(canvas, 0, 0, scaledCanvas.width, scaledCanvas.height);
        
        return scaledCanvas.convertToBlob().then(blob => blob.arrayBuffer());
      }
      
      self.onmessage = async function(event) {
        const { id, type, payload } = event.data;
        
        console.log('🔧 Fallback worker received:', type, 'for task:', id);
        
        if (type === 'test') {
          self.postMessage({ id, type: 'test-response' });
          return;
        }
        
        if (type === 'process') {
          try {
            const { imageData, type: processType, parameters } = payload;
            
            // Send progress updates
            self.postMessage({
              id, type: 'progress',
              payload: { progress: 25, message: 'Processing with fallback worker...' }
            });
            
            let result;
            
            // Simple processing based on type
            switch (processType) {
              case 'binarization':
                const threshold = parameters.binarization?.threshold || 128;
                result = simpleBinarization(imageData, threshold);
                break;
                
              case 'scaling':
                const factor = parameters.scaling?.factor || 2;
                result = await simpleScale(imageData, factor);
                break;
                
              default:
                // For other types, just return a copy of the original
                result = imageData.data.slice(0);
            }
            
            self.postMessage({
              id, type: 'progress',
              payload: { progress: 75, message: 'Finalizing...' }
            });
            
            // Send result
            self.postMessage({
              id, type: 'result',
              payload: { result }
            }, result instanceof ArrayBuffer ? [result] : []);
            
          } catch (error) {
            console.error('🔧 Fallback worker error:', error);
            self.postMessage({
              id, type: 'error',
              payload: { error: 'Fallback processing failed: ' + error.message }
            });
          }
        }
      };
      
      console.log('🔧 Inline fallback worker initialized and ready');
    `
  }

  /**
   * Set callback for task updates
   */
  setTaskUpdateCallback(callback: (task: ProcessingTask) => void) {
    this.onTaskUpdate = callback
  }

  /**
   * Submit a processing task with enhanced logging
   */
  async submitTask(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): Promise<string> {
    console.group(`📋 Submitting Task: ${type}`)
    
    if (!this.isInitialized) {
      const error = this.initializationError || 'Worker orchestrator not initialized'
      console.error('❌', error)
      console.groupEnd()
      throw new Error(error)
    }

    if (this.workers.length === 0) {
      const error = 'No workers available for processing'
      console.error('❌', error)
      console.groupEnd()
      throw new Error(error)
    }

    const task: ProcessingTask = {
      id: this.generateTaskId(),
      type,
      parameters,
      status: 'pending',
      progress: 0,
      createdAt: new Date()
    }

    console.log(`Task ID: ${task.id}`)
    console.log(`Image: ${imageData.width}x${imageData.height} (${imageData.format})`)
    console.log(`Parameters:`, parameters)
    console.log(`Available workers: ${this.availableWorkers.length}/${this.workers.length}`)
    console.log(`Queue length: ${this.taskQueue.length}`)
    
    this.taskQueue.push(task)
    this.notifyTaskUpdate(task)
    
    // Process the queue with the image data
    await this.processQueue(imageData)
    
    console.groupEnd()
    return task.id
  }

  /**
   * Cancel a task with logging
   */
  cancelTask(taskId: string): boolean {
    console.log(`🚫 Cancelling task: ${taskId}`)
    
    // Remove from queue if pending
    const queueIndex = this.taskQueue.findIndex(task => task.id === taskId)
    if (queueIndex !== -1) {
      const task = this.taskQueue[queueIndex]
      task.status = 'cancelled'
      this.taskQueue.splice(queueIndex, 1)
      this.notifyTaskUpdate(task)
      console.log(`✅ Task ${taskId} removed from queue`)
      return true
    }

    // Cancel active task
    const activeTask = this.activeTasksMap.get(taskId)
    if (activeTask) {
      try {
        activeTask.worker.postMessage({
          id: taskId,
          type: 'cancel'
        } as WorkerMessage)
        
        activeTask.task.status = 'cancelled'
        this.activeTasksMap.delete(taskId)
        this.availableWorkers.push(activeTask.worker)
        this.notifyTaskUpdate(activeTask.task)
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
   * Process the task queue with enhanced logging
   */
  private async processQueue(imageData?: ImageData) {
    console.log(`🔄 Processing queue: ${this.taskQueue.length} queued, ${this.availableWorkers.length} available workers`)
    
    while (this.taskQueue.length > 0 && this.availableWorkers.length > 0) {
      const task = this.taskQueue.shift()!
      const worker = this.availableWorkers.shift()!
      
      console.log(`▶️ Assigning task ${task.id} to worker`)
      await this.executeTask(worker, task, imageData)
    }
    
    if (this.taskQueue.length > 0 && this.availableWorkers.length === 0) {
      console.warn(`⚠️ ${this.taskQueue.length} tasks waiting for available workers`)
    }
  }

  /**
   * Execute a task on a worker with enhanced error handling
   */
  private async executeTask(worker: Worker, task: ProcessingTask, imageData?: ImageData) {
    try {
      console.log(`🎯 Executing task ${task.id} of type ${task.type}`)
      
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

      console.log(`📤 Sending message to worker for task ${task.id}`)
      worker.postMessage(message)
    } catch (error) {
      console.error(`❌ Failed to execute task ${task.id}:`, error)
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
   * Handle messages from workers with enhanced logging
   */
  private handleWorkerMessage(event: MessageEvent<WorkerResponse>) {
    const { id, type, payload } = event.data
    
    // Don't log test messages
    if (id.startsWith('test-')) {
      return
    }
    
    console.log(`📨 Worker message: ${type} for task ${id}`)
    
    const activeTask = this.activeTasksMap.get(id)
    
    if (!activeTask) {
      console.warn(`⚠️ Received message for unknown task: ${id}`)
      return
    }

    const { worker, task } = activeTask

    switch (type) {
      case 'progress':
        if (payload && typeof payload.progress === 'number') {
          console.log(`📊 Task ${id} progress: ${payload.progress}%`)
          task.progress = payload.progress
          this.notifyTaskUpdate(task)
        }
        break

      case 'result':
        if (payload && payload.result) {
          console.log(`✅ Task ${id} completed successfully`)
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
          console.error(`❌ Task ${id} failed: ${payload.error}`)
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
   * Handle worker errors with enhanced logging
   */
  private handleWorkerError(event: ErrorEvent, index?: number) {
    console.error(`💥 Worker ${index ?? 'unknown'} error:`, {
      message: event.message,
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      error: event.error
    })
    
    // Find and handle any tasks using the failed worker
    for (const [taskId, { worker, task }] of this.activeTasksMap.entries()) {
      if (worker === event.target) {
        console.log(`🔄 Failing task ${taskId} due to worker error`)
        task.status = 'failed'
        task.error = 'Worker error: ' + (event.message || 'Unknown worker error')
        task.completedAt = new Date()
        this.activeTasksMap.delete(taskId)
        this.notifyTaskUpdate(task)
        break
      }
    }

    // Try to replace the failed worker
    this.replaceFailedWorker(event.target as Worker, index)
  }

  /**
   * Replace a failed worker with logging
   */
  private replaceFailedWorker(failedWorker: Worker, index?: number) {
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
        const workerUrls = this.getWorkerUrls()
        this.createWorker(Date.now(), workerUrls).then(newWorker => {
          if (newWorker) {
            this.workers.push(newWorker)
            this.availableWorkers.push(newWorker)
            console.log(`✅ Successfully replaced failed worker`)
          } else {
            console.error(`❌ Failed to create replacement worker`)
          }
        }).catch(createError => {
          console.error(`❌ Failed to create replacement worker:`, createError)
        })
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
   * Get worker pool status with enhanced information
   */
  getWorkerStatus() {
    return {
      totalWorkers: this.workers.length,
      availableWorkers: this.availableWorkers.length,
      activeWorkers: this.workers.length - this.availableWorkers.length,
      queuedTasks: this.taskQueue.length,
      activeTasks: this.activeTasksMap.size,
      initialized: this.isInitialized,
      initializationError: this.initializationError,
      environment: this.detectEnvironment()
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
    console.log('⏸️ Pausing worker orchestrator')
    // Move all available workers to a paused state
    // This prevents new tasks from being assigned
    this.availableWorkers = []
  }

  /**
   * Resume processing
   */
  resume() {
    console.log('▶️ Resuming worker orchestrator')
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
    console.log('🗑️ Destroying worker orchestrator')
    
    // Cancel all active tasks
    for (const [taskId] of this.activeTasksMap) {
      this.cancelTask(taskId)
    }

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
    this.taskQueue = []
    this.activeTasksMap.clear()
    this.isInitialized = false
    this.initializationError = null
    
    console.log('✅ Worker orchestrator destroyed')
  }
}