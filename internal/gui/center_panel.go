// internal/gui/center_panel.go
// Perfect UI Center Panel: Image Workspace (1000px wide)
package gui

import (
	"fmt"
	"image"
	"image/color"
	"log/slog"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/core"
)

type CenterPanel struct {
	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *slog.Logger

	container *fyne.Container

	// Toolbar (40px height)
	toolbar *fyne.Container

	// Selection tools (Layer Mode only)
	rectTool  *widget.Button
	lassoTool *widget.Button
	wandTool  *widget.Button

	// Zoom controls
	zoomSlider *widget.Slider
	zoomLabel  *widget.Label
	fitBtn     *widget.Button

	// View toggle
	viewToggle *widget.Button

	// Image Display (910px height)
	imageDisplay *fyne.Container
	splitView    *container.Split
	singleView   *fyne.Container

	// Interactive canvases
	originalCanvas *InteractiveCanvas
	previewCanvas  *StaticCanvas

	// State
	currentZoom float64
	activeTool  string
	isSplitView bool
	isLayerMode bool

	// Callbacks
	onToolChanged      func(string)
	onZoomChanged      func(float64)
	onSelectionChanged func(bool)
}

func NewCenterPanel(imageData *core.ImageData, regionManager *core.RegionManager, logger *slog.Logger) *CenterPanel {
	panel := &CenterPanel{
		imageData:     imageData,
		regionManager: regionManager,
		logger:        logger,
		currentZoom:   1.0,
		activeTool:    "none",
		isSplitView:   true,
		isLayerMode:   false,
	}

	panel.initializeUI()
	return panel
}

func (cp *CenterPanel) initializeUI() {
	// Create interactive canvases
	cp.originalCanvas = NewInteractiveCanvas(cp.imageData, cp.regionManager, cp.logger)
	cp.previewCanvas = NewStaticCanvas(cp.logger)

	// Set canvas callbacks
	cp.originalCanvas.SetSelectionCallback(func(hasSelection bool) {
		if cp.onSelectionChanged != nil {
			cp.onSelectionChanged(hasSelection)
		}
	})

	// Initialize toolbar (40px height)
	cp.initializeToolbar()

	// Initialize image display (910px height)
	cp.initializeImageDisplay()

	// Main container: Toolbar (top) + Image Display (center)
	cp.container = container.NewBorder(
		cp.toolbar,      // top
		nil,             // bottom
		nil,             // left
		nil,             // right
		cp.imageDisplay, // center
	)

	// Set fixed width to 1000px as per specification
	cp.container.Resize(fyne.NewSize(1000, 950)) // 40px toolbar + 910px display
}

