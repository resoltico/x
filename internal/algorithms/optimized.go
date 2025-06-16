// Optimized algorithms using GoCV standard APIs
package algorithms

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// AdaptiveThreshold using GoCV built-in adaptive threshold
type AdaptiveThreshold struct{}

func NewAdaptiveThreshold() *AdaptiveThreshold {
	return &AdaptiveThreshold{}
}

func (a *AdaptiveThreshold) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	gray := a.ensureGrayscale(input)
	defer func() {
		if gray.Ptr() != input.Ptr() {
			gray.Close()
		}
	}()

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

	// Ensure block size is odd
	if blockSize%2 == 0 {
		blockSize++
	}

	C := 2.0
	if val, ok := params["C"]; ok {
		if v, ok := val.(float64); ok {
			C = v
		}
	}

	method := gocv.AdaptiveThresholdMean
	if val, ok := params["method"]; ok {
		if v, ok := val.(string); ok && v == "gaussian" {
			method = gocv.AdaptiveThresholdGaussian
		}
	}

	// Apply adaptive threshold using GoCV built-in
	output := gocv.NewMat()
	gocv.AdaptiveThreshold(gray, &output, maxValue, method, gocv.ThresholdBinary, blockSize, C)

	return output, nil
}

func (a *AdaptiveThreshold) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"max_value":  255.0,
		"block_size": 11.0,
		"C":          2.0,
		"method":     "mean",
	}
}

func (a *AdaptiveThreshold) GetName() string {
	return "Adaptive Threshold"
}

func (a *AdaptiveThreshold) GetDescription() string {
	return "Adaptive thresholding using GoCV optimized implementation"
}

func (a *AdaptiveThreshold) Validate(params map[string]interface{}) error {
	if val, ok := params["block_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 101 {
				return fmt.Errorf("block_size must be between 3 and 101")
			}
		}
	}
	return nil
}

func (a *AdaptiveThreshold) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "max_value",
			Type:        "float",
			Min:         0.0,
			Max:         255.0,
			Default:     255.0,
			Description: "Maximum output value",
		},
		{
			Name:        "block_size",
			Type:        "int",
			Min:         3.0,
			Max:         101.0,
			Default:     11.0,
			Description: "Size of neighborhood area",
		},
		{
			Name:        "C",
			Type:        "float",
			Min:         -10.0,
			Max:         10.0,
			Default:     2.0,
			Description: "Constant subtracted from mean",
		},
		{
			Name:        "method",
			Type:        "enum",
			Default:     "mean",
			Description: "Adaptive method",
			Options:     []string{"mean", "gaussian"},
		},
	}
}

func (a *AdaptiveThreshold) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}
	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

// MultiOtsu using GoCV built-in threshold
type MultiOtsu struct{}

func NewMultiOtsu() *MultiOtsu {
	return &MultiOtsu{}
}

func (m *MultiOtsu) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	gray := m.ensureGrayscale(input)
	defer func() {
		if gray.Ptr() != input.Ptr() {
			gray.Close()
		}
	}()

	maxValue := 255.0
	if val, ok := params["max_value"]; ok {
		if v, ok := val.(float64); ok {
			maxValue = v
		}
	}

	// Use GoCV's built-in Otsu threshold
	output := gocv.NewMat()
	gocv.Threshold(gray, &output, 0, maxValue, gocv.ThresholdBinary+gocv.ThresholdOtsu)

	return output, nil
}

func (m *MultiOtsu) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"max_value": 255.0,
	}
}

func (m *MultiOtsu) GetName() string {
	return "Otsu Threshold"
}

func (m *MultiOtsu) GetDescription() string {
	return "Otsu thresholding using GoCV optimized implementation"
}

func (m *MultiOtsu) Validate(params map[string]interface{}) error {
	return nil
}

func (m *MultiOtsu) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "max_value",
			Type:        "float",
			Min:         0.0,
			Max:         255.0,
			Default:     255.0,
			Description: "Maximum output value",
		},
	}
}

func (m *MultiOtsu) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}
	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

// Utility functions
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}