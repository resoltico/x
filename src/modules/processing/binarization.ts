import type { BinarizationParams, ImageData } from '@/types'

/**
 * Binarization algorithms for converting grayscale images to black and white
 */

export class BinarizationProcessor {
  /**
   * Process image with specified binarization method
   */
  static async process(
    imageData: ImageData,
    params: BinarizationParams,
    vipsImage?: any
  ): Promise<any> {
    const { method, windowSize = 15, k = 0.2, threshold = 128 } = params

    if (vipsImage) {
      return this.processWithVips(vipsImage, method, windowSize, k, threshold)
    } else {
      return this.processWithCanvas(imageData, method, windowSize, k, threshold)
    }
  }

  /**
   * Process using wasm-vips
   */
  private static async processWithVips(
    image: any,
    method: string,
    windowSize: number,
    k: number,
    threshold: number
  ): Promise<any> {
    switch (method) {
      case 'otsu':
        return this.otsuThresholding(image, threshold)
      case 'sauvola':
        return this.sauvolaThresholding(image, windowSize, k)
      case 'niblack':
        return this.niblackThresholding(image, windowSize, k)
      default:
        return this.globalThresholding(image, threshold)
    }
  }

  /**
   * Process using Canvas API (fallback)
   */
  private static async processWithCanvas(
    imageData: ImageData,
    method: string,
    windowSize: number,
    k: number,
    threshold: number
  ): Promise<ArrayBuffer> {
    const canvas = new globalThis.OffscreenCanvas(imageData.width, imageData.height)
    const ctx = canvas.getContext('2d')!
    
    // Create ImageData object for canvas
    const canvasImageData = new globalThis.ImageData(
      new Uint8ClampedArray(imageData.data),
      imageData.width,
      imageData.height
    )
    
    ctx.putImageData(canvasImageData, 0, 0)

    switch (method) {
      case 'otsu':
        await this.canvasOtsuThresholding(ctx, imageData)
        break
      case 'sauvola':
        await this.canvasSauvolaThresholding(ctx, imageData, windowSize, k)
        break
      case 'niblack':
        await this.canvasNiblackThresholding(ctx, imageData, windowSize, k)
        break
      default:
        await this.canvasGlobalThresholding(ctx, imageData, threshold)
    }
    
    // Convert canvas to blob and then to ArrayBuffer
    const blob = await canvas.convertToBlob({ type: 'image/png' })
    return await blob.arrayBuffer()
  }

  /**
   * Global Otsu thresholding using VIPS
   */
  private static otsuThresholding(image: any, fallbackThreshold: number): any {
    try {
      // Calculate histogram
      const hist = image.histFind()
      const otsuThreshold = this.calculateOtsuThreshold(hist)
      return image.more(otsuThreshold)
    } catch (error) {
      console.warn('Otsu thresholding failed, using fallback threshold:', error)
      return image.more(fallbackThreshold)
    }
  }

  /**
   * Sauvola adaptive thresholding using VIPS
   */
  private static sauvolaThresholding(image: any, windowSize: number, k: number): any {
    try {
      // Calculate local mean and standard deviation
      const mean = image.conv(this.createGaussianKernel(windowSize))
      const variance = image.multiply(image).conv(this.createGaussianKernel(windowSize)).subtract(mean.multiply(mean))
      const stddev = variance.pow(0.5)
      
      // Sauvola threshold: T = mean * (1 + k * (stddev / 128 - 1))
      const threshold = mean.multiply(
        stddev.divide(128).subtract(1).multiply(k).add(1)
      )
      
      return image.more(threshold)
    } catch (error) {
      console.warn('Sauvola thresholding failed, using global threshold:', error)
      return image.more(128)
    }
  }

  /**
   * Niblack adaptive thresholding using VIPS
   */
  private static niblackThresholding(image: any, windowSize: number, k: number): any {
    try {
      // Calculate local mean and standard deviation
      const mean = image.conv(this.createGaussianKernel(windowSize))
      const variance = image.multiply(image).conv(this.createGaussianKernel(windowSize)).subtract(mean.multiply(mean))
      const stddev = variance.pow(0.5)
      
      // Niblack threshold: T = mean + k * stddev
      const threshold = mean.add(stddev.multiply(k))
      
      return image.more(threshold)
    } catch (error) {
      console.warn('Niblack thresholding failed, using global threshold:', error)
      return image.more(128)
    }
  }

