import type { NoiseReductionParams, ImageData } from '@/types'

/**
 * Noise reduction algorithms for image processing
 */

export class NoiseReductionProcessor {
  /**
   * Process image with specified noise reduction method
   */
  static async process(
    imageData: ImageData,
    params: NoiseReductionParams,
    vipsImage?: any
  ): Promise<any> {
    const { method, kernelSize = 3, threshold = 128 } = params

    if (vipsImage) {
      return this.processWithVips(vipsImage, method, kernelSize, threshold)
    } else {
      return this.processWithCanvas(imageData, method, kernelSize, threshold)
    }
  }

  /**
   * Process using wasm-vips
   */
  private static async processWithVips(
    image: any,
    method: string,
    kernelSize: number,
    threshold: number
  ): Promise<any> {
    switch (method) {
      case 'median':
        return this.medianFilter(image, kernelSize)
      case 'binary-noise-removal':
        return this.removeBinaryNoise(image, threshold)
      case 'gaussian':
        return this.gaussianBlur(image, kernelSize)
      case 'bilateral':
        return this.bilateralFilter(image, kernelSize)
      default:
        return this.medianFilter(image, kernelSize)
    }
  }

  /**
   * Process using Canvas API (fallback)
   */
  private static async processWithCanvas(
    imageData: ImageData,
    method: string,
    kernelSize: number,
    threshold: number
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

    switch (method) {
      case 'median':
        await this.canvasMedianFilter(ctx, imageData, kernelSize)
        break
      case 'binary-noise-removal':
        await this.canvasBinaryNoiseRemoval(ctx, imageData, threshold)
        break
      case 'gaussian':
        await this.canvasGaussianBlur(ctx, imageData, kernelSize)
        break
      case 'bilateral':
        await this.canvasBilateralFilter(ctx, imageData, kernelSize)
        break
      default:
        await this.canvasMedianFilter(ctx, imageData, kernelSize)
    }
    
    // Convert canvas to blob and then to ArrayBuffer
    const blob = await canvas.convertToBlob({ type: 'image/png' })
    return await blob.arrayBuffer()
  }

  /**
   * Median filter using VIPS
   */
  private static medianFilter(image: any, kernelSize: number): any {
    try {
      return image.median(kernelSize)
    } catch (error) {
      console.warn('VIPS median filter failed:', error)
      return image
    }
  }

  /**
   * Gaussian blur using VIPS
   */
  private static gaussianBlur(image: any, kernelSize: number): any {
    try {
      const sigma = kernelSize / 3
      return image.gaussblur(sigma)
    } catch (error) {
      console.warn('VIPS Gaussian blur failed:', error)
      return image
    }
  }

  /**
   * Bilateral filter using VIPS (approximation)
   */
  private static bilateralFilter(image: any, kernelSize: number): any {
    try {
      // VIPS doesn't have direct bilateral filter, use edge-preserving smoothing
      const sigma = kernelSize / 3
      return image.gaussblur(sigma * 0.5)
    } catch (error) {
      console.warn('VIPS bilateral filter failed:', error)
      return image
    }
  }

  /**
   * Remove small binary noise components using VIPS
   */
  private static removeBinaryNoise(image: any, minSize: number): any {
    try {
      // Use connected component analysis to remove small components
      const labels = image.labelregions()
      const stats = labels.regionShrink('mean')
      
      // Filter out small regions
      let result = image
      for (let i = 1; i < stats.height; i++) {
        const area = stats.getpoint(3, i)[0] // Area is in column 3
        if (area < minSize) {
          const mask = labels.equal(i)
          result = result.ifthenelse(0, result, mask)
        }
      }
      
      return result
    } catch (error) {
      console.warn('Binary noise removal failed, returning original:', error)
      return image
    }
  }

