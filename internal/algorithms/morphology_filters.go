// Morphological operations and filters
package algorithms

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Erosion
type Erosion struct{}

func NewErosion() *Erosion {
	return &Erosion{}
}

func (e *Erosion) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

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

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	output := gocv.NewMat()
	gocv.Erode(input, &output, kernel)

	for i := 1; i < iterations; i++ {
		temp := gocv.NewMat()
		gocv.Erode(output, &temp, kernel)
		output.Close()
		output = temp
	}

	return output, nil
}

func (e *Erosion) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 3.0,
		"iterations":  1.0,
	}
}

func (e *Erosion) GetName() string {
	return "Erosion"
}

func (e *Erosion) GetDescription() string {
	return "Morphological erosion to remove small noise"
}

func (e *Erosion) Validate(params map[string]interface{}) error {
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 1 || v > 15 {
				return fmt.Errorf("kernel_size must be between 1 and 15")
			}
		}
	}
	return nil
}

func (e *Erosion) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "kernel_size",
			Type:        "int",
			Min:         1.0,
			Max:         15.0,
			Default:     3.0,
			Description: "Size of the morphological kernel",
		},
		{
			Name:        "iterations",
			Type:        "int",
			Min:         1.0,
			Max:         10.0,
			Default:     1.0,
			Description: "Number of erosion iterations",
		},
	}
}

// Dilation
type Dilation struct{}

func NewDilation() *Dilation {
	return &Dilation{}
}

func (d *Dilation) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

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

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	output := gocv.NewMat()
	gocv.Dilate(input, &output, kernel)

	for i := 1; i < iterations; i++ {
		temp := gocv.NewMat()
		gocv.Dilate(output, &temp, kernel)
		output.Close()
		output = temp
	}

	return output, nil
}

func (d *Dilation) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 3.0,
		"iterations":  1.0,
	}
}

func (d *Dilation) GetName() string {
	return "Dilation"
}

func (d *Dilation) GetDescription() string {
	return "Morphological dilation to fill gaps in text"
}

func (d *Dilation) Validate(params map[string]interface{}) error {
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 1 || v > 15 {
				return fmt.Errorf("kernel_size must be between 1 and 15")
			}
		}
	}
	return nil
}

func (d *Dilation) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "kernel_size",
			Type:        "int",
			Min:         1.0,
			Max:         15.0,
			Default:     3.0,
			Description: "Size of the morphological kernel",
		},
		{
			Name:        "iterations",
			Type:        "int",
			Min:         1.0,
			Max:         10.0,
			Default:     1.0,
			Description: "Number of dilation iterations",
		},
	}
}

// Opening
type Opening struct{}

func NewOpening() *Opening {
	return &Opening{}
}

func (o *Opening) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	kernelSize := 3
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			kernelSize = int(v)
		}
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	output := gocv.NewMat()
	gocv.MorphologyEx(input, &output, gocv.MorphOpen, kernel)

	return output, nil
}

func (o *Opening) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 3.0,
	}
}

func (o *Opening) GetName() string {
	return "Opening"
}

func (o *Opening) GetDescription() string {
	return "Morphological opening to remove noise while preserving text"
}

func (o *Opening) Validate(params map[string]interface{}) error {
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 1 || v > 15 {
				return fmt.Errorf("kernel_size must be between 1 and 15")
			}
		}
	}
	return nil
}

func (o *Opening) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "kernel_size",
			Type:        "int",
			Min:         1.0,
			Max:         15.0,
			Default:     3.0,
			Description: "Size of the morphological kernel",
		},
	}
}

// Closing
type Closing struct{}

func NewClosing() *Closing {
	return &Closing{}
}

func (c *Closing) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	kernelSize := 3
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			kernelSize = int(v)
		}
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	output := gocv.NewMat()
	gocv.MorphologyEx(input, &output, gocv.MorphClose, kernel)

	return output, nil
}

func (c *Closing) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"kernel_size": 3.0,
	}
}

func (c *Closing) GetName() string {
	return "Closing"
}

func (c *Closing) GetDescription() string {
	return "Morphological closing to connect broken characters"
}

func (c *Closing) Validate(params map[string]interface{}) error {
	if val, ok := params["kernel_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 1 || v > 15 {
				return fmt.Errorf("kernel_size must be between 1 and 15")
			}
		}
	}
	return nil
}

func (c *Closing) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "kernel_size",
			Type:        "int",
			Min:         1.0,
			Max:         15.0,
			Default:     3.0,
			Description: "Size of the morphological kernel",
		},
	}
}

// GaussianFilter
type GaussianFilter struct{}

func NewGaussianFilter() *GaussianFilter {
	return &GaussianFilter{}
}

func (g *GaussianFilter) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

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

	output := gocv.NewMat()
	gocv.GaussianBlur(input, &output, image.Pt(kernelSize, kernelSize), sigmaX, sigmaY, gocv.BorderDefault)

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

// MedianFilter
type MedianFilter struct{}

func NewMedianFilter() *MedianFilter {
	return &MedianFilter{}
}

func (m *MedianFilter) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

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

// BilateralFilter
type BilateralFilter struct{}

func NewBilateralFilter() *BilateralFilter {
	return &BilateralFilter{}
}

func (b *BilateralFilter) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

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
