// internal/gui/app.go
// Simplified main application with working layout
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

	// GUI components
	toolbar        *ModernToolbar
	leftPanel      *ControlPanel
	imageWorkspace *ImageWorkspace
	rightPanel     *InfoPanel
	statusManager  *StatusManager

	// Layout
	mainContainer *fyne.Container
}

func NewApplication(app fyne.App, logger *slog.Logger, debugMode bool) *Application {
	window := app.NewWindow("Advanced Image Processing v2.0")
	window.Resize(fyne.NewSize(1600, 1000))
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
	a.toolbar = NewModernToolbar(a.imageData, a.loader, a.logger)
	a.leftPanel = NewControlPanel(a.pipeline, a.regionManager, a.logger)
	a.imageWorkspace = NewImageWorkspace(a.imageData, a.regionManager, a.logger)
	a.rightPanel = NewInfoPanel(a.logger)
	a.statusManager = NewStatusManager()
}

func (a *Application) setupLayout() {
	// Create three-panel layout using HSplit containers

	// Left panel with fixed minimum width
	leftContent := a.leftPanel.GetContainer()

	// Center and right content
	centerContent := a.imageWorkspace.GetContainer()
	rightContent := a.rightPanel.GetContainer()

	// Create right split (center + right)
	rightSplit := container.NewHSplit(centerContent, rightContent)
	rightSplit.SetOffset(0.75) // 75% for center, 25% for right panel

	// Create main split (left + [center+right])
	mainSplit := container.NewHSplit(leftContent, rightSplit)
	mainSplit.SetOffset(0.2) // 20% for left panel, 80% for center+right

	// Main layout with toolbar and status
	a.mainContainer = container.NewBorder(
		a.toolbar.GetContainer(),    // top
		a.statusManager.GetWidget(), // bottom
		nil,                         // left
		nil,                         // right
		mainSplit,                   // center
	)

	a.window.SetContent(a.mainContainer)
}

func (a *Application) setupCallbacks() {
	// Pipeline callbacks for real-time preview
	a.pipeline.SetCallbacks(
		func(preview image.Image, metrics map[string]float64) {
			a.imageWorkspace.UpdatePreview(preview)
			a.rightPanel.UpdateMetrics(metrics)
		},
		func(err error) {
			a.statusManager.ShowError(err)
		},
	)

	// Toolbar callbacks
	a.toolbar.SetCallbacks(
		func(filepath string) {
			a.imageWorkspace.UpdateOriginal()
			a.leftPanel.Enable()
			a.statusManager.ShowSuccess(fmt.Sprintf("Loaded: %s", filepath))

			metadata := a.imageData.GetMetadata()
			a.rightPanel.ShowImageInfo(filepath, metadata.Width, metadata.Height, metadata.Channels)

			// Show original image in preview initially
			original := a.imageData.GetOriginal()
			if !original.Empty() {
				if img, err := original.ToImage(); err == nil {
					a.imageWorkspace.UpdatePreview(img)
				}
			}
			original.Close()
		},
		func(filepath string) {
			a.statusManager.ShowSuccess(fmt.Sprintf("Saved: %s", filepath))
		},
		func(tool string) {
			a.imageWorkspace.SetActiveTool(tool)
			a.statusManager.ShowInfo(fmt.Sprintf("Tool: %s", tool))
		},
		func() {
			a.pipeline.ClearAll()
			a.imageData.ResetToOriginal()
			a.imageWorkspace.Reset()
			a.leftPanel.Reset()
			a.rightPanel.Clear()
			a.statusManager.ShowInfo("Reset to original")
		},
	)

	a.toolbar.SetZoomCallback(func(zoom float64) {
		a.imageWorkspace.SetZoom(zoom)
	})

	a.toolbar.SetViewCallback(func() {
		a.statusManager.ShowInfo("View toggled")
	})

	// Left panel callbacks
	a.leftPanel.SetCallbacks(
		func(layerMode bool) {
			a.pipeline.SetProcessingMode(layerMode)
			mode := "Sequential"
			if layerMode {
				mode = "Layer"
			}
			a.statusManager.ShowInfo(fmt.Sprintf("Mode: %s", mode))
		},
		func() {
			a.imageWorkspace.RefreshSelections()
		},
	)

	// Image workspace callbacks
	a.imageWorkspace.SetCallbacks(
		func(hasSelection bool) {
			a.leftPanel.UpdateSelectionState(hasSelection)
			a.toolbar.UpdateSelectionState(hasSelection)
			if hasSelection {
				a.statusManager.ShowInfo("Region selected")
			}
		},
		func(zoom float64) {
			a.statusManager.ShowInfo(fmt.Sprintf("Zoom: %.0f%%", zoom*100))
		},
	)
}