  /**
   * Canvas-based median filter
   */
  private static async canvasMedianFilter(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    kernelSize: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    const newPixels = new Uint8ClampedArray(pixels)
    const halfKernel = Math.floor(kernelSize / 2)

    for (let y = halfKernel; y < imageData.height - halfKernel; y++) {
      for (let x = halfKernel; x < imageData.width - halfKernel; x++) {
        const idx = (y * imageData.width + x) * 4
        
        const rValues: number[] = []
        const gValues: number[] = []
        const bValues: number[] = []
        
        // Collect neighborhood values
        for (let ky = -halfKernel; ky <= halfKernel; ky++) {
          for (let kx = -halfKernel; kx <= halfKernel; kx++) {
            const kidx = ((y + ky) * imageData.width + (x + kx)) * 4
            rValues.push(pixels[kidx])
            gValues.push(pixels[kidx + 1])
            bValues.push(pixels[kidx + 2])
          }
        }
        
        // Sort and find median
        rValues.sort((a, b) => a - b)
        gValues.sort((a, b) => a - b)
        bValues.sort((a, b) => a - b)
        
        const medianIndex = Math.floor(rValues.length / 2)
        
        newPixels[idx] = rValues[medianIndex]
        newPixels[idx + 1] = gValues[medianIndex]
        newPixels[idx + 2] = bValues[medianIndex]
        // Alpha channel remains unchanged
      }
    }
    
    const newData = new ImageData(newPixels, imageData.width, imageData.height)
    ctx.putImageData(newData, 0, 0)
  }

  /**
   * Canvas-based binary noise removal
   */
  private static async canvasBinaryNoiseRemoval(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    threshold: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    const visited = new Array(imageData.width * imageData.height).fill(false)
    const newPixels = new Uint8ClampedArray(pixels)

    // Find connected components and remove small ones
    for (let y = 0; y < imageData.height; y++) {
      for (let x = 0; x < imageData.width; x++) {
        const idx = y * imageData.width + x
        
        if (!visited[idx]) {
          const pixelIdx = idx * 4
          const gray = pixels[pixelIdx] * 0.299 + pixels[pixelIdx + 1] * 0.587 + pixels[pixelIdx + 2] * 0.114
          
          // If it's a foreground pixel (assuming dark text on light background)
          if (gray < 128) {
            const component = this.floodFill(pixels, visited, x, y, imageData.width, imageData.height)
            
            // If component is too small, remove it
            if (component.length < threshold) {
              for (const pos of component) {
                const removeIdx = pos * 4
                newPixels[removeIdx] = 255
                newPixels[removeIdx + 1] = 255
                newPixels[removeIdx + 2] = 255
              }
            }
          }
        }
      }
    }
    
    const newData = new ImageData(newPixels, imageData.width, imageData.height)
    ctx.putImageData(newData, 0, 0)
  }

  /**
   * Canvas-based Gaussian blur
   */
  private static async canvasGaussianBlur(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    kernelSize: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    const newPixels = new Uint8ClampedArray(pixels)
    const kernel = this.createGaussianKernel(kernelSize)
    const halfKernel = Math.floor(kernelSize / 2)

    for (let y = halfKernel; y < imageData.height - halfKernel; y++) {
      for (let x = halfKernel; x < imageData.width - halfKernel; x++) {
        const idx = (y * imageData.width + x) * 4
        
        let r = 0, g = 0, b = 0
        
        // Apply Gaussian kernel
        for (let ky = 0; ky < kernelSize; ky++) {
          for (let kx = 0; kx < kernelSize; kx++) {
            const py = y + ky - halfKernel
            const px = x + kx - halfKernel
            const pidx = (py * imageData.width + px) * 4
            const weight = kernel[ky][kx]
            
            r += pixels[pidx] * weight
            g += pixels[pidx + 1] * weight
            b += pixels[pidx + 2] * weight
          }
        }
        
        newPixels[idx] = Math.round(r)
        newPixels[idx + 1] = Math.round(g)
        newPixels[idx + 2] = Math.round(b)
      }
    }
    
    const newData = new ImageData(newPixels, imageData.width, imageData.height)
    ctx.putImageData(newData, 0, 0)
  }

