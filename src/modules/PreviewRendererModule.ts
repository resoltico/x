import type { ImageData, CanvasState } from '@/types'
import { calculateFitZoom, getImageCenter } from '@/utils/imageHelpers'

/**
 * PreviewRendererModule handles canvas rendering and interaction
 */
export class PreviewRendererModule {
  private canvas: HTMLCanvasElement | null = null
  private ctx: CanvasRenderingContext2D | null = null
  private currentImage: ImageData | null = null
  private processedImage: ImageData | null = null
  private state: CanvasState = {
    zoom: 1,
    offsetX: 0,
    offsetY: 0,
    showOriginal: true
  }

  private isDragging = false
  private lastMousePos = { x: 0, y: 0 }
  private eventListeners: { element: EventTarget; type: string; listener: (event: Event) => void }[] = []

  /**
   * Initialize the renderer with a canvas element
   */
  initialize(canvas: HTMLCanvasElement) {
    try {
      this.canvas = canvas
      this.ctx = canvas.getContext('2d')
      
      if (!this.ctx) {
        throw new Error('Failed to get 2D context from canvas')
      }

      this.setupEventListeners()
      this.resize()
      
      console.log('PreviewRendererModule initialized successfully')
    } catch (error) {
      console.error('Failed to initialize PreviewRendererModule:', error)
      throw error
    }
  }

  /**
   * Set up mouse and touch event listeners for pan and zoom
   */
  private setupEventListeners() {
    if (!this.canvas) return

    // Store event listeners for cleanup
    const addListener = (element: EventTarget, type: string, listener: (event: Event) => void, options?: boolean | { passive?: boolean; once?: boolean; capture?: boolean }) => {
      element.addEventListener(type, listener, options)
      this.eventListeners.push({ element, type, listener })
    }

    // Mouse events for panning
    addListener(this.canvas, 'mousedown', this.handleMouseDown.bind(this))
    addListener(this.canvas, 'mousemove', this.handleMouseMove.bind(this))
    addListener(this.canvas, 'mouseup', this.handleMouseUp.bind(this))
    addListener(this.canvas, 'mouseleave', this.handleMouseUp.bind(this))

    // Wheel event for zooming
    addListener(this.canvas, 'wheel', this.handleWheel.bind(this), { passive: false })

    // Touch events for mobile support
    addListener(this.canvas, 'touchstart', this.handleTouchStart.bind(this), { passive: false })
    addListener(this.canvas, 'touchmove', this.handleTouchMove.bind(this), { passive: false })
    addListener(this.canvas, 'touchend', this.handleTouchEnd.bind(this))

    // Window resize
    addListener(window, 'resize', this.resize.bind(this))
  }

  /**
   * Handle mouse down for drag start
   */
  private handleMouseDown(event: Event) {
    const mouseEvent = event as MouseEvent
    if (mouseEvent.button === 0) { // Left mouse button
      this.isDragging = true
      this.lastMousePos = { x: mouseEvent.clientX, y: mouseEvent.clientY }
      if (this.canvas) {
        this.canvas.style.cursor = 'grabbing'
      }
    }
  }

  /**
   * Handle mouse move for dragging
   */
  private handleMouseMove(event: Event) {
    const mouseEvent = event as MouseEvent
    if (this.isDragging) {
      const deltaX = mouseEvent.clientX - this.lastMousePos.x
      const deltaY = mouseEvent.clientY - this.lastMousePos.y
      
      this.state.offsetX += deltaX
      this.state.offsetY += deltaY
      
      this.lastMousePos = { x: mouseEvent.clientX, y: mouseEvent.clientY }
      this.render()
    }
  }

  /**
   * Handle mouse up for drag end
   */
  private handleMouseUp() {
    this.isDragging = false
    if (this.canvas) {
      this.canvas.style.cursor = 'grab'
    }
  }

