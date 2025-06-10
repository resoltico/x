/**
 * @fileoverview Main pipeline orchestrator for image processing
 * @license MIT
 * @author Ervins Strauhmanis
 */

import { Sauvola } from '../algorithms/binarization/Sauvola.js';
import { Niblack } from '../algorithms/binarization/Niblack.js';
import { Close } from '../algorithms/morphology/Close.js';
import { BinaryNoise } from '../algorithms/noise/BinaryNoise.js';
import { Scale2x } from '../algorithms/scaling/Scale2x.js';
import { Scale3x } from '../algorithms/scaling/Scale3x.js';
import { Scale4x } from '../algorithms/scaling/Scale4x.js';

export class ProcessingPipeline {
  constructor() {
    this.stages = [];
    this.algorithms = {
      binarization: {
        sauvola: Sauvola,
        niblack: Niblack
      },
      morphology: {
        close: Close
      },
      noise: {
        binary: BinaryNoise
      },
      scaling: {
        scale2x: Scale2x,
        scale3x: Scale3x,
        scale4x: Scale4x
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
      const ScalerClass = this.algorithms.scaling[`scale${parameters.scaling.method}`];
      if (ScalerClass) {
        const scaler = new ScalerClass();
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

      result = stage.processor.process(result);

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