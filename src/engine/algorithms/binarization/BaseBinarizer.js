/**
 * @fileoverview Abstract base class for binarization algorithms
 * @license MIT
 * @author Ervins Strauhmanis
 */

export class BaseBinarizer {
  constructor() {
    if (new.target === BaseBinarizer) {
      throw new Error('BaseBinarizer is an abstract class and cannot be instantiated directly');
    }
  }

  /**
   * Process an image and return a binarized version
   * @param {ImageData} imageData - Input image
   * @returns {ImageData} Binarized image
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

  /**
   * Get parameter metadata for UI controls
   * @returns {Object} Parameter metadata
   */
  static getParameterMetadata() {
    return {};
  }
}