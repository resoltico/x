// internal/gui/image_workspace.go
// Modern image workspace with enhanced viewing and interaction
package gui

import (
	"image"
	"image/color"
	"log/slog"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/core"
)

// ImageWorkspace provides the main image viewing and interaction area
type ImageWorkspace struct {
	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *slog.Logger

	container *fyne.Container

	// View mode toggle
	viewToggle *widget.RadioGroup

	// Image display containers
	splitView  *container.Split
	singleView *fyne.Container

	// Interactive canvases
	originalCanvas *InteractiveImageCanvas
	previewCanvas  *StaticImageCanvas

	// Current state
	isSplitView bool
	currentZoom float64
	activeTool  string

	// Callbacks
	onSelectionChanged func(bool)
	onZoomChanged      func(float64)
}

func NewImageWorkspace(imageData *core.ImageData, regionManager *core.RegionManager, logger *slog.Logger) *ImageWorkspace {
	workspace := &ImageWorkspace{
		imageData:     imageData,
		regionManager: regionManager,
		logger:        logger,
		isSplitView:   true,
		currentZoom:   1.0,
		activeTool:    "none",
	}

	workspace.initializeUI()
	return workspace
}

func (iw *ImageWorkspace) initializeUI() {
	// Create interactive canvases FIRST
	iw.originalCanvas = NewInteractiveImageCanvas(iw.imageData, iw.regionManager, iw.logger)
	iw.previewCanvas = NewStaticImageCanvas(iw.logger)

	// Set up canvas callbacks
	iw.originalCanvas.SetSelectionCallback(func(hasSelection bool) {
		if iw.onSelectionChanged != nil {
			iw.onSelectionChanged(hasSelection)
		}
	})

	// Create view containers
	originalCard := widget.NewCard("Original", "", iw.originalCanvas)
	previewCard := widget.NewCard("Preview", "", iw.previewCanvas)

	// Split view (default)
	iw.splitView = container.NewVSplit(originalCard, previewCard)
	iw.splitView.SetOffset(0.5) // Equal split

	// Single view (preview only)
	iw.singleView = container.NewBorder(nil, nil, nil, nil, previewCard)

	// View mode toggle - DON'T set callback yet
	iw.viewToggle = widget.NewRadioGroup([]string{"Split View", "Preview Only"}, nil)
	iw.viewToggle.Horizontal = true

	// Main container with view toggle at top
	viewControls := container.NewHBox(
		widget.NewLabel("View Mode:"),
		iw.viewToggle,
	)

	iw.container = container.NewBorder(
		viewControls, // top
		nil,          // bottom
		nil,          // left
		nil,          // right
		iw.splitView, // center (initial view)
	)
	iw.container.Resize(fyne.NewSize(1000, 950))

	// NOW set the callback and selected value after container is initialized
	iw.viewToggle.OnChanged = func(value string) {
		iw.isSplitView = (value == "Split View")
		iw.updateViewMode()
	}
	iw.viewToggle.SetSelected("Split View")
}

func (iw *ImageWorkspace) updateViewMode() {
	// Remove current view
	iw.container.Objects = iw.container.Objects[:1] // Keep only top (view controls)

	if iw.isSplitView {
		iw.container.Add(iw.splitView)
	} else {
		iw.container.Add(iw.singleView)
	}

	iw.container.Refresh()
}

func (iw *ImageWorkspace) GetContainer() fyne.CanvasObject {
	return iw.container
}

func (iw *ImageWorkspace) UpdateOriginal() {
	if !iw.imageData.HasImage() {
		return
	}

	original := iw.imageData.GetOriginal()
	defer original.Close()

	if !original.Empty() {
		img, err := original.ToImage()
		if err != nil {
			iw.logger.Error("Failed to convert original to image", "error", err)
			return
		}
		iw.originalCanvas.UpdateImage(img)
	}
}

func (iw *ImageWorkspace) UpdatePreview(preview image.Image) {
	iw.previewCanvas.UpdateImage(preview)
}

func (iw *ImageWorkspace) SetActiveTool(tool string) {
	iw.activeTool = tool
	iw.originalCanvas.SetActiveTool(tool)
}

func (iw *ImageWorkspace) SetZoom(zoom float64) {
	iw.currentZoom = zoom
	iw.originalCanvas.SetZoom(zoom)
	iw.previewCanvas.SetZoom(zoom)

	if iw.onZoomChanged != nil {
		iw.onZoomChanged(zoom)
	}
}

func (iw *ImageWorkspace) RefreshSelections() {
	iw.originalCanvas.RefreshSelections()
}

