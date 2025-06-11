/**
 * @fileoverview Otsu's method for automatic threshold selection
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { BaseBinarizer } from './BaseBinarizer.js';

export class Otsu extends BaseBinarizer {
  constructor(options = {}) {
    super();
    // Otsu doesn't use window-based processing, but we keep these for consistency
    this.windowSize = options.windowSize || 15;
    this.k = options.k || 0.5;
    this.r = options.r || 128;
  }

  process(imageData) {
    const grayImage = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const result = ImageData.createEmpty(grayImage.width, grayImage.height, 1);
    
    // Calculate histogram
    const histogram = new Array(256).fill(0);
    const totalPixels = grayImage.width * grayImage.height;
    
    for (let i = 0; i < grayImage.data.length; i++) {
      histogram[grayImage.data[i]]++;
    }
    
    // Calculate optimal threshold using Otsu's method
    const threshold = this.calculateOtsuThreshold(histogram, totalPixels);
    
    // Apply threshold
    for (let i = 0; i < grayImage.data.length; i++) {
      result.data[i] = grayImage.data[i] > threshold ? 255 : 0;
    }
    
    return result;
  }

  calculateOtsuThreshold(histogram, totalPixels) {
    // Calculate probabilities
    const probabilities = histogram.map(count => count / totalPixels);
    
    // Calculate cumulative sums
    let omega = 0;
    let mu = 0;
    const omega0 = new Array(256);
    const mu0 = new Array(256);
    
    for (let i = 0; i < 256; i++) {
      omega += probabilities[i];
      mu += i * probabilities[i];
      omega0[i] = omega;
      mu0[i] = mu;
    }
    
    // Calculate between-class variance for each threshold
    let maxVariance = 0;
    let bestThreshold = 0;
    
    for (let t = 0; t < 256; t++) {
      const omega1 = omega0[t];
      const omega2 = 1 - omega1;
      
      if (omega1 === 0 || omega2 === 0) continue;
      
      const mu1 = mu0[t] / omega1;
      const mu2 = (mu - mu0[t]) / omega2;
      
      const variance = omega1 * omega2 * Math.pow(mu1 - mu2, 2);
      
      if (variance > maxVariance) {
        maxVariance = variance;
        bestThreshold = t;
      }
    }
    
    return bestThreshold;
  }

  getParameters() {
    return {
      windowSize: this.windowSize,
      k: this.k,
      r: this.r
    };
  }

  setParameters(params) {
    // Otsu doesn't actually use these parameters, but we maintain them for UI consistency
    if (params.windowSize !== undefined) {
      this.windowSize = Math.max(3, Math.min(51, params.windowSize));
      if (this.windowSize % 2 === 0) this.windowSize++;
    }
    if (params.k !== undefined) {
      this.k = Math.max(0.1, Math.min(1.0, params.k));
    }
    if (params.r !== undefined) {
      this.r = Math.max(1, Math.min(255, params.r));
    }
  }

  static getParameterMetadata() {
    return {
      windowSize: {
        min: 3,
        max: 51,
        default: 15,
        description: 'Not used in Otsu method',
        enabled: false
      },
      k: {
        min: 0.1,
        max: 1.0,
        default: 0.5,
        description: 'Not used in Otsu method',
        enabled: false
      },
      r: {
        min: 1,
        max: 255,
        default: 128,
        description: 'Not used in Otsu method',
        enabled: false
      }
    };
  }
}