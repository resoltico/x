// public/workers/imageProcessingWorker.js
// Fallback worker file for production builds
(function() {
  'use strict';
  
  console.log('🔧 Production fallback worker starting...');

  function simpleBinarization(imageData, threshold = 128) {
    return new Promise(async (resolve) => {
      const canvas = new OffscreenCanvas(imageData.width, imageData.height);
      const ctx = canvas.getContext('2d');
      const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
      ctx.putImageData(canvasImageData, 0, 0);
      
      const data = ctx.getImageData(0, 0, imageData.width, imageData.height);
      const pixels = data.data;
      
      for (let i = 0; i < pixels.length; i += 4) {
        const gray = pixels[i] * 0.299 + pixels[i + 1] * 0.587 + pixels[i + 2] * 0.114;
        const binary = gray > threshold ? 255 : 0;
        pixels[i] = binary;
        pixels[i + 1] = binary;
        pixels[i + 2] = binary;
      }
      
      ctx.putImageData(data, 0, 0);
      const blob = await canvas.convertToBlob();
      const result = await blob.arrayBuffer();
      resolve(result);
    });
  }
  
  function simpleScale(imageData, factor) {
    return new Promise(async (resolve) => {
      const canvas = new OffscreenCanvas(imageData.width, imageData.height);
      const ctx = canvas.getContext('2d');
      const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
      ctx.putImageData(canvasImageData, 0, 0);
      
      const scaledCanvas = new OffscreenCanvas(
        Math.round(imageData.width * factor),
        Math.round(imageData.height * factor)
      );
      const scaledCtx = scaledCanvas.getContext('2d');
      scaledCtx.imageSmoothingEnabled = false;
      scaledCtx.drawImage(canvas, 0, 0, scaledCanvas.width, scaledCanvas.height);
      
      const blob = await scaledCanvas.convertToBlob();
      const result = await blob.arrayBuffer();
      resolve(result);
    });
  }
  
  function simpleMorphology(imageData, operation) {
    return new Promise(async (resolve) => {
      const canvas = new OffscreenCanvas(imageData.width, imageData.height);
      const ctx = canvas.getContext('2d');
      const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
      ctx.putImageData(canvasImageData, 0, 0);
      
      const data = ctx.getImageData(0, 0, imageData.width, imageData.height);
      const pixels = data.data;
      const newPixels = new Uint8ClampedArray(pixels);
      
      for (let y = 1; y < imageData.height - 1; y++) {
        for (let x = 1; x < imageData.width - 1; x++) {
          const idx = (y * imageData.width + x) * 4;
          let value = pixels[idx];
          
          // Simple 3x3 kernel operation
          if (operation === 'erosion' || operation === 'opening') {
            for (let dy = -1; dy <= 1; dy++) {
              for (let dx = -1; dx <= 1; dx++) {
                const nIdx = ((y + dy) * imageData.width + (x + dx)) * 4;
                value = Math.min(value, pixels[nIdx]);
              }
            }
          } else if (operation === 'dilation' || operation === 'closing') {
            for (let dy = -1; dy <= 1; dy++) {
              for (let dx = -1; dx <= 1; dx++) {
                const nIdx = ((y + dy) * imageData.width + (x + dx)) * 4;
                value = Math.max(value, pixels[nIdx]);
              }
            }
          }
          
          newPixels[idx] = value;
          newPixels[idx + 1] = value;
          newPixels[idx + 2] = value;
        }
      }
      
      const newData = new ImageData(newPixels, imageData.width, imageData.height);
      ctx.putImageData(newData, 0, 0);
      const blob = await canvas.convertToBlob();
      const result = await blob.arrayBuffer();
      resolve(result);
    });
  }
  
  function simpleNoiseReduction(imageData) {
    return new Promise(async (resolve) => {
      const canvas = new OffscreenCanvas(imageData.width, imageData.height);
      const ctx = canvas.getContext('2d');
      const canvasImageData = new ImageData(new Uint8ClampedArray(imageData.data), imageData.width, imageData.height);
      ctx.putImageData(canvasImageData, 0, 0);
      
      const data = ctx.getImageData(0, 0, imageData.width, imageData.height);
      const pixels = data.data;
      const newPixels = new Uint8ClampedArray(pixels);
      
      for (let y = 1; y < imageData.height - 1; y++) {
        for (let x = 1; x < imageData.width - 1; x++) {
          const idx = (y * imageData.width + x) * 4;
          const values = [];
          
          for (let dy = -1; dy <= 1; dy++) {
            for (let dx = -1; dx <= 1; dx++) {
              const nIdx = ((y + dy) * imageData.width + (x + dx)) * 4;
              values.push(pixels[nIdx]);
            }
          }
          
          values.sort((a, b) => a - b);
          const median = values[Math.floor(values.length / 2)];
          
          newPixels[idx] = median;
          newPixels[idx + 1] = median;
          newPixels[idx + 2] = median;
        }
      }
      
      const newData = new ImageData(newPixels, imageData.width, imageData.height);
      ctx.putImageData(newData, 0, 0);
      const blob = await canvas.convertToBlob();
      const result = await blob.arrayBuffer();
      resolve(result);
    });
  }
  
  self.onmessage = async function(event) {
    const { id, type, payload } = event.data;
    
    console.log('🔧 Production worker received:', type, 'for task:', id);
    
    if (type === 'test') {
      self.postMessage({ id, type: 'test-response' });
      return;
    }
    
    if (type === 'process') {
      try {
        const { imageData, type: processType, parameters } = payload;
        
        self.postMessage({
          id, type: 'progress',
          payload: { progress: 25, message: 'Processing with production worker...' }
        });
        
        let result;
        
        switch (processType) {
          case 'binarization':
            const threshold = parameters.binarization?.threshold || 128;
            result = await simpleBinarization(imageData, threshold);
            break;
            
          case 'scaling':
            const factor = parameters.scaling?.factor || 2;
            result = await simpleScale(imageData, factor);
            break;
            
          case 'morphology':
            const operation = parameters.morphology?.operation || 'opening';
            result = await simpleMorphology(imageData, operation);
            break;
            
          case 'noise-reduction':
            result = await simpleNoiseReduction(imageData);
            break;
            
          default:
            // Return original data if processing type is unknown
            result = imageData.data.slice(0);
        }
        
        self.postMessage({
          id, type: 'progress',
          payload: { progress: 75, message: 'Finalizing...' }
        });
        
        self.postMessage({
          id, type: 'result',
          payload: { result }
        }, result instanceof ArrayBuffer ? [result] : []);
        
      } catch (error) {
        console.error('🔧 Production worker error:', error);
        self.postMessage({
          id, type: 'error',
          payload: { error: 'Production processing failed: ' + error.message }
        });
      }
    }
  };
  
  self.onerror = function(error) {
    console.error('🔧 Production worker script error:', error);
  };
  
  self.onunhandledrejection = function(event) {
    console.error('🔧 Production worker unhandled rejection:', event.reason);
    event.preventDefault();
  };
  
  console.log('🔧 Production worker initialized and ready');
})();