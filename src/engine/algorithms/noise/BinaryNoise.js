/**
 * @fileoverview Binary noise reduction for removing isolated pixels
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { NoiseReducer } from './NoiseReducer.js';

export class BinaryNoise extends NoiseReducer {
  constructor(options = {}) {
    super();
    this.threshold = options.threshold || 4;
  }

  process(imageData) {
    const src = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const result = src.clone();
    
    for (let y = 1; y < src.height - 1; y++) {
      for (let x = 1; x < src.width - 1; x++) {
        const pixel = src.getPixel(x, y);
        const targetValue = pixel > 127 ? 255 : 0;
        
        // Count neighbors with same value
        let sameCount = 0;
        for (let dy = -1; dy <= 1; dy++) {
          for (let dx = -1; dx <= 1; dx++) {
            if (dx === 0 && dy === 0) continue;
            
            const neighbor = src.getPixel(x + dx, y + dy);
            const neighborValue = neighbor > 127 ? 255 : 0;
            
            if (neighborValue === targetValue) {
              sameCount++;
            }
          }
        }
        
        // If too few neighbors have same value, flip the pixel
        if (sameCount < this.threshold) {
          result.setPixel(x, y, targetValue === 255 ? 0 : 255);
        }
      }
    }
    
    return result;
  }

  getParameters() {
    return {
      threshold: this.threshold
    };
  }

  setParameters(params) {
    if (params.threshold !== undefined) {
      this.threshold = Math.max(1, Math.min(8, params.threshold));
    }
  }
}