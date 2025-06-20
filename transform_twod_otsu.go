package main

import (
	"fmt"
	"image"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

// TwoDOtsu implements 2D Otsu thresholding with GoCV API usage
type TwoDOtsu struct {
	ThreadSafeTransformation
	debugImage *DebugImage

	// Thread-safe parameters with validation
	paramMutex      sync.RWMutex
	windowRadius    int
	epsilon         float64
	morphKernelSize int

	// Callback for parameter changes
	onParameterChanged func()
}

// NewTwoDOtsu creates a new 2D Otsu transformation with validated parameters
func NewTwoDOtsu(config *DebugConfig) *TwoDOtsu {
	return &TwoDOtsu{
		debugImage:      NewDebugImage(config),
		windowRadius:    5,
		epsilon:         0.02,
		morphKernelSize: 3, // Must be odd
	}
}

func (t *TwoDOtsu) Name() string {
	return "2D Otsu"
}

func (t *TwoDOtsu) GetParameters() map[string]interface{} {
	t.paramMutex.RLock()
	defer t.paramMutex.RUnlock()

	return map[string]interface{}{
		"windowRadius":    t.windowRadius,
		"epsilon":         t.epsilon,
		"morphKernelSize": t.morphKernelSize,
	}
}

func (t *TwoDOtsu) SetParameters(params map[string]interface{}) {
	t.paramMutex.Lock()
	defer t.paramMutex.Unlock()

	if radius, ok := params["windowRadius"].(int); ok {
		if radius >= 1 && radius <= 20 {
			t.windowRadius = radius
		}
	}
	if eps, ok := params["epsilon"].(float64); ok {
		if eps > 0.001 && eps <= 1.0 {
			t.epsilon = eps
		}
	}
	if kernel, ok := params["morphKernelSize"].(int); ok {
		if kernel >= 1 && kernel <= 15 && kernel%2 == 1 {
			t.morphKernelSize = kernel
		}
	}
}

func (t *TwoDOtsu) Apply(src gocv.Mat) gocv.Mat {
	return t.applyWithScale(src, 1.0)
}

func (t *TwoDOtsu) ApplyPreview(src gocv.Mat) gocv.Mat {
	return t.applyWithScale(src, 0.5)
}

func (t *TwoDOtsu) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	t.onParameterChanged = onParameterChanged
	return t.createParameterUI()
}

func (t *TwoDOtsu) Close() {
	// No resources to cleanup - GoCV MatProfile handles tracking
}

func (t *TwoDOtsu) applyWithScale(src gocv.Mat, scale float64) gocv.Mat {
	defer func() {
		if r := recover(); r != nil {
			t.debugImage.LogError(fmt.Errorf("panic in 2D Otsu: %v", r))
		}
	}()

	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Starting binarization (scale: %.2f)", scale))

	if src.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Input matrix is empty")
		return gocv.NewMat()
	}

	// Get thread-safe parameters
	t.paramMutex.RLock()
	windowRadius := t.windowRadius
	epsilon := t.epsilon
	morphKernelSize := t.morphKernelSize
	t.paramMutex.RUnlock()

	// Scale input if needed
	var workingImage gocv.Mat
	if scale != 1.0 {
		newWidth := int(float64(src.Cols()) * scale)
		newHeight := int(float64(src.Rows()) * scale)

		if newWidth <= 0 || newHeight <= 0 || newWidth > 16384 || newHeight > 16384 {
			t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("ERROR: Invalid scaled dimensions: %dx%d", newWidth, newHeight))
			return gocv.NewMat()
		}

		workingImage = gocv.NewMat()
		gocv.Resize(src, &workingImage, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)
		defer workingImage.Close()
		t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Scaled to %dx%d", newWidth, newHeight))
	} else {
		workingImage = src.Clone()
		defer workingImage.Close()
	}

	// Convert to grayscale if needed
	var grayscale gocv.Mat
	if workingImage.Channels() > 1 {
		grayscale = gocv.NewMat()
		gocv.CvtColor(workingImage, &grayscale, gocv.ColorBGRToGray)
		defer grayscale.Close()
	} else {
		grayscale = workingImage.Clone()
		defer grayscale.Close()
	}

	if grayscale.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Grayscale conversion failed")
		return gocv.NewMat()
	}

	t.debugImage.LogMatInfo("grayscale", grayscale)

	// Apply guided filter using GoCV APIs
	guided := t.applyGuidedFilterCorrected(grayscale, windowRadius, epsilon)
	defer guided.Close()

	if guided.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Guided filter failed")
		guided = grayscale.Clone()
		defer guided.Close()
	}

	// Apply 2D Otsu thresholding
	binaryResult := t.apply2DOtsuFixed(grayscale, guided)
	defer func() {
		if scale != 1.0 && !binaryResult.Empty() {
			binaryResult.Close()
		}
	}()

	if binaryResult.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Binarization failed")
		return gocv.NewMat()
	}

	// Post-processing with morphological operations using standard GoCV
	processed := t.applyMorphologicalOps(binaryResult, morphKernelSize)

	// Scale back to original size if needed
	var result gocv.Mat
	if scale != 1.0 {
		result = gocv.NewMat()
		gocv.Resize(processed, &result, image.Point{X: src.Cols(), Y: src.Rows()}, 0, 0, gocv.InterpolationLinear)
		processed.Close()
		t.debugImage.LogAlgorithmStep("2D Otsu", "Scaled back to original size")
	} else {
		result = processed
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", "Completed successfully")
	return result
}

