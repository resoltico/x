import { useState, useCallback, useRef, useEffect } from "react";
import { Upload, Download, Info, RefreshCw, Loader2 } from "lucide-react";
import { ImageUploader } from "~/components/ImageUploader";
import { PreviewCanvas } from "~/components/PreviewCanvas";
import { ParameterControls } from "~/components/ParameterControls";
import { ProgressBar } from "~/components/ProgressBar";
import { HistogramDisplay } from "~/components/HistogramDisplay";
import { useWebSocket } from "~/utils/websocketClient";
import type { ProcessingParameters } from "~/types";

export default function Index() {
  const [uploadedImage, setUploadedImage] = useState<{
    id: string;
    preview: string;
    metadata: any;
  } | null>(null);
  const [processedPreview, setProcessedPreview] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [processingProgress, setProcessingProgress] = useState(0);
  const [histogram, setHistogram] = useState<number[]>([]);
  const [parameters, setParameters] = useState<ProcessingParameters>({
    binarization: {
      method: 'sauvola',
      windowSize: 15,
      k: 0.34,
      r: 128,
    },
    morphology: {
      enabled: false,
      operation: 'close',
      kernelSize: 3,
      iterations: 1,
    },
    noise: {
      enabled: false,
      method: 'binary',
      threshold: 4,
      windowSize: 3,
    },
    scaling: {
      method: 'none',
      algorithm: 'scale2x',
    },
  });

  const ws = useWebSocket();
  const downloadLinkRef = useRef<HTMLAnchorElement>(null);

  useEffect(() => {
    if (ws) {
      ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        
        switch (data.type) {
          case 'preview.result':
            setProcessedPreview(data.payload.preview);
            setHistogram(data.payload.histogram || []);
            break;
          case 'processing.progress':
            setProcessingProgress(data.payload.progress);
            break;
        }
      };
    }
  }, [ws]);

  const handleImageUpload = async (file: File) => {
    const formData = new FormData();
    formData.append('image', file);

    try {
      const response = await fetch('/api/upload', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) throw new Error('Upload failed');

      const data = await response.json();
      setUploadedImage(data);
      setProcessedPreview(null);
      
      // Request initial preview
      requestPreview(data.id);
    } catch (error) {
      console.error('Upload error:', error);
      alert('Failed to upload image. Please try again.');
    }
  };

  const requestPreview = useCallback((imageId: string) => {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify({
        type: 'preview.update',
        payload: {
          imageId,
          parameters,
        },
      }));
    }
  }, [ws, parameters]);

  const handleParameterChange = useCallback((newParams: ProcessingParameters) => {
    setParameters(newParams);
    if (uploadedImage) {
      requestPreview(uploadedImage.id);
    }
  }, [uploadedImage, requestPreview]);

  const handleProcess = async () => {
    if (!uploadedImage) return;

    setIsProcessing(true);
    setProcessingProgress(0);

    try {
      const response = await fetch('/api/process', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          imageId: uploadedImage.id,
          parameters,
        }),
      });

      if (!response.ok) throw new Error('Processing failed');

      const data = await response.json();
      const jobId = data.jobId;

      // Poll for job completion
      const pollInterval = setInterval(async () => {
        const statusResponse = await fetch(`/api/job/${jobId}`);
        const statusData = await statusResponse.json();

        if (statusData.status === 'completed') {
          clearInterval(pollInterval);
          setIsProcessing(false);
          
          // Download the result
          if (downloadLinkRef.current) {
            downloadLinkRef.current.href = `/api/download/${jobId}`;
            downloadLinkRef.current.click();
          }
        } else if (statusData.status === 'failed') {
          clearInterval(pollInterval);
          setIsProcessing(false);
          alert('Processing failed. Please try again.');
        }
      }, 1000);
    } catch (error) {
      console.error('Processing error:', error);
      setIsProcessing(false);
      alert('Failed to process image. Please try again.');
    }
  };

  const handleReset = () => {
    setParameters({
      binarization: {
        method: 'sauvola',
        windowSize: 15,
        k: 0.34,
        r: 128,
      },
      morphology: {
        enabled: false,
        operation: 'close',
        kernelSize: 3,
        iterations: 1,
      },
      noise: {
        enabled: false,
        method: 'binary',
        threshold: 4,
        windowSize: 3,
      },
      scaling: {
        method: 'none',
        algorithm: 'scale2x',
      },
    });
    
    if (uploadedImage) {
      requestPreview(uploadedImage.id);
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <h1 className="text-2xl font-bold text-gray-900">
            Engraving Processor Pro
          </h1>
          <p className="text-sm text-gray-600 mt-1">
            Advanced image processing for historical engravings and documents
          </p>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Left Panel - Upload and Info */}
          <div className="space-y-6">
            <ImageUploader onUpload={handleImageUpload} />
            
            {uploadedImage && (
              <div className="bg-white rounded-lg shadow p-6">
                <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
                  <Info className="w-5 h-5 text-primary-600" />
                  Image Information
                </h3>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <span className="text-gray-600">Size:</span>
                    <span>{uploadedImage.metadata.width} × {uploadedImage.metadata.height}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Format:</span>
                    <span className="uppercase">{uploadedImage.metadata.format}</span>
                  </div>
                  <div className="flex justify-between">
                    <span className="text-gray-600">Channels:</span>
                    <span>{uploadedImage.metadata.channels}</span>
                  </div>
                </div>
                
                {/* Original Thumbnail */}
                <div className="mt-4">
                  <p className="text-sm text-gray-600 mb-2">Original:</p>
                  <img
                    src={uploadedImage.preview}
                    alt="Original"
                    className="w-full rounded border border-gray-200"
                  />
                </div>
              </div>
            )}
          </div>

          {/* Center Panel - Preview */}
          <div className="lg:col-span-2 space-y-6">
            <PreviewCanvas
              originalImage={uploadedImage?.preview}
              processedImage={processedPreview}
              isProcessing={isProcessing}
            />
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <HistogramDisplay data={histogram} />
              
              {/* Zoom Controls - Placeholder */}
              <div className="bg-white rounded-lg shadow p-4">
                <h3 className="text-sm font-semibold mb-2">Zoom Controls</h3>
                <div className="flex gap-2">
                  <button className="btn-secondary text-sm">Fit</button>
                  <button className="btn-secondary text-sm">100%</button>
                  <button className="btn-secondary text-sm">200%</button>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Parameter Controls */}
        {uploadedImage && (
          <div className="mt-8">
            <ParameterControls
              parameters={parameters}
              onChange={handleParameterChange}
            />
          </div>
        )}

        {/* Action Buttons */}
        {uploadedImage && (
          <div className="mt-8 flex items-center justify-between">
            <div className="flex gap-4">
              <button
                onClick={handleReset}
                className="btn-secondary flex items-center gap-2"
              >
                <RefreshCw className="w-4 h-4" />
                Reset
              </button>
            </div>
            
            <div className="flex gap-4">
              <button
                onClick={handleProcess}
                disabled={isProcessing}
                className="btn-primary flex items-center gap-2"
              >
                {isProcessing ? (
                  <>
                    <Loader2 className="w-4 h-4 animate-spin" />
                    Processing...
                  </>
                ) : (
                  <>
                    <Download className="w-4 h-4" />
                    Process Full Resolution
                  </>
                )}
              </button>
            </div>
          </div>
        )}

        {/* Progress Bar */}
        {isProcessing && (
          <div className="mt-4">
            <ProgressBar progress={processingProgress} />
          </div>
        )}

        {/* Hidden download link */}
        <a ref={downloadLinkRef} className="hidden" download="processed-image.png" />
      </main>
    </div>
  );
}