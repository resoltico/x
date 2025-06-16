// Main application GUI with modern Fyne v2.6 architecture
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"
)

// Application represents the main application
type Application struct {
	app       fyne.App
	window    fyne.Window
	logger    *logrus.Logger
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

// NewApplication creates a new application instance
func NewApplication(app fyne.App, logger *logrus.Logger, debugMode bool) *Application {
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

// initializeCore initializes core components
func (a *Application) initializeCore() {
	a.imageData = core.NewImageData()
	a.regionManager = core.NewRegionManager()
	a.pipeline = core.NewProcessingPipeline(a.imageData, a.regionManager, a.logger)
	a.loader = io.NewImageLoader(a.logger)
}

// initializeGUI initializes GUI components
func (a *Application) initializeGUI() {
	// Create main GUI components
	a.canvas = NewImageCanvas(a.imageData, a.regionManager, a.logger)
	a.toolbar = NewToolbar()
	a.properties = NewEnhancedPropertiesPanel(a.pipeline, a.logger)
	a.metricsPanel = NewMetricsPanel()
	a.menuHandler = NewMenuHandler(a.window, a.imageData, a.loader, a.logger)
}

// setupLayout creates the application layout
func (a *Application) setupLayout() {
	// Create toolbar
	toolbarContainer := container.NewVBox(
		a.toolbar.GetContainer(),
		widget.NewSeparator(),
	)

	// Create main canvas area
	canvasContainer := container.NewBorder(
		toolbarContainer, // top
		nil,              // bottom
		nil,              // left
		nil,              // right
		a.canvas.GetContainer(),
	)

	// Create right panel with properties and metrics
	a.rightPanel = container.NewVSplit(
		a.properties.GetContainer(),
		a.metricsPanel.GetContainer(),
	)
	a.rightPanel.SetOffset(0.6) // 60% for properties, 40% for metrics

	// Create main content split
	a.mainContent = container.NewHSplit(
		canvasContainer,
		a.rightPanel,
	)
	a.mainContent.SetOffset(0.75) // 75% for canvas, 25% for right panel

	// Set main menu
	a.window.SetMainMenu(a.menuHandler.GetMainMenu())

	// Set window content
	a.window.SetContent(a.mainContent)
}

// setupCallbacks sets up component callbacks
func (a *Application) setupCallbacks() {
	// Pipeline callbacks
	a.pipeline.SetCallbacks(
		// onProgress
		func(step, total int, stepName string) {
			fyne.Do(func() {
				a.properties.UpdateProgress(step, total, stepName)
			})
		},
		// onComplete
		func(result gocv.Mat, metrics map[string]float64) {
			fyne.Do(func() {
				a.canvas.UpdateProcessedImage(result)
				a.metricsPanel.UpdateMetrics(metrics)
				a.properties.ClearProgress()
			})
		},
		// onError
		func(err error) {
			fyne.Do(func() {
				a.showError("Processing Error", err)
				a.properties.ClearProgress()
			})
		},
	)

	// Menu callbacks
	a.menuHandler.SetCallbacks(
		// onImageLoaded
		func(filepath string) {
			fyne.Do(func() {
				a.canvas.UpdateOriginalImage()
				a.properties.Enable()
				a.toolbar.Enable()
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
			a.imageData.ResetToOriginal()
			a.pipeline.ClearSteps()
			a.canvas.UpdateOriginalImage()
			a.canvas.UpdateProcessedImage(a.imageData.GetOriginal())
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

// ShowAndRun displays the application window and starts the event loop
func (a *Application) ShowAndRun() {
	a.logger.Info("Showing main application window")

	// Set up window close handler
	a.window.SetCloseIntercept(func() {
		a.cleanup()
		a.app.Quit()
	})

	a.window.ShowAndRun()
}

// cleanup performs cleanup when the application closes
func (a *Application) cleanup() {
	a.logger.Info("Cleaning up application resources")

	// Stop any ongoing processing
	a.pipeline.Stop()

	// Close image data
	a.imageData.Close()

	// Clear regions
	a.regionManager.ClearAll()
}

// showError displays an error dialog
func (a *Application) showError(title string, err error) {
	a.logger.WithError(err).Error(title)
	dialog.ShowError(err, a.window)
}

// showInfo displays an information dialog
func (a *Application) showInfo(title, message string) {
	a.logger.WithField("message", message).Info(title)
	dialog.ShowInformation(title, message, a.window)
}

// showWarning displays a warning dialog
func (a *Application) showWarning(title, message string) {
	a.logger.WithField("message", message).Warn(title)

	content := container.NewVBox(
		widget.NewIcon(theme.WarningIcon()),
		widget.NewLabel(message),
	)

	warningDialog := dialog.NewCustom(title, "OK", content, a.window)
	warningDialog.Show()
}

// GetWindow returns the main window
func (a *Application) GetWindow() fyne.Window {
	return a.window
}

// GetImageData returns the image data
func (a *Application) GetImageData() *core.ImageData {
	return a.imageData
}

// GetRegionManager returns the region manager
func (a *Application) GetRegionManager() *core.RegionManager {
	return a.regionManager
}

// GetPipeline returns the processing pipeline
func (a *Application) GetPipeline() *core.ProcessingPipeline {
	return a.pipeline
}

// RefreshUI refreshes the entire UI
func (a *Application) RefreshUI() {
	fyne.Do(func() {
		a.canvas.Refresh()
		a.properties.Refresh()
		a.metricsPanel.Refresh()
		a.toolbar.Refresh()
	})
}

// SetStatus sets the status message (if we add a status bar later)
func (a *Application) SetStatus(message string) {
	a.logger.WithField("status", message).Debug("Status update")
	// TODO: Implement status bar if needed
}

// ToggleDebugMode toggles debug mode display
func (a *Application) ToggleDebugMode() {
	a.debugMode = !a.debugMode
	a.logger.WithField("debug_mode", a.debugMode).Info("Debug mode toggled")

	// Update UI elements that depend on debug mode
	a.RefreshUI()
}

// GetDebugMode returns current debug mode state
func (a *Application) GetDebugMode() bool {
	return a.debugMode
}

// LoadImageFromPath loads an image from file path
func (a *Application) LoadImageFromPath(filepath string) error {
	mat, err := a.loader.LoadImage(filepath)
	if err != nil {
		return fmt.Errorf("failed to load image: %w", err)
	}
	defer mat.Close()

	// Validate the image
	if err := core.ValidateImage(mat); err != nil {
		return fmt.Errorf("invalid image: %w", err)
	}

	// Set the image
	if err := a.imageData.SetOriginal(mat, filepath); err != nil {
		return fmt.Errorf("failed to set image: %w", err)
	}

	// Clear any existing regions and processing
	a.regionManager.ClearAll()
	a.pipeline.ClearSteps()

	// Update UI
	fyne.Do(func() {
		a.canvas.UpdateOriginalImage()
		a.properties.Enable()
		a.toolbar.Enable()
		a.metricsPanel.Clear()
	})

	a.logger.WithField("filepath", filepath).Info("Image loaded successfully")
	return nil
}

// SaveProcessedImage saves the processed image
func (a *Application) SaveProcessedImage(filepath string) error {
	if !a.imageData.HasImage() {
		return fmt.Errorf("no image to save")
	}

	processed := a.imageData.GetProcessed()
	defer processed.Close()

	if processed.Empty() {
		// Save original if no processing applied
		original := a.imageData.GetOriginal()
		defer original.Close()
		processed = original
	}

	if err := a.loader.SaveImage(processed, filepath); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	a.logger.WithField("filepath", filepath).Info("Image saved successfully")
	return nil
}
