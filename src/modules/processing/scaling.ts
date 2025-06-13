import type { ScalingParams, ImageData } from '@/types'

/**
 * Image scaling algorithms including pixel art scaling
 */

export class ScalingProcessor {
  /**
   * Process image with specified scaling method
   */
  static async process(
    imageData: ImageData,
    params: ScalingParams,
    vipsImage?: any
  ): Promise<any> {
    const { method, factor } = params

    if (vipsImage) {
      return this.processWithVips(vipsImage, method, factor)
    } else {
      return this.processWithCanvas(imageData, method, factor)
    }
  }

  /**
   * Process using wasm-vips
   */
  private static async processWithVips(
    image: any,
    method: string,
    factor: number
  ): Promise<any> {
    switch (method) {
      case 'nearest':
        return image.resize(factor, { kernel: 'nearest' })
      case 'bilinear':
        return image.resize(factor, { kernel: 'linear' })
      case 'bicubic':
        return image.resize(factor, { kernel: 'cubic' })
      case 'lanczos':
        return image.resize(factor, { kernel: 'lanczos3' })
      case 'scale2x':
      case 'scale3x':
      case 'scale4x':
        // Use custom pixel art scaling algorithms
        return this.pixelArtScaling(image, method)
      default:
        return image.resize(factor, { kernel: 'linear' })
    }
  }

  /**
   * Process using Canvas API (fallback)
   */
  private static async processWithCanvas(
    imageData: ImageData,
    method: string,
    factor: number
  ): Promise<ArrayBuffer> {
    const canvas = new OffscreenCanvas(imageData.width, imageData.height)
    const ctx = canvas.getContext('2d')!
    
    // Create ImageData object for canvas
    const canvasImageData = new globalThis.ImageData(
      new Uint8ClampedArray(imageData.data),
      imageData.width,
      imageData.height
    )
    
    ctx.putImageData(canvasImageData, 0, 0)

    switch (method) {
      case 'nearest':
        return this.canvasNearestNeighbor(canvas, factor)
      case 'bilinear':
        return this.canvasBilinear(canvas, factor)
      case 'scale2x':
        return this.canvasScale2x(canvas)
      case 'scale3x':
        return this.canvasScale3x(canvas)
      case 'scale4x':
        return this.canvasScale4x(canvas)
      default:
        return this.canvasBilinear(canvas, factor)
    }
  }

  /**
   * Pixel art scaling using VIPS (simplified)
   */
  private static pixelArtScaling(image: any, method: string): any {
    // For now, use nearest neighbor scaling with the appropriate factor
    // In a full implementation, these would use the actual Scale2x/3x/4x algorithms
    const factor = parseInt(method.replace('scale', '').replace('x', ''))
    return image.resize(factor, { kernel: 'nearest' })
  }

  /**
   * Canvas nearest neighbor scaling
   */
  private static async canvasNearestNeighbor(
    canvas: OffscreenCanvas,
    factor: number
  ): Promise<ArrayBuffer> {
    const scaledCanvas = new OffscreenCanvas(
      Math.round(canvas.width * factor),
      Math.round(canvas.height * factor)
    )
    const scaledCtx = scaledCanvas.getContext('2d')!
    
    scaledCtx.imageSmoothingEnabled = false
    scaledCtx.drawImage(
      canvas,
      0, 0, canvas.width, canvas.height,
      0, 0, scaledCanvas.width, scaledCanvas.height
    )
    
    const blob = await scaledCanvas.convertToBlob({ type: 'image/png' })
    return await blob.arrayBuffer()
  }

  /**
   * Canvas bilinear scaling
   */
  private static async canvasBilinear(
    canvas: OffscreenCanvas,
    factor: number
  ): Promise<ArrayBuffer> {
    const scaledCanvas = new OffscreenCanvas(
      Math.round(canvas.width * factor),
      Math.round(canvas.height * factor)
    )
    const scaledCtx = scaledCanvas.getContext('2d')!
    
    scaledCtx.imageSmoothingEnabled = true
    scaledCtx.imageSmoothingQuality = 'high'
    scaledCtx.drawImage(
      canvas,
      0, 0, canvas.width, canvas.height,
      0, 0, scaledCanvas.width, scaledCanvas.height
    )
    
    const blob = await scaledCanvas.convertToBlob({ type: 'image/png' })
    return await blob.arrayBuffer()
  }

