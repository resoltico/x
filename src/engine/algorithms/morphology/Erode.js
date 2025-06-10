/**
 * @fileoverview Morphological erosion operation
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { MorphologyBase } from './MorphologyBase.js';

export class Erode extends MorphologyBase {
  processSingle(imageData) {
    const src = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const result = ImageData.createEmpty(src.width, src.height, 1);
    const kernel = this.getKernel();
    const halfKernel = Math.floor(this.kernelSize / 2);
    
    for (let y = 0; y < src.height; y++) {
      for (let x = 0; x < src.width; x++) {
        let minValue = 255;
        
        // Apply kernel
        for (let ky = 0; ky < this.kernelSize; ky++) {
          for (let kx = 0; kx < this.kernelSize; kx++) {
            if (kernel[ky][kx] === 0) continue;
            
            const px = x + kx - halfKernel;
            const py = y + ky - halfKernel;
            
            if (px >= 0 && px < src.width && py >= 0 && py < src.height) {
              const value = src.getPixel(px, py);
              minValue = Math.min(minValue, value);
            } else {
              // Treat out-of-bounds as white (255) for erosion
              minValue = Math.min(minValue, 255);
            }
          }
        }
        
        result.setPixel(x, y, minValue);
      }
    }
    
    return result;
  }

  process(imageData) {
    let result = imageData;
    for (let i = 0; i < this.iterations; i++) {
      result = this.processSingle(result);
    }
    return result;
  }
}