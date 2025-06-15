// Filter algorithms for noise reduction and enhancement
package algorithms

import (
	"fmt"

	"gocv.io/x/gocv"
)

// GaussianFilter implements Gaussian blur filter
type GaussianFilter struct{}

// NewGaussianFilter creates a new Gaussian filter algorithm
func NewGaussianFilter() *GaussianFilter {
	return &GaussianFilter{}
}

func (g *GaussianFilter) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	// Ensure kernel size is odd
	if kernelSize%2 == 0 {
		kernelSize++
	}

	// Apply Gaussian blur
	output := gocv.NewMat()
	gocv.GaussianBlur(input, &output, gocv.NewPoint(kernelSize, kernelSize), sigmaX, sigmaY, gocv.BorderDefault)

	return output, nil
}

func (g *GaussianFilter) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 5.0,
		"sigma_x":     1.0,
		"sigma_y":     1.0,
	}
}

func (g *GaussianFilter) GetName() string {
	return "Gaussian Filter"
}

func (g *GaussianFilter) GetDescription() string {
	return "Gaussian blur for general noise reduction"
}

func (g *GaussianFilter) Validate(params map[string]interface{}) error {
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 21 {
				return fmt.Errorf("kernel_size must be between 3 and 21")
			}
		}
	}
	
	if val, ok := params["sigma_x"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0.1 || v > 10.0 {
				return fmt.Errorf("sigma_x must be between 0.1 and 10.0")
			}
		}
	}
	
	if val, ok := params["sigma_y"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0.1 || v > 10.0 {
				return fmt.Errorf("sigma_y must be between 0.1 and 10.0")
			}
		}
	}
	
	return nil
}

func (g *GaussianFilter) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "kernel_size",
			Type:        "int",
			Min:         3.0,
			Max:         21.0,
			Default:     5.0,
			Description: "Size of the Gaussian kernel (must be odd)",
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

// MedianFilter implements median filter
type MedianFilter struct{}

// NewMedianFilter creates a new median filter algorithm
func NewMedianFilter() *MedianFilter {
	return &MedianFilter{}
}

func (m *MedianFilter) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	// Ensure kernel size is odd
	if kernelSize%2 == 0 {
		kernelSize++
	}

	// Apply median blur
	output := gocv.NewMat()
	gocv.MedianBlur(input, &output, kernelSize)

	return output, nil
}

func (m *MedianFilter) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 5.0,
	}
}

func (m *MedianFilter) GetName() string {
	return "Median Filter"
}

func (m *MedianFilter) GetDescription() string {
	return "Median filter to remove salt-and-pepper noise"
}

func (m *MedianFilter) Validate(params map[string]interface{}) error {
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 15 {
				return fmt.Errorf("kernel_size must be between 3 and 15")
			}
		}
	}
	
	return nil
}

func (m *MedianFilter) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "kernel_size",
			Type:        "int",
			Min:         3.0,
			Max:         15.0,
			Default:     5.0,
			Description: "Size of the median filter kernel (must be odd)",
		},
	}
}

// BilateralFilter implements bilateral filter
type BilateralFilter struct{}

// NewBilateralFilter creates a new bilateral filter algorithm
func NewBilateralFilter() *BilateralFilter {
	return &BilateralFilter{}
}

func (b *BilateralFilter) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// Get parameters
	d := 9
	if val, ok := params["d"]; ok {
		if v, ok := val.(float64); ok {
			d = int(v)
		}
	}

	sigmaColor := 75.0
	if val, ok := params["sigma_color"]; ok {
		if v, ok := val.(float64); ok {
			sigmaColor = v
		}
	}

	sigmaSpace := 75.0
	if val, ok := params["sigma_space"]; ok {
		if v, ok := val.(float64); ok {
			sigmaSpace = v
		}
	}

	// Apply bilateral filter
	output := gocv.NewMat()
	gocv.BilateralFilter(input, &output, d, sigmaColor, sigmaSpace)

	return output, nil
}

func (b *BilateralFilter) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"d":           9.0,
		"sigma_color": 75.0,
		"sigma_space": 75.0,
	}
}

func (b *BilateralFilter) GetName() string {
	return "Bilateral Filter"
}

func (b *BilateralFilter) GetDescription() string {
	return "Bilateral filter for edge-preserving smoothing"
}

func (b *BilateralFilter) Validate(params map[string]interface{}) error {
	if val, ok := params["d"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 15 {
				return fmt.Errorf("d must be between 3 and 15")
			}
		}
	}
	
	if val, ok := params["sigma_color"]; ok {
		if v, ok := val.(float64); ok {
			if v < 10.0 || v > 200.0 {
				return fmt.Errorf("sigma_color must be between 10.0 and 200.0")
			}
		}
	}
	
	if val, ok := params["sigma_space"]; ok {
		if v, ok := val.(float64); ok {
			if v < 10.0 || v > 200.0 {
				return fmt.Errorf("sigma_space must be between 10.0 and 200.0")
			}
		}
	}
	
	return nil
}

func (b *BilateralFilter) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "d",
			Type:        "int",
			Min:         3.0,
			Max:         15.0,
			Default:     9.0,
			Description: "Diameter of each pixel neighborhood",
		},
		{
			Name:        "sigma_color",
			Type:        "float",
			Min:         10.0,
			Max:         200.0,
			Default:     75.0,
			Description: "Filter sigma in the color space",
		},
		{
			Name:        "sigma_space",
			Type:        "float",
			Min:         10.0,
			Max:         200.0,
			Default:     75.0,
			Description: "Filter sigma in the coordinate space",
		},
	}
}