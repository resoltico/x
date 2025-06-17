// internal/gui/toolbar.go
// Perfect UI Top Toolbar with comprehensive debug integration
package gui

import (
	"fmt"
	"log/slog"
	"time"

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
	start := time.Now()

	// DEBUG: Log toolbar initialization
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "initialize_start", map[string]interface{}{
			"default_zoom": tb.currentZoom,
			"default_view": tb.currentView,
		})
	}

	// File operation buttons (left side)
	tb.openBtn = widget.NewButtonWithIcon("OPEN IMAGE", theme.FolderOpenIcon(), tb.openImage)
	tb.openBtn.Resize(fyne.NewSize(120, 40))
	tb.openBtn.Importance = widget.HighImportance

	tb.saveBtn = widget.NewButtonWithIcon("SAVE IMAGE", theme.DocumentSaveIcon(), tb.saveImage)
	tb.saveBtn.Resize(fyne.NewSize(120, 40))
	tb.saveBtn.Importance = widget.HighImportance
	tb.saveBtn.Disable()

	tb.resetBtn = widget.NewButtonWithIcon("Reset", theme.ViewRefreshIcon(), func() {
		tb.handleReset()
	})
	tb.resetBtn.Resize(fyne.NewSize(80, 40))
	tb.resetBtn.Importance = widget.HighImportance
	tb.resetBtn.Disable()

	leftSection := container.NewHBox(
		tb.openBtn,
		tb.saveBtn,
		tb.resetBtn,
	)

	// Zoom controls (center)
	zoomLabel := widget.NewLabel("Zoom:")

	tb.zoomOutBtn = widget.NewButtonWithIcon("", theme.ZoomOutIcon(), func() {
		tb.handleZoomOut()
	})
	tb.zoomOutBtn.Resize(fyne.NewSize(24, 24))

	tb.zoomSlider = widget.NewSlider(0.25, 4.0)
	tb.zoomSlider.SetValue(1.0)
	tb.zoomSlider.Step = 0.25
	tb.zoomSlider.Resize(fyne.NewSize(100, 25))
	tb.zoomSlider.OnChanged = func(value float64) {
		tb.handleZoomSliderChange(value)
	}

	tb.zoomInBtn = widget.NewButtonWithIcon("", theme.ZoomInIcon(), func() {
		tb.handleZoomIn()
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
		tb.handleViewChange("single")
	})
	tb.singleViewBtn.Resize(fyne.NewSize(24, 24))

	tb.splitViewBtn = widget.NewButtonWithIcon("", theme.ListIcon(), func() {
		tb.handleViewChange("split")
	})
	tb.splitViewBtn.Resize(fyne.NewSize(24, 24))
	tb.splitViewBtn.Importance = widget.HighImportance // Default active

	tb.overlayViewBtn = widget.NewButtonWithIcon("", theme.ViewRestoreIcon(), func() {
		tb.handleViewChange("overlay")
	})
	tb.overlayViewBtn.Resize(fyne.NewSize(24, 24))

	rightSection := container.NewHBox(
		viewLabel,
		tb.singleViewBtn,
		tb.splitViewBtn,
		tb.overlayViewBtn,
	)

	// Main toolbar layout
	tb.container = container.NewBorder(
		nil, nil,
		leftSection,   // left
		rightSection,  // right
		centerSection, // center
	)

	// Set fixed height to 50px
	tb.container.Resize(fyne.NewSize(1600, 50))

	// DEBUG: Log toolbar initialization complete
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "initialize_complete", map[string]interface{}{
			"duration":     duration,
			"button_count": 8, // open, save, reset, zoom-, zoom+, single, split, overlay
		})
	}

	tb.logger.Debug("TOOLBAR: Initialized successfully")
}

func (tb *Toolbar) handleReset() {
	start := time.Now()
	tb.logger.Info("TOOLBAR: Reset button clicked")

	// DEBUG: Log reset attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "reset_clicked", map[string]interface{}{
			"has_image": tb.imageData.HasImage(),
		})
	}

	if tb.onReset != nil {
		tb.onReset()
	}

	// DEBUG: Log reset completion
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "reset_completed", map[string]interface{}{
			"duration": duration,
		})
	}
}

