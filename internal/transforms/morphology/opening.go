// Author: Ervins Strauhmanis
// License: MIT

package morphology

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
	"advanced-image-processing/internal/transforms"
)

// OpeningTransform implements morphological opening (erosion followed by dilation)
type OpeningTransform struct{}

// NewOpeningTransform creates a new opening transformation
func NewOpeningTransform() *OpeningTransform {
	return &OpeningTransform{}
}

// Apply applies morphological opening to the input image
func (o *OpeningTransform) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	// Apply opening (erosion followed by dilation)
	// Updated function signature - removed extra parameters
	output := gocv.NewMat()
	for i := 0; i < iterations; i++ {
		if i == 0 {
			gocv.MorphologyEx(input, &output, gocv.MorphOpen, kernel)
		} else {
			temp := output.Clone()
			gocv.MorphologyEx(temp, &output, gocv.MorphOpen, kernel)
			temp.Close()
		}
	}

	return output, nil
}

// GetDefaultParams returns the default parameters for opening
func (o *OpeningTransform) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 3.0,
		"iterations":  1.0,
	}
}

// GetName returns the name of this transformation
func (o *OpeningTransform) GetName() string {
	return "Opening"
}

// GetDescription returns a description of this transformation
func (o *OpeningTransform) GetDescription() string {
	return "Applies morphological opening (erosion followed by dilation) to remove small objects"
}

// Validate validates the parameters
func (o *OpeningTransform) Validate(params map[string]interface{}) error {
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
func (o *OpeningTransform) GetParameterInfo() []transforms.ParameterInfo {
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
			Description: "Number of times opening is applied",
		},
	}
}