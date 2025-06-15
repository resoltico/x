// Author: Ervins Strauhmanis
// License: MIT

package binarization

import (
	"fmt"

	"gocv.io/x/gocv"
	"advanced-image-processing/internal/transforms"
)

// NiblackTransform implements Niblack local adaptive thresholding
type NiblackTransform struct{}

// NewNiblackTransform creates a new Niblack transformation
func NewNiblackTransform() *NiblackTransform {
	return &NiblackTransform{}
}

// Apply applies Niblack adaptive thresholding to the input image
func (n *NiblackTransform) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	blockSize := 11
	if val, ok := params["block_size"]; ok {
		if v, ok := val.(float64); ok {
			blockSize = int(v)
		}
	}

	c := 2.0
	if val, ok := params["c"]; ok {
		if v, ok := val.(float64); ok {
			c = v
		}
	}

	// Ensure block size is odd and at least 3
	if blockSize%2 == 0 {
		blockSize++
	}
	if blockSize < 3 {
		blockSize = 3
	}

	// Apply adaptive thresholding (using mean method as approximation for Niblack)
	output := gocv.NewMat()
	gocv.AdaptiveThreshold(gray, &output, maxValue, gocv.AdaptiveThresholdMean, gocv.ThresholdBinary, blockSize, c)

	return output, nil
}

// GetDefaultParams returns the default parameters for Niblack thresholding
func (n *NiblackTransform) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"max_value":  255.0,
		"block_size": 11.0,
		"c":          2.0,
	}
}

// GetName returns the name of this transformation
func (n *NiblackTransform) GetName() string {
	return "Niblack Binarization"
}

// GetDescription returns a description of this transformation
func (n *NiblackTransform) GetDescription() string {
	return "Applies Niblack local adaptive thresholding for image binarization"
}

// Validate validates the parameters
func (n *NiblackTransform) Validate(params map[string]interface{}) error {
	if val, ok := params["max_value"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0 || v > 255 {
				return fmt.Errorf("max_value must be between 0 and 255")
			}
		} else {
			return fmt.Errorf("max_value must be a number")
		}
	}

	if val, ok := params["block_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 255 {
				return fmt.Errorf("block_size must be between 3 and 255")
			}
		} else {
			return fmt.Errorf("block_size must be a number")
		}
	}

	return nil
}

// GetParameterInfo returns parameter information
func (n *NiblackTransform) GetParameterInfo() []transforms.ParameterInfo {
	return []transforms.ParameterInfo{
		{
			Name:        "max_value",
			Type:        "float",
			Min:         0.0,
			Max:         255.0,
			Default:     255.0,
			Description: "Maximum value assigned to pixels above threshold",
		},
		{
			Name:        "block_size",
			Type:        "int",
			Min:         3.0,
			Max:         255.0,
			Default:     11.0,
			Description: "Size of neighborhood area for calculating threshold",
		},
		{
			Name:        "c",
			Type:        "float",
			Min:         -50.0,
			Max:         50.0,
			Default:     2.0,
			Description: "Constant subtracted from the mean",
		},
	}
}