  /**
   * Scale2x algorithm implementation
   */
  private static async canvasScale2x(canvas: OffscreenCanvas): Promise<ArrayBuffer> {
    const ctx = canvas.getContext('2d')!
    const srcData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const srcPixels = srcData.data
    
    const scaledCanvas = new OffscreenCanvas(canvas.width * 2, canvas.height * 2)
    const scaledCtx = scaledCanvas.getContext('2d')!
    const dstData = scaledCtx.createImageData(canvas.width * 2, canvas.height * 2)
    const dstPixels = dstData.data
    
    for (let y = 0; y < canvas.height; y++) {
      for (let x = 0; x < canvas.width; x++) {
        const srcIdx = (y * canvas.width + x) * 4
        
        // Get neighboring pixels
        const A = this.getPixel(srcPixels, x, y - 1, canvas.width, canvas.height)
        const B = this.getPixel(srcPixels, x + 1, y, canvas.width, canvas.height)
        const C = this.getPixel(srcPixels, x, y, canvas.width, canvas.height) // Center pixel
        const D = this.getPixel(srcPixels, x - 1, y, canvas.width, canvas.height)
        const E = this.getPixel(srcPixels, x, y + 1, canvas.width, canvas.height)
        
        // Scale2x algorithm
        let E0 = C, E1 = C, E2 = C, E3 = C
        
        if (!this.pixelsEqual(D, B) && !this.pixelsEqual(A, E)) {
          E0 = this.pixelsEqual(D, A) ? D : C
          E1 = this.pixelsEqual(A, B) ? B : C
          E2 = this.pixelsEqual(D, E) ? D : C
          E3 = this.pixelsEqual(E, B) ? B : C
        }
        
        // Write to destination
        this.setPixel(dstPixels, x * 2, y * 2, E0, scaledCanvas.width)
        this.setPixel(dstPixels, x * 2 + 1, y * 2, E1, scaledCanvas.width)
        this.setPixel(dstPixels, x * 2, y * 2 + 1, E2, scaledCanvas.width)
        this.setPixel(dstPixels, x * 2 + 1, y * 2 + 1, E3, scaledCanvas.width)
      }
    }
    
    scaledCtx.putImageData(dstData, 0, 0)
    
    const blob = await scaledCanvas.convertToBlob({ type: 'image/png' })
    return await blob.arrayBuffer()
  }

  /**
   * Scale3x algorithm implementation
   */
  private static async canvasScale3x(canvas: OffscreenCanvas): Promise<ArrayBuffer> {
    const ctx = canvas.getContext('2d')!
    const srcData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const srcPixels = srcData.data
    
    const scaledCanvas = new OffscreenCanvas(canvas.width * 3, canvas.height * 3)
    const scaledCtx = scaledCanvas.getContext('2d')!
    const dstData = scaledCtx.createImageData(canvas.width * 3, canvas.height * 3)
    const dstPixels = dstData.data
    
    for (let y = 0; y < canvas.height; y++) {
      for (let x = 0; x < canvas.width; x++) {
        // Get 3x3 neighborhood
        const A = this.getPixel(srcPixels, x - 1, y - 1, canvas.width, canvas.height)
        const B = this.getPixel(srcPixels, x, y - 1, canvas.width, canvas.height)
        const C = this.getPixel(srcPixels, x + 1, y - 1, canvas.width, canvas.height)
        const D = this.getPixel(srcPixels, x - 1, y, canvas.width, canvas.height)
        const E = this.getPixel(srcPixels, x, y, canvas.width, canvas.height) // Center
        const F = this.getPixel(srcPixels, x + 1, y, canvas.width, canvas.height)
        const G = this.getPixel(srcPixels, x - 1, y + 1, canvas.width, canvas.height)
        const H = this.getPixel(srcPixels, x, y + 1, canvas.width, canvas.height)
        const I = this.getPixel(srcPixels, x + 1, y + 1, canvas.width, canvas.height)
        
        // Scale3x algorithm - simplified version
        let E0 = E, E1 = E, E2 = E
        let E3 = E, E4 = E, E5 = E
        let E6 = E, E7 = E, E8 = E
        
        if (!this.pixelsEqual(D, F) && !this.pixelsEqual(B, H)) {
          E0 = this.pixelsEqual(D, B) ? D : E
          E1 = (this.pixelsEqual(D, B) && !this.pixelsEqual(E, C)) || 
               (this.pixelsEqual(B, F) && !this.pixelsEqual(E, A)) ? B : E
          E2 = this.pixelsEqual(B, F) ? F : E
          E3 = (this.pixelsEqual(D, B) && !this.pixelsEqual(E, G)) || 
               (this.pixelsEqual(D, H) && !this.pixelsEqual(E, A)) ? D : E
          E4 = E
          E5 = (this.pixelsEqual(B, F) && !this.pixelsEqual(E, I)) || 
               (this.pixelsEqual(F, H) && !this.pixelsEqual(E, C)) ? F : E
          E6 = this.pixelsEqual(D, H) ? D : E
          E7 = (this.pixelsEqual(D, H) && !this.pixelsEqual(E, I)) || 
               (this.pixelsEqual(H, F) && !this.pixelsEqual(E, G)) ? H : E
          E8 = this.pixelsEqual(H, F) ? F : E
        }
        
        // Write 3x3 block to destination
        this.setPixel(dstPixels, x * 3, y * 3, E0, scaledCanvas.width)
        this.setPixel(dstPixels, x * 3 + 1, y * 3, E1, scaledCanvas.width)
        this.setPixel(dstPixels, x * 3 + 2, y * 3, E2, scaledCanvas.width)
        this.setPixel(dstPixels, x * 3, y * 3 + 1, E3, scaledCanvas.width)
        this.setPixel(dstPixels, x * 3 + 1, y * 3 + 1, E4, scaledCanvas.width)
        this.setPixel(dstPixels, x * 3 + 2, y * 3 + 1, E5, scaledCanvas.width)
        this.setPixel(dstPixels, x * 3, y * 3 + 2, E6, scaledCanvas.width)
        this.setPixel(dstPixels, x * 3 + 1, y * 3 + 2, E7, scaledCanvas.width)
        this.setPixel(dstPixels, x * 3 + 2, y * 3 + 2, E8, scaledCanvas.width)
      }
    }
    
    scaledCtx.putImageData(dstData, 0, 0)
    
    const blob = await scaledCanvas.convertToBlob({ type: 'image/png' })
    return await blob.arrayBuffer()
  }

