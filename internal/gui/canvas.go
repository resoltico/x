// Real-time image canvas with preview support
package gui

import (
	"fmt"
	"image"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/core"
)

// ImageCanvas handles image display with real-time preview
type ImageCanvas struct {
	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *slog.Logger

	split           *container.Split
	originalView    *widget.Card
	previewView     *widget.Card
	interactiveOrig *InteractiveCanvas
	previewImage    *canvas.Image

	activeTool         string
	onSelectionChanged func(bool)
}

func NewImageCanvas(imageData *core.ImageData, regionManager *core.RegionManager, logger *slog.Logger) *ImageCanvas {
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

	// Create real-time preview image view
	ic.previewImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ic.previewImage.FillMode = canvas.ImageFillContain
	ic.previewView = widget.NewCard("Preview", "", ic.previewImage)

	// Create split container
	ic.split = container.NewHSplit(ic.originalView, ic.previewView)
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
		img, err := original.ToImage()
		if err != nil {
			ic.logger.Error("Failed to convert Mat to image", "error", err)
			return
		}

		ic.interactiveOrig.UpdateImage(img)
	}
}

func (ic *ImageCanvas) UpdatePreview(preview gocv.Mat) {
	defer preview.Close()

	if preview.Empty() {
		ic.previewImage.Image = image.NewRGBA(image.Rect(0, 0, 1, 1))
		ic.previewImage.Refresh()
		return
	}

	img, err := preview.ToImage()
	if err != nil {
		ic.logger.Error("Failed to convert preview Mat to image", "error", err)
		return
	}

	ic.previewImage.Image = img
	ic.previewImage.Refresh()
}

func (ic *ImageCanvas) ClearPreview() {
	ic.previewImage.Image = image.NewRGBA(image.Rect(0, 0, 1, 1))
	ic.previewImage.Refresh()
}

func (ic *ImageCanvas) SetActiveTool(tool string) {
	ic.activeTool = tool
	ic.interactiveOrig.SetActiveTool(tool)
	ic.logger.Debug("Active tool changed", "tool", tool)
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

// MetricsPanel displays quality metrics
type MetricsPanel struct {
	vbox    *fyne.Container
	metrics map[string]float64
}

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
			widget.NewLabel("Real-time metrics will appear here")),
	)
}

func (mp *MetricsPanel) GetContainer() fyne.CanvasObject {
	return mp.vbox
}

func (mp *MetricsPanel) UpdateMetrics(metrics map[string]float64) {
	mp.metrics = metrics

	content := container.NewVBox()

	for name, value := range metrics {
		var displayText string
		switch name {
		case "psnr":
			displayText = fmt.Sprintf("PSNR: %.2f dB", value)
		case "ssim":
			displayText = fmt.Sprintf("SSIM: %.3f", value)
		case "mse":
			displayText = fmt.Sprintf("MSE: %.2f", value)
		default:
			displayText = fmt.Sprintf("%s: %.3f", name, value)
		}

		label := widget.NewLabel(displayText)
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
		widget.NewLabel("Real-time metrics will appear here")))
}

func (mp *MetricsPanel) Refresh() {
	mp.vbox.Refresh()
}