  /**
   * Handle wheel event for zooming
   */
  private handleWheel(event: Event) {
    const wheelEvent = event as WheelEvent
    wheelEvent.preventDefault()
    
    if (!this.canvas) return
    
    const rect = this.canvas.getBoundingClientRect()
    const mouseX = wheelEvent.clientX - rect.left
    const mouseY = wheelEvent.clientY - rect.top
    
    const zoomFactor = wheelEvent.deltaY > 0 ? 0.9 : 1.1
    const newZoom = Math.max(0.1, Math.min(5, this.state.zoom * zoomFactor))
    
    // Zoom towards mouse position
    const zoomRatio = newZoom / this.state.zoom
    this.state.offsetX = mouseX - (mouseX - this.state.offsetX) * zoomRatio
    this.state.offsetY = mouseY - (mouseY - this.state.offsetY) * zoomRatio
    this.state.zoom = newZoom
    
    this.render()
  }

  /**
   * Handle touch events for mobile panning
   */
  private handleTouchStart(event: Event) {
    const touchEvent = event as TouchEvent
    touchEvent.preventDefault()
    if (touchEvent.touches.length === 1) {
      const touch = touchEvent.touches[0]
      this.isDragging = true
      this.lastMousePos = { x: touch.clientX, y: touch.clientY }
    }
  }

  private handleTouchMove(event: Event) {
    const touchEvent = event as TouchEvent
    touchEvent.preventDefault()
    if (this.isDragging && touchEvent.touches.length === 1) {
      const touch = touchEvent.touches[0]
      const deltaX = touch.clientX - this.lastMousePos.x
      const deltaY = touch.clientY - this.lastMousePos.y
      
      this.state.offsetX += deltaX
      this.state.offsetY += deltaY
      
      this.lastMousePos = { x: touch.clientX, y: touch.clientY }
      this.render()
    }
  }

  private handleTouchEnd() {
    this.isDragging = false
  }

  /**
   * Resize canvas to fit container
   */
  resize() {
    if (!this.canvas) return

    try {
      const container = this.canvas.parentElement
      if (container) {
        const rect = container.getBoundingClientRect()
        this.canvas.width = rect.width || 800
        this.canvas.height = rect.height || 600
        this.render()
      }
    } catch (error) {
      console.warn('Failed to resize canvas:', error)
    }
  }

  /**
   * Set the current original image
   */
  setCurrentImage(imageData: ImageData) {
    try {
      this.currentImage = imageData
      this.fitToCanvas()
      this.render()
    } catch (error) {
      console.error('Failed to set current image:', error)
    }
  }

  /**
   * Set the processed image
   */
  setProcessedImage(imageData: ImageData) {
    try {
      this.processedImage = imageData
      this.render()
    } catch (error) {
      console.error('Failed to set processed image:', error)
    }
  }

  /**
   * Toggle between original and processed image view
   */
  toggleImageView() {
    this.state.showOriginal = !this.state.showOriginal
    this.render()
  }

  /**
   * Update canvas state
   */
  updateState(newState: Partial<CanvasState>) {
    this.state = { ...this.state, ...newState }
    this.render()
  }

  /**
   * Get current canvas state
   */
  getState(): CanvasState {
    return { ...this.state }
  }

  /**
   * Fit image to canvas view
   */
  fitToCanvas() {
    if (!this.canvas || !this.currentImage) return

    try {
      const zoom = calculateFitZoom(this.currentImage, this.canvas)
      const center = getImageCenter(this.currentImage, this.canvas)
      
      this.state.zoom = zoom
      this.state.offsetX = center.x - (this.currentImage.width * zoom) / 2
      this.state.offsetY = center.y - (this.currentImage.height * zoom) / 2
    } catch (error) {
      console.warn('Failed to fit image to canvas:', error)
    }
  }

  /**
   * Reset view to original position and zoom
   */
  resetView() {
    this.state.zoom = 1
    this.state.offsetX = 0
    this.state.offsetY = 0
    this.render()
  }

  /**
   * Zoom to actual size (100%)
   */
  zoomToActualSize() {
    if (!this.canvas || !this.currentImage) return

    try {
      const center = getImageCenter(this.currentImage, this.canvas)
      this.state.zoom = 1
      this.state.offsetX = center.x - this.currentImage.width / 2
      this.state.offsetY = center.y - this.currentImage.height / 2
      this.render()
    } catch (error) {
      console.warn('Failed to zoom to actual size:', error)
    }
  }