  /**
   * Scale4x algorithm implementation (Scale2x applied twice)
   */
  private static async canvasScale4x(canvas: OffscreenCanvas): Promise<ArrayBuffer> {
    // Apply Scale2x twice
    const intermediate = await this.canvasScale2x(canvas)
    
    // Create intermediate canvas from the result
    const blob = new Blob([intermediate], { type: 'image/png' })
    
    // Use createImageBitmap instead of Image constructor for worker compatibility
    const imageBitmap = await createImageBitmap(blob)
    
    const intermediateCanvas = new OffscreenCanvas(canvas.width * 2, canvas.height * 2)
    const intermediateCtx = intermediateCanvas.getContext('2d')!
    
    intermediateCtx.drawImage(imageBitmap, 0, 0)
    
    // Apply Scale2x again
    const result = await this.canvasScale2x(intermediateCanvas)
    
    // Clean up the image bitmap
    imageBitmap.close()
    
    return result
  }

  /**
   * Get pixel from image data
   */
  private static getPixel(
    pixels: Uint8ClampedArray,
    x: number,
    y: number,
    width: number,
    height: number
  ): [number, number, number, number] {
    // Clamp coordinates to image bounds
    x = Math.max(0, Math.min(width - 1, x))
    y = Math.max(0, Math.min(height - 1, y))
    
    const idx = (y * width + x) * 4
    return [pixels[idx], pixels[idx + 1], pixels[idx + 2], pixels[idx + 3]]
  }

  /**
   * Set pixel in image data
   */
  private static setPixel(
    pixels: Uint8ClampedArray,
    x: number,
    y: number,
    color: [number, number, number, number],
    width: number
  ) {
    const idx = (y * width + x) * 4
    pixels[idx] = color[0]
    pixels[idx + 1] = color[1]
    pixels[idx + 2] = color[2]
    pixels[idx + 3] = color[3]
  }

  /**
   * Compare two pixels for equality
   */
  private static pixelsEqual(
    a: [number, number, number, number],
    b: [number, number, number, number],
    threshold: number = 0
  ): boolean {
    return Math.abs(a[0] - b[0]) <= threshold &&
           Math.abs(a[1] - b[1]) <= threshold &&
           Math.abs(a[2] - b[2]) <= threshold &&
           Math.abs(a[3] - b[3]) <= threshold
  }

  /**
   * Get parameter constraints for scaling
   */
  static getParameterConstraints() {
    return {
      factor: { min: 0.1, max: 8.0, step: 0.1, default: 2.0 },
      methods: [
        'nearest',
        'bilinear', 
        'bicubic',
        'lanczos',
        'scale2x',
        'scale3x',
        'scale4x'
      ]
    }
  }

