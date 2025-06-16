// Updated main application with improved layout and proportions
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

// Application represents the main application with enhanced UI
type Application struct {
	app       fyne.App
	window    fyne.Window
	logger    *slog.Logger
	debugMode bool

	// Core components
	imageData     *core.ImageData
	regionManager *core.RegionManager
	pipeline      *core.EnhancedPipeline
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
	leftPanels  *container.Split
	centerPanel *fyne.Container
	rightPanels *container.Split
	statusCard  *widget.Card
}

func NewApplication(app fyne.App, logger *slog.Logger, debugMode bool) *Application {
	window := app.NewWindow("🎨 Advanced Image Processing v2.0 - Layer Edition")
	window.Resize(fyne.NewSize(1800, 1200))
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
	// Create enhanced toolbar with spacing
	toolbarContainer := container.NewVBox(
		widget.NewCard("🛠️ Tools", "", a.toolbar.GetContainer()),
		widget.NewSeparator(),
	)

	// Create center panel with image canvas
	a.centerPanel = container.NewBorder(
		toolbarContainer, // top
		nil,              // bottom
		nil,              // left
		nil,              // right
		container.NewPadded(a.canvas.GetContainer()), // center with padding
	)

	// Create left panel with layer management and properties
	a.leftPanels = container.NewVSplit(
		container.NewScroll(a.layerPanel.GetContainer()),
		container.NewScroll(a.properties.GetContainer()),
	)
	a.leftPanels.SetOffset(0.6) // Give more space to layer panel

	// Create status card for tracking
	a.statusCard = widget.NewCard("📊 Status", "",
		widget.NewLabel("Application ready for image processing"))

	// Create right panel with metrics and status
	a.rightPanels = container.NewVSplit(
		a.statusCard,
		a.metricsPanel.GetContainer(),
	)
	a.rightPanels.SetOffset(0.3) // Give more space to metrics

	// Create main three-panel layout
	centerAndRight := container.NewHSplit(
		a.centerPanel,
		a.rightPanels,
	)
	centerAndRight.SetOffset(0.75) // Give most space to center panel

	a.mainContent = container.NewHSplit(
		a.leftPanels,
		centerAndRight,
	)
	a.mainContent.SetOffset(0.3) // Balanced left panel size

	// Set window properties
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
				a.updateStatusMessage(fmt.Sprintf("✅ Loaded: %s", filepath))
			})
		},
		// onImageSaved
		func(filepath string) {
			fyne.Do(func() {
				a.showInfo("💾 Image Saved", fmt.Sprintf("Image successfully saved to:\n%s", filepath))
				a.updateStatusMessage(fmt.Sprintf("💾 Saved: %s", filepath))
			})
		},
	)

	// Toolbar callbacks
	a.toolbar.SetCallbacks(
		// onToolChanged
		func(tool string) {
			a.canvas.SetActiveTool(tool)
			a.updateStatusMessage(fmt.Sprintf("🛠️ Tool: %s", tool))
		},
		// onClearSelection
		func() {
			a.regionManager.ClearAll()
			a.canvas.RefreshSelections()
			a.layerPanel.Refresh()
			a.updateStatusMessage("🗑️ Selection cleared")
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
			a.updateStatusMessage("↻ Reset to original image")
		}
	})

	// Canvas callbacks
	a.canvas.SetCallbacks(
		// onSelectionChanged
		func(hasSelection bool) {
			fyne.Do(func() {
				a.toolbar.SetSelectionState(hasSelection)
				a.layerPanel.Refresh()
				if hasSelection {
					a.updateStatusMessage("🎯 Region selected")
				} else {
					a.updateStatusMessage("📄 No selection")
				}
			})
		},
	)

	// Layer panel selection change callback
	a.layerPanel.SetSelectionChangedCallback(func() {
		a.layerPanel.Refresh()
	})
}

func (a *Application) updateStatusMessage(message string) {
	// Update the status card directly
	if a.statusCard != nil {
		a.statusCard.SetContent(widget.NewLabel(message))
	}
}

func (a *Application) ShowAndRun() {
	a.logger.Info("Showing main application window with enhanced UI")

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
	a.updateStatusMessage(fmt.Sprintf("❌ Error: %s", err.Error()))
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
		a.updateStatusMessage(fmt.Sprintf("✅ Image loaded: %s", filepath))
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