  /**
   * Simple global thresholding using VIPS
   */
  private static globalThresholding(image: any, threshold: number): any {
    return image.more(threshold)
  }

  /**
   * Canvas-based Otsu thresholding
   */
  private static async canvasOtsuThresholding(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    
    // Calculate histogram
    const histogram = new Array(256).fill(0)
    for (let i = 0; i < pixels.length; i += 4) {
      const gray = Math.round(pixels[i] * 0.299 + pixels[i + 1] * 0.587 + pixels[i + 2] * 0.114)
      histogram[gray]++
    }
    
    // Calculate Otsu threshold
    const threshold = this.calculateOtsuThresholdFromHistogram(histogram)
    
    // Apply threshold
    for (let i = 0; i < pixels.length; i += 4) {
      const gray = pixels[i] * 0.299 + pixels[i + 1] * 0.587 + pixels[i + 2] * 0.114
      const binary = gray > threshold ? 255 : 0
      pixels[i] = binary
      pixels[i + 1] = binary
      pixels[i + 2] = binary
    }
    
    ctx.putImageData(data, 0, 0)
  }

  /**
   * Canvas-based Sauvola thresholding
   */
  private static async canvasSauvolaThresholding(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    windowSize: number,
    k: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    const newPixels = new Uint8ClampedArray(pixels)
    const halfWindow = Math.floor(windowSize / 2)
    
    for (let y = halfWindow; y < imageData.height - halfWindow; y++) {
      for (let x = halfWindow; x < imageData.width - halfWindow; x++) {
        const idx = (y * imageData.width + x) * 4
        
        // Calculate local mean and standard deviation
        let sum = 0
        let sumSq = 0
        let count = 0
        
        for (let wy = -halfWindow; wy <= halfWindow; wy++) {
          for (let wx = -halfWindow; wx <= halfWindow; wx++) {
            const widx = ((y + wy) * imageData.width + (x + wx)) * 4
            const gray = pixels[widx] * 0.299 + pixels[widx + 1] * 0.587 + pixels[widx + 2] * 0.114
            sum += gray
            sumSq += gray * gray
            count++
          }
        }
        
        const mean = sum / count
        const variance = (sumSq / count) - (mean * mean)
        const stddev = Math.sqrt(Math.max(0, variance))
        
        // Sauvola threshold: T = mean * (1 + k * (stddev / 128 - 1))
        const threshold = mean * (1 + k * (stddev / 128 - 1))
        
        const gray = pixels[idx] * 0.299 + pixels[idx + 1] * 0.587 + pixels[idx + 2] * 0.114
        const binary = gray > threshold ? 255 : 0
        
        newPixels[idx] = binary
        newPixels[idx + 1] = binary
        newPixels[idx + 2] = binary
      }
    }
    
    const newData = new globalThis.ImageData(newPixels, imageData.width, imageData.height)
    ctx.putImageData(newData, 0, 0)
  }

