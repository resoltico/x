// Author: Ervins Strauhmanis
// License: MIT

package morphology

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
	"advanced-image-processing/internal/transforms"
)

// ErosionTransform implements morphological erosion
type ErosionTransform struct{}

// NewErosionTransform creates a new erosion transformation
func NewErosionTransform() *ErosionTransform {
	return &ErosionTransform{}
}

// Apply applies morphological erosion to the input image
func (e *ErosionTransform) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	// Apply erosion
	output := gocv.NewMat()
	gocv.Erode(input, &output, kernel, image.Pt(-1, -1), iterations, gocv.BorderConstant, image.Black)

	return output, nil
}

// GetDefaultParams returns the default parameters for erosion
func (e *ErosionTransform) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 3.0,
		"iterations":  1.0,
	}
}

// GetName returns the name of this transformation
func (e *ErosionTransform) GetName() string {
	return "Erosion"
}

// GetDescription returns a description of this transformation
func (e *ErosionTransform) GetDescription() string {
	return "Applies morphological erosion to reduce white regions in binary images"
}

// Validate validates the parameters
func (e *ErosionTransform) Validate(params map[string]interface{}) error {
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
func (e *ErosionTransform) GetParameterInfo() []transforms.ParameterInfo {
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
			Description: "Number of times erosion is applied",
		},
	}
}