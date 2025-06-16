// internal/gui/canvas.go
// Fixed image canvas with enhanced debugging and preview display
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
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/core"
)

// ImageCanvas handles image display with improved proportions
type ImageCanvas struct {
	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *slog.Logger

	vbox            *container.Split // Changed from *fyne.Container to *container.Split
	originalView    *widget.Card
	previewView     *widget.Card
	interactiveOrig *InteractiveCanvas
	previewImage    *canvas.Image

	activeTool         string
	onSelectionChanged func(bool)
}

func NewImageCanvas(imageData *core.ImageData, regionManager *core.RegionManager, logger *slog.Logger) *ImageCanvas {
	canvas := &ImageCanvas{
		imageData:     imageData,
		regionManager: regionManager,
		logger:        logger,
		activeTool:    "none",
	}

	canvas.initializeUI()
	return canvas
}

func (ic *ImageCanvas) initializeUI() {
	// Create interactive original image view
	ic.interactiveOrig = NewInteractiveCanvas(ic.imageData, ic.regionManager, ic.logger)
	ic.originalView = widget.NewCard("Original", "", ic.interactiveOrig)

	// Create real-time preview image view with placeholder
	placeholderImg := image.NewRGBA(image.Rect(0, 0, 200, 150))
	// Fill with light gray
	for y := 0; y < 150; y++ {
		for x := 0; x < 200; x++ {
			placeholderImg.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}

	ic.previewImage = canvas.NewImageFromImage(placeholderImg)
	ic.previewImage.FillMode = canvas.ImageFillContain
	ic.previewImage.ScaleMode = canvas.ImageScalePixels

	// Set minimum size to prevent height collapse, but allow scaling
	ic.previewImage.SetMinSize(fyne.NewSize(200, 150))

	ic.previewView = widget.NewCard("Preview", "", ic.previewImage)

	// Create vertical split container with equal split
	ic.vbox = container.NewVSplit(
		ic.originalView,
		ic.previewView,
	)
	// Set equal split (50/50)
	ic.vbox.SetOffset(0.5)
}

func (ic *ImageCanvas) GetContainer() fyne.CanvasObject {
	return ic.vbox
}

func (ic *ImageCanvas) UpdateOriginalImage() {
	if !ic.imageData.HasImage() {
		ic.logger.Debug("No image data available for original update")
		return
	}

	original := ic.imageData.GetOriginal()
	defer original.Close()

	if !original.Empty() {
		img, err := original.ToImage()
		if err != nil {
			ic.logger.Error("Failed to convert Mat to image", "error", err)
			return
		}

		ic.interactiveOrig.UpdateImage(img)
		ic.logger.Debug("Updated original image display")
	}
}

// DEPRECATED: Old method that caused thread safety issues
// Kept for compatibility but should not be used
func (ic *ImageCanvas) UpdatePreview(preview gocv.Mat) {
	ic.logger.Warn("UpdatePreview(gocv.Mat) called - this method is deprecated and unsafe")
	// Do nothing to prevent crashes - use UpdatePreviewFromImage instead
}

// NEW: Thread-safe preview update using Go image.Image
func (ic *ImageCanvas) UpdatePreviewFromImage(preview image.Image) {
	ic.logger.Debug("UpdatePreviewFromImage called",
		"width", preview.Bounds().Dx(),
		"height", preview.Bounds().Dy())

	// Additional debugging
	bounds := preview.Bounds()
	ic.logger.Debug("Preview image details",
		"bounds", bounds,
		"empty", bounds.Empty(),
		"min", bounds.Min,
		"max", bounds.Max)

	// Validate the image before setting
	if bounds.Empty() {
		ic.logger.Error("Preview image has empty bounds")
		return
	}

	// Enhanced debugging - check current state
	ic.logger.Debug("BEFORE update - current previewImage state",
		"current_image_nil", ic.previewImage.Image == nil,
		"current_file", ic.previewImage.File,
		"current_resource_nil", ic.previewImage.Resource == nil,
		"visible", ic.previewImage.Visible(),
		"size", ic.previewImage.Size(),
		"position", ic.previewImage.Position())

	// CRITICAL FIX: Clear File property and set Image directly
	// Based on Fyne issue #1886, we need to clear File before setting Image
	ic.previewImage.File = ""
	ic.previewImage.Resource = nil
	ic.previewImage.Image = preview

	ic.logger.Debug("AFTER setting - previewImage state",
		"new_image_nil", ic.previewImage.Image == nil,
		"new_file", ic.previewImage.File,
		"new_resource_nil", ic.previewImage.Resource == nil,
		"image_bounds", ic.previewImage.Image.Bounds(),
		"fill_mode", ic.previewImage.FillMode,
		"scale_mode", ic.previewImage.ScaleMode)

	// Force immediate refresh
	ic.logger.Debug("Forcing preview image refresh")
	ic.previewImage.Refresh()

	// Check if the widget is properly set up
	ic.logger.Debug("Widget hierarchy check",
		"previewImage_in_previewView", ic.previewView.Content == ic.previewImage,
		"previewView_visible", ic.previewView.Visible(),
		"split_leading", ic.vbox.Leading != nil,
		"split_trailing", ic.vbox.Trailing != nil)

	// Force refresh of parent containers
	if ic.previewView != nil {
		ic.logger.Debug("Refreshing preview card")
		ic.previewView.Refresh()
	}
	if ic.vbox != nil {
		ic.logger.Debug("Refreshing main container")
		ic.vbox.Refresh()
	}

	// Additional forced refresh to ensure display update using canvas.Refresh
	if canvas := fyne.CurrentApp().Driver().CanvasForObject(ic.previewImage); canvas != nil {
		ic.logger.Debug("Canvas found, forcing refresh")
		canvas.Refresh(ic.previewImage)
		canvas.Refresh(ic.previewView)
		canvas.Refresh(ic.vbox)
		ic.logger.Debug("Forced canvas refresh completed")
	} else {
		ic.logger.Error("NO CANVAS FOUND for previewImage - this might be the problem!")
	}

	// Try alternative approach: recreate the image widget entirely
	ic.logger.Debug("Trying alternative approach - recreating image widget")
	newImage := canvas.NewImageFromImage(preview)
	newImage.FillMode = canvas.ImageFillContain
	newImage.ScaleMode = canvas.ImageScalePixels

	// Set minimum size to prevent height collapse but allow scaling
	newImage.SetMinSize(fyne.NewSize(200, 150))

	// Replace the content of the preview card
	ic.previewView.SetContent(newImage)
	ic.previewImage = newImage

	ic.logger.Debug("Recreated image widget",
		"new_widget_image_nil", ic.previewImage.Image == nil,
		"new_widget_bounds", ic.previewImage.Image.Bounds(),
		"new_widget_size", ic.previewImage.Size(),
		"new_widget_min_size", ic.previewImage.MinSize())

	ic.logger.Debug("Preview image updated and all containers refreshed")
}

func (ic *ImageCanvas) ClearPreview() {
	ic.logger.Debug("Clearing preview")

	// Get original preview or use placeholder
	if ic.imageData.HasImage() {
		originalPreview := ic.imageData.GetPreview()
		defer originalPreview.Close()

		if !originalPreview.Empty() {
			img, err := originalPreview.ToImage()
			if err == nil {
				ic.logger.Debug("Clearing preview with original image")
				ic.UpdatePreviewFromImage(img)
				return
			}
		}
	}

	// Fall back to placeholder
	ic.logger.Debug("Clearing preview with placeholder")
	placeholderImg := image.NewRGBA(image.Rect(0, 0, 200, 150))
	for y := 0; y < 150; y++ {
		for x := 0; x < 200; x++ {
			placeholderImg.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}
	ic.UpdatePreviewFromImage(placeholderImg)
}

func (ic *ImageCanvas) SetActiveTool(tool string) {
	ic.activeTool = tool
	ic.interactiveOrig.SetActiveTool(tool)
	ic.logger.Debug("Active tool changed", "tool", tool)
}

func (ic *ImageCanvas) RefreshSelections() {
	ic.interactiveOrig.RefreshSelections()
	ic.logger.Debug("Refreshing selections")
}

func (ic *ImageCanvas) SetCallbacks(onSelectionChanged func(bool)) {
	ic.onSelectionChanged = onSelectionChanged
	ic.interactiveOrig.SetSelectionChangedCallback(onSelectionChanged)
}

func (ic *ImageCanvas) Refresh() {
	ic.vbox.Refresh()
	ic.originalView.Refresh()
	ic.previewView.Refresh()
	ic.previewImage.Refresh()
}

// Toolbar handles tool selection with improved styling
type Toolbar struct {
	hbox          *fyne.Container
	rectangleTool *widget.Button
	freehandTool  *widget.Button
	clearButton   *widget.Button
	resetButton   *widget.Button

	onToolChanged    func(string)
	onClearSelection func()
	onResetImage     func()
}

func NewToolbar() *Toolbar {
	toolbar := &Toolbar{}
	toolbar.initializeUI()
	return toolbar
}

func (t *Toolbar) initializeUI() {
	t.rectangleTool = widget.NewButton("📐 Rectangle", func() {
		if t.onToolChanged != nil {
			t.onToolChanged("rectangle")
		}
	})
	t.rectangleTool.Importance = widget.MediumImportance

	t.freehandTool = widget.NewButton("✏️ Freehand", func() {
		if t.onToolChanged != nil {
			t.onToolChanged("freehand")
		}
	})
	t.freehandTool.Importance = widget.MediumImportance

	t.clearButton = widget.NewButton("🗑️ Clear Selection", func() {
		if t.onClearSelection != nil {
			t.onClearSelection()
		}
	})
	t.clearButton.Importance = widget.LowImportance

	t.resetButton = widget.NewButton("↻ Reset to Original", func() {
		if t.onResetImage != nil {
			t.onResetImage()
		}
	})
	t.resetButton.Importance = widget.HighImportance

	t.hbox = container.NewHBox(
		widget.NewLabel("Selection Tools:"),
		t.rectangleTool,
		t.freehandTool,
		widget.NewSeparator(),
		t.clearButton,
		t.resetButton,
	)

	t.Disable()
}

func (t *Toolbar) GetContainer() fyne.CanvasObject {
	return t.hbox
}

func (t *Toolbar) Enable() {
	t.rectangleTool.Enable()
	t.freehandTool.Enable()
	t.clearButton.Enable()
	t.resetButton.Enable()
}

func (t *Toolbar) Disable() {
	t.rectangleTool.Disable()
	t.freehandTool.Disable()
	t.clearButton.Disable()
	t.resetButton.Disable()
}

func (t *Toolbar) SetSelectionState(hasSelection bool) {
	if hasSelection {
		t.clearButton.Enable()
	} else {
		t.clearButton.Disable()
	}
}

func (t *Toolbar) SetCallbacks(onToolChanged func(string), onClearSelection func()) {
	t.onToolChanged = onToolChanged
	t.onClearSelection = onClearSelection
}

func (t *Toolbar) SetResetCallback(onResetImage func()) {
	t.onResetImage = onResetImage
}

func (t *Toolbar) Refresh() {
	t.hbox.Refresh()
}

// MetricsPanel displays quality metrics with improved styling
type MetricsPanel struct {
	vbox    *fyne.Container
	metrics map[string]float64
}

func NewMetricsPanel() *MetricsPanel {
	panel := &MetricsPanel{
		metrics: make(map[string]float64),
	}

	panel.initializeUI()
	return panel
}

func (mp *MetricsPanel) initializeUI() {
	mp.vbox = container.NewVBox(
		widget.NewCard("📊 Quality Metrics", "",
			widget.NewLabel("Real-time quality metrics will appear here during processing.")),
	)
}

func (mp *MetricsPanel) GetContainer() fyne.CanvasObject {
	return mp.vbox
}

func (mp *MetricsPanel) UpdateMetrics(metrics map[string]float64) {
	mp.metrics = metrics

	content := container.NewVBox()

	if len(metrics) == 0 {
		content.Add(widget.NewLabel("🔄 Processing..."))
	} else {
		for name, value := range metrics {
			var displayText string
			var quality string

			switch name {
			case "psnr":
				displayText = fmt.Sprintf("📡 PSNR: %.2f dB", value)
				if value > 40 {
					quality = " ✅ Excellent"
				} else if value > 30 {
					quality = " ✔️ Good"
				} else if value > 20 {
					quality = " ⚠️ Fair"
				} else {
					quality = " ❌ Poor"
				}
			case "ssim":
				displayText = fmt.Sprintf("📈 SSIM: %.3f", value)
				if value > 0.95 {
					quality = " ✅ Excellent"
				} else if value > 0.8 {
					quality = " ✔️ Good"
				} else if value > 0.6 {
					quality = " ⚠️ Fair"
				} else {
					quality = " ❌ Poor"
				}
			case "mse":
				displayText = fmt.Sprintf("📊 MSE: %.2f", value)
				quality = ""
			default:
				displayText = fmt.Sprintf("📈 %s: %.3f", name, value)
				quality = ""
			}

			label := widget.NewLabel(displayText + quality)
			content.Add(label)
		}
	}

	card := widget.NewCard("📊 Quality Metrics", "", content)
	mp.vbox.RemoveAll()
	mp.vbox.Add(card)
}

func (mp *MetricsPanel) Clear() {
	mp.metrics = make(map[string]float64)
	mp.vbox.RemoveAll()
	mp.vbox.Add(widget.NewCard("📊 Quality Metrics", "",
		widget.NewLabel("Real-time quality metrics will appear here during processing.")))
}

func (mp *MetricsPanel) Refresh() {
	mp.vbox.Refresh()
}