func (cp *CenterPanel) initializeToolbar() {
	// Selection Tools (disabled in Sequential Mode)
	cp.rectTool = widget.NewButtonWithIcon("", theme.ContentAddIcon(), func() {
		cp.setActiveTool("rectangle")
	})
	cp.rectTool.Resize(fyne.NewSize(20, 20))
	cp.rectTool.Disable()

	cp.lassoTool = widget.NewButtonWithIcon("", theme.ContentCutIcon(), func() {
		cp.setActiveTool("lasso")
	})
	cp.lassoTool.Resize(fyne.NewSize(20, 20))
	cp.lassoTool.Disable()

	// Magic wand tool disabled - FloodFill not available in GoCV yet
	cp.wandTool = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		// TODO: Implement when gocv.FloodFill becomes available
	})
	cp.wandTool.Resize(fyne.NewSize(20, 20))
	cp.wandTool.Disable() // Keep disabled until implemented

	selectionTools := container.NewHBox(
		widget.NewLabel("Selection:"),
		cp.rectTool,
		cp.lassoTool,
		// cp.wandTool, // Comment out until implemented
	)

	// Zoom Controls
	cp.zoomSlider = widget.NewSlider(0.25, 4.0)
	cp.zoomSlider.SetValue(1.0)
	cp.zoomSlider.Step = 0.25
	cp.zoomSlider.Resize(fyne.NewSize(80, 25))
	cp.zoomSlider.OnChanged = func(value float64) {
		cp.currentZoom = value
		cp.zoomLabel.SetText(fmt.Sprintf("%.0f%%", value*100))
		cp.originalCanvas.SetZoom(value)
		cp.previewCanvas.SetZoom(value)
		if cp.onZoomChanged != nil {
			cp.onZoomChanged(value)
		}
	}

	zoomInBtn := widget.NewButtonWithIcon("", theme.ZoomInIcon(), func() {
		newZoom := cp.currentZoom + 0.25
		if newZoom <= 4.0 {
			cp.zoomSlider.SetValue(newZoom)
		}
	})
	zoomInBtn.Resize(fyne.NewSize(15, 15))

	zoomOutBtn := widget.NewButtonWithIcon("", theme.ZoomOutIcon(), func() {
		newZoom := cp.currentZoom - 0.25
		if newZoom >= 0.25 {
			cp.zoomSlider.SetValue(newZoom)
		}
	})
	zoomOutBtn.Resize(fyne.NewSize(15, 15))

	cp.zoomLabel = widget.NewLabel("100%")
	cp.zoomLabel.Resize(fyne.NewSize(50, 25))

	cp.fitBtn = widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), func() {
		cp.zoomSlider.SetValue(1.0)
	})
	cp.fitBtn.Resize(fyne.NewSize(20, 20))

	zoomControls := container.NewHBox(
		widget.NewLabel("Zoom:"),
		zoomOutBtn,
		cp.zoomSlider,
		zoomInBtn,
		cp.zoomLabel,
		cp.fitBtn,
	)

	// View Toggle
	cp.viewToggle = widget.NewButtonWithIcon("", theme.ViewFullScreenIcon(), func() {
		cp.isSplitView = !cp.isSplitView
		cp.updateImageDisplay()
	})
	cp.viewToggle.Resize(fyne.NewSize(20, 20))

	viewControls := container.NewHBox(
		widget.NewLabel("View:"),
		cp.viewToggle,
	)

	// Toolbar layout with 5px padding
	cp.toolbar = container.NewHBox(
		selectionTools,
		widget.NewSeparator(),
		zoomControls,
		widget.NewSeparator(),
		viewControls,
	)
}

func (cp *CenterPanel) initializeImageDisplay() {
	// Create image containers with cards
	originalCard := widget.NewCard("Original", "", cp.originalCanvas)
	previewCard := widget.NewCard("Preview", "", cp.previewCanvas)

	// Split View: Two 455px canvases with 1px divider
	cp.splitView = container.NewVSplit(originalCard, previewCard)
	cp.splitView.SetOffset(0.5) // 50/50 split

	// Single View: 910px canvas (preview only)
	cp.singleView = container.NewBorder(nil, nil, nil, nil, previewCard)

	// Start with Split View
	cp.imageDisplay = container.NewBorder(nil, nil, nil, nil, cp.splitView)
}

func (cp *CenterPanel) updateImageDisplay() {
	cp.imageDisplay.RemoveAll()

	if cp.isSplitView {
		cp.imageDisplay.Add(cp.splitView)
	} else {
		cp.imageDisplay.Add(cp.singleView)
	}

	cp.imageDisplay.Refresh()
}

func (cp *CenterPanel) setActiveTool(tool string) {
	// Reset tool button importance
	cp.rectTool.Importance = widget.MediumImportance
	cp.lassoTool.Importance = widget.MediumImportance
	// cp.wandTool.Importance = widget.MediumImportance // Disabled until implemented

	// Highlight active tool
	switch tool {
	case "rectangle":
		cp.rectTool.Importance = widget.HighImportance
	case "lasso":
		cp.lassoTool.Importance = widget.HighImportance
		// case "wand": // Disabled until implemented
		//	cp.wandTool.Importance = widget.HighImportance
	}

	cp.activeTool = tool
	cp.originalCanvas.SetActiveTool(tool)

	// Refresh buttons
	cp.rectTool.Refresh()
	cp.lassoTool.Refresh()
	// cp.wandTool.Refresh() // Disabled until implemented

	if cp.onToolChanged != nil {
		cp.onToolChanged(tool)
	}
}

