// Concrete implementations of quality metrics
package metrics

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"
)

// PSNR implements Peak Signal-to-Noise Ratio metric
type PSNR struct{}

// NewPSNR creates a new PSNR metric
func NewPSNR() *PSNR {
	return &PSNR{}
}

func (p *PSNR) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}

	// Ensure same dimensions
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("image dimensions mismatch")
	}

	// Calculate MSE
	mse := p.calculateMSE(original, processed)
	if mse == 0 {
		return math.Inf(1), nil // Perfect match
	}

	// Calculate PSNR
	maxVal := 255.0
	psnr := 20 * math.Log10(maxVal/math.Sqrt(mse))

	return psnr, nil
}

func (p *PSNR) calculateMSE(original, processed gocv.Mat) float64 {
	// Convert to grayscale if needed
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

	sumSquaredDiff := 0.0
	totalPixels := gray1.Rows() * gray1.Cols()

	for y := 0; y < gray1.Rows(); y++ {
		for x := 0; x < gray1.Cols(); x++ {
			val1 := float64(gray1.GetUCharAt(y, x))
			val2 := float64(gray2.GetUCharAt(y, x))
			diff := val1 - val2
			sumSquaredDiff += diff * diff
		}
	}

	return sumSquaredDiff / float64(totalPixels)
}

func (p *PSNR) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}

	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (p *PSNR) GetName() string {
	return "PSNR"
}

func (p *PSNR) GetDescription() string {
	return "Peak Signal-to-Noise Ratio - measures image quality"
}

func (p *PSNR) GetRange() (float64, float64) {
	return 0, 100 // Practical range, can go higher
}

func (p *PSNR) IsHigherBetter() bool {
	return true
}

// SSIM implements Structural Similarity Index metric
type SSIM struct{}

// NewSSIM creates a new SSIM metric
func NewSSIM() *SSIM {
	return &SSIM{}
}

func (s *SSIM) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}

	// Ensure same dimensions
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("image dimensions mismatch")
	}

	// Convert to grayscale if needed
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

	// Calculate SSIM using sliding window
	ssim := s.calculateSSIM(gray1, gray2)

	return ssim, nil
}