  /**
   * Main render method
   */
  render() {
    if (!this.canvas || !this.ctx) return

    try {
      // Clear canvas
      this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height)

      // Draw background pattern
      this.drawBackground()

      // Determine which image to show
      const imageToShow = this.state.showOriginal ? this.currentImage : 
                        (this.processedImage || this.currentImage)

      if (imageToShow) {
        this.drawImage(imageToShow)
      }

      // Draw overlay info
      this.drawOverlay()
    } catch (error) {
      console.warn('Failed to render canvas:', error)
    }
  }

  /**
   * Draw checkerboard background pattern
   */
  private drawBackground() {
    if (!this.ctx || !this.canvas) return

    try {
      const size = 20
      this.ctx.fillStyle = '#f0f0f0'
      this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height)

      this.ctx.fillStyle = '#e0e0e0'
      for (let x = 0; x < this.canvas.width; x += size) {
        for (let y = 0; y < this.canvas.height; y += size) {
          if ((x / size + y / size) % 2 === 0) {
            this.ctx.fillRect(x, y, size, size)
          }
        }
      }
    } catch (error) {
      console.warn('Failed to draw background:', error)
    }
  }

  /**
   * Draw the image with current transformations
   */
  private drawImage(imageData: ImageData) {
    if (!this.ctx) return

    try {
      // Check if we're in a browser environment
      if (typeof Image === 'undefined' || typeof URL === 'undefined' || !URL.createObjectURL) {
        console.warn('Image rendering not supported in this environment')
        return
      }

      const img = new Image()
      img.onload = () => {
        try {
          if (!this.ctx) return
          
          const width = imageData.width * this.state.zoom
          const height = imageData.height * this.state.zoom
          
          this.ctx.imageSmoothingEnabled = this.state.zoom < 1
          this.ctx.drawImage(
            img,
            this.state.offsetX,
            this.state.offsetY,
            width,
            height
          )
          
          // Clean up object URL
          URL.revokeObjectURL(img.src)
        } catch (error) {
          console.warn('Failed to draw image:', error)
        }
      }

      img.onerror = () => {
        console.warn('Failed to load image for rendering')
      }

      // Convert ArrayBuffer to object URL
      const blob = new Blob([imageData.data])
      img.src = URL.createObjectURL(blob)
    } catch (error) {
      console.warn('Failed to setup image drawing:', error)
    }
  }

  /**
   * Draw overlay information
   */
  private drawOverlay() {
    if (!this.ctx || !this.currentImage) return

    try {
      const padding = 10
      const fontSize = 12
      this.ctx.font = `${fontSize}px Inter, sans-serif`
      this.ctx.fillStyle = 'rgba(0, 0, 0, 0.7)'
      this.ctx.fillRect(padding, padding, 200, 60)

      this.ctx.fillStyle = 'white'
      this.ctx.fillText(`Zoom: ${(this.state.zoom * 100).toFixed(0)}%`, padding + 10, padding + 20)
      this.ctx.fillText(`Size: ${this.currentImage.width}×${this.currentImage.height}`, padding + 10, padding + 35)
      this.ctx.fillText(
        `View: ${this.state.showOriginal ? 'Original' : 'Processed'}`,
        padding + 10,
        padding + 50
      )
    } catch (error) {
      console.warn('Failed to draw overlay:', error)
    }
  }

  /**
   * Export current view as image
   */
  exportView(): string {
    if (!this.canvas) return ''
    
    try {
      return this.canvas.toDataURL('image/png')
    } catch (error) {
      console.warn('Failed to export view:', error)
      return ''
    }
  }

  /**
   * Cleanup resources
   */
  destroy() {
    try {
      // Remove all event listeners
      this.eventListeners.forEach(({ element, type, listener }) => {
        element.removeEventListener(type, listener)
      })
      this.eventListeners = []
      
      this.canvas = null
      this.ctx = null
      this.currentImage = null
      this.processedImage = null
      
      console.log('PreviewRendererModule destroyed')
    } catch (error) {
      console.warn('Error during PreviewRendererModule cleanup:', error)
    }
  }
}