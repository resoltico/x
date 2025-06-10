import { type LoaderFunctionArgs } from "@remix-run/node";
import { jobStore } from '~/services/jobStore.server';

export async function loader({ params }: LoaderFunctionArgs) {
  const { jobId } = params;
  
  if (!jobId) {
    return new Response('Job ID required', { status: 400 });
  }

  const job = jobStore.get(jobId);
  
  if (!job) {
    return new Response('Job not found', { status: 404 });
  }

  if (job.status !== 'completed' || !job.result) {
    return new Response('Job not completed', { status: 400 });
  }

  // Extract base64 data
  const base64Data = job.result.split(',')[1];
  const buffer = Buffer.from(base64Data, 'base64');

  // Return as PNG download
  return new Response(buffer, {
    status: 200,
    headers: {
      'Content-Type': 'image/png',
      'Content-Disposition': 'attachment; filename="processed-engraving.png"',
      'Content-Length': buffer.length.toString(),
    },
  });
}