func (s *SSIM) calculateSSIM(img1, img2 gocv.Mat) float64 {
	// SSIM constants
	const (
		C1 = 6.5025  // (0.01 * 255)^2
		C2 = 58.5225 // (0.03 * 255)^2
	)

	// Convert to float
	f1 := gocv.NewMat()
	defer f1.Close()
	img1.ConvertTo(&f1, gocv.MatTypeCV32F)

	f2 := gocv.NewMat()
	defer f2.Close()
	img2.ConvertTo(&f2, gocv.MatTypeCV32F)

	// Calculate means
	mu1 := gocv.NewMat()
	defer mu1.Close()
	gocv.GaussianBlur(f1, &mu1, image.Pt(11, 11), 1.5, 1.5, gocv.BorderDefault)

	mu2 := gocv.NewMat()
	defer mu2.Close()
	gocv.GaussianBlur(f2, &mu2, image.Pt(11, 11), 1.5, 1.5, gocv.BorderDefault)

	// Calculate squared means
	mu1Sq := gocv.NewMat()
	defer mu1Sq.Close()
	gocv.Multiply(mu1, mu1, &mu1Sq)

	mu2Sq := gocv.NewMat()
	defer mu2Sq.Close()
	gocv.Multiply(mu2, mu2, &mu2Sq)

	mu1Mu2 := gocv.NewMat()
	defer mu1Mu2.Close()
	gocv.Multiply(mu1, mu2, &mu1Mu2)

	// Calculate variances
	sigma1Sq := gocv.NewMat()
	defer sigma1Sq.Close()
	f1Sq := gocv.NewMat()
	defer f1Sq.Close()
	gocv.Multiply(f1, f1, &f1Sq)
	gocv.GaussianBlur(f1Sq, &sigma1Sq, image.Pt(11, 11), 1.5, 1.5, gocv.BorderDefault)
	gocv.Subtract(sigma1Sq, mu1Sq, &sigma1Sq)

	sigma2Sq := gocv.NewMat()
	defer sigma2Sq.Close()
	f2Sq := gocv.NewMat()
	defer f2Sq.Close()
	gocv.Multiply(f2, f2, &f2Sq)
	gocv.GaussianBlur(f2Sq, &sigma2Sq, image.Pt(11, 11), 1.5, 1.5, gocv.BorderDefault)
	gocv.Subtract(sigma2Sq, mu2Sq, &sigma2Sq)

	sigma12 := gocv.NewMat()
	defer sigma12.Close()
	f1f2 := gocv.NewMat()
	defer f1f2.Close()
	gocv.Multiply(f1, f2, &f1f2)
	gocv.GaussianBlur(f1f2, &sigma12, image.Pt(11, 11), 1.5, 1.5, gocv.BorderDefault)
	gocv.Subtract(sigma12, mu1Mu2, &sigma12)

	// Calculate SSIM
	numerator1 := gocv.NewMat()
	defer numerator1.Close()
	gocv.Multiply(mu1Mu2, gocv.NewMatFromScalar(gocv.Scalar{Val1: 2.0, Val2: 2.0, Val3: 2.0, Val4: 2.0}, gocv.MatTypeCV32F), &numerator1)
	gocv.Add(numerator1, gocv.NewMatFromScalar(gocv.Scalar{Val1: C1, Val2: C1, Val3: C1, Val4: C1}, gocv.MatTypeCV32F), &numerator1)

	numerator2 := gocv.NewMat()
	defer numerator2.Close()
	gocv.Multiply(sigma12, gocv.NewMatFromScalar(gocv.Scalar{Val1: 2.0, Val2: 2.0, Val3: 2.0, Val4: 2.0}, gocv.MatTypeCV32F), &numerator2)
	gocv.Add(numerator2, gocv.NewMatFromScalar(gocv.Scalar{Val1: C2, Val2: C2, Val3: C2, Val4: C2}, gocv.MatTypeCV32F), &numerator2)

	numerator := gocv.NewMat()
	defer numerator.Close()
	gocv.Multiply(numerator1, numerator2, &numerator)

	denominator1 := gocv.NewMat()
	defer denominator1.Close()
	gocv.Add(mu1Sq, mu2Sq, &denominator1)
	gocv.Add(denominator1, gocv.NewMatFromScalar(gocv.Scalar{Val1: C1, Val2: C1, Val3: C1, Val4: C1}, gocv.MatTypeCV32F), &denominator1)

	denominator2 := gocv.NewMat()
	defer denominator2.Close()
	gocv.Add(sigma1Sq, sigma2Sq, &denominator2)
	gocv.Add(denominator2, gocv.NewMatFromScalar(gocv.Scalar{Val1: C2, Val2: C2, Val3: C2, Val4: C2}, gocv.MatTypeCV32F), &denominator2)

	denominator := gocv.NewMat()
	defer denominator.Close()
	gocv.Multiply(denominator1, denominator2, &denominator)

	ssimMap := gocv.NewMat()
	defer ssimMap.Close()
	gocv.Divide(numerator, denominator, &ssimMap)

	// Calculate mean SSIM
	meanSSIM := calculateMean(ssimMap)
	return meanSSIM
}

func (s *SSIM) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}

	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (s *SSIM) GetName() string {
	return "SSIM"
}

func (s *SSIM) GetDescription() string {
	return "Structural Similarity Index - measures perceptual quality"
}

func (s *SSIM) GetRange() (float64, float64) {
	return 0, 1
}

func (s *SSIM) IsHigherBetter() bool {
	return true
}

// FMeasure implements F-measure for binarization quality
type FMeasure struct{}

// NewFMeasure creates a new F-measure metric
func NewFMeasure() *FMeasure {
	return &FMeasure{}
}

