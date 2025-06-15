// Author: Ervins Strauhmanis
// License: MIT

package gui

import (
	"fmt"
	"image"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/models"
)

// Preview handles the display of original and processed images
type Preview struct {
	mu        sync.RWMutex
	imageData *models.ImageData
	logger    *logrus.Logger

	// GUI components
	container      *container.AppTabs
	originalImage  *canvas.Image
	processedImage *canvas.Image
	originalLabel  *widget.Label
	processedLabel *widget.Label
	tabs           *container.AppTabs

	// State
	hasImage bool
}

// NewPreview creates a new image preview component
func NewPreview(imageData *models.ImageData, logger *logrus.Logger) *Preview {
	p := &Preview{
		imageData: imageData,
		logger:    logger,
		hasImage:  false,
	}

	p.initializeComponents()
	p.setupLayout()

	return p
}

// initializeComponents initializes the preview components
func (p *Preview) initializeComponents() {
	// Create image canvases
	p.originalImage = canvas.NewImageFromResource(nil)
	p.originalImage.FillMode = canvas.ImageFillContain
	p.originalImage.SetMinSize(fyne.NewSize(400, 300))

	p.processedImage = canvas.NewImageFromResource(nil)
	p.processedImage.FillMode = canvas.ImageFillContain
	p.processedImage.SetMinSize(fyne.NewSize(400, 300))

	// Create labels
	p.originalLabel = widget.NewLabel("No image loaded")
	p.processedLabel = widget.NewLabel("No processing applied")
}

// setupLayout creates the preview layout
func (p *Preview) setupLayout() {
	// Original image tab
	originalTab := container.NewBorder(
		p.originalLabel, // top
		nil,             // bottom
		nil,             // left
		nil,             // right
		container.NewScroll(p.originalImage),
	)

	// Processed image tab
	processedTab := container.NewBorder(
		p.processedLabel, // top
		nil,              // bottom
		nil,              // left
		nil,              // right
		container.NewScroll(p.processedImage),
	)

	// Create tab container
	p.tabs = container.NewAppTabs(
		container.NewTabItem("Original", originalTab),
		container.NewTabItem("Processed", processedTab),
	)

	p.container = p.tabs
}

// UpdateOriginal updates the original image display
func (p *Preview) UpdateOriginal() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.imageData.HasImage() {
		p.clearImages()
		return
	}

	// Get original image
	original := p.imageData.GetOriginal()
	defer original.Close()

	if original.Empty() {
		p.clearImages()
		return
	}

	// Convert to displayable format
	img, err := p.matToImage(original)
	if err != nil {
		p.logger.WithError(err).Error("Failed to convert original image for display")
		return
	}

	// Update display
	p.originalImage.Image = img
	p.originalImage.Refresh()

	// Update label
	width, height := p.imageData.GetDimensions()
	p.originalLabel.SetText(fmt.Sprintf("Original Image (%dx%d)", width, height))

	p.hasImage = true

	p.logger.Debug("Updated original image preview")
}

// UpdateProcessed updates the processed image display
func (p *Preview) UpdateProcessed(processed gocv.Mat) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if processed.Empty() {
		p.processedImage.Image = nil
		p.processedImage.Refresh()
		p.processedLabel.SetText("Processing failed")
		return
	}

	// Convert to displayable format
	img, err := p.matToImage(processed)
	if err != nil {
		p.logger.WithError(err).Error("Failed to convert processed image for display")
		return
	}

	// Update display
	p.processedImage.Image = img
	p.processedImage.Refresh()

	// Update label
	p.processedLabel.SetText(fmt.Sprintf("Processed Image (%dx%d)", processed.Cols(), processed.Rows()))

	// Switch to processed tab to show result
	if p.hasImage {
		p.tabs.SelectIndex(1)
	}

	p.logger.Debug("Updated processed image preview")
}

// matToImage converts a gocv.Mat to a Go image.Image
func (p *Preview) matToImage(mat gocv.Mat) (image.Image, error) {
	if mat.Empty() {
		return nil, fmt.Errorf("empty matrix")
	}

	// Convert to RGB if needed
	var displayMat gocv.Mat
	defer func() {
		if !displayMat.Empty() && displayMat.Ptr() != mat.Ptr() {
			displayMat.Close()
		}
	}()

	if mat.Channels() == 1 {
		// Grayscale to RGB
		displayMat = gocv.NewMat()
		gocv.CvtColor(mat, &displayMat, gocv.ColorGrayToRGB)
	} else if mat.Channels() == 3 {
		// BGR to RGB
		displayMat = gocv.NewMat()
		gocv.CvtColor(mat, &displayMat, gocv.ColorBGRToRGB)
	} else {
		displayMat = mat
	}

	// Convert to Go image
	img, err := displayMat.ToImage()
	if err != nil {
		return nil, fmt.Errorf("failed to convert matrix to image: %w", err)
	}

	return img, nil
}

// clearImages clears both image displays
func (p *Preview) clearImages() {
	p.originalImage.Image = nil
	p.originalImage.Refresh()
	p.processedImage.Image = nil
	p.processedImage.Refresh()

	p.originalLabel.SetText("No image loaded")
	p.processedLabel.SetText("No processing applied")

	p.hasImage = false
}

// GetContainer returns the preview container
func (p *Preview) GetContainer() *container.AppTabs {
	return p.container
}

// SwitchToOriginal switches the view to the original image
func (p *Preview) SwitchToOriginal() {
	p.tabs.SelectIndex(0)
}

// SwitchToProcessed switches the view to the processed image
func (p *Preview) SwitchToProcessed() {
	p.tabs.SelectIndex(1)
}

// HasImage returns true if an image is currently loaded
func (p *Preview) HasImage() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.hasImage
}