func (iw *ImageWorkspace) Reset() {
	iw.originalCanvas.Clear()
	iw.previewCanvas.Clear()
	iw.currentZoom = 1.0
	iw.activeTool = "none"
}

func (iw *ImageWorkspace) SetCallbacks(onSelectionChanged func(bool), onZoomChanged func(float64)) {
	iw.onSelectionChanged = onSelectionChanged
	iw.onZoomChanged = onZoomChanged
}

// InteractiveImageCanvas handles user interaction for ROI selection
type InteractiveImageCanvas struct {
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

func NewInteractiveImageCanvas(imageData *core.ImageData, regionManager *core.RegionManager, logger *slog.Logger) *InteractiveImageCanvas {
	canvas := &InteractiveImageCanvas{
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

func (iic *InteractiveImageCanvas) CreateRenderer() fyne.WidgetRenderer {
	// Create placeholder image
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}

	iic.currentImage = canvas.NewImageFromImage(placeholder)
	iic.currentImage.FillMode = canvas.ImageFillContain
	iic.currentImage.ScaleMode = canvas.ImageScalePixels

	iic.overlay = canvas.NewRaster(iic.createOverlay)

	return &interactiveCanvasRenderer{
		canvas:  iic,
		image:   iic.currentImage,
		overlay: iic.overlay,
	}
}

func (iic *InteractiveImageCanvas) UpdateImage(img image.Image) {
	iic.currentImage.Image = img
	iic.currentImage.Refresh()
	iic.overlay.Refresh()
}

func (iic *InteractiveImageCanvas) SetActiveTool(tool string) {
	iic.activeTool = tool
	iic.overlay.Refresh()
}

func (iic *InteractiveImageCanvas) SetZoom(zoom float64) {
	iic.currentZoom = zoom
	iic.currentImage.Refresh()
	iic.overlay.Refresh()
}

func (iic *InteractiveImageCanvas) Clear() {
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, color.RGBA{240, 240, 240, 255})
		}
	}
	iic.UpdateImage(placeholder)
}

func (iic *InteractiveImageCanvas) RefreshSelections() {
	iic.overlay.Refresh()
}

func (iic *InteractiveImageCanvas) SetSelectionCallback(callback func(bool)) {
	iic.onSelectionChanged = callback
}

// Mouse event handlers
func (iic *InteractiveImageCanvas) MouseDown(event *desktop.MouseEvent) {
	if iic.activeTool == "none" || !iic.imageData.HasImage() {
		return
	}

	iic.isDrawing = true
	iic.startPoint = event.Position
	iic.mousePos = event.Position

	imagePoint := iic.screenToImageCoords(event.Position)

	switch iic.activeTool {
	case "rectangle":
		iic.currentPoints = []image.Point{imagePoint}
	case "freehand":
		iic.currentPoints = []image.Point{imagePoint}
	}
}

