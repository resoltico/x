import { describe, it, expect } from 'vitest'
import { 
  validateImageFile, 
  getImageFormat, 
  formatFileSize, 
  createPreviewSize,
  validateImageDimensions 
} from '@/utils/fileValidation'

describe('File Validation Utils', () => {
  describe('validateImageFile', () => {
    it('accepts valid PNG file', () => {
      const file = new File(['dummy content'], 'test.png', { type: 'image/png' })
      const result = validateImageFile(file)
      
      expect(result.isValid).toBe(true)
      expect(result.error).toBeUndefined()
    })

    it('accepts valid JPEG file', () => {
      const file = new File(['dummy content'], 'test.jpg', { type: 'image/jpeg' })
      const result = validateImageFile(file)
      
      expect(result.isValid).toBe(true)
      expect(result.error).toBeUndefined()
    })

    it('rejects unsupported file type', () => {
      const file = new File(['dummy content'], 'test.gif', { type: 'image/gif' })
      const result = validateImageFile(file)
      
      expect(result.isValid).toBe(false)
      expect(result.error).toContain('Unsupported file type')
    })

    it('rejects file that is too large', () => {
      // Create a mock file that's larger than 10MB
      const largeContent = new Array(11 * 1024 * 1024).fill('x').join('')
      const file = new File([largeContent], 'large.png', { type: 'image/png' })
      const result = validateImageFile(file)
      
      expect(result.isValid).toBe(false)
      expect(result.error).toContain('File size exceeds maximum limit')
    })

    it('warns about large files', () => {
      // Create a mock file that's larger than 5MB but smaller than 10MB
      const largeContent = new Array(6 * 1024 * 1024).fill('x').join('')
      const file = new File([largeContent], 'large.png', { type: 'image/png' })
      const result = validateImageFile(file)
      
      expect(result.isValid).toBe(true)
      expect(result.warnings).toBeDefined()
      expect(result.warnings?.[0]).toContain('Large file detected')
    })

    it('warns about TIFF files', () => {
      const file = new File(['dummy content'], 'test.tiff', { type: 'image/tiff' })
      const result = validateImageFile(file)
      
      expect(result.isValid).toBe(true)
      expect(result.warnings).toBeDefined()
      expect(result.warnings?.[0]).toContain('TIFF files may have variable support')
    })
  })

  describe('getImageFormat', () => {
    it('returns correct format for PNG', () => {
      const file = new File(['content'], 'test.png', { type: 'image/png' })
      expect(getImageFormat(file)).toBe('PNG')
    })

    it('returns correct format for JPEG', () => {
      const file = new File(['content'], 'test.jpg', { type: 'image/jpeg' })
      expect(getImageFormat(file)).toBe('JPEG')
    })

    it('returns PNG as default for unknown types', () => {
      const file = new File(['content'], 'test.unknown', { type: 'image/unknown' })
      expect(getImageFormat(file)).toBe('PNG')
    })
  })

  describe('formatFileSize', () => {
    it('formats bytes correctly', () => {
      expect(formatFileSize(0)).toBe('0 Bytes')
      expect(formatFileSize(1024)).toBe('1 KB')
      expect(formatFileSize(1024 * 1024)).toBe('1 MB')
      expect(formatFileSize(1024 * 1024 * 1024)).toBe('1 GB')
    })

    it('formats fractional sizes correctly', () => {
      expect(formatFileSize(1536)).toBe('1.5 KB')
      expect(formatFileSize(1024 * 1024 * 1.5)).toBe('1.5 MB')
    })
  })

  describe('createPreviewSize', () => {
    it('returns original size when within max dimension', () => {
      const result = createPreviewSize(500, 400, 800)
      expect(result).toEqual({ width: 500, height: 400, scale: 1 })
    })

    it('scales down when exceeding max dimension', () => {
      const result = createPreviewSize(1600, 800, 800)
      expect(result.width).toBe(800)
      expect(result.height).toBe(400)
      expect(result.scale).toBe(0.5)
    })

    it('maintains aspect ratio when scaling', () => {
      const result = createPreviewSize(2000, 1000, 500)
      expect(result.width).toBe(500)
      expect(result.height).toBe(250)
      expect(result.scale).toBe(0.25)
    })
  })

  describe('validateImageDimensions', () => {
    it('accepts valid dimensions', () => {
      const result = validateImageDimensions(1024, 768)
      expect(result.isValid).toBe(true)
      expect(result.error).toBeUndefined()
    })

    it('rejects dimensions that are too small', () => {
      const result = validateImageDimensions(16, 16)
      expect(result.isValid).toBe(false)
      expect(result.error).toContain('Image dimensions too small')
    })

    it('rejects dimensions that are too large', () => {
      const result = validateImageDimensions(10000, 10000)
      expect(result.isValid).toBe(false)
      expect(result.error).toContain('Image dimensions too large')
    })

    it('warns about very large images', () => {
      const result = validateImageDimensions(5000, 5000)
      expect(result.isValid).toBe(true)
      expect(result.warnings).toBeDefined()
      expect(result.warnings?.[0]).toContain('Large image detected')
    })

    it('warns about unusual aspect ratios', () => {
      const result = validateImageDimensions(5000, 100)
      expect(result.isValid).toBe(true)
      expect(result.warnings).toBeDefined()
      expect(result.warnings?.[0]).toContain('Unusual aspect ratio detected')
    })
  })
})