  /**
   * Get recommended parameters for different use cases
   */
  static getRecommendedParameters(useCase: string) {
    const recommendations = {
      'pixel-art': {
        method: 'scale2x',
        factor: 2
      },
      'pixel-art-large': {
        method: 'scale4x',
        factor: 4
      },
      'photograph': {
        method: 'lanczos',
        factor: 2.0
      },
      'line-art': {
        method: 'nearest',
        factor: 2.0
      },
      'smooth-upscale': {
        method: 'bicubic',
        factor: 2.0
      },
      'fast-upscale': {
        method: 'bilinear',
        factor: 2.0
      },
      'downscale': {
        method: 'lanczos',
        factor: 0.5
      }
    }

    return recommendations[useCase as keyof typeof recommendations] || recommendations['photograph']
  }

  /**
   * Validate scaling parameters
   */
  static validateParameters(params: ScalingParams): { isValid: boolean; errors: string[] } {
    const errors: string[] = []
    const constraints = this.getParameterConstraints()

    // Validate factor
    if (params.factor < constraints.factor.min || params.factor > constraints.factor.max) {
      errors.push(`Scale factor must be between ${constraints.factor.min} and ${constraints.factor.max}`)
    }

    // Validate method
    if (!constraints.methods.includes(params.method)) {
      errors.push(`Method must be one of: ${constraints.methods.join(', ')}`)
    }

    // Special validation for pixel art methods
    if (params.method.startsWith('scale') && !Number.isInteger(params.factor)) {
      errors.push(`Pixel art scaling methods require integer scale factors`)
    }

    // Warn about very large scale factors
    if (params.factor > 4) {
      errors.push(`Warning: Large scale factors may require significant memory and processing time`)
    }

    return {
      isValid: errors.filter(e => !e.startsWith('Warning:')).length === 0,
      errors
    }
  }

  /**
   * Get method descriptions
   */
  static getMethodDescriptions() {
    return {
      nearest: 'Nearest neighbor interpolation. Fast but may produce blocky results. Good for pixel art.',
      bilinear: 'Bilinear interpolation. Smooth results with some blur. Good general purpose method.',
      bicubic: 'Bicubic interpolation. Smoother than bilinear but may introduce ringing artifacts.',
      lanczos: 'Lanczos resampling. High quality results, best for photographic images.',
      scale2x: 'Scale2x algorithm. Designed specifically for pixel art, produces sharp 2x upscaling.',
      scale3x: 'Scale3x algorithm. Pixel art scaling to 3x size with edge detection.',
      scale4x: 'Scale4x algorithm. High quality 4x pixel art upscaling (Scale2x applied twice).'
    }
  }

  /**
   * Calculate output dimensions
   */
  static calculateOutputDimensions(
    inputWidth: number,
    inputHeight: number,
    params: ScalingParams
  ): { width: number; height: number } {
    let factor = params.factor

    // For pixel art methods, extract factor from method name
    if (params.method.startsWith('scale')) {
      factor = parseInt(params.method.replace('scale', '').replace('x', ''))
    }

    return {
      width: Math.round(inputWidth * factor),
      height: Math.round(inputHeight * factor)
    }
  }

  /**
   * Estimate memory usage for scaling operation
   */
  static estimateMemoryUsage(
    inputWidth: number,
    inputHeight: number,
    params: ScalingParams
  ): { inputMB: number; outputMB: number; totalMB: number } {
    const outputDims = this.calculateOutputDimensions(inputWidth, inputHeight, params)
    
    const inputPixels = inputWidth * inputHeight
    const outputPixels = outputDims.width * outputDims.height
    
    // Assume 4 bytes per pixel (RGBA)
    const inputMB = (inputPixels * 4) / (1024 * 1024)
    const outputMB = (outputPixels * 4) / (1024 * 1024)
    
    // Some algorithms need intermediate buffers
    const intermediateMultiplier = params.method === 'scale4x' ? 2 : 1
    const totalMB = inputMB + (outputMB * intermediateMultiplier)
    
    return { inputMB, outputMB, totalMB }
  }

