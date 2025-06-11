/**
 * @fileoverview Main pipeline orchestrator for image processing
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { Sauvola } from '../algorithms/binarization/Sauvola.js';
import { Niblack } from '../algorithms/binarization/Niblack.js';
import { Otsu } from '../algorithms/binarization/Otsu.js';
import { Close } from '../algorithms/morphology/Close.js';
import { Open } from '../algorithms/morphology/Open.js';
import { Dilate } from '../algorithms/morphology/Dilate.js';
import { Erode } from '../algorithms/morphology/Erode.js';
import { BinaryNoise } from '../algorithms/noise/BinaryNoise.js';
import { MedianFilter } from '../algorithms/noise/MedianFilter.js';
import { Scale2x } from '../algorithms/scaling/Scale2x.js';
import { Scale3x } from '../algorithms/scaling/Scale3x.js';
import { Scale4x } from '../algorithms/scaling/Scale4x.js';
import { NearestNeighbor } from '../algorithms/scaling/NearestNeighbor.js';
import { Bilinear } from '../algorithms/scaling/Bilinear.js';

export class ProcessingPipeline {
  constructor() {
    this.stages = [];
    this.algorithms = {
      binarization: {
        sauvola: Sauvola,
        niblack: Niblack,
        otsu: Otsu
      },
      morphology: {
        close: Close,
        open: Open,
        dilate: Dilate,
        erode: Erode
      },
      noise: {
        binary: BinaryNoise,
        median: MedianFilter
      },
      scaling: {
        scale2x: Scale2x,
        scale3x: Scale3x,
        scale4x: Scale4x,
        nearest: NearestNeighbor,
        bilinear: Bilinear
      }
    };
  }

  configure(parameters) {
    this.stages = [];

    // 1. Binarization stage (always first)
    if (parameters.binarization) {
      const BinarizerClass = this.algorithms.binarization[parameters.binarization.method];
      if (BinarizerClass) {
        const binarizer = new BinarizerClass({
          windowSize: parameters.binarization.windowSize,
          k: parameters.binarization.k,
          r: parameters.binarization.r
        });
        this.stages.push({
          name: 'binarization',
          processor: binarizer
        });
      }
    }

    // 2. Morphology stage
    if (parameters.morphology && parameters.morphology.enabled) {
      const MorphClass = this.algorithms.morphology[parameters.morphology.operation];
      if (MorphClass) {
        const morph = new MorphClass({
          kernelSize: parameters.morphology.kernelSize,
          iterations: parameters.morphology.iterations
        });
        this.stages.push({
          name: 'morphology',
          processor: morph
        });
      }
    }

    // 3. Noise reduction stage
    if (parameters.noise && parameters.noise.enabled) {
      const NoiseClass = this.algorithms.noise[parameters.noise.method];
      if (NoiseClass) {
        const noiseReducer = new NoiseClass({
          threshold: parameters.noise.threshold,
          windowSize: parameters.noise.windowSize
        });
        this.stages.push({
          name: 'noise',
          processor: noiseReducer
        });
      }
    }

    // 4. Scaling stage (always last)
    if (parameters.scaling && parameters.scaling.method !== 'none') {
      let ScalerClass;
      let scaleFactor = 2; // Default scale factor
      
      // Parse scale factor from method
      if (parameters.scaling.method === '2x') {
        scaleFactor = 2;
      } else if (parameters.scaling.method === '3x') {
        scaleFactor = 3;
      } else if (parameters.scaling.method === '4x') {
        scaleFactor = 4;
      }
      
      // Determine which scaler to use based on algorithm
      if (parameters.scaling.algorithm === 'scale2x' && scaleFactor === 2) {
        ScalerClass = Scale2x;
      } else if (parameters.scaling.algorithm === 'scale3x' && scaleFactor === 3) {
        ScalerClass = Scale3x;
      } else if (parameters.scaling.algorithm === 'scale2x' && scaleFactor === 4) {
        // Use Scale4x (which applies Scale2x twice)
        ScalerClass = Scale4x;
      } else if (parameters.scaling.algorithm === 'nearest') {
        ScalerClass = NearestNeighbor;
      } else if (parameters.scaling.algorithm === 'bilinear') {
        ScalerClass = Bilinear;
      } else {
        // Default to nearest neighbor for unsupported combinations
        console.warn(`Using nearest neighbor for ${parameters.scaling.algorithm} at ${scaleFactor}x`);
        ScalerClass = NearestNeighbor;
      }
      
      if (ScalerClass) {
        let scaler;
        if (ScalerClass === NearestNeighbor || ScalerClass === Bilinear) {
          scaler = new ScalerClass(scaleFactor);
        } else {
          scaler = new ScalerClass();
        }
        
        this.stages.push({
          name: 'scaling',
          processor: scaler
        });
      }
    }
  }

  async process(imageData, progressCallback = null) {
    let result = imageData;
    const totalStages = this.stages.length;

    console.log(`📋 Processing pipeline with ${totalStages} stages`);

    for (let i = 0; i < totalStages; i++) {
      const stage = this.stages[i];
      
      if (progressCallback) {
        progressCallback({
          stage: stage.name,
          stageIndex: i,
          totalStages: totalStages,
          progress: (i / totalStages) * 100
        });
      }

      console.log(`⚙️ Processing stage ${i + 1}/${totalStages}: ${stage.name}`);
      const stageStart = Date.now();
      
      try {
        result = stage.processor.process(result);
        console.log(`✅ Stage ${stage.name} completed in ${Date.now() - stageStart}ms`);
      } catch (error) {
        console.error(`❌ Stage ${stage.name} failed:`, error);
        throw new Error(`Processing failed at ${stage.name} stage: ${error.message}`);
      }

      // Allow for async breaks to prevent blocking
      if (i < totalStages - 1) {
        await new Promise(resolve => setTimeout(resolve, 0));
      }
    }

    if (progressCallback) {
      progressCallback({
        stage: 'complete',
        stageIndex: totalStages,
        totalStages: totalStages,
        progress: 100
      });
    }

    return result;
  }

  getStageNames() {
    return this.stages.map(stage => stage.name);
  }

  getStageCount() {
    return this.stages.length;
  }
}