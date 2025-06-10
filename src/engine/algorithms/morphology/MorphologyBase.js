/**
 * @fileoverview Base class for morphological operations
 * @license MIT
 * @author Ervins Strauhmanis
 */

export class MorphologyBase {
  constructor(options = {}) {
    this.kernelSize = options.kernelSize || 3;
    this.iterations = options.iterations || 1;
    this.kernelType = options.kernelType || 'square';
    
    if (new.target === MorphologyBase) {
      throw new Error('MorphologyBase is an abstract class and cannot be instantiated directly');
    }
  }

  getKernel() {
    const kernel = Array(this.kernelSize).fill(null).map(() => Array(this.kernelSize).fill(0));
    
    if (this.kernelType === 'square') {
      // Fill entire kernel with 1s
      for (let i = 0; i < this.kernelSize; i++) {
        for (let j = 0; j < this.kernelSize; j++) {
          kernel[i][j] = 1;
        }
      }
    } else if (this.kernelType === 'cross') {
      // Create cross-shaped kernel
      const center = Math.floor(this.kernelSize / 2);
      for (let i = 0; i < this.kernelSize; i++) {
        kernel[i][center] = 1;
        kernel[center][i] = 1;
      }
    }
    
    return kernel;
  }

  process(imageData) {
    throw new Error('process method must be implemented by subclass');
  }

  getParameters() {
    return {
      kernelSize: this.kernelSize,
      iterations: this.iterations,
      kernelType: this.kernelType
    };
  }

  setParameters(params) {
    if (params.kernelSize !== undefined) {
      this.kernelSize = Math.max(3, Math.min(9, params.kernelSize));
      // Ensure odd kernel size
      if (this.kernelSize % 2 === 0) this.kernelSize++;
    }
    if (params.iterations !== undefined) {
      this.iterations = Math.max(1, Math.min(5, params.iterations));
    }
    if (params.kernelType !== undefined) {
      this.kernelType = params.kernelType;
    }
  }
}