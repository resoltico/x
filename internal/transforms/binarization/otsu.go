// Author: Ervins Strauhmanis
// License: MIT

package binarization

import (
	"fmt"

	"gocv.io/x/gocv"
	"advanced-image-processing/internal/transforms"
)

// OtsuTransform implements Otsu's automatic threshold selection
type OtsuTransform struct{}

// NewOtsuTransform creates a new Otsu transformation
func NewOtsuTransform() *OtsuTransform {
	return &OtsuTransform{}
}

// Apply applies Otsu thresholding to the input image
func (o *OtsuTransform) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// Convert to grayscale if needed
	var gray gocv.Mat
	if input.Channels() > 1 {
		gray = gocv.NewMat()
		gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
		defer gray.Close()
	} else {
		gray = input
	}

	// Get parameters
	maxValue := 255.0
	if val, ok := params["max_value"]; ok {
		if v, ok := val.(float64); ok {
			maxValue = v
		}
	}

	// Apply Otsu thresholding
	// Convert float64 to float32 for gocv function
	output := gocv.NewMat()
	gocv.Threshold(gray, &output, 0, float32(maxValue), gocv.ThresholdBinary+gocv.ThresholdOtsu)

	return output, nil
}

// GetDefaultParams returns the default parameters for Otsu thresholding
func (o *OtsuTransform) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"max_value": 255.0,
	}
}

// GetName returns the name of this transformation
func (o *OtsuTransform) GetName() string {
	return "Otsu Binarization"
}

// GetDescription returns a description of this transformation
func (o *OtsuTransform) GetDescription() string {
	return "Applies Otsu's automatic threshold selection for image binarization"
}

// Validate validates the parameters
func (o *OtsuTransform) Validate(params map[string]interface{}) error {
	if val, ok := params["max_value"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0 || v > 255 {
				return fmt.Errorf("max_value must be between 0 and 255")
			}
		} else {
			return fmt.Errorf("max_value must be a number")
		}
	}
	return nil
}

// GetParameterInfo returns parameter information
func (o *OtsuTransform) GetParameterInfo() []transforms.ParameterInfo {
	return []transforms.ParameterInfo{
		{
			Name:        "max_value",
			Type:        "float",
			Min:         0.0,
			Max:         255.0,
			Default:     255.0,
			Description: "Maximum value assigned to pixels above threshold",
		},
	}
}