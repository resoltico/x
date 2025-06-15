// Author: Ervins Strauhmanis
// License: MIT

package noise_reduction

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
	"advanced-image-processing/internal/transforms"
)

// GaussianTransform implements Gaussian blur for noise reduction
type GaussianTransform struct{}

// NewGaussianTransform creates a new Gaussian blur transformation
func NewGaussianTransform() *GaussianTransform {
	return &GaussianTransform{}
}

// Apply applies Gaussian blur to the input image
func (g *GaussianTransform) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// Get parameters
	kernelSize := 5
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			kernelSize = int(v)
		}
	}

	sigmaX := 1.0
	if val, ok := params["sigma_x"]; ok {
		if v, ok := val.(float64); ok {
			sigmaX = v
		}
	}

	sigmaY := 1.0
	if val, ok := params["sigma_y"]; ok {
		if v, ok := val.(float64); ok {
			sigmaY = v
		}
	}

	// Ensure kernel size is odd and at least 3
	if kernelSize%2 == 0 {
		kernelSize++
	}
	if kernelSize < 3 {
		kernelSize = 3
	}

	// Apply Gaussian blur
	output := gocv.NewMat()
	gocv.GaussianBlur(input, &output, image.Pt(kernelSize, kernelSize), sigmaX, sigmaY, gocv.BorderDefault)

	return output, nil
}

// GetDefaultParams returns the default parameters for Gaussian blur
func (g *GaussianTransform) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 5.0,
		"sigma_x":     1.0,
		"sigma_y":     1.0,
	}
}

// GetName returns the name of this transformation
func (g *GaussianTransform) GetName() string {
	return "Gaussian Blur"
}

// GetDescription returns a description of this transformation
func (g *GaussianTransform) GetDescription() string {
	return "Applies Gaussian blur to reduce noise in images"
}

// Validate validates the parameters
func (g *GaussianTransform) Validate(params map[string]interface{}) error {
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 31 {
				return fmt.Errorf("kernel_size must be between 3 and 31")
			}
		} else {
			return fmt.Errorf("kernel_size must be a number")
		}
	}

	if val, ok := params["sigma_x"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0.1 || v > 10 {
				return fmt.Errorf("sigma_x must be between 0.1 and 10")
			}
		} else {
			return fmt.Errorf("sigma_x must be a number")
		}
	}

	if val, ok := params["sigma_y"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0.1 || v > 10 {
				return fmt.Errorf("sigma_y must be between 0.1 and 10")
			}
		} else {
			return fmt.Errorf("sigma_y must be a number")
		}
	}

	return nil
}

// GetParameterInfo returns parameter information
func (g *GaussianTransform) GetParameterInfo() []transforms.ParameterInfo {
	return []transforms.ParameterInfo{
		{
			Name:        "kernel_size",
			Type:        "int",
			Min:         3.0,
			Max:         31.0,
			Default:     5.0,
			Description: "Size of the Gaussian kernel",
		},
		{
			Name:        "sigma_x",
			Type:        "float",
			Min:         0.1,
			Max:         10.0,
			Default:     1.0,
			Description: "Standard deviation in X direction",
		},
		{
			Name:        "sigma_y",
			Type:        "float",
			Min:         0.1,
			Max:         10.0,
			Default:     1.0,
			Description: "Standard deviation in Y direction",
		},
	}
}