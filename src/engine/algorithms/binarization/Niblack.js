/**
 * @fileoverview Niblack adaptive binarization algorithm
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { IntegralImage } from '../../core/IntegralImage.js';
import { BaseBinarizer } from './BaseBinarizer.js';

export class Niblack extends BaseBinarizer {
  constructor(options = {}) {
    super();
    this.windowSize = options.windowSize || 15;
    this.k = options.k || -0.2;
  }

  process(imageData) {
    const grayImage = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const integral = new IntegralImage(grayImage);
    const result = ImageData.createEmpty(grayImage.width, grayImage.height, 1);
    
    const halfWindow = Math.floor(this.windowSize / 2);
    
    for (let y = 0; y < grayImage.height; y++) {
      for (let x = 0; x < grayImage.width; x++) {
        // Define window boundaries
        const x1 = Math.max(0, x - halfWindow);
        const y1 = Math.max(0, y - halfWindow);
        const x2 = Math.min(grayImage.width - 1, x + halfWindow);
        const y2 = Math.min(grayImage.height - 1, y + halfWindow);
        
        // Get local statistics
        const mean = integral.getMean(x1, y1, x2, y2);
        const stdDev = integral.getStdDev(x1, y1, x2, y2);
        
        // Niblack threshold formula: T = mean + k * stdDev
        const threshold = mean + this.k * stdDev;
        
        // Apply threshold
        const pixel = grayImage.getPixel(x, y);
        result.setPixel(x, y, pixel > threshold ? 255 : 0);
      }
    }
    
    return result;
  }

  getParameters() {
    return {
      windowSize: this.windowSize,
      k: this.k
    };
  }

  setParameters(params) {
    if (params.windowSize !== undefined) {
      this.windowSize = Math.max(3, Math.min(51, params.windowSize));
      // Ensure odd window size
      if (this.windowSize % 2 === 0) this.windowSize++;
    }
    if (params.k !== undefined) {
      this.k = Math.max(-1.0, Math.min(0.5, params.k));
    }
  }
}