func (tb *Toolbar) handleZoomOut() {
	start := time.Now()
	oldZoom := tb.currentZoom
	newZoom := tb.currentZoom - 0.25

	tb.logger.Debug("TOOLBAR: Zoom out clicked", "current_zoom", oldZoom, "target_zoom", newZoom)

	// DEBUG: Log zoom out attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_out_clicked", map[string]interface{}{
			"old_zoom":      oldZoom,
			"target_zoom":   newZoom,
			"within_bounds": newZoom >= 0.25,
		})
	}

	if newZoom >= 0.25 {
		tb.setZoom(newZoom)

		// DEBUG: Log successful zoom out
		if GlobalGUIDebugger != nil {
			duration := time.Since(start)
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_out_success", map[string]interface{}{
				"old_zoom": oldZoom,
				"new_zoom": newZoom,
				"duration": duration,
			})
		}
	} else {
		tb.logger.Debug("TOOLBAR: Zoom out rejected - would go below minimum", "minimum", 0.25)

		// DEBUG: Log zoom out rejected
		if GlobalGUIDebugger != nil {
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_out_rejected", map[string]interface{}{
				"reason":      "below_minimum",
				"target_zoom": newZoom,
				"minimum":     0.25,
			})
		}
	}
}

func (tb *Toolbar) handleZoomIn() {
	start := time.Now()
	oldZoom := tb.currentZoom
	newZoom := tb.currentZoom + 0.25

	tb.logger.Debug("TOOLBAR: Zoom in clicked", "current_zoom", oldZoom, "target_zoom", newZoom)

	// DEBUG: Log zoom in attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_in_clicked", map[string]interface{}{
			"old_zoom":      oldZoom,
			"target_zoom":   newZoom,
			"within_bounds": newZoom <= 4.0,
		})
	}

	if newZoom <= 4.0 {
		tb.setZoom(newZoom)

		// DEBUG: Log successful zoom in
		if GlobalGUIDebugger != nil {
			duration := time.Since(start)
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_in_success", map[string]interface{}{
				"old_zoom": oldZoom,
				"new_zoom": newZoom,
				"duration": duration,
			})
		}
	} else {
		tb.logger.Debug("TOOLBAR: Zoom in rejected - would go above maximum", "maximum", 4.0)

		// DEBUG: Log zoom in rejected
		if GlobalGUIDebugger != nil {
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_in_rejected", map[string]interface{}{
				"reason":      "above_maximum",
				"target_zoom": newZoom,
				"maximum":     4.0,
			})
		}
	}
}

func (tb *Toolbar) handleZoomSliderChange(value float64) {
	start := time.Now()
	oldZoom := tb.currentZoom

	tb.logger.Debug("TOOLBAR: Zoom slider changed", "old_zoom", oldZoom, "new_value", value)

	// DEBUG: Log slider change
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_slider_changed", map[string]interface{}{
			"old_zoom":  oldZoom,
			"new_value": value,
			"source":    "slider",
		})
	}

	tb.setZoom(value)

	// DEBUG: Log slider change completion
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_slider_applied", map[string]interface{}{
			"old_zoom": oldZoom,
			"new_zoom": value,
			"duration": duration,
		})
	}
}

func (tb *Toolbar) setZoom(zoom float64) {
	start := time.Now()
	oldZoom := tb.currentZoom
	tb.currentZoom = zoom
	tb.zoomSlider.SetValue(zoom)
	tb.zoomPercentage.SetText(fmt.Sprintf("%.0f%%", zoom*100))

	tb.logger.Info("TOOLBAR: Zoom set", "old_zoom", oldZoom, "new_zoom", zoom)

	// DEBUG: Log zoom setting
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_set", map[string]interface{}{
			"old_zoom":   oldZoom,
			"new_zoom":   zoom,
			"percentage": fmt.Sprintf("%.0f%%", zoom*100),
		})
	}

	if tb.onZoomChanged != nil {
		tb.onZoomChanged(zoom)

		// DEBUG: Log callback execution
		if GlobalGUIDebugger != nil {
			duration := time.Since(start)
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_callback_executed", map[string]interface{}{
				"zoom":     zoom,
				"duration": duration,
			})
		}
	} else {
		tb.logger.Warn("TOOLBAR: No zoom changed callback set")

		// DEBUG: Log missing callback
		if GlobalGUIDebugger != nil {
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "zoom_callback_missing", map[string]interface{}{
				"zoom": zoom,
			})
		}
	}
}

func (tb *Toolbar) handleViewChange(view string) {
	start := time.Now()
	oldView := tb.currentView

	tb.logger.Info("TOOLBAR: View change requested", "old_view", oldView, "new_view", view)

	// DEBUG: Log view change attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "view_change_requested", map[string]interface{}{
			"old_view": oldView,
			"new_view": view,
		})
	}

	tb.setView(view)

	// DEBUG: Log view change completion
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "view_change_completed", map[string]interface{}{
			"old_view": oldView,
			"new_view": view,
			"duration": duration,
		})
	}
}

