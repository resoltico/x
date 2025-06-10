/**
 * @fileoverview Scale3x pixel art scaling algorithm
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { ImageData } from '../../core/ImageData.js';
import { BaseScaler } from './BaseScaler.js';

export class Scale3x extends BaseScaler {
  constructor() {
    super();
    this.scaleFactor = 3;
  }

  process(imageData) {
    const src = imageData.channels === 1 ? imageData : imageData.toGrayscale();
    const newWidth = src.width * 3;
    const newHeight = src.height * 3;
    const result = ImageData.createEmpty(newWidth, newHeight, 1);

    for (let y = 0; y < src.height; y++) {
      for (let x = 0; x < src.width; x++) {
        // Get neighboring pixels (using array indices for clarity)
        //  A B C
        //  D E F
        //  G H I
        const e = src.getPixel(x, y); // center
        const a = (y > 0 && x > 0) ? src.getPixel(x - 1, y - 1) : e;
        const b = y > 0 ? src.getPixel(x, y - 1) : e;
        const c = (y > 0 && x < src.width - 1) ? src.getPixel(x + 1, y - 1) : e;
        const d = x > 0 ? src.getPixel(x - 1, y) : e;
        const f = x < src.width - 1 ? src.getPixel(x + 1, y) : e;
        const g = (y < src.height - 1 && x > 0) ? src.getPixel(x - 1, y + 1) : e;
        const h = y < src.height - 1 ? src.getPixel(x, y + 1) : e;
        const i = (y < src.height - 1 && x < src.width - 1) ? src.getPixel(x + 1, y + 1) : e;

        // Apply Scale3x rules
        // Output pixels:
        // E0 E1 E2
        // E3 E4 E5
        // E6 E7 E8
        let e0, e1, e2, e3, e4, e5, e6, e7, e8;

        if (b !== h && d !== f) {
          e0 = d === b ? d : e;
          e1 = (d === b && e !== c) || (b === f && e !== a) ? b : e;
          e2 = b === f ? f : e;
          e3 = (d === b && e !== g) || (d === h && e !== a) ? d : e;
          e4 = e;
          e5 = (b === f && e !== i) || (h === f && e !== c) ? f : e;
          e6 = d === h ? d : e;
          e7 = (d === h && e !== i) || (h === f && e !== g) ? h : e;
          e8 = h === f ? f : e;
        } else {
          e0 = e1 = e2 = e3 = e4 = e5 = e6 = e7 = e8 = e;
        }

        // Set output pixels
        const outX = x * 3;
        const outY = y * 3;
        result.setPixel(outX, outY, e0);
        result.setPixel(outX + 1, outY, e1);
        result.setPixel(outX + 2, outY, e2);
        result.setPixel(outX, outY + 1, e3);
        result.setPixel(outX + 1, outY + 1, e4);
        result.setPixel(outX + 2, outY + 1, e5);
        result.setPixel(outX, outY + 2, e6);
        result.setPixel(outX + 1, outY + 2, e7);
        result.setPixel(outX + 2, outY + 2, e8);
      }
    }

    return result;
  }

  getScaleFactor() {
    return 3;
  }
}