  /**
   * Canvas-based bilateral filter (simplified)
   */
  private static async canvasBilateralFilter(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    kernelSize: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    const newPixels = new Uint8ClampedArray(pixels)
    const halfKernel = Math.floor(kernelSize / 2)
    const sigmaSpatial = kernelSize / 3
    const sigmaIntensity = 50

    for (let y = halfKernel; y < imageData.height - halfKernel; y++) {
      for (let x = halfKernel; x < imageData.width - halfKernel; x++) {
        const idx = (y * imageData.width + x) * 4
        const centerR = pixels[idx]
        const centerG = pixels[idx + 1]
        const centerB = pixels[idx + 2]
        
        let r = 0, g = 0, b = 0, weightSum = 0
        
        // Apply bilateral filter
        for (let ky = -halfKernel; ky <= halfKernel; ky++) {
          for (let kx = -halfKernel; kx <= halfKernel; kx++) {
            const py = y + ky
            const px = x + kx
            const pidx = (py * imageData.width + px) * 4
            
            // Spatial weight
            const spatialDist = Math.sqrt(kx * kx + ky * ky)
            const spatialWeight = Math.exp(-(spatialDist * spatialDist) / (2 * sigmaSpatial * sigmaSpatial))
            
            // Intensity weight
            const intensityDist = Math.sqrt(
              Math.pow(pixels[pidx] - centerR, 2) +
              Math.pow(pixels[pidx + 1] - centerG, 2) +
              Math.pow(pixels[pidx + 2] - centerB, 2)
            )
            const intensityWeight = Math.exp(-(intensityDist * intensityDist) / (2 * sigmaIntensity * sigmaIntensity))
            
            const weight = spatialWeight * intensityWeight
            
            r += pixels[pidx] * weight
            g += pixels[pidx + 1] * weight
            b += pixels[pidx + 2] * weight
            weightSum += weight
          }
        }
        
        newPixels[idx] = Math.round(r / weightSum)
        newPixels[idx + 1] = Math.round(g / weightSum)
        newPixels[idx + 2] = Math.round(b / weightSum)
      }
    }
    
    const newData = new ImageData(newPixels, imageData.width, imageData.height)
    ctx.putImageData(newData, 0, 0)
  }

  /**
   * Flood fill algorithm for connected component analysis
   */
  private static floodFill(
    pixels: Uint8ClampedArray,
    visited: boolean[],
    startX: number,
    startY: number,
    width: number,
    height: number
  ): number[] {
    const component: number[] = []
    const stack: Array<[number, number]> = [[startX, startY]]
    
    while (stack.length > 0) {
      const [x, y] = stack.pop()!
      const idx = y * width + x
      
      if (x < 0 || x >= width || y < 0 || y >= height || visited[idx]) {
        continue
      }
      
      const pixelIdx = idx * 4
      const gray = pixels[pixelIdx] * 0.299 + pixels[pixelIdx + 1] * 0.587 + pixels[pixelIdx + 2] * 0.114
      
      // If not a foreground pixel, skip
      if (gray >= 128) {
        continue
      }
      
      visited[idx] = true
      component.push(idx)
      
      // Add neighbors to stack
      stack.push([x + 1, y])
      stack.push([x - 1, y])
      stack.push([x, y + 1])
      stack.push([x, y - 1])
    }
    
    return component
  }

  /**
   * Create Gaussian kernel
   */
  private static createGaussianKernel(size: number): number[][] {
    const kernel: number[][] = []
    const sigma = size / 3
    const center = Math.floor(size / 2)
    let sum = 0
    
    for (let y = 0; y < size; y++) {
      kernel[y] = []
      for (let x = 0; x < size; x++) {
        const distance = (x - center) ** 2 + (y - center) ** 2
        const value = Math.exp(-distance / (2 * sigma ** 2))
        kernel[y][x] = value
        sum += value
      }
    }
    
    // Normalize kernel
    for (let y = 0; y < size; y++) {
      for (let x = 0; x < size; x++) {
        kernel[y][x] /= sum
      }
    }
    
    return kernel
  }

