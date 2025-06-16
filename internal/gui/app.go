// Main application GUI with real-time preview
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"
)

// Application represents the main application
type Application struct {
	app       fyne.App
	window    fyne.Window
	logger    *slog.Logger
	debugMode bool

	// Core components
	imageData     *core.ImageData
	regionManager *core.RegionManager
	pipeline      *core.ProcessingPipeline
	loader        *io.ImageLoader

	// GUI components
	canvas       *ImageCanvas
	toolbar      *Toolbar
	properties   *EnhancedPropertiesPanel
	metricsPanel *MetricsPanel
	menuHandler  *MenuHandler

	// Layout containers
	mainContent *container.Split
	rightPanel  *container.Split
}

func NewApplication(app fyne.App, logger *slog.Logger, debugMode bool) *Application {
	window := app.NewWindow("Advanced Image Processing v2.0")
	window.Resize(fyne.NewSize(1400, 900))
	window.CenterOnScreen()

	appInstance := &Application{
		app:       app,
		window:    window,
		logger:    logger,
		debugMode: debugMode,
	}

	appInstance.initializeCore()
	appInstance.initializeGUI()
	appInstance.setupLayout()
	appInstance.setupCallbacks()

	return appInstance
}

func (a *Application) initializeCore() {
	a.imageData = core.NewImageData()
	a.regionManager = core.NewRegionManager()
	a.pipeline = core.NewProcessingPipeline(a.imageData, a.regionManager, a.logger)
	a.loader = io.NewImageLoader(a.logger)
}

func (a *Application) initializeGUI() {
	a.canvas = NewImageCanvas(a.imageData, a.regionManager, a.logger)
	a.toolbar = NewToolbar()
	a.properties = NewEnhancedPropertiesPanel(a.pipeline, a.logger)
	a.metricsPanel = NewMetricsPanel()
	a.menuHandler = NewMenuHandler(a.window, a.imageData, a.loader, a.logger)
}

func (a *Application) setupLayout() {
	// Create toolbar
	toolbarContainer := container.NewVBox(
		a.toolbar.GetContainer(),
		widget.NewSeparator(),
	)

	// Create main canvas area
	canvasContainer := container.NewBorder(
		toolbarContainer,
		nil,
		nil,
		nil,
		a.canvas.GetContainer(),
	)

	// Create right panel with properties and metrics
	a.rightPanel = container.NewVSplit(
		a.properties.GetContainer(),
		a.metricsPanel.GetContainer(),
	)
	a.rightPanel.SetOffset(0.6)

	// Create main content split
	a.mainContent = container.NewHSplit(
		canvasContainer,
		a.rightPanel,
	)
	a.mainContent.SetOffset(0.75)

	a.window.SetMainMenu(a.menuHandler.GetMainMenu())
	a.window.SetContent(a.mainContent)
}

func (a *Application) setupCallbacks() {
	// Pipeline callbacks for real-time preview
	a.pipeline.SetCallbacks(
		// onPreviewUpdate
		func(preview gocv.Mat, metrics map[string]float64) {
			fyne.Do(func() {
				a.canvas.UpdatePreview(preview)
				a.metricsPanel.UpdateMetrics(metrics)
			})
		},
		// onError
		func(err error) {
			fyne.Do(func() {
				a.showError("Processing Error", err)
			})
		},
	)

	// Menu callbacks
	a.menuHandler.SetCallbacks(
		// onImageLoaded
		func(filepath string) {
			fyne.Do(func() {
				// Clear preview on new image load
				a.canvas.ClearPreview()
				a.canvas.UpdateOriginalImage()
				a.properties.Enable()
				a.toolbar.Enable()
				a.metricsPanel.Clear()
			})
		},
		// onImageSaved
		func(filepath string) {
			fyne.Do(func() {
				a.showInfo("Image Saved", fmt.Sprintf("Image saved to: %s", filepath))
			})
		},
	)

	// Toolbar callbacks
	a.toolbar.SetCallbacks(
		// onToolChanged
		func(tool string) {
			a.canvas.SetActiveTool(tool)
		},
		// onClearSelection
		func() {
			a.regionManager.ClearAll()
			a.canvas.RefreshSelections()
		},
	)

	a.toolbar.SetResetCallback(func() {
		if a.imageData.HasImage() {
			a.pipeline.ClearSteps()
			a.imageData.ResetToOriginal()
			a.canvas.UpdateOriginalImage()
			a.canvas.ClearPreview()
			a.metricsPanel.Clear()
		}
	})

	// Canvas callbacks
	a.canvas.SetCallbacks(
		// onSelectionChanged
		func(hasSelection bool) {
			fyne.Do(func() {
				a.toolbar.SetSelectionState(hasSelection)
			})
		},
	)
}