// Guided filter using GoCV API signatures
func (t *TwoDOtsu) applyGuidedFilterCorrected(src gocv.Mat, windowRadius int, epsilon float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("GuidedFilter", "Starting guided filter with API")

	if src.Empty() {
		return gocv.NewMat()
	}

	// Validate epsilon to prevent division by zero
	if epsilon <= 0 {
		epsilon = 0.001
	}

	// Convert to float32 for processing using API
	srcFloat := gocv.NewMat()
	defer srcFloat.Close()
	src.ConvertTo(&srcFloat, gocv.MatTypeCV32F) // Only dst and type parameters
	srcFloat.DivideFloat(255.0)                 // Normalize to [0,1] range

	kernelSize := 2*windowRadius + 1

	meanI := gocv.NewMat()
	defer meanI.Close()
	gocv.Blur(srcFloat, &meanI, image.Point{X: kernelSize, Y: kernelSize})

	// Compute mean of I*I
	meanII := gocv.NewMat()
	defer meanII.Close()
	srcSquared := gocv.NewMat()
	defer srcSquared.Close()
	gocv.Multiply(srcFloat, srcFloat, &srcSquared)
	gocv.Blur(srcSquared, &meanII, image.Point{X: kernelSize, Y: kernelSize})

	// Variance calculation using standard GoCV
	varI := gocv.NewMat()
	defer varI.Close()
	meanISquared := gocv.NewMat()
	defer meanISquared.Close()
	gocv.Multiply(meanI, meanI, &meanISquared)
	gocv.Subtract(meanII, meanISquared, &varI)

	// Coefficient calculation using standard GoCV operations
	a := gocv.NewMat()
	defer a.Close()
	denominator := gocv.NewMat()
	defer denominator.Close()
	varI.CopyTo(&denominator)
	denominator.AddFloat(float32(epsilon))
	gocv.Divide(varI, denominator, &a)

	// b = mean(I) * (1 - a) using standard operations
	b := gocv.NewMat()
	defer b.Close()
	oneMinusA := gocv.NewMat()
	defer oneMinusA.Close()

	// Create ones matrix using standard GoCV
	ones := gocv.NewMatWithSize(a.Rows(), a.Cols(), a.Type())
	defer ones.Close()
	ones.SetTo(gocv.NewScalar(1, 0, 0, 0))
	gocv.Subtract(ones, a, &oneMinusA)
	gocv.Multiply(meanI, oneMinusA, &b)

	// Smooth coefficients using standard blur
	meanA := gocv.NewMat()
	defer meanA.Close()
	gocv.Blur(a, &meanA, image.Point{X: kernelSize, Y: kernelSize})

	meanB := gocv.NewMat()
	defer meanB.Close()
	gocv.Blur(b, &meanB, image.Point{X: kernelSize, Y: kernelSize})

	// Final result: q = mean_a * I + mean_b
	resultFloat := gocv.NewMat()
	defer resultFloat.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	gocv.Multiply(meanA, srcFloat, &temp)
	gocv.Add(temp, meanB, &resultFloat)

	// Convert back to uint8 using API
	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U) // Only dst and type parameters

	t.debugImage.LogFilter("GuidedFilter", fmt.Sprintf("radius=%d epsilon=%.3f", windowRadius, epsilon))
	return result
}

