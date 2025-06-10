/**
 * @fileoverview Abstract base class for scaling algorithms
 * @license MIT
 * @author Ervins Strauhmanis
 */

export class BaseScaler {
  constructor() {
    if (new.target === BaseScaler) {
      throw new Error('BaseScaler is an abstract class and cannot be instantiated directly');
    }
  }

  /**
   * Process an image and return a scaled version
   * @param {ImageData} imageData - Input image
   * @returns {ImageData} Scaled image
   */
  process(imageData) {
    throw new Error('process method must be implemented by subclass');
  }

  /**
   * Get the scale factor of this algorithm
   * @returns {number} Scale factor
   */
  getScaleFactor() {
    throw new Error('getScaleFactor method must be implemented by subclass');
  }

  /**
   * Get current parameters
   * @returns {Object} Current algorithm parameters
   */
  getParameters() {
    return {};
  }

  /**
   * Set algorithm parameters
   * @param {Object} params - New parameters
   */
  setParameters(params) {
    // Override in subclass if parameters are supported
  }
}