func (cp *CenterPanel) SetProcessingMode(layerMode bool) {
	cp.isLayerMode = layerMode

	if layerMode {
		// Enable selection tools in Layer Mode (except wand - not implemented yet)
		cp.rectTool.Enable()
		cp.lassoTool.Enable()
		// cp.wandTool.Enable() // Keep disabled until FloodFill is available
	} else {
		// Disable selection tools in Sequential Mode
		cp.rectTool.Disable()
		cp.lassoTool.Disable()
		// cp.wandTool.Disable() // Already disabled
		cp.setActiveTool("none")
	}
}

func (cp *CenterPanel) UpdateOriginal() {
	if !cp.imageData.HasImage() {
		return
	}

	original := cp.imageData.GetOriginal()
	defer original.Close()

	if !original.Empty() {
		if img, err := original.ToImage(); err == nil {
			cp.originalCanvas.UpdateImage(img)
		}
	}
}

func (cp *CenterPanel) UpdatePreview(preview image.Image) {
	cp.previewCanvas.UpdateImage(preview)
}

func (cp *CenterPanel) Reset() {
	cp.originalCanvas.Clear()
	cp.previewCanvas.Clear()
	cp.setActiveTool("none")
	cp.zoomSlider.SetValue(1.0)
}

func (cp *CenterPanel) GetContainer() fyne.CanvasObject {
	return cp.container
}

func (cp *CenterPanel) SetCallbacks(
	onToolChanged func(string),
	onZoomChanged func(float64),
	onSelectionChanged func(bool),
) {
	cp.onToolChanged = onToolChanged
	cp.onZoomChanged = onZoomChanged
	cp.onSelectionChanged = onSelectionChanged
}

// InteractiveCanvas handles user interaction for ROI selection
type InteractiveCanvas struct {
	widget.BaseWidget

	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *slog.Logger

	currentImage *canvas.Image
	overlay      *canvas.Raster

	activeTool    string
	currentZoom   float64
	isDrawing     bool
	startPoint    fyne.Position
	currentPoints []image.Point
	mousePos      fyne.Position

	onSelectionChanged func(bool)
}

func NewInteractiveCanvas(imageData *core.ImageData, regionManager *core.RegionManager, logger *slog.Logger) *InteractiveCanvas {
	canvas := &InteractiveCanvas{
		imageData:     imageData,
		regionManager: regionManager,
		logger:        logger,
		activeTool:    "none",
		currentZoom:   1.0,
		currentPoints: make([]image.Point, 0),
	}

	canvas.ExtendBaseWidget(canvas)
	return canvas
}

func (ic *InteractiveCanvas) CreateRenderer() fyne.WidgetRenderer {
	// Create placeholder image
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}

	ic.currentImage = canvas.NewImageFromImage(placeholder)
	ic.currentImage.FillMode = canvas.ImageFillContain
	ic.currentImage.ScaleMode = canvas.ImageScalePixels

	ic.overlay = canvas.NewRaster(ic.createOverlay)

	return &interactiveCanvasRenderer{
		canvas:  ic,
		image:   ic.currentImage,
		overlay: ic.overlay,
	}
}

func (ic *InteractiveCanvas) UpdateImage(img image.Image) {
	ic.currentImage.Image = img
	ic.currentImage.Refresh()
	ic.overlay.Refresh()
}

func (ic *InteractiveCanvas) SetActiveTool(tool string) {
	ic.activeTool = tool
	ic.overlay.Refresh()
}

func (ic *InteractiveCanvas) SetZoom(zoom float64) {
	ic.currentZoom = zoom
	ic.currentImage.Refresh()
	ic.overlay.Refresh()
}

func (ic *InteractiveCanvas) Clear() {
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}
	ic.UpdateImage(placeholder)
}

func (ic *InteractiveCanvas) SetSelectionCallback(callback func(bool)) {
	ic.onSelectionChanged = callback
}

