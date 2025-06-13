// src/modules/processing/scaling/index.ts
import type { ScalingParams, ImageData } from '@/types'
import { NearestNeighborScaler } from './nearest-neighbor'
import { BilinearScaler } from './bilinear'
import { PixelArtScaler } from './pixel-art'
import { LanczosScaler } from './lanczos'
import { ScalingValidator } from './validator'
import { ScalingUtils } from './utils'

/**
 * Main scaling processor that delegates to specific scaling algorithms
 */
export class ScalingProcessor {
  /**
   * Process image with specified scaling method
   */
  static async process(
    imageData: ImageData,
    params: ScalingParams,
    vipsImage?: any
  ): Promise<ArrayBuffer> {
    const { method, factor } = params

    // Validate parameters first
    const validation = ScalingValidator.validateParameters(params)
    if (!validation.isValid) {
      throw new Error(`Invalid scaling parameters: ${validation.errors.join(', ')}`)
    }

    // Check feasibility
    const feasibility = ScalingValidator.checkFeasibility(
      imageData.width,
      imageData.height,
      params
    )
    if (!feasibility.feasible) {
      throw new Error(`Scaling not feasible: ${feasibility.reason}`)
    }

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
  ): Promise<ArrayBuffer> {
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
        return NearestNeighborScaler.scale(canvas, factor)
      case 'bilinear':
        return BilinearScaler.scale(canvas, factor)
      case 'lanczos':
        return LanczosScaler.scale(canvas, factor)
      case 'scale2x':
        return PixelArtScaler.scale2x(canvas)
      case 'scale3x':
        return PixelArtScaler.scale3x(canvas)
      case 'scale4x':
        return PixelArtScaler.scale4x(canvas)
      default:
        return BilinearScaler.scale(canvas, factor)
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
   * Get parameter constraints for scaling
   */
  static getParameterConstraints() {
    return ScalingValidator.getParameterConstraints()
  }

  /**
   * Get recommended parameters for different use cases
   */
  static getRecommendedParameters(useCase: string) {
    return ScalingValidator.getRecommendedParameters(useCase)
  }

  /**
   * Validate scaling parameters
   */
  static validateParameters(params: ScalingParams) {
    return ScalingValidator.validateParameters(params)
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
  ) {
    return ScalingUtils.calculateOutputDimensions(inputWidth, inputHeight, params)
  }

  /**
   * Estimate memory usage for scaling operation
   */
  static estimateMemoryUsage(
    inputWidth: number,
    inputHeight: number,
    params: ScalingParams
  ) {
    return ScalingUtils.estimateMemoryUsage(inputWidth, inputHeight, params)
  }

  /**
   * Auto-detect best scaling method for image content
   */
  static autoDetectBestMethod(imageData: ImageData, targetFactor: number) {
    return ScalingUtils.autoDetectBestMethod(imageData, targetFactor)
  }

  /**
   * Get quality comparison between methods
   */
  static getMethodQualityComparison() {
    return ScalingUtils.getMethodQualityComparison()
  }

  /**
   * Check if scaling operation is feasible
   */
  static checkFeasibility(
    inputWidth: number,
    inputHeight: number,
    params: ScalingParams,
    maxMemoryMB: number = 500
  ) {
    return ScalingValidator.checkFeasibility(inputWidth, inputHeight, params, maxMemoryMB)
  }
}