func (iic *InteractiveImageCanvas) MouseUp(event *desktop.MouseEvent) {
	if !iic.isDrawing {
		return
	}

	iic.isDrawing = false
	imagePoint := iic.screenToImageCoords(event.Position)

	switch iic.activeTool {
	case "rectangle":
		if len(iic.currentPoints) > 0 {
			startPoint := iic.currentPoints[0]
			rect := image.Rect(
				int(math.Min(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Min(float64(startPoint.Y), float64(imagePoint.Y))),
				int(math.Max(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Max(float64(startPoint.Y), float64(imagePoint.Y))),
			)
			if !rect.Empty() {
				iic.regionManager.CreateRectangleSelection(rect)
				iic.notifySelectionChanged(true)
			}
		}
	case "freehand":
		if len(iic.currentPoints) >= 3 {
			iic.regionManager.CreateFreehandSelection(iic.currentPoints)
			iic.notifySelectionChanged(true)
		}
	}

	iic.currentPoints = make([]image.Point, 0)
	iic.overlay.Refresh()
}

func (iic *InteractiveImageCanvas) Dragged(event *fyne.DragEvent) {
	if !iic.isDrawing {
		return
	}

	iic.mousePos = event.Position

	switch iic.activeTool {
	case "rectangle":
		iic.overlay.Refresh()
	case "freehand":
		imagePoint := iic.screenToImageCoords(event.Position)
		iic.currentPoints = append(iic.currentPoints, imagePoint)
		iic.overlay.Refresh()
	}
}

func (iic *InteractiveImageCanvas) DoubleTapped(event *fyne.PointEvent) {
	if iic.activeTool == "freehand" && len(iic.currentPoints) >= 3 {
		iic.regionManager.CreateFreehandSelection(iic.currentPoints)
		iic.notifySelectionChanged(true)
		iic.currentPoints = make([]image.Point, 0)
		iic.overlay.Refresh()
	}
}

func (iic *InteractiveImageCanvas) screenToImageCoords(screenPos fyne.Position) image.Point {
	if !iic.imageData.HasImage() {
		return image.Point{}
	}

	metadata := iic.imageData.GetMetadata()
	widgetSize := iic.Size()

	// Calculate scaling and offset for centered image
	scaleX := float64(widgetSize.Width) / float64(metadata.Width)
	scaleY := float64(widgetSize.Height) / float64(metadata.Height)
	scale := math.Min(scaleX, scaleY) * iic.currentZoom

	displayWidth := float64(metadata.Width) * scale
	displayHeight := float64(metadata.Height) * scale
	offsetX := (float64(widgetSize.Width) - displayWidth) / 2
	offsetY := (float64(widgetSize.Height) - displayHeight) / 2

	imageX := (float64(screenPos.X) - offsetX) / scale
	imageY := (float64(screenPos.Y) - offsetY) / scale

	// Clamp to image bounds
	imageX = math.Max(0, math.Min(imageX, float64(metadata.Width-1)))
	imageY = math.Max(0, math.Min(imageY, float64(metadata.Height-1)))

	return image.Point{X: int(imageX), Y: int(imageY)}
}

func (iic *InteractiveImageCanvas) createOverlay(w, h int) image.Image {
	overlay := image.NewRGBA(image.Rect(0, 0, w, h))

	// Draw existing selections
	selections := iic.regionManager.GetAllSelections()
	for _, selection := range selections {
		iic.drawSelection(overlay, selection, w, h)
	}

	// Draw current selection being drawn
	if iic.isDrawing && len(iic.currentPoints) > 0 {
		iic.drawCurrentSelection(overlay, w, h)
	}

	return overlay
}

func (iic *InteractiveImageCanvas) drawSelection(overlay *image.RGBA, selection *core.Selection, w, h int) {
	selectionColor := color.RGBA{R: 255, G: 100, B: 0, A: 180} // Modern orange

	switch selection.Type {
	case core.SelectionRectangle:
		if len(selection.Points) >= 2 {
			rect := image.Rect(selection.Points[0].X, selection.Points[0].Y,
				selection.Points[1].X, selection.Points[1].Y)
			iic.drawRectangleOverlay(overlay, rect, selectionColor, w, h)
		}
	case core.SelectionFreehand:
		iic.drawPolygonOverlay(overlay, selection.Points, selectionColor, w, h)
	}
}

func (iic *InteractiveImageCanvas) drawCurrentSelection(overlay *image.RGBA, w, h int) {
	currentColor := color.RGBA{R: 0, G: 200, B: 255, A: 180} // Modern cyan

	switch iic.activeTool {
	case "rectangle":
		if len(iic.currentPoints) > 0 && iic.isDrawing {
			imagePoint := iic.screenToImageCoords(iic.mousePos)
			startPoint := iic.currentPoints[0]
			rect := image.Rect(
				int(math.Min(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Min(float64(startPoint.Y), float64(imagePoint.Y))),
				int(math.Max(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Max(float64(startPoint.Y), float64(imagePoint.Y))),
			)
			iic.drawRectangleOverlay(overlay, rect, currentColor, w, h)
		}
	case "freehand":
		if len(iic.currentPoints) > 1 {
			iic.drawPolygonOverlay(overlay, iic.currentPoints, currentColor, w, h)
		}
	}
}

func (iic *InteractiveImageCanvas) drawRectangleOverlay(overlay *image.RGBA, rect image.Rectangle, col color.RGBA, w, h int) {
	screenRect := iic.imageToScreenRect(rect, w, h)

	// Draw border with modern thick line
	thickness := 3
	for t := 0; t < thickness; t++ {
		// Top and bottom lines
		for x := screenRect.Min.X; x <= screenRect.Max.X; x++ {
			if x >= 0 && x < w {
				if screenRect.Min.Y+t >= 0 && screenRect.Min.Y+t < h {
					overlay.Set(x, screenRect.Min.Y+t, col)
				}
				if screenRect.Max.Y-t >= 0 && screenRect.Max.Y-t < h {
					overlay.Set(x, screenRect.Max.Y-t, col)
				}
			}
		}
		// Left and right lines
		for y := screenRect.Min.Y; y <= screenRect.Max.Y; y++ {
			if y >= 0 && y < h {
				if screenRect.Min.X+t >= 0 && screenRect.Min.X+t < w {
					overlay.Set(screenRect.Min.X+t, y, col)
				}
				if screenRect.Max.X-t >= 0 && screenRect.Max.X-t < w {
					overlay.Set(screenRect.Max.X-t, y, col)
				}
			}
		}
	}
}

func (iic *InteractiveImageCanvas) drawPolygonOverlay(overlay *image.RGBA, points []image.Point, col color.RGBA, w, h int) {
	if len(points) < 2 {
		return
	}

	// Draw lines between points with modern thick lines
	for i := 0; i < len(points)-1; i++ {
		p1 := iic.imageToScreenPoint(points[i], w, h)
		p2 := iic.imageToScreenPoint(points[i+1], w, h)
		iic.drawThickLine(overlay, p1, p2, col, w, h, 3)
	}

	// Close the polygon if we have enough points
	if len(points) >= 3 {
		p1 := iic.imageToScreenPoint(points[len(points)-1], w, h)
		p2 := iic.imageToScreenPoint(points[0], w, h)
		iic.drawThickLine(overlay, p1, p2, col, w, h, 3)
	}
}

func (iic *InteractiveImageCanvas) imageToScreenRect(rect image.Rectangle, w, h int) image.Rectangle {
	p1 := iic.imageToScreenPoint(rect.Min, w, h)
	p2 := iic.imageToScreenPoint(rect.Max, w, h)
	return image.Rect(p1.X, p1.Y, p2.X, p2.Y)
}

func (iic *InteractiveImageCanvas) imageToScreenPoint(imagePoint image.Point, w, h int) image.Point {
	if !iic.imageData.HasImage() {
		return image.Point{}
	}

	metadata := iic.imageData.GetMetadata()

	scaleX := float64(w) / float64(metadata.Width)
	scaleY := float64(h) / float64(metadata.Height)
	scale := math.Min(scaleX, scaleY) * iic.currentZoom

	displayWidth := float64(metadata.Width) * scale
	displayHeight := float64(metadata.Height) * scale
	offsetX := (float64(w) - displayWidth) / 2
	offsetY := (float64(h) - displayHeight) / 2

	screenX := float64(imagePoint.X)*scale + offsetX
	screenY := float64(imagePoint.Y)*scale + offsetY

	return image.Point{X: int(screenX), Y: int(screenY)}
}

func (iic *InteractiveImageCanvas) drawThickLine(overlay *image.RGBA, p1, p2 image.Point, col color.RGBA, w, h int, thickness int) {
	// Bresenham's line algorithm with thickness
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

	for {
		// Draw thick point
		for tx := -thickness / 2; tx <= thickness/2; tx++ {
			for ty := -thickness / 2; ty <= thickness/2; ty++ {
				px, py := x+tx, y+ty
				if px >= 0 && px < w && py >= 0 && py < h {
					overlay.Set(px, py, col)
				}
			}
		}

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

func (iic *InteractiveImageCanvas) notifySelectionChanged(hasSelection bool) {
	if iic.onSelectionChanged != nil {
		iic.onSelectionChanged(hasSelection)
	}
}

// StaticImageCanvas displays non-interactive preview
type StaticImageCanvas struct {
	widget.BaseWidget

	logger *slog.Logger

	currentImage *canvas.Image
	currentZoom  float64
}

func NewStaticImageCanvas(logger *slog.Logger) *StaticImageCanvas {
	canvas := &StaticImageCanvas{
		logger:      logger,
		currentZoom: 1.0,
	}

	canvas.ExtendBaseWidget(canvas)
	return canvas
}

func (sic *StaticImageCanvas) CreateRenderer() fyne.WidgetRenderer {
	// Create placeholder image
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, color.RGBA{248, 250, 252, 255}) // Light background
		}
	}

	sic.currentImage = canvas.NewImageFromImage(placeholder)
	sic.currentImage.FillMode = canvas.ImageFillContain
	sic.currentImage.ScaleMode = canvas.ImageScalePixels

	return &staticCanvasRenderer{
		canvas: sic,
		image:  sic.currentImage,
	}
}

func (sic *StaticImageCanvas) UpdateImage(img image.Image) {
	sic.currentImage.Image = img
	sic.currentImage.Refresh()
}

func (sic *StaticImageCanvas) SetZoom(zoom float64) {
	sic.currentZoom = zoom
	sic.currentImage.Refresh()
}

func (sic *StaticImageCanvas) Clear() {
	placeholder := image.NewRGBA(image.Rect(0, 0, 400, 300))
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			placeholder.Set(x, y, color.RGBA{248, 250, 252, 255})
		}
	}
	sic.UpdateImage(placeholder)
}

// Renderer implementations
type interactiveCanvasRenderer struct {
	canvas  *InteractiveImageCanvas
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
	canvas *StaticImageCanvas
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
