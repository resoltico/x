// Author: Ervins Strauhmanis
// License: MIT

package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/image_processing"
	"advanced-image-processing/internal/models"
	"advanced-image-processing/internal/presets"
	"advanced-image-processing/internal/transforms"
	"advanced-image-processing/internal/transforms/binarization"
	"advanced-image-processing/internal/transforms/morphology"
	"advanced-image-processing/internal/transforms/noise_reduction"
	"advanced-image-processing/internal/transforms/color_manipulation"
)

// MainWindow represents the main application window
type MainWindow struct {
	app          fyne.App
	window       fyne.Window
	logger       *logrus.Logger
	debugMode    bool
	
	// Core components
	imageData    *models.ImageData
	pipeline     *image_processing.Pipeline
	registry     *transforms.TransformRegistry
	loader       *image_processing.ImageLoader
	presetMgr    *presets.Manager
	
	// GUI components
	preview      *Preview
	controls     *Controls
	menu         *Menu
	errorReport  *ErrorReport
	
	// Layout containers
	sidebar      *fyne.Container
	imageArea    *fyne.Container
	paramPanel   *fyne.Container
	mainContent  *fyne.Container
}

// NewMainWindow creates and initializes the main window
func NewMainWindow(app fyne.App, logger *logrus.Logger, debugMode bool) *MainWindow {
	window := app.NewWindow("Advanced Image Processing")
	window.Resize(fyne.NewSize(1200, 800))
	window.CenterOnScreen()

	mw := &MainWindow{
		app:       app,
		window:    window,
		logger:    logger,
		debugMode: debugMode,
	}

	mw.initializeComponents()
	mw.setupLayout()
	mw.setupCallbacks()

	return mw
}

// initializeComponents initializes all core components
func (mw *MainWindow) initializeComponents() {
	// Initialize data models
	mw.imageData = models.NewImageData()
	
	// Initialize transform registry and register all transforms
	mw.registry = transforms.NewTransformRegistry()
	mw.registerTransforms()
	
	// Initialize core components
	mw.loader = image_processing.NewImageLoader(mw.logger)
	mw.pipeline = image_processing.NewPipeline(mw.registry, mw.imageData, mw.logger)
	mw.presetMgr = presets.NewManager(mw.logger)
	
	// Initialize GUI components
	mw.preview = NewPreview(mw.imageData, mw.logger)
	mw.controls = NewControls(mw.registry, mw.pipeline, mw.logger)
	mw.menu = NewMenu(mw.loader, mw.imageData, mw.pipeline, mw.presetMgr, mw.logger)
	mw.errorReport = NewErrorReport(mw.logger, mw.debugMode)
}

// registerTransforms registers all available transformations
func (mw *MainWindow) registerTransforms() {
	// Binarization transforms
	mw.registry.Register("otsu", binarization.NewOtsuTransform())
	mw.registry.Register("niblack", binarization.NewNiblackTransform())
	mw.registry.Register("sauvola", binarization.NewSauvolaTransform())
	
	// Morphology transforms
	mw.registry.Register("erosion", morphology.NewErosionTransform())
	mw.registry.Register("dilation", morphology.NewDilationTransform())
	
	// Noise reduction transforms
	mw.registry.Register("gaussian", noise_reduction.NewGaussianTransform())
	
	// Color manipulation transforms
	mw.registry.Register("grayscale", color_manipulation.NewGrayscaleTransform())
}

// setupLayout creates and organizes the GUI layout
func (mw *MainWindow) setupLayout() {
	// Create sidebar with transformation categories
	mw.sidebar = mw.createSidebar()
	
	// Create image preview area
	mw.imageArea = container.NewBorder(
		nil, // top
		nil, // bottom
		nil, // left
		nil, // right
		mw.preview.GetContainer(),
	)
	
	// Create parameter panel
	mw.paramPanel = mw.controls.GetContainer()
	
	// Create main content area with split layout
	imageAndParams := container.NewHSplit(
		mw.imageArea,
		mw.paramPanel,
	)
	imageAndParams.SetOffset(0.7) // 70% for image, 30% for parameters
	
	// Create main layout with sidebar
	mw.mainContent = container.NewBorder(
		nil, // top
		nil, // bottom
		mw.sidebar, // left
		nil, // right
		imageAndParams,
	)
	
	// Set window content
	mw.window.SetContent(mw.mainContent)
	
	// Set menu bar
	mw.window.SetMainMenu(mw.menu.GetMainMenu())
}