// 2D Otsu algorithm implementation
func (t *TwoDOtsu) apply2DOtsuFixed(gray, guided gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", "Constructing 2D histogram")

	if gray.Empty() || guided.Empty() {
		return gocv.NewMat()
	}

	// Use standard GoCV ToBytes for data access
	grayData := gray.ToBytes()
	guidedData := guided.ToBytes()

	if len(grayData) == 0 || len(guidedData) == 0 || len(grayData) != len(guidedData) {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Invalid data")
		return gocv.NewMat()
	}

	// Build 2D histogram - optimized memory allocation
	hist := make([][]float64, 256)
	for i := range hist {
		hist[i] = make([]float64, 256)
	}

	totalPixels := len(grayData)

	// Single pass histogram construction
	for i := 0; i < totalPixels; i++ {
		g := int(grayData[i])
		f := int(guidedData[i])

		// Bounds checking
		if g >= 0 && g < 256 && f >= 0 && f < 256 {
			hist[g][f]++
		}
	}

	// Normalize histogram to probabilities in-place
	invTotalPixels := 1.0 / float64(totalPixels)
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			hist[g][f] *= invTotalPixels
		}
	}

	// Find optimal thresholds using 2D Otsu formulation
	bestS, bestT, maxVariance := t.findOptimalThresholdsFixed(hist)
	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	// Apply thresholding using GoCV's Mat operations
	t.debugImage.LogAlgorithmStep("2D Otsu", "Applying 2D Otsu classification")

	size := gray.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	// Process pixel by pixel using 2D Otsu classification
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grayVal := int(gray.GetUCharAt(y, x))
			guidedVal := int(guided.GetUCharAt(y, x))

			// Use proper statistical classification
			isBackground := t.classifyPixelStatistically(grayVal, guidedVal, bestS, bestT, hist)

			if isBackground {
				result.SetUCharAt(y, x, 255) // Background
			} else {
				result.SetUCharAt(y, x, 0) // Foreground
			}
		}
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarization completed")
	return result
}

// 2D Otsu optimal threshold finding using between-class scatter matrix
func (t *TwoDOtsu) findOptimalThresholdsFixed(hist [][]float64) (int, int, float64) {
	maxVariance := 0.0
	bestS, bestT := 0, 0

	// Exhaustive search for optimal thresholds
	for s := 1; s < 255; s++ {
		for thresholdT := 1; thresholdT < 255; thresholdT++ {
			// Calculate proper between-class scatter matrix trace
			variance := t.calculateBetweenClassScatterTrace(hist, s, thresholdT)
			if variance > maxVariance {
				maxVariance = variance
				bestS = s
				bestT = thresholdT
			}
		}
	}

	return bestS, bestT, maxVariance
}

// 2D between-class scatter matrix trace calculation
func (t *TwoDOtsu) calculateBetweenClassScatterTrace(hist [][]float64, s, thresholdT int) float64 {
	// Calculate statistics for 4 quadrants/regions
	var w [4]float64   // weights (probabilities)
	var muG [4]float64 // mean gray values
	var muF [4]float64 // mean guided values

	// Region 0: g <= s, f <= t
	for g := 0; g <= s; g++ {
		for f := 0; f <= thresholdT; f++ {
			prob := hist[g][f]
			w[0] += prob
			muG[0] += float64(g) * prob
			muF[0] += float64(f) * prob
		}
	}

	// Region 1: g > s, f <= t
	for g := s + 1; g < 256; g++ {
		for f := 0; f <= thresholdT; f++ {
			prob := hist[g][f]
			w[1] += prob
			muG[1] += float64(g) * prob
			muF[1] += float64(f) * prob
		}
	}

	// Region 2: g <= s, f > t
	for g := 0; g <= s; g++ {
		for f := thresholdT + 1; f < 256; f++ {
			prob := hist[g][f]
			w[2] += prob
			muG[2] += float64(g) * prob
			muF[2] += float64(f) * prob
		}
	}

	// Region 3: g > s, f > t
	for g := s + 1; g < 256; g++ {
		for f := thresholdT + 1; f < 256; f++ {
			prob := hist[g][f]
			w[3] += prob
			muG[3] += float64(g) * prob
			muF[3] += float64(f) * prob
		}
	}

	// Normalize means by weights and handle empty regions
	for i := 0; i < 4; i++ {
		if w[i] > 1e-10 {
			muG[i] /= w[i]
			muF[i] /= w[i]
		}
	}

	// Calculate overall means
	muGTotal := 0.0
	muFTotal := 0.0
	for i := 0; i < 4; i++ {
		muGTotal += w[i] * muG[i]
		muFTotal += w[i] * muF[i]
	}

	// Calculate between-class scatter matrix trace
	betweenClassVariance := 0.0
	for i := 0; i < 4; i++ {
		if w[i] > 1e-10 {
			diffG := muG[i] - muGTotal
			diffF := muF[i] - muFTotal
			betweenClassVariance += w[i] * (diffG*diffG + diffF*diffF)
		}
	}

	return betweenClassVariance
}

