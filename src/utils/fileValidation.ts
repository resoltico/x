import type { FileValidation, ImageFormat } from '@/types'

// Supported file types and their MIME types
const SUPPORTED_TYPES: Record<string, ImageFormat> = {
  'image/png': 'PNG',
  'image/jpeg': 'JPEG',
  'image/jpg': 'JPEG',
  'image/tiff': 'TIFF',
  'image/tif': 'TIFF',
  'image/webp': 'WebP'
}

// Maximum file size (10MB)
const MAX_FILE_SIZE = 10 * 1024 * 1024

/**
 * Validates an uploaded file for image processing
 */
export const validateImageFile = (file: File): FileValidation => {
  const warnings: string[] = []

  // Check file type
  if (!SUPPORTED_TYPES[file.type]) {
    return {
      isValid: false,
      error: `Unsupported file type: ${file.type}. Supported formats: PNG, JPEG, TIFF, WebP`
    }
  }

  // Check file size
  if (file.size > MAX_FILE_SIZE) {
    return {
      isValid: false,
      error: `File size exceeds maximum limit of ${formatFileSize(MAX_FILE_SIZE)}`
    }
  }

  // Warning for very large files
  if (file.size > 5 * 1024 * 1024) {
    warnings.push(`Large file detected (${formatFileSize(file.size)}). Processing may take longer.`)
  }

  // Warning for TIFF files (can be complex)
  if (file.type.includes('tiff')) {
    warnings.push('TIFF files may have variable support depending on compression and color space.')
  }

  return {
    isValid: true,
    warnings: warnings.length > 0 ? warnings : undefined
  }
}

/**
 * Determines the image format from a file
 */
export const getImageFormat = (file: File): ImageFormat => {
  return SUPPORTED_TYPES[file.type] || 'PNG'
}

/**
 * Formats file size in human-readable format
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 Bytes'
  
  const k = 1024
  const sizes = ['Bytes', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * Creates a preview-sized version of image data for faster processing
 */
export const createPreviewSize = (width: number, height: number, maxDimension = 800) => {
  if (width <= maxDimension && height <= maxDimension) {
    return { width, height, scale: 1 }
  }

  const scale = Math.min(maxDimension / width, maxDimension / height)
  return {
    width: Math.round(width * scale),
    height: Math.round(height * scale),
    scale
  }
}

/**
 * Validates image dimensions for processing
 */
export const validateImageDimensions = (width: number, height: number): FileValidation => {
  const warnings: string[] = []

  // Minimum dimensions
  if (width < 32 || height < 32) {
    return {
      isValid: false,
      error: 'Image dimensions too small. Minimum size is 32x32 pixels.'
    }
  }

  // Maximum dimensions (for memory safety)
  const maxDimension = 8192
  if (width > maxDimension || height > maxDimension) {
    return {
      isValid: false,
      error: `Image dimensions too large. Maximum size is ${maxDimension}x${maxDimension} pixels.`
    }
  }

  // Warning for very large images
  if (width * height > 4096 * 4096) {
    warnings.push('Large image detected. Processing may require significant memory and time.')
  }

  // Warning for unusual aspect ratios
  const aspectRatio = Math.max(width, height) / Math.min(width, height)
  if (aspectRatio > 10) {
    warnings.push('Unusual aspect ratio detected. Some algorithms may not work optimally.')
  }

  return {
    isValid: true,
    warnings: warnings.length > 0 ? warnings : undefined
  }
}