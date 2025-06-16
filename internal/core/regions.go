// ROI (Region of Interest) selection and management system
package core

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"gocv.io/x/gocv"
)

// SelectionType defines the type of selection
type SelectionType int

const (
	SelectionNone SelectionType = iota
	SelectionRectangle
	SelectionFreehand
)

// Selection represents a region of interest
type Selection struct {
	ID     string
	Type   SelectionType
	Points []image.Point
	Bounds image.Rectangle
	Active bool
}

// RegionManager manages multiple ROI selections
type RegionManager struct {
	mu         sync.RWMutex
	selections map[string]*Selection
	active     string
	nextID     int
}

func NewRegionManager() *RegionManager {
	return &RegionManager{
		selections: make(map[string]*Selection),
		nextID:     1,
	}
}

func (rm *RegionManager) CreateRectangleSelection(rect image.Rectangle) string {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	id := fmt.Sprintf("rect_%d", rm.nextID)
	rm.nextID++

	selection := &Selection{
		ID:     id,
		Type:   SelectionRectangle,
		Points: rectangleToPoints(rect),
		Bounds: rect,
		Active: true,
	}

	// Deactivate other selections
	for _, sel := range rm.selections {
		sel.Active = false
	}

	rm.selections[id] = selection
	rm.active = id

	return id
}

func (rm *RegionManager) CreateFreehandSelection(points []image.Point) string {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if len(points) < 3 {
		return ""
	}

	id := fmt.Sprintf("freehand_%d", rm.nextID)
	rm.nextID++

	bounds := calculateBounds(points)

	selection := &Selection{
		ID:     id,
		Type:   SelectionFreehand,
		Points: make([]image.Point, len(points)),
		Bounds: bounds,
		Active: true,
	}

	copy(selection.Points, points)

	// Deactivate other selections
	for _, sel := range rm.selections {
		sel.Active = false
	}

	rm.selections[id] = selection
	rm.active = id

	return id
}

func (rm *RegionManager) GetActiveSelection() *Selection {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.active == "" {
		return nil
	}

	selection, exists := rm.selections[rm.active]
	if !exists {
		return nil
	}

	result := &Selection{
		ID:     selection.ID,
		Type:   selection.Type,
		Points: make([]image.Point, len(selection.Points)),
		Bounds: selection.Bounds,
		Active: selection.Active,
	}
	copy(result.Points, selection.Points)

	return result
}

func (rm *RegionManager) GetSelection(id string) *Selection {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	selection, exists := rm.selections[id]
	if !exists {
		return nil
	}

	result := &Selection{
		ID:     selection.ID,
		Type:   selection.Type,
		Points: make([]image.Point, len(selection.Points)),
		Bounds: selection.Bounds,
		Active: selection.Active,
	}
	copy(result.Points, selection.Points)

	return result
}

func (rm *RegionManager) GetAllSelections() []*Selection {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	result := make([]*Selection, 0, len(rm.selections))
	for _, selection := range rm.selections {
		selectionCopy := &Selection{
			ID:     selection.ID,
			Type:   selection.Type,
			Points: make([]image.Point, len(selection.Points)),
			Bounds: selection.Bounds,
			Active: selection.Active,
		}
		copy(selectionCopy.Points, selection.Points)
		result = append(result, selectionCopy)
	}

	return result
}

func (rm *RegionManager) SetActiveSelection(id string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	selection, exists := rm.selections[id]
	if !exists {
		return false
	}

	// Deactivate all selections
	for _, sel := range rm.selections {
		sel.Active = false
	}

	// Activate the selected one
	selection.Active = true
	rm.active = id

	return true
}

func (rm *RegionManager) RemoveSelection(id string) bool {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	_, exists := rm.selections[id]
	if !exists {
		return false
	}

	delete(rm.selections, id)

	if rm.active == id {
		rm.active = ""
	}

	return true
}

func (rm *RegionManager) ClearAll() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.selections = make(map[string]*Selection)
	rm.active = ""
}

func (rm *RegionManager) CreateMask(imgWidth, imgHeight int) gocv.Mat {
	selection := rm.GetActiveSelection()
	if selection == nil {
		return gocv.NewMat()
	}

	return rm.CreateMaskForSelection(selection, imgWidth, imgHeight)
}

func (rm *RegionManager) CreateMaskForSelection(selection *Selection, imgWidth, imgHeight int) gocv.Mat {
	mask := gocv.Zeros(imgHeight, imgWidth, gocv.MatTypeCV8UC1)

	switch selection.Type {
	case SelectionRectangle:
		if len(selection.Points) >= 2 {
			rect := image.Rect(
				selection.Points[0].X,
				selection.Points[0].Y,
				selection.Points[1].X,
				selection.Points[1].Y,
			)

			rect = rect.Intersect(image.Rect(0, 0, imgWidth, imgHeight))

			if !rect.Empty() {
				roi := mask.Region(rect)
				roi.SetTo(gocv.Scalar{Val1: 255, Val2: 255, Val3: 255, Val4: 255})
				roi.Close()
			}
		}

	case SelectionFreehand:
		if len(selection.Points) >= 3 {
			// Use GoCV's optimized FillPoly
			pointsVector := gocv.NewPointsVector()
			defer pointsVector.Close()

			pointVector := gocv.NewPointVectorFromPoints(selection.Points)
			defer pointVector.Close()

			pointsVector.Append(pointVector)
			gocv.FillPoly(&mask, pointsVector, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}

	return mask
}

func (rm *RegionManager) HasActiveSelection() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.active != ""
}

func rectangleToPoints(rect image.Rectangle) []image.Point {
	return []image.Point{
		{X: rect.Min.X, Y: rect.Min.Y},
		{X: rect.Max.X, Y: rect.Max.Y},
	}
}

func calculateBounds(points []image.Point) image.Rectangle {
	if len(points) == 0 {
		return image.Rectangle{}
	}

	minX, minY := points[0].X, points[0].Y
	maxX, maxY := points[0].X, points[0].Y

	for _, point := range points {
		if point.X < minX {
			minX = point.X
		}
		if point.X > maxX {
			maxX = point.X
		}
		if point.Y < minY {
			minY = point.Y
		}
		if point.Y > maxY {
			maxY = point.Y
		}
	}

	return image.Rect(minX, minY, maxX, maxY)
}

func (rm *RegionManager) IsPointInSelection(point image.Point) bool {
	selection := rm.GetActiveSelection()
	if selection == nil {
		return false
	}

	switch selection.Type {
	case SelectionRectangle:
		return point.In(selection.Bounds)
	case SelectionFreehand:
		return isPointInPolygon(point, selection.Points)
	}

	return false
}

func isPointInPolygon(point image.Point, polygon []image.Point) bool {
	if len(polygon) < 3 {
		return false
	}

	x, y := float64(point.X), float64(point.Y)
	inside := false

	j := len(polygon) - 1
	for i := 0; i < len(polygon); i++ {
		xi, yi := float64(polygon[i].X), float64(polygon[i].Y)
		xj, yj := float64(polygon[j].X), float64(polygon[j].Y)

		if ((yi > y) != (yj > y)) && (x < (xj-xi)*(y-yi)/(yj-yi)+xi) {
			inside = !inside
		}
		j = i
	}

	return inside
}
