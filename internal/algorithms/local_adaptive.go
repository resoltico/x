// True Local Adaptive Algorithm implementations based on latest research
package algorithms

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"
)

// TrueNiblack implements the true Niblack algorithm with proper local statistics
type TrueNiblack struct{}

// NewTrueNiblack creates a new true Niblack algorithm
func NewTrueNiblack() *TrueNiblack {
	return &TrueNiblack{}
}

func (n *TrueNiblack) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// Convert to grayscale if needed
	gray := n.ensureGrayscale(input)
	defer func() {
		if gray.Ptr() != input.Ptr() {
			gray.Close()
		}
	}()

	// Get parameters
	windowSize := 15
	if val, ok := params["window_size"]; ok {
		if v, ok := val.(float64); ok {
			windowSize = int(v)
		}
	}

	k := -0.2 // Niblack parameter
	if val, ok := params["k"]; ok {
		if v, ok := val.(float64); ok {
			k = v
		}
	}

	// Ensure window size is odd
	if windowSize%2 == 0 {
		windowSize++
	}

	// Apply true Niblack algorithm
	output := n.applyTrueNiblack(gray, windowSize, k)
	return output, nil
}

func (n *TrueNiblack) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"window_size": 15.0,
		"k":           -0.2,
	}
}

func (n *TrueNiblack) GetName() string {
	return "True Niblack"
}

func (n *TrueNiblack) GetDescription() string {
	return "True Niblack local adaptive thresholding with proper local mean and standard deviation"
}

func (n *TrueNiblack) Validate(params map[string]interface{}) error {
	if val, ok := params["window_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 101 {
				return fmt.Errorf("window_size must be between 3 and 101")
			}
		}
	}
	
	if val, ok := params["k"]; ok {
		if v, ok := val.(float64); ok {
			if v < -1.0 || v > 1.0 {
				return fmt.Errorf("k must be between -1.0 and 1.0")
			}
		}
	}
	
	return nil
}

func (n *TrueNiblack) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "window_size",
			Type:        "int",
			Min:         3.0,
			Max:         101.0,
			Default:     15.0,
			Description: "Local window size for statistics calculation",
		},
		{
			Name:        "k",
			Type:        "float",
			Min:         -1.0,
			Max:         1.0,
			Default:     -0.2,
			Description: "Niblack parameter (negative values preserve more text)",
		},
	}
}

func (n *TrueNiblack) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}
	
	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (n *TrueNiblack) applyTrueNiblack(gray gocv.Mat, windowSize int, k float64) gocv.Mat {
	height := gray.Rows()
	width := gray.Cols()
	output := gocv.NewMat()
	gray.CopyTo(&output)
	
	halfWindow := windowSize / 2
	
	// Pre-calculate integral images for efficiency
	integralSum := n.calculateIntegralImage(gray)
	defer integralSum.Close()
	
	integralSumSq := n.calculateIntegralImageSquared(gray)
	defer integralSumSq.Close()
	
	// Process each pixel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate window bounds
			x1 := max(0, x-halfWindow)
			y1 := max(0, y-halfWindow)
			x2 := min(width-1, x+halfWindow)
			y2 := min(height-1, y+halfWindow)
			
			// Calculate local mean and standard deviation using integral images
			mean, stddev := n.calculateLocalStats(integralSum, integralSumSq, x1, y1, x2, y2)
			
			// Apply Niblack formula: T = mean + k * stddev
			threshold := mean + k*stddev
			
			// Apply threshold
			intensity := float64(gray.GetUCharAt(y, x))
			if intensity <= threshold {
				output.SetUCharAt(y, x, 0)
			} else {
				output.SetUCharAt(y, x, 255)
			}
		}
	}
	
	return output
}

func (n *TrueNiblack) calculateIntegralImage(gray gocv.Mat) gocv.Mat {
	integral := gocv.NewMatWithSize(gray.Rows(), gray.Cols(), gocv.MatTypeCV64F)
	
	for y := 0; y < gray.Rows(); y++ {
		for x := 0; x < gray.Cols(); x++ {
			val := float64(gray.GetUCharAt(y, x))
			
			sum := val
			if x > 0 {
				sum += integral.GetDoubleAt(y, x-1)
			}
			if y > 0 {
				sum += integral.GetDoubleAt(y-1, x)
			}
			if x > 0 && y > 0 {
				sum -= integral.GetDoubleAt(y-1, x-1)
			}
			
			integral.SetDoubleAt(y, x, sum)
		}
	}
	
	return integral
}

