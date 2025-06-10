/**
 * @fileoverview Morphological closing operation (dilation followed by erosion)
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { MorphologyBase } from './MorphologyBase.js';
import { Dilate } from './Dilate.js';
import { Erode } from './Erode.js';

export class Close extends MorphologyBase {
  constructor(options = {}) {
    super(options);
    this.dilate = new Dilate(options);
    this.erode = new Erode(options);
  }

  process(imageData) {
    // Closing = Dilation followed by Erosion
    let result = imageData;
    
    // Apply dilation first
    for (let i = 0; i < this.iterations; i++) {
      result = this.dilate.processSingle(result);
    }
    
    // Then apply erosion
    for (let i = 0; i < this.iterations; i++) {
      result = this.erode.processSingle(result);
    }
    
    return result;
  }

  setParameters(params) {
    super.setParameters(params);
    this.dilate.setParameters(params);
    this.erode.setParameters(params);
  }
}