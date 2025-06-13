// src/modules/processing/scaling/validator.ts
import type { ScalingParams } from '@/types'
import { ScalingUtils } from './utils'

/**
 * Parameter validation for scaling operations
 */
export class ScalingValidator {
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
   * Check if scaling operation is feasible
   */
  static checkFeasibility(
    inputWidth: number,
    inputHeight: number,
    params: ScalingParams,
    maxMemoryMB: number = 500
  ): { feasible: boolean; reason?: string; suggestion?: string } {
    const outputDims = ScalingUtils.calculateOutputDimensions(inputWidth, inputHeight, params)
    const memoryUsage = ScalingUtils.estimateMemoryUsage(inputWidth, inputHeight, params)
    
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
}