// Core image data structure with thread-safe operations
package core

import (
	"fmt"
	"sync"

	"gocv.io/x/gocv"
)

// ImageData manages original and processed images with thread safety
type ImageData struct {
	mu           sync.RWMutex
	original     gocv.Mat
	processed    gocv.Mat
	hasImage     bool
	filepath     string
	metadata     ImageMetadata
}

// ImageMetadata contains image information
type ImageMetadata struct {
	Width    int
	Height   int
	Channels int
	Type     gocv.MatType
	Format   string
	Size     int64 // File size in bytes
}

// NewImageData creates a new thread-safe image data container
func NewImageData() *ImageData {
	return &ImageData{
		original:  gocv.NewMat(),
		processed: gocv.NewMat(),
		hasImage:  false,
	}
}

// SetOriginal sets the original image with validation
func (img *ImageData) SetOriginal(mat gocv.Mat, filepath string) error {
	img.mu.Lock()
	defer img.mu.Unlock()

	// Validate input
	if mat.Empty() {
		return fmt.Errorf("cannot set empty image")
	}

	// Validate image properties
	if mat.Cols() <= 0 || mat.Rows() <= 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", mat.Cols(), mat.Rows())
	}

	channels := mat.Channels()
	if channels != 1 && channels != 3 && channels != 4 {
		return fmt.Errorf("unsupported number of channels: %d", channels)
	}

	// Close existing images
	if !img.original.Empty() {
		img.original.Close()
	}
	if !img.processed.Empty() {
		img.processed.Close()
	}

	// Clone and store the image
	img.original = mat.Clone()
	img.processed = mat.Clone() // Start with original as processed
	img.hasImage = true
	img.filepath = filepath

	// Store metadata
	img.metadata = ImageMetadata{
		Width:    mat.Cols(),
		Height:   mat.Rows(),
		Channels: channels,
		Type:     mat.Type(),
		Format:   getFormatFromPath(filepath),
	}

	return nil
}

// SetProcessed sets the processed image
func (img *ImageData) SetProcessed(mat gocv.Mat) error {
	img.mu.Lock()
	defer img.mu.Unlock()

	if !img.hasImage {
		return fmt.Errorf("no original image loaded")
	}

	if mat.Empty() {
		return fmt.Errorf("cannot set empty processed image")
	}

	// Close existing processed image
	if !img.processed.Empty() {
		img.processed.Close()
	}

	img.processed = mat.Clone()
	return nil
}

// GetOriginal returns a copy of the original image
func (img *ImageData) GetOriginal() gocv.Mat {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if !img.hasImage || img.original.Empty() {
		return gocv.NewMat()
	}
	return img.original.Clone()
}

// GetProcessed returns a copy of the processed image
func (img *ImageData) GetProcessed() gocv.Mat {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if img.processed.Empty() {
		return gocv.NewMat()
	}
	return img.processed.Clone()
}

// HasImage returns true if an image is loaded
func (img *ImageData) HasImage() bool {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return img.hasImage
}

// GetMetadata returns image metadata
func (img *ImageData) GetMetadata() ImageMetadata {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return img.metadata
}

// GetFilepath returns the current file path
func (img *ImageData) GetFilepath() string {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return img.filepath
}

// Clear clears all image data
func (img *ImageData) Clear() {
	img.mu.Lock()
	defer img.mu.Unlock()

	if !img.original.Empty() {
		img.original.Close()
	}
	if !img.processed.Empty() {
		img.processed.Close()
	}

	img.original = gocv.NewMat()
	img.processed = gocv.NewMat()
	img.hasImage = false
	img.filepath = ""
	img.metadata = ImageMetadata{}
}

// Close releases all resources
func (img *ImageData) Close() {
	img.Clear()
}

// ResetToOriginal resets processed image to original
func (img *ImageData) ResetToOriginal() error {
	img.mu.Lock()
	defer img.mu.Unlock()

	if !img.hasImage || img.original.Empty() {
		return fmt.Errorf("no original image available")
	}

	// Close existing processed image
	if !img.processed.Empty() {
		img.processed.Close()
	}

	// Reset to original
	img.processed = img.original.Clone()
	return nil
}

// getFormatFromPath extracts image format from file path
func getFormatFromPath(filepath string) string {
	if filepath == "" {
		return "unknown"
	}
	
	// Extract extension
	for i := len(filepath) - 1; i >= 0; i-- {
		if filepath[i] == '.' {
			return filepath[i+1:]
		}
	}
	return "unknown"
}

// ValidateImage validates an OpenCV Mat for basic requirements
func ValidateImage(mat gocv.Mat) error {
	if mat.Empty() {
		return fmt.Errorf("image is empty")
	}

	if mat.Cols() <= 0 || mat.Rows() <= 0 {
		return fmt.Errorf("invalid dimensions: %dx%d", mat.Cols(), mat.Rows())
	}

	channels := mat.Channels()
	if channels < 1 || channels > 4 {
		return fmt.Errorf("unsupported channel count: %d", channels)
	}

	// Check for reasonable size limits (prevent memory issues)
	const maxDimension = 16384
	if mat.Cols() > maxDimension || mat.Rows() > maxDimension {
		return fmt.Errorf("image too large: %dx%d (max: %d)", mat.Cols(), mat.Rows(), maxDimension)
	}

	return nil
}
