// Interactive canvas widget for ROI selection
package gui

import (
	"image"
	"image/color"
	"log/slog"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/core"
)

// InteractiveCanvas is a custom widget that handles image display and ROI selection
type InteractiveCanvas struct {
	widget.BaseWidget

	imageData     *core.ImageData
	regionManager *core.RegionManager
	logger        *slog.Logger

	currentImage  *canvas.Image
	overlayRaster *canvas.Raster

	activeTool      string
	isDrawing       bool
	startPoint      fyne.Position
	currentPoints   []image.Point
	currentMousePos fyne.Position

	onSelectionChanged func(bool)
}

func NewInteractiveCanvas(imageData *core.ImageData, regionManager *core.RegionManager, logger *slog.Logger) *InteractiveCanvas {
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
}ic *InteractiveCanvas) CreateRenderer() fyne.WidgetRenderer {
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

func (ic *InteractiveCanvas) SetActiveTool(tool string) {
	ic.activeTool = tool
	ic.logger.Debug("Active tool changed in interactive canvas", "tool", tool)
}

func (ic *InteractiveCanvas) UpdateImage(img image.Image) {
	if img != nil {
		ic.currentImage.Image = img
		ic.currentImage.Refresh()
	}
}

func (ic *InteractiveCanvas) MouseDown(event *desktop.MouseEvent) {
	if ic.activeTool == "none" || !ic.imageData.HasImage() {
		return
	}

	ic.isDrawing = true
	ic.startPoint = event.Position
	ic.currentMousePos = event.Position

	imagePoint := ic.screenToImageCoords(event.Position)

	switch ic.activeTool {
	case "rectangle":
		ic.currentPoints = []image.Point{imagePoint}
	case "freehand":
		ic.currentPoints = []image.Point{imagePoint}
	}

	ic.logger.Debug("Mouse down in interactive canvas", "tool", ic.activeTool, "point", imagePoint)
}

func (ic *InteractiveCanvas) MouseUp(event *desktop.MouseEvent) {
	if !ic.isDrawing {
		return
	}

	if ic.activeTool == "freehand" {
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
				selectionID := ic.regionManager.CreateRectangleSelection(rect)
				ic.logger.Debug("Created rectangle selection", "selection_id", selectionID)
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
		ic.overlayRaster.Refresh()
	case "freehand":
		ic.currentPoints = append(ic.currentPoints, imagePoint)
		ic.overlayRaster.Refresh()
		ic.logger.Debug("Added freehand point", "points_count", len(ic.currentPoints))
	}
}

func (ic *InteractiveCanvas) DragEnd() {
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
				ic.logger.Debug("Created rectangle selection", "selection_id", selectionID)
				ic.notifySelectionChanged(true)
			}
		}
	case "freehand":
		if len(ic.currentPoints) >= 3 {
			selectionID := ic.regionManager.CreateFreehandSelection(ic.currentPoints)
			ic.logger.Debug("Created freehand selection", "selection_id", selectionID)
			ic.notifySelectionChanged(true)
		}
	}

	ic.currentPoints = make([]image.Point, 0)
	ic.overlayRaster.Refresh()
}

func (ic *InteractiveCanvas) DoubleTapped(event *fyne.PointEvent) {
	if ic.activeTool == "freehand" && len(ic.currentPoints) >= 3 {
		selectionID := ic.regionManager.CreateFreehandSelection(ic.currentPoints)
		ic.logger.Debug("Finished freehand selection", "selection_id", selectionID)
		ic.notifySelectionChanged(true)
		ic.currentPoints = make([]image.Point, 0)
		ic.overlayRaster.Refresh()
	}
}

