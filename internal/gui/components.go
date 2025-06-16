// Missing GUI components for the application
package gui

import (
	"fmt"
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"
)

// ImageCanvas handles image display and ROI selection
type ImageCanvas struct {
	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *logrus.Logger

	split           *container.Split
	originalView    *widget.Card
	processedView   *widget.Card
	interactiveOrig *InteractiveCanvas
	processedImage  *canvas.Image

	activeTool         string
	onSelectionChanged func(bool)
}

// NewImageCanvas creates a new image canvas
func NewImageCanvas(imageData *core.ImageData, regionManager *core.RegionManager, logger *logrus.Logger) *ImageCanvas {
	canvas := &ImageCanvas{
		imageData:     imageData,
		regionManager: regionManager,
		logger:        logger,
		activeTool:    "none",
	}

	canvas.initializeUI()
	return canvas
}

func (ic *ImageCanvas) initializeUI() {
	// Create interactive original image view
	ic.interactiveOrig = NewInteractiveCanvas(ic.imageData, ic.regionManager, ic.logger)
	ic.originalView = widget.NewCard("Original", "", ic.interactiveOrig)

	// Create processed image view
	ic.processedImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ic.processedImage.FillMode = canvas.ImageFillContain
	ic.processedView = widget.NewCard("Processed", "", ic.processedImage)

	// Create split container
	ic.split = container.NewHSplit(ic.originalView, ic.processedView)
	ic.split.SetOffset(0.5)
}

func (ic *ImageCanvas) GetContainer() fyne.CanvasObject {
	return ic.split
}

func (ic *ImageCanvas) UpdateOriginalImage() {
	if !ic.imageData.HasImage() {
		return
	}

	original := ic.imageData.GetOriginal()
	defer original.Close()

	if !original.Empty() {
		// Convert Mat to image.Image
		img, err := original.ToImage()
		if err != nil {
			ic.logger.WithError(err).Error("Failed to convert Mat to image")
			return
		}

		// Update interactive canvas
		ic.interactiveOrig.UpdateImage(img)
	}
}

func (ic *ImageCanvas) UpdateProcessedImage(processed gocv.Mat) {
	defer processed.Close()

	if processed.Empty() {
		// Show placeholder
		ic.processedImage.Image = image.NewRGBA(image.Rect(0, 0, 1, 1))
		ic.processedImage.Refresh()
		return
	}

	// Convert Mat to image.Image
	img, err := processed.ToImage()
	if err != nil {
		ic.logger.WithError(err).Error("Failed to convert processed Mat to image")
		return
	}

	// Update processed image
	ic.processedImage.Image = img
	ic.processedImage.Refresh()
}

func (ic *ImageCanvas) SetActiveTool(tool string) {
	ic.activeTool = tool
	ic.interactiveOrig.SetActiveTool(tool)
	ic.logger.WithField("tool", tool).Debug("Active tool changed")
}

func (ic *ImageCanvas) RefreshSelections() {
	ic.interactiveOrig.RefreshSelections()
	ic.logger.Debug("Refreshing selections")
}

func (ic *ImageCanvas) SetCallbacks(onSelectionChanged func(bool)) {
	ic.onSelectionChanged = onSelectionChanged
	ic.interactiveOrig.SetSelectionChangedCallback(onSelectionChanged)
}

func (ic *ImageCanvas) Refresh() {
	ic.split.Refresh()
}

// Toolbar handles tool selection
type Toolbar struct {
	hbox          *fyne.Container
	rectangleTool *widget.Button
	freehandTool  *widget.Button
	clearButton   *widget.Button
	resetButton   *widget.Button

	onToolChanged    func(string)
	onClearSelection func()
	onResetImage     func()
}

// NewToolbar creates a new toolbar
func NewToolbar() *Toolbar {
	toolbar := &Toolbar{}
	toolbar.initializeUI()
	return toolbar
}

func (t *Toolbar) initializeUI() {
	t.rectangleTool = widget.NewButton("Rectangle", func() {
		if t.onToolChanged != nil {
			t.onToolChanged("rectangle")
		}
	})

	t.freehandTool = widget.NewButton("Freehand", func() {
		if t.onToolChanged != nil {
			t.onToolChanged("freehand")
		}
	})

	t.clearButton = widget.NewButton("Clear Selection", func() {
		if t.onClearSelection != nil {
			t.onClearSelection()
		}
	})

	t.resetButton = widget.NewButton("Reset to Original", func() {
		if t.onResetImage != nil {
			t.onResetImage()
		}
	})

	t.hbox = container.NewHBox(
		widget.NewLabel("Tools:"),
		t.rectangleTool,
		t.freehandTool,
		widget.NewSeparator(),
		t.clearButton,
		t.resetButton,
	)

	// Initially disabled
	t.Disable()
}

func (t *Toolbar) GetContainer() fyne.CanvasObject {
	return t.hbox
}

