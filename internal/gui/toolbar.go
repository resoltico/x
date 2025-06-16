// internal/gui/toolbar.go
// Modern toolbar with enhanced functionality and visual design
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"
)

// ModernToolbar provides a streamlined toolbar with essential actions
type ModernToolbar struct {
	imageData *core.ImageData
	loader    *io.ImageLoader
	logger    *slog.Logger

	container *fyne.Container

	// File operations
	openBtn *widget.Button
	saveBtn *widget.Button

	// Selection tools
	rectTool     *widget.Button
	freehandTool *widget.Button
	clearSelBtn  *widget.Button

	// View controls
	zoomSlider    *widget.Slider
	zoomLabel     *widget.Label
	resetViewBtn  *widget.Button
	toggleViewBtn *widget.Button

	// Global actions
	resetBtn *widget.Button

	// Callbacks
	onImageLoaded func(string)
	onImageSaved  func(string)
	onToolChanged func(string)
	onResetImage  func()
	onZoomChanged func(float64)
	onViewToggled func()

	currentTool string
	currentZoom float64
}

func NewModernToolbar(imageData *core.ImageData, loader *io.ImageLoader, logger *slog.Logger) *ModernToolbar {
	toolbar := &ModernToolbar{
		imageData:   imageData,
		loader:      loader,
		logger:      logger,
		currentTool: "none",
		currentZoom: 1.0,
	}

	toolbar.initializeUI()
	return toolbar
}

func (mt *ModernToolbar) initializeUI() {
	// File operations group
	mt.openBtn = widget.NewButtonWithIcon("Open", theme.FolderOpenIcon(), mt.openImage)
	mt.openBtn.Importance = widget.HighImportance

	mt.saveBtn = widget.NewButtonWithIcon("Save", theme.DocumentSaveIcon(), mt.saveImage)
	mt.saveBtn.Importance = widget.MediumImportance
	mt.saveBtn.Disable()

	fileGroup := container.NewHBox(
		mt.openBtn,
		mt.saveBtn,
		widget.NewSeparator(),
	)

	// Selection tools group
	mt.rectTool = widget.NewButtonWithIcon("Rectangle", theme.ContentAddIcon(), func() {
		mt.setActiveTool("rectangle")
	})
	mt.rectTool.Importance = widget.MediumImportance
	mt.rectTool.Disable()

	mt.freehandTool = widget.NewButtonWithIcon("Freehand", theme.ContentCutIcon(), func() {
		mt.setActiveTool("freehand")
	})
	mt.freehandTool.Importance = widget.MediumImportance
	mt.freehandTool.Disable()

	mt.clearSelBtn = widget.NewButtonWithIcon("Clear Selection", theme.ContentClearIcon(), func() {
		mt.setActiveTool("none")
		// Clear selection logic would be handled by callback
	})
	mt.clearSelBtn.Importance = widget.LowImportance
	mt.clearSelBtn.Disable()

	selectionGroup := container.NewHBox(
		widget.NewLabel("Selection:"),
		mt.rectTool,
		mt.freehandTool,
		mt.clearSelBtn,
		widget.NewSeparator(),
	)

	// View controls group
	mt.zoomSlider = widget.NewSlider(0.1, 5.0)
	mt.zoomSlider.SetValue(1.0)
	mt.zoomSlider.Step = 0.1
	mt.zoomSlider.Resize(fyne.NewSize(150, 30))
	mt.zoomSlider.OnChanged = func(value float64) {
		mt.currentZoom = value
		mt.zoomLabel.SetText(fmt.Sprintf("%.0f%%", value*100))
		if mt.onZoomChanged != nil {
			mt.onZoomChanged(value)
		}
	}

	mt.zoomLabel = widget.NewLabel("100%")
	mt.zoomLabel.Resize(fyne.NewSize(50, 30))

	zoomInBtn := widget.NewButtonWithIcon("", theme.ZoomInIcon(), func() {
		newZoom := mt.currentZoom + 0.2
		if newZoom <= 5.0 {
			mt.zoomSlider.SetValue(newZoom)
		}
	})
	zoomInBtn.Importance = widget.LowImportance

	zoomOutBtn := widget.NewButtonWithIcon("", theme.ZoomOutIcon(), func() {
		newZoom := mt.currentZoom - 0.2
		if newZoom >= 0.1 {
			mt.zoomSlider.SetValue(newZoom)
		}
	})
	zoomOutBtn.Importance = widget.LowImportance

	mt.resetViewBtn = widget.NewButtonWithIcon("Reset View", theme.ViewRefreshIcon(), func() {
		mt.zoomSlider.SetValue(1.0)
	})
	mt.resetViewBtn.Importance = widget.LowImportance

	mt.toggleViewBtn = widget.NewButtonWithIcon("Toggle View", theme.ViewFullScreenIcon(), func() {
		if mt.onViewToggled != nil {
			mt.onViewToggled()
		}
	})
	mt.toggleViewBtn.Importance = widget.LowImportance

	viewGroup := container.NewHBox(
		widget.NewLabel("Zoom:"),
		zoomOutBtn,
		mt.zoomSlider,
		zoomInBtn,
		mt.zoomLabel,
		widget.NewSeparator(),
		mt.resetViewBtn,
		mt.toggleViewBtn,
		widget.NewSeparator(),
	)

	// Global actions
	mt.resetBtn = widget.NewButtonWithIcon("Reset to Original", theme.ViewRefreshIcon(), func() {
		if mt.onResetImage != nil {
			mt.onResetImage()
		}
	})
	mt.resetBtn.Importance = widget.HighImportance
	mt.resetBtn.Disable()

	globalGroup := container.NewHBox(
		mt.resetBtn,
	)

	// Main toolbar container with modern spacing and padding
	toolbarContent := container.NewHBox(
		fileGroup,
		selectionGroup,
		viewGroup,
		globalGroup,
	)

	// Wrap in a card for modern appearance
	mt.container = container.NewBorder(
		nil, nil, nil, nil,
		widget.NewCard("", "", toolbarContent),
	)
}

