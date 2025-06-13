// src/modules/processing/scaling/utils.ts
import type { ScalingParams, ImageData } from '@/types'

/**
 * Utility functions for scaling operations
 */
export class ScalingUtils {
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
    const colorVariations = new Set<string>()
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

  /**
   * Get pixel from image data
   */
  static getPixel(
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
  static setPixel(
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
  static pixelsEqual(
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
   * Convert canvas to ArrayBuffer
   */
  static async canvasToArrayBuffer(canvas: OffscreenCanvas): Promise<ArrayBuffer> {
    const blob = await canvas.convertToBlob({ type: 'image/png' })
    return await blob.arrayBuffer()
  }
}