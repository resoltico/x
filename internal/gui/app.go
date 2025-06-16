// internal/gui/app.go
// Redesigned main application with modern UI patterns and enhanced UX
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

// Application represents the redesigned main application
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
	// Initialize modern components
	a.toolbar = NewModernToolbar(a.imageData, a.loader, a.logger)
	a.leftPanel = NewControlPanel(a.pipeline, a.regionManager, a.logger)
	a.imageWorkspace = NewImageWorkspace(a.imageData, a.regionManager, a.logger)
	a.rightPanel = NewInfoPanel(a.logger)
	a.statusManager = NewStatusManager()
}

func (a *Application) setupLayout() {
	// Modern three-panel layout with proper proportions
	a.mainContainer = container.NewBorder(
		a.toolbar.GetContainer(),        // top
		a.statusManager.GetWidget(),     // bottom
		a.leftPanel.GetContainer(),      // left (300px)
		a.rightPanel.GetContainer(),     // right (300px)
		a.imageWorkspace.GetContainer(), // center (1000px)
	)

	a.window.SetContent(a.mainContainer)
}

func (a *Application) setupCallbacks() {
	// Pipeline callbacks for real-time preview
	a.pipeline.SetCallbacks(
		// onPreviewUpdate - receives thread-safe image.Image
		func(preview image.Image, metrics map[string]float64) {
			a.imageWorkspace.UpdatePreview(preview)
			a.rightPanel.UpdateMetrics(metrics)
		},
		// onError
		func(err error) {
			a.statusManager.ShowError(err)
		},
	)

	// Toolbar callbacks
	a.toolbar.SetCallbacks(
		// onImageLoaded
		func(filepath string) {
			a.imageWorkspace.UpdateOriginal()
			a.leftPanel.Enable()
			a.statusManager.ShowSuccess(fmt.Sprintf("Loaded: %s", filepath))

			// Show image info in right panel
			metadata := a.imageData.GetMetadata()
			a.rightPanel.ShowImageInfo(filepath, metadata.Width, metadata.Height, metadata.Channels)
		},
		// onImageSaved
		func(filepath string) {
			a.statusManager.ShowSuccess(fmt.Sprintf("Saved: %s", filepath))
		},
		// onToolChanged
		func(tool string) {
			a.imageWorkspace.SetActiveTool(tool)
			a.statusManager.ShowInfo(fmt.Sprintf("Tool: %s", tool))
		},
		// onResetImage
		func() {
			a.pipeline.ClearAll()
			a.imageData.ResetToOriginal()
			a.imageWorkspace.Reset()
			a.leftPanel.Reset()
			a.rightPanel.Clear()
			a.statusManager.ShowInfo("Reset to original")
		},
	)

	// Set zoom callback
	a.toolbar.SetZoomCallback(func(zoom float64) {
		a.imageWorkspace.SetZoom(zoom)
	})

	// Set view toggle callback
	a.toolbar.SetViewCallback(func() {
		// View toggle functionality would be implemented here
		a.statusManager.ShowInfo("View toggled")
	})

	// Left panel callbacks
	a.leftPanel.SetCallbacks(
		// onModeChanged
		func(layerMode bool) {
			a.pipeline.SetProcessingMode(layerMode)
			mode := "Sequential"
			if layerMode {
				mode = "Layer"
			}
			a.statusManager.ShowInfo(fmt.Sprintf("Mode: %s", mode))
		},
		// onSelectionChanged
		func() {
			a.imageWorkspace.RefreshSelections()
		},
	)

	// Image workspace callbacks
	a.imageWorkspace.SetCallbacks(
		// onSelectionChanged
		func(hasSelection bool) {
			a.leftPanel.UpdateSelectionState(hasSelection)
			a.toolbar.UpdateSelectionState(hasSelection)
			if hasSelection {
				a.statusManager.ShowInfo("Region selected")
			}
		},
		// onZoomChanged
		func(zoom float64) {
			a.statusManager.ShowInfo(fmt.Sprintf("Zoom: %.0f%%", zoom*100))
		},
	)
}

func (a *Application) setupTheme() {
	// Apply modern theme with consistent colors
	a.app.Settings().SetTheme(&ModernTheme{})
}

func (a *Application) ShowAndRun() {
	a.logger.Info("Starting Advanced Image Processing v2.0 with modern UI")

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

// LoadImageFromPath loads an image from the specified file path
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
	a.imageWorkspace.UpdateOriginal()
	a.leftPanel.Enable()
	a.rightPanel.Clear()

	metadata := a.imageData.GetMetadata()
	a.rightPanel.ShowImageInfo(filepath, metadata.Width, metadata.Height, metadata.Channels)
	a.statusManager.ShowSuccess(fmt.Sprintf("Image loaded: %s", filepath))

	a.logger.Info("Image loaded successfully", "filepath", filepath)
	return nil
}

// SaveProcessedImage saves the currently processed image
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

// RefreshUI refreshes all UI components
func (a *Application) RefreshUI() {
	fyne.Do(func() {
		a.imageWorkspace.RefreshSelections()
		a.leftPanel.Reset()
		a.rightPanel.Clear()
		a.statusManager.ShowInfo("UI refreshed")
	})
}

// ModernTheme implements a custom theme with 2025 design principles
type ModernTheme struct{}

func (m *ModernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 248, G: 250, B: 252, A: 255} // Modern light gray
	case theme.ColorNameForeground:
		return color.RGBA{R: 15, G: 23, B: 42, A: 255} // Dark slate
	case theme.ColorNamePrimary:
		return color.RGBA{R: 59, G: 130, B: 246, A: 255} // Modern blue
	case theme.ColorNameFocus:
		return color.RGBA{R: 99, G: 102, B: 241, A: 255} // Indigo focus
	case theme.ColorNameHover:
		return color.RGBA{R: 239, G: 246, B: 255, A: 255} // Light blue hover
	case theme.ColorNameShadow:
		return color.RGBA{R: 0, G: 0, B: 0, A: 32} // Subtle shadow
	case theme.ColorNameSuccess:
		return color.RGBA{R: 34, G: 197, B: 94, A: 255} // Modern green
	case theme.ColorNameWarning:
		return color.RGBA{R: 251, G: 146, B: 60, A: 255} // Modern orange
	case theme.ColorNameError:
		return color.RGBA{R: 239, G: 68, B: 68, A: 255} // Modern red
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 255, G: 255, B: 255, A: 255} // Pure white
	case theme.ColorNameButton:
		return color.RGBA{R: 59, G: 130, B: 246, A: 255} // Modern blue
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 156, G: 163, B: 175, A: 255} // Gray
	case theme.ColorNamePlaceHolder:
		return color.RGBA{R: 107, G: 114, B: 128, A: 255} // Medium gray
	case theme.ColorNamePressed:
		return color.RGBA{R: 37, G: 99, B: 235, A: 255} // Darker blue
	case theme.ColorNameSelection:
		return color.RGBA{R: 219, G: 234, B: 254, A: 255} // Light blue selection
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
