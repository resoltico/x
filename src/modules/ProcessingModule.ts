// src/modules/ProcessingModule.ts
import type { 
  ImageData, 
  ProcessingParameters, 
  ProcessingType
} from '@/types'

import { BinarizationProcessor } from './processing/binarization'
import { MorphologyProcessor } from './processing/morphology'
import { NoiseReductionProcessor } from './processing/noiseReduction'
import { ScalingProcessor } from './processing/scaling'

/**
 * Main processing module that coordinates different image processing algorithms
 */
export class ProcessingModule {
  private static instance: ProcessingModule
  private vips: any = null
  private isInitialized = false
  private initializationPromise: Promise<void> | null = null

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
    
    // Prevent multiple simultaneous initialization attempts
    if (this.initializationPromise) {
      return this.initializationPromise
    }

    this.initializationPromise = this.performInitialization()
    return this.initializationPromise
  }

  private async performInitialization(): Promise<void> {
    try {
      console.log('Initializing ProcessingModule...')
      
      // Check if we're in a test environment
      if (typeof process !== 'undefined' && process.env?.NODE_ENV === 'test') {
        console.log('Test environment detected, skipping WASM initialization')
        this.isInitialized = true
        return
      }

      // Check if we're in a browser environment
      if (typeof window === 'undefined' || typeof navigator === 'undefined') {
        console.log('Non-browser environment detected, skipping WASM initialization')
        this.isInitialized = true
        return
      }

      // Dynamic import of wasm-vips with proper error handling
      try {
        const wasmVips = await import('wasm-vips')
        
        // Initialize with CDN path - this avoids eval issues
        this.vips = await wasmVips.default({
          dynamicLibraries: [`https://cdn.jsdelivr.net/npm/wasm-vips@0.0.13/lib/vips.wasm`],
          locateFile: (file: string) => {
            // Return the full CDN path for WASM files
            if (file.endsWith('.wasm')) {
              return `https://cdn.jsdelivr.net/npm/wasm-vips@0.0.13/lib/${file}`
            }
            return file
          }
        })
        
        this.isInitialized = true
        console.log('ProcessingModule initialized with wasm-vips successfully')
      } catch (wasmError) {
        console.warn('Failed to initialize ProcessingModule with WASM:', wasmError)
        // Fallback to manual processing without WASM
        console.log('Falling back to JavaScript-based processing')
        this.isInitialized = true
      }
    } catch (error) {
      console.error('Failed to initialize ProcessingModule:', error)
      // Still mark as initialized to allow fallback processing
      this.isInitialized = true
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
      console.log(`Processing image with type: ${type}`, parameters)
      
      if (this.vips) {
        return await this.processWithVips(imageData, type, parameters)
      } else {
        return await this.processWithCanvas(imageData, type, parameters)
      }
    } catch (error) {
      console.error('Processing error:', error)
      throw new Error(`Processing failed: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  }

  /**
   * Process using wasm-vips
   */
  private async processWithVips(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): Promise<ArrayBuffer> {
    console.log('Processing with WASM-VIPS')
    
    try {
      // Load image into vips
      const vipsImage = this.vips.Image.newFromBuffer(new Uint8Array(imageData.data))

      let result: any

      switch (type) {
        case 'binarization':
          result = await BinarizationProcessor.process(imageData, parameters.binarization!, vipsImage)
          break
        case 'morphology':
          result = await MorphologyProcessor.process(imageData, parameters.morphology!, vipsImage)
          break
        case 'noise-reduction':
          result = await NoiseReductionProcessor.process(imageData, parameters.noise!, vipsImage)
          break
        case 'scaling':
          result = await ScalingProcessor.process(imageData, parameters.scaling!, vipsImage)
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
      console.warn('WASM processing failed, falling back to Canvas API:', error)
      return this.processWithCanvas(imageData, type, parameters)
    }
  }

  /**
   * Fallback processing using Canvas API (when WASM is not available)
   */
  private async processWithCanvas(
    imageData: ImageData,
    type: ProcessingType,
    parameters: ProcessingParameters
  ): Promise<ArrayBuffer> {
    console.log('Processing with Canvas API fallback')
    
    try {
      switch (type) {
        case 'binarization':
          return await BinarizationProcessor.process(imageData, parameters.binarization!)
        case 'morphology':
          return await MorphologyProcessor.process(imageData, parameters.morphology!)
        case 'noise-reduction':
          return await NoiseReductionProcessor.process(imageData, parameters.noise!)
        case 'scaling':
          return await ScalingProcessor.process(imageData, parameters.scaling!)
        default:
          throw new Error(`Unsupported processing type: ${type}`)
      }
    } catch (error) {
      console.error('Canvas processing failed:', error)
      throw error
    }
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
      console.log(`Processing preview with type: ${type}, max dimension: ${maxDimension}`)
      
      // Calculate preview scale
      const scale = Math.min(maxDimension / imageData.width, maxDimension / imageData.height, 1)
      
      let previewData = imageData
      
      // If we need to scale down for preview
      if (scale < 1) {
        const previewWidth = Math.round(imageData.width * scale)
        const previewHeight = Math.round(imageData.height * scale)
        
        // Check if we can use OffscreenCanvas
        if (typeof OffscreenCanvas !== 'undefined') {
          // Create preview using OffscreenCanvas scaling
          const canvas = new OffscreenCanvas(imageData.width, imageData.height)
          const ctx = canvas.getContext('2d')!
          
          const canvasImageData = new ImageData(
            new Uint8ClampedArray(imageData.data),
            imageData.width,
            imageData.height
          )
          ctx.putImageData(canvasImageData, 0, 0)
          
          const previewCanvas = new OffscreenCanvas(previewWidth, previewHeight)
          const previewCtx = previewCanvas.getContext('2d')!
          previewCtx.drawImage(canvas, 0, 0, previewWidth, previewHeight)
          
          const previewBlob = await previewCanvas.convertToBlob()
          const previewArrayBuffer = await previewBlob.arrayBuffer()
          
          previewData = {
            ...imageData,
            data: previewArrayBuffer,
            width: previewWidth,
            height: previewHeight
          }
        } else {
          // Fallback: use original data (in test environment)
          console.log('OffscreenCanvas not available, using original data for preview')
        }
      }

      // Process the preview data
      return await this.processImage(previewData, type, parameters)
    } catch (error) {
      console.error('Preview processing error:', error)
      throw error
    }
  }

  /**
   * Get available processing algorithms
   */
  getAvailableAlgorithms(): Record<ProcessingType, string[]> {
    return {
      'binarization': ['otsu', 'sauvola', 'niblack'],
      'morphology': ['opening', 'closing', 'dilation', 'erosion'],
      'noise-reduction': ['median', 'binary-noise-removal', 'gaussian', 'bilateral'],
      'scaling': ['scale2x', 'scale3x', 'scale4x', 'nearest', 'bilinear', 'bicubic', 'lanczos']
    }
  }

  /**
   * Validate processing parameters
   */
  validateParameters(type: ProcessingType, parameters: ProcessingParameters): { isValid: boolean; errors: string[] } {
    switch (type) {
      case 'binarization':
        return parameters.binarization 
          ? { isValid: true, errors: [] }
          : { isValid: false, errors: ['Binarization parameters required'] }
      
      case 'morphology':
        return parameters.morphology 
          ? MorphologyProcessor.validateParameters(parameters.morphology)
          : { isValid: false, errors: ['Morphology parameters required'] }
      
      case 'noise-reduction':
        return parameters.noise 
          ? NoiseReductionProcessor.validateParameters(parameters.noise)
          : { isValid: false, errors: ['Noise reduction parameters required'] }
      
      case 'scaling':
        return parameters.scaling 
          ? ScalingProcessor.validateParameters(parameters.scaling)
          : { isValid: false, errors: ['Scaling parameters required'] }
      
      default:
        return { isValid: false, errors: [`Unknown processing type: ${type}`] }
    }
  }

  /**
   * Get parameter constraints for a processing type
   */
  getParameterConstraints(type: ProcessingType) {
    switch (type) {
      case 'binarization':
        return BinarizationProcessor.getParameterConstraints()
      case 'morphology':
        return MorphologyProcessor.getParameterConstraints()
      case 'noise-reduction':
        return NoiseReductionProcessor.getParameterConstraints()
      case 'scaling':
        return ScalingProcessor.getParameterConstraints()
      default:
        return {}
    }
  }

  /**
   * Get recommended parameters for specific use cases
   */
  getRecommendedParameters(type: ProcessingType, useCase: string) {
    switch (type) {
      case 'binarization':
        return BinarizationProcessor.getRecommendedParameters(useCase)
      case 'morphology':
        return MorphologyProcessor.getRecommendedParameters(useCase)
      case 'noise-reduction':
        return NoiseReductionProcessor.getRecommendedParameters(useCase)
      case 'scaling':
        return ScalingProcessor.getRecommendedParameters(useCase)
      default:
        return {}
    }
  }

  /**
   * Auto-detect best processing parameters for image
   */
  autoDetectParameters(imageData: ImageData, type: ProcessingType): any {
    switch (type) {
      case 'noise-reduction':
        const noiseType = NoiseReductionProcessor.detectNoiseType(imageData)
        return NoiseReductionProcessor.getRecommendedParameters(noiseType)
      
      case 'scaling':
        // For scaling, we need a target factor - default to 2x
        return ScalingProcessor.autoDetectBestMethod(imageData, 2)
      
      default:
        return this.getRecommendedParameters(type, 'default')
    }
  }

  /**
   * Estimate processing time (rough heuristic)
   */
  estimateProcessingTime(imageData: ImageData, type: ProcessingType): number {
    const pixelCount = imageData.width * imageData.height
    const baseTime = pixelCount / 1000000 // Base time in seconds per megapixel
    
    const complexityMultipliers = {
      'binarization': 1,
      'morphology': 2,
      'noise-reduction': 3,
      'scaling': 1.5
    }
    
    const wasmSpeedup = this.vips ? 0.3 : 1 // WASM is ~3x faster
    
    return baseTime * complexityMultipliers[type] * wasmSpeedup
  }

  /**
   * Check if the module is initialized
   */
  isReady(): boolean {
    return this.isInitialized
  }

  /**
   * Check if WASM is available
   */
  hasWasmSupport(): boolean {
    return this.vips !== null
  }

  /**
   * Get module status information
   */
  getStatus() {
    return {
      initialized: this.isInitialized,
      hasWasm: this.hasWasmSupport(),
      availableAlgorithms: this.getAvailableAlgorithms(),
      version: '1.0.0'
    }
  }

  /**
   * Cleanup resources
   */
  destroy() {
    if (this.vips) {
      // VIPS cleanup if needed
      this.vips = null
    }
    this.isInitialized = false
    this.initializationPromise = null
  }
}