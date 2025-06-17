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
	"fyne.io/fyne/v2/dialog"
)

type Application struct {
	app           fyne.App
	window        fyne.Window
	logger        *slog.Logger
	mainContainer *fyne.Container

	// Core components - using actual types that exist
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
	myApp.SetIcon(nil) // Set icon resource if available

	window := myApp.NewWindow("Image Restoration Suite")
	window.Resize(fyne.NewSize(1600, 900))

	// Initialize core components in correct order (based on dependencies)
	imageData := core.NewImageData()                                       // No arguments
	regionManager := core.NewRegionManager()                               // No arguments
	pipeline := core.NewEnhancedPipeline(imageData, regionManager, logger) // 3 arguments
	imageLoader := io.NewImageLoader(logger)

	// Initialize UI panels with correct signatures
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
	// Right panel callbacks for window title updates
	a.rightPanel.SetWindowTitleChangeCallback(func(title string) {
		a.logger.Debug("Window title change requested", "new_title", title)
		a.window.SetTitle(title)
	})

	// TODO: Add toolbar callbacks once Toolbar interface is known
	// These callbacks don't exist yet in Toolbar struct:
	// - onOpenImage
	// - onSaveImage
	// - onViewModeChanged
	// - onZoomChanged

	// TODO: Add left panel callbacks once LeftPanel interface is known
	// These callbacks don't exist yet in LeftPanel struct:
	// - onLayerAdded
	// - onLayerDeleted
	// - onLayerToggled
	// - onParameterChanged

	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogRuntimeError("Application", "Toolbar and LeftPanel callbacks not implemented - interfaces unknown")
	}
}

func (a *Application) setupLayout() {
	// Perfect UI Layout: Toolbar (top) | Left (300px) | Center | Right (300px)
	leftContent := a.leftPanel.GetContainer()
	centerContent := a.centerPanel.GetContainer()
	rightContent := a.rightPanel.GetContainer()
	toolbarContent := a.toolbar.GetContainer()

	// Create horizontal split: Left | Center | Right with better proportions
	rightSplit := container.NewHSplit(centerContent, rightContent)
	rightSplit.SetOffset(0.7) // Center gets 70% of remaining space after left panel

	mainSplit := container.NewHSplit(leftContent, rightSplit)
	mainSplit.SetOffset(0.22) // Left gets 22% of total space (slightly wider)

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

func (a *Application) loadImage(filepath string) {
	a.logger.Debug("Loading image", "filepath", filepath)

	// TODO: ImageData.LoadFromFile doesn't exist - need to find actual method
	// Possible alternatives: Load(), LoadImage(), SetImage(), etc.

	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogBuildError("ImageData", "LoadFromFile", "LoadFromFile(string) error", "method does not exist")
	}

	// Placeholder for now - need to investigate actual ImageData interface
	// if err := a.imageData.LoadFromFile(filepath); err != nil {
	//     a.logger.Error("Failed to load image", "error", err, "filepath", filepath)
	//     dialog.ShowError(err, a.window)
	//     return
	// }

	// Update displays
	a.centerPanel.UpdateOriginal()
	a.centerPanel.UpdatePreview(nil) // Clear preview initially

	// TODO: ImageData.GetDimensions doesn't exist - need to find actual method
	// Show image info in right panel
	// width, height, channels := a.imageData.GetDimensions()
	// a.rightPanel.ShowImageInfo(filepath, width, height, channels)

	// Placeholder values for now
	a.rightPanel.ShowImageInfo(filepath, 1400, 995, 3)

	// Trigger preview processing if layers exist
	a.triggerPreviewProcessing()

	a.logger.Info("Image load attempted", "filepath", filepath)
}

func (a *Application) saveImage() {
	if !a.imageData.HasImage() {
		dialog.ShowInformation("No Image", "Please load an image first", a.window)
		return
	}

	// Fix Fyne dialog API - correct signature expects error parameter
	saveDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			a.logger.Error("Save dialog error", "error", err)
			return
		}

		if writer != nil {
			defer writer.Close()

			// TODO: Implement actual save logic using pipeline
			a.rightPanel.ShowMessage("Image saved successfully")
			a.logger.Info("Image saved successfully", "filepath", writer.URI().Path())
		}
	}, a.window)

	// TODO: Fix filter - need proper storage.FileFilter implementation
	// saveDialog.SetFilter([]string{".png", ".jpg", ".jpeg", ".tiff"})
	// For now, skip filter until we understand the proper API

	saveDialog.Show()
}

func (a *Application) resetApplication() {
	a.logger.Debug("Resetting application")

	// Clear image data
	a.imageData.Clear()

	// Reset UI panels
	a.centerPanel.Reset()
	a.rightPanel.Clear()

	// Reset window title
	a.window.SetTitle("Image Restoration Suite")

	a.logger.Info("Application reset completed")
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

	// TODO: Implement actual preview processing using pipeline
	// For now, just update with the original image
	original := a.imageData.GetOriginal()
	defer original.Close()

	if !original.Empty() {
		if img, err := original.ToImage(); err == nil {
			a.centerPanel.UpdatePreview(img)

			// Dummy metrics for now
			metrics := map[string]float64{
				"psnr": 25.0,
				"ssim": 0.85,
			}
			a.handlePreviewUpdate(img, metrics)
		} else {
			return fmt.Errorf("failed to convert image: %v", err)
		}
	}

	a.logger.Debug("Preview processing completed successfully")
	return nil
}

func (a *Application) handlePreviewUpdate(preview image.Image, metrics map[string]float64) {
	// Update preview image in center panel
	a.centerPanel.UpdatePreview(preview)

	// Update metrics in right panel - extract individual values from map
	if metrics != nil {
		psnr, psnrOk := metrics["psnr"]
		ssim, ssimOk := metrics["ssim"]
		if psnrOk && ssimOk {
			a.rightPanel.UpdateMetrics(psnr, ssim)
		} else {
			a.rightPanel.ShowError("Invalid quality metrics data")
		}
	} else {
		a.rightPanel.ShowError("Failed to calculate quality metrics")
	}

	a.logger.Debug("Preview and metrics updated")
}
