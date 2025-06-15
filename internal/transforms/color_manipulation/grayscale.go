// Author: Ervins Strauhmanis
// License: MIT

package color_manipulation

import (
	"fmt"

	"gocv.io/x/gocv"
	"advanced-image-processing/internal/transforms"
)

// GrayscaleTransform implements color to grayscale conversion
type GrayscaleTransform struct{}

// NewGrayscaleTransform creates a new grayscale transformation
func NewGrayscaleTransform() *GrayscaleTransform {
	return &GrayscaleTransform{}
}

// Apply converts the input image to grayscale
func (g *GrayscaleTransform) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// If already grayscale, return copy
	if input.Channels() == 1 {
		return input.Clone(), nil
	}

	// Convert to grayscale
	output := gocv.NewMat()
	gocv.CvtColor(input, &output, gocv.ColorBGRToGray)

	return output, nil
}

// GetDefaultParams returns the default parameters for grayscale conversion
func (g *GrayscaleTransform) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		// No parameters for basic grayscale conversion
	}
}

// GetName returns the name of this transformation
func (g *GrayscaleTransform) GetName() string {
	return "Grayscale"
}

// GetDescription returns a description of this transformation
func (g *GrayscaleTransform) GetDescription() string {
	return "Converts color images to grayscale"
}

// Validate validates the parameters
func (g *GrayscaleTransform) Validate(params map[string]interface{}) error {
	// No parameters to validate for basic grayscale conversion
	return nil
}

// GetParameterInfo returns parameter information
func (g *GrayscaleTransform) GetParameterInfo() []transforms.ParameterInfo {
	return []transforms.ParameterInfo{
		// No parameters for basic grayscale conversion
	}
}