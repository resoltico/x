// src/modules/processing/scaling/nearest-neighbor.ts
import { ScalingUtils } from './utils'

/**
 * Nearest neighbor scaling implementation
 */
export class NearestNeighborScaler {
  /**
   * Scale canvas using nearest neighbor interpolation
   */
  static async scale(canvas: OffscreenCanvas, factor: number): Promise<ArrayBuffer> {
    const scaledCanvas = new OffscreenCanvas(
      Math.round(canvas.width * factor),
      Math.round(canvas.height * factor)
    )
    const scaledCtx = scaledCanvas.getContext('2d')!
    
    scaledCtx.imageSmoothingEnabled = false
    scaledCtx.drawImage(
      canvas,
      0, 0, canvas.width, canvas.height,
      0, 0, scaledCanvas.width, scaledCanvas.height
    )
    
    return ScalingUtils.canvasToArrayBuffer(scaledCanvas)
  }
}