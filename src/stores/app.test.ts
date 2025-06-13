import { describe, it, expect, beforeEach } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'
import { useAppStore } from '@/stores/app'
import type { ImageData } from '@/types'

describe('App Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('initializes with correct default state', () => {
    const store = useAppStore()
    
    expect(store.currentImage).toBe(null)
    expect(store.processedImage).toBe(null)
    expect(store.activeTasks).toEqual([])
    expect(store.isProcessing).toBe(false)
    expect(store.hasImage).toBe(false)
    expect(store.hasProcessedImage).toBe(false)
  })

  it('sets current image correctly', () => {
    const store = useAppStore()
    const mockImage: ImageData = {
      data: new ArrayBuffer(100),
      width: 100,
      height: 100,
      channels: 4,
      format: 'PNG',
      filename: 'test.png',
      size: 1024
    }

    store.setCurrentImage(mockImage)

    // Use toStrictEqual for deep comparison instead of toBe for object reference
    expect(store.currentImage).toStrictEqual(mockImage)
    expect(store.hasImage).toBe(true)
    expect(store.processedImage).toBe(null) // Should reset processed image
  })

  it('adds and updates tasks correctly', () => {
    const store = useAppStore()
    
    const task = store.addTask('binarization', {
      binarization: { method: 'otsu', threshold: 128 }
    })

    expect(store.activeTasks).toHaveLength(1)
    expect(task.type).toBe('binarization')
    expect(task.status).toBe('pending')

    store.updateTask(task.id, { status: 'processing', progress: 50 })
    
    const updatedTask = store.activeTasks.find(t => t.id === task.id)
    expect(updatedTask?.status).toBe('processing')
    expect(updatedTask?.progress).toBe(50)
  })

  it('updates canvas state correctly', () => {
    const store = useAppStore()
    
    store.updateCanvasState({ zoom: 2, offsetX: 100, offsetY: 50 })
    
    expect(store.canvasState.zoom).toBe(2)
    expect(store.canvasState.offsetX).toBe(100)
    expect(store.canvasState.offsetY).toBe(50)
    expect(store.canvasState.showOriginal).toBe(true) // Should maintain other properties
  })

  it('clears completed tasks', () => {
    const store = useAppStore()
    
    const task1 = store.addTask('binarization', { binarization: { method: 'otsu' } })
    const task2 = store.addTask('scaling', { scaling: { method: 'scale2x', factor: 2 } })
    
    store.updateTask(task1.id, { status: 'completed' })
    store.updateTask(task2.id, { status: 'processing' })
    
    expect(store.activeTasks).toHaveLength(2)
    
    store.clearCompletedTasks()
    
    expect(store.activeTasks).toHaveLength(1)
    expect(store.activeTasks[0].id).toBe(task2.id)
  })

  it('resets state correctly', () => {
    const store = useAppStore()
    const mockImage: ImageData = {
      data: new ArrayBuffer(100),
      width: 100,
      height: 100,
      channels: 4,
      format: 'PNG',
      size: 1024
    }

    store.setCurrentImage(mockImage)
    store.addTask('binarization', { binarization: { method: 'otsu' } })
    store.updateCanvasState({ zoom: 2 })

    store.reset()

    expect(store.currentImage).toBe(null)
    expect(store.processedImage).toBe(null)
    expect(store.activeTasks).toEqual([])
    expect(store.canvasState.zoom).toBe(1)
    expect(store.canvasState.offsetX).toBe(0)
    expect(store.canvasState.offsetY).toBe(0)
  })

  it('handles processed image setting', () => {
    const store = useAppStore()
    const mockProcessedImage: ImageData = {
      data: new ArrayBuffer(200),
      width: 200,
      height: 200,
      channels: 4,
      format: 'PNG',
      filename: 'processed.png',
      size: 2048
    }

    store.setProcessedImage(mockProcessedImage)

    expect(store.processedImage).toStrictEqual(mockProcessedImage)
    expect(store.hasProcessedImage).toBe(true)
  })

  it('handles task cancellation', () => {
    const store = useAppStore()
    
    const task = store.addTask('binarization', { binarization: { method: 'otsu' } })
    expect(store.activeTasks).toHaveLength(1)

    store.cancelTask(task.id)
    
    const cancelledTask = store.activeTasks.find(t => t.id === task.id)
    expect(cancelledTask?.status).toBe('cancelled')
  })

  it('handles task removal', () => {
    const store = useAppStore()
    
    const task = store.addTask('binarization', { binarization: { method: 'otsu' } })
    expect(store.activeTasks).toHaveLength(1)

    store.removeTask(task.id)
    
    expect(store.activeTasks).toHaveLength(0)
  })

  it('computes processing state correctly', () => {
    const store = useAppStore()
    
    expect(store.isProcessing).toBe(false)
    
    const task = store.addTask('binarization', { binarization: { method: 'otsu' } })
    expect(store.isProcessing).toBe(false) // Still pending
    
    store.updateTask(task.id, { status: 'processing', progress: 50 })
    expect(store.isProcessing).toBe(true)
    
    store.updateTask(task.id, { status: 'completed', progress: 100 })
    expect(store.isProcessing).toBe(false)
  })

  it('computes current task correctly', () => {
    const store = useAppStore()
    
    expect(store.currentTask).toBeUndefined()
    
    const task1 = store.addTask('binarization', { binarization: { method: 'otsu' } })
    const task2 = store.addTask('scaling', { scaling: { method: 'scale2x', factor: 2 } })
    
    expect(store.currentTask).toBeUndefined() // None processing yet
    
    store.updateTask(task1.id, { status: 'processing', progress: 50 })
    expect(store.currentTask?.id).toBe(task1.id)
    
    // Second task starts processing
    store.updateTask(task2.id, { status: 'processing', progress: 25 })
    // Should still return first processing task found
    expect(store.currentTask?.id).toBe(task1.id)
  })

  it('computes completed tasks correctly', () => {
    const store = useAppStore()
    
    const task1 = store.addTask('binarization', { binarization: { method: 'otsu' } })
    const task2 = store.addTask('scaling', { scaling: { method: 'scale2x', factor: 2 } })
    
    expect(store.completedTasks).toHaveLength(0)
    
    store.updateTask(task1.id, { status: 'completed', progress: 100 })
    expect(store.completedTasks).toHaveLength(1)
    expect(store.completedTasks[0].id).toBe(task1.id)
    
    store.updateTask(task2.id, { status: 'failed', progress: 0 })
    expect(store.completedTasks).toHaveLength(1) // Failed tasks not included in completed
  })

  it('handles canvas reset', () => {
    const store = useAppStore()
    
    store.updateCanvasState({ zoom: 3, offsetX: 200, offsetY: 100, showOriginal: false })
    
    store.resetCanvas()
    
    expect(store.canvasState).toEqual({
      zoom: 1,
      offsetX: 0,
      offsetY: 0,
      showOriginal: true
    })
  })
})