  /**
   * Get parameter constraints for noise reduction
   */
  static getParameterConstraints() {
    return {
      median: {
        kernelSize: { min: 3, max: 9, step: 2, default: 3 }
      },
      gaussian: {
        kernelSize: { min: 3, max: 15, step: 2, default: 5 }
      },
      bilateral: {
        kernelSize: { min: 3, max: 11, step: 2, default: 5 }
      },
      'binary-noise-removal': {
        threshold: { min: 1, max: 1000, default: 50 }
      }
    }
  }

  /**
   * Get recommended parameters for different noise types
   */
  static getRecommendedParameters(noiseType: string) {
    const recommendations = {
      'salt-and-pepper': {
        method: 'median',
        kernelSize: 3
      },
      'gaussian-noise': {
        method: 'gaussian',
        kernelSize: 5
      },
      'speckle-noise': {
        method: 'bilateral',
        kernelSize: 5
      },
      'impulse-noise': {
        method: 'median',
        kernelSize: 5
      },
      'small-artifacts': {
        method: 'binary-noise-removal',
        threshold: 50
      },
      'texture-noise': {
        method: 'bilateral',
        kernelSize: 7
      }
    }

    return recommendations[noiseType as keyof typeof recommendations] || recommendations['salt-and-pepper']
  }

  /**
   * Validate noise reduction parameters
   */
  static validateParameters(params: NoiseReductionParams): { isValid: boolean; errors: string[] } {
    const errors: string[] = []
    const constraints = this.getParameterConstraints()

    if (params.method in constraints) {
      const methodConstraints = constraints[params.method as keyof typeof constraints] as any

      if (methodConstraints.kernelSize && params.kernelSize) {
        if (params.kernelSize < methodConstraints.kernelSize.min || 
            params.kernelSize > methodConstraints.kernelSize.max) {
          errors.push(`Kernel size must be between ${methodConstraints.kernelSize.min} and ${methodConstraints.kernelSize.max}`)
        }

        if (params.kernelSize % 2 === 0) {
          errors.push('Kernel size must be odd')
        }
      }

      if (methodConstraints.threshold && params.threshold) {
        if (params.threshold < methodConstraints.threshold.min || 
            params.threshold > methodConstraints.threshold.max) {
          errors.push(`Threshold must be between ${methodConstraints.threshold.min} and ${methodConstraints.threshold.max}`)
        }
      }
    }

    return {
      isValid: errors.length === 0,
      errors
    }
  }

  /**
   * Get method descriptions
   */
  static getMethodDescriptions() {
    return {
      median: 'Replaces each pixel with the median value of its neighborhood. Excellent for salt-and-pepper noise.',
      gaussian: 'Applies Gaussian blur to smooth the image. Good for reducing Gaussian noise but may blur edges.',
      bilateral: 'Edge-preserving smoothing that reduces noise while maintaining sharp edges.',
      'binary-noise-removal': 'Removes small connected components from binary images. Good for cleaning up scanned documents.'
    }
  }

  /**
   * Detect noise type automatically (simplified heuristic)
   */
  static detectNoiseType(imageData: ImageData): string {
    // This is a simplified heuristic - in practice, noise detection is complex
    const canvas = new OffscreenCanvas(imageData.width, imageData.height)
    const ctx = canvas.getContext('2d')!
    const data = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height)
    ctx.putImageData(data, 0, 0)
    
    const pixels = data.data
    let extremePixels = 0
    let totalPixels = 0
    
    for (let i = 0; i < pixels.length; i += 4) {
      const gray = pixels[i] * 0.299 + pixels[i + 1] * 0.587 + pixels[i + 2] * 0.114
      if (gray < 10 || gray > 245) {
        extremePixels++
      }
      totalPixels++
    }
    
    const extremeRatio = extremePixels / totalPixels
    
    if (extremeRatio > 0.1) {
      return 'salt-and-pepper'
    } else if (extremeRatio > 0.05) {
      return 'impulse-noise'
    } else {
      return 'gaussian-noise'
    }
  }
}