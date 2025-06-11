/**
 * @fileoverview Median filter for noise reduction
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { NoiseReducer } from './NoiseReducer.js';

export class MedianFilter extends NoiseReducer {
  constructor(options = {}) {
    super();
    this.windowSize = options.windowSize || 3;
  }

  process(imageData) {
    const src = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const result = ImageData.createEmpty(src.width, src.height, 1);
    const halfWindow = Math.floor(this.windowSize / 2);
    
    for (let y = 0; y < src.height; y++) {
      for (let x = 0; x < src.width; x++) {
        const values = [];
        
        // Collect neighborhood values
        for (let dy = -halfWindow; dy <= halfWindow; dy++) {
          for (let dx = -halfWindow; dx <= halfWindow; dx++) {
            const ny = y + dy;
            const nx = x + dx;
            
            if (ny >= 0 && ny < src.height && nx >= 0 && nx < src.width) {
              values.push(src.getPixel(nx, ny));
            }
          }
        }
        
        // Sort values and find median
        values.sort((a, b) => a - b);
        const median = values[Math.floor(values.length / 2)];
        
        result.setPixel(x, y, median);
      }
    }
    
    return result;
  }

  getParameters() {
    return {
      windowSize: this.windowSize
    };
  }

  setParameters(params) {
    if (params.windowSize !== undefined) {
      this.windowSize = Math.max(3, Math.min(7, params.windowSize));
      // Ensure odd window size
      if (this.windowSize % 2 === 0) this.windowSize++;
    }
  }
}