/**
 * @fileoverview Bilinear interpolation scaling algorithm
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { BaseScaler } from './BaseScaler.js';

export class Bilinear extends BaseScaler {
  constructor(scaleFactor = 2) {
    super();
    this.scaleFactor = scaleFactor;
  }

  process(imageData) {
    const src = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const newWidth = Math.round(src.width * this.scaleFactor);
    const newHeight = Math.round(src.height * this.scaleFactor);
    const result = ImageData.createEmpty(newWidth, newHeight, 1);

    const xRatio = src.width / newWidth;
    const yRatio = src.height / newHeight;

    for (let y = 0; y < newHeight; y++) {
      for (let x = 0; x < newWidth; x++) {
        // Calculate source coordinates
        const srcX = x * xRatio;
        const srcY = y * yRatio;
        
        // Get integer and fractional parts
        const x0 = Math.floor(srcX);
        const y0 = Math.floor(srcY);
        const x1 = Math.min(x0 + 1, src.width - 1);
        const y1 = Math.min(y0 + 1, src.height - 1);
        
        const fx = srcX - x0;
        const fy = srcY - y0;
        
        // Get pixel values
        const p00 = src.getPixel(x0, y0) || 0;
        const p10 = src.getPixel(x1, y0) || 0;
        const p01 = src.getPixel(x0, y1) || 0;
        const p11 = src.getPixel(x1, y1) || 0;
        
        // Bilinear interpolation
        const value = Math.round(
          p00 * (1 - fx) * (1 - fy) +
          p10 * fx * (1 - fy) +
          p01 * (1 - fx) * fy +
          p11 * fx * fy
        );
        
        result.setPixel(x, y, Math.max(0, Math.min(255, value)));
      }
    }

    return result;
  }

  getScaleFactor() {
    return this.scaleFactor;
  }

  setScaleFactor(factor) {
    this.scaleFactor = factor;
  }
}