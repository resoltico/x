/**
 * @fileoverview Base class for noise reduction algorithms
 * @license MIT
 * @author Ervins Strauhmanis
 */

export class NoiseReducer {
  constructor() {
    if (new.target === NoiseReducer) {
      throw new Error('NoiseReducer is an abstract class and cannot be instantiated directly');
    }
  }

  /**
   * Process an image and return a noise-reduced version
   * @param {ImageData} imageData - Input image
   * @returns {ImageData} Noise-reduced image
   */
  process(imageData) {
    throw new Error('process method must be implemented by subclass');
  }

  /**
   * Get current parameters
   * @returns {Object} Current algorithm parameters
   */
  getParameters() {
    throw new Error('getParameters method must be implemented by subclass');
  }

  /**
   * Set algorithm parameters
   * @param {Object} params - New parameters
   */
  setParameters(params) {
    throw new Error('setParameters method must be implemented by subclass');
  }
}