func (t *Toolbar) Enable() {
	t.rectangleTool.Enable()
	t.freehandTool.Enable()
	t.clearButton.Enable()
	t.resetButton.Enable()
}

func (t *Toolbar) Disable() {
	t.rectangleTool.Disable()
	t.freehandTool.Disable()
	t.clearButton.Disable()
	t.resetButton.Disable()
}

func (t *Toolbar) SetSelectionState(hasSelection bool) {
	if hasSelection {
		t.clearButton.Enable()
	} else {
		t.clearButton.Disable()
	}
}

func (t *Toolbar) SetCallbacks(onToolChanged func(string), onClearSelection func()) {
	t.onToolChanged = onToolChanged
	t.onClearSelection = onClearSelection
}

func (t *Toolbar) SetResetCallback(onResetImage func()) {
	t.onResetImage = onResetImage
}

func (t *Toolbar) Refresh() {
	t.hbox.Refresh()
}

// PropertiesPanel handles algorithm parameters
type PropertiesPanel struct {
	pipeline *core.ProcessingPipeline
	logger   *logrus.Logger

	vbox        *fyne.Container
	enabled     bool
	progressBar *widget.ProgressBar
	statusLabel *widget.Label
}

// NewPropertiesPanel creates a new properties panel
func NewPropertiesPanel(pipeline *core.ProcessingPipeline, logger *logrus.Logger) *PropertiesPanel {
	panel := &PropertiesPanel{
		pipeline: pipeline,
		logger:   logger,
		enabled:  false,
	}

	panel.initializeUI()
	return panel
}

func (pp *PropertiesPanel) initializeUI() {
	pp.progressBar = widget.NewProgressBar()
	pp.progressBar.Hide()

	pp.statusLabel = widget.NewLabel("")
	pp.statusLabel.Hide()

	content := container.NewVBox(
		widget.NewLabel("Select an algorithm to adjust parameters"),
		pp.progressBar,
		pp.statusLabel,
	)

	pp.vbox = container.NewVBox(
		widget.NewCard("Algorithm Properties", "", content),
	)
}

func (pp *PropertiesPanel) GetContainer() fyne.CanvasObject {
	return pp.vbox
}

func (pp *PropertiesPanel) Enable() {
	pp.enabled = true
	// TODO: Enable algorithm selection UI
}

func (pp *PropertiesPanel) Disable() {
	pp.enabled = false
}

func (pp *PropertiesPanel) UpdateProgress(step, total int, stepName string) {
	pp.progressBar.Show()
	pp.statusLabel.Show()

	if total > 0 {
		pp.progressBar.SetValue(float64(step) / float64(total))
	}
	pp.statusLabel.SetText(fmt.Sprintf("Step %d/%d: %s", step, total, stepName))

	pp.logger.WithFields(logrus.Fields{
		"step":      step,
		"total":     total,
		"step_name": stepName,
	}).Debug("Processing progress")
}

func (pp *PropertiesPanel) ClearProgress() {
	pp.progressBar.Hide()
	pp.statusLabel.Hide()
}

func (pp *PropertiesPanel) Refresh() {
	pp.vbox.Refresh()
}

// MetricsPanel displays quality metrics
type MetricsPanel struct {
	vbox    *fyne.Container
	metrics map[string]float64
}

// NewMetricsPanel creates a new metrics panel
func NewMetricsPanel() *MetricsPanel {
	panel := &MetricsPanel{
		metrics: make(map[string]float64),
	}

	panel.initializeUI()
	return panel
}

func (mp *MetricsPanel) initializeUI() {
	mp.vbox = container.NewVBox(
		widget.NewCard("Quality Metrics", "",
			widget.NewLabel("Process an image to see quality metrics")),
	)
}

func (mp *MetricsPanel) GetContainer() fyne.CanvasObject {
	return mp.vbox
}

func (mp *MetricsPanel) UpdateMetrics(metrics map[string]float64) {
	mp.metrics = metrics

	// Create metrics display
	content := container.NewVBox()

	for name, value := range metrics {
		label := widget.NewLabel(fmt.Sprintf("%s: %.3f", name, value))
		content.Add(label)
	}

	if len(metrics) == 0 {
		content.Add(widget.NewLabel("No metrics available"))
	}

	card := widget.NewCard("Quality Metrics", "", content)
	mp.vbox.RemoveAll()
	mp.vbox.Add(card)
}

func (mp *MetricsPanel) Clear() {
	mp.metrics = make(map[string]float64)
	mp.vbox.RemoveAll()
	mp.vbox.Add(widget.NewCard("Quality Metrics", "",
		widget.NewLabel("Process an image to see quality metrics")))
}

func (mp *MetricsPanel) Refresh() {
	mp.vbox.Refresh()
}

// MenuHandler handles menu actions
type MenuHandler struct {
	window    fyne.Window
	imageData *core.ImageData
	loader    *io.ImageLoader
	logger    *logrus.Logger

	onImageLoaded func(string)
	onImageSaved  func(string)
}

