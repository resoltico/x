// internal/gui/center_panel.go
// Perfect UI Center Panel: Image Display Area with debug integration
package gui

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"time"

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
	// Create better placeholder images
	originalPlaceholder := cp.createPlaceholderImage("Load an image to start processing", color.RGBA{240, 248, 255, 255})
	previewPlaceholder := cp.createPlaceholderImage("Preview will appear here", color.RGBA{248, 255, 240, 255})

	// Create image canvases
	cp.originalImage = canvas.NewImageFromImage(originalPlaceholder)
	cp.originalImage.FillMode = canvas.ImageFillContain
	cp.originalImage.ScaleMode = canvas.ImageScalePixels
	cp.originalImage.SetMinSize(fyne.NewSize(400, 400))

	cp.previewImage = canvas.NewImageFromImage(previewPlaceholder)
	cp.previewImage.FillMode = canvas.ImageFillContain
	cp.previewImage.ScaleMode = canvas.ImageScalePixels
	cp.previewImage.SetMinSize(fyne.NewSize(400, 400))

	cp.logger.Debug("Initialized image canvases with better placeholders")

	// Create view containers
	cp.setupViewContainers()

	// Start with split view as default
	cp.container = container.NewBorder(nil, nil, nil, nil, cp.splitView)
}

func (cp *CenterPanel) setupViewContainers() {
	// Single view - preview only
	singleLabel := widget.NewLabelWithStyle("Preview", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	singleContent := container.NewBorder(
		singleLabel,     // top
		nil,             // bottom
		nil,             // left
		nil,             // right
		cp.previewImage, // center
	)
	cp.singleView = singleContent

	// Split view - original and preview side by side
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
	cp.splitView.SetOffset(0.5)

	// Overlay view
	cp.opacitySlider = widget.NewSlider(0, 1)
	cp.opacitySlider.SetValue(0.5)
	cp.opacitySlider.OnChanged = func(value float64) {
		cp.updateOverlayOpacity(value)
	}

	opacityLabel := widget.NewLabel("Opacity:")
	opacityControls := container.NewHBox(
		opacityLabel,
		cp.opacitySlider,
		widget.NewLabel("50%"),
	)

	cp.overlayStack = container.NewStack(cp.originalImage, cp.previewImage)

	overlayContent := container.NewBorder(
		widget.NewLabelWithStyle("Overlay", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		opacityControls,
		nil,
		nil,
		cp.overlayStack,
	)

	cp.overlayView = overlayContent
}

func (cp *CenterPanel) createPlaceholderImage(text string, bgColor color.RGBA) image.Image {
	// Create 400x400 placeholder
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 400))

	// Fill with background color
	for y := 0; y < 400; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, bgColor)
		}
	}

	// Add dashed border
	borderColor := color.RGBA{180, 180, 180, 255}
	for i := 0; i < 400; i += 10 {
		// Top and bottom borders
		for j := 0; j < 5; j++ {
			if i+j < 400 {
				placeholder.Set(i+j, 10, borderColor)
				placeholder.Set(i+j, 390, borderColor)
			}
		}
		// Left and right borders
		for j := 0; j < 5; j++ {
			if i+j < 400 {
				placeholder.Set(10, i+j, borderColor)
				placeholder.Set(390, i+j, borderColor)
			}
		}
	}

	// Add centered icon area (simple rectangle)
	iconColor := color.RGBA{200, 200, 200, 255}
	for y := 160; y < 240; y++ {
		for x := 160; x < 240; x++ {
			placeholder.Set(x, y, iconColor)
		}
	}

	// Add text background area
	textBgColor := color.RGBA{255, 255, 255, 200}
	for y := 260; y < 280; y++ {
		for x := 100; x < 300; x++ {
			placeholder.Set(x, y, textBgColor)
		}
	}

	return placeholder
}

func (cp *CenterPanel) updateOverlayOpacity(opacity float64) {
	overlayContainer := cp.overlayView.Objects[0].(*fyne.Container)
	opacityControls := overlayContainer.Objects[2].(*fyne.Container)
	percentLabel := opacityControls.Objects[2].(*widget.Label)
	percentLabel.SetText(fmt.Sprintf("%.0f%%", opacity*100))
}

func (cp *CenterPanel) SetViewMode(mode string) {
	start := time.Now()
	cp.viewMode = mode
	cp.container.RemoveAll()

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
	cp.logger.Debug("View mode changed", "mode", mode)

	// DEBUG: Log view mode change
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("CenterPanel", "view_mode_change", map[string]interface{}{
			"new_mode": mode,
			"duration": duration,
		})
	}
}

func (cp *CenterPanel) SetZoom(zoom float64) {
	start := time.Now()
	cp.currentZoom = zoom
	cp.originalImage.Refresh()
	cp.previewImage.Refresh()
	cp.logger.Debug("Zoom changed", "zoom", zoom)

	// DEBUG: Log zoom change
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("CenterPanel", "zoom_change", map[string]interface{}{
			"zoom":     zoom,
			"duration": duration,
		})
	}
}

