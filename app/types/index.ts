export interface ProcessingParameters {
  binarization: {
    method: 'sauvola' | 'niblack' | 'otsu';
    windowSize: number;
    k: number;
    r: number;
  };
  morphology: {
    enabled: boolean;
    operation: 'close' | 'open' | 'dilate' | 'erode';
    kernelSize: number;
    iterations: number;
  };
  noise: {
    enabled: boolean;
    method: 'binary' | 'median';
    threshold: number;
    windowSize: number;
  };
  scaling: {
    method: 'none' | '2x' | '3x' | '4x';
    algorithm: 'scale2x' | 'scale3x' | 'nearest' | 'bilinear';
  };
}

export interface ImageMetadata {
  width: number;
  height: number;
  channels: number;
  format: string;
  size?: number;
}

export interface ProcessingJob {
  id: string;
  imageId: string;
  parameters: ProcessingParameters;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress: number;
  result?: string;
  error?: string;
  createdAt: Date;
  updatedAt: Date;
}