func (tb *Toolbar) setView(view string) {
	start := time.Now()
	oldView := tb.currentView
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

	tb.logger.Info("TOOLBAR: View set", "old_view", oldView, "new_view", view)

	// DEBUG: Log view setting
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "view_set", map[string]interface{}{
			"old_view": oldView,
			"new_view": view,
		})
	}

	if tb.onViewChanged != nil {
		tb.onViewChanged(view)

		// DEBUG: Log callback execution
		if GlobalGUIDebugger != nil {
			duration := time.Since(start)
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "view_callback_executed", map[string]interface{}{
				"view":     view,
				"duration": duration,
			})
		}
	} else {
		tb.logger.Warn("TOOLBAR: No view changed callback set")

		// DEBUG: Log missing callback
		if GlobalGUIDebugger != nil {
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "view_callback_missing", map[string]interface{}{
				"view": view,
			})
		}
	}
}

func (tb *Toolbar) openImage() {
	start := time.Now()
	tb.logger.Info("TOOLBAR: Open image clicked")

	// DEBUG: Log open image attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "open_image_clicked", map[string]interface{}{
			"has_existing_image": tb.imageData.HasImage(),
		})
	}

	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		tb.handleFileDialogResult(reader, err, start)
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	imageFilter := storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()

	// DEBUG: Log dialog shown
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "file_dialog_shown", map[string]interface{}{
			"supported_formats": []string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"},
		})
	}
}

func (tb *Toolbar) handleFileDialogResult(reader fyne.URIReadCloser, err error, startTime time.Time) {
	if err != nil || reader == nil {
		tb.logger.Debug("TOOLBAR: File dialog cancelled or error", "error", err)

		// DEBUG: Log dialog cancellation
		if GlobalGUIDebugger != nil {
			duration := time.Since(startTime)
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "file_dialog_cancelled", map[string]interface{}{
				"error":    err,
				"duration": duration,
			})
		}
		return
	}
	defer reader.Close()

	filepath := reader.URI().Path()
	tb.logger.Info("TOOLBAR: File selected", "filepath", filepath)

	// DEBUG: Log file selection
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "file_selected", map[string]interface{}{
			"filepath": filepath,
		})
	}

	loadStart := time.Now()
	mat, err := tb.loader.LoadImage(filepath)
	if err != nil {
		tb.logger.Error("TOOLBAR: Failed to load image", "error", err, "filepath", filepath)

		// DEBUG: Log load failure
		if GlobalGUIDebugger != nil {
			duration := time.Since(loadStart)
			GlobalGUIDebugger.LogImageOperation("image_load_failed", false, map[string]interface{}{
				"filepath": filepath,
				"error":    err.Error(),
				"duration": duration,
			})
		}

		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}
	defer mat.Close()

	if err := tb.imageData.SetOriginal(mat, filepath); err != nil {
		tb.logger.Error("TOOLBAR: Failed to set image", "error", err, "filepath", filepath)

		// DEBUG: Log set image failure
		if GlobalGUIDebugger != nil {
			duration := time.Since(loadStart)
			GlobalGUIDebugger.LogImageOperation("image_set_failed", false, map[string]interface{}{
				"filepath": filepath,
				"error":    err.Error(),
				"duration": duration,
			})
		}

		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	tb.enableProcessingButtons()

	// DEBUG: Log successful image loading
	if GlobalGUIDebugger != nil {
		totalDuration := time.Since(startTime)
		loadDuration := time.Since(loadStart)
		metadata := tb.imageData.GetMetadata()
		GlobalGUIDebugger.LogImageOperation("image_load_success", true, map[string]interface{}{
			"filepath":       filepath,
			"load_duration":  loadDuration,
			"total_duration": totalDuration,
			"width":          metadata.Width,
			"height":         metadata.Height,
			"channels":       metadata.Channels,
		})
	}

	if tb.onImageLoaded != nil {
		tb.onImageLoaded(filepath)
	}

	tb.logger.Info("TOOLBAR: Image loaded successfully", "filepath", filepath)
}

func (tb *Toolbar) saveImage() {
	start := time.Now()
	tb.logger.Info("TOOLBAR: Save image clicked")

	// DEBUG: Log save attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "save_image_clicked", map[string]interface{}{
			"has_image": tb.imageData.HasImage(),
		})
	}

	if !tb.imageData.HasImage() {
		tb.logger.Warn("TOOLBAR: Save clicked but no image loaded")

		// DEBUG: Log no image to save
		if GlobalGUIDebugger != nil {
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "save_no_image", map[string]interface{}{
				"reason": "no_image_loaded",
			})
		}
		return
	}

	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		tb.handleSaveDialogResult(writer, err, start)
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	imageFilter := storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".tiff", ".tif"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()

	// DEBUG: Log save dialog shown
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "save_dialog_shown", map[string]interface{}{
			"supported_formats": []string{".png", ".jpg", ".jpeg", ".tiff", ".tif"},
		})
	}
}

