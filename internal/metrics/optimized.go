// Optimized metrics using GoCV built-in functions
package metrics

import (
	"fmt"
	"math"

	"gocv.io/x/gocv"
)

// Metric defines the interface for quality metrics
type Metric interface {
	Calculate(original, processed gocv.Mat) (float64, error)
	GetName() string
	GetDescription() string
	GetRange() (float64, float64)
	IsHigherBetter() bool
}

// Evaluator manages and calculates multiple metrics
type Evaluator struct {
	metrics map[string]Metric
}

func NewEvaluator() *Evaluator {
	e := &Evaluator{
		metrics: make(map[string]Metric),
	}
	e.RegisterDefaultMetrics()
	return e
}

func (e *Evaluator) RegisterDefaultMetrics() {
	e.Register("psnr", NewPSNR())
	e.Register("ssim", NewSSIM())
	e.Register("mse", NewMSE())
}

func (e *Evaluator) Register(name string, metric Metric) {
	e.metrics[name] = metric
}

func (e *Evaluator) Calculate(name string, original, processed gocv.Mat) (float64, error) {
	metric, exists := e.metrics[name]
	if !exists {
		return 0, fmt.Errorf("metric not found: %s", name)
	}
	return metric.Calculate(original, processed)
}

func (e *Evaluator) CalculateAll(original, processed gocv.Mat) map[string]float64 {
	results := make(map[string]float64)
	for name, metric := range e.metrics {
		if value, err := metric.Calculate(original, processed); err == nil {
			results[name] = value
		}
	}
	return results
}

func (e *Evaluator) CalculatePSNR(original, processed gocv.Mat) (float64, error) {
	return e.Calculate("psnr", original, processed)
}

func (e *Evaluator) CalculateSSIM(original, processed gocv.Mat) (float64, error) {
	return e.Calculate("ssim", original, processed)
}

func (e *Evaluator) EvaluateStep(before, after gocv.Mat, stepName string) map[string]float64 {
	metrics := make(map[string]float64)
	
	if psnr, err := e.CalculatePSNR(before, after); err == nil {
		metrics["psnr"] = psnr
	}
	
	if ssim, err := e.CalculateSSIM(before, after); err == nil {
		metrics["ssim"] = ssim
	}
	
	return metrics
}

// PSNR using GoCV built-in functions
type PSNR struct{}

func NewPSNR() *PSNR { return &PSNR{} }

func (p *PSNR) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("dimension mismatch")
	}

	gray1 := p.ensureGrayscale(original)
	defer func() {
		if gray1.Ptr() != original.Ptr() {
			gray1.Close()
		}
	}()

	gray2 := p.ensureGrayscale(processed)
	defer func() {
		if gray2.Ptr() != processed.Ptr() {
			gray2.Close()
		}
	}()

	// Use GoCV's optimized PSNR calculation
	psnr := gocv.PSNR(gray1, gray2)
	if math.IsInf(psnr, 1) {
		return 100.0, nil // Cap infinite PSNR at 100
	}
	return psnr, nil
}

func (p *PSNR) ensureGrayscale(m gocv.Mat) gocv.Mat {
	if m.Channels() == 1 {
		return m
	}
	g := gocv.NewMat()
	gocv.CvtColor(m, &g, gocv.ColorBGRToGray)
	return g
}

func (p *PSNR) GetName() string              { return "PSNR" }
func (p *PSNR) GetDescription() string       { return "Peak Signal-to-Noise Ratio" }
func (p *PSNR) GetRange() (float64, float64) { return 0, 100 }
func (p *PSNR) IsHigherBetter() bool         { return true }

// SSIM using optimized calculation
type SSIM struct{}

func NewSSIM() *SSIM { return &SSIM{} }

func (s *SSIM) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("dimension mismatch")
	}

	gray1 := s.ensureGrayscale(original)
	defer func() {
		if gray1.Ptr() != original.Ptr() {
			gray1.Close()
		}
	}()

	gray2 := s.ensureGrayscale(processed)
	defer func() {
		if gray2.Ptr() != processed.Ptr() {
			gray2.Close()
		}
	}()

	// Convert to float for precision
	f1, f2 := gocv.NewMat(), gocv.NewMat()
	defer f1.Close()
	defer f2.Close()
	gray1.ConvertTo(&f1, gocv.MatTypeCV32F)
	gray2.ConvertTo(&f2, gocv.MatTypeCV32F)

	// SSIM constants
	const C1, C2 = 6.5025, 58.5225

	// Calculate means using GoCV optimized functions
	mu1 := f1.Mean().Val1
	mu2 := f2.Mean().Val1

	// Calculate squares and cross multiplication
	f1Sq, f2Sq, f1f2 := gocv.NewMat(), gocv.NewMat(), gocv.NewMat()
	defer f1Sq.Close()
	defer f2Sq.Close()
	defer f1f2.Close()

	gocv.Multiply(f1, f1, &f1Sq)
	gocv.Multiply(f2, f2, &f2Sq)
	gocv.Multiply(f1, f2, &f1f2)

	// Calculate variances and covariance
	sigma1Sq := f1Sq.Mean().Val1 - mu1*mu1
	sigma2Sq := f2Sq.Mean().Val1 - mu2*mu2
	sigma12 := f1f2.Mean().Val1 - mu1*mu2

	// SSIM formula
	num := (2*mu1*mu2 + C1) * (2*sigma12 + C2)
	den := (mu1*mu1 + mu2*mu2 + C1) * (sigma1Sq + sigma2Sq + C2)
	
	if den == 0 {
		return 1.0, nil
	}
	return num / den, nil
}

func (s *SSIM) ensureGrayscale(m gocv.Mat) gocv.Mat {
	if m.Channels() == 1 {
		return m
	}
	g := gocv.NewMat()
	gocv.CvtColor(m, &g, gocv.ColorBGRToGray)
	return g
}

func (s *SSIM) GetName() string              { return "SSIM" }
func (s *SSIM) GetDescription() string       { return "Structural Similarity Index" }
func (s *SSIM) GetRange() (float64, float64) { return 0, 1 }
func (s *SSIM) IsHigherBetter() bool         { return true }

// MSE using GoCV optimized functions
type MSE struct{}

func NewMSE() *MSE { return &MSE{} }

func (m *MSE) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("dimension mismatch")
	}

	gray1 := m.ensureGrayscale(original)
	defer func() {
		if gray1.Ptr() != original.Ptr() {
			gray1.Close()
		}
	}()

	gray2 := m.ensureGrayscale(processed)
	defer func() {
		if gray2.Ptr() != processed.Ptr() {
			gray2.Close()
		}
	}()

	// Calculate difference using GoCV
	diff := gocv.NewMat()
	defer diff.Close()
	gocv.AbsDiff(gray1, gray2, &diff)

	// Square the difference
	diffSq := gocv.NewMat()
	defer diffSq.Close()
	gocv.Multiply(diff, diff, &diffSq)

	// Return mean using GoCV optimized mean calculation
	return diffSq.Mean().Val1, nil
}

func (m *MSE) ensureGrayscale(mat gocv.Mat) gocv.Mat {
	if mat.Channels() == 1 {
		return mat
	}
	g := gocv.NewMat()
	gocv.CvtColor(mat, &g, gocv.ColorBGRToGray)
	return g
}

func (m *MSE) GetName() string              { return "MSE" }
func (m *MSE) GetDescription() string       { return "Mean Squared Error" }
func (m *MSE) GetRange() (float64, float64) { return 0, 65025 }
func (m *MSE) IsHigherBetter() bool         { return false }