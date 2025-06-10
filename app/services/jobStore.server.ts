import type { ProcessingJob } from '~/types';

class JobStore {
  private store: Map<string, ProcessingJob> = new Map();
  private maxAge = 60 * 60 * 1000; // 1 hour

  set(id: string, job: ProcessingJob) {
    this.store.set(id, job);
    this.cleanup();
  }

  get(id: string): ProcessingJob | undefined {
    return this.store.get(id);
  }

  delete(id: string) {
    this.store.delete(id);
  }

  private cleanup() {
    const now = Date.now();
    for (const [id, job] of this.store.entries()) {
      if (now - job.createdAt.getTime() > this.maxAge) {
        this.store.delete(id);
      }
    }
  }
}

// Singleton instance
export const jobStore = new JobStore();