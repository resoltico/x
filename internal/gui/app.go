package gui

import (
	"fmt"
	"image"
	"log/slog"
	"time"

	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

type Application struct {
	app           fyne.App
	window        fyne.Window
	logger        *slog.Logger
	mainContainer *fyne.Container

	// Core components
	imageData     *core.ImageData
	pipeline      *core.EnhancedPipeline
	regionManager *core.RegionManager
	imageLoader   *io.ImageLoader

	// UI panels
	toolbar     *Toolbar
	leftPanel   *LeftPanel
	centerPanel *CenterPanel
	rightPanel  *RightPanel

	// Preview processing
	previewTimer *time.Timer
}

func NewApplication(logger *slog.Logger) *Application {
	// Initialize GUI debugger
	InitGUIDebugger(logger)

	myApp := app.NewWithID("advanced-image-processing")
	myApp.SetIcon(nil)

	window := myApp.NewWindow("Image Restoration Suite")
	window.Resize(fyne.NewSize(1600, 900))

	// Initialize core components
	imageData := core.NewImageData()
	regionManager := core.NewRegionManager()
	pipeline := core.NewEnhancedPipeline(imageData, regionManager, logger)
	imageLoader := io.NewImageLoader(logger)

	// Initialize UI panels
	toolbar := NewToolbar(imageData, imageLoader, pipeline, logger)
	leftPanel := NewLeftPanel(pipeline, regionManager, imageData, logger)
	centerPanel := NewCenterPanel(imageData, regionManager, logger)
	rightPanel := NewRightPanel(logger)

	application := &Application{
		app:           myApp,
		window:        window,
		logger:        logger,
		imageData:     imageData,
		pipeline:      pipeline,
		regionManager: regionManager,
		imageLoader:   imageLoader,
		toolbar:       toolbar,
		leftPanel:     leftPanel,
		centerPanel:   centerPanel,
		rightPanel:    rightPanel,
	}

	return application
}

func (a *Application) Initialize() error {
	a.logger.Info("Starting Image Restoration Suite with Perfect UI")

	// Setup UI layout
	a.setupLayout()

	// Setup callbacks between components
	a.SetupCallbacks()

	return nil
}

func (a *Application) SetupCallbacks() {
	// Connect toolbar callbacks
	a.toolbar.SetCallbacks(
		a.handleImageLoaded, // onImageLoaded
		a.handleImageSaved,  // onImageSaved
		a.handleReset,       // onReset
		a.handleZoomChanged, // onZoomChanged
		a.handleViewChanged, // onViewChanged
	)

	// Connect pipeline preview callback
	a.pipeline.SetCallbacks(
		a.handlePreviewUpdate, // onPreviewUpdate
		a.handleError,         // onError
	)

	// Connect right panel window title callback
	a.rightPanel.SetWindowTitleChangeCallback(func(title string) {
		a.window.SetTitle(title)
	})

	// Connect left panel parameter change callback
	a.leftPanel.SetCallbacks(a.handleParameterChanged)

	// Connect center panel selection callback
	a.centerPanel.SetCallbacks(a.handleSelectionChanged)
}

func (a *Application) setupLayout() {
	leftContent := a.leftPanel.GetContainer()
	centerContent := a.centerPanel.GetContainer()
	rightContent := a.rightPanel.GetContainer()
	toolbarContent := a.toolbar.GetContainer()

	// Create horizontal split
	rightSplit := container.NewHSplit(centerContent, rightContent)
	rightSplit.SetOffset(0.7)

	mainSplit := container.NewHSplit(leftContent, rightSplit)
	mainSplit.SetOffset(0.22)

	// Main container with toolbar at top
	a.mainContainer = container.NewBorder(
		toolbarContent, // top
		nil,            // bottom
		nil,            // left
		nil,            // right
		mainSplit,      // center
	)

	a.window.SetContent(a.mainContainer)
}

func (a *Application) Run() {
	a.window.ShowAndRun()
}

// Callback handlers
func (a *Application) handleImageLoaded(filepath string) {
	a.logger.Debug("Image loaded callback", "filepath", filepath)

	// Update displays
	a.centerPanel.UpdateOriginal()
	a.centerPanel.UpdatePreview(nil) // Clear preview initially

	// Show image info
	metadata := a.imageData.GetMetadata()
	a.rightPanel.ShowImageInfo(filepath, metadata.Width, metadata.Height, metadata.Channels)

	// Enable left panel
	a.leftPanel.EnableProcessing()

	// Reset pipeline to original
	a.imageData.ResetToOriginal()

	// Trigger initial preview (show original as preview)
	a.triggerPreviewProcessing()

	a.logger.Info("Image loaded and UI updated", "filepath", filepath)
}

func (a *Application) handleImageSaved(filepath string) {
	a.rightPanel.ShowMessage("Image saved successfully")
	a.logger.Info("Image saved", "filepath", filepath)
}

func (a *Application) handleReset() {
	a.logger.Debug("Reset callback")

	// Clear image data
	a.imageData.Clear()

	// Reset pipeline
	a.pipeline.ClearAll()

	// Reset UI panels
	a.centerPanel.Reset()
	a.rightPanel.Clear()
	a.leftPanel.Reset()
	a.leftPanel.Disable()

	// Reset window title
	a.window.SetTitle("Image Restoration Suite")

	a.logger.Info("Application reset completed")
}

func (a *Application) handleZoomChanged(zoom float64) {
	a.centerPanel.SetZoom(zoom)
	a.logger.Debug("Zoom changed", "zoom", zoom)
}

func (a *Application) handleViewChanged(view string) {
	a.centerPanel.SetViewMode(view)
	a.logger.Debug("View changed", "view", view)
}

func (a *Application) handlePreviewUpdate(preview image.Image, metrics map[string]float64) {
	a.logger.Debug("Preview update callback", "metrics_count", len(metrics))

	// Update preview image in center panel
	a.centerPanel.UpdatePreview(preview)

	// Update metrics in right panel
	if metrics != nil {
		if psnr, ok := metrics["psnr"]; ok {
			if ssim, ok := metrics["ssim"]; ok {
				a.rightPanel.UpdateMetrics(psnr, ssim)
			}
		}
	}

	a.logger.Debug("Preview and metrics updated")
}

func (a *Application) handleError(err error) {
	a.logger.Error("Pipeline error", "error", err)
	a.rightPanel.ShowError(err.Error())
}

func (a *Application) handleParameterChanged(layerID string, params map[string]interface{}) {
	a.logger.Debug("Parameter changed", "layer_id", layerID)
	// Pipeline automatically handles real-time updates
}

func (a *Application) handleSelectionChanged(hasSelection bool) {
	a.leftPanel.UpdateSelectionState(hasSelection)
	a.logger.Debug("Selection changed", "has_selection", hasSelection)
}

func (a *Application) triggerPreviewProcessing() {
	// Cancel existing timer
	if a.previewTimer != nil {
		a.previewTimer.Stop()
	}

	// Schedule new processing with debounce delay
	delay := 200 * time.Millisecond
	a.logger.Debug("Scheduling preview processing", "delay_ms", delay.Milliseconds())

	a.previewTimer = time.AfterFunc(delay, func() {
		a.logger.Debug("Starting preview processing")
		if err := a.handlePreviewProcessing(); err != nil {
			a.logger.Error("Preview processing failed", "error", err)
			a.rightPanel.ShowError("Preview processing failed: " + err.Error())
		}
	})
}

func (a *Application) handlePreviewProcessing() error {
	if !a.imageData.HasImage() {
		a.logger.Debug("No image available for preview processing")
		return nil
	}

	// Get original as fallback preview
	original := a.imageData.GetOriginal()
	defer original.Close()

	if !original.Empty() {
		if img, err := original.ToImage(); err == nil {
			a.centerPanel.UpdatePreview(img)

			// Show original image metrics (comparing to itself = perfect)
			metrics := map[string]float64{
				"psnr": 100.0, // Perfect match
				"ssim": 1.0,   // Perfect match
			}
			a.handlePreviewUpdate(img, metrics)
		} else {
			return fmt.Errorf("failed to convert image: %v", err)
		}
	}

	a.logger.Debug("Preview processing completed successfully")
	return nil
}
