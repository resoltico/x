// src/modules/processing/scaling/lanczos.ts
import { ScalingUtils } from './utils'

/**
 * Lanczos scaling implementation
 */
export class LanczosScaler {
  /**
   * Scale canvas using Lanczos resampling
   */
  static async scale(canvas: OffscreenCanvas, factor: number): Promise<ArrayBuffer> {
    const ctx = canvas.getContext('2d')!
    const srcData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const srcPixels = srcData.data
    
    const newWidth = Math.round(canvas.width * factor)
    const newHeight = Math.round(canvas.height * factor)
    
    const scaledCanvas = new OffscreenCanvas(newWidth, newHeight)
    const scaledCtx = scaledCanvas.getContext('2d')!
    const dstData = scaledCtx.createImageData(newWidth, newHeight)
    const dstPixels = dstData.data
    
    // Lanczos-3 kernel
    const lanczos = (x: number): number => {
      if (x === 0) return 1
      if (Math.abs(x) >= 3) return 0
      
      const piX = Math.PI * x
      return (3 * Math.sin(piX) * Math.sin(piX / 3)) / (piX * piX)
    }
    
    const scaleX = canvas.width / newWidth
    const scaleY = canvas.height / newHeight
    
    for (let dstY = 0; dstY < newHeight; dstY++) {
      for (let dstX = 0; dstX < newWidth; dstX++) {
        const srcX = (dstX + 0.5) * scaleX - 0.5
        const srcY = (dstY + 0.5) * scaleY - 0.5
        
        let r = 0, g = 0, b = 0, a = 0, weightSum = 0
        
        // Sample 6x6 neighborhood
        for (let sy = Math.floor(srcY) - 2; sy <= Math.floor(srcY) + 3; sy++) {
          for (let sx = Math.floor(srcX) - 2; sx <= Math.floor(srcX) + 3; sx++) {
            if (sx >= 0 && sx < canvas.width && sy >= 0 && sy < canvas.height) {
              const weight = lanczos(sx - srcX) * lanczos(sy - srcY)
              const idx = (sy * canvas.width + sx) * 4
              
              r += srcPixels[idx] * weight
              g += srcPixels[idx + 1] * weight
              b += srcPixels[idx + 2] * weight
              a += srcPixels[idx + 3] * weight
              weightSum += weight
            }
          }
        }
        
        const dstIdx = (dstY * newWidth + dstX) * 4
        if (weightSum > 0) {
          dstPixels[dstIdx] = Math.round(r / weightSum)
          dstPixels[dstIdx + 1] = Math.round(g / weightSum)
          dstPixels[dstIdx + 2] = Math.round(b / weightSum)
          dstPixels[dstIdx + 3] = Math.round(a / weightSum)
        }
      }
    }
    
    scaledCtx.putImageData(dstData, 0, 0)
    return ScalingUtils.canvasToArrayBuffer(scaledCanvas)
  }
}