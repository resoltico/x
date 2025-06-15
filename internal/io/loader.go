// Image loading and saving functionality
package io

import (
	"fmt"
	"strings"

	"gocv.io/x/gocv"
	"github.com/sirupsen/logrus"
)

// ImageLoader handles image file operations
type ImageLoader struct {
	logger *logrus.Logger
}

// NewImageLoader creates a new image loader
func NewImageLoader(logger *logrus.Logger) *ImageLoader {
	return &ImageLoader{
		logger: logger,
	}
}

// LoadImage loads an image from file
func (il *ImageLoader) LoadImage(filepath string) (gocv.Mat, error) {
	il.logger.WithField("filepath", filepath).Debug("Loading image")
	
	// Validate file extension
	if !il.isSupportedImageFormat(filepath) {
		return gocv.NewMat(), fmt.Errorf("unsupported image format: %s", filepath)
	}
	
	// Load image using OpenCV
	mat := gocv.IMRead(filepath, gocv.IMReadColor)
	if mat.Empty() {
		return fmt.Errorf("cannot save empty image")
	}
	
	// Validate file extension
	if !il.isSupportedImageFormat(filepath) {
		return fmt.Errorf("unsupported image format: %s", filepath)
	}
	
	// Save image using OpenCV
	success := gocv.IMWrite(filepath, mat)
	if !success {
		return fmt.Errorf("failed to save image: %s", filepath)
	}
	
	il.logger.WithFields(logrus.Fields{
		"filepath": filepath,
		"width":    mat.Cols(),
		"height":   mat.Rows(),
		"channels": mat.Channels(),
	}).Info("Image saved successfully")
	
	return nil
}

// isSupportedImageFormat checks if the file format is supported
func (il *ImageLoader) isSupportedImageFormat(filepath string) bool {
	ext := strings.ToLower(getFileExtension(filepath))
	supportedFormats := []string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"}
	
	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}
	
	return false
}

// getFileExtension extracts the file extension from a filepath
func getFileExtension(filepath string) string {
	for i := len(filepath) - 1; i >= 0; i-- {
		if filepath[i] == '.' {
			return filepath[i:]
		}
		if filepath[i] == '/' || filepath[i] == '\\' {
			break
		}
	}
	return ""
}

// GetSupportedFormats returns a list of supported image formats
func (il *ImageLoader) GetSupportedFormats() []string {
	return []string{"JPEG", "PNG", "TIFF", "BMP"}
}

// LoadImageGrayscale loads an image as grayscale
func (il *ImageLoader) LoadImageGrayscale(filepath string) (gocv.Mat, error) {
	il.logger.WithField("filepath", filepath).Debug("Loading image as grayscale")
	
	// Validate file extension
	if !il.isSupportedImageFormat(filepath) {
		return gocv.NewMat(), fmt.Errorf("unsupported image format: %s", filepath)
	}
	
	// Load image as grayscale using OpenCV
	mat := gocv.IMRead(filepath, gocv.IMReadGrayScale)
	if mat.Empty() {
		return gocv.NewMat(), fmt.Errorf("failed to load image: %s", filepath)
	}
	
	il.logger.WithFields(logrus.Fields{
		"filepath": filepath,
		"width":    mat.Cols(),
		"height":   mat.Rows(),
		"channels": mat.Channels(),
	}).Info("Grayscale image loaded successfully")
	
	return mat, nil
}

// ValidateImageFile checks if a file is a valid image
func (il *ImageLoader) ValidateImageFile(filepath string) error {
	if !il.isSupportedImageFormat(filepath) {
		return fmt.Errorf("unsupported image format")
	}
	
	// Try to load the image header to validate
	mat := gocv.IMRead(filepath, gocv.IMReadGrayScale)
	defer mat.Close()
	
	if mat.Empty() {
		return fmt.Errorf("invalid or corrupted image file")
	}
	
	if mat.Cols() <= 0 || mat.Rows() <= 0 {
		return fmt.Errorf("invalid image dimensions")
	}
	
	return nil
}
		return gocv.NewMat(), fmt.Errorf("failed to load image: %s", filepath)
	}
	
	il.logger.WithFields(logrus.Fields{
		"filepath": filepath,
		"width":    mat.Cols(),
		"height":   mat.Rows(),
		"channels": mat.Channels(),
	}).Info("Image loaded successfully")
	
	return mat, nil
}

// SaveImage saves an image to file
func (il *ImageLoader) SaveImage(mat gocv.Mat, filepath string) error {
	il.logger.WithField("filepath", filepath).Debug("Saving image")
	
	if mat.Empty() {
		