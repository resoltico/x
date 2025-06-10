/**
 * @fileoverview Scale2x pixel art scaling algorithm
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { BaseScaler } from './BaseScaler.js';

export class Scale2x extends BaseScaler {
  constructor() {
    super();
    this.scaleFactor = 2;
  }

  process(imageData) {
    const src = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const newWidth = src.width * 2;
    const newHeight = src.height * 2;
    const result = ImageData.createEmpty(newWidth, newHeight, 1);

    for (let y = 0; y < src.height; y++) {
      for (let x = 0; x < src.width; x++) {
        // Get neighboring pixels
        const p = src.getPixel(x, y);
        const a = y > 0 ? src.getPixel(x, y - 1) : p;
        const b = x < src.width - 1 ? src.getPixel(x + 1, y) : p;
        const c = x > 0 ? src.getPixel(x - 1, y) : p;
        const d = y < src.height - 1 ? src.getPixel(x, y + 1) : p;

        // Scale2x rules
        let e0, e1, e2, e3;

        if (c === a && c !== d && a !== b) {
          e0 = a;
        } else {
          e0 = p;
        }

        if (a === b && a !== c && b !== d) {
          e1 = b;
        } else {
          e1 = p;
        }

        if (c === d && c !== b && d !== a) {
          e2 = c;
        } else {
          e2 = p;
        }

        if (b === d && b !== a && d !== c) {
          e3 = d;
        } else {
          e3 = p;
        }

        // Set output pixels
        const outX = x * 2;
        const outY = y * 2;
        result.setPixel(outX, outY, e0);
        result.setPixel(outX + 1, outY, e1);
        result.setPixel(outX, outY + 1, e2);
        result.setPixel(outX + 1, outY + 1, e3);
      }
    }

    return result;
  }

  getScaleFactor() {
    return 2;
  }
}