  /**
   * Check if scaling operation is feasible
   */
  static checkFeasibility(
    inputWidth: number,
    inputHeight: number,
    params: ScalingParams,
    maxMemoryMB: number = 500
  ): { feasible: boolean; reason?: string; suggestion?: string } {
    const outputDims = this.calculateOutputDimensions(inputWidth, inputHeight, params)
    const memoryUsage = this.estimateMemoryUsage(inputWidth, inputHeight, params)
    
    // Check output dimensions
    const maxDimension = 16384
    if (outputDims.width > maxDimension || outputDims.height > maxDimension) {
      return {
        feasible: false,
        reason: `Output dimensions too large (${outputDims.width}x${outputDims.height}). Maximum: ${maxDimension}x${maxDimension}`,
        suggestion: 'Reduce scale factor or use a different scaling method'
      }
    }
    
    // Check memory usage
    if (memoryUsage.totalMB > maxMemoryMB) {
      return {
        feasible: false,
        reason: `Estimated memory usage (${memoryUsage.totalMB.toFixed(1)}MB) exceeds limit (${maxMemoryMB}MB)`,
        suggestion: 'Reduce scale factor, resize input image, or increase memory limit'
      }
    }
    
    return { feasible: true }
  }

  /**
   * Get quality comparison between methods
   */
  static getMethodQualityComparison() {
    return {
      'pixel-art': {
        best: ['scale2x', 'scale3x', 'scale4x'],
        good: ['nearest'],
        poor: ['bilinear', 'bicubic', 'lanczos']
      },
      'photographs': {
        best: ['lanczos', 'bicubic'],
        good: ['bilinear'],
        poor: ['nearest', 'scale2x', 'scale3x', 'scale4x']
      },
      'line-art': {
        best: ['nearest', 'scale2x'],
        good: ['bilinear'],
        poor: ['bicubic', 'lanczos']
      },
      'mixed-content': {
        best: ['bilinear', 'bicubic'],
        good: ['lanczos'],
        poor: ['nearest', 'scale2x']
      }
    }
  }

  /**
   * Auto-detect best scaling method for image content
   */
  static autoDetectBestMethod(
    imageData: ImageData,
    targetFactor: number
  ): { method: string; confidence: number; reason: string } {
    // This is a simplified heuristic - in practice, content detection is complex
    const canvas = new OffscreenCanvas(imageData.width, imageData.height)
    const ctx = canvas.getContext('2d')!
    const data = new globalThis.ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height)
    ctx.putImageData(data, 0, 0)
    
    const pixels = data.data
    let sharpEdges = 0
    let colorVariations = new Set<string>()
    let totalPixels = 0
    
    // Analyze image characteristics
    for (let y = 1; y < imageData.height - 1; y++) {
      for (let x = 1; x < imageData.width - 1; x++) {
        const idx = (y * imageData.width + x) * 4
        const color = `${pixels[idx]},${pixels[idx + 1]},${pixels[idx + 2]}`
        colorVariations.add(color)
        
        // Check for sharp edges
        const centerGray = pixels[idx] * 0.299 + pixels[idx + 1] * 0.587 + pixels[idx + 2] * 0.114
        const rightIdx = (y * imageData.width + (x + 1)) * 4
        const rightGray = pixels[rightIdx] * 0.299 + pixels[rightIdx + 1] * 0.587 + pixels[rightIdx + 2] * 0.114
        
        if (Math.abs(centerGray - rightGray) > 50) {
          sharpEdges++
        }
        
        totalPixels++
      }
    }
    
    const sharpEdgeRatio = sharpEdges / totalPixels
    const colorCount = colorVariations.size
    const pixelCount = imageData.width * imageData.height
    const colorDensity = colorCount / pixelCount
    
    // Decision logic
    if (colorCount < 256 && sharpEdgeRatio > 0.3) {
      // Likely pixel art
      if (targetFactor === 2) {
        return { method: 'scale2x', confidence: 0.9, reason: 'Low color count and sharp edges suggest pixel art' }
      } else if (targetFactor === 3) {
        return { method: 'scale3x', confidence: 0.9, reason: 'Low color count and sharp edges suggest pixel art' }
      } else if (targetFactor === 4) {
        return { method: 'scale4x', confidence: 0.9, reason: 'Low color count and sharp edges suggest pixel art' }
      } else {
        return { method: 'nearest', confidence: 0.8, reason: 'Pixel art characteristics detected' }
      }
    } else if (sharpEdgeRatio > 0.4) {
      // Line art or diagrams
      return { method: 'nearest', confidence: 0.7, reason: 'High edge density suggests line art' }
    } else if (colorDensity > 0.5) {
      // Photographic content
      return { method: 'lanczos', confidence: 0.8, reason: 'High color variation suggests photographic content' }
    } else {
      // Mixed or unknown content
      return { method: 'bilinear', confidence: 0.6, reason: 'Mixed content characteristics' }
    }
  }
}