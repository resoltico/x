/**
 * @fileoverview Scale4x pixel art scaling algorithm (2x Scale2x)
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { BaseScaler } from './BaseScaler.js';
import { Scale2x } from './Scale2x.js';

export class Scale4x extends BaseScaler {
  constructor() {
    super();
    this.scaleFactor = 4;
    this.scale2x = new Scale2x();
  }

  process(imageData) {
    // Apply Scale2x twice
    const scaled2x = this.scale2x.process(imageData);
    const scaled4x = this.scale2x.process(scaled2x);
    return scaled4x;
  }

  getScaleFactor() {
    return 4;
  }
}