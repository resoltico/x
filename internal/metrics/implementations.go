// Corrected optimized metrics using actual GoCV API
package metrics

import (
	"fmt"
	"math"

	"gocv.io/x/gocv"
)

// PSNR using Mean for MSE calculation
type PSNR struct{}

func NewPSNR() *PSNR { return &PSNR{} }

func (p *PSNR) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("dimension mismatch")
	}

	gray1 := p.gray(original)
	defer func() {
		if gray1.Ptr() != original.Ptr() {
			gray1.Close()
		}
	}()

	gray2 := p.gray(processed)
	defer func() {
		if gray2.Ptr() != processed.Ptr() {
			gray2.Close()
		}
	}()

	diff := gocv.NewMat()
	defer diff.Close()
	gocv.AbsDiff(gray1, gray2, &diff)

	diffSq := gocv.NewMat()
	defer diffSq.Close()
	gocv.Multiply(diff, diff, &diffSq)

	mse := diffSq.Mean().Val1
	if mse == 0 {
		return math.Inf(1), nil
	}
	return 20 * math.Log10(255.0/math.Sqrt(mse)), nil
}

func (p *PSNR) gray(m gocv.Mat) gocv.Mat {
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

// SSIM using Mean for statistics
type SSIM struct{}

func NewSSIM() *SSIM { return &SSIM{} }

func (s *SSIM) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("dimension mismatch")
	}

	gray1 := s.gray(original)
	defer func() {
		if gray1.Ptr() != original.Ptr() {
			gray1.Close()
		}
	}()

	gray2 := s.gray(processed)
	defer func() {
		if gray2.Ptr() != processed.Ptr() {
			gray2.Close()
		}
	}()

	f1, f2 := gocv.NewMat(), gocv.NewMat()
	defer f1.Close()
	defer f2.Close()
	gray1.ConvertTo(&f1, gocv.MatTypeCV32F)
	gray2.ConvertTo(&f2, gocv.MatTypeCV32F)

	const C1, C2 = 6.5025, 58.5225
	mu1, mu2 := f1.Mean().Val1, f2.Mean().Val1

	f1Sq, f2Sq, f1f2 := gocv.NewMat(), gocv.NewMat(), gocv.NewMat()
	defer f1Sq.Close()
	defer f2Sq.Close()
	defer f1f2.Close()
	gocv.Multiply(f1, f1, &f1Sq)
	gocv.Multiply(f2, f2, &f2Sq)
	gocv.Multiply(f1, f2, &f1f2)

	sigma1Sq := f1Sq.Mean().Val1 - mu1*mu1
	sigma2Sq := f2Sq.Mean().Val1 - mu2*mu2
	sigma12 := f1f2.Mean().Val1 - mu1*mu2

	num := (2*mu1*mu2 + C1) * (2*sigma12 + C2)
	den := (mu1*mu1 + mu2*mu2 + C1) * (sigma1Sq + sigma2Sq + C2)
	if den == 0 {
		return 1.0, nil
	}
	return num / den, nil
}

func (s *SSIM) gray(m gocv.Mat) gocv.Mat {
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

// FMeasure using CountNonZero
type FMeasure struct{}

func NewFMeasure() *FMeasure { return &FMeasure{} }

func (f *FMeasure) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("dimension mismatch")
	}

	orig, proc := f.binary(original), f.binary(processed)
	defer func() {
		if orig.Ptr() != original.Ptr() {
			orig.Close()
		}
		if proc.Ptr() != processed.Ptr() {
			proc.Close()
		}
	}()

	tp, fp, fn := f.confusionCounts(orig, proc)
	precision := tp / (tp + fp + 1e-10)
	recall := tp / (tp + fn + 1e-10)
	if precision+recall == 0 {
		return 0, nil
	}
	return 2 * precision * recall / (precision + recall), nil
}

func (f *FMeasure) binary(m gocv.Mat) gocv.Mat {
	gray := gocv.NewMat()
	if m.Channels() == 1 {
		m.CopyTo(&gray)
	} else {
		gocv.CvtColor(m, &gray, gocv.ColorBGRToGray)
	}

	min, max, _, _ := gocv.MinMaxLoc(gray)
	if min == 0.0 && max == 255.0 {
		return gray
	}

	binary := gocv.NewMat()
	gocv.Threshold(gray, &binary, 0, 255, gocv.ThresholdBinary+gocv.ThresholdOtsu)
	gray.Close()
	return binary
}

