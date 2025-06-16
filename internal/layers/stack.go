// Layers package for region-based processing with proper imports
package layers

import (
	"fmt"
	"sync"

	"gocv.io/x/gocv"

	"advanced-image-processing/internal/algorithms"
)

// Layer represents a processing layer with optional region mask
type Layer struct {
	ID         string
	Name       string
	Algorithm  string
	Parameters map[string]interface{}
	RegionID   string // Optional region selection ID
	Enabled    bool
	BlendMode  BlendMode
	Opacity    float64 // 0.0 to 1.0
}

// BlendMode defines how layers combine
type BlendMode int

const (
	BlendNormal BlendMode = iota
	BlendOverlay
	BlendMultiply
	BlendScreen
)

// RegionManager interface to avoid import cycle
type RegionManager interface {
	GetSelection(id string) Selection
	CreateMaskForSelection(selection Selection, width, height int) gocv.Mat
}

// Selection interface to avoid import cycle
type Selection interface {
	GetID() string
	GetType() int
	GetPoints() []Point
	GetBounds() Rectangle
	IsActive() bool
}

type Point struct {
	X, Y int
}

type Rectangle struct {
	Min, Max Point
}

// LayerStack manages multiple processing layers
type LayerStack struct {
	mu            sync.RWMutex
	layers        []*Layer
	regionManager RegionManager
	nextID        int
}

func NewLayerStack(regionManager RegionManager) *LayerStack {
	return &LayerStack{
		layers:        make([]*Layer, 0),
		regionManager: regionManager,
		nextID:        1,
	}
}

// AddLayer adds a processing layer
func (ls *LayerStack) AddLayer(name, algorithm string, params map[string]interface{}, regionID string) string {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	id := fmt.Sprintf("layer_%d", ls.nextID)
	ls.nextID++

	layer := &Layer{
		ID:         id,
		Name:       name,
		Algorithm:  algorithm,
		Parameters: params,
		RegionID:   regionID,
		Enabled:    true,
		BlendMode:  BlendNormal,
		Opacity:    1.0,
	}

	ls.layers = append(ls.layers, layer)
	return id
}

// GetLayers returns all layers
func (ls *LayerStack) GetLayers() []*Layer {
	ls.mu.RLock()
	defer ls.mu.RUnlock()

	result := make([]*Layer, len(ls.layers))
	copy(result, ls.layers)
	return result
}

// ProcessLayers applies all enabled layers to input image
func (ls *LayerStack) ProcessLayers(input gocv.Mat) (gocv.Mat, error) {
	result := input.Clone()

	for _, layer := range ls.GetLayers() {
		if !layer.Enabled {
			continue
		}

		processed, err := ls.processLayer(result, layer)
		if err != nil {
			processed.Close()
			result.Close()
			return gocv.NewMat(), err
		}

		// Blend the processed result
		blended := ls.blendLayers(result, processed, layer.BlendMode, layer.Opacity, layer.RegionID)
		result.Close()
		processed.Close()
		result = blended
	}

	return result, nil
}

// processLayer applies algorithm to image with optional region mask
func (ls *LayerStack) processLayer(input gocv.Mat, layer *Layer) (gocv.Mat, error) {
	// Apply algorithm using existing pipeline
	return algorithms.Apply(layer.Algorithm, input, layer.Parameters)
}

// blendLayers combines two images using blend mode and opacity
func (ls *LayerStack) blendLayers(base, overlay gocv.Mat, mode BlendMode, opacity float64, regionID string) gocv.Mat {
	result := base.Clone()

	// Create mask if region specified
	var mask gocv.Mat
	if regionID != "" && ls.regionManager != nil {
		if selection := ls.regionManager.GetSelection(regionID); selection != nil {
			mask = ls.regionManager.CreateMaskForSelection(selection, base.Cols(), base.Rows())
			defer mask.Close()
		}
	}

	// Apply blend mode with opacity
	switch mode {
	case BlendNormal:
		ls.blendNormal(result, overlay, opacity, mask)
	case BlendOverlay:
		ls.blendOverlay(result, overlay, opacity, mask)
		// Add other blend modes as needed
	}

	return result
}

// blendNormal performs normal alpha blending
func (ls *LayerStack) blendNormal(base, overlay gocv.Mat, opacity float64, mask gocv.Mat) {
	if mask.Empty() {
		// Global blend
		gocv.AddWeighted(base, 1.0-opacity, overlay, opacity, 0, &base)
	} else {
		// Masked blend using GoCV's optimized operations
		temp := gocv.NewMat()
		defer temp.Close()

		gocv.AddWeighted(base, 1.0-opacity, overlay, opacity, 0, &temp)
		temp.CopyToWithMask(&base, mask)
	}
}

// blendOverlay performs overlay blending (simplified)
func (ls *LayerStack) blendOverlay(base, overlay gocv.Mat, opacity float64, mask gocv.Mat) {
	// Use GoCV's built-in operations for performance
	// This is a simplified overlay - for production use proper overlay math
	ls.blendNormal(base, overlay, opacity*0.5, mask)
}
