// src/modules/processing/scaling/pixel-art.ts
import { ScalingUtils } from './utils'

/**
 * Pixel art scaling algorithms (Scale2x, Scale3x, Scale4x)
 */
export class PixelArtScaler {
  /**
   * Scale2x algorithm implementation
   */
  static async scale2x(canvas: OffscreenCanvas): Promise<ArrayBuffer> {
    const ctx = canvas.getContext('2d')!
    const srcData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const srcPixels = srcData.data
    
    const scaledCanvas = new OffscreenCanvas(canvas.width * 2, canvas.height * 2)
    const scaledCtx = scaledCanvas.getContext('2d')!
    const dstData = scaledCtx.createImageData(canvas.width * 2, canvas.height * 2)
    const dstPixels = dstData.data
    
    for (let y = 0; y < canvas.height; y++) {
      for (let x = 0; x < canvas.width; x++) {
        // Get neighboring pixels
        const A = ScalingUtils.getPixel(srcPixels, x, y - 1, canvas.width, canvas.height)
        const B = ScalingUtils.getPixel(srcPixels, x + 1, y, canvas.width, canvas.height)
        const C = ScalingUtils.getPixel(srcPixels, x, y, canvas.width, canvas.height) // Center pixel
        const D = ScalingUtils.getPixel(srcPixels, x - 1, y, canvas.width, canvas.height)
        const E = ScalingUtils.getPixel(srcPixels, x, y + 1, canvas.width, canvas.height)
        
        // Scale2x algorithm
        let E0 = C, E1 = C, E2 = C, E3 = C
        
        if (!ScalingUtils.pixelsEqual(D, B) && !ScalingUtils.pixelsEqual(A, E)) {
          E0 = ScalingUtils.pixelsEqual(D, A) ? D : C
          E1 = ScalingUtils.pixelsEqual(A, B) ? B : C
          E2 = ScalingUtils.pixelsEqual(D, E) ? D : C
          E3 = ScalingUtils.pixelsEqual(E, B) ? B : C
        }
        
        // Write to destination
        ScalingUtils.setPixel(dstPixels, x * 2, y * 2, E0, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 2 + 1, y * 2, E1, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 2, y * 2 + 1, E2, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 2 + 1, y * 2 + 1, E3, scaledCanvas.width)
      }
    }
    
    scaledCtx.putImageData(dstData, 0, 0)
    return ScalingUtils.canvasToArrayBuffer(scaledCanvas)
  }

  /**
   * Scale3x algorithm implementation
   */
  static async scale3x(canvas: OffscreenCanvas): Promise<ArrayBuffer> {
    const ctx = canvas.getContext('2d')!
    const srcData = ctx.getImageData(0, 0, canvas.width, canvas.height)
    const srcPixels = srcData.data
    
    const scaledCanvas = new OffscreenCanvas(canvas.width * 3, canvas.height * 3)
    const scaledCtx = scaledCanvas.getContext('2d')!
    const dstData = scaledCtx.createImageData(canvas.width * 3, canvas.height * 3)
    const dstPixels = dstData.data
    
    for (let y = 0; y < canvas.height; y++) {
      for (let x = 0; x < canvas.width; x++) {
        // Get 3x3 neighborhood
        const A = ScalingUtils.getPixel(srcPixels, x - 1, y - 1, canvas.width, canvas.height)
        const B = ScalingUtils.getPixel(srcPixels, x, y - 1, canvas.width, canvas.height)
        const C = ScalingUtils.getPixel(srcPixels, x + 1, y - 1, canvas.width, canvas.height)
        const D = ScalingUtils.getPixel(srcPixels, x - 1, y, canvas.width, canvas.height)
        const E = ScalingUtils.getPixel(srcPixels, x, y, canvas.width, canvas.height) // Center
        const F = ScalingUtils.getPixel(srcPixels, x + 1, y, canvas.width, canvas.height)
        const G = ScalingUtils.getPixel(srcPixels, x - 1, y + 1, canvas.width, canvas.height)
        const H = ScalingUtils.getPixel(srcPixels, x, y + 1, canvas.width, canvas.height)
        const I = ScalingUtils.getPixel(srcPixels, x + 1, y + 1, canvas.width, canvas.height)
        
        // Scale3x algorithm - simplified version
        let E0 = E, E1 = E, E2 = E
        let E3 = E, E4 = E, E5 = E
        let E6 = E, E7 = E, E8 = E
        
        if (!ScalingUtils.pixelsEqual(D, F) && !ScalingUtils.pixelsEqual(B, H)) {
          E0 = ScalingUtils.pixelsEqual(D, B) ? D : E
          E1 = (ScalingUtils.pixelsEqual(D, B) && !ScalingUtils.pixelsEqual(E, C)) || 
               (ScalingUtils.pixelsEqual(B, F) && !ScalingUtils.pixelsEqual(E, A)) ? B : E
          E2 = ScalingUtils.pixelsEqual(B, F) ? F : E
          E3 = (ScalingUtils.pixelsEqual(D, B) && !ScalingUtils.pixelsEqual(E, G)) || 
               (ScalingUtils.pixelsEqual(D, H) && !ScalingUtils.pixelsEqual(E, A)) ? D : E
          E4 = E
          E5 = (ScalingUtils.pixelsEqual(B, F) && !ScalingUtils.pixelsEqual(E, I)) || 
               (ScalingUtils.pixelsEqual(F, H) && !ScalingUtils.pixelsEqual(E, C)) ? F : E
          E6 = ScalingUtils.pixelsEqual(D, H) ? D : E
          E7 = (ScalingUtils.pixelsEqual(D, H) && !ScalingUtils.pixelsEqual(E, I)) || 
               (ScalingUtils.pixelsEqual(H, F) && !ScalingUtils.pixelsEqual(E, G)) ? H : E
          E8 = ScalingUtils.pixelsEqual(H, F) ? F : E
        }
        
        // Write 3x3 block to destination
        ScalingUtils.setPixel(dstPixels, x * 3, y * 3, E0, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 3 + 1, y * 3, E1, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 3 + 2, y * 3, E2, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 3, y * 3 + 1, E3, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 3 + 1, y * 3 + 1, E4, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 3 + 2, y * 3 + 1, E5, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 3, y * 3 + 2, E6, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 3 + 1, y * 3 + 2, E7, scaledCanvas.width)
        ScalingUtils.setPixel(dstPixels, x * 3 + 2, y * 3 + 2, E8, scaledCanvas.width)
      }
    }
    
    scaledCtx.putImageData(dstData, 0, 0)
    return ScalingUtils.canvasToArrayBuffer(scaledCanvas)
  }

  /**
   * Scale4x algorithm implementation (Scale2x applied twice)
   */
  static async scale4x(canvas: OffscreenCanvas): Promise<ArrayBuffer> {
    // Apply Scale2x twice
    const intermediate = await this.scale2x(canvas)
    
    // Create intermediate canvas from the result
    const blob = new Blob([intermediate], { type: 'image/png' })
    
    // Use createImageBitmap for worker compatibility
    const imageBitmap = await createImageBitmap(blob)
    
    const intermediateCanvas = new OffscreenCanvas(canvas.width * 2, canvas.height * 2)
    const intermediateCtx = intermediateCanvas.getContext('2d')!
    
    intermediateCtx.drawImage(imageBitmap, 0, 0)
    
    // Apply Scale2x again
    const result = await this.scale2x(intermediateCanvas)
    
    // Clean up the image bitmap
    imageBitmap.close()
    
    return result
  }
}