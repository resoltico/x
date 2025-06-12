import type { ProcessingType, ProcessingParameters } from '@/types'

/**
 * Pre-configured processing presets for common use cases
 */

export interface ProcessingPreset {
  name: string
  description: string
  type: ProcessingType
  parameters: ProcessingParameters
  icon: string
}

export const PROCESSING_PRESETS: Record<string, ProcessingPreset> = {
  document: {
    name: 'Document',
    description: 'Optimized for scanned documents with varying lighting conditions',
    type: 'binarization',
    icon: '📄',
    parameters: {
      binarization: {
        method: 'sauvola',
        windowSize: 15,
        k: 0.2,
        threshold: 128
      }
    }
  },

  engraving: {
    name: 'Engraving',
    description: 'Best for historical engravings and artwork with clear contrast',
    type: 'binarization',
    icon: '🖼️',
    parameters: {
      binarization: {
        method: 'otsu',
        windowSize: 15,
        k: 0.2,
        threshold: 128
      }
    }
  },

  'pixel-art': {
    name: 'Pixel Art',
    description: 'Scale pixel art images while preserving sharp edges',
    type: 'scaling',
    icon: '🎮',
    parameters: {
      scaling: {
        method: 'scale2x',
        factor: 2
      }
    }
  },

  'noise-clean': {
    name: 'Noise Clean',
    description: 'Remove salt-and-pepper noise from scanned images',
    type: 'noise-reduction',
    icon: '✨',
    parameters: {
      noise: {
        method: 'median',
        kernelSize: 3,
        threshold: 50
      }
    }
  },

  'morphology-clean': {
    name: 'Morphology Clean',
    description: 'Clean up binary images by removing small artifacts',
    type: 'morphology',
    icon: '🔧',
    parameters: {
      morphology: {
        operation: 'opening',
        kernelSize: 3,
        iterations: 1
      }
    }
  },

  'enhance-text': {
    name: 'Enhance Text',
    description: 'Multi-stage processing for text documents',
    type: 'binarization',
    icon: '📖',
    parameters: {
      binarization: {
        method: 'niblack',
        windowSize: 21,
        k: -0.2,
        threshold: 128
      }
    }
  },

  'high-quality-scale': {
    name: 'High Quality Scale',
    description: 'Smooth scaling for photographs and artwork',
    type: 'scaling',
    icon: '🔍',
    parameters: {
      scaling: {
        method: 'bilinear',
        factor: 2.0
      }
    }
  },

  'heavy-noise-reduction': {
    name: 'Heavy Noise Reduction',
    description: 'Aggressive noise removal for heavily corrupted images',
    type: 'noise-reduction',
    icon: '💥',
    parameters: {
      noise: {
        method: 'median',
        kernelSize: 5,
        threshold: 100
      }
    }
  }
}

/**
 * Get all available presets
 */
export const getAllPresets = (): ProcessingPreset[] => {
  return Object.values(PROCESSING_PRESETS)
}

/**
 * Get presets by processing type
 */
export const getPresetsByType = (type: ProcessingType): ProcessingPreset[] => {
  return Object.values(PROCESSING_PRESETS).filter(preset => preset.type === type)
}

/**
 * Get a specific preset by name
 */
export const getPreset = (name: string): ProcessingPreset | undefined => {
  return PROCESSING_PRESETS[name]
}