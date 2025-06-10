import { useCallback, useRef, useState } from "react";
import { Upload, Image as ImageIcon, Check } from "lucide-react";

interface ImageUploaderProps {
  onUpload: (file: File) => void;
}

export function ImageUploader({ onUpload }: ImageUploaderProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadSuccess, setUploadSuccess] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragEnter = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    // Only set dragging to false if we're leaving the drop zone entirely
    const rect = e.currentTarget.getBoundingClientRect();
    const x = e.clientX;
    const y = e.clientY;
    
    if (x <= rect.left || x >= rect.right || y <= rect.top || y >= rect.bottom) {
      setIsDragging(false);
    }
  }, []);

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);

    const files = Array.from(e.dataTransfer.files);
    const imageFile = files.find(file => file.type.startsWith('image/'));
    
    if (imageFile) {
      processFile(imageFile);
    } else {
      alert('Please drop an image file (PNG, JPEG, TIFF, or WebP)');
    }
  }, []);

  const handleFileChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      processFile(file);
    }
  }, []);

  const processFile = async (file: File) => {
    setIsUploading(true);
    setUploadSuccess(false);
    
    try {
      await onUpload(file);
      setUploadSuccess(true);
      // Reset success indicator after 2 seconds
      setTimeout(() => setUploadSuccess(false), 2000);
    } catch (error) {
      console.error('Upload failed:', error);
    } finally {
      setIsUploading(false);
    }
  };

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className="bg-white rounded-lg shadow">
      <div
        className={`
          relative p-8 border-3 border-dashed rounded-lg cursor-pointer 
          transition-all duration-200 min-h-[200px] flex items-center justify-center
          ${isDragging 
            ? 'border-blue-500 bg-blue-50 scale-[1.02] shadow-lg' 
            : 'border-gray-300 hover:border-gray-400 hover:bg-gray-50'
          }
          ${isUploading ? 'opacity-50 cursor-wait' : ''}
          ${uploadSuccess ? 'border-green-500 bg-green-50' : ''}
        `}
        onDragOver={handleDragOver}
        onDragEnter={handleDragEnter}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={!isUploading ? handleClick : undefined}
      >
        {/* Invisible overlay to ensure consistent drag behavior */}
        <div className="absolute inset-0 pointer-events-none" />
        
        <div className="text-center relative z-10">
          <div className="mb-4">
            {isUploading ? (
              <div className="mx-auto h-12 w-12 text-blue-500">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
              </div>
            ) : uploadSuccess ? (
              <Check className="mx-auto h-12 w-12 text-green-500 animate-bounce" />
            ) : isDragging ? (
              <ImageIcon className="mx-auto h-12 w-12 text-blue-500 animate-pulse" />
            ) : (
              <Upload className="mx-auto h-12 w-12 text-gray-400" />
            )}
          </div>
          
          <p className="text-sm font-medium text-gray-900">
            {isUploading ? 'Uploading...' : 
             uploadSuccess ? 'Upload successful!' :
             isDragging ? 'Drop your image here!' : 
             'Drag and drop your image here'}
          </p>
          <p className="mt-1 text-xs text-gray-500">
            {!isUploading && !uploadSuccess && (
              <>or click to browse • Auto-uploads on selection</>
            )}
          </p>
          <p className="mt-2 text-xs text-gray-400">
            Supports PNG, JPEG, TIFF, and WebP formats (max 10MB)
          </p>
          
          <input
            ref={fileInputRef}
            type="file"
            className="hidden"
            accept="image/*"
            onChange={handleFileChange}
            disabled={isUploading}
          />
        </div>
      </div>
    </div>
  );
}