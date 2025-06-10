import { describe, it, expect } from 'vitest';
import { Sauvola } from './Sauvola.js';
import { ImageData } from '../../core/ImageData.js';

describe('Sauvola Binarization', () => {
  it('should create instance with default parameters', () => {
    const sauvola = new Sauvola();
    const params = sauvola.getParameters();
    
    expect(params.windowSize).toBe(15);
    expect(params.k).toBe(0.34);
    expect(params.r).toBe(128);
  });

  it('should accept custom parameters', () => {
    const sauvola = new Sauvola({
      windowSize: 21,
      k: 0.5,
      r: 100
    });
    const params = sauvola.getParameters();
    
    expect(params.windowSize).toBe(21);
    expect(params.k).toBe(0.5);
    expect(params.r).toBe(100);
  });

  it('should process grayscale image', () => {
    // Create a simple test image
    const data = new Uint8ClampedArray([
      128, 200, 50, 128,
      200, 50, 128, 200,
      50, 128, 200, 50,
      128, 200, 50, 128
    ]);
    const image = ImageData.fromGrayscale(data, 4, 4);
    
    const sauvola = new Sauvola({ windowSize: 3, k: 0.3, r: 128 });
    const result = sauvola.process(image);
    
    expect(result).toBeDefined();
    expect(result.width).toBe(4);
    expect(result.height).toBe(4);
    expect(result.channels).toBe(1);
    
    // Check that result is binary (only 0 or 255)
    for (let i = 0; i < result.data.length; i++) {
      expect([0, 255]).toContain(result.data[i]);
    }
  });

  it('should enforce odd window sizes', () => {
    const sauvola = new Sauvola();
    
    sauvola.setParameters({ windowSize: 10 });
    expect(sauvola.getParameters().windowSize).toBe(11);
    
    sauvola.setParameters({ windowSize: 15 });
    expect(sauvola.getParameters().windowSize).toBe(15);
  });

  it('should clamp parameters to valid ranges', () => {
    const sauvola = new Sauvola();
    
    // Test window size limits
    sauvola.setParameters({ windowSize: 2 });
    expect(sauvola.getParameters().windowSize).toBe(3);
    
    sauvola.setParameters({ windowSize: 100 });
    expect(sauvola.getParameters().windowSize).toBe(51);
    
    // Test k parameter limits
    sauvola.setParameters({ k: 0 });
    expect(sauvola.getParameters().k).toBe(0.1);
    
    sauvola.setParameters({ k: 2 });
    expect(sauvola.getParameters().k).toBe(1.0);
    
    // Test r parameter limits
    sauvola.setParameters({ r: -10 });
    expect(sauvola.getParameters().r).toBe(1);
    
    sauvola.setParameters({ r: 300 });
    expect(sauvola.getParameters().r).toBe(255);
  });
});