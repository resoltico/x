import type { MorphologyParams, ImageData } from '@/types'

/**
 * Morphological operations for binary image processing
 */

export class MorphologyProcessor {
  /**
   * Process image with specified morphological operation
   */
  static async process(
    imageData: ImageData,
    params: MorphologyParams,
    vipsImage?: any
  ): Promise<any> {
    const { operation, kernelSize, iterations } = params

    if (vipsImage) {
      return this.processWithVips(vipsImage, operation, kernelSize, iterations)
    } else {
      return this.processWithCanvas(imageData, operation, kernelSize, iterations)
    }
  }

  /**
   * Process using wasm-vips
   */
  private static async processWithVips(
    image: any,
    operation: string,
    kernelSize: number,
    iterations: number
  ): Promise<any> {
    // Create morphological kernel
    const kernel = image.constructor.newFromArray(
      this.createMorphologyKernel(kernelSize)
    )

    let result = image
    for (let i = 0; i < iterations; i++) {
      switch (operation) {
        case 'erosion':
          result = result.erode(kernel)
          break
        case 'dilation':
          result = result.dilate(kernel)
          break
        case 'opening':
          result = result.erode(kernel).dilate(kernel)
          break
        case 'closing':
          result = result.dilate(kernel).erode(kernel)
          break
        default:
          throw new Error(`Unknown morphological operation: ${operation}`)
      }
    }

    return result
  }

  /**
   * Process using Canvas API (fallback)
   */
  private static async processWithCanvas(
    imageData: ImageData,
    operation: string,
    kernelSize: number,
    iterations: number
  ): Promise<ArrayBuffer> {
    const canvas = new OffscreenCanvas(imageData.width, imageData.height)
    const ctx = canvas.getContext('2d')!
    
    // Create ImageData object for canvas
    const canvasImageData = new ImageData(
      new Uint8ClampedArray(imageData.data),
      imageData.width,
      imageData.height
    )
    
    ctx.putImageData(canvasImageData, 0, 0)

    // Apply morphological operations
    for (let i = 0; i < iterations; i++) {
      switch (operation) {
        case 'erosion':
          await this.canvasErosion(ctx, imageData, kernelSize)
          break
        case 'dilation':
          await this.canvasDilation(ctx, imageData, kernelSize)
          break
        case 'opening':
          await this.canvasErosion(ctx, imageData, kernelSize)
          await this.canvasDilation(ctx, imageData, kernelSize)
          break
        case 'closing':
          await this.canvasDilation(ctx, imageData, kernelSize)
          await this.canvasErosion(ctx, imageData, kernelSize)
          break
        default:
          throw new Error(`Unknown morphological operation: ${operation}`)
      }
    }
    
    // Convert canvas to blob and then to ArrayBuffer
    const blob = await canvas.convertToBlob({ type: 'image/png' })
    return await blob.arrayBuffer()
  }

  /**
   * Canvas-based erosion operation
   */
  private static async canvasErosion(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    kernelSize: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    const newPixels = new Uint8ClampedArray(pixels)
    const halfKernel = Math.floor(kernelSize / 2)
    const kernel = this.createMorphologyKernel(kernelSize)

    for (let y = halfKernel; y < imageData.height - halfKernel; y++) {
      for (let x = halfKernel; x < imageData.width - halfKernel; x++) {
        const idx = (y * imageData.width + x) * 4
        
        let minValue = 255
        
        // Apply kernel
        for (let ky = 0; ky < kernelSize; ky++) {
          for (let kx = 0; kx < kernelSize; kx++) {
            if (kernel[ky][kx] === 1) {
              const py = y + ky - halfKernel
              const px = x + kx - halfKernel
              const pidx = (py * imageData.width + px) * 4
              
              if (px >= 0 && px < imageData.width && py >= 0 && py < imageData.height) {
                const gray = pixels[pidx] * 0.299 + pixels[pidx + 1] * 0.587 + pixels[pidx + 2] * 0.114
                minValue = Math.min(minValue, gray)
              }
            }
          }
        }
        
        newPixels[idx] = minValue
        newPixels[idx + 1] = minValue
        newPixels[idx + 2] = minValue
      }
    }
    
    const newData = new ImageData(newPixels, imageData.width, imageData.height)
    ctx.putImageData(newData, 0, 0)
  }

