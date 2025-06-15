// Author: Ervins Strauhmanis
// License: MIT

package models

import (
	"sync"

	"gocv.io/x/gocv"
)

// ImageData holds original and processed image matrices
type ImageData struct {
	mu        sync.RWMutex
	original  gocv.Mat
	processed gocv.Mat
	hasImage  bool
}

// NewImageData creates a new ImageData instance
func NewImageData() *ImageData {
	return &ImageData{
		hasImage: false,
	}
}

// SetOriginal sets the original image matrix
func (img *ImageData) SetOriginal(mat gocv.Mat) {
	img.mu.Lock()
	defer img.mu.Unlock()

	if !img.original.Empty() {
		img.original.Close()
	}
	
	img.original = mat.Clone()
	img.hasImage = true
}

// SetProcessed sets the processed image matrix
func (img *ImageData) SetProcessed(mat gocv.Mat) {
	img.mu.Lock()
	defer img.mu.Unlock()

	if !img.processed.Empty() {
		img.processed.Close()
	}
	
	img.processed = mat.Clone()
}

// GetOriginal returns a copy of the original image matrix
func (img *ImageData) GetOriginal() gocv.Mat {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if img.original.Empty() {
		return gocv.NewMat()
	}
	return img.original.Clone()
}

// GetProcessed returns a copy of the processed image matrix
func (img *ImageData) GetProcessed() gocv.Mat {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if img.processed.Empty() {
		return gocv.NewMat()
	}
	return img.processed.Clone()
}

// HasImage returns true if an image has been loaded
func (img *ImageData) HasImage() bool {
	img.mu.RLock()
	defer img.mu.RUnlock()
	return img.hasImage
}

// Clear clears both original and processed images
func (img *ImageData) Clear() {
	img.mu.Lock()
	defer img.mu.Unlock()

	if !img.original.Empty() {
		img.original.Close()
	}
	if !img.processed.Empty() {
		img.processed.Close()
	}
	
	img.hasImage = false
}

// GetDimensions returns the dimensions of the original image
func (img *ImageData) GetDimensions() (int, int) {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if img.original.Empty() {
		return 0, 0
	}
	return img.original.Cols(), img.original.Rows()
}