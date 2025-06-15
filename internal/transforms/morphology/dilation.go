// Author: Ervins Strauhmanis
// License: MIT

package morphology

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
	"advanced-image-processing/internal/transforms"
)

// DilationTransform implements morphological dilation
type DilationTransform struct{}

// NewDilationTransform creates a new dilation transformation
func NewDilationTransform() *DilationTransform {
	return &DilationTransform{}
}

// Apply applies morphological dilation to the input image
func (d *DilationTransform) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// Get parameters
	kernelSize := 3
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			kernelSize = int(v)
		}
	}

	iterations := 1
	if val, ok := params["iterations"]; ok {
		if v, ok := val.(float64); ok {
			iterations = int(v)
		}
	}

	// Ensure kernel size is odd and at least 3
	if kernelSize%2 == 0 {
		kernelSize++
	}
	if kernelSize < 3 {
		kernelSize = 3
	}

	// Create structuring element
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	// Apply dilation
	output := gocv.NewMat()
	gocv.Dilate(input, &output, kernel, image.Pt(-1, -1), iterations, gocv.BorderConstant, image.Black)

	return output, nil
}

// GetDefaultParams returns the default parameters for dilation
func (d *DilationTransform) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 3.0,
		"iterations":  1.0,
	}
}

// GetName returns the name of this transformation
func (d *DilationTransform) GetName() string {
	return "Dilation"
}

// GetDescription returns a description of this transformation
func (d *DilationTransform) GetDescription() string {
	return "Applies morphological dilation to expand white regions in binary images"
}

// Validate validates the parameters
func (d *DilationTransform) Validate(params map[string]interface{}) error {
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 31 {
				return fmt.Errorf("kernel_size must be between 3 and 31")
			}
		} else {
			return fmt.Errorf("kernel_size must be a number")
		}
	}

	if val, ok := params["iterations"]; ok {
		if v, ok := val.(float64); ok {
			if v < 1 || v > 10 {
				return fmt.Errorf("iterations must be between 1 and 10")
			}
		} else {
			return fmt.Errorf("iterations must be a number")
		}
	}

	return nil
}

// GetParameterInfo returns parameter information
func (d *DilationTransform) GetParameterInfo() []transforms.ParameterInfo {
	return []transforms.ParameterInfo{
		{
			Name:        "kernel_size",
			Type:        "int",
			Min:         3.0,
			Max:         31.0,
			Default:     3.0,
			Description: "Size of the morphological kernel",
		},
		{
			Name:        "iterations",
			Type:        "int",
			Min:         1.0,
			Max:         10.0,
			Default:     1.0,
			Description: "Number of times dilation is applied",
		},
	}
}