func (f *FMeasure) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}

	// Ensure same dimensions
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("image dimensions mismatch")
	}

	// Convert to grayscale and binarize if needed
	origBinary := f.ensureBinary(original)
	defer func() {
		if origBinary.Ptr() != original.Ptr() {
			origBinary.Close()
		}
	}()

	procBinary := f.ensureBinary(processed)
	defer func() {
		if procBinary.Ptr() != processed.Ptr() {
			procBinary.Close()
		}
	}()

	// Calculate true positives, false positives, false negatives
	tp, fp, fn := f.calculateConfusionMatrix(origBinary, procBinary)

	// Calculate precision and recall
	precision := 0.0
	if tp+fp > 0 {
		precision = tp / (tp + fp)
	}

	recall := 0.0
	if tp+fn > 0 {
		recall = tp / (tp + fn)
	}

	// Calculate F-measure
	if precision+recall == 0 {
		return 0, nil
	}

	fMeasure := 2 * (precision * recall) / (precision + recall)
	return fMeasure, nil
}

func (f *FMeasure) ensureBinary(input gocv.Mat) gocv.Mat {
	// Convert to grayscale first
	gray := gocv.NewMat()
	if input.Channels() == 1 {
		input.CopyTo(&gray)
	} else {
		gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	}

	// Check if already binary
	minVal, maxVal, _, _ := gocv.MinMaxLoc(gray)
	if minVal == 0.0 && maxVal == 255.0 {
		// Already binary
		return gray
	}

	// Apply Otsu thresholding to make binary
	binary := gocv.NewMat()
	gocv.Threshold(gray, &binary, 0, 255, gocv.ThresholdBinary+gocv.ThresholdOtsu)
	gray.Close()

	return binary
}

func (f *FMeasure) calculateConfusionMatrix(original, processed gocv.Mat) (float64, float64, float64) {
	tp := 0.0 // True positives (foreground correctly identified)
	fp := 0.0 // False positives (background incorrectly identified as foreground)
	fn := 0.0 // False negatives (foreground incorrectly identified as background)

	for y := 0; y < original.Rows(); y++ {
		for x := 0; x < original.Cols(); x++ {
			origVal := original.GetUCharAt(y, x)
			procVal := processed.GetUCharAt(y, x)

			// Assuming 255 = foreground, 0 = background
			origForeground := origVal > 127
			procForeground := procVal > 127

			if origForeground && procForeground {
				tp++
			} else if !origForeground && procForeground {
				fp++
			} else if origForeground && !procForeground {
				fn++
			}
			// tn (true negatives) not needed for F-measure
		}
	}

	return tp, fp, fn
}

func (f *FMeasure) GetName() string {
	return "F-Measure"
}

func (f *FMeasure) GetDescription() string {
	return "F-measure for binarization quality assessment"
}

func (f *FMeasure) GetRange() (float64, float64) {
	return 0, 1
}

func (f *FMeasure) IsHigherBetter() bool {
	return true
}

// MSE implements Mean Squared Error metric
type MSE struct{}

// NewMSE creates a new MSE metric
func NewMSE() *MSE {
	return &MSE{}
}

func (m *MSE) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}

	// Ensure same dimensions
	if original.Rows() != processed.Rows() || original.Cols() != processed.Cols() {
		return 0, fmt.Errorf("image dimensions mismatch")
	}

	// Convert to grayscale if needed
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

	sumSquaredDiff := 0.0
	totalPixels := gray1.Rows() * gray1.Cols()

	for y := 0; y < gray1.Rows(); y++ {
		for x := 0; x < gray1.Cols(); x++ {
			val1 := float64(gray1.GetUCharAt(y, x))
			val2 := float64(gray2.GetUCharAt(y, x))
			diff := val1 - val2
			sumSquaredDiff += diff * diff
		}
	}

	mse := sumSquaredDiff / float64(totalPixels)
	return mse, nil
}

func (m *MSE) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}

	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (m *MSE) GetName() string {
	return "MSE"
}

func (m *MSE) GetDescription() string {
	return "Mean Squared Error between images"
}

func (m *MSE) GetRange() (float64, float64) {
	return 0, 65025 // 255^2
}

func (m *MSE) IsHigherBetter() bool {
	return false
}

// ContrastRatio implements contrast ratio metric
type ContrastRatio struct{}

// NewContrastRatio creates a new contrast ratio metric
func NewContrastRatio() *ContrastRatio {
	return &ContrastRatio{}
}

