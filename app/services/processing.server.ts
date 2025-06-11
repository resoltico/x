import { ProcessingPipeline } from '../../src/engine/pipeline/ProcessingPipeline.js';
import { ImageSaver } from '../../src/engine/utils/ImageSaver.js';
import { jobStore } from './jobStore.server';
import { wsManager } from './websocket.server';
import type { ProcessingParameters } from '~/types';
import type { ImageData } from '../../src/engine/core/ImageData.js';

interface StoredImage {
  imageData: ImageData;
  metadata: any;
  originalBuffer: Buffer;
}

export async function processImage(
  jobId: string,
  storedImage: StoredImage,
  parameters: ProcessingParameters
) {
  console.log(`🎯 Starting full image processing for job ${jobId}`);
  const startTime = Date.now();
  
  const job = jobStore.get(jobId);
  if (!job) {
    console.error(`❌ Job not found: ${jobId}`);
    return;
  }

  // Update job status
  job.status = 'processing';
  job.updatedAt = new Date();
  jobStore.set(jobId, job);

  try {
    console.log(`📐 Processing image: ${storedImage.metadata.width}x${storedImage.metadata.height}`);
    console.log(`⚙️ Parameters:`, JSON.stringify(parameters));

    // Create processing pipeline
    const pipeline = new ProcessingPipeline();
    pipeline.configure(parameters);
    
    const stages = pipeline.getStageNames();
    console.log(`📋 Processing stages: ${stages.join(' → ')}`);

    // Process image with progress callback
    const result = await pipeline.process(
      storedImage.imageData,
      (progress) => {
        const currentJob = jobStore.get(jobId);
        if (currentJob) {
          currentJob.progress = Math.round(progress.progress);
          currentJob.updatedAt = new Date();
          jobStore.set(jobId, currentJob);
        }

        // Send progress via WebSocket
        wsManager.broadcast({
          type: 'processing.progress',
          payload: {
            jobId,
            stage: progress.stage,
            stageIndex: progress.stageIndex,
            totalStages: progress.totalStages,
            progress: Math.round(progress.progress),
          },
        });

        console.log(`📊 Progress: ${progress.stage} (${Math.round(progress.progress)}%)`);
      }
    );

    console.log(`🖼️ Result dimensions: ${result.width}x${result.height}`);

    // Convert result to base64
    console.log('💾 Encoding result as PNG...');
    const resultBase64 = await ImageSaver.toBase64(result, 'png');
    
    const processingTime = Date.now() - startTime;
    console.log(`✅ Processing completed in ${processingTime}ms`);

    // Update job with result
    const finalJob = jobStore.get(jobId);
    if (finalJob) {
      finalJob.status = 'completed';
      finalJob.progress = 100;
      finalJob.result = resultBase64;
      finalJob.updatedAt = new Date();
      jobStore.set(jobId, finalJob);
      
      // Broadcast completion
      wsManager.broadcast({
        type: 'processing.complete',
        payload: {
          jobId,
          processingTime,
        },
      });
    }

  } catch (error) {
    const processingTime = Date.now() - startTime;
    console.error(`❌ Processing error after ${processingTime}ms:`, error);
    
    const errorJob = jobStore.get(jobId);
    if (errorJob) {
      errorJob.status = 'failed';
      errorJob.error = error instanceof Error ? error.message : 'Unknown error';
      errorJob.updatedAt = new Date();
      jobStore.set(jobId, errorJob);
      
      // Broadcast error
      wsManager.broadcast({
        type: 'processing.error',
        payload: {
          jobId,
          error: errorJob.error,
          processingTime,
        },
      });
    }
    throw error;
  }
}

export async function processPreview(
  imageData: ImageData,
  parameters: ProcessingParameters
): Promise<{ preview: string; histogram: number[]; processingTime: number }> {
  const startTime = Date.now();
  console.log(`🔍 Processing preview: ${imageData.width}x${imageData.height}`);

  try {
    // Validate parameters
    if (!parameters || !parameters.binarization) {
      throw new Error('Invalid parameters: binarization settings required');
    }

    // Create processing pipeline
    const pipeline = new ProcessingPipeline();
    pipeline.configure(parameters);
    
    const stages = pipeline.getStageNames();
    console.log(`📋 Preview stages: ${stages.join(' → ')}`);

    // Process image (no progress callback for preview)
    const result = await pipeline.process(imageData);

    // Get histogram
    console.log('📊 Computing histogram...');
    const histogram = result.getHistogram();

    // Convert to base64
    console.log('💾 Encoding preview...');
    const preview = await ImageSaver.toBase64(result, 'png', {
      compressionLevel: 6 // Faster compression for previews
    });

    const processingTime = Date.now() - startTime;
    console.log(`✅ Preview processed in ${processingTime}ms`);

    return { preview, histogram, processingTime };
  } catch (error) {
    const processingTime = Date.now() - startTime;
    console.error(`❌ Preview processing error after ${processingTime}ms:`, error);
    
    // Provide more context in error message
    let errorMessage = 'Failed to process preview';
    if (error instanceof Error) {
      errorMessage = `${errorMessage}: ${error.message}`;
      
      // Add specific guidance for common errors
      if (error.message.includes('parameters')) {
        errorMessage += '. Please check your processing parameters.';
      } else if (error.message.includes('memory')) {
        errorMessage += '. The image may be too large for preview processing.';
      }
    }
    
    throw new Error(errorMessage);
  }
}