// Mouse event handlers
func (ic *InteractiveCanvas) MouseDown(event *desktop.MouseEvent) {
	if ic.activeTool == "none" || !ic.imageData.HasImage() {
		return
	}

	ic.isDrawing = true
	ic.startPoint = event.Position
	ic.mousePos = event.Position

	imagePoint := ic.screenToImageCoords(event.Position)

	switch ic.activeTool {
	case "rectangle":
		ic.currentPoints = []image.Point{imagePoint}
	case "lasso":
		ic.currentPoints = []image.Point{imagePoint}
	}
}

func (ic *InteractiveCanvas) MouseUp(event *desktop.MouseEvent) {
	if !ic.isDrawing {
		return
	}

	ic.isDrawing = false
	imagePoint := ic.screenToImageCoords(event.Position)

	switch ic.activeTool {
	case "rectangle":
		if len(ic.currentPoints) > 0 {
			startPoint := ic.currentPoints[0]
			rect := image.Rect(
				int(math.Min(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Min(float64(startPoint.Y), float64(imagePoint.Y))),
				int(math.Max(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Max(float64(startPoint.Y), float64(imagePoint.Y))),
			)
			if !rect.Empty() {
				ic.regionManager.CreateRectangleSelection(rect)
				if ic.onSelectionChanged != nil {
					ic.onSelectionChanged(true)
				}
			}
		}
	case "lasso":
		if len(ic.currentPoints) >= 3 {
			ic.regionManager.CreateFreehandSelection(ic.currentPoints)
			if ic.onSelectionChanged != nil {
				ic.onSelectionChanged(true)
			}
		}
	}

	ic.currentPoints = make([]image.Point, 0)
	ic.overlay.Refresh()
}

func (ic *InteractiveCanvas) Dragged(event *fyne.DragEvent) {
	if !ic.isDrawing {
		return
	}

	ic.mousePos = event.Position

	switch ic.activeTool {
	case "rectangle":
		ic.overlay.Refresh()
	case "lasso":
		imagePoint := ic.screenToImageCoords(event.Position)
		ic.currentPoints = append(ic.currentPoints, imagePoint)
		ic.overlay.Refresh()
	}
}

func (ic *InteractiveCanvas) screenToImageCoords(screenPos fyne.Position) image.Point {
	if !ic.imageData.HasImage() {
		return image.Point{}
	}

	metadata := ic.imageData.GetMetadata()
	widgetSize := ic.Size()

	scaleX := float64(widgetSize.Width) / float64(metadata.Width)
	scaleY := float64(widgetSize.Height) / float64(metadata.Height)
	scale := math.Min(scaleX, scaleY) * ic.currentZoom

	displayWidth := float64(metadata.Width) * scale
	displayHeight := float64(metadata.Height) * scale
	offsetX := (float64(widgetSize.Width) - displayWidth) / 2
	offsetY := (float64(widgetSize.Height) - displayHeight) / 2

	imageX := (float64(screenPos.X) - offsetX) / scale
	imageY := (float64(screenPos.Y) - offsetY) / scale

	imageX = math.Max(0, math.Min(imageX, float64(metadata.Width-1)))
	imageY = math.Max(0, math.Min(imageY, float64(metadata.Height-1)))

	return image.Point{X: int(imageX), Y: int(imageY)}
}

func (ic *InteractiveCanvas) createOverlay(w, h int) image.Image {
	overlay := image.NewRGBA(image.Rect(0, 0, w, h))

	// Draw existing selections
	selections := ic.regionManager.GetAllSelections()
	for _, selection := range selections {
		ic.drawSelection(overlay, selection, w, h)
	}

	// Draw current selection being drawn
	if ic.isDrawing && len(ic.currentPoints) > 0 {
		ic.drawCurrentSelection(overlay, w, h)
	}

	return overlay
}

func (ic *InteractiveCanvas) drawSelection(overlay *image.RGBA, selection *core.Selection, w, h int) {
	selectionColor := color.RGBA{R: 255, G: 255, B: 0, A: 180} // Yellow dashed outline

	switch selection.Type {
	case core.SelectionRectangle:
		if len(selection.Points) >= 2 {
			rect := image.Rect(selection.Points[0].X, selection.Points[0].Y,
				selection.Points[1].X, selection.Points[1].Y)
			ic.drawDashedRectangle(overlay, rect, selectionColor, w, h)
		}
	case core.SelectionFreehand:
		ic.drawDashedPolygon(overlay, selection.Points, selectionColor, w, h)
	}
}

func (ic *InteractiveCanvas) drawCurrentSelection(overlay *image.RGBA, w, h int) {
	currentColor := color.RGBA{R: 0, G: 200, B: 255, A: 180} // Blue for current selection

	switch ic.activeTool {
	case "rectangle":
		if len(ic.currentPoints) > 0 && ic.isDrawing {
			imagePoint := ic.screenToImageCoords(ic.mousePos)
			startPoint := ic.currentPoints[0]
			rect := image.Rect(
				int(math.Min(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Min(float64(startPoint.Y), float64(imagePoint.Y))),
				int(math.Max(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Max(float64(startPoint.Y), float64(imagePoint.Y))),
			)
			ic.drawDashedRectangle(overlay, rect, currentColor, w, h)
		}
	case "lasso":
		if len(ic.currentPoints) > 1 {
			ic.drawDashedPolygon(overlay, ic.currentPoints, currentColor, w, h)
		}
	}
}

func (ic *InteractiveCanvas) drawDashedRectangle(overlay *image.RGBA, rect image.Rectangle, col color.RGBA, w, h int) {
	screenRect := ic.imageToScreenRect(rect, w, h)

	// Draw dashed lines
	dashLength := 5
	for x := screenRect.Min.X; x <= screenRect.Max.X; x += dashLength * 2 {
		for i := 0; i < dashLength && x+i <= screenRect.Max.X; i++ {
			if x+i >= 0 && x+i < w {
				if screenRect.Min.Y >= 0 && screenRect.Min.Y < h {
					overlay.Set(x+i, screenRect.Min.Y, col)
				}
				if screenRect.Max.Y >= 0 && screenRect.Max.Y < h {
					overlay.Set(x+i, screenRect.Max.Y, col)
				}
			}
		}
	}
	for y := screenRect.Min.Y; y <= screenRect.Max.Y; y += dashLength * 2 {
		for i := 0; i < dashLength && y+i <= screenRect.Max.Y; i++ {
			if y+i >= 0 && y+i < h {
				if screenRect.Min.X >= 0 && screenRect.Min.X < w {
					overlay.Set(screenRect.Min.X, y+i, col)
				}
				if screenRect.Max.X >= 0 && screenRect.Max.X < w {
					overlay.Set(screenRect.Max.X, y+i, col)
				}
			}
		}
	}
}

func (ic *InteractiveCanvas) drawDashedPolygon(overlay *image.RGBA, points []image.Point, col color.RGBA, w, h int) {
	if len(points) < 2 {
		return
	}

	for i := 0; i < len(points)-1; i++ {
		p1 := ic.imageToScreenPoint(points[i], w, h)
		p2 := ic.imageToScreenPoint(points[i+1], w, h)
		ic.drawDashedLine(overlay, p1, p2, col, w, h)
	}

	if len(points) >= 3 {
		p1 := ic.imageToScreenPoint(points[len(points)-1], w, h)
		p2 := ic.imageToScreenPoint(points[0], w, h)
		ic.drawDashedLine(overlay, p1, p2, col, w, h)
	}
}

func (ic *InteractiveCanvas) drawDashedLine(overlay *image.RGBA, p1, p2 image.Point, col color.RGBA, w, h int) {
	dx := int(math.Abs(float64(p2.X - p1.X)))
	dy := int(math.Abs(float64(p2.Y - p1.Y)))
	sx := -1
	if p1.X < p2.X {
		sx = 1
	}
	sy := -1
	if p1.Y < p2.Y {
		sy = 1
	}
	err := dx - dy

	x, y := p1.X, p1.Y
	dashCount := 0
	dashLength := 5

	for {
		if dashCount < dashLength {
			if x >= 0 && x < w && y >= 0 && y < h {
				overlay.Set(x, y, col)
			}
		}
		dashCount = (dashCount + 1) % (dashLength * 2)

		if x == p2.X && y == p2.Y {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

func (ic *InteractiveCanvas) imageToScreenRect(rect image.Rectangle, w, h int) image.Rectangle {
	p1 := ic.imageToScreenPoint(rect.Min, w, h)
	p2 := ic.imageToScreenPoint(rect.Max, w, h)
	return image.Rect(p1.X, p1.Y, p2.X, p2.Y)
}

func (ic *InteractiveCanvas) imageToScreenPoint(imagePoint image.Point, w, h int) image.Point {
	if !ic.imageData.HasImage() {
		return image.Point{}
	}

	metadata := ic.imageData.GetMetadata()

	scaleX := float64(w) / float64(metadata.Width)
	scaleY := float64(h) / float64(metadata.Height)
	scale := math.Min(scaleX, scaleY) * ic.currentZoom

	displayWidth := float64(metadata.Width) * scale
	displayHeight := float64(metadata.Height) * scale
	offsetX := (float64(w) - displayWidth) / 2
	offsetY := (float64(h) - displayHeight) / 2

	screenX := float64(imagePoint.X)*scale + offsetX
	screenY := float64(imagePoint.Y)*scale + offsetY

	return image.Point{X: int(screenX), Y: int(screenY)}
}

// StaticCanvas displays non-interactive preview
type StaticCanvas struct {
	widget.BaseWidget

	logger *slog.Logger

	currentImage *canvas.Image
	currentZoom  float64
}

func NewStaticCanvas(logger *slog.Logger) *StaticCanvas {
	canvas := &StaticCanvas{
		logger:      logger,
		currentZoom: 1.0,
	}

	canvas.ExtendBaseWidget(canvas)
	return canvas
}

func (sc *StaticCanvas) CreateRenderer() fyne.WidgetRenderer {
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, color.RGBA{248, 250, 252, 255})
		}
	}

	sc.currentImage = canvas.NewImageFromImage(placeholder)
	sc.currentImage.FillMode = canvas.ImageFillContain
	sc.currentImage.ScaleMode = canvas.ImageScalePixels

	return &staticCanvasRenderer{
		canvas: sc,
		image:  sc.currentImage,
	}
}

func (sc *StaticCanvas) UpdateImage(img image.Image) {
	sc.currentImage.Image = img
	sc.currentImage.Refresh()
}

func (sc *StaticCanvas) SetZoom(zoom float64) {
	sc.currentZoom = zoom
	sc.currentImage.Refresh()
}

func (sc *StaticCanvas) Clear() {
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, color.RGBA{248, 250, 252, 255})
		}
	}
	sc.UpdateImage(placeholder)
}

// Renderer implementations
type interactiveCanvasRenderer struct {
	canvas  *InteractiveCanvas
	image   *canvas.Image
	overlay *canvas.Raster
}

func (r *interactiveCanvasRenderer) Layout(size fyne.Size) {
	r.image.Resize(size)
	r.overlay.Resize(size)
}

func (r *interactiveCanvasRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, 300)
}

func (r *interactiveCanvasRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.image, r.overlay}
}

func (r *interactiveCanvasRenderer) Refresh() {
	r.image.Refresh()
	r.overlay.Refresh()
}

func (r *interactiveCanvasRenderer) Destroy() {}

type staticCanvasRenderer struct {
	canvas *StaticCanvas
	image  *canvas.Image
}

func (r *staticCanvasRenderer) Layout(size fyne.Size) {
	r.image.Resize(size)
}

func (r *staticCanvasRenderer) MinSize() fyne.Size {
	return fyne.NewSize(400, 300)
}

func (r *staticCanvasRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.image}
}

func (r *staticCanvasRenderer) Refresh() {
	r.image.Refresh()
}

func (r *staticCanvasRenderer) Destroy() {}
