// internal/gui/app.go
// Perfect UI implementation following specification document
package gui

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"
)

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

	// GUI components following Perfect UI spec
	toolbar     *Toolbar     // Top toolbar (50px height)
	leftPanel   *LeftPanel   // 250px wide control sidebar
	centerPanel *CenterPanel // Central image display area
	rightPanel  *RightPanel  // 250px wide metrics sidebar

	// Layout containers
	mainContainer *fyne.Container
}

func NewApplication(app fyne.App, logger *slog.Logger, debugMode bool) *Application {
	window := app.NewWindow("Advanced Image Processing v2.0")
	window.Resize(fyne.NewSize(1600, 900))
	window.CenterOnScreen()

	appInstance := &Application{
		app:       app,
		window:    window,
		logger:    logger,
		debugMode: debugMode,
	}

	appInstance.initializeCore()
	appInstance.initializeComponents()
	appInstance.setupLayout()
	appInstance.setupCallbacks()
	appInstance.setupTheme()

	return appInstance
}

func (a *Application) initializeCore() {
	a.imageData = core.NewImageData()
	a.regionManager = core.NewRegionManager()
	a.pipeline = core.NewEnhancedPipeline(a.imageData, a.regionManager, a.logger)
	a.loader = io.NewImageLoader(a.logger)
}

func (a *Application) initializeComponents() {
	a.toolbar = NewToolbar(a.imageData, a.loader, a.pipeline, a.logger)
	a.leftPanel = NewLeftPanel(a.pipeline, a.regionManager, a.imageData, a.logger)
	a.centerPanel = NewCenterPanel(a.imageData, a.regionManager, a.logger)
	a.rightPanel = NewRightPanel(a.logger)
}

func (a *Application) setupLayout() {
	// Perfect UI Layout: Toolbar (top) | Left (250px) | Center | Right (250px)
	leftContent := a.leftPanel.GetContainer()
	centerContent := a.centerPanel.GetContainer()
	rightContent := a.rightPanel.GetContainer()
	toolbarContent := a.toolbar.GetContainer()

	// Create horizontal split: Left | Center | Right
	rightSplit := container.NewHSplit(centerContent, rightContent)
	rightSplit.SetOffset(0.8) // Center gets 80% of remaining space

	mainSplit := container.NewHSplit(leftContent, rightSplit)
	mainSplit.SetOffset(0.2) // Left gets 20% of total space

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

func (a *Application) setupCallbacks() {
	// Pipeline callbacks for real-time preview
	a.pipeline.SetCallbacks(
		func(preview image.Image, metrics map[string]float64) {
			a.centerPanel.UpdatePreview(preview)
			a.rightPanel.UpdateMetrics(metrics)
		},
		func(err error) {
			a.rightPanel.ShowError(err)
		},
	)

	// Toolbar callbacks
	a.toolbar.SetCallbacks(
		func(filepath string) {
			a.onImageLoaded(filepath)
		},
		func(filepath string) {
			a.onImageSaved(filepath)
		},
		func() {
			a.onReset()
		},
		func(zoom float64) {
			a.centerPanel.SetZoom(zoom)
		},
		func(viewMode string) {
			a.centerPanel.SetViewMode(viewMode)
		},
	)

	// Left panel callbacks
	a.leftPanel.SetCallbacks(
		func(layerMode bool) {
			a.pipeline.SetProcessingMode(layerMode)
		},
	)

	// Center panel callbacks
	a.centerPanel.SetCallbacks(
		func(hasSelection bool) {
			a.leftPanel.UpdateSelectionState(hasSelection)
		},
	)
}

func (a *Application) onImageLoaded(filepath string) {
	a.centerPanel.UpdateOriginal()
	a.leftPanel.EnableProcessing()

	metadata := a.imageData.GetMetadata()
	a.rightPanel.ShowImageInfo(filepath, metadata.Width, metadata.Height, metadata.Channels)

	// Show original image in preview initially
	original := a.imageData.GetOriginal()
	if !original.Empty() {
		if img, err := original.ToImage(); err == nil {
			a.centerPanel.UpdatePreview(img)
		}
	}
	original.Close()

	a.logger.Info("Image loaded successfully", "filepath", filepath)
}

func (a *Application) onImageSaved(filepath string) {
	a.rightPanel.ShowMessage(fmt.Sprintf("Saved: %s", filepath))
	a.logger.Info("Image saved successfully", "filepath", filepath)
}

func (a *Application) onReset() {
	a.pipeline.ClearAll()
	a.imageData.ResetToOriginal()
	a.centerPanel.Reset()
	a.leftPanel.Reset()
	a.rightPanel.Clear()

	// Show original image after reset
	original := a.imageData.GetOriginal()
	if !original.Empty() {
		if img, err := original.ToImage(); err == nil {
			a.centerPanel.UpdatePreview(img)
		}
	}
	original.Close()
}

func (a *Application) setupTheme() {
	a.app.Settings().SetTheme(&PerfectUITheme{})
}

func (a *Application) ShowAndRun() {
	a.logger.Info("Starting Advanced Image Processing v2.0 with Perfect UI")

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

// PerfectUITheme implements the Perfect UI color scheme
type PerfectUITheme struct{}

func (t *PerfectUITheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255} // #FFFFFF
	case theme.ColorNameForeground:
		return color.RGBA{R: 0, G: 0, B: 0, A: 255} // #000000
	case theme.ColorNamePrimary:
		return color.RGBA{R: 0, G: 123, B: 255, A: 255} // #007BFF
	case theme.ColorNameFocus:
		return color.RGBA{R: 0, G: 86, B: 179, A: 255} // #0056B3
	case theme.ColorNameHover:
		return color.RGBA{R: 0, G: 86, B: 179, A: 255} // #0056B3
	case theme.ColorNameSuccess:
		return color.RGBA{R: 40, G: 167, B: 69, A: 255} // #28A745
	case theme.ColorNameError:
		return color.RGBA{R: 220, G: 20, B: 60, A: 255} // #DC143C
	case theme.ColorNameWarning:
		return color.RGBA{R: 255, G: 193, B: 7, A: 255} // #FFC107
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255} // #FFFFFF
	case theme.ColorNameButton:
		return color.RGBA{R: 0, G: 123, B: 255, A: 255} // #007BFF
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 211, G: 211, B: 211, A: 255} // #D3D3D3
	case theme.ColorNameSeparator:
		return color.RGBA{R: 211, G: 211, B: 211, A: 255} // #D3D3D3
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (t *PerfectUITheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (t *PerfectUITheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (t *PerfectUITheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 5
	case theme.SizeNameInlineIcon:
		return 16
	case theme.SizeNameText:
		return 12
	case theme.SizeNameCaptionText:
		return 10
	case theme.SizeNameHeadingText:
		return 16
	default:
		return theme.DefaultTheme().Size(name)
	}
}
