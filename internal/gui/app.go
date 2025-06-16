// Updated main application with layer support
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"
)

// Application represents the main application with layer support
type Application struct {
	app       fyne.App
	window    fyne.Window
	logger    *slog.Logger
	debugMode bool

	// Core components
	imageData     *core.ImageData
	regionManager *core.RegionManager
	pipeline      *core.EnhancedPipeline // Updated to enhanced pipeline
	loader        *io.ImageLoader

	// GUI components
	canvas       *ImageCanvas
	toolbar      *Toolbar
	properties   *EnhancedPropertiesPanel
	layerPanel   *LayerPanel
	metricsPanel *MetricsPanel
	menuHandler  *MenuHandler

	// Layout containers
	mainContent *container.Split
	rightPanel  *container.Split
	leftPanel   *container.Split
}

func NewApplication(app fyne.App, logger *slog.Logger, debugMode bool) *Application {
	window := app.NewWindow("Advanced Image Processing v2.0 - Layer Edition")
	window.Resize(fyne.NewSize(1600, 1000))
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
	a.pipeline = core.NewEnhancedPipeline(a.imageData, a.regionManager, a.logger)
	a.loader = io.NewImageLoader(a.logger)
}

func (a *Application) initializeGUI() {
	a.canvas = NewImageCanvas(a.imageData, a.regionManager, a.logger)
	a.toolbar = NewToolbar()
	a.properties = NewEnhancedPropertiesPanel(a.pipeline, a.logger)
	a.layerPanel = NewLayerPanel(a.pipeline, a.regionManager, a.logger)
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

	// Create left panel with layer management
	a.leftPanel = container.NewVSplit(
		a.layerPanel.GetContainer(),
		a.properties.GetContainer(),
	)
	a.leftPanel.SetOffset(0.5)

	// Create right panel with metrics
	a.rightPanel = container.NewVSplit(
		container.NewVBox(), // Placeholder for future panels
		a.metricsPanel.GetContainer(),
	)
	a.rightPanel.SetOffset(0.3)

	// Create main content with three-panel layout
	centerAndRight := container.NewHSplit(
		canvasContainer,
		a.rightPanel,
	)
	centerAndRight.SetOffset(0.8)

	a.mainContent = container.NewHSplit(
		a.leftPanel,
		centerAndRight,
	)
	a.mainContent.SetOffset(0.25)

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
				a.canvas.ClearPreview()
				a.canvas.UpdateOriginalImage()
				a.properties.Enable()
				a.layerPanel.Enable()
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
			a.layerPanel.Refresh()
		},
	)

	a.toolbar.SetResetCallback(func() {
		if a.imageData.HasImage() {
			a.pipeline.ClearAll()
			a.imageData.ResetToOriginal()
			a.canvas.UpdateOriginalImage()
			a.canvas.ClearPreview()
			a.metricsPanel.Clear()
			a.layerPanel.Refresh()
		}
	})

	// Canvas callbacks
	a.canvas.SetCallbacks(
		// onSelectionChanged
		func(hasSelection bool) {
			fyne.Do(func() {
				a.toolbar.SetSelectionState(hasSelection)
				a.layerPanel.Refresh()
			})
		},
	)

	// Layer panel selection change callback
	a.layerPanel.SetSelectionChangedCallback(func() {
		// Refresh UI when selections change
		a.layerPanel.Refresh()
	})
}

func (a *Application) ShowAndRun() {
	a.logger.Info("Showing main application window with layer support")

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

func (a *Application) RefreshUI() {
	fyne.Do(func() {
		a.canvas.Refresh()
		a.properties.Refresh()
		a.layerPanel.Refresh()
		a.metricsPanel.Refresh()
		a.toolbar.Refresh()
	})
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
	a.pipeline.ClearAll()

	// Update UI
	fyne.Do(func() {
		a.canvas.ClearPreview()
		a.canvas.UpdateOriginalImage()
		a.properties.Enable()
		a.layerPanel.Enable()
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