// NewMenuHandler creates a new menu handler
func NewMenuHandler(window fyne.Window, imageData *core.ImageData, loader *io.ImageLoader, logger *logrus.Logger) *MenuHandler {
	return &MenuHandler{
		window:    window,
		imageData: imageData,
		loader:    loader,
		logger:    logger,
	}
}

func (mh *MenuHandler) GetMainMenu() *fyne.MainMenu {
	// File menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Image...", mh.openImage),
		fyne.NewMenuItem("Save Image...", mh.saveImage),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() {
			mh.window.Close()
		}),
	)

	// Edit menu
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Clear Selection", func() {
			// TODO: Clear selection
		}),
		fyne.NewMenuItem("Reset to Original", func() {
			if mh.imageData.HasImage() {
				mh.imageData.ResetToOriginal()
				mh.logger.Info("Reset to original image")
				// Trigger UI update
				if mh.onImageLoaded != nil {
					filepath := mh.imageData.GetFilepath()
					mh.onImageLoaded(filepath)
				}
			}
		}),
	)

	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", mh.showAbout),
	)

	return fyne.NewMainMenu(fileMenu, editMenu, helpMenu)
}

func (mh *MenuHandler) openImage() {
	mh.logger.Info("Opening file dialog for image selection")

	// Create file dialog for opening images
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			mh.showError("File Dialog Error", err)
			return
		}
		if reader == nil {
			return // User cancelled
		}
		defer reader.Close()

		uri := reader.URI()
		filepath := uri.Path()

		mh.logger.WithField("filepath", filepath).Info("Loading selected image")

		// Load the image
		mat, err := mh.loader.LoadImage(filepath)
		if err != nil {
			mh.showError("Failed to Load Image", err)
			return
		}
		defer mat.Close()

		// Validate the image
		if err := core.ValidateImage(mat); err != nil {
			mh.showError("Invalid Image", err)
			return
		}

		// Set the image
		if err := mh.imageData.SetOriginal(mat, filepath); err != nil {
			mh.showError("Failed to Set Image", err)
			return
		}

		mh.logger.WithField("filepath", filepath).Info("Image loaded successfully")

		if mh.onImageLoaded != nil {
			mh.onImageLoaded(filepath)
		}

	}, mh.window)

	// Set file filter for images
	imageFilter := storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"})
	fileDialog.SetFilter(imageFilter)

	fileDialog.Show()
}

func (mh *MenuHandler) saveImage() {
	if !mh.imageData.HasImage() {
		mh.showError("No Image", fmt.Errorf("no image loaded to save"))
		return
	}

	mh.logger.Info("Opening file dialog for image saving")

	// Create file dialog for saving images
	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			mh.showError("File Dialog Error", err)
			return
		}
		if writer == nil {
			return // User cancelled
		}
		defer writer.Close()

		uri := writer.URI()
		filepath := uri.Path()

		mh.logger.WithField("filepath", filepath).Info("Saving image")

		// Get processed image (or original if no processing)
		processed := mh.imageData.GetProcessed()
		defer processed.Close()

		if processed.Empty() {
			// Use original if no processed image
			processed = mh.imageData.GetOriginal()
		}

		// Save the image
		if err := mh.loader.SaveImage(processed, filepath); err != nil {
			mh.showError("Failed to Save Image", err)
			return
		}

		mh.logger.WithField("filepath", filepath).Info("Image saved successfully")

		if mh.onImageSaved != nil {
			mh.onImageSaved(filepath)
		}

	}, mh.window)

	// Set default filename and filter
	fileDialog.SetFileName("processed_image.png")
	imageFilter := storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".tiff", ".tif", ".bmp"})
	fileDialog.SetFilter(imageFilter)

	fileDialog.Show()
}

func (mh *MenuHandler) showAbout() {
	content := container.NewVBox(
		widget.NewLabel("Advanced Image Processing v2.0"),
		widget.NewSeparator(),
		widget.NewLabel("A professional-grade image processing application"),
		widget.NewLabel("for historical documents with ROI selection,"),
		widget.NewLabel("Local Adaptive Algorithms (LAA), and"),
		widget.NewLabel("comprehensive quality metrics."),
		widget.NewSeparator(),
		widget.NewLabel("Author: Ervins Strauhmanis"),
		widget.NewLabel("Built with Go, Fyne v2.6, and OpenCV 4.11"),
		widget.NewSeparator(),
		widget.NewLabel("License: MIT"),
	)

	aboutDialog := dialog.NewCustom("About", "Close", content, mh.window)
	aboutDialog.Resize(fyne.NewSize(400, 300))
	aboutDialog.Show()
}

func (mh *MenuHandler) showError(title string, err error) {
	mh.logger.WithError(err).Error(title)
	dialog.ShowError(err, mh.window)
}

func (mh *MenuHandler) SetCallbacks(onImageLoaded, onImageSaved func(string)) {
	mh.onImageLoaded = onImageLoaded
	mh.onImageSaved = onImageSaved
}
