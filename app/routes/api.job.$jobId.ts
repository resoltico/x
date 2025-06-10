import { type LoaderFunctionArgs, json } from "@remix-run/node";
import { jobStore } from '~/services/jobStore.server';

export async function loader({ params }: LoaderFunctionArgs) {
  const { jobId } = params;
  
  if (!jobId) {
    return json({ error: 'Job ID required' }, { status: 400 });
  }

  const job = jobStore.get(jobId);
  
  if (!job) {
    return json({ error: 'Job not found' }, { status: 404 });
  }

  return json({
    id: job.id,
    status: job.status,
    progress: job.progress,
    result: job.result,
    error: job.error,
  });
}