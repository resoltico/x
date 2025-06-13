// src/modules/worker/TaskQueueManager.ts
// Manages task queuing and execution logic

import type { ProcessingTask, ProcessingType, ProcessingParameters, ImageData } from '@/types'
import type { WorkerMessage } from '@/types/worker-messages'

export interface TaskSubmissionResult {
  taskId: string
  task: ProcessingTask
}

export class TaskQueueManager {
  private taskQueue: ProcessingTask[] = []
  private allTasks = new Map<string, ProcessingTask>() // Store all tasks for lookup
  private onTaskUpdate?: (task: ProcessingTask) => void

  /**
   * Generate unique task ID
   */
  generateTaskId(): string {
    return `task_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  /**
   * Create a new processing task
   */
  createTask(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): ProcessingTask {
    const task: ProcessingTask = {
      id: this.generateTaskId(),
      type,
      parameters,
      status: 'pending',
      progress: 0,
      createdAt: new Date()
    }

    return task
  }

  /**
   * Add task to queue
   */
  queueTask(task: ProcessingTask): void {
    console.group(`📋 Queuing Task: ${task.type}`)
    console.log(`Task ID: ${task.id}`)
    console.log(`Parameters:`, task.parameters)
    
    this.taskQueue.push(task)
    this.allTasks.set(task.id, task)
    this.notifyTaskUpdate(task)
    
    console.log(`Queue length: ${this.taskQueue.length}`)
    console.log(`Total tasks tracked: ${this.allTasks.size}`)
    console.groupEnd()
  }

  /**
   * Get next task from queue
   */
  getNextTask(): ProcessingTask | undefined {
    const task = this.taskQueue.shift()
    if (task) {
      console.log(`📤 Dequeued task: ${task.id}`)
    }
    return task
  }

  /**
   * Remove task from queue
   */
  removeFromQueue(taskId: string): boolean {
    const index = this.taskQueue.findIndex(task => task.id === taskId)
    if (index !== -1) {
      const task = this.taskQueue[index]
      task.status = 'cancelled'
      this.taskQueue.splice(index, 1)
      this.allTasks.set(taskId, task) // Keep cancelled task in allTasks for tracking
      this.notifyTaskUpdate(task)
      console.log(`✅ Task ${taskId} removed from queue`)
      return true
    }
    return false
  }

  /**
   * Get queue status
   */
  getQueueStatus() {
    return {
      length: this.taskQueue.length,
      tasks: this.taskQueue.map(task => ({
        id: task.id,
        type: task.type,
        status: task.status,
        createdAt: task.createdAt
      }))
    }
  }

  /**
   * Clear all pending tasks
   */
  clearQueue(): void {
    console.log(`🧹 Clearing task queue (${this.taskQueue.length} tasks)`)
    
    // Mark all queued tasks as cancelled
    this.taskQueue.forEach(task => {
      task.status = 'cancelled'
      this.allTasks.set(task.id, task)
      this.notifyTaskUpdate(task)
    })
    
    this.taskQueue = []
  }

  /**
   * Create worker message for task
   */
  createWorkerMessage(task: ProcessingTask, imageData?: ImageData): WorkerMessage {
    return {
      id: task.id,
      type: 'process',
      payload: {
        type: task.type,
        parameters: task.parameters,
        imageData: imageData || null
      }
    }
  }

  /**
   * Update task status
   */
  updateTask(taskId: string, updates: Partial<ProcessingTask>): void {
    // Check if task exists in allTasks
    const existingTask = this.allTasks.get(taskId)
    if (!existingTask) {
      console.warn(`⚠️ Attempting to update unknown task: ${taskId}`)
      return
    }

    // Update the task
    const updatedTask = {
      ...existingTask,
      ...updates
    }
    
    // Set completion time if task is completed or failed
    if (updates.status === 'completed' || updates.status === 'failed') {
      updatedTask.completedAt = new Date()
    }
    
    this.allTasks.set(taskId, updatedTask)
    
    // Also update in queue if it's still there
    const queueIndex = this.taskQueue.findIndex(task => task.id === taskId)
    if (queueIndex !== -1) {
      this.taskQueue[queueIndex] = updatedTask
    }
    
    console.log(`📝 Task ${taskId} updated:`, updates)
    this.notifyTaskUpdate(updatedTask)
  }

  /**
   * Get task by ID (from all tasks)
   */
  getTask(taskId: string): ProcessingTask | undefined {
    return this.allTasks.get(taskId)
  }

  /**
   * Check if queue has pending tasks
   */
  hasPendingTasks(): boolean {
    return this.taskQueue.length > 0
  }

  /**
   * Get tasks by status
   */
  getTasksByStatus(status: ProcessingTask['status']): ProcessingTask[] {
    return Array.from(this.allTasks.values()).filter(task => task.status === status)
  }

  /**
   * Set task update callback
   */
  setTaskUpdateCallback(callback: (task: ProcessingTask) => void): void {
    this.onTaskUpdate = callback
  }

  /**
   * Notify about task updates
   */
  private notifyTaskUpdate(task: ProcessingTask): void {
    if (this.onTaskUpdate) {
      this.onTaskUpdate({ ...task })
    }
  }

  /**
   * Get queue statistics
   */
  getStatistics() {
    const now = Date.now()
    const allTasksArray = Array.from(this.allTasks.values())

    return {
      total: allTasksArray.length,
      pending: allTasksArray.filter(t => t.status === 'pending').length,
      processing: allTasksArray.filter(t => t.status === 'processing').length,
      completed: allTasksArray.filter(t => t.status === 'completed').length,
      failed: allTasksArray.filter(t => t.status === 'failed').length,
      cancelled: allTasksArray.filter(t => t.status === 'cancelled').length,
      averageAge: allTasksArray.length > 0 ? allTasksArray.reduce((sum, task) => sum + (now - task.createdAt.getTime()), 0) / allTasksArray.length : 0,
      oldestTask: allTasksArray.length > 0 ? Math.min(...allTasksArray.map(t => t.createdAt.getTime())) : 0
    }
  }

  /**
   * Estimate processing time for queue
   */
  estimateQueueProcessingTime(averageTaskTime: number = 5000): number {
    return this.taskQueue.length * averageTaskTime
  }

  /**
   * Get tasks older than specified time
   */
  getOldTasks(maxAge: number = 30000): ProcessingTask[] {
    const cutoff = Date.now() - maxAge
    return Array.from(this.allTasks.values()).filter(task => task.createdAt.getTime() < cutoff)
  }

  /**
   * Clean up old cancelled/failed tasks
   */
  cleanup(maxAge: number = 60000): number {
    const cutoff = Date.now() - maxAge
    const tasksToRemove: string[] = []
    
    this.allTasks.forEach((task, taskId) => {
      const isOld = task.createdAt.getTime() < cutoff
      const isFinished = ['cancelled', 'failed', 'completed'].includes(task.status)
      
      if (isOld && isFinished) {
        tasksToRemove.push(taskId)
      }
    })
    
    tasksToRemove.forEach(taskId => {
      this.allTasks.delete(taskId)
    })
    
    // Also clean up the queue
    const initialQueueLength = this.taskQueue.length
    this.taskQueue = this.taskQueue.filter(task => {
      const isOld = task.createdAt.getTime() < cutoff
      const isFinished = ['cancelled', 'failed'].includes(task.status)
      return !(isOld && isFinished)
    })
    
    const totalRemoved = tasksToRemove.length + (initialQueueLength - this.taskQueue.length)
    
    if (totalRemoved > 0) {
      console.log(`🧹 Cleaned up ${totalRemoved} old tasks`)
    }
    
    return totalRemoved
  }

  /**
   * Get all tasks (for debugging)
   */
  getAllTasks(): ProcessingTask[] {
    return Array.from(this.allTasks.values())
  }
}