import { useState, useEffect } from "react";
import { Loader2, ToggleLeft, ToggleRight, AlertCircle, Image as ImageIcon, RefreshCw } from "lucide-react";

interface PreviewCanvasProps {
  originalImage?: string | null;
  processedImage?: string | null;
  isProcessing?: boolean;
  onRetryPreview?: () => void;
}

export function PreviewCanvas({ 
  originalImage, 
  processedImage, 
  isProcessing,
  onRetryPreview
}: PreviewCanvasProps) {
  const [showOriginal, setShowOriginal] = useState(false);
  const [imageLoadError, setImageLoadError] = useState(false);
  const [processedLoadError, setProcessedLoadError] = useState(false);
  const [loadingStartTime, setLoadingStartTime] = useState<number | null>(null);
  const [showSlowWarning, setShowSlowWarning] = useState(false);

  const displayImage = showOriginal ? originalImage : processedImage;
  const hasError = showOriginal ? imageLoadError : processedLoadError;

  // Track loading time
  useEffect(() => {
    if (isProcessing && !loadingStartTime) {
      setLoadingStartTime(Date.now());
      setShowSlowWarning(false);
      
      // Show warning after 5 seconds
      const warningTimeout = setTimeout(() => {
        setShowSlowWarning(true);
      }, 5000);
      
      return () => clearTimeout(warningTimeout);
    } else if (!isProcessing && loadingStartTime) {
      const duration = Date.now() - loadingStartTime;
      console.log(`Preview generation took ${duration}ms`);
      setLoadingStartTime(null);
      setShowSlowWarning(false);
    }
  }, [isProcessing, loadingStartTime]);

  // Reset error states when images change
  useEffect(() => {
    if (processedImage) {
      setProcessedLoadError(false);
    }
  }, [processedImage]);

  useEffect(() => {
    if (originalImage) {
      setImageLoadError(false);
    }
  }, [originalImage]);

  const handleImageError = () => {
    if (showOriginal) {
      setImageLoadError(true);
    } else {
      setProcessedLoadError(true);
    }
  };

  const handleImageLoad = () => {
    if (showOriginal) {
      setImageLoadError(false);
    } else {
      setProcessedLoadError(false);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow">
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Preview</h2>
          
          <div className="flex items-center gap-4">
            {isProcessing && (
              <span className="text-sm text-gray-500">
                Generating preview...
              </span>
            )}
            
            {originalImage && processedImage && !isProcessing && (
              <button
                onClick={() => setShowOriginal(!showOriginal)}
                className="flex items-center gap-2 text-sm font-medium text-gray-700 hover:text-gray-900 transition-colors"
              >
                {showOriginal ? (
                  <>
                    <ToggleLeft className="w-5 h-5" />
                    <span>Showing: Original</span>
                  </>
                ) : (
                  <>
                    <ToggleRight className="w-5 h-5" />
                    <span>Showing: Processed</span>
                  </>
                )}
              </button>
            )}
          </div>
        </div>
      </div>
      
      <div className="relative aspect-[4/3] bg-gray-100 overflow-hidden">
        {displayImage && !hasError ? (
          <img
            src={displayImage}
            alt={showOriginal ? "Original" : "Processed"}
            className="w-full h-full object-contain"
            onError={handleImageError}
            onLoad={handleImageLoad}
          />
        ) : (
          <div className="absolute inset-0 flex items-center justify-center">
            {hasError ? (
              <div className="text-center p-4">
                <AlertCircle className="w-8 h-8 text-red-500 mx-auto mb-2" />
                <p className="text-sm text-gray-600 mb-2">Failed to load image</p>
                <button
                  onClick={() => {
                    if (showOriginal) {
                      setImageLoadError(false);
                    } else {
                      setProcessedLoadError(false);
                      if (onRetryPreview) {
                        onRetryPreview();
                      }
                    }
                  }}
                  className="text-xs text-blue-600 hover:text-blue-700 flex items-center gap-1 mx-auto"
                >
                  <RefreshCw className="w-3 h-3" />
                  Try again
                </button>
              </div>
            ) : isProcessing ? (
              <div className="text-center p-4">
                <Loader2 className="w-8 h-8 text-blue-600 animate-spin mx-auto mb-2" />
                <p className="text-sm text-gray-600">
                  Generating preview...
                </p>
                {showSlowWarning && (
                  <div className="mt-3 space-y-1">
                    <p className="text-xs text-amber-600">
                      This is taking longer than usual
                    </p>
                    <p className="text-xs text-gray-500">
                      Large images or complex parameters may take more time
                    </p>
                  </div>
                )}
              </div>
            ) : originalImage && !processedImage ? (
              <div className="text-center p-4">
                <div className="relative">
                  <ImageIcon className="w-8 h-8 text-gray-400 mx-auto mb-2" />
                  <div className="absolute inset-0 flex items-center justify-center">
                    <Loader2 className="w-12 h-12 text-blue-200 animate-spin" />
                  </div>
                </div>
                <p className="text-sm text-gray-600">Preparing preview...</p>
                <p className="text-xs text-gray-500 mt-1">
                  Processing will begin shortly
                </p>
              </div>
            ) : (
              <div className="text-center p-4">
                <ImageIcon className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                <p className="text-gray-500">Upload an image to begin</p>
                <p className="text-xs text-gray-400 mt-1">
                  Drag and drop or click to browse
                </p>
              </div>
            )}
          </div>
        )}
      </div>
      
      {/* Status bar */}
      {(originalImage || processedImage) && (
        <div className="px-4 py-2 bg-gray-50 border-t border-gray-200 text-xs text-gray-600">
          <div className="flex items-center justify-between">
            <span>
              {showOriginal ? 'Original Image' : 'Processed Preview (512px max)'}
            </span>
            <div className="flex items-center gap-3">
              {processedImage && !showOriginal && (
                <span className="text-green-600 flex items-center gap-1">
                  <span className="w-2 h-2 bg-green-600 rounded-full animate-pulse"></span>
                  Ready
                </span>
              )}
              {isProcessing && (
                <span className="text-blue-600 flex items-center gap-1">
                  <span className="w-2 h-2 bg-blue-600 rounded-full animate-pulse"></span>
                  Processing
                </span>
              )}
              {hasError && (
                <span className="text-red-600 flex items-center gap-1">
                  <span className="w-2 h-2 bg-red-600 rounded-full"></span>
                  Error
                </span>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}