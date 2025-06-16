// internal/layers/stack.go
// Fixed layers package for region-based processing
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
	fmt.Printf("DEBUG: ProcessLayers called with input empty=%v, rows=%d, cols=%d\n",
		input.Empty(), input.Rows(), input.Cols())

	if input.Empty() {
		fmt.Printf("DEBUG: Input is empty, returning NewMat\n")
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	layers := ls.GetLayers()
	fmt.Printf("DEBUG: Processing %d layers\n", len(layers))

	// If no layers, return clone of input - NEVER return zero-value Mat
	if len(layers) == 0 {
		fmt.Printf("DEBUG: No layers, cloning input\n")
		result := input.Clone()
		fmt.Printf("DEBUG: Clone result: empty=%v, rows=%d, cols=%d\n",
			result.Empty(), result.Rows(), result.Cols())

		if result.Empty() {
			fmt.Printf("DEBUG: Clone failed, creating new Mat and copying\n")
			// Fallback: create a proper Mat copy
			result = gocv.NewMatWithSize(input.Rows(), input.Cols(), input.Type())
			input.CopyTo(&result)
			fmt.Printf("DEBUG: After manual copy: empty=%v, rows=%d, cols=%d\n",
				result.Empty(), result.Rows(), result.Cols())
		}
		return result, nil
	}

	result := input.Clone()
	fmt.Printf("DEBUG: Initial clone for processing: empty=%v, rows=%d, cols=%d\n",
		result.Empty(), result.Rows(), result.Cols())

	for i, layer := range layers {
		if !layer.Enabled {
			fmt.Printf("DEBUG: Layer %d (%s) is disabled, skipping\n", i, layer.Name)
			continue
		}

		fmt.Printf("DEBUG: Processing layer %d: %s with algorithm %s\n", i, layer.Name, layer.Algorithm)

		// Apply algorithm to current result
		processed, err := algorithms.Apply(layer.Algorithm, result, layer.Parameters)
		if err != nil {
			fmt.Printf("DEBUG: Algorithm failed: %v\n", err)
			result.Close()
			return gocv.NewMat(), err
		}

		if processed.Empty() {
			fmt.Printf("DEBUG: Algorithm returned empty result, skipping\n")
			processed.Close()
			continue
		}

		fmt.Printf("DEBUG: Algorithm applied successfully, result size: %dx%d\n", processed.Cols(), processed.Rows())

		// Apply region mask if specified
		if layer.RegionID != "" && ls.regionManager != nil {
			if selection := ls.regionManager.GetSelection(layer.RegionID); selection != nil {
				mask := ls.regionManager.CreateMaskForSelection(selection, result.Cols(), result.Rows())
				if !mask.Empty() {
					fmt.Printf("DEBUG: Applying region mask for %s\n", layer.RegionID)
					// Apply mask to processed result
					maskedResult := gocv.NewMat()
					processed.CopyTo(&maskedResult)
					result.CopyToWithMask(&maskedResult, mask)
					processed.Close()
					processed = maskedResult
					mask.Close()
				}
			}
		}

		// Replace result with processed
		result.Close()
		result = processed

		fmt.Printf("DEBUG: Layer %d processing completed\n", i)
	}

	fmt.Printf("DEBUG: ProcessLayers returning: empty=%v, rows=%d, cols=%d\n",
		result.Empty(), result.Rows(), result.Cols())
	return result, nil
}