func (c *ContrastRatio) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}

	origContrast := c.calculateContrast(original)
	procContrast := c.calculateContrast(processed)

	if origContrast == 0 {
		return 1.0, nil
	}

	ratio := procContrast / origContrast
	return ratio, nil
}

func (c *ContrastRatio) calculateContrast(input gocv.Mat) float64 {
	gray := c.ensureGrayscale(input)
	defer func() {
		if gray.Ptr() != input.Ptr() {
			gray.Close()
		}
	}()

	// Calculate standard deviation as a measure of contrast
	meanVal := calculateMean(gray)

	sumSquaredDiff := 0.0
	totalPixels := gray.Rows() * gray.Cols()

	for y := 0; y < gray.Rows(); y++ {
		for x := 0; x < gray.Cols(); x++ {
			val := float64(gray.GetUCharAt(y, x))
			diff := val - meanVal
			sumSquaredDiff += diff * diff
		}
	}

	variance := sumSquaredDiff / float64(totalPixels)
	return math.Sqrt(variance)
}

func (c *ContrastRatio) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}

	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (c *ContrastRatio) GetName() string {
	return "Contrast Ratio"
}

func (c *ContrastRatio) GetDescription() string {
	return "Ratio of contrast preservation"
}

func (c *ContrastRatio) GetRange() (float64, float64) {
	return 0, 2
}

func (c *ContrastRatio) IsHigherBetter() bool {
	return true
}

// Sharpness implements sharpness metric
type Sharpness struct{}

// NewSharpness creates a new sharpness metric
func NewSharpness() *Sharpness {
	return &Sharpness{}
}

func (s *Sharpness) Calculate(original, processed gocv.Mat) (float64, error) {
	if original.Empty() || processed.Empty() {
		return 0, fmt.Errorf("empty images")
	}

	origSharpness := s.calculateSharpness(original)
	procSharpness := s.calculateSharpness(processed)

	if origSharpness == 0 {
		return 1.0, nil
	}

	ratio := procSharpness / origSharpness
	return ratio, nil
}

func (s *Sharpness) calculateSharpness(input gocv.Mat) float64 {
	gray := s.ensureGrayscale(input)
	defer func() {
		if gray.Ptr() != input.Ptr() {
			gray.Close()
		}
	}()

	// Apply Laplacian to detect edges
	laplacian := gocv.NewMat()
	defer laplacian.Close()
	gocv.Laplacian(gray, &laplacian, gocv.MatTypeCV64F, 1, 1, 0, gocv.BorderDefault)

	// Calculate variance of Laplacian as sharpness measure
	meanVal := calculateMean(laplacian)

	sumSquaredDiff := 0.0
	totalPixels := laplacian.Rows() * laplacian.Cols()

	for y := 0; y < laplacian.Rows(); y++ {
		for x := 0; x < laplacian.Cols(); x++ {
			val := laplacian.GetDoubleAt(y, x)
			diff := val - meanVal
			sumSquaredDiff += diff * diff
		}
	}

	variance := sumSquaredDiff / float64(totalPixels)
	return variance
}

func (s *Sharpness) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}

	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (s *Sharpness) GetName() string {
	return "Sharpness"
}

func (s *Sharpness) GetDescription() string {
	return "Edge preservation measure"
}

func (s *Sharpness) GetRange() (float64, float64) {
	return 0, 2
}

func (s *Sharpness) IsHigherBetter() bool {
	return true
}

// Helper function to calculate mean of a matrix
func calculateMean(mat gocv.Mat) float64 {
	sum := 0.0
	totalPixels := mat.Rows() * mat.Cols()

	if mat.Type() == gocv.MatTypeCV64F {
		// For double precision matrices
		for y := 0; y < mat.Rows(); y++ {
			for x := 0; x < mat.Cols(); x++ {
				sum += mat.GetDoubleAt(y, x)
			}
		}
	} else {
		// For 8-bit matrices
		for y := 0; y < mat.Rows(); y++ {
			for x := 0; x < mat.Cols(); x++ {
				sum += float64(mat.GetUCharAt(y, x))
			}
		}
	}

	return sum / float64(totalPixels)
}