  /**
   * Canvas-based Niblack thresholding
   */
  private static async canvasNiblackThresholding(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    windowSize: number,
    k: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    const newPixels = new Uint8ClampedArray(pixels)
    const halfWindow = Math.floor(windowSize / 2)
    
    for (let y = halfWindow; y < imageData.height - halfWindow; y++) {
      for (let x = halfWindow; x < imageData.width - halfWindow; x++) {
        const idx = (y * imageData.width + x) * 4
        
        // Calculate local mean and standard deviation
        let sum = 0
        let sumSq = 0
        let count = 0
        
        for (let wy = -halfWindow; wy <= halfWindow; wy++) {
          for (let wx = -halfWindow; wx <= halfWindow; wx++) {
            const widx = ((y + wy) * imageData.width + (x + wx)) * 4
            const gray = pixels[widx] * 0.299 + pixels[widx + 1] * 0.587 + pixels[widx + 2] * 0.114
            sum += gray
            sumSq += gray * gray
            count++
          }
        }
        
        const mean = sum / count
        const variance = (sumSq / count) - (mean * mean)
        const stddev = Math.sqrt(Math.max(0, variance))
        
        // Niblack threshold: T = mean + k * stddev
        const threshold = mean + k * stddev
        
        const gray = pixels[idx] * 0.299 + pixels[idx + 1] * 0.587 + pixels[idx + 2] * 0.114
        const binary = gray > threshold ? 255 : 0
        
        newPixels[idx] = binary
        newPixels[idx + 1] = binary
        newPixels[idx + 2] = binary
      }
    }
    
    const newData = new globalThis.ImageData(newPixels, imageData.width, imageData.height)
    ctx.putImageData(newData, 0, 0)
  }

  /**
   * Canvas-based global thresholding
   */
  private static async canvasGlobalThresholding(
    ctx: OffscreenCanvasRenderingContext2D,
    imageData: ImageData,
    threshold: number
  ) {
    const data = ctx.getImageData(0, 0, imageData.width, imageData.height)
    const pixels = data.data
    
    for (let i = 0; i < pixels.length; i += 4) {
      const gray = pixels[i] * 0.299 + pixels[i + 1] * 0.587 + pixels[i + 2] * 0.114
      const binary = gray > threshold ? 255 : 0
      pixels[i] = binary
      pixels[i + 1] = binary
      pixels[i + 2] = binary
    }
    
    ctx.putImageData(data, 0, 0)
  }

  /**
   * Calculate Otsu threshold from VIPS histogram
   */
  private static calculateOtsuThreshold(histogram: any): number {
    try {
      const data = histogram.getpoint(0, 0)
      return this.calculateOtsuThresholdFromHistogram(data)
    } catch (error) {
      console.warn('Failed to calculate Otsu threshold from VIPS histogram:', error)
      return 128 // fallback
    }
  }

  /**
   * Calculate Otsu threshold from histogram array
   */
  private static calculateOtsuThresholdFromHistogram(histogram: number[]): number {
    const total = histogram.reduce((sum, val) => sum + val, 0)
    if (total === 0) return 128

    let sum = 0
    for (let i = 0; i < 256; i++) {
      sum += i * histogram[i]
    }

    let sumB = 0
    let wB = 0
    let wF = 0
    let maxVariance = 0
    let threshold = 0

    for (let i = 0; i < 256; i++) {
      wB += histogram[i]
      if (wB === 0) continue

      wF = total - wB
      if (wF === 0) break

      sumB += i * histogram[i]
      const mB = sumB / wB
      const mF = (sum - sumB) / wF

      const variance = wB * wF * (mB - mF) * (mB - mF)
      if (variance > maxVariance) {
        maxVariance = variance
        threshold = i
      }
    }

    return threshold
  }

  /**
   * Create Gaussian kernel for smoothing
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
   * Get algorithm-specific parameter constraints
   */
  static getParameterConstraints() {
    return {
      otsu: {
        threshold: { min: 0, max: 255, default: 128 }
      },
      sauvola: {
        windowSize: { min: 3, max: 51, step: 2, default: 15 },
        k: { min: -1, max: 1, step: 0.01, default: 0.2 }
      },
      niblack: {
        windowSize: { min: 3, max: 51, step: 2, default: 15 },
        k: { min: -1, max: 1, step: 0.01, default: -0.2 }
      }
    }
  }

  /**
   * Get recommended parameters for different image types
   */
  static getRecommendedParameters(imageType: string) {
    const recommendations = {
      document: {
        method: 'sauvola',
        windowSize: 15,
        k: 0.2
      },
      engraving: {
        method: 'otsu',
        threshold: 128
      },
      manuscript: {
        method: 'niblack',
        windowSize: 21,
        k: -0.2
      },
      photograph: {
        method: 'otsu',
        threshold: 128
      }
    }

    return recommendations[imageType as keyof typeof recommendations] || recommendations.document
  }
}