  /**
   * Canvas-based dilation operation
   */
  private static async canvasDilation(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    kernelSize: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    const newPixels = new Uint8ClampedArray(pixels)
    const halfKernel = Math.floor(kernelSize / 2)
    const kernel = this.createMorphologyKernel(kernelSize)

    for (let y = halfKernel; y < imageData.height - halfKernel; y++) {
      for (let x = halfKernel; x < imageData.width - halfKernel; x++) {
        const idx = (y * imageData.width + x) * 4
        
        let maxValue = 0
        
        // Apply kernel
        for (let ky = 0; ky < kernelSize; ky++) {
          for (let kx = 0; kx < kernelSize; kx++) {
            if (kernel[ky][kx] === 1) {
              const py = y + ky - halfKernel
              const px = x + kx - halfKernel
              const pidx = (py * imageData.width + px) * 4
              
              if (px >= 0 && px < imageData.width && py >= 0 && py < imageData.height) {
                const gray = pixels[pidx] * 0.299 + pixels[pidx + 1] * 0.587 + pixels[pidx + 2] * 0.114
                maxValue = Math.max(maxValue, gray)
              }
            }
          }
        }
        
        newPixels[idx] = maxValue
        newPixels[idx + 1] = maxValue
        newPixels[idx + 2] = maxValue
      }
    }
    
    const newData = new ImageData(newPixels, imageData.width, imageData.height)
    ctx.putImageData(newData, 0, 0)
  }

  /**
   * Create morphological structuring element (kernel)
   */
  private static createMorphologyKernel(size: number, shape: 'circle' | 'square' | 'cross' = 'circle'): number[][] {
    const kernel: number[][] = []
    const center = Math.floor(size / 2)
    
    for (let y = 0; y < size; y++) {
      kernel[y] = []
      for (let x = 0; x < size; x++) {
        switch (shape) {
          case 'circle': {
            // Create circular kernel
            const distance = Math.sqrt((x - center) ** 2 + (y - center) ** 2)
            kernel[y][x] = distance <= center ? 1 : 0
            break
          }
          case 'square': {
            // Create square kernel (all 1s)
            kernel[y][x] = 1
            break
          }
          case 'cross': {
            // Create cross-shaped kernel
            kernel[y][x] = (x === center || y === center) ? 1 : 0
            break
          }
          default:
            kernel[y][x] = 1
        }
      }
    }
    
    return kernel
  }

  /**
   * Get parameter constraints for morphological operations
   */
  static getParameterConstraints() {
    return {
      kernelSize: { min: 3, max: 15, step: 2, default: 3 },
      iterations: { min: 1, max: 5, default: 1 },
      operations: ['opening', 'closing', 'erosion', 'dilation']
    }
  }

  /**
   * Get recommended parameters for different use cases
   */
  static getRecommendedParameters(useCase: string) {
    const recommendations = {
      'noise-removal': {
        operation: 'opening',
        kernelSize: 3,
        iterations: 1
      },
      'gap-filling': {
        operation: 'closing',
        kernelSize: 3,
        iterations: 1
      },
      'edge-smoothing': {
        operation: 'opening',
        kernelSize: 3,
        iterations: 2
      },
      'text-cleanup': {
        operation: 'closing',
        kernelSize: 3,
        iterations: 1
      },
      'skeleton-extraction': {
        operation: 'erosion',
        kernelSize: 3,
        iterations: 3
      },
      'boundary-extraction': {
        operation: 'dilation',
        kernelSize: 3,
        iterations: 1
      }
    }

    return recommendations[useCase as keyof typeof recommendations] || recommendations['noise-removal']
  }

  /**
   * Validate morphological parameters
   */
  static validateParameters(params: MorphologyParams): { isValid: boolean; errors: string[] } {
    const errors: string[] = []
    const constraints = this.getParameterConstraints()

    // Validate kernel size
    if (params.kernelSize < constraints.kernelSize.min || params.kernelSize > constraints.kernelSize.max) {
      errors.push(`Kernel size must be between ${constraints.kernelSize.min} and ${constraints.kernelSize.max}`)
    }

    // Kernel size must be odd
    if (params.kernelSize % 2 === 0) {
      errors.push('Kernel size must be odd')
    }

    // Validate iterations
    if (params.iterations < constraints.iterations.min || params.iterations > constraints.iterations.max) {
      errors.push(`Iterations must be between ${constraints.iterations.min} and ${constraints.iterations.max}`)
    }

    // Validate operation
    if (!constraints.operations.includes(params.operation)) {
      errors.push(`Operation must be one of: ${constraints.operations.join(', ')}`)
    }

    return {
      isValid: errors.length === 0,
      errors
    }
  }

  /**
   * Get operation descriptions
   */
  static getOperationDescriptions() {
    return {
      erosion: 'Shrinks white regions and expands black regions. Useful for removing noise and separating connected objects.',
      dilation: 'Expands white regions and shrinks black regions. Useful for filling small gaps and connecting nearby objects.',
      opening: 'Erosion followed by dilation. Removes small noise while preserving the shape of larger objects.',
      closing: 'Dilation followed by erosion. Fills small gaps and holes while preserving the shape of objects.'
    }
  }

  /**
   * Get kernel shape options
   */
  static getKernelShapes() {
    return {
      circle: 'Circular structuring element - good for general morphological operations',
      square: 'Square structuring element - preserves rectangular features',
      cross: 'Cross-shaped structuring element - preserves linear features'
    }
  }
}