func (a *Application) setupTheme() {
	a.app.Settings().SetTheme(&ModernTheme{})
}

func (a *Application) ShowAndRun() {
	a.logger.Info("Starting Advanced Image Processing v2.0")

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

	a.regionManager.ClearAll()
	a.pipeline.ClearAll()

	a.imageWorkspace.UpdateOriginal()
	a.leftPanel.Enable()
	a.rightPanel.Clear()

	metadata := a.imageData.GetMetadata()
	a.rightPanel.ShowImageInfo(filepath, metadata.Width, metadata.Height, metadata.Channels)
	a.statusManager.ShowSuccess(fmt.Sprintf("Image loaded: %s", filepath))

	a.logger.Info("Image loaded successfully", "filepath", filepath)
	return nil
}

func (a *Application) SaveProcessedImage(filepath string) error {
	if !a.imageData.HasImage() {
		return fmt.Errorf("no image to save")
	}

	processed, err := a.pipeline.ProcessFullResolution()
	if err != nil {
		processed = a.imageData.GetOriginal()
	}
	defer processed.Close()

	if err := a.loader.SaveImage(processed, filepath); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	a.logger.Info("Image saved successfully", "filepath", filepath)
	return nil
}

func (a *Application) RefreshUI() {
	fyne.Do(func() {
		a.imageWorkspace.RefreshSelections()
		a.leftPanel.Reset()
		a.rightPanel.Clear()
		a.statusManager.ShowInfo("UI refreshed")
	})
}

// ModernTheme implements a custom theme
type ModernTheme struct{}

func (m *ModernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 248, G: 250, B: 252, A: 255}
	case theme.ColorNameForeground:
		return color.RGBA{R: 15, G: 23, B: 42, A: 255}
	case theme.ColorNamePrimary:
		return color.RGBA{R: 59, G: 130, B: 246, A: 255}
	case theme.ColorNameFocus:
		return color.RGBA{R: 99, G: 102, B: 241, A: 255}
	case theme.ColorNameHover:
		return color.RGBA{R: 239, G: 246, B: 255, A: 255}
	case theme.ColorNameShadow:
		return color.RGBA{R: 0, G: 0, B: 0, A: 32}
	case theme.ColorNameSuccess:
		return color.RGBA{R: 34, G: 197, B: 94, A: 255}
	case theme.ColorNameWarning:
		return color.RGBA{R: 251, G: 146, B: 60, A: 255}
	case theme.ColorNameError:
		return color.RGBA{R: 239, G: 68, B: 68, A: 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255}
	case theme.ColorNameButton:
		return color.RGBA{R: 59, G: 130, B: 246, A: 255}
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 156, G: 163, B: 175, A: 255}
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 107, G: 114, B: 128, A: 255}
	case theme.ColorNamePressed:
		return color.RGBA{R: 37, G: 99, B: 235, A: 255}
	case theme.ColorNameSelection:
		return color.RGBA{R: 219, G: 234, B: 254, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m *ModernTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m *ModernTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *ModernTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameInlineIcon:
		return 20
	case theme.SizeNameScrollBar:
		return 12
	case theme.SizeNameSeparatorThickness:
		return 1
	case theme.SizeNameInputBorder:
		return 2
	case theme.SizeNameInputRadius:
		return 6
	case theme.SizeNameSelectionRadius:
		return 4
	default:
		return theme.DefaultTheme().Size(name)
	}
}