func (a *Application) ShowAndRun() {
	a.logger.Info("Showing main application window")

	a.window.SetCloseIntercept(func() {
		a.cleanup()
		a.app.Quit()
	})

	a.window.ShowAndRun()
}

func (a *Application) cleanup() {
	a.logger.Info("Cleaning up application resources")
	a.pipeline.Stop()
	a.imageData.Close()
	a.regionManager.ClearAll()
}

func (a *Application) showError(title string, err error) {
	a.logger.Error(title, "error", err)
	dialog.ShowError(err, a.window)
}

func (a *Application) showInfo(title, message string) {
	a.logger.Info(title, "message", message)
	dialog.ShowInformation(title, message, a.window)
}

func (a *Application) showWarning(title, message string) {
	a.logger.Warn(title, "message", message)

	content := container.NewVBox(
		widget.NewIcon(theme.WarningIcon()),
		widget.NewLabel(message),
	)

	warningDialog := dialog.NewCustom(title, "OK", content, a.window)
	warningDialog.Show()
}

func (a *Application) GetWindow() fyne.Window {
	return a.window
}

func (a *Application) GetImageData() *core.ImageData {
	return a.imageData
}

func (a *Application) GetRegionManager() *core.RegionManager {
	return a.regionManager
}

func (a *Application) GetPipeline() *core.ProcessingPipeline {
	return a.pipeline
}

func (a *Application) RefreshUI() {
	fyne.Do(func() {
		a.canvas.Refresh()
		a.properties.Refresh()
		a.metricsPanel.Refresh()
		a.toolbar.Refresh()
	})
}

func (a *Application) SetStatus(message string) {
	a.logger.Debug("Status update", "status", message)
}

func (a *Application) ToggleDebugMode() {
	a.debugMode = !a.debugMode
	a.logger.Info("Debug mode toggled", "debug_mode", a.debugMode)
	a.RefreshUI()
}

func (a *Application) GetDebugMode() bool {
	return a.debugMode
}

func (a *Application) LoadImageFromPath(filepath string) error {
	mat, err := a.loader.LoadImage(filepath)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}
	defer mat.Close()

	if err := core.ValidateImage(mat); err != nil {
		return fmt.Errorf("invalid image: %w", err)
	}

	if err := a.imageData.SetOriginal(mat, filepath); err != nil {
		return fmt.Errorf("failed to set image: %w", err)
	}

	// Clear previous processing state
	a.regionManager.ClearAll()
	a.pipeline.ClearSteps()

	// Update UI
	fyne.Do(func() {
		a.canvas.ClearPreview()
		a.canvas.UpdateOriginalImage()
		a.properties.Enable()
		a.toolbar.Enable()
		a.metricsPanel.Clear()
	})

	a.logger.Info("Image loaded successfully", "filepath", filepath)
	return nil
}

func (a *Application) SaveProcessedImage(filepath string) error {
	if !a.imageData.HasImage() {
		return fmt.Errorf("no image to save")
	}

	// Process full resolution
	processed, err := a.pipeline.ProcessFullResolution()
	if err != nil {
		// Fall back to original if processing fails
		processed = a.imageData.GetOriginal()
	}
	defer processed.Close()

	if err := a.loader.SaveImage(processed, filepath); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	a.logger.Info("Image saved successfully", "filepath", filepath)
	return nil
}
