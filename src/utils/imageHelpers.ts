import type { ImageData, ImageFormat } from '@/types'

/**
 * Converts a File to ImageData format
 */
export const fileToImageData = async (file: File): Promise<ImageData> => {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    
    reader.onload = async (event) => {
      try {
        const arrayBuffer = event.target?.result as ArrayBuffer
        const imageElement = new Image()
        
        imageElement.onload = () => {
          const canvas = document.createElement('canvas')
          const ctx = canvas.getContext('2d')!
          
          canvas.width = imageElement.width
          canvas.height = imageElement.height
          
          ctx.drawImage(imageElement, 0, 0)
          
          // Get image data from canvas
          const imageDataFromCanvas = ctx.getImageData(0, 0, canvas.width, canvas.height)
          
          resolve({
            data: arrayBuffer,
            width: imageElement.width,
            height: imageElement.height,
            channels: 4, // RGBA
            format: getFormatFromMimeType(file.type),
            filename: file.name,
            size: file.size
          })
        }
        
        imageElement.onerror = () => reject(new Error('Failed to load image'))
        imageElement.src = URL.createObjectURL(file)
      } catch (error) {
        reject(error)
      }
    }
    
    reader.onerror = () => reject(new Error('Failed to read file'))
    reader.readAsArrayBuffer(file)
  })
}

/**
 * Converts ArrayBuffer to a downloadable blob
 */
export const arrayBufferToBlob = (buffer: ArrayBuffer, format: ImageFormat): Blob => {
  const mimeType = formatToMimeType(format)
  return new Blob([buffer], { type: mimeType })
}

/**
 * Creates a download link for processed image
 */
export const downloadImage = (buffer: ArrayBuffer, filename: string, format: ImageFormat) => {
  const blob = arrayBufferToBlob(buffer, format)
  const url = URL.createObjectURL(blob)
  
  const link = document.createElement('a')
  link.href = url
  link.download = `processed_${filename}`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  // Clean up object URL
  setTimeout(() => URL.revokeObjectURL(url), 100)
}

/**
 * Renders ImageData to a canvas element
 */
export const renderImageToCanvas = (
  imageData: ImageData, 
  canvas: HTMLCanvasElement,
  options: {
    zoom?: number
    offsetX?: number
    offsetY?: number
    fitToCanvas?: boolean
  } = {}
) => {
  const ctx = canvas.getContext('2d')!
  const { zoom = 1, offsetX = 0, offsetY = 0, fitToCanvas = false } = options
  
  // Clear canvas
  ctx.clearRect(0, 0, canvas.width, canvas.height)
  
  // Create image element from data
  const img = new Image()
  img.onload = () => {
    let drawWidth = img.width
    let drawHeight = img.height
    let drawX = offsetX
    let drawY = offsetY
    
    if (fitToCanvas) {
      // Calculate scale to fit image in canvas
      const scaleX = canvas.width / img.width
      const scaleY = canvas.height / img.height
      const scale = Math.min(scaleX, scaleY)
      
      drawWidth = img.width * scale
      drawHeight = img.height * scale
      drawX = (canvas.width - drawWidth) / 2
      drawY = (canvas.height - drawHeight) / 2
    } else {
      drawWidth *= zoom
      drawHeight *= zoom
    }
    
    ctx.drawImage(img, drawX, drawY, drawWidth, drawHeight)
  }
  
  // Convert ArrayBuffer to object URL
  const blob = new Blob([imageData.data])
  img.src = URL.createObjectURL(blob)
}

/**
 * Gets the center point of an image for zoom operations
 */
export const getImageCenter = (imageData: ImageData, canvas: HTMLCanvasElement) => {
  return {
    x: canvas.width / 2,
    y: canvas.height / 2
  }
}

/**
 * Calculates optimal zoom level to fit image in canvas
 */
export const calculateFitZoom = (imageData: ImageData, canvas: HTMLCanvasElement): number => {
  const scaleX = canvas.width / imageData.width
  const scaleY = canvas.height / imageData.height
  return Math.min(scaleX, scaleY, 1) // Don't zoom in beyond 100%
}

/**
 * Converts image format to MIME type
 */
const formatToMimeType = (format: ImageFormat): string => {
  const mimeTypes: Record<ImageFormat, string> = {
    'PNG': 'image/png',
    'JPEG': 'image/jpeg',
    'TIFF': 'image/tiff',
    'WebP': 'image/webp'
  }
  return mimeTypes[format]
}

/**
 * Converts MIME type to image format
 */
const getFormatFromMimeType = (mimeType: string): ImageFormat => {
  if (mimeType.includes('png')) return 'PNG'
  if (mimeType.includes('jpeg') || mimeType.includes('jpg')) return 'JPEG'
  if (mimeType.includes('tiff') || mimeType.includes('tif')) return 'TIFF'
  if (mimeType.includes('webp')) return 'WebP'
  return 'PNG' // Default fallback
}

/**
 * Creates a thumbnail from image data
 */
export const createThumbnail = (
  imageData: ImageData, 
  maxSize: number = 200
): Promise<string> => {
  return new Promise((resolve) => {
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')!
    
    // Calculate thumbnail dimensions
    const scale = Math.min(maxSize / imageData.width, maxSize / imageData.height)
    canvas.width = imageData.width * scale
    canvas.height = imageData.height * scale
    
    const img = new Image()
    img.onload = () => {
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
      resolve(canvas.toDataURL('image/png'))
    }
    
    const blob = new Blob([imageData.data])
    img.src = URL.createObjectURL(blob)
  })
}