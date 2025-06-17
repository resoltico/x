// internal/gui/center_panel.go
// Perfect UI Center Panel: Image Display Area with synchronized views
package gui

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/core"
)

type CenterPanel struct {
	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *slog.Logger

	container *fyne.Container

	// Image displays
	originalImage *canvas.Image
	previewImage  *canvas.Image

	// View containers
	singleView  *fyne.Container
	splitView   *container.Split
	overlayView *fyne.Container

	// Overlay controls
	opacitySlider *widget.Slider
	overlayStack  *fyne.Container

	// Current state
	currentZoom float64
	viewMode    string

	// Callbacks
	onSelectionChanged func(bool)
}

func NewCenterPanel(imageData *core.ImageData, regionManager *core.RegionManager, logger *slog.Logger) *CenterPanel {
	panel := &CenterPanel{
		imageData:     imageData,
		regionManager: regionManager,
		logger:        logger,
		currentZoom:   1.0,
		viewMode:      "split",
	}

	panel.initializeUI()
	return panel
}

func (cp *CenterPanel) initializeUI() {
	// Create placeholder images
	placeholder := cp.createPlaceholderImage()

	// Create image canvases with proper sizing
	cp.originalImage = canvas.NewImageFromImage(placeholder)
	cp.originalImage.FillMode = canvas.ImageFillContain
	cp.originalImage.ScaleMode = canvas.ImageScalePixels
	// CRITICAL: Set minimum size so image is visible
	cp.originalImage.SetMinSize(fyne.NewSize(400, 400))

	cp.previewImage = canvas.NewImageFromImage(placeholder)
	cp.previewImage.FillMode = canvas.ImageFillContain
	cp.previewImage.ScaleMode = canvas.ImageScalePixels
	// CRITICAL: Set minimum size so image is visible
	cp.previewImage.SetMinSize(fyne.NewSize(400, 400))

	cp.logger.Debug("Initialized image canvases with min size", "size", "400x400")

	// Create view containers following Perfect UI specification
	cp.setupViewContainers()

	// Start with split view as default
	cp.container = container.NewBorder(nil, nil, nil, nil, cp.splitView)
}

func (cp *CenterPanel) setupViewContainers() {
	// Single view - preview only with "Preview" label
	singleLabel := widget.NewLabelWithStyle("Preview", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	singleContent := container.NewBorder(
		singleLabel,     // top
		nil,             // bottom
		nil,             // left
		nil,             // right
		cp.previewImage, // center
	)
	cp.singleView = singleContent

	// Split view - original and preview side by side with labels
	originalLabel := widget.NewLabelWithStyle("Original", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	previewLabel := widget.NewLabelWithStyle("Preview", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	originalContent := container.NewBorder(
		originalLabel,    // top
		nil,              // bottom
		nil,              // left
		nil,              // right
		cp.originalImage, // center
	)

	previewContent := container.NewBorder(
		previewLabel,    // top
		nil,             // bottom
		nil,             // left
		nil,             // right
		cp.previewImage, // center
	)

	cp.splitView = container.NewHSplit(originalContent, previewContent)
	cp.splitView.SetOffset(0.5) // Equal split as per specification

	// Overlay view - with opacity slider for blending
	cp.opacitySlider = widget.NewSlider(0, 1)
	cp.opacitySlider.SetValue(0.5)
	cp.opacitySlider.Resize(fyne.NewSize(100, 20))
	cp.opacitySlider.OnChanged = func(value float64) {
		cp.updateOverlayOpacity(value)
	}

	opacityLabel := widget.NewLabel("Opacity:")
	opacityControls := container.NewHBox(
		opacityLabel,
		cp.opacitySlider,
		widget.NewLabel("50%"),
	)

	// Create stack for overlay with both images
	cp.overlayStack = container.NewStack(cp.originalImage, cp.previewImage)

	overlayContent := container.NewBorder(
		widget.NewLabelWithStyle("Overlay", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), // top
		opacityControls, // bottom
		nil,             // left
		nil,             // right
		cp.overlayStack, // center
	)

	cp.overlayView = overlayContent
}

func (cp *CenterPanel) createPlaceholderImage() image.Image {
	// Create 400x400 placeholder with Perfect UI colors
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 400))
	gray := color.RGBA{211, 211, 211, 255} // #D3D3D3 as per specification

	// Fill with gray background
	for y := 0; y < 400; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, gray)
		}
	}

	// Add "No image loaded" text area (simplified)
	textArea := color.RGBA{245, 245, 245, 255} // #F5F5F5
	for y := 180; y < 220; y++ {
		for x := 120; x < 280; x++ {
			if x < 400 && y < 400 {
				placeholder.Set(x, y, textArea)
			}
		}
	}

	return placeholder
}

