// src/modules/processing/SystemStatusManager.ts
// Manages system status and initialization state

export interface SystemStatus {
  initialized: boolean
  totalWorkers: number
  availableWorkers: number
  queuedTasks: number
  environment: string
  error: string | null
}

export class SystemStatusManager {
  private static instance: SystemStatusManager
  private status: SystemStatus = {
    initialized: false,
    totalWorkers: 0,
    availableWorkers: 0,
    queuedTasks: 0,
    environment: 'Unknown',
    error: null
  }

  private initializationAttempts = 0
  private maxInitializationAttempts = 3
  private statusUpdateCallbacks: Set<(status: SystemStatus) => void> = new Set()

  static getInstance(): SystemStatusManager {
    if (!SystemStatusManager.instance) {
      SystemStatusManager.instance = new SystemStatusManager()
    }
    return SystemStatusManager.instance
  }

  /**
   * Update system status
   */
  updateStatus(updates: Partial<SystemStatus>): void {
    this.status = { ...this.status, ...updates }
    this.notifyStatusUpdate()
  }

  /**
   * Get current status
   */
  getStatus(): SystemStatus {
    return { ...this.status }
  }

  /**
   * Set initialization error
   */
  setInitializationError(error: string): void {
    this.status.error = error
    this.status.initialized = false
    this.notifyStatusUpdate()
  }

  /**
   * Clear initialization error
   */
  clearInitializationError(): void {
    this.status.error = null
    this.notifyStatusUpdate()
  }

  /**
   * Mark as initialized
   */
  markInitialized(): void {
    this.status.initialized = true
    this.status.error = null
    this.notifyStatusUpdate()
  }

  /**
   * Get initialization attempts
   */
  getInitializationAttempts(): number {
    return this.initializationAttempts
  }

  /**
   * Increment initialization attempts
   */
  incrementInitializationAttempts(): number {
    return ++this.initializationAttempts
  }

  /**
   * Reset initialization attempts
   */
  resetInitializationAttempts(): void {
    this.initializationAttempts = 0
  }

  /**
   * Check if max attempts reached
   */
  isMaxAttemptsReached(): boolean {
    return this.initializationAttempts >= this.maxInitializationAttempts
  }

  /**
   * Get max attempts
   */
  getMaxAttempts(): number {
    return this.maxInitializationAttempts
  }

  /**
   * Add status update callback
   */
  addStatusUpdateCallback(callback: (status: SystemStatus) => void): void {
    this.statusUpdateCallbacks.add(callback)
  }

  /**
   * Remove status update callback
   */
  removeStatusUpdateCallback(callback: (status: SystemStatus) => void): void {
    this.statusUpdateCallbacks.delete(callback)
  }

  /**
   * Notify all callbacks of status update
   */
  private notifyStatusUpdate(): void {
    this.statusUpdateCallbacks.forEach(callback => {
      try {
        callback(this.getStatus())
      } catch (error) {
        console.error('Error in status update callback:', error)
      }
    })
  }

  /**
   * Reset system status
   */
  reset(): void {
    this.status = {
      initialized: false,
      totalWorkers: 0,
      availableWorkers: 0,
      queuedTasks: 0,
      environment: 'Unknown',
      error: null
    }
    this.initializationAttempts = 0
    this.notifyStatusUpdate()
  }
}