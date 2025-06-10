import type { ImageData } from '../../src/engine/core/ImageData.js';
import type { ImageMetadata } from '~/types';

interface StoredImage {
  imageData: ImageData;
  metadata: ImageMetadata;
  originalBuffer: Buffer;
  createdAt?: Date;
}

class ImageStore {
  private store: Map<string, StoredImage> = new Map();
  private maxAge = 30 * 60 * 1000; // 30 minutes

  set(id: string, image: StoredImage) {
    this.store.set(id, {
      ...image,
      createdAt: new Date(),
    });
    
    // Clean up old images
    this.cleanup();
  }

  get(id: string): StoredImage | undefined {
    return this.store.get(id);
  }

  delete(id: string) {
    this.store.delete(id);
  }

  private cleanup() {
    const now = Date.now();
    for (const [id, image] of this.store.entries()) {
      if (image.createdAt && now - image.createdAt.getTime() > this.maxAge) {
        this.store.delete(id);
      }
    }
  }
}

// Singleton instance
export const imageStore = new ImageStore();