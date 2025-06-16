// Morphological operations algorithms
package algorithms

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Erosion implements morphological erosion
type Erosion struct{}

// NewErosion creates a new erosion algorithm
func NewErosion() *Erosion {
	return &Erosion{}
}

func (e *Erosion) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	// Create kernel
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	// Apply erosion
	output := gocv.NewMat()
	gocv.Erode(input, &output, kernel)

	// Apply multiple iterations if needed
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

	if val, ok := params["iterations"]; ok {
		if v, ok := val.(float64); ok {
			if v < 1 || v > 10 {
				return fmt.Errorf("iterations must be between 1 and 10")
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

// Dilation implements morphological dilation
type Dilation struct{}

// NewDilation creates a new dilation algorithm
func NewDilation() *Dilation {
	return &Dilation{}
}

func (d *Dilation) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	// Create kernel
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	// Apply dilation
	output := gocv.NewMat()
	gocv.Dilate(input, &output, kernel)

	// Apply multiple iterations if needed
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

	if val, ok := params["iterations"]; ok {
		if v, ok := val.(float64); ok {
			if v < 1 || v > 10 {
				return fmt.Errorf("iterations must be between 1 and 10")
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

// Opening implements morphological opening
type Opening struct{}

// NewOpening creates a new opening algorithm
func NewOpening() *Opening {
	return &Opening{}
}

func (o *Opening) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	// Create kernel
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	// Apply opening (erosion followed by dilation)
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

// Closing implements morphological closing
type Closing struct{}

// NewClosing creates a new closing algorithm
func NewClosing() *Closing {
	return &Closing{}
}

func (c *Closing) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
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

	// Create kernel
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(kernelSize, kernelSize))
	defer kernel.Close()

	// Apply closing (dilation followed by erosion)
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