// Statistical classification based on 2D Otsu theory
func (t *TwoDOtsu) classifyPixelStatistically(grayVal, guidedVal, bestS, bestT int, hist [][]float64) bool {
	// Calculate which region this pixel belongs to
	var region int
	if grayVal <= bestS && guidedVal <= bestT {
		region = 0 // Q1
	} else if grayVal > bestS && guidedVal <= bestT {
		region = 1 // Q2
	} else if grayVal <= bestS && guidedVal > bestT {
		region = 2 // Q3
	} else {
		region = 3 // Q4
	}

	// Use statistical approach: calculate likelihood
	foregroundLikelihood := hist[grayVal][guidedVal] * (1.0 + float64(256-grayVal)/256.0)
	backgroundLikelihood := hist[grayVal][guidedVal] * (1.0 + float64(grayVal)/256.0)

	// Classification based on region and likelihood
	if region == 3 {
		return true // High gray, high guided -> likely background
	}
	if region == 0 {
		return false // Low gray, low guided -> likely foreground
	}

	// For edge regions, use likelihood comparison
	return backgroundLikelihood > foregroundLikelihood
}

// Use standard GoCV morphological operations
func (t *TwoDOtsu) applyMorphologicalOps(src gocv.Mat, morphKernelSize int) gocv.Mat {
	if morphKernelSize <= 1 {
		return src.Clone()
	}

	// Use standard GoCV getStructuringElement
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: morphKernelSize, Y: morphKernelSize})
	defer kernel.Close()

	// Closing (dilation followed by erosion) to fill small holes
	closed := gocv.NewMat()
	defer closed.Close()
	gocv.MorphologyEx(src, &closed, gocv.MorphClose, kernel)

	// Opening (erosion followed by dilation) to remove small noise
	result := gocv.NewMat()
	gocv.MorphologyEx(closed, &result, gocv.MorphOpen, kernel)

	return result
}

func (t *TwoDOtsu) createParameterUI() *fyne.Container {
	// Window Radius parameter with validation
	radiusLabel := widget.NewLabel("Window Radius (1-20):")
	radiusEntry := widget.NewEntry()

	t.paramMutex.RLock()
	radiusEntry.SetText(fmt.Sprintf("%d", t.windowRadius))
	t.paramMutex.RUnlock()

	radiusEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value >= 1 && value <= 20 {
			t.paramMutex.Lock()
			oldValue := t.windowRadius
			t.windowRadius = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Window radius changed: %d -> %d", oldValue, value))
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid window radius: %s (must be 1-20)", text))
		}
	}

	// Epsilon parameter with validation
	epsilonLabel := widget.NewLabel("Epsilon (0.001-1.0):")
	epsilonEntry := widget.NewEntry()

	t.paramMutex.RLock()
	epsilonEntry.SetText(fmt.Sprintf("%.3f", t.epsilon))
	t.paramMutex.RUnlock()

	epsilonEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value > 0.001 && value <= 1.0 {
			t.paramMutex.Lock()
			oldValue := t.epsilon
			t.epsilon = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Epsilon changed: %.3f -> %.3f", oldValue, value))
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid epsilon: %s (must be 0.001-1.0)", text))
		}
	}

	// Morphological Kernel Size parameter with validation
	kernelLabel := widget.NewLabel("Morphological Kernel Size (1-15, odd):")
	kernelEntry := widget.NewEntry()

	t.paramMutex.RLock()
	kernelEntry.SetText(fmt.Sprintf("%d", t.morphKernelSize))
	t.paramMutex.RUnlock()

	kernelEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value >= 1 && value <= 15 && value%2 == 1 {
			t.paramMutex.Lock()
			oldValue := t.morphKernelSize
			t.morphKernelSize = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Morphological kernel size changed: %d -> %d", oldValue, value))
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid kernel size: %s (must be 1-15 and odd)", text))
		}
	}

	return container.NewVBox(
		radiusLabel, radiusEntry,
		epsilonLabel, epsilonEntry,
		kernelLabel, kernelEntry,
	)
}