func (ic *InteractiveCanvas) screenToImageCoords(screenPos fyne.Position) image.Point {
	if !ic.imageData.HasImage() {
		return image.Point{}
	}

	widgetSize := ic.Size()
	metadata := ic.imageData.GetMetadata()
	imageSize := image.Point{X: metadata.Width, Y: metadata.Height}

	scaleX := float64(widgetSize.Width) / float64(imageSize.X)
	scaleY := float64(widgetSize.Height) / float64(imageSize.Y)
	scale := math.Min(scaleX, scaleY)

	displayWidth := float64(imageSize.X) * scale
	displayHeight := float64(imageSize.Y) * scale
	offsetX := (float64(widgetSize.Width) - displayWidth) / 2
	offsetY := (float64(widgetSize.Height) - displayHeight) / 2

	imageX := (float64(screenPos.X) - offsetX) / scale
	imageY := (float64(screenPos.Y) - offsetY) / scale

	imageX = math.Max(0, math.Min(imageX, float64(imageSize.X-1)))
	imageY = math.Max(0, math.Min(imageY, float64(imageSize.Y-1)))

	return image.Point{X: int(imageX), Y: int(imageY)}
}

func (ic *InteractiveCanvas) createOverlay(w, h int) image.Image {
	overlay := image.NewRGBA(image.Rect(0, 0, w, h))

	selections := ic.regionManager.GetAllSelections()
	for _, selection := range selections {
		ic.drawSelection(overlay, selection, w, h)
	}

	if ic.isDrawing && len(ic.currentPoints) > 0 {
		ic.drawCurrentSelection(overlay, w, h)
	}

	return overlay
}

func (ic *InteractiveCanvas) drawSelection(overlay *image.RGBA, selection *core.Selection, w, h int) {
	selectionColor := color.RGBA{R: 255, G: 0, B: 0, A: 128}

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

func (ic *InteractiveCanvas) drawCurrentSelection(overlay *image.RGBA, w, h int) {
	currentColor := color.RGBA{R: 0, G: 255, B: 0, A: 128}

	switch ic.activeTool {
	case "rectangle":
		if len(ic.currentPoints) > 0 && ic.isDrawing {
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

func (ic *InteractiveCanvas) drawRectangleOverlay(overlay *image.RGBA, rect image.Rectangle, col color.RGBA, w, h int) {
	screenRect := ic.imageToScreenRect(rect, w, h)

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

func (ic *InteractiveCanvas) drawPolygonOverlay(overlay *image.RGBA, points []image.Point, col color.RGBA, w, h int) {
	if len(points) < 2 {
		return
	}

	for i := 0; i < len(points)-1; i++ {
		p1 := ic.imageToScreenPoint(points[i], w, h)
		p2 := ic.imageToScreenPoint(points[i+1], w, h)
		ic.drawLine(overlay, p1, p2, col, w, h)
	}

	if len(points) >= 3 {
		p1 := ic.imageToScreenPoint(points[len(points)-1], w, h)
		p2 := ic.imageToScreenPoint(points[0], w, h)
		ic.drawLine(overlay, p1, p2, col, w, h)
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
	imageSize := image.Point{X: metadata.Width, Y: metadata.Height}

	scaleX := float64(w) / float64(imageSize.X)
	scaleY := float64(h) / float64(imageSize.Y)
	scale := math.Min(scaleX, scaleY)

	displayWidth := float64(imageSize.X) * scale
	displayHeight := float64(imageSize.Y) * scale
	offsetX := (float64(w) - displayWidth) / 2
	offsetY := (float64(h) - displayHeight) / 2

	screenX := float64(imagePoint.X)*scale + offsetX
	screenY := float64(imagePoint.Y)*scale + offsetY

	return image.Point{X: int(screenX), Y: int(screenY)}
}

func (ic *InteractiveCanvas) drawLine(overlay *image.RGBA, p1, p2 image.Point, col color.RGBA, w, h int) {
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

func (ic *InteractiveCanvas) SetSelectionChangedCallback(callback func(bool)) {
	ic.onSelectionChanged = callback
}

func (ic *InteractiveCanvas) notifySelectionChanged(hasSelection bool) {
	if ic.onSelectionChanged != nil {
		ic.onSelectionChanged(hasSelection)
	}
}

func (ic *InteractiveCanvas) RefreshSelections() {
	ic.overlayRaster.Refresh()
}

type interactiveCanvasRenderer struct {
	canvas  *InteractiveCanvas
	image   *canvas.Image
	overlay *canvas.Raster
}

func (