func (f *FMeasure) confusionCounts(orig, proc gocv.Mat) (float64, float64, float64) {
	tp, fp, fn := gocv.NewMat(), gocv.NewMat(), gocv.NewMat()
	defer tp.Close()
	defer fp.Close()
	defer fn.Close()

	gocv.BitwiseAnd(orig, proc, &tp)

	origInv := gocv.NewMat()
	defer origInv.Close()
	gocv.BitwiseNot(orig, &origInv)
	gocv.BitwiseAnd(origInv, proc, &fp)

	procInv := gocv.NewMat()
	defer procInv.Close()
	gocv.BitwiseNot(proc, &procInv)
	gocv.BitwiseAnd(orig, procInv, &fn)

	return float64(gocv.CountNonZero(tp)), float64(gocv.CountNonZero(fp)), float64(gocv.CountNonZero(fn))
}

func (f *FMeasure) GetName() string              { return "F-Measure" }
func (f *FMeasure) GetDescription() string       { return "F-measure for binarization quality" }
func (f *FMeasure) GetRange() (float64, float64) { return 0, 1 }
func (f *FMeasure) IsHigherBetter() bool         { return true }

// MSE using Mean
type MSE struct{}

func NewMSE() *MSE { return &MSE{} }

func (m *MSE) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("dimension mismatch")
	}

	gray1 := m.gray(original)
	defer func() {
		if gray1.Ptr() != original.Ptr() {
			gray1.Close()
		}
	}()

	gray2 := m.gray(processed)
	defer func() {
		if gray2.Ptr() != processed.Ptr() {
			gray2.Close()
		}
	}()

	diff := gocv.NewMat()
	defer diff.Close()
	gocv.AbsDiff(gray1, gray2, &diff)

	diffSq := gocv.NewMat()
	defer diffSq.Close()
	gocv.Multiply(diff, diff, &diffSq)

	return diffSq.Mean().Val1, nil
}

func (m *MSE) gray(mat gocv.Mat) gocv.Mat {
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

// ContrastRatio using MeanStdDev
type ContrastRatio struct{}

func NewContrastRatio() *ContrastRatio { return &ContrastRatio{} }

func (c *ContrastRatio) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}

	origContrast := c.contrast(original)
	procContrast := c.contrast(processed)
	if origContrast == 0 {
		return 1.0, nil
	}
	return procContrast / origContrast, nil
}

func (c *ContrastRatio) contrast(m gocv.Mat) float64 {
	gray := c.gray(m)
	defer func() {
		if gray.Ptr() != m.Ptr() {
			gray.Close()
		}
	}()

	meanMat, stdMat := gocv.NewMat(), gocv.NewMat()
	defer meanMat.Close()
	defer stdMat.Close()
	gocv.MeanStdDev(gray, &meanMat, &stdMat)
	return stdMat.GetDoubleAt(0, 0)
}

func (c *ContrastRatio) gray(m gocv.Mat) gocv.Mat {
	if m.Channels() == 1 {
		return m
	}
	g := gocv.NewMat()
	gocv.CvtColor(m, &g, gocv.ColorBGRToGray)
	return g
}

func (c *ContrastRatio) GetName() string              { return "Contrast Ratio" }
func (c *ContrastRatio) GetDescription() string       { return "Ratio of contrast preservation" }
func (c *ContrastRatio) GetRange() (float64, float64) { return 0, 2 }
func (c *ContrastRatio) IsHigherBetter() bool         { return true }

// Sharpness using Laplacian variance
type Sharpness struct{}

func NewSharpness() *Sharpness { return &Sharpness{} }

func (s *Sharpness) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}

	origSharp := s.sharpness(original)
	procSharp := s.sharpness(processed)
	if origSharp == 0 {
		return 1.0, nil
	}
	return procSharp / origSharp, nil
}

func (s *Sharpness) sharpness(m gocv.Mat) float64 {
	gray := s.gray(m)
	defer func() {
		if gray.Ptr() != m.Ptr() {
			gray.Close()
		}
	}()

	lap := gocv.NewMat()
	defer lap.Close()
	gocv.Laplacian(gray, &lap, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)

	meanMat, stdMat := gocv.NewMat(), gocv.NewMat()
	defer meanMat.Close()
	defer stdMat.Close()
	gocv.MeanStdDev(lap, &meanMat, &stdMat)

	std := stdMat.GetDoubleAt(0, 0)
	return std * std // variance = stddev^2
}

func (s *Sharpness) gray(m gocv.Mat) gocv.Mat {
	if m.Channels() == 1 {
		return m
	}
	g := gocv.NewMat()
	gocv.CvtColor(m, &g, gocv.ColorBGRToGray)
	return g
}

func (s *Sharpness) GetName() string              { return "Sharpness" }
func (s *Sharpness) GetDescription() string       { return "Edge preservation measure" }
func (s *Sharpness) GetRange() (float64, float64) { return 0, 2 }
func (s *Sharpness) IsHigherBetter() bool         { return true }