func (cp *CenterPanel) UpdateOriginal() {
	start := time.Now()
	cp.logger.Debug("DISPLAY: UpdateOriginal called")

	// DEBUG: Log original update attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogImageOperation("update_original_attempt", true, map[string]interface{}{
			"has_image": cp.imageData.HasImage(),
		})
	}

	if !cp.imageData.HasImage() {
		cp.logger.Debug("DISPLAY: No image to update original display")

		// DEBUG: Log no image available
		if GlobalGUIDebugger != nil {
			GlobalGUIDebugger.LogImageOperation("update_original_no_image", false, map[string]interface{}{
				"reason": "no_image_loaded",
			})
		}
		return
	}

	original := cp.imageData.GetOriginal()
	defer original.Close()

	if original.Empty() {
		cp.logger.Error("DISPLAY: Original Mat is empty")

		// DEBUG: Log empty mat
		if GlobalGUIDebugger != nil {
			GlobalGUIDebugger.LogImageOperation("update_original_empty_mat", false, map[string]interface{}{
				"reason": "empty_opencv_mat",
			})
		}
		return
	}

	cp.logger.Debug("DISPLAY: Converting original Mat to image.Image",
		"size", fmt.Sprintf("%dx%d", original.Cols(), original.Rows()))

	img, err := original.ToImage()
	if err != nil {
		cp.logger.Error("DISPLAY: Failed to convert original to image", "error", err)

		// DEBUG: Log conversion failure
		if GlobalGUIDebugger != nil {
			duration := time.Since(start)
			GlobalGUIDebugger.LogImageOperation("update_original_conversion_failed", false, map[string]interface{}{
				"error":    err.Error(),
				"duration": duration,
			})
		}
		return
	}

	cp.logger.Debug("DISPLAY: Successfully converted Mat to image.Image", "bounds", img.Bounds())

	cp.originalImage.Image = img
	cp.originalImage.Refresh()

	cp.logger.Info("DISPLAY: Original image updated successfully")

	// DEBUG: Log successful original update
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogImageOperation("update_original_success", true, map[string]interface{}{
			"image_bounds": img.Bounds().String(),
			"duration":     duration,
		})
	}
}

func (cp *CenterPanel) UpdatePreview(preview image.Image) {
	start := time.Now()
	cp.logger.Debug("DISPLAY: UpdatePreview called", "preview_nil", preview == nil)

	// DEBUG: Log preview update attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogImageOperation("update_preview_attempt", true, map[string]interface{}{
			"preview_provided": preview != nil,
			"has_image":        cp.imageData.HasImage(),
		})
	}

	if preview == nil {
		cp.logger.Debug("DISPLAY: Preview is nil, using original as preview")

		if cp.imageData.HasImage() {
			original := cp.imageData.GetOriginal()
			defer original.Close()

			if !original.Empty() {
				if img, err := original.ToImage(); err == nil {
					cp.previewImage.Image = img
					cp.previewImage.Refresh()
					cp.logger.Info("DISPLAY: Used original as preview")

					// DEBUG: Log original used as preview
					if GlobalGUIDebugger != nil {
						duration := time.Since(start)
						GlobalGUIDebugger.LogImageOperation("update_preview_used_original", true, map[string]interface{}{
							"image_bounds": img.Bounds().String(),
							"duration":     duration,
						})
					}
					return
				} else {
					// DEBUG: Log original conversion failure
					if GlobalGUIDebugger != nil {
						GlobalGUIDebugger.LogImageOperation("update_preview_original_conversion_failed", false, map[string]interface{}{
							"error": err.Error(),
						})
					}
				}
			}
		}

		// Keep current placeholder if no image loaded
		cp.logger.Debug("DISPLAY: No image available, keeping placeholder")

		// DEBUG: Log keeping placeholder
		if GlobalGUIDebugger != nil {
			GlobalGUIDebugger.LogImageOperation("update_preview_keeping_placeholder", true, map[string]interface{}{
				"reason": "no_image_or_conversion_failed",
			})
		}
		return
	}

	cp.logger.Debug("DISPLAY: Updating preview image", "bounds", preview.Bounds())

	cp.previewImage.Image = preview
	cp.previewImage.Refresh()

	cp.logger.Info("DISPLAY: Preview image updated successfully")

	// DEBUG: Log successful preview update
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogImageOperation("update_preview_success", true, map[string]interface{}{
			"image_bounds": preview.Bounds().String(),
			"duration":     duration,
		})
	}
}

func (cp *CenterPanel) Reset() {
	start := time.Now()
	originalPlaceholder := cp.createPlaceholderImage("Load an image to start processing", color.RGBA{240, 248, 255, 255})
	previewPlaceholder := cp.createPlaceholderImage("Preview will appear here", color.RGBA{248, 255, 240, 255})

	cp.originalImage.Image = originalPlaceholder
	cp.previewImage.Image = previewPlaceholder

	cp.originalImage.Refresh()
	cp.previewImage.Refresh()

	cp.currentZoom = 1.0
	cp.SetViewMode("split")

	cp.logger.Debug("DISPLAY: Center panel reset to defaults with better placeholders")

	// DEBUG: Log reset operation
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogImageOperation("center_panel_reset", true, map[string]interface{}{
			"duration":  duration,
			"view_mode": "split",
			"zoom":      1.0,
		})
	}
}

func (cp *CenterPanel) GetContainer() fyne.CanvasObject {
	return cp.container
}

func (cp *CenterPanel) SetCallbacks(onSelectionChanged func(bool)) {
	cp.onSelectionChanged = onSelectionChanged
}
