// WebSocket message types

export interface WebSocketMessage {
  type: string;
  payload: any;
}

export interface PreviewUpdatePayload {
  imageId: string;
  parameters: any;
}

export interface PreviewResultPayload {
  preview: string;
  histogram: number[];
  imageId: string;
  processingTime: number;
}

export interface ProcessingProgressPayload {
  jobId: string;
  stage: string;
  stageIndex: number;
  totalStages: number;
  progress: number;
}

export interface ErrorPayload {
  message: string;
  error?: string;
  imageId?: string;
  processingTime?: number;
  suggestion?: string;
}