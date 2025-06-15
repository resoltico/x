// Author: Ervins Strauhmanis
// License: MIT

package image_processing

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
)

// ImageLoader handles loading and saving of images
type ImageLoader struct {
	logger *logrus.Logger
}

// NewImageLoader creates a new image loader
func NewImageLoader(logger *logrus.Logger) *ImageLoader {
	return &ImageLoader{
		logger: logger,
	}
}

// LoadImage loads an image from the specified file path
func (il *ImageLoader) LoadImage(filepath string) (gocv.Mat, error) {
	il.logger.WithField("filepath", filepath).Info("Loading image")

	// Check if file extension is supported
	if !il.isSupportedFormat(filepath) {
		return gocv.NewMat(), fmt.Errorf("unsupported file format: %s", filepath)
	}

	// Load the image
	mat := gocv.IMRead(filepath, gocv.IMReadColor)
	if mat.Empty() {
		return gocv.NewMat(), fmt.Errorf("failed to load image: %s (file may be corrupted or not exist)", filepath)
	}

	il.logger.WithFields(logrus.Fields{
		"filepath": filepath,
		"width":    mat.Cols(),
		"height":   mat.Rows(),
		"channels": mat.Channels(),
	}).Info("Image loaded successfully")

	return mat, nil
}

// SaveImage saves an image to the specified file path
func (il *ImageLoader) SaveImage(mat gocv.Mat, filepath string) error {
	if mat.Empty() {
		return fmt.Errorf("cannot save empty image")
	}

	il.logger.WithField("filepath", filepath).Info("Saving image")

	// Ensure the output format is PNG for best compatibility
	if !strings.HasSuffix(strings.ToLower(filepath), ".png") {
		ext := filepath[strings.LastIndex(filepath, "."):]
		filepath = strings.TrimSuffix(filepath, ext) + ".png"
	}

	// Save the image
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

// isSupportedFormat checks if the file format is supported for loading
func (il *ImageLoader) isSupportedFormat(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	supportedFormats := []string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"}

	for _, format := range supportedFormats {
		if ext == format {
			return true
		}
	}
	return false
}

// GetSupportedFormats returns a list of supported file formats
func (il *ImageLoader) GetSupportedFormats() []string {
	return []string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp", ".JPG", ".JPEG", ".PNG", ".TIFF", ".TIF", ".BMP"}
}

// ValidateImage performs basic validation on a loaded image
func (il *ImageLoader) ValidateImage(mat gocv.Mat) error {
	if mat.Empty() {
		return fmt.Errorf("image is empty")
	}

	// Check dimensions
	if mat.Cols() <= 0 || mat.Rows() <= 0 {
		return fmt.Errorf("invalid image dimensions: %dx%d", mat.Cols(), mat.Rows())
	}

	// Check if image is too large (arbitrary limit for performance)
	maxDimension := 8192
	if mat.Cols() > maxDimension || mat.Rows() > maxDimension {
		return fmt.Errorf("image too large: %dx%d (max: %d)", mat.Cols(), mat.Rows(), maxDimension)
	}

	// Check channels
	channels := mat.Channels()
	if channels != 1 && channels != 3 && channels != 4 {
		return fmt.Errorf("unsupported number of channels: %d", channels)
	}

	// Check data type
	matType := mat.Type()
	if matType != gocv.MatTypeCV8UC1 && matType != gocv.MatTypeCV8UC3 && matType != gocv.MatTypeCV8UC4 {
		il.logger.WithField("type", matType).Warn("Unusual image data type detected")
	}

	il.logger.WithFields(logrus.Fields{
		"width":    mat.Cols(),
		"height":   mat.Rows(),
		"channels": channels,
		"type":     matType,
	}).Debug("Image validation passed")

	return nil
}
