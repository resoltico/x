import type { ImageData, CanvasState } from '@/types'
import { renderImageToCanvas, calculateFitZoom, getImageCenter } from '@/utils/imageHelpers'

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

  /**
   * Initialize the renderer with a canvas element
   */
  initialize(canvas: HTMLCanvasElement) {
    this.canvas = canvas
    this.ctx = canvas.getContext('2d')
    
    if (!this.ctx) {
      throw new Error('Failed to get 2D context from canvas')
    }

    this.setupEventListeners()
    this.resize()
  }

  /**
   * Set up mouse and touch event listeners for pan and zoom
   */
  private setupEventListeners() {
    if (!this.canvas) return

    // Mouse events for panning
    this.canvas.addEventListener('mousedown', this.handleMouseDown.bind(this))
    this.canvas.addEventListener('mousemove', this.handleMouseMove.bind(this))
    this.canvas.addEventListener('mouseup', this.handleMouseUp.bind(this))
    this.canvas.addEventListener('mouseleave', this.handleMouseUp.bind(this))

    // Wheel event for zooming
    this.canvas.addEventListener('wheel', this.handleWheel.bind(this))

    // Touch events for mobile support
    this.canvas.addEventListener('touchstart', this.handleTouchStart.bind(this))
    this.canvas.addEventListener('touchmove', this.handleTouchMove.bind(this))
    this.canvas.addEventListener('touchend', this.handleTouchEnd.bind(this))

    // Window resize
    window.addEventListener('resize', this.resize.bind(this))
  }

  /**
   * Handle mouse down for drag start
   */
  private handleMouseDown(event: MouseEvent) {
    if (event.button === 0) { // Left mouse button
      this.isDragging = true
      this.lastMousePos = { x: event.clientX, y: event.clientY }
      this.canvas!.style.cursor = 'grabbing'
    }
  }

  /**
   * Handle mouse move for dragging
   */
  private handleMouseMove(event: MouseEvent) {
    if (this.isDragging) {
      const deltaX = event.clientX - this.lastMousePos.x
      const deltaY = event.clientY - this.lastMousePos.y
      
      this.state.offsetX += deltaX
      this.state.offsetY += deltaY
      
      this.lastMousePos = { x: event.clientX, y: event.clientY }
      this.render()
    }
  }

  /**
   * Handle mouse up for drag end
   */
  private handleMouseUp() {
    this.isDragging = false
    this.canvas!.style.cursor = 'grab'
  }

  /**
   * Handle wheel event for zooming
   */
  private handleWheel(event: WheelEvent) {
    event.preventDefault()
    
    const rect = this.canvas!.getBoundingClientRect()
    const mouseX = event.clientX - rect.left
    const mouseY = event.clientY - rect.top
    
    const zoomFactor = event.deltaY > 0 ? 0.9 : 1.1
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
  private handleTouchStart(event: TouchEvent) {
    event.preventDefault()
    if (event.touches.length === 1) {
      const touch = event.touches[0]
      this.isDragging = true
      this.lastMousePos = { x: touch.clientX, y: touch.clientY }
    }
  }

  private handleTouchMove(event: TouchEvent) {
    event.preventDefault()
    if (this.isDragging && event.touches.length === 1) {
      const touch = event.touches[0]
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

    const container = this.canvas.parentElement
    if (container) {
      this.canvas.width = container.clientWidth
      this.canvas.height = container.clientHeight
      this.render()
    }
  }

  /**
   * Set the current original image
   */
  setCurrentImage(imageData: ImageData) {
    this.currentImage = imageData
    this.fitToCanvas()
    this.render()
  }

  /**
   * Set the processed image
   */
  setProcessedImage(imageData: ImageData) {
    this.processedImage = imageData
    this.render()
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

    const zoom = calculateFitZoom(this.currentImage, this.canvas)
    const center = getImageCenter(this.currentImage, this.canvas)
    
    this.state.zoom = zoom
    this.state.offsetX = center.x - (this.currentImage.width * zoom) / 2
    this.state.offsetY = center.y - (this.currentImage.height * zoom) / 2
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

    const center = getImageCenter(this.currentImage, this.canvas)
    this.state.zoom = 1
    this.state.offsetX = center.x - this.currentImage.width / 2
    this.state.offsetY = center.y - this.currentImage.height / 2
    this.render()
  }

  /**
   * Main render method
   */
  render() {
    if (!this.canvas || !this.ctx) return

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
  }

  /**
   * Draw checkerboard background pattern
   */
  private drawBackground() {
    if (!this.ctx || !this.canvas) return

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
  }

  /**
   * Draw the image with current transformations
   */
  private drawImage(imageData: ImageData) {
    if (!this.ctx) return

    const img = new Image()
    img.onload = () => {
      const width = imageData.width * this.state.zoom
      const height = imageData.height * this.state.zoom
      
      this.ctx!.imageSmoothingEnabled = this.state.zoom < 1
      this.ctx!.drawImage(
        img,
        this.state.offsetX,
        this.state.offsetY,
        width,
        height
      )
    }

    // Convert ArrayBuffer to object URL
    const blob = new Blob([imageData.data])
    img.src = URL.createObjectURL(blob)
  }

  /**
   * Draw overlay information
   */
  private drawOverlay() {
    if (!this.ctx || !this.currentImage) return

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
  }

  /**
   * Export current view as image
   */
  exportView(): string {
    if (!this.canvas) return ''
    return this.canvas.toDataURL('image/png')
  }

  /**
   * Cleanup resources
   */
  destroy() {
    if (this.canvas) {
      // Remove event listeners
      this.canvas.removeEventListener('mousedown', this.handleMouseDown)
      this.canvas.removeEventListener('mousemove', this.handleMouseMove)
      this.canvas.removeEventListener('mouseup', this.handleMouseUp)
      this.canvas.removeEventListener('wheel', this.handleWheel)
      window.removeEventListener('resize', this.resize)
    }
    
    this.canvas = null
    this.ctx = null
    this.currentImage = null
    this.processedImage = null
  }
}