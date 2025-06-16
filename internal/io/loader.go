// Optimized image loading and saving functionality
package io

import (
	"fmt"
	"log/slog"
	"strings"

	"gocv.io/x/gocv"
)

// ImageLoader handles image file operations
type ImageLoader struct {
	logger *slog.Logger
}

func NewImageLoader(logger *slog.Logger) *ImageLoader {
	return &ImageLoader{
		logger: logger,
	}
}

func (il *ImageLoader) LoadImage(filepath string) (gocv.Mat, error) {
	il.logger.Debug("Loading image", "filepath", filepath)

	if !il.isSupportedImageFormat(filepath) {
		return gocv.NewMat(), fmt.Errorf("unsupported image format: %s", filepath)
	}

	// Load image using OpenCV
	mat := gocv.IMRead(filepath, gocv.IMReadColor)
	if mat.Empty() {
		return gocv.NewMat(), fmt.Errorf("failed to load image: %s", filepath)
	}

	il.logger.Info("Image loaded successfully",
		"filepath", filepath,
		"width", mat.Cols(),
		"height", mat.Rows(),
		"channels", mat.Channels())

	return mat, nil
}

func (il *ImageLoader) SaveImage(mat gocv.Mat, filepath string) error {
	il.logger.Debug("Saving image", "filepath", filepath)

	if mat.Empty() {
		return fmt.Errorf("cannot save empty image")
	}

	if !il.isSupportedImageFormat(filepath) {
		return fmt.Errorf("unsupported image format: %s", filepath)
	}

	success := gocv.IMWrite(filepath, mat)
	if !success {
		return fmt.Errorf("failed to save image: %s", filepath)
	}

	il.logger.Info("Image saved successfully",
		"filepath", filepath,
		"width", mat.Cols(),
		"height", mat.Rows(),
		"channels", mat.Channels())

	return nil
}

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

func (il *ImageLoader) GetSupportedFormats() []string {
	return []string{"JPEG", "PNG", "TIFF", "BMP"}
}

func (il *ImageLoader) LoadImageGrayscale(filepath string) (gocv.Mat, error) {
	il.logger.Debug("Loading image as grayscale", "filepath", filepath)

	if !il.isSupportedImageFormat(filepath) {
		return gocv.NewMat(), fmt.Errorf("unsupported image format: %s", filepath)
	}

	mat := gocv.IMRead(filepath, gocv.IMReadGrayScale)
	if mat.Empty() {
		return gocv.NewMat(), fmt.Errorf("failed to load image: %s", filepath)
	}

	il.logger.Info("Grayscale image loaded successfully",
		"filepath", filepath,
		"width", mat.Cols(),
		"height", mat.Rows(),
		"channels", mat.Channels())

	return mat, nil
}

func (il *ImageLoader) ValidateImageFile(filepath string) error {
	if !il.isSupportedImageFormat(filepath) {
		return fmt.Errorf("unsupported image format")
	}

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
