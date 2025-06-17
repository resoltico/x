// internal/gui/center_panel.go
// Perfect UI Center Panel: Image Display Area
package gui

import (
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

	// Create image canvases
	cp.originalImage = canvas.NewImageFromImage(placeholder)
	cp.originalImage.FillMode = canvas.ImageFillContain
	cp.originalImage.ScaleMode = canvas.ImageScalePixels

	cp.previewImage = canvas.NewImageFromImage(placeholder)
	cp.previewImage.FillMode = canvas.ImageFillContain
	cp.previewImage.ScaleMode = canvas.ImageScalePixels

	// Create view containers
	cp.setupViewContainers()

	// Start with split view
	cp.container = container.NewBorder(nil, nil, nil, nil, cp.splitView)
}

func (cp *CenterPanel) setupViewContainers() {
	// Single view - preview only
	previewCard := widget.NewCard("Preview", "", cp.previewImage)
	cp.singleView = container.NewBorder(nil, nil, nil, nil, previewCard)

	// Split view - original and preview side by side
	originalCard := widget.NewCard("Original", "", cp.originalImage)
	previewCardSplit := widget.NewCard("Preview", "", cp.previewImage)
	cp.splitView = container.NewHSplit(originalCard, previewCardSplit)
	cp.splitView.SetOffset(0.5)

	// Overlay view - with opacity slider for blending
	overlaySlider := widget.NewSlider(0, 1)
	overlaySlider.SetValue(0.5)
	overlaySlider.OnChanged = func(value float64) {
		// TODO: Implement overlay blending
	}

	overlayContainer := container.NewVBox(
		container.NewStack(cp.originalImage, cp.previewImage),
		container.NewHBox(
			widget.NewLabel("Opacity:"),
			overlaySlider,
		),
	)
	overlayCard := widget.NewCard("Overlay", "", overlayContainer)
	cp.overlayView = container.NewBorder(nil, nil, nil, nil, overlayCard)
}

func (cp *CenterPanel) createPlaceholderImage() image.Image {
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	gray := color.RGBA{245, 245, 245, 255} // #F5F5F5

	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, gray)
		}
	}

	return placeholder
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
		cp.container.Add(cp.splitView)
	}

	cp.container.Refresh()
}

func (cp *CenterPanel) SetZoom(zoom float64) {
	cp.currentZoom = zoom
	// Note: Zoom functionality would require custom image widgets
	// For now, we'll use the built-in scaling
	cp.originalImage.Refresh()
	cp.previewImage.Refresh()
}

func (cp *CenterPanel) UpdateOriginal() {
	if !cp.imageData.HasImage() {
		return
	}

	original := cp.imageData.GetOriginal()
	defer original.Close()

	if !original.Empty() {
		if img, err := original.ToImage(); err == nil {
			cp.originalImage.Image = img
			cp.originalImage.Refresh()
		}
	}
}

func (cp *CenterPanel) UpdatePreview(preview image.Image) {
	cp.previewImage.Image = preview
	cp.previewImage.Refresh()
}

func (cp *CenterPanel) Reset() {
	placeholder := cp.createPlaceholderImage()
	cp.originalImage.Image = placeholder
	cp.previewImage.Image = placeholder
	cp.originalImage.Refresh()
	cp.previewImage.Refresh()
}

func (cp *CenterPanel) GetContainer() fyne.CanvasObject {
	return cp.container
}

func (cp *CenterPanel) SetCallbacks(onSelectionChanged func(bool)) {
	cp.onSelectionChanged = onSelectionChanged
}
