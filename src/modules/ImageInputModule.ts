import type { ImageData, FileValidation } from '@/types'
import { validateImageFile, validateImageDimensions } from '@/utils/fileValidation'
import { fileToImageData } from '@/utils/imageHelpers'

/**
 * ImageInputModule handles file upload, validation, and conversion to ImageData
 */
export class ImageInputModule {
  private static instance: ImageInputModule
  
  static getInstance(): ImageInputModule {
    if (!ImageInputModule.instance) {
      ImageInputModule.instance = new ImageInputModule()
    }
    return ImageInputModule.instance
  }

  /**
   * Validates and processes a file input
   */
  async processFile(file: File): Promise<{
    success: boolean
    imageData?: ImageData
    validation: FileValidation
    error?: string
  }> {
    try {
      // Validate file
      const fileValidation = validateImageFile(file)
      if (!fileValidation.isValid) {
        return {
          success: false,
          validation: fileValidation,
          error: fileValidation.error
        }
      }

      // Convert to ImageData
      const imageData = await fileToImageData(file)
      
      // Validate dimensions
      const dimensionValidation = validateImageDimensions(imageData.width, imageData.height)
      if (!dimensionValidation.isValid) {
        return {
          success: false,
          validation: dimensionValidation,
          error: dimensionValidation.error
        }
      }

      // Combine warnings from both validations
      const allWarnings = [
        ...(fileValidation.warnings || []),
        ...(dimensionValidation.warnings || [])
      ]

      return {
        success: true,
        imageData,
        validation: {
          isValid: true,
          warnings: allWarnings.length > 0 ? allWarnings : undefined
        }
      }
    } catch (error) {
      return {
        success: false,
        validation: { isValid: false },
        error: error instanceof Error ? error.message : 'Unknown error occurred'
      }
    }
  }

  /**
   * Handles multiple file inputs (for future batch processing)
   */
  async processFiles(files: FileList): Promise<{
    successful: ImageData[]
    failed: Array<{ file: File; error: string }>
    warnings: string[]
  }> {
    const successful: ImageData[] = []
    const failed: Array<{ file: File; error: string }> = []
    const warnings: string[] = []

    for (const file of Array.from(files)) {
      const result = await this.processFile(file)
      
      if (result.success && result.imageData) {
        successful.push(result.imageData)
        if (result.validation.warnings) {
          warnings.push(...result.validation.warnings)
        }
      } else {
        failed.push({
          file,
          error: result.error || 'Unknown error'
        })
      }
    }

    return { successful, failed, warnings }
  }

  /**
   * Creates a drag and drop handler
   */
  createDropHandler(
    onSuccess: (imageData: ImageData, warnings?: string[]) => void,
    onError: (error: string) => void
  ) {
    return {
      handleDrop: async (event: DragEvent) => {
        event.preventDefault()
        
        const files = event.dataTransfer?.files
        if (!files || files.length === 0) {
          onError('No files detected in drop')
          return
        }

        if (files.length > 1) {
          onError('Please drop only one file at a time')
          return
        }

        const result = await this.processFile(files[0])
        if (result.success && result.imageData) {
          onSuccess(result.imageData, result.validation.warnings)
        } else {
          onError(result.error || 'Failed to process file')
        }
      },

      handleDragOver: (event: DragEvent) => {
        event.preventDefault()
        event.dataTransfer!.dropEffect = 'copy'
      },

      handleDragEnter: (event: DragEvent) => {
        event.preventDefault()
      },

      handleDragLeave: (event: DragEvent) => {
        event.preventDefault()
      }
    }
  }

  /**
   * Validates drag and drop data types
   */
  validateDropData(event: DragEvent): boolean {
    if (!event.dataTransfer) return false

    const items = Array.from(event.dataTransfer.items)
    return items.some(item => 
      item.kind === 'file' && 
      item.type.startsWith('image/')
    )
  }

  /**
   * Gets supported file extensions for file input
   */
  getSupportedExtensions(): string {
    return '.png,.jpg,.jpeg,.tiff,.tif,.webp'
  }

  /**
   * Gets human-readable list of supported formats
   */
  getSupportedFormats(): string[] {
    return ['PNG', 'JPEG', 'TIFF', 'WebP']
  }
}