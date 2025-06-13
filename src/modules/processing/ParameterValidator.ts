// src/modules/processing/ParameterValidator.ts
// Validates processing parameters for different algorithms

import type { 
  ProcessingType, 
  ProcessingParameters,
  BinarizationParams,
  MorphologyParams,
  NoiseReductionParams,
  ScalingParams
} from '@/types'

export interface ValidationResult {
  isValid: boolean
  errors: string[]
  warnings: string[]
}

export class ParameterValidator {
  /**
   * Validate processing parameters based on type
   */
  static validate(type: ProcessingType, parameters: ProcessingParameters): ValidationResult {
    switch (type) {
      case 'binarization':
        return this.validateBinarization(parameters.binarization)
      case 'morphology':
        return this.validateMorphology(parameters.morphology)
      case 'noise-reduction':
        return this.validateNoiseReduction(parameters.noise)
      case 'scaling':
        return this.validateScaling(parameters.scaling)
      default:
        return {
          isValid: false,
          errors: [`Unknown processing type: ${type}`],
          warnings: []
        }
    }
  }

  /**
   * Validate binarization parameters
   */
  private static validateBinarization(params?: BinarizationParams): ValidationResult {
    const errors: string[] = []
    const warnings: string[] = []

    if (!params) {
      return {
        isValid: false,
        errors: ['Binarization parameters are required'],
        warnings: []
      }
    }

    // Validate method
    const validMethods = ['sauvola', 'niblack', 'otsu']
    if (!validMethods.includes(params.method)) {
      errors.push(`Invalid binarization method: ${params.method}. Valid methods: ${validMethods.join(', ')}`)
    }

    // Validate window size for adaptive methods
    if ((params.method === 'sauvola' || params.method === 'niblack') && params.windowSize !== undefined) {
      if (params.windowSize < 3 || params.windowSize > 51) {
        errors.push('Window size must be between 3 and 51')
      }
      if (params.windowSize % 2 === 0) {
        errors.push('Window size must be odd')
      }
    }

    // Validate k factor
    if (params.k !== undefined) {
      if (params.k < -1 || params.k > 1) {
        errors.push('K factor must be between -1 and 1')
      }
    }

    // Validate threshold
    if (params.threshold !== undefined) {
      if (params.threshold < 0 || params.threshold > 255) {
        errors.push('Threshold must be between 0 and 255')
      }
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    }
  }

  /**
   * Validate morphology parameters
   */
  private static validateMorphology(params?: MorphologyParams): ValidationResult {
    const errors: string[] = []
    const warnings: string[] = []

    if (!params) {
      return {
        isValid: false,
        errors: ['Morphology parameters are required'],
        warnings: []
      }
    }

    // Validate operation
    const validOperations = ['opening', 'closing', 'dilation', 'erosion']
    if (!validOperations.includes(params.operation)) {
      errors.push(`Invalid morphology operation: ${params.operation}. Valid operations: ${validOperations.join(', ')}`)
    }

    // Validate kernel size
    if (params.kernelSize < 3 || params.kernelSize > 15) {
      errors.push('Kernel size must be between 3 and 15')
    }
    if (params.kernelSize % 2 === 0) {
      errors.push('Kernel size must be odd')
    }

    // Validate iterations
    if (params.iterations < 1 || params.iterations > 5) {
      errors.push('Iterations must be between 1 and 5')
    }

    // Warnings
    if (params.iterations > 3) {
      warnings.push('High iteration count may significantly increase processing time')
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    }
  }

  /**
   * Validate noise reduction parameters
   */
  private static validateNoiseReduction(params?: NoiseReductionParams): ValidationResult {
    const errors: string[] = []
    const warnings: string[] = []

    if (!params) {
      return {
        isValid: false,
        errors: ['Noise reduction parameters are required'],
        warnings: []
      }
    }

    // Validate method
    const validMethods = ['median', 'binary-noise-removal']
    if (!validMethods.includes(params.method)) {
      errors.push(`Invalid noise reduction method: ${params.method}. Valid methods: ${validMethods.join(', ')}`)
    }

    // Validate kernel size for median filter
    if (params.method === 'median' && params.kernelSize !== undefined) {
      if (params.kernelSize < 3 || params.kernelSize > 9) {
        errors.push('Kernel size for median filter must be between 3 and 9')
      }
      if (params.kernelSize % 2 === 0) {
        errors.push('Kernel size must be odd')
      }
    }

    // Validate threshold for binary noise removal
    if (params.method === 'binary-noise-removal' && params.threshold !== undefined) {
      if (params.threshold < 1 || params.threshold > 1000) {
        errors.push('Threshold for binary noise removal must be between 1 and 1000')
      }
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    }
  }

  /**
   * Validate scaling parameters
   */
  private static validateScaling(params?: ScalingParams): ValidationResult {
    const errors: string[] = []
    const warnings: string[] = []

    if (!params) {
      return {
        isValid: false,
        errors: ['Scaling parameters are required'],
        warnings: []
      }
    }

    // Validate method
    const validMethods = ['scale2x', 'scale3x', 'scale4x', 'nearest', 'bilinear']
    if (!validMethods.includes(params.method)) {
      errors.push(`Invalid scaling method: ${params.method}. Valid methods: ${validMethods.join(', ')}`)
    }

    // Validate factor
    if (params.factor < 0.1 || params.factor > 8) {
      errors.push('Scale factor must be between 0.1 and 8')
    }

    // Special validation for pixel art methods
    if (params.method.startsWith('scale') && !Number.isInteger(params.factor)) {
      errors.push('Pixel art scaling methods require integer scale factors')
    }

    // Warnings
    if (params.factor > 4) {
      warnings.push('Large scale factors may require significant memory and processing time')
    }

    if (params.factor < 1 && params.method !== 'bilinear' && params.method !== 'nearest') {
      warnings.push('Downscaling with pixel art methods may produce poor results')
    }

    return {
      isValid: errors.length === 0,
      errors,
      warnings
    }
  }

  /**
   * Get parameter constraints for a processing type
   */
  static getConstraints(type: ProcessingType) {
    switch (type) {
      case 'binarization':
        return {
          method: ['sauvola', 'niblack', 'otsu'],
          windowSize: { min: 3, max: 51, step: 2 },
          k: { min: -1, max: 1, step: 0.01 },
          threshold: { min: 0, max: 255, step: 1 }
        }
      case 'morphology':
        return {
          operation: ['opening', 'closing', 'dilation', 'erosion'],
          kernelSize: { min: 3, max: 15, step: 2 },
          iterations: { min: 1, max: 5, step: 1 }
        }
      case 'noise-reduction':
        return {
          method: ['median', 'binary-noise-removal'],
          kernelSize: { min: 3, max: 9, step: 2 },
          threshold: { min: 1, max: 1000, step: 10 }
        }
      case 'scaling':
        return {
          method: ['scale2x', 'scale3x', 'scale4x', 'nearest', 'bilinear'],
          factor: { min: 0.1, max: 8, step: 0.1 }
        }
      default:
        return {}
    }
  }
}