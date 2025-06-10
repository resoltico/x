// Type declarations for the JavaScript engine modules

declare module '../../src/engine/core/ImageData.js' {
  export class ImageData {
    constructor(data: Uint8ClampedArray, width: number, height: number, channels?: number);
    data: Uint8ClampedArray;
    width: number;
    height: number;
    channels: number;
    
    static fromRGBA(rgbaData: Uint8ClampedArray, width: number, height: number): ImageData;
    static fromGrayscale(grayData: Uint8ClampedArray, width: number, height: number): ImageData;
    static createEmpty(width: number, height: number, channels?: number): ImageData;
    
    getPixel(x: number, y: number): number | number[] | null;
    setPixel(x: number, y: number, value: number | number[]): void;
    toGrayscale(): ImageData;
    toRGBA(): ImageData;
    clone(): ImageData;
    getHistogram(): number[];
  }
}

declare module '../../src/engine/utils/ImageLoader.js' {
  import type { ImageData } from '../../src/engine/core/ImageData.js';
  
  interface ImageMetadata {
    width: number;
    height: number;
    channels: number;
    format: string;
    size?: number;
    path?: string;
  }
  
  export class ImageLoader {
    static loadFromBuffer(buffer: Buffer, options?: any): Promise<{
      imageData: ImageData;
      metadata: ImageMetadata;
    }>;
    static loadFromPath(path: string, options?: any): Promise<{
      imageData: ImageData;
      metadata: ImageMetadata;
    }>;
    static createPreview(imageData: ImageData, maxSize?: number): Promise<ImageData>;
    static toBuffer(imageData: ImageData): Promise<Buffer>;
  }
}

declare module '../../src/engine/utils/ImageSaver.js' {
  import type { ImageData } from '../../src/engine/core/ImageData.js';
  
  export class ImageSaver {
    static saveToBuffer(imageData: ImageData, format?: string, options?: any): Promise<Buffer>;
    static saveToFile(imageData: ImageData, path: string, format?: string | null, options?: any): Promise<{
      path: string;
      format: string;
      size: number;
    }>;
    static toBase64(imageData: ImageData, format?: string, options?: any): Promise<string>;
  }
}

declare module '../../src/engine/pipeline/ProcessingPipeline.js' {
  import type { ImageData } from '../../src/engine/core/ImageData.js';
  import type { ProcessingParameters } from '../index';
  
  interface ProgressCallback {
    (progress: {
      stage: string;
      stageIndex: number;
      totalStages: number;
      progress: number;
    }): void;
  }
  
  export class ProcessingPipeline {
    configure(parameters: ProcessingParameters): void;
    process(imageData: ImageData, progressCallback?: ProgressCallback | null): Promise<ImageData>;
    getStageNames(): string[];
    getStageCount(): number;
  }
}