func (n *TrueNiblack) calculateIntegralImageSquared(gray gocv.Mat) gocv.Mat {
	integral := gocv.NewMatWithSize(gray.Rows(), gray.Cols(), gocv.MatTypeCV64F)
	
	for y := 0; y < gray.Rows(); y++ {
		for x := 0; x < gray.Cols(); x++ {
			val := float64(gray.GetUCharAt(y, x))
			valSq := val * val
			
			sum := valSq
			if x > 0 {
				sum += integral.GetDoubleAt(y, x-1)
			}
			if y > 0 {
				sum += integral.GetDoubleAt(y-1, x)
			}
			if x > 0 && y > 0 {
				sum -= integral.GetDoubleAt(y-1, x-1)
			}
			
			integral.SetDoubleAt(y, x, sum)
		}
	}
	
	return integral
}

func (n *TrueNiblack) calculateLocalStats(integralSum, integralSumSq gocv.Mat, x1, y1, x2, y2 int) (float64, float64) {
	// Calculate sum using integral image
	sum := integralSum.GetDoubleAt(y2, x2)
	if x1 > 0 {
		sum -= integralSum.GetDoubleAt(y2, x1-1)
	}
	if y1 > 0 {
		sum -= integralSum.GetDoubleAt(y1-1, x2)
	}
	if x1 > 0 && y1 > 0 {
		sum += integralSum.GetDoubleAt(y1-1, x1-1)
	}
	
	// Calculate sum of squares
	sumSq := integralSumSq.GetDoubleAt(y2, x2)
	if x1 > 0 {
		sumSq -= integralSumSq.GetDoubleAt(y2, x1-1)
	}
	if y1 > 0 {
		sumSq -= integralSumSq.GetDoubleAt(y1-1, x2)
	}
	if x1 > 0 && y1 > 0 {
		sumSq += integralSumSq.GetDoubleAt(y1-1, x1-1)
	}
	
	// Calculate area
	area := float64((x2 - x1 + 1) * (y2 - y1 + 1))
	
	// Calculate mean and standard deviation
	mean := sum / area
	variance := (sumSq / area) - (mean * mean)
	if variance < 0 {
		variance = 0 // Avoid numerical errors
	}
	stddev := math.Sqrt(variance)
	
	return mean, stddev
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
} * (y2 - y1 + 1))
	
	// Calculate mean and standard deviation
	mean := sum / area
	variance := (sumSq / area) - (mean * mean)
	if variance < 0 {
		variance = 0 // Avoid numerical errors
	}
	stddev := math.Sqrt(variance)
	
	return mean, stddev
}

// TrueSauvola implements the true Sauvola algorithm
type TrueSauvola struct{}

// NewTrueSauvola creates a new true Sauvola algorithm
func NewTrueSauvola() *TrueSauvola {
	return &TrueSauvola{}
}

func (s *TrueSauvola) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// Convert to grayscale if needed
	gray := s.ensureGrayscale(input)
	defer func() {
		if gray.Ptr() != input.Ptr() {
			gray.Close()
		}
	}()

	// Get parameters
	windowSize := 15
	if val, ok := params["window_size"]; ok {
		if v, ok := val.(float64); ok {
			windowSize = int(v)
		}
	}

	k := 0.5 // Sauvola parameter
	if val, ok := params["k"]; ok {
		if v, ok := val.(float64); ok {
			k = v
		}
	}

	R := 128.0 // Dynamic range of standard deviation
	if val, ok := params["R"]; ok {
		if v, ok := val.(float64); ok {
			R = v
		}
	}

	// Ensure window size is odd
	if windowSize%2 == 0 {
		windowSize++
	}

	// Apply true Sauvola algorithm
	output := s.applyTrueSauvola(gray, windowSize, k, R)
	return output, nil
}

func (s *TrueSauvola) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"window_size": 15.0,
		"k":           0.5,
		"R":           128.0,
	}
}

func (s *TrueSauvola) GetName() string {
	return "True Sauvola"
}

func (s *TrueSauvola) GetDescription() string {
	return "True Sauvola local adaptive thresholding with dynamic range normalization"
}

func (s *TrueSauvola) Validate(params map[string]interface{}) error {
	if val, ok := params["window_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 3 || v > 101 {
				return fmt.Errorf("window_size must be between 3 and 101")
			}
		}
	}
	
	if val, ok := params["k"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0.1 || v > 1.0 {
				return fmt.Errorf("k must be between 0.1 and 1.0")
			}
		}
	}
	
	if val, ok := params["R"]; ok {
		if v, ok := val.(float64); ok {
			if v < 50.0 || v > 255.0 {
				return fmt.Errorf("R must be between 50 and 255")
			}
		}
	}
	
	return nil
}

