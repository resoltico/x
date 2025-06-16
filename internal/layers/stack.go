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
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	result := input.Clone()
	layers := ls.GetLayers()

	fmt.Printf("DEBUG: Processing %d layers\n", len(layers))

	for i, layer := range layers {
		if !layer.Enabled {
			fmt.Printf("DEBUG: Layer %d (%s) is disabled, skipping\n", i, layer.Name)
			continue
		}

		fmt.Printf("DEBUG: Processing layer %d: %s with algorithm %s\n", i, layer.Name, layer.Algorithm)

		processed, err := ls.processLayer(result, layer)
		if err != nil {
			fmt.Printf("DEBUG: Layer processing failed: %v\n", err)
			if !processed.Empty() {
				processed.Close()
			}
			result.Close()
			return gocv.NewMat(), err
		}

		if processed.Empty() {
			fmt.Printf("DEBUG: Processed result is empty, skipping layer\n")
			continue
		}

		fmt.Printf("DEBUG: Layer processed successfully, blending...\n")

		// Blend the processed result
		blended := ls.blendLayers(result, processed, layer.BlendMode, layer.Opacity, layer.RegionID)
		result.Close()
		processed.Close()
		result = blended

		fmt.Printf("DEBUG: Layer %d blending completed\n", i)
	}

	return result, nil
}

// processLayer applies algorithm to image with optional region mask
func (ls *LayerStack) processLayer(input gocv.Mat, layer *Layer) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	fmt.Printf("DEBUG: Applying algorithm %s with params %+v\n", layer.Algorithm, layer.Parameters)

	// Apply algorithm using existing pipeline
	result, err := algorithms.Apply(layer.Algorithm, input, layer.Parameters)
	if err != nil {
		fmt.Printf("DEBUG: Algorithm application failed: %v\n", err)
		return gocv.NewMat(), err
	}

	if result.Empty() {
		fmt.Printf("DEBUG: Algorithm returned empty result\n")
		return gocv.NewMat(), fmt.Errorf("algorithm returned empty result")
	}

	fmt.Printf("DEBUG: Algorithm applied successfully, result size: %dx%d\n", result.Cols(), result.Rows())
	return result, nil
}

// blendLayers combines two images using blend mode and opacity
func (ls *LayerStack) blendLayers(base, overlay gocv.Mat, mode BlendMode, opacity float64, regionID string) gocv.Mat {
	if base.Empty() || overlay.Empty() {
		return base.Clone()
	}

	result := base.Clone()

	// Create mask if region specified
	var mask gocv.Mat
	hasMask := false
	if regionID != "" && ls.regionManager != nil {
		if selection := ls.regionManager.GetSelection(regionID); selection != nil {
			mask = ls.regionManager.CreateMaskForSelection(selection, base.Cols(), base.Rows())
			hasMask = !mask.Empty()
		}
	}

	// Ensure mask cleanup
	defer func() {
		if hasMask && !mask.Empty() {
			mask.Close()
		}
	}()

	// Apply blend mode with opacity
	switch mode {
	case BlendNormal:
		ls.blendNormal(result, overlay, opacity, mask, hasMask)
	case BlendOverlay:
		ls.blendOverlay(result, overlay, opacity, mask, hasMask)
		// Add other blend modes as needed
	}

	return result
}

// blendNormal performs normal alpha blending
func (ls *LayerStack) blendNormal(base, overlay gocv.Mat, opacity float64, mask gocv.Mat, hasMask bool) {
	if base.Empty() || overlay.Empty() {
		fmt.Printf("DEBUG: Blend called with empty mats: base=%v, overlay=%v\n", base.Empty(), overlay.Empty())
		return
	}

	fmt.Printf("DEBUG: Blending with opacity=%.2f, hasMask=%v\n", opacity, hasMask)

	if !hasMask {
		// Global blend - when opacity is 1.0, completely replace with overlay
		if opacity >= 1.0 {
			fmt.Printf("DEBUG: Full replacement (opacity=1.0)\n")
			overlay.CopyTo(&base)
		} else {
			fmt.Printf("DEBUG: Alpha blending with opacity=%.2f\n", opacity)
			gocv.AddWeighted(base, 1.0-opacity, overlay, opacity, 0, &base)
		}
	} else {
		// Masked blend using GoCV's optimized operations
		temp := gocv.NewMat()
		defer temp.Close()

		if opacity >= 1.0 {
			fmt.Printf("DEBUG: Full masked replacement\n")
			overlay.CopyToWithMask(&base, mask)
		} else {
			fmt.Printf("DEBUG: Alpha masked blending\n")
			gocv.AddWeighted(base, 1.0-opacity, overlay, opacity, 0, &temp)
			temp.CopyToWithMask(&base, mask)
		}
	}
}

// blendOverlay performs overlay blending (simplified)
func (ls *LayerStack) blendOverlay(base, overlay gocv.Mat, opacity float64, mask gocv.Mat, hasMask bool) {
	if base.Empty() || overlay.Empty() {
		return
	}

	// Use GoCV's built-in operations for performance
	// This is a simplified overlay - for production use proper overlay math
	ls.blendNormal(base, overlay, opacity*0.5, mask, hasMask)
}
