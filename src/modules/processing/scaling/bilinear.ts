// src/modules/processing/scaling/bilinear.ts
import { ScalingUtils } from './utils'

/**
 * Bilinear scaling implementation
 */
export class BilinearScaler {
  /**
   * Scale canvas using bilinear interpolation
   */
  static async scale(canvas: OffscreenCanvas, factor: number): Promise<ArrayBuffer> {
    const scaledCanvas = new OffscreenCanvas(
      Math.round(canvas.width * factor),
      Math.round(canvas.height * factor)
    )
    const scaledCtx = scaledCanvas.getContext('2d')!
    
    scaledCtx.imageSmoothingEnabled = true
    scaledCtx.imageSmoothingQuality = 'high'
    scaledCtx.drawImage(
      canvas,
      0, 0, canvas.width, canvas.height,
      0, 0, scaledCanvas.width, scaledCanvas.height
    )
    
    return ScalingUtils.canvasToArrayBuffer(scaledCanvas)
  }
}