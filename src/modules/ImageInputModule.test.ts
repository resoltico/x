import { describe, it, expect, vi, beforeEach } from 'vitest'
import { ImageInputModule } from '@/modules/ImageInputModule'

// Mock the fileToImageData utility function
vi.mock('@/utils/imageHelpers', () => ({
  fileToImageData: vi.fn()
}))

describe('ImageInputModule', () => {
  let mockFileToImageData: any

  beforeEach(async () => {
    // Reset all mocks before each test
    vi.clearAllMocks()
    
    // Get the mocked function
    const imageHelpers = await import('@/utils/imageHelpers')
    mockFileToImageData = imageHelpers.fileToImageData
  })

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
      
      // Create a proper mock DragEvent
      const mockEvent = {
        dataTransfer: {
          items: [
            { kind: 'file', type: 'image/png' }
          ]
        },
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        type: 'dragover',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      const result = module.validateDropData(mockEvent)
      expect(result).toBe(true)
    })

    it('rejects non-image files', () => {
      const module = ImageInputModule.getInstance()
      
      const mockEvent = {
        dataTransfer: {
          items: [
            { kind: 'file', type: 'text/plain' }
          ]
        },
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        type: 'dragover',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      const result = module.validateDropData(mockEvent)
      expect(result).toBe(false)
    })

    it('handles missing dataTransfer', () => {
      const module = ImageInputModule.getInstance()
      
      const mockEvent = {
        dataTransfer: null,
        preventDefault: vi.fn(),
        stopPropagation: vi.fn(),
        type: 'dragover',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      const result = module.validateDropData(mockEvent)
      expect(result).toBe(false)
    })
  })

  describe('processFile', () => {
    it('processes valid file successfully', async () => {
      const module = ImageInputModule.getInstance()
      
      // Create a mock file
      const file = new File(['test content'], 'test.png', { type: 'image/png' })
      
      // Mock the fileToImageData function to return success
      const mockImageData = {
        data: new ArrayBuffer(100),
        width: 100,
        height: 100,
        channels: 4,
        format: 'PNG' as const,
        filename: 'test.png',
        size: 1024
      }

      mockFileToImageData.mockResolvedValue(mockImageData)

      const result = await module.processFile(file)
      
      expect(result.success).toBe(true)
      expect(result.validation.isValid).toBe(true)
      expect(result.imageData).toEqual(mockImageData)
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
      mockFileToImageData.mockRejectedValue(new Error('Processing failed'))

      const result = await module.processFile(file)
      
      expect(result.success).toBe(false)
      expect(result.error).toBe('Processing failed')
    })

    it('handles large file warning', async () => {
      const module = ImageInputModule.getInstance()
      
      // Create a large file (6MB)
      const largeContent = new Array(6 * 1024 * 1024).fill('x').join('')
      const file = new File([largeContent], 'large.png', { type: 'image/png' })
      
      const mockImageData = {
        data: new ArrayBuffer(100),
        width: 100,
        height: 100,
        channels: 4,
        format: 'PNG' as const,
        filename: 'large.png',
        size: file.size
      }

      mockFileToImageData.mockResolvedValue(mockImageData)

      const result = await module.processFile(file)
      
      expect(result.success).toBe(true)
      expect(result.validation.isValid).toBe(true)
      expect(result.validation.warnings).toBeDefined()
      expect(result.validation.warnings![0]).toContain('Large file detected')
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
        dataTransfer: { files: [] },
        type: 'drop',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopPropagation: vi.fn(),
        stopImmediatePropagation: vi.fn()
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
        dataTransfer: { files: [file1, file2] },
        type: 'drop',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopPropagation: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      await handler.handleDrop(mockEvent)
      
      expect(onError).toHaveBeenCalledWith('Please drop only one file at a time')
      expect(onSuccess).not.toHaveBeenCalled()
    })

    it('handles successful file drop', async () => {
      const module = ImageInputModule.getInstance()
      const onSuccess = vi.fn()
      const onError = vi.fn()
      
      const handler = module.createDropHandler(onSuccess, onError)
      
      const file = new File(['content'], 'test.png', { type: 'image/png' })
      
      const mockImageData = {
        data: new ArrayBuffer(100),
        width: 100,
        height: 100,
        channels: 4,
        format: 'PNG' as const,
        filename: 'test.png',
        size: 1024
      }

      mockFileToImageData.mockResolvedValue(mockImageData)
      
      const mockEvent = {
        preventDefault: vi.fn(),
        dataTransfer: { files: [file] },
        type: 'drop',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopPropagation: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      await handler.handleDrop(mockEvent)
      
      expect(onSuccess).toHaveBeenCalledWith(mockImageData, undefined)
      expect(onError).not.toHaveBeenCalled()
    })

    it('handles file drop processing error', async () => {
      const module = ImageInputModule.getInstance()
      const onSuccess = vi.fn()
      const onError = vi.fn()
      
      const handler = module.createDropHandler(onSuccess, onError)
      
      const file = new File(['content'], 'test.png', { type: 'image/png' })

      mockFileToImageData.mockRejectedValue(new Error('Drop processing failed'))
      
      const mockEvent = {
        preventDefault: vi.fn(),
        dataTransfer: { files: [file] },
        type: 'drop',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopPropagation: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      await handler.handleDrop(mockEvent)
      
      expect(onError).toHaveBeenCalledWith('Drop processing failed')
      expect(onSuccess).not.toHaveBeenCalled()
    })

    it('handles drag over event', () => {
      const module = ImageInputModule.getInstance()
      const onSuccess = vi.fn()
      const onError = vi.fn()
      
      const handler = module.createDropHandler(onSuccess, onError)
      
      const mockEvent = {
        preventDefault: vi.fn(),
        dataTransfer: { dropEffect: 'none' },
        type: 'dragover',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopPropagation: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      handler.handleDragOver(mockEvent)
      
      expect(mockEvent.preventDefault).toHaveBeenCalled()
      expect((mockEvent.dataTransfer as any).dropEffect).toBe('copy')
    })

    it('handles drag enter event', () => {
      const module = ImageInputModule.getInstance()
      const onSuccess = vi.fn()
      const onError = vi.fn()
      
      const handler = module.createDropHandler(onSuccess, onError)
      
      const mockEvent = {
        preventDefault: vi.fn(),
        type: 'dragenter',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopPropagation: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      handler.handleDragEnter(mockEvent)
      
      expect(mockEvent.preventDefault).toHaveBeenCalled()
    })

    it('handles drag leave event', () => {
      const module = ImageInputModule.getInstance()
      const onSuccess = vi.fn()
      const onError = vi.fn()
      
      const handler = module.createDropHandler(onSuccess, onError)
      
      const mockEvent = {
        preventDefault: vi.fn(),
        type: 'dragleave',
        bubbles: false,
        cancelable: true,
        composed: false,
        currentTarget: null,
        defaultPrevented: false,
        eventPhase: 0,
        isTrusted: true,
        target: null,
        timeStamp: Date.now(),
        initEvent: vi.fn(),
        stopPropagation: vi.fn(),
        stopImmediatePropagation: vi.fn()
      } as unknown as DragEvent
      
      handler.handleDragLeave(mockEvent)
      
      expect(mockEvent.preventDefault).toHaveBeenCalled()
    })
  })

  describe('processFiles (batch processing)', () => {
    it('processes multiple files successfully', async () => {
      const module = ImageInputModule.getInstance()
      
      const file1 = new File(['content1'], 'test1.png', { type: 'image/png' })
      const file2 = new File(['content2'], 'test2.png', { type: 'image/png' })
      const files = [file1, file2] as unknown as FileList
      
      const mockImageData1 = {
        data: new ArrayBuffer(100),
        width: 100,
        height: 100,
        channels: 4,
        format: 'PNG' as const,
        filename: 'test1.png',
        size: 1024
      }
      
      const mockImageData2 = {
        data: new ArrayBuffer(200),
        width: 200,
        height: 200,
        channels: 4,
        format: 'PNG' as const,
        filename: 'test2.png',
        size: 2048
      }

      mockFileToImageData
        .mockResolvedValueOnce(mockImageData1)
        .mockResolvedValueOnce(mockImageData2)

      const result = await module.processFiles(files)
      
      expect(result.successful).toHaveLength(2)
      expect(result.failed).toHaveLength(0)
      expect(result.warnings).toHaveLength(0)
      expect(result.successful[0]).toEqual(mockImageData1)
      expect(result.successful[1]).toEqual(mockImageData2)
    })

    it('handles mixed success and failure in batch processing', async () => {
      const module = ImageInputModule.getInstance()
      
      const file1 = new File(['content1'], 'test1.png', { type: 'image/png' })
      const file2 = new File(['content2'], 'test2.txt', { type: 'text/plain' }) // Invalid type
      const files = [file1, file2] as unknown as FileList
      
      const mockImageData1 = {
        data: new ArrayBuffer(100),
        width: 100,
        height: 100,
        channels: 4,
        format: 'PNG' as const,
        filename: 'test1.png',
        size: 1024
      }

      mockFileToImageData.mockResolvedValueOnce(mockImageData1)

      const result = await module.processFiles(files)
      
      expect(result.successful).toHaveLength(1)
      expect(result.failed).toHaveLength(1)
      expect(result.successful[0]).toEqual(mockImageData1)
      expect(result.failed[0].file).toBe(file2)
      expect(result.failed[0].error).toContain('Unsupported file type')
    })
  })
})