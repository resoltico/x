import { describe, it, expect, vi } from 'vitest'
import { ImageInputModule } from '@/modules/ImageInputModule'

describe('ImageInputModule', () => {
  describe('getInstance', () => {
    it('returns singleton instance', () => {
      const instance1 = ImageInputModule.getInstance()
      const instance2 = ImageInputModule.getInstance()
      
      expect(instance1).toBe(instance2)
    })
  })

  describe('getSupportedExtensions', () => {
    it('returns correct extensions string', () => {
      const module = ImageInputModule.getInstance()
      const extensions = module.getSupportedExtensions()
      
      expect(extensions).toBe('.png,.jpg,.jpeg,.tiff,.tif,.webp')
    })
  })

  describe('getSupportedFormats', () => {
    it('returns correct formats array', () => {
      const module = ImageInputModule.getInstance()
      const formats = module.getSupportedFormats()
      
      expect(formats).toEqual(['PNG', 'JPEG', 'TIFF', 'WebP'])
    })
  })

  describe('validateDropData', () => {
    it('validates drop data correctly', () => {
      const module = ImageInputModule.getInstance()
      
      // Mock DataTransfer with image file
      const mockDataTransfer = {
        items: [
          { kind: 'file', type: 'image/png' }
        ]
      }
      
      const mockEvent = {
        dataTransfer: mockDataTransfer
      } as DragEvent
      
      const result = module.validateDropData(mockEvent)
      expect(result).toBe(true)
    })

    it('rejects non-image files', () => {
      const module = ImageInputModule.getInstance()
      
      const mockDataTransfer = {
        items: [
          { kind: 'file', type: 'text/plain' }
        ]
      }
      
      const mockEvent = {
        dataTransfer: mockDataTransfer
      } as DragEvent
      
      const result = module.validateDropData(mockEvent)
      expect(result).toBe(false)
    })

    it('handles missing dataTransfer', () => {
      const module = ImageInputModule.getInstance()
      
      const mockEvent = {
        dataTransfer: null
      } as DragEvent
      
      const result = module.validateDropData(mockEvent)
      expect(result).toBe(false)
    })
  })

  describe('processFile', () => {
    it('processes valid file successfully', async () => {
      const module = ImageInputModule.getInstance()
      
      // Create a mock file
      const file = new File(['test content'], 'test.png', { type: 'image/png' })
      
      // Mock fileToImageData to avoid actual file processing
      const mockImageData = {
        data: new ArrayBuffer(100),
        width: 100,
        height: 100,
        channels: 4,
        format: 'PNG' as const,
        filename: 'test.png',
        size: 1024
      }

      // Mock the fileToImageData function
      vi.doMock('@/utils/imageHelpers', () => ({
        fileToImageData: vi.fn().mockResolvedValue(mockImageData)
      }))

      const result = await module.processFile(file)
      
      expect(result.success).toBe(true)
      expect(result.validation.isValid).toBe(true)
    })

    it('handles invalid file type', async () => {
      const module = ImageInputModule.getInstance()
      
      const file = new File(['test content'], 'test.txt', { type: 'text/plain' })
      
      const result = await module.processFile(file)
      
      expect(result.success).toBe(false)
      expect(result.validation.isValid).toBe(false)
      expect(result.error).toContain('Unsupported file type')
    })

    it('handles processing errors', async () => {
      const module = ImageInputModule.getInstance()
      
      const file = new File(['test content'], 'test.png', { type: 'image/png' })
      
      // Mock fileToImageData to throw an error
      vi.doMock('@/utils/imageHelpers', () => ({
        fileToImageData: vi.fn().mockRejectedValue(new Error('Processing failed'))
      }))

      const result = await module.processFile(file)
      
      expect(result.success).toBe(false)
      expect(result.error).toBe('Processing failed')
    })
  })

  describe('createDropHandler', () => {
    it('creates drop handler with success callback', () => {
      const module = ImageInputModule.getInstance()
      const onSuccess = vi.fn()
      const onError = vi.fn()
      
      const handler = module.createDropHandler(onSuccess, onError)
      
      expect(handler).toBeDefined()
      expect(handler.handleDrop).toBeDefined()
      expect(handler.handleDragOver).toBeDefined()
      expect(handler.handleDragEnter).toBeDefined()
      expect(handler.handleDragLeave).toBeDefined()
    })

    it('handles drop with no files', async () => {
      const module = ImageInputModule.getInstance()
      const onSuccess = vi.fn()
      const onError = vi.fn()
      
      const handler = module.createDropHandler(onSuccess, onError)
      
      const mockEvent = {
        preventDefault: vi.fn(),
        dataTransfer: { files: [] }
      } as unknown as DragEvent
      
      await handler.handleDrop(mockEvent)
      
      expect(onError).toHaveBeenCalledWith('No files detected in drop')
      expect(onSuccess).not.toHaveBeenCalled()
    })

    it('handles drop with multiple files', async () => {
      const module = ImageInputModule.getInstance()
      const onSuccess = vi.fn()
      const onError = vi.fn()
      
      const handler = module.createDropHandler(onSuccess, onError)
      
      const file1 = new File(['content1'], 'test1.png', { type: 'image/png' })
      const file2 = new File(['content2'], 'test2.png', { type: 'image/png' })
      
      const mockEvent = {
        preventDefault: vi.fn(),
        dataTransfer: { files: [file1, file2] }
      } as unknown as DragEvent
      
      await handler.handleDrop(mockEvent)
      
      expect(onError).toHaveBeenCalledWith('Please drop only one file at a time')
      expect(onSuccess).not.toHaveBeenCalled()
    })
  })
})