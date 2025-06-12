import type { 
  ImageData, 
  ProcessingParameters, 
  ProcessingType,
  BinarizationParams,
  MorphologyParams,
  NoiseReductionParams,
  ScalingParams
} from '@/types'

/**
 * ProcessingModule handles image processing operations using wasm-vips and custom algorithms
 */
export class ProcessingModule {
  private static instance: ProcessingModule
  private vips: any = null
  private isInitialized = false

  static getInstance(): ProcessingModule {
    if (!ProcessingModule.instance) {
      ProcessingModule.instance = new ProcessingModule()
    }
    return ProcessingModule.instance
  }

  /**
   * Initialize the processing module with wasm-vips
   */
  async initialize(): Promise<void> {
    if (this.isInitialized) return

    try {
      // Dynamic import of wasm-vips
      const { Vips } = await import('wasm-vips')
      this.vips = await Vips({
        locateFile: (file: string) => `https://cdn.jsdelivr.net/npm/wasm-vips@0.0.13/lib/${file}`
      })
      
      this.isInitialized = true
      console.log('ProcessingModule initialized with wasm-vips')
    } catch (error) {
      console.error('Failed to initialize ProcessingModule:', error)
      throw new Error('Failed to initialize image processing engine')
    }
  }

  /**
   * Process an image with specified parameters
   */
  async processImage(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): Promise<ArrayBuffer> {
    if (!this.isInitialized) {
      await this.initialize()
    }

    try {
      // Load image into vips
      const vipsImage = this.vips.Image.newFromBuffer(new Uint8Array(imageData.data))

      let result: any

      switch (type) {
        case 'binarization':
          result = await this.processBinarization(vipsImage, parameters.binarization!)
          break
        case 'morphology':
          result = await this.processMorphology(vipsImage, parameters.morphology!)
          break
        case 'noise-reduction':
          result = await this.processNoiseReduction(vipsImage, parameters.noise!)
          break
        case 'scaling':
          result = await this.processScaling(vipsImage, parameters.scaling!)
          break
        default:
          throw new Error(`Unsupported processing type: ${type}`)
      }

      // Convert result back to ArrayBuffer
      const outputBuffer = result.writeToBuffer('.png')
      return outputBuffer.buffer.slice(
        outputBuffer.byteOffset,
        outputBuffer.byteOffset + outputBuffer.byteLength
      )
    } catch (error) {
      console.error('Processing error:', error)
      throw new Error(`Processing failed: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  }

  /**
   * Process binarization using various algorithms
   */
  private async processBinarization(
    image: any,
    params: BinarizationParams
  ): Promise<any> {
    const { method, windowSize = 15, k = 0.2, threshold = 128 } = params

    switch (method) {
      case 'otsu':
        // Global Otsu thresholding
        const hist = image.histFind()
        const otsuThreshold = this.calculateOtsuThreshold(hist)
        return image.more(otsuThreshold)

      case 'sauvola':
        // Sauvola adaptive thresholding
        return this.sauvolaThresholding(image, windowSize, k)

      case 'niblack':
        // Niblack adaptive thresholding
        return this.niblackThresholding(image, windowSize, k)

      default:
        // Simple global thresholding
        return image.more(threshold)
    }
  }

  /**
   * Process morphological operations
   */
  private async processMorphology(
    image: any,
    params: MorphologyParams
  ): Promise<any> {
    const { operation, kernelSize, iterations } = params

    // Create morphological kernel
    const kernel = this.vips.Image.newFromArray(
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
      }
    }

    return result
  }

  /**
   * Process noise reduction
   */
  private async processNoiseReduction(
    image: any,
    params: NoiseReductionParams
  ): Promise<any> {
    const { method, kernelSize = 3, threshold = 128 } = params

    switch (method) {
      case 'median':
        return image.median(kernelSize)

      case 'binary-noise-removal':
        // Remove small binary noise components
        return this.removeBinaryNoise(image, threshold)

      default:
        return image.median(kernelSize)
    }
  }

  /**
   * Process scaling using various algorithms
   */
  private async processScaling(
    image: any,
    params: ScalingParams
  ): Promise<any> {
    const { method, factor } = params

    switch (method) {
      case 'nearest':
        return image.resize(factor, { kernel: 'nearest' })

      case 'bilinear':
        return image.resize(factor, { kernel: 'linear' })

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
   * Calculate Otsu threshold from histogram
   */
  private calculateOtsuThreshold(histogram: any): number {
    const data = histogram.getpoint(0, 0)
    const total = data.reduce((sum: number, val: number) => sum + val, 0)
    
    let sum = 0
    for (let i = 0; i < 256; i++) {
      sum += i * data[i]
    }

    let sumB = 0
    let wB = 0
    let wF = 0
    let maxVariance = 0
    let threshold = 0

    for (let i = 0; i < 256; i++) {
      wB += data[i]
      if (wB === 0) continue

      wF = total - wB
      if (wF === 0) break

      sumB += i * data[i]
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
   * Sauvola adaptive thresholding
   */
  private sauvolaThresholding(image: any, windowSize: number, k: number): any {
    // Calculate local mean and standard deviation
    const mean = image.conv(this.createGaussianKernel(windowSize))
    const variance = image.multiply(image).conv(this.createGaussianKernel(windowSize)).subtract(mean.multiply(mean))
    const stddev = variance.pow(0.5)
    
    // Sauvola threshold: T = mean * (1 + k * (stddev / 128 - 1))
    const threshold = mean.multiply(
      stddev.divide(128).subtract(1).multiply(k).add(1)
    )
    
    return image.more(threshold)
  }

  /**
   * Niblack adaptive thresholding
   */
  private niblackThresholding(image: any, windowSize: number, k: number): any {
    // Calculate local mean and standard deviation
    const mean = image.conv(this.createGaussianKernel(windowSize))
    const variance = image.multiply(image).conv(this.createGaussianKernel(windowSize)).subtract(mean.multiply(mean))
    const stddev = variance.pow(0.5)
    
    // Niblack threshold: T = mean + k * stddev
    const threshold = mean.add(stddev.multiply(k))
    
    return image.more(threshold)
  }

  /**
   * Remove small binary noise components
   */
  private removeBinaryNoise(image: any, minSize: number): any {
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
  }

  /**
   * Pixel art scaling algorithms (Scale2x, Scale3x, Scale4x)
   */
  private pixelArtScaling(image: any, method: string): any {
    // For now, use simple nearest neighbor scaling
    // In a full implementation, these would use the actual Scale2x/3x/4x algorithms
    const factor = parseInt(method.replace('scale', '').replace('x', ''))
    return image.resize(factor, { kernel: 'nearest' })
  }

  /**
   * Create morphology kernel
   */
  private createMorphologyKernel(size: number): number[][] {
    const kernel: number[][] = []
    const center = Math.floor(size / 2)
    
    for (let y = 0; y < size; y++) {
      kernel[y] = []
      for (let x = 0; x < size; x++) {
        // Create circular kernel
        const distance = Math.sqrt((x - center) ** 2 + (y - center) ** 2)
        kernel[y][x] = distance <= center ? 1 : 0
      }
    }
    
    return kernel
  }

  /**
   * Create Gaussian kernel for smoothing
   */
  private createGaussianKernel(size: number): number[][] {
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
   * Create preview version of processing (lower resolution for speed)
   */
  async processPreview(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters,
    maxDimension: number = 400
  ): Promise<ArrayBuffer> {
    if (!this.isInitialized) {
      await this.initialize()
    }

    try {
      // Load and resize image for preview
      const vipsImage = this.vips.Image.newFromBuffer(new Uint8Array(imageData.data))
      const scale = Math.min(maxDimension / vipsImage.width, maxDimension / vipsImage.height, 1)
      const previewImage = scale < 1 ? vipsImage.resize(scale) : vipsImage

      // Process the preview
      const result = await this.processImageInternal(previewImage, type, parameters)
      
      // Convert result back to ArrayBuffer
      const outputBuffer = result.writeToBuffer('.png')
      return outputBuffer.buffer.slice(
        outputBuffer.byteOffset,
        outputBuffer.byteOffset + outputBuffer.byteLength
      )
    } catch (error) {
      console.error('Preview processing error:', error)
      throw error
    }
  }

  /**
   * Internal processing method (without initialization check)
   */
  private async processImageInternal(
    vipsImage: any,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): Promise<any> {
    switch (type) {
      case 'binarization':
        return this.processBinarization(vipsImage, parameters.binarization!)
      case 'morphology':
        return this.processMorphology(vipsImage, parameters.morphology!)
      case 'noise-reduction':
        return this.processNoiseReduction(vipsImage, parameters.noise!)
      case 'scaling':
        return this.processScaling(vipsImage, parameters.scaling!)
      default:
        throw new Error(`Unsupported processing type: ${type}`)
    }
  }

  /**
   * Get available processing algorithms
   */
  getAvailableAlgorithms(): Record<ProcessingType, string[]> {
    return {
      'binarization': ['otsu', 'sauvola', 'niblack'],
      'morphology': ['opening', 'closing', 'dilation', 'erosion'],
      'noise-reduction': ['median', 'binary-noise-removal'],
      'scaling': ['scale2x', 'scale3x', 'scale4x', 'nearest', 'bilinear']
    }
  }

  /**
   * Check if the module is initialized
   */
  isReady(): boolean {
    return this.isInitialized
  }
}