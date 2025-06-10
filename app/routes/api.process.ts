import { type ActionFunctionArgs, json } from "@remix-run/node";
import { v4 as uuidv4 } from 'uuid';
import { imageStore } from '~/services/imageStore.server';
import { jobStore } from '~/services/jobStore.server';
import { processImage } from '~/services/processing.server';
import type { ProcessingParameters } from '~/types';

export async function action({ request }: ActionFunctionArgs) {
  try {
    const body = await request.json();
    const { imageId, parameters } = body as {
      imageId: string;
      parameters: ProcessingParameters;
    };

    // Validate image exists
    const storedImage = imageStore.get(imageId);
    if (!storedImage) {
      return json({ error: 'Image not found' }, { status: 404 });
    }

    // Create job
    const jobId = uuidv4();
    const job = {
      id: jobId,
      imageId,
      parameters,
      status: 'pending' as const,
      progress: 0,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    jobStore.set(jobId, job);

    // Start processing in background
    processImage(jobId, storedImage, parameters).catch((error) => {
      console.error('Processing error:', error);
      const job = jobStore.get(jobId);
      if (job) {
        job.status = 'failed';
        job.error = error.message;
        job.updatedAt = new Date();
        jobStore.set(jobId, job);
      }
    });

    return json({ jobId });
  } catch (error) {
    console.error('Process error:', error);
    return json(
      { error: 'Failed to start processing' },
      { status: 500 }
    );
  }
}