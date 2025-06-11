/**
 * @fileoverview Nearest neighbor scaling algorithm
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { BaseScaler } from './BaseScaler.js';

export class NearestNeighbor extends BaseScaler {
  constructor(scaleFactor = 2) {
    super();
    this.scaleFactor = scaleFactor;
  }

  process(imageData) {
    const src = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const newWidth = Math.round(src.width * this.scaleFactor);
    const newHeight = Math.round(src.height * this.scaleFactor);
    const result = ImageData.createEmpty(newWidth, newHeight, 1);

    for (let y = 0; y < newHeight; y++) {
      for (let x = 0; x < newWidth; x++) {
        // Find corresponding source pixel
        const srcX = Math.floor(x / this.scaleFactor);
        const srcY = Math.floor(y / this.scaleFactor);
        
        const value = src.getPixel(srcX, srcY);
        result.setPixel(x, y, value);
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