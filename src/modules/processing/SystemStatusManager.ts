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
  private maxInitializationAttempts = 5 // Increased from 3
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
   * Set initialization error with detailed information
   */
  setInitializationError(error: string): void {
    this.status.error = `Initialization failed (attempt ${this.initializationAttempts}/${this.maxInitializationAttempts}): ${error}`
    this.status.initialized = false
    this.notifyStatusUpdate()
    
    console.error(`❌ System initialization error (attempt ${this.initializationAttempts}):`, error)
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
    this.resetInitializationAttempts() // Reset on successful initialization
    this.notifyStatusUpdate()
    
    console.log('✅ System marked as initialized successfully')
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
    this.initializationAttempts++
    console.log(`🔄 Initialization attempt ${this.initializationAttempts}/${this.maxInitializationAttempts}`)
    return this.initializationAttempts
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
   * Get detailed status for debugging
   */
  getDetailedStatus(): SystemStatus & {
    initializationAttempts: number
    maxAttempts: number
    hasCallbacks: boolean
  } {
    return {
      ...this.getStatus(),
      initializationAttempts: this.initializationAttempts,
      maxAttempts: this.maxInitializationAttempts,
      hasCallbacks: this.statusUpdateCallbacks.size > 0
    }
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
    
    console.log('🔄 System status reset')
  }

  /**
   * Force initialize with fallback worker
   */
  forceInitializeWithFallback(): void {
    console.warn('⚠️ Forcing initialization with fallback worker only')
    
    this.status = {
      initialized: true,
      totalWorkers: 1,
      availableWorkers: 1,
      queuedTasks: 0,
      environment: 'Fallback Mode',
      error: null
    }
    
    this.resetInitializationAttempts()
    this.notifyStatusUpdate()
    
    console.log('✅ System initialized in fallback mode')
  }
}