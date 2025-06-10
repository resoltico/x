import { useState } from "react";
import { Loader2, ToggleLeft, ToggleRight } from "lucide-react";

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

  const displayImage = showOriginal ? originalImage : processedImage;

  return (
    <div className="bg-white rounded-lg shadow">
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold">Preview</h2>
          
          {originalImage && processedImage && (
            <button
              onClick={() => setShowOriginal(!showOriginal)}
              className="flex items-center gap-2 text-sm font-medium text-gray-700 hover:text-gray-900"
            >
              {showOriginal ? (
                <>
                  <ToggleLeft className="w-5 h-5" />
                  Showing: Original
                </>
              ) : (
                <>
                  <ToggleRight className="w-5 h-5" />
                  Showing: Processed
                </>
              )}
            </button>
          )}
        </div>
      </div>
      
      <div className="relative aspect-[4/3] bg-gray-100">
        {displayImage ? (
          <img
            src={displayImage}
            alt={showOriginal ? "Original" : "Processed"}
            className="w-full h-full object-contain"
          />
        ) : (
          <div className="absolute inset-0 flex items-center justify-center">
            {isProcessing ? (
              <div className="text-center">
                <Loader2 className="w-8 h-8 text-blue-600 animate-spin mx-auto mb-2" />
                <p className="text-sm text-gray-600">Processing preview...</p>
              </div>
            ) : (
              <p className="text-gray-500">Upload an image to begin</p>
            )}
          </div>
        )}
      </div>
    </div>
  );
}