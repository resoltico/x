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
  if (!job) return;

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
        job.progress = progress.progress;
        job.updatedAt = new Date();
        jobStore.set(jobId, job);

        // Send progress via WebSocket
        wsManager.broadcast({
          type: 'processing.progress',
          payload: {
            jobId,
            stage: progress.stage,
            progress: progress.progress,
          },
        });
      }
    );

    // Convert result to base64
    const resultBase64 = await ImageSaver.toBase64(result, 'png');

    // Update job with result
    job.status = 'completed';
    job.progress = 100;
    job.result = resultBase64;
    job.updatedAt = new Date();
    jobStore.set(jobId, job);

  } catch (error) {
    job.status = 'failed';
    job.error = error instanceof Error ? error.message : 'Unknown error';
    job.updatedAt = new Date();
    jobStore.set(jobId, job);
    throw error;
  }
}

export async function processPreview(
  imageData: ImageData,
  parameters: ProcessingParameters
): Promise<{ preview: string; histogram: number[]; processingTime: number }> {
  const startTime = Date.now();

  // Create processing pipeline
  const pipeline = new ProcessingPipeline();
  pipeline.configure(parameters);

  // Process image
  const result = await pipeline.process(imageData);

  // Get histogram
  const histogram = result.getHistogram();

  // Convert to base64
  const preview = await ImageSaver.toBase64(result, 'png');

  const processingTime = Date.now() - startTime;

  return { preview, histogram, processingTime };
}