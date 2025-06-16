// Core image data structure with thread-safe operations
package core

import (
	"fmt"
	"image"
	"sync"

	"gocv.io/x/gocv"
)

// ImageData manages original and processed images with thread safety
type ImageData struct {
	mu        sync.RWMutex
	original  gocv.Mat
	processed gocv.Mat
	preview   gocv.Mat // Low-res preview for real-time processing
	hasImage  bool
	filepath  string
	metadata  ImageMetadata
}

// ImageMetadata contains image information
type ImageMetadata struct {
	Width    int
	Height   int
	Channels int
	Type     gocv.MatType
	Format   string
	Size     int64
}

const PreviewMaxSize = 400 // Max width/height for preview

func NewImageData() *ImageData {
	return &ImageData{
		original:  gocv.NewMat(),
		processed: gocv.NewMat(),
		preview:   gocv.NewMat(),
		hasImage:  false,
	}
}

func (img *ImageData) SetOriginal(mat gocv.Mat, filepath string) error {
	img.mu.Lock()
	defer img.mu.Unlock()

	if mat.Empty() {
		return fmt.Errorf("cannot set empty image")
	}

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
	if !img.preview.Empty() {
		img.preview.Close()
	}

	// Clone and store the image
	img.original = mat.Clone()
	img.processed = mat.Clone()
	img.hasImage = true
	img.filepath = filepath

	// Create preview
	img.preview = img.createPreview(mat)

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

func (img *ImageData) createPreview(mat gocv.Mat) gocv.Mat {
	if mat.Empty() {
		return gocv.NewMat()
	}

	width := mat.Cols()
	height := mat.Rows()

	// Calculate scale factor
	scale := 1.0
	if width > PreviewMaxSize || height > PreviewMaxSize {
		scaleW := float64(PreviewMaxSize) / float64(width)
		scaleH := float64(PreviewMaxSize) / float64(height)
		if scaleW < scaleH {
			scale = scaleW
		} else {
			scale = scaleH
		}
	}

	if scale >= 1.0 {
		return mat.Clone()
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	preview := gocv.NewMat()
	gocv.Resize(mat, &preview, image.Pt(newWidth, newHeight), 0, 0, gocv.InterpolationArea)
	return preview
}

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

func (img *ImageData) GetOriginal() gocv.Mat {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if !img.hasImage || img.original.Empty() {
		return gocv.NewMat()
	}
	return img.original.Clone()
}

func (img *ImageData) GetProcessed() gocv.Mat {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if img.processed.Empty() {
		return gocv.NewMat()
	}
	return img.processed.Clone()
}

func (img *ImageData) GetPreview() gocv.Mat {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if img.preview.Empty() {
		return gocv.NewMat()
	}
	return img.preview.Clone()
}

func (img *ImageData) HasImage() bool {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return img.hasImage
}

func (img *ImageData) GetMetadata() ImageMetadata {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return img.metadata
}

func (img *ImageData) GetFilepath() string {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return img.filepath
}

func (img *ImageData) Clear() {
	img.mu.Lock()
	defer img.mu.Unlock()

	if !img.original.Empty() {
		img.original.Close()
	}
	if !img.processed.Empty() {
		img.processed.Close()
	}
	if !img.preview.Empty() {
		img.preview.Close()
	}

	img.original = gocv.NewMat()
	img.processed = gocv.NewMat()
	img.preview = gocv.NewMat()
	img.hasImage = false
	img.filepath = ""
	img.metadata = ImageMetadata{}
}

func (img *ImageData) Close() {
	img.Clear()
}

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

func getFormatFromPath(filepath string) string {
	if filepath == "" {
		return "unknown"
	}

	for i := len(filepath) - 1; i >= 0; i-- {
		if filepath[i] == '.' {
			return filepath[i+1:]
		}
	}
	return "unknown"
}

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

	const maxDimension = 16384
	if mat.Cols() > maxDimension || mat.Rows() > maxDimension {
		return fmt.Errorf("image too large: %dx%d (max: %d)", mat.Cols(), mat.Rows(), maxDimension)
	}

	return nil
}
