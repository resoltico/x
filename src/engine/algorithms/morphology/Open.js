/**
 * @fileoverview Morphological opening operation (erosion followed by dilation)
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { MorphologyBase } from './MorphologyBase.js';
import { Erode } from './Erode.js';
import { Dilate } from './Dilate.js';

export class Open extends MorphologyBase {
  constructor(options = {}) {
    super(options);
    this.erode = new Erode(options);
    this.dilate = new Dilate(options);
  }

  process(imageData) {
    // Opening = Erosion followed by Dilation
    let result = imageData;
    
    // Apply erosion first
    for (let i = 0; i < this.iterations; i++) {
      result = this.erode.processSingle(result);
    }
    
    // Then apply dilation
    for (let i = 0; i < this.iterations; i++) {
      result = this.dilate.processSingle(result);
    }
    
    return result;
  }

  setParameters(params) {
    super.setParameters(params);
    this.erode.setParameters(params);
    this.dilate.setParameters(params);
  }
}