func (tb *Toolbar) handleSaveDialogResult(writer fyne.URIWriteCloser, err error, startTime time.Time) {
	if err != nil || writer == nil {
		tb.logger.Debug("TOOLBAR: Save dialog cancelled or error", "error", err)

		// DEBUG: Log save cancellation
		if GlobalGUIDebugger != nil {
			duration := time.Since(startTime)
			GlobalGUIDebugger.LogUIInteraction("Toolbar", "save_dialog_cancelled", map[string]interface{}{
				"error":    err,
				"duration": duration,
			})
		}
		return
	}
	defer writer.Close()

	filepath := writer.URI().Path()
	tb.logger.Info("TOOLBAR: Save path selected", "filepath", filepath)

	// Process full resolution first
	processStart := time.Now()
	if _, err := tb.pipeline.ProcessFullResolution(); err != nil {
		tb.logger.Error("TOOLBAR: Failed to process full resolution", "error", err)

		// DEBUG: Log processing failure
		if GlobalGUIDebugger != nil {
			duration := time.Since(processStart)
			GlobalGUIDebugger.LogImageOperation("save_processing_failed", false, map[string]interface{}{
				"filepath": filepath,
				"error":    err.Error(),
				"duration": duration,
			})
		}

		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	processed := tb.imageData.GetProcessed()
	defer processed.Close()

	if processed.Empty() {
		processed = tb.imageData.GetOriginal()
	}

	saveStart := time.Now()
	if err := tb.loader.SaveImage(processed, filepath); err != nil {
		tb.logger.Error("TOOLBAR: Failed to save image", "error", err, "filepath", filepath)

		// DEBUG: Log save failure
		if GlobalGUIDebugger != nil {
			duration := time.Since(saveStart)
			GlobalGUIDebugger.LogImageOperation("image_save_failed", false, map[string]interface{}{
				"filepath": filepath,
				"error":    err.Error(),
				"duration": duration,
			})
		}

		dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
		return
	}

	// DEBUG: Log successful save
	if GlobalGUIDebugger != nil {
		totalDuration := time.Since(startTime)
		saveDuration := time.Since(saveStart)
		processDuration := saveStart.Sub(processStart)
		GlobalGUIDebugger.LogImageOperation("image_save_success", true, map[string]interface{}{
			"filepath":         filepath,
			"save_duration":    saveDuration,
			"process_duration": processDuration,
			"total_duration":   totalDuration,
		})
	}

	if tb.onImageSaved != nil {
		tb.onImageSaved(filepath)
	}

	tb.logger.Info("TOOLBAR: Image saved successfully", "filepath", filepath)
}

func (tb *Toolbar) enableProcessingButtons() {
	start := time.Now()
	tb.saveBtn.Enable()
	tb.resetBtn.Enable()

	tb.logger.Debug("TOOLBAR: Processing buttons enabled")

	// DEBUG: Log button state change
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "processing_buttons_enabled", map[string]interface{}{
			"save_enabled":  true,
			"reset_enabled": true,
			"duration":      duration,
		})
	}
}

func (tb *Toolbar) disableProcessingButtons() {
	start := time.Now()
	tb.saveBtn.Disable()
	tb.resetBtn.Disable()

	tb.logger.Debug("TOOLBAR: Processing buttons disabled")

	// DEBUG: Log button state change
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "processing_buttons_disabled", map[string]interface{}{
			"save_enabled":  false,
			"reset_enabled": false,
			"duration":      duration,
		})
	}
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
	start := time.Now()
	tb.onImageLoaded = onImageLoaded
	tb.onImageSaved = onImageSaved
	tb.onReset = onReset
	tb.onZoomChanged = onZoomChanged
	tb.onViewChanged = onViewChanged

	tb.logger.Debug("TOOLBAR: Callbacks set")

	// DEBUG: Log callback registration
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogUIInteraction("Toolbar", "callbacks_set", map[string]interface{}{
			"image_loaded_set": onImageLoaded != nil,
			"image_saved_set":  onImageSaved != nil,
			"reset_set":        onReset != nil,
			"zoom_changed_set": onZoomChanged != nil,
			"view_changed_set": onViewChanged != nil,
			"duration":         duration,
		})
	}
}