func (cp *CenterPanel) updateOverlayOpacity(opacity float64) {
	// Update opacity percentage label
	overlayContainer := cp.overlayView.Objects[0].(*fyne.Container)
	opacityControls := overlayContainer.Objects[2].(*fyne.Container)
	percentLabel := opacityControls.Objects[2].(*widget.Label)
	percentLabel.SetText(fmt.Sprintf("%.0f%%", opacity*100))

	// Apply opacity to preview image in overlay
	// Note: Fyne doesn't have direct opacity control, this is a placeholder
	// In a real implementation, you'd need custom rendering or image blending
}

func (cp *CenterPanel) SetViewMode(mode string) {
	cp.viewMode = mode

	// Remove current view
	cp.container.RemoveAll()

	// Add new view based on mode
	switch mode {
	case "single":
		cp.container.Add(cp.singleView)
	case "split":
		cp.container.Add(cp.splitView)
	case "overlay":
		cp.container.Add(cp.overlayView)
	default:
		cp.container.Add(cp.splitView) // Default to split view
	}

	cp.container.Refresh()
	cp.logger.Debug("View mode changed", "mode", mode)
}

func (cp *CenterPanel) SetZoom(zoom float64) {
	cp.currentZoom = zoom

	// Apply zoom scaling to images
	// Note: This is simplified - real zoom would require custom image widgets
	// with pan and scroll capabilities

	cp.originalImage.Refresh()
	cp.previewImage.Refresh()

	cp.logger.Debug("Zoom changed", "zoom", zoom)
}

func (cp *CenterPanel) UpdateOriginal() {
	if !cp.imageData.HasImage() {
		cp.logger.Debug("No image to update original display")
		return
	}

	original := cp.imageData.GetOriginal()
	defer original.Close()

	if !original.Empty() {
		cp.logger.Debug("Converting original Mat to image.Image", "size", fmt.Sprintf("%dx%d", original.Cols(), original.Rows()))

		if img, err := original.ToImage(); err == nil {
			cp.logger.Debug("Successfully converted Mat to image.Image", "bounds", img.Bounds())

			// Set the image directly and refresh
			cp.originalImage.Image = img
			cp.originalImage.Refresh()

			cp.logger.Debug("Original image set and refreshed", "fill_mode", cp.originalImage.FillMode, "min_size", cp.originalImage.MinSize())
		} else {
			cp.logger.Error("Failed to convert original to image", "error", err)
		}
	} else {
		cp.logger.Error("Original Mat is empty")
	}
}

func (cp *CenterPanel) UpdatePreview(preview image.Image) {
	if preview == nil {
		cp.logger.Debug("Preview is nil, skipping update")
		return
	}

	cp.logger.Debug("Updating preview image", "bounds", preview.Bounds())

	// Set the image and refresh
	cp.previewImage.Image = preview
	cp.previewImage.Refresh()

	cp.logger.Debug("Preview image set and refreshed", "fill_mode", cp.previewImage.FillMode, "min_size", cp.previewImage.MinSize())
}

func (cp *CenterPanel) Reset() {
	placeholder := cp.createPlaceholderImage()

	// Reset both images to placeholder
	cp.originalImage.Image = placeholder
	cp.previewImage.Image = placeholder

	// Force refresh both images
	cp.originalImage.Refresh()
	cp.previewImage.Refresh()

	// Reset zoom to 100%
	cp.currentZoom = 1.0

	// Reset to split view
	cp.SetViewMode("split")

	cp.logger.Debug("Center panel reset to defaults")
}

func (cp *CenterPanel) GetContainer() fyne.CanvasObject {
	return cp.container
}

func (cp *CenterPanel) SetCallbacks(onSelectionChanged func(bool)) {
	cp.onSelectionChanged = onSelectionChanged
}