// createSidebar creates the transformation category sidebar
func (mw *MainWindow) createSidebar() *fyne.Container {
	title := widget.NewRichTextFromMarkdown("## Transformations")
	
	// Create category accordions
	categories := mw.registry.GetByCategory()
	accordions := make([]fyne.CanvasObject, 0, len(categories))
	
	for categoryName, transformNames := range categories {
		items := make([]fyne.CanvasObject, 0, len(transformNames))
		
		for _, transformName := range transformNames {
			transform, exists := mw.registry.Get(transformName)
			if !exists {
				continue
			}
			
			// Create button for this transformation
			btn := widget.NewButton(transform.GetName(), func(name string) func() {
				return func() {
					mw.addTransformation(name)
				}
			}(transformName))
			
			btn.Resize(fyne.NewSize(200, 35))
			items = append(items, btn)
		}
		
		// Create accordion item
		accordion := widget.NewAccordion()
		accordion.Append(categoryName, container.NewVBox(items...))
		accordions = append(accordions, accordion)
	}
	
	// Create scrollable sidebar
	sidebarContent := container.NewVBox(append([]fyne.CanvasObject{title}, accordions...)...)
	scroll := container.NewScroll(sidebarContent)
	scroll.Resize(fyne.NewSize(250, 600))
	
	return container.NewBorder(
		nil, // top
		nil, // bottom
		nil, // left
		nil, // right
		scroll,
	)
}

// setupCallbacks sets up event callbacks between components
func (mw *MainWindow) setupCallbacks() {
	// Set pipeline callbacks
	mw.pipeline.SetCallbacks(
		// onProgress
		func(step, total int, stepName string) {
			mw.logger.WithFields(logrus.Fields{
				"step":      step,
				"total":     total,
				"transform": stepName,
			}).Debug("Pipeline progress")
		},
		// onComplete
		func(result gocv.Mat) {
			mw.preview.UpdateProcessed(result)
			result.Close()
		},
		// onError
		func(err error) {
			mw.errorReport.ShowError(err)
		},
	)
	
	// Set menu callbacks
	mw.menu.SetCallbacks(
		// onImageLoaded
		func() {
			mw.preview.UpdateOriginal()
			mw.controls.Enable()
		},
		// onPresetLoaded
		func() {
			mw.controls.RefreshSequence()
		},
	)
	
	// Set controls callbacks
	mw.controls.SetCallbacks(
		// onTransformationChanged
		func() {
			// Pipeline will automatically reprocess
		},
	)
}

// addTransformation adds a transformation to the pipeline
func (mw *MainWindow) addTransformation(transformName string) {
	transform, exists := mw.registry.Get(transformName)
	if !exists {
		mw.errorReport.ShowError(fmt.Errorf("unknown transformation: %s", transformName))
		return
	}
	
	// Add with default parameters
	params := transform.GetDefaultParams()
	if err := mw.pipeline.AddTransformation(transformName, params); err != nil {
		mw.errorReport.ShowError(err)
		return
	}
	
	// Update controls to show the new transformation
	mw.controls.RefreshSequence()
	
	mw.logger.WithField("transform", transformName).Info("Added transformation")
}

// ShowAndRun displays the window and starts the application
func (mw *MainWindow) ShowAndRun() {
	mw.window.ShowAndRun()
}

// GetWindow returns the underlying Fyne window
func (mw *MainWindow) GetWindow() fyne.Window {
	return mw.window
}