func (mt *ModernToolbar) GetContainer() fyne.CanvasObject {
	return mt.container
}

func (mt *ModernToolbar) setActiveTool(tool string) {
	// Reset all tool button styles
	mt.rectTool.Importance = widget.MediumImportance
	mt.freehandTool.Importance = widget.MediumImportance

	// Highlight active tool
	switch tool {
	case "rectangle":
		mt.rectTool.Importance = widget.HighImportance
	case "freehand":
		mt.freehandTool.Importance = widget.HighImportance
	case "none":
		// No tool highlighted
	}

	mt.currentTool = tool

	if mt.onToolChanged != nil {
		mt.onToolChanged(tool)
	}

	// Refresh buttons to show importance changes
	mt.rectTool.Refresh()
	mt.freehandTool.Refresh()
}

func (mt *ModernToolbar) openImage() {
	mt.logger.Info("Opening file dialog for image selection")

	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			mt.logger.Error("File dialog error", "error", err)
			mt.showErrorDialog("File Dialog Error", err)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		uri := reader.URI()
		filepath := uri.Path()

		mt.logger.Info("Loading selected image", "filepath", filepath)

		mat, err := mt.loader.LoadImage(filepath)
		if err != nil {
			mt.logger.Error("Failed to load image", "error", err)
			mt.showErrorDialog("Failed to Load Image", err)
			return
		}
		defer mat.Close()

		if err := core.ValidateImage(mat); err != nil {
			mt.logger.Error("Invalid image", "error", err)
			mt.showErrorDialog("Invalid Image", err)
			return
		}

		if err := mt.imageData.SetOriginal(mat, filepath); err != nil {
			mt.logger.Error("Failed to set image", "error", err)
			mt.showErrorDialog("Failed to Set Image", err)
			return
		}

		// Enable controls
		mt.saveBtn.Enable()
		mt.rectTool.Enable()
		mt.freehandTool.Enable()
		mt.clearSelBtn.Enable()
		mt.resetBtn.Enable()

		mt.logger.Info("Image loaded successfully", "filepath", filepath)

		if mt.onImageLoaded != nil {
			mt.onImageLoaded(filepath)
		}

	}, mt.getParentWindow())

	// Set file filters for supported image formats
	imageFilter := storage.NewExtensionFileFilter([]string{
		".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp", ".gif", ".webp",
	})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Resize(fyne.NewSize(800, 600))
	fileDialog.Show()
}

func (mt *ModernToolbar) saveImage() {
	if !mt.imageData.HasImage() {
		mt.showErrorDialog("No Image", fmt.Errorf("no image loaded to save"))
		return
	}

	mt.logger.Info("Opening file dialog for image saving")

	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			mt.logger.Error("File dialog error", "error", err)
			mt.showErrorDialog("File Dialog Error", err)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		uri := writer.URI()
		filepath := uri.Path()

		mt.logger.Info("Saving image", "filepath", filepath)

		processed := mt.imageData.GetProcessed()
		defer processed.Close()

		// If no processing has been done, save the original
		if processed.Empty() {
			processed = mt.imageData.GetOriginal()
		}

		if err := mt.loader.SaveImage(processed, filepath); err != nil {
			mt.logger.Error("Failed to save image", "error", err)
			mt.showErrorDialog("Failed to Save Image", err)
			return
		}

		mt.logger.Info("Image saved successfully", "filepath", filepath)

		if mt.onImageSaved != nil {
			mt.onImageSaved(filepath)
		}

		// Show success message
		mt.showInfoDialog("Image Saved", fmt.Sprintf("Image successfully saved to:\n%s", filepath))

	}, mt.getParentWindow())

	// Set default filename and filter
	originalPath := mt.imageData.GetFilepath()
	if originalPath != "" {
		// Extract filename without extension and add _processed
		baseName := getBaseName(originalPath)
		fileDialog.SetFileName(baseName + "_processed.png")
	} else {
		fileDialog.SetFileName("processed_image.png")
	}

	imageFilter := storage.NewExtensionFileFilter([]string{
		".png", ".jpg", ".jpeg", ".tiff", ".tif", ".bmp",
	})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Resize(fyne.NewSize(800, 600))
	fileDialog.Show()
}

func (mt *ModernToolbar) showErrorDialog(title string, err error) {
	dialog.ShowError(err, mt.getParentWindow())
}

func (mt *ModernToolbar) showInfoDialog(title, message string) {
	dialog.ShowInformation(title, message, mt.getParentWindow())
}

func (mt *ModernToolbar) getParentWindow() fyne.Window {
	// Try to get the parent window from the current app
	windows := fyne.CurrentApp().Driver().AllWindows()
	if len(windows) > 0 {
		return windows[0]
	}
	// Fallback: create a new window (should not normally happen)
	return fyne.CurrentApp().NewWindow("Dialog")
}

func (mt *ModernToolbar) SetCallbacks(
	onImageLoaded func(string),
	onImageSaved func(string),
	onToolChanged func(string),
	onResetImage func(),
) {
	mt.onImageLoaded = onImageLoaded
	mt.onImageSaved = onImageSaved
	mt.onToolChanged = onToolChanged
	mt.onResetImage = onResetImage
}

func (mt *ModernToolbar) SetZoomCallback(onZoomChanged func(float64)) {
	mt.onZoomChanged = onZoomChanged
}

func (mt *ModernToolbar) SetViewCallback(onViewToggled func()) {
	mt.onViewToggled = onViewToggled
}

func (mt *ModernToolbar) UpdateSelectionState(hasSelection bool) {
	if hasSelection {
		mt.clearSelBtn.Enable()
	} else {
		mt.clearSelBtn.Disable()
	}
}

func (mt *ModernToolbar) GetCurrentTool() string {
	return mt.currentTool
}

func (mt *ModernToolbar) GetCurrentZoom() float64 {
	return mt.currentZoom
}

func (mt *ModernToolbar) SetZoom(zoom float64) {
	if zoom >= 0.1 && zoom <= 5.0 {
		mt.zoomSlider.SetValue(zoom)
	}
}

func (mt *ModernToolbar) Enable() {
	mt.openBtn.Enable()
	if mt.imageData.HasImage() {
		mt.saveBtn.Enable()
		mt.rectTool.Enable()
		mt.freehandTool.Enable()
		mt.resetBtn.Enable()
	}
}

func (mt *ModernToolbar) Disable() {
	mt.saveBtn.Disable()
	mt.rectTool.Disable()
	mt.freehandTool.Disable()
	mt.clearSelBtn.Disable()
	mt.resetBtn.Disable()
}

func (mt *ModernToolbar) Refresh() {
	mt.container.Refresh()
}

// Utility function to extract base filename without extension
func getBaseName(filepath string) string {
	// Find the last slash or backslash
	lastSlash := -1
	for i := len(filepath) - 1; i >= 0; i-- {
		if filepath[i] == '/' || filepath[i] == '\\' {
			lastSlash = i
			break
		}
	}

	// Extract filename
	filename := filepath
	if lastSlash != -1 {
		filename = filepath[lastSlash+1:]
	}

	// Remove extension
	lastDot := -1
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '.' {
			lastDot = i
			break
		}
	}

	if lastDot != -1 {
		return filename[:lastDot]
	}

	return filename
}

// GetSupportedImageFormats returns the list of supported image formats
func (mt *ModernToolbar) GetSupportedImageFormats() []string {
	return []string{
		"JPEG (.jpg, .jpeg)",
		"PNG (.png)",
		"TIFF (.tiff, .tif)",
		"BMP (.bmp)",
		"GIF (.gif)",
		"WebP (.webp)",
	}
}

// IsImageLoaded checks if an image is currently loaded
func (mt *ModernToolbar) IsImageLoaded() bool {
	return mt.imageData.HasImage()
}

// ClearSelection clears the current selection
func (mt *ModernToolbar) ClearSelection() {
	mt.setActiveTool("none")
}

// ResetView resets the zoom to 100%
func (mt *ModernToolbar) ResetView() {
	mt.zoomSlider.SetValue(1.0)
}
