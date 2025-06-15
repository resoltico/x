// Author: Ervins Strauhmanis
// License: MIT

package models

import (
	"sync"
	"unsafe"

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
		hasImage:  false,
		original:  gocv.NewMat(),
		processed: gocv.NewMat(),
	}
}

// SetOriginal sets the original image matrix
func (img *ImageData) SetOriginal(mat gocv.Mat) {
	img.mu.Lock()
	defer img.mu.Unlock()

	// Close existing original if it exists and is valid
	if !img.original.Empty() {
		img.original.Close()
	}

	// Check if the incoming mat has a valid internal pointer
	// The issue is that sometimes a Mat with nil internal pointer is passed
	if !isMatPointerValid(mat) {
		img.original = gocv.NewMat()
		img.hasImage = false
		return
	}

	// Now it's safe to call Empty() since we verified the pointer
	if mat.Empty() {
		img.original = gocv.NewMat()
		img.hasImage = false
		return
	}

	img.original = mat.Clone()
	img.hasImage = true
}

// SetProcessed sets the processed image matrix
func (img *ImageData) SetProcessed(mat gocv.Mat) {
	img.mu.Lock()
	defer img.mu.Unlock()

	// Close existing processed if it exists and is valid
	if !img.processed.Empty() {
		img.processed.Close()
	}

	// Check if the incoming mat has a valid internal pointer
	if !isMatPointerValid(mat) {
		img.processed = gocv.NewMat()
		return
	}

	// Now it's safe to call Empty()
	if mat.Empty() {
		img.processed = gocv.NewMat()
		return
	}

	img.processed = mat.Clone()
}

// GetOriginal returns a copy of the original image matrix
func (img *ImageData) GetOriginal() gocv.Mat {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if !img.hasImage || img.original.Empty() {
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

	img.original = gocv.NewMat()
	img.processed = gocv.NewMat()
	img.hasImage = false
}

// GetDimensions returns the dimensions of the original image
func (img *ImageData) GetDimensions() (int, int) {
	img.mu.RLock()
	defer img.mu.RUnlock()

	if !img.hasImage || img.original.Empty() {
		return 0, 0
	}
	return img.original.Cols(), img.original.Rows()
}

// isMatPointerValid checks if the Mat's internal C pointer is valid
// GoCV Mat wraps a C++ cv::Mat pointer, and if this is nil, method calls segfault
func isMatPointerValid(mat gocv.Mat) bool {
	// Get the internal pointer using unsafe operations
	// GoCV Mat struct has a 'p' field that's a C pointer
	// We need to check if this pointer is nil before calling any methods

	// Use unsafe to access the first field of the Mat struct (the C pointer)
	matPtr := unsafe.Pointer(&mat)
	if matPtr == nil {
		return false
	}

	// The first field in the GoCV Mat struct is the C pointer 'p'
	// If this is nil (0x0), then the Mat is invalid
	cPtr := *(*unsafe.Pointer)(matPtr)
	return cPtr != nil
}
