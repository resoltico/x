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
  const job = jobStore.get(jobId);
  if (!job) {
    console.error('Job not found:', jobId);
    return;
  }

  // Update job status
  job.status = 'processing';
  job.updatedAt = new Date();
  jobStore.set(jobId, job);

  try {
    // Create processing pipeline
    const pipeline = new ProcessingPipeline();
    pipeline.configure(parameters);

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
            progress: Math.round(progress.progress),
          },
        });
      }
    );

    // Convert result to base64
    const resultBase64 = await ImageSaver.toBase64(result, 'png');

    // Update job with result
    const finalJob = jobStore.get(jobId);
    if (finalJob) {
      finalJob.status = 'completed';
      finalJob.progress = 100;
      finalJob.result = resultBase64;
      finalJob.updatedAt = new Date();
      jobStore.set(jobId, finalJob);
    }

  } catch (error) {
    console.error('Processing error:', error);
    const errorJob = jobStore.get(jobId);
    if (errorJob) {
      errorJob.status = 'failed';
      errorJob.error = error instanceof Error ? error.message : 'Unknown error';
      errorJob.updatedAt = new Date();
      jobStore.set(jobId, errorJob);
    }
    throw error;
  }
}

export async function processPreview(
  imageData: ImageData,
  parameters: ProcessingParameters
): Promise<{ preview: string; histogram: number[]; processingTime: number }> {
  const startTime = Date.now();

  try {
    // Create processing pipeline
    const pipeline = new ProcessingPipeline();
    pipeline.configure(parameters);

    // Process image (no progress callback for preview)
    const result = await pipeline.process(imageData);

    // Get histogram
    const histogram = result.getHistogram();

    // Convert to base64
    const preview = await ImageSaver.toBase64(result, 'png', {
      compressionLevel: 6 // Faster compression for previews
    });

    const processingTime = Date.now() - startTime;

    return { preview, histogram, processingTime };
  } catch (error) {
    console.error('Preview processing error:', error);
    throw new Error(`Failed to process preview: ${error.message}`);
  }
}