// Interactive canvas widget for ROI selection
package gui

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"

	"advanced-image-processing/internal/core"
)

// InteractiveCanvas is a custom widget that handles image display and ROI selection
type InteractiveCanvas struct {
	widget.BaseWidget

	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *logrus.Logger

	currentImage  *canvas.Image
	overlayRaster *canvas.Raster

	activeTool      string
	isDrawing       bool
	startPoint      fyne.Position
	currentPoints   []image.Point
	currentMousePos fyne.Position

	onSelectionChanged func(bool)
}

// NewInteractiveCanvas creates a new interactive canvas
func NewInteractiveCanvas(imageData *core.ImageData, regionManager *core.RegionManager, logger *logrus.Logger) *InteractiveCanvas {
	ic := &InteractiveCanvas{
		imageData:     imageData,
		regionManager: regionManager,
		logger:        logger,
		activeTool:    "none",
		currentPoints: make([]image.Point, 0),
	}

	ic.ExtendBaseWidget(ic)
	return ic
}

// CreateRenderer creates the renderer for the interactive canvas
func (ic *InteractiveCanvas) CreateRenderer() fyne.WidgetRenderer {
	ic.currentImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ic.currentImage.FillMode = canvas.ImageFillContain

	ic.overlayRaster = canvas.NewRaster(func(w, h int) image.Image {
		return ic.createOverlay(w, h)
	})

	return &interactiveCanvasRenderer{
		canvas:  ic,
		image:   ic.currentImage,
		overlay: ic.overlayRaster,
	}
}

// SetActiveTool sets the active selection tool
func (ic *InteractiveCanvas) SetActiveTool(tool string) {
	ic.activeTool = tool
	ic.logger.WithField("tool", tool).Debug("Active tool changed in interactive canvas")
}

// UpdateImage updates the displayed image
func (ic *InteractiveCanvas) UpdateImage(img image.Image) {
	if img != nil {
		ic.currentImage.Image = img
		ic.currentImage.Refresh()
	}
}

// Mouse event handlers
func (ic *InteractiveCanvas) MouseDown(event *desktop.MouseEvent) {
	if ic.activeTool == "none" || !ic.imageData.HasImage() {
		return
	}

	ic.isDrawing = true
	ic.startPoint = event.Position
	ic.currentMousePos = event.Position

	// Convert screen coordinates to image coordinates
	imagePoint := ic.screenToImageCoords(event.Position)

	switch ic.activeTool {
	case "rectangle":
		ic.currentPoints = []image.Point{imagePoint}
	case "freehand":
		ic.currentPoints = []image.Point{imagePoint}
	}

	ic.logger.WithFields(logrus.Fields{
		"tool":  ic.activeTool,
		"point": imagePoint,
	}).Debug("Mouse down in interactive canvas")
}

func (ic *InteractiveCanvas) MouseUp(event *desktop.MouseEvent) {
	// Only handle mouse up for non-drag operations
	if !ic.isDrawing {
		return
	}

	// For freehand, rely on drag events instead
	if ic.activeTool == "freehand" {
		return
	}

	ic.isDrawing = false

	// Convert screen coordinates to image coordinates
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
				selectionID := ic.regionManager.CreateRectangleSelection(rect)
				ic.logger.WithField("selection_id", selectionID).Debug("Created rectangle selection")
				ic.notifySelectionChanged(true)
			}
		}
	}

	ic.currentPoints = make([]image.Point, 0)
	ic.overlayRaster.Refresh()
}

func (ic *InteractiveCanvas) Dragged(event *fyne.DragEvent) {
	if !ic.isDrawing {
		return
	}

	ic.currentMousePos = event.Position
	imagePoint := ic.screenToImageCoords(event.Position)

	switch ic.activeTool {
	case "rectangle":
		// For rectangle, just refresh for preview
		ic.overlayRaster.Refresh()
	case "freehand":
		// Add point to freehand path
		ic.currentPoints = append(ic.currentPoints, imagePoint)
		ic.overlayRaster.Refresh()
		ic.logger.WithField("points_count", len(ic.currentPoints)).Debug("Added freehand point")
	}
}

func (ic *InteractiveCanvas) DragEnd() {
	// Same as MouseUp for drag completion
	if !ic.isDrawing {
		return
	}

	ic.isDrawing = false

	switch ic.activeTool {
	case "rectangle":
		if len(ic.currentPoints) > 0 {
			imagePoint := ic.screenToImageCoords(ic.currentMousePos)
			startPoint := ic.currentPoints[0]
			rect := image.Rect(
				int(math.Min(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Min(float64(startPoint.Y), float64(imagePoint.Y))),
				int(math.Max(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Max(float64(startPoint.Y), float64(imagePoint.Y))),
			)
			if !rect.Empty() {
				selectionID := ic.regionManager.CreateRectangleSelection(rect)
				ic.logger.WithField("selection_id", selectionID).Debug("Created rectangle selection")
				ic.notifySelectionChanged(true)
			}
		}
	case "freehand":
		if len(ic.currentPoints) >= 3 {
			selectionID := ic.regionManager.CreateFreehandSelection(ic.currentPoints)
			ic.logger.WithField("selection_id", selectionID).Debug("Created freehand selection")
			ic.notifySelectionChanged(true)
		}
	}

	ic.currentPoints = make([]image.Point, 0)
	ic.overlayRaster.Refresh()
}

// Double-click to finish freehand selection
func (ic *InteractiveCanvas) DoubleTapped(event *fyne.PointEvent) {
	if ic.activeTool == "freehand" && len(ic.currentPoints) >= 3 {
		selectionID := ic.regionManager.CreateFreehandSelection(ic.currentPoints)
		ic.logger.WithField("selection_id", selectionID).Debug("Finished freehand selection")
		ic.notifySelectionChanged(true)
		ic.currentPoints = make([]image.Point, 0)
		ic.overlayRaster.Refresh()
	}
}

// screenToImageCoords converts screen coordinates to image coordinates
func (ic *InteractiveCanvas) screenToImageCoords(screenPos fyne.Position) image.Point {
	if !ic.imageData.HasImage() {
		return image.Point{}
	}

	// Get widget size and image size
	widgetSize := ic.Size()
	metadata := ic.imageData.GetMetadata()
	imageSize := image.Point{X: metadata.Width, Y: metadata.Height}

	// Calculate scale factor (considering ImageFillContain behavior)
	scaleX := float64(widgetSize.Width) / float64(imageSize.X)
	scaleY := float64(widgetSize.Height) / float64(imageSize.Y)
	scale := math.Min(scaleX, scaleY)

	// Calculate displayed image size and offset
	displayWidth := float64(imageSize.X) * scale
	displayHeight := float64(imageSize.Y) * scale
	offsetX := (float64(widgetSize.Width) - displayWidth) / 2
	offsetY := (float64(widgetSize.Height) - displayHeight) / 2

	// Convert screen coordinates to image coordinates
	imageX := (float64(screenPos.X) - offsetX) / scale
	imageY := (float64(screenPos.Y) - offsetY) / scale

	// Clamp to image bounds
	imageX = math.Max(0, math.Min(imageX, float64(imageSize.X-1)))
	imageY = math.Max(0, math.Min(imageY, float64(imageSize.Y-1)))

	return image.Point{X: int(imageX), Y: int(imageY)}
}

// createOverlay creates the selection overlay
func (ic *InteractiveCanvas) createOverlay(w, h int) image.Image {
	overlay := image.NewRGBA(image.Rect(0, 0, w, h))

	// Draw existing selections
	selections := ic.regionManager.GetAllSelections()
	for _, selection := range selections {
		ic.drawSelection(overlay, selection, w, h)
	}

	// Draw current drawing
	if ic.isDrawing && len(ic.currentPoints) > 0 {
		ic.drawCurrentSelection(overlay, w, h)
	}

	return overlay
}

// drawSelection draws a selection on the overlay
func (ic *InteractiveCanvas) drawSelection(overlay *image.RGBA, selection *core.Selection, w, h int) {
	selectionColor := color.RGBA{R: 255, G: 0, B: 0, A: 128} // Semi-transparent red

	switch selection.Type {
	case core.SelectionRectangle:
		if len(selection.Points) >= 2 {
			rect := image.Rect(selection.Points[0].X, selection.Points[0].Y,
				selection.Points[1].X, selection.Points[1].Y)
			ic.drawRectangleOverlay(overlay, rect, selectionColor, w, h)
		}
	case core.SelectionFreehand:
		ic.drawPolygonOverlay(overlay, selection.Points, selectionColor, w, h)
	}
}

// drawCurrentSelection draws the current selection being drawn
func (ic *InteractiveCanvas) drawCurrentSelection(overlay *image.RGBA, w, h int) {
	currentColor := color.RGBA{R: 0, G: 255, B: 0, A: 128} // Semi-transparent green

	switch ic.activeTool {
	case "rectangle":
		if len(ic.currentPoints) > 0 && ic.isDrawing {
			// Use current mouse position for rectangle preview
			imagePoint := ic.screenToImageCoords(ic.currentMousePos)
			startPoint := ic.currentPoints[0]
			rect := image.Rect(
				int(math.Min(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Min(float64(startPoint.Y), float64(imagePoint.Y))),
				int(math.Max(float64(startPoint.X), float64(imagePoint.X))),
				int(math.Max(float64(startPoint.Y), float64(imagePoint.Y))),
			)
			ic.drawRectangleOverlay(overlay, rect, currentColor, w, h)
		}
	case "freehand":
		if len(ic.currentPoints) > 1 {
			ic.drawPolygonOverlay(overlay, ic.currentPoints, currentColor, w, h)
		}
	}
}

// drawRectangleOverlay draws a rectangle overlay
func (ic *InteractiveCanvas) drawRectangleOverlay(overlay *image.RGBA, rect image.Rectangle, col color.RGBA, w, h int) {
	// Convert image coordinates to screen coordinates
	screenRect := ic.imageToScreenRect(rect, w, h)

	// Draw rectangle border
	for x := screenRect.Min.X; x <= screenRect.Max.X; x++ {
		if x >= 0 && x < w {
			if screenRect.Min.Y >= 0 && screenRect.Min.Y < h {
				overlay.Set(x, screenRect.Min.Y, col)
			}
			if screenRect.Max.Y >= 0 && screenRect.Max.Y < h {
				overlay.Set(x, screenRect.Max.Y, col)
			}
		}
	}
	for y := screenRect.Min.Y; y <= screenRect.Max.Y; y++ {
		if y >= 0 && y < h {
			if screenRect.Min.X >= 0 && screenRect.Min.X < w {
				overlay.Set(screenRect.Min.X, y, col)
			}
			if screenRect.Max.X >= 0 && screenRect.Max.X < w {
				overlay.Set(screenRect.Max.X, y, col)
			}
		}
	}
}

// drawPolygonOverlay draws a polygon overlay
func (ic *InteractiveCanvas) drawPolygonOverlay(overlay *image.RGBA, points []image.Point, col color.RGBA, w, h int) {
	if len(points) < 2 {
		return
	}

	// Convert points to screen coordinates and draw lines between them
	for i := 0; i < len(points)-1; i++ {
		p1 := ic.imageToScreenPoint(points[i], w, h)
		p2 := ic.imageToScreenPoint(points[i+1], w, h)
		ic.drawLine(overlay, p1, p2, col, w, h)
	}

	// Close the polygon if we have enough points
	if len(points) >= 3 {
		p1 := ic.imageToScreenPoint(points[len(points)-1], w, h)
		p2 := ic.imageToScreenPoint(points[0], w, h)
		ic.drawLine(overlay, p1, p2, col, w, h)
	}
}

// imageToScreenRect converts image rectangle to screen rectangle
func (ic *InteractiveCanvas) imageToScreenRect(rect image.Rectangle, w, h int) image.Rectangle {
	p1 := ic.imageToScreenPoint(rect.Min, w, h)
	p2 := ic.imageToScreenPoint(rect.Max, w, h)
	return image.Rect(p1.X, p1.Y, p2.X, p2.Y)
}

// imageToScreenPoint converts image point to screen point
func (ic *InteractiveCanvas) imageToScreenPoint(imagePoint image.Point, w, h int) image.Point {
	if !ic.imageData.HasImage() {
		return image.Point{}
	}

	metadata := ic.imageData.GetMetadata()
	imageSize := image.Point{X: metadata.Width, Y: metadata.Height}

	// Calculate scale factor
	scaleX := float64(w) / float64(imageSize.X)
	scaleY := float64(h) / float64(imageSize.Y)
	scale := math.Min(scaleX, scaleY)

	// Calculate displayed image size and offset
	displayWidth := float64(imageSize.X) * scale
	displayHeight := float64(imageSize.Y) * scale
	offsetX := (float64(w) - displayWidth) / 2
	offsetY := (float64(h) - displayHeight) / 2

	// Convert image coordinates to screen coordinates
	screenX := float64(imagePoint.X)*scale + offsetX
	screenY := float64(imagePoint.Y)*scale + offsetY

	return image.Point{X: int(screenX), Y: int(screenY)}
}

// drawLine draws a line between two points
func (ic *InteractiveCanvas) drawLine(overlay *image.RGBA, p1, p2 image.Point, col color.RGBA, w, h int) {
	// Simple line drawing using Bresenham's algorithm
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
		if x >= 0 && x < w && y >= 0 && y < h {
			overlay.Set(x, y, col)
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

// SetSelectionChangedCallback sets the callback for selection changes
func (ic *InteractiveCanvas) SetSelectionChangedCallback(callback func(bool)) {
	ic.onSelectionChanged = callback
}

// notifySelectionChanged notifies about selection changes
func (ic *InteractiveCanvas) notifySelectionChanged(hasSelection bool) {
	if ic.onSelectionChanged != nil {
		ic.onSelectionChanged(hasSelection)
	}
}

// RefreshSelections refreshes the selection overlay
func (ic *InteractiveCanvas) RefreshSelections() {
	ic.overlayRaster.Refresh()
}

// interactiveCanvasRenderer is the renderer for the interactive canvas
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

func (r *interactiveCanvasRenderer) Destroy() {
}