func (s *TrueSauvola) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "window_size",
			Type:        "int",
			Min:         3.0,
			Max:         101.0,
			Default:     15.0,
			Description: "Local window size for statistics calculation",
		},
		{
			Name:        "k",
			Type:        "float",
			Min:         0.1,
			Max:         1.0,
			Default:     0.5,
			Description: "Sauvola parameter controlling threshold sensitivity",
		},
		{
			Name:        "R",
			Type:        "float",
			Min:         50.0,
			Max:         255.0,
			Default:     128.0,
			Description: "Dynamic range of standard deviation",
		},
	}
}

func (s *TrueSauvola) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}
	
	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (s *TrueSauvola) applyTrueSauvola(gray gocv.Mat, windowSize int, k, R float64) gocv.Mat {
	height := gray.Rows()
	width := gray.Cols()
	output := gocv.NewMat()
	gray.CopyTo(&output)
	
	halfWindow := windowSize / 2
	
	// Pre-calculate integral images for efficiency
	integralSum := s.calculateIntegralImage(gray)
	defer integralSum.Close()
	
	integralSumSq := s.calculateIntegralImageSquared(gray)
	defer integralSumSq.Close()
	
	// Process each pixel
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate window bounds
			x1 := max(0, x-halfWindow)
			y1 := max(0, y-halfWindow)
			x2 := min(width-1, x+halfWindow)
			y2 := min(height-1, y+halfWindow)
			
			// Calculate local mean and standard deviation
			mean, stddev := s.calculateLocalStats(integralSum, integralSumSq, x1, y1, x2, y2)
			
			// Apply Sauvola formula: T = mean * (1 + k * ((stddev / R) - 1))
			threshold := mean * (1.0 + k*((stddev/R)-1.0))
			
			// Apply threshold
			intensity := float64(gray.GetUCharAt(y, x))
			if intensity <= threshold {
				output.SetUCharAt(y, x, 0)
			} else {
				output.SetUCharAt(y, x, 255)
			}
		}
	}
	
	return output
}

func (s *TrueSauvola) calculateIntegralImage(gray gocv.Mat) gocv.Mat {
	integral := gocv.NewMatWithSize(gray.Rows(), gray.Cols(), gocv.MatTypeCV64F)
	
	for y := 0; y < gray.Rows(); y++ {
		for x := 0; x < gray.Cols(); x++ {
			val := float64(gray.GetUCharAt(y, x))
			
			sum := val
			if x > 0 {
				sum += integral.GetDoubleAt(y, x-1)
			}
			if y > 0 {
				sum += integral.GetDoubleAt(y-1, x)
			}
			if x > 0 && y > 0 {
				sum -= integral.GetDoubleAt(y-1, x-1)
			}
			
			integral.SetDoubleAt(y, x, sum)
		}
	}
	
	return integral
}

func (s *TrueSauvola) calculateIntegralImageSquared(gray gocv.Mat) gocv.Mat {
	integral := gocv.NewMatWithSize(gray.Rows(), gray.Cols(), gocv.MatTypeCV64F)
	
	for y := 0; y < gray.Rows(); y++ {
		for x := 0; x < gray.Cols(); x++ {
			val := float64(gray.GetUCharAt(y, x))
			valSq := val * val
			
			sum := valSq
			if x > 0 {
				sum += integral.GetDoubleAt(y, x-1)
			}
			if y > 0 {
				sum += integral.GetDoubleAt(y-1, x)
			}
			if x > 0 && y > 0 {
				sum -= integral.GetDoubleAt(y-1, x-1)
			}
			
			integral.SetDoubleAt(y, x, sum)
		}
	}
	
	return integral
}

func (s *TrueSauvola) calculateLocalStats(integralSum, integralSumSq gocv.Mat, x1, y1, x2, y2 int) (float64, float64) {
	// Calculate sum using integral image
	sum := integralSum.GetDoubleAt(y2, x2)
	if x1 > 0 {
		sum -= integralSum.GetDoubleAt(y2, x1-1)
	}
	if y1 > 0 {
		sum -= integralSum.GetDoubleAt(y1-1, x2)
	}
	if x1 > 0 && y1 > 0 {
		sum += integralSum.GetDoubleAt(y1-1, x1-1)
	}
	
	// Calculate sum of squares
	sumSq := integralSumSq.GetDoubleAt(y2, x2)
	if x1 > 0 {
		sumSq -= integralSumSq.GetDoubleAt(y2, x1-1)
	}
	if y1 > 0 {
		sumSq -= integralSumSq.GetDoubleAt(y1-1, x2)
	}
	if x1 > 0 && y1 > 0 {
		sumSq += integralSumSq.GetDoubleAt(y1-1, x1-1)
	}
	
	// Calculate area
	area := float64((x2 - x1 + 1)