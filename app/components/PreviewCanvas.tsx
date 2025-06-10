import { useState, useEffect } from "react";
import { Loader2, ToggleLeft, ToggleRight, AlertCircle, Image as ImageIcon } from "lucide-react";

interface PreviewCanvasProps {
  originalImage?: string | null;
  processedImage?: string | null;
  isProcessing?: boolean;
}

export function PreviewCanvas({ 
  originalImage, 
  processedImage, 
  isProcessing 
}: PreviewCanvasProps) {
  const [showOriginal, setShowOriginal] = useState(false);
  const [imageLoadError, setImageLoadError] = useState(false);
  const [isLoadingPreview, setIsLoadingPreview] = useState(false);

  const displayImage = showOriginal ? originalImage : processedImage;

  // Reset loading state when processed image changes
  useEffect(() => {
    if (processedImage) {
      setIsLoadingPreview(false);
      setImageLoadError(false);
    }
  }, [processedImage]);

  // Start loading when original image is set but no processed image yet
  useEffect(() => {
    if (originalImage && !processedImage && !isProcessing) {
      setIsLoadingPreview(true);
    }
  }, [originalImage, processedImage, isProcessing]);

  return (
    <div className="bg-white rounded-lg shadow">
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Preview</h2>
          
          {originalImage && processedImage && (
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
      
      <div className="relative aspect-[4/3] bg-gray-100 overflow-hidden">
        {displayImage && !imageLoadError ? (
          <img
            src={displayImage}
            alt={showOriginal ? "Original" : "Processed"}
            className="w-full h-full object-contain"
            onError={() => setImageLoadError(true)}
            onLoad={() => setImageLoadError(false)}
          />
        ) : (
          <div className="absolute inset-0 flex items-center justify-center">
            {imageLoadError ? (
              <div className="text-center">
                <AlertCircle className="w-8 h-8 text-red-500 mx-auto mb-2" />
                <p className="text-sm text-gray-600">Failed to load image</p>
                <button
                  onClick={() => setImageLoadError(false)}
                  className="mt-2 text-xs text-blue-600 hover:text-blue-700"
                >
                  Try again
                </button>
              </div>
            ) : isProcessing || isLoadingPreview ? (
              <div className="text-center">
                <Loader2 className="w-8 h-8 text-blue-600 animate-spin mx-auto mb-2" />
                <p className="text-sm text-gray-600">
                  {isProcessing ? 'Processing image...' : 'Generating preview...'}
                </p>
                <p className="text-xs text-gray-500 mt-1">
                  This may take a few seconds
                </p>
              </div>
            ) : originalImage && !processedImage ? (
              <div className="text-center">
                <ImageIcon className="w-8 h-8 text-gray-400 mx-auto mb-2" />
                <p className="text-sm text-gray-600">Preparing preview...</p>
              </div>
            ) : (
              <div className="text-center">
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
            {processedImage && !showOriginal && (
              <span className="text-green-600 flex items-center gap-1">
                <span className="w-2 h-2 bg-green-600 rounded-full"></span>
                Ready
              </span>
            )}
          </div>
        </div>
      )}
    </div>
  );
}