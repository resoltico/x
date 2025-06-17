// internal/gui/toolbar.go
// Perfect UI Top Toolbar (50px height) with exact specification
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

type Toolbar struct {
	imageData *core.ImageData
	loader    *io.ImageLoader
	pipeline  *core.EnhancedPipeline
	logger    *slog.Logger

	container *fyne.Container

	// File operations (left side)
	openBtn  *widget.Button
	saveBtn  *widget.Button
	resetBtn *widget.Button

	// Zoom controls (center)
	zoomSlider     *widget.Slider
	zoomLabel      *widget.Label
	zoomInBtn      *widget.Button
	zoomOutBtn     *widget.Button
	zoomPercentage *widget.Label

	// View toggles (right side)
	singleViewBtn  *widget.Button
	splitViewBtn   *widget.Button
	overlayViewBtn *widget.Button

	currentZoom float64
	currentView string

	// Callbacks
	onImageLoaded func(string)
	onImageSaved  func(string)
	onReset       func()
	onZoomChanged func(float64)
	onViewChanged func(string)
}

func NewToolbar(imageData *core.ImageData, loader *io.ImageLoader, pipeline *core.EnhancedPipeline, logger *slog.Logger) *Toolbar {
	toolbar := &Toolbar{
		imageData:   imageData,
		loader:      loader,
		pipeline:    pipeline,
		logger:      logger,
		currentZoom: 1.0,
		currentView: "split",
	}

	toolbar.initializeUI()
	return toolbar
}

func (tb *Toolbar) initializeUI() {
	// Title
	titleLabel := widget.NewLabelWithStyle("Image Restoration Suite v2.0", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	titleLabel.Resize(fyne.NewSize(300, 30))

	// File operation buttons (left side)
	tb.openBtn = widget.NewButtonWithIcon("OPEN IMAGE", theme.FolderOpenIcon(), tb.openImage)
	tb.openBtn.Resize(fyne.NewSize(120, 40))
	tb.openBtn.Importance = widget.HighImportance

	tb.saveBtn = widget.NewButtonWithIcon("SAVE IMAGE", theme.DocumentSaveIcon(), tb.saveImage)
	tb.saveBtn.Resize(fyne.NewSize(120, 40))
	tb.saveBtn.Importance = widget.HighImportance
	tb.saveBtn.Disable()

	tb.resetBtn = widget.NewButtonWithIcon("Reset", theme.ViewRefreshIcon(), func() {
		if tb.onReset != nil {
			tb.onReset()
		}
	})
	tb.resetBtn.Resize(fyne.NewSize(80, 40))
	tb.resetBtn.Importance = widget.HighImportance
	tb.resetBtn.Disable()

	leftSection := container.NewHBox(
		titleLabel,
		widget.NewSeparator(),
		tb.openBtn,
		tb.saveBtn,
		tb.resetBtn,
	)

	// Zoom controls (center)
	zoomLabel := widget.NewLabel("Zoom:")

	tb.zoomOutBtn = widget.NewButtonWithIcon("", theme.ZoomOutIcon(), func() {
		newZoom := tb.currentZoom - 0.25
		if newZoom >= 0.25 {
			tb.setZoom(newZoom)
		}
	})
	tb.zoomOutBtn.Resize(fyne.NewSize(24, 24))

	tb.zoomSlider = widget.NewSlider(0.25, 4.0)
	tb.zoomSlider.SetValue(1.0)
	tb.zoomSlider.Step = 0.25
	tb.zoomSlider.Resize(fyne.NewSize(100, 25))
	tb.zoomSlider.OnChanged = func(value float64) {
		tb.setZoom(value)
	}

	tb.zoomInBtn = widget.NewButtonWithIcon("", theme.ZoomInIcon(), func() {
		newZoom := tb.currentZoom + 0.25
		if newZoom <= 4.0 {
			tb.setZoom(newZoom)
		}
	})
	tb.zoomInBtn.Resize(fyne.NewSize(24, 24))

	tb.zoomPercentage = widget.NewLabel("100%")
	tb.zoomPercentage.Resize(fyne.NewSize(50, 25))

	centerSection := container.NewHBox(
		zoomLabel,
		tb.zoomOutBtn,
		tb.zoomSlider,
		tb.zoomInBtn,
		tb.zoomPercentage,
	)

	// View toggle buttons (right side)
	viewLabel := widget.NewLabel("View:")

	tb.singleViewBtn = widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), func() {
		tb.setView("single")
	})
	tb.singleViewBtn.Resize(fyne.NewSize(24, 24))

	tb.splitViewBtn = widget.NewButtonWithIcon("", theme.ListIcon(), func() {
		tb.setView("split")
	})
	tb.splitViewBtn.Resize(fyne.NewSize(24, 24))
	tb.splitViewBtn.Importance = widget.HighImportance // Default active

	tb.overlayViewBtn = widget.NewButtonWithIcon("", theme.ViewRestoreIcon(), func() {
		tb.setView("overlay")
	})
	tb.overlayViewBtn.Resize(fyne.NewSize(24, 24))

	rightSection := container.NewHBox(
		viewLabel,
		tb.singleViewBtn,
		tb.splitViewBtn,
		tb.overlayViewBtn,
	)

	// Main toolbar layout with proper spacing and fixed height
	tb.container = container.NewBorder(
		nil, nil,
		leftSection,   // left
		rightSection,  // right
		centerSection, // center
	)

	// Set background color and fixed height to 50px as per specification
	tb.container.Resize(fyne.NewSize(1600, 50))
}

func (tb *Toolbar) setZoom(zoom float64) {
	tb.currentZoom = zoom
	tb.zoomSlider.SetValue(zoom)
	tb.zoomPercentage.SetText(fmt.Sprintf("%.0f%%", zoom*100))
	if tb.onZoomChanged != nil {
		tb.onZoomChanged(zoom)
	}
}

func (tb *Toolbar) setView(view string) {
	tb.currentView = view

	// Reset button importance
	tb.singleViewBtn.Importance = widget.MediumImportance
	tb.splitViewBtn.Importance = widget.MediumImportance
	tb.overlayViewBtn.Importance = widget.MediumImportance

	// Highlight active view
	switch view {
	case "single":
		tb.singleViewBtn.Importance = widget.HighImportance
	case "split":
		tb.splitViewBtn.Importance = widget.HighImportance
	case "overlay":
		tb.overlayViewBtn.Importance = widget.HighImportance
	}

	// Refresh buttons
	tb.singleViewBtn.Refresh()
	tb.splitViewBtn.Refresh()
	tb.overlayViewBtn.Refresh()

	if tb.onViewChanged != nil {
		tb.onViewChanged(view)
	}
}

func (tb *Toolbar) openImage() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		filepath := reader.URI().Path()
		mat, err := tb.loader.LoadImage(filepath)
		if err != nil {
			tb.logger.Error("Failed to load image", "error", err)
			return
		}
		defer mat.Close()

		if err := tb.imageData.SetOriginal(mat, filepath); err != nil {
			tb.logger.Error("Failed to set image", "error", err)
			return
		}

		tb.enableProcessingButtons()

		if tb.onImageLoaded != nil {
			tb.onImageLoaded(filepath)
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	imageFilter := storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()
}

func (tb *Toolbar) saveImage() {
	if !tb.imageData.HasImage() {
		return
	}

	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		filepath := writer.URI().Path()

		// Process full resolution first
		if _, err := tb.pipeline.ProcessFullResolution(); err != nil {
			tb.logger.Error("Failed to process full resolution", "error", err)
			return
		}

		processed := tb.imageData.GetProcessed()
		defer processed.Close()

		if processed.Empty() {
			processed = tb.imageData.GetOriginal()
		}

		if err := tb.loader.SaveImage(processed, filepath); err != nil {
			tb.logger.Error("Failed to save image", "error", err)
			return
		}

		if tb.onImageSaved != nil {
			tb.onImageSaved(filepath)
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	imageFilter := storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".tiff", ".tif"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()
}

func (tb *Toolbar) enableProcessingButtons() {
	tb.saveBtn.Enable()
	tb.resetBtn.Enable()
}

func (tb *Toolbar) disableProcessingButtons() {
	tb.saveBtn.Disable()
	tb.resetBtn.Disable()
}

func (tb *Toolbar) GetContainer() fyne.CanvasObject {
	return tb.container
}

func (tb *Toolbar) SetCallbacks(
	onImageLoaded func(string),
	onImageSaved func(string),
	onReset func(),
	onZoomChanged func(float64),
	onViewChanged func(string),
) {
	tb.onImageLoaded = onImageLoaded
	tb.onImageSaved = onImageSaved
	tb.onReset = onReset
	tb.onZoomChanged = onZoomChanged
	tb.onViewChanged = onViewChanged
}
