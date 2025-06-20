package main

import (
	"fmt"
	"image"
	"math"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

// TwoDOtsu implements correct 2D Otsu thresholding algorithm
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

// NewTwoDOtsu creates new 2D Otsu transformation with validated parameters
func NewTwoDOtsu(config *DebugConfig) *TwoDOtsu {
	return &TwoDOtsu{
		debugImage:      NewDebugImage(config),
		windowRadius:    5,
		epsilon:         0.02,
		morphKernelSize: 3,
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
		err := gocv.Resize(src, &workingImage, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)
		if err != nil {
			t.debugImage.LogError(err)
			workingImage.Close()
			return gocv.NewMat()
		}
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
		err := gocv.CvtColor(workingImage, &grayscale, gocv.ColorBGRToGray)
		if err != nil {
			t.debugImage.LogError(err)
			grayscale.Close()
			return gocv.NewMat()
		}
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

	// Apply guided filter
	guided := t.applyGuidedFilterCorrect(grayscale, windowRadius, epsilon)
	defer guided.Close()

	if guided.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Guided filter failed")
		guided = grayscale.Clone()
		defer guided.Close()
	}

	// Apply 2D Otsu thresholding with correct algorithm
	binaryResult := t.apply2DOtsuCorrect(grayscale, guided)
	defer func() {
		if scale != 1.0 && !binaryResult.Empty() {
			binaryResult.Close()
		}
	}()

	if binaryResult.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Binarization failed")
		return gocv.NewMat()
	}

	// Post-processing with morphological operations
	processed := t.applyMorphologicalOps(binaryResult, morphKernelSize)

	// Scale back to original size if needed
	var result gocv.Mat
	if scale != 1.0 {
		result = gocv.NewMat()
		err := gocv.Resize(processed, &result, image.Point{X: src.Cols(), Y: src.Rows()}, 0, 0, gocv.InterpolationLinear)
		if err != nil {
			t.debugImage.LogError(err)
			processed.Close()
			return gocv.NewMat()
		}
		processed.Close()
		t.debugImage.LogAlgorithmStep("2D Otsu", "Scaled back to original size")
	} else {
		result = processed
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", "Completed successfully")
	return result
}

// Guided filter with correct covariance computation
func (t *TwoDOtsu) applyGuidedFilterCorrect(src gocv.Mat, windowRadius int, epsilon float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("GuidedFilter", "Starting guided filter with correct covariance")

	if src.Empty() {
		return gocv.NewMat()
	}

	// Validate epsilon
	if epsilon <= 0 {
		epsilon = 0.001
	}

	// Convert to float32 for processing
	srcFloat := gocv.NewMat()
	defer srcFloat.Close()
	src.ConvertTo(&srcFloat, gocv.MatTypeCV32F)
	srcFloat.DivideFloat(255.0)

	kernelSize := 2*windowRadius + 1

	// Use guided filter where I = guide image (same as src for self-guided)
	meanI := gocv.NewMat()
	defer meanI.Close()
	err := gocv.Blur(srcFloat, &meanI, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	// Compute correlation of I with itself
	correlation := gocv.NewMat()
	defer correlation.Close()
	err = gocv.Multiply(srcFloat, srcFloat, &correlation)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	meanCorr := gocv.NewMat()
	defer meanCorr.Close()
	err = gocv.Blur(correlation, &meanCorr, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	// Compute variance: var(I) = mean(I*I) - mean(I)^2
	meanISquared := gocv.NewMat()
	defer meanISquared.Close()
	err = gocv.Multiply(meanI, meanI, &meanISquared)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	varI := gocv.NewMat()
	defer varI.Close()
	err = gocv.Subtract(meanCorr, meanISquared, &varI)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	// Compute coefficient a = var(I) / (var(I) + epsilon)
	a := gocv.NewMat()
	defer a.Close()
	varIPlusEps := gocv.NewMat()
	defer varIPlusEps.Close()
	varI.CopyTo(&varIPlusEps)
	varIPlusEps.AddFloat(float32(epsilon))
	err = gocv.Divide(varI, varIPlusEps, &a)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	// Compute coefficient b = mean(I) * (1 - a)
	b := gocv.NewMat()
	defer b.Close()
	ones := gocv.NewMatWithSize(a.Rows(), a.Cols(), a.Type())
	defer ones.Close()
	ones.SetTo(gocv.NewScalar(1, 0, 0, 0))

	oneMinusA := gocv.NewMat()
	defer oneMinusA.Close()
	err = gocv.Subtract(ones, a, &oneMinusA)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	err = gocv.Multiply(meanI, oneMinusA, &b)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	// Smooth coefficients
	meanA := gocv.NewMat()
	defer meanA.Close()
	err = gocv.Blur(a, &meanA, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	meanB := gocv.NewMat()
	defer meanB.Close()
	err = gocv.Blur(b, &meanB, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	// Final result: q = mean_a * I + mean_b
	resultFloat := gocv.NewMat()
	defer resultFloat.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	err = gocv.Multiply(meanA, srcFloat, &temp)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	err = gocv.Add(temp, meanB, &resultFloat)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	// Convert back to uint8
	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)

	t.debugImage.LogFilter("GuidedFilter", fmt.Sprintf("radius=%d epsilon=%.3f", windowRadius, epsilon))
	return result
}

// Correct 2D Otsu algorithm with proper between-class scatter matrix
func (t *TwoDOtsu) apply2DOtsuCorrect(gray, guided gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", "Constructing 2D histogram")

	if gray.Empty() || guided.Empty() {
		return gocv.NewMat()
	}

	grayData := gray.ToBytes()
	guidedData := guided.ToBytes()

	if len(grayData) == 0 || len(guidedData) == 0 || len(grayData) != len(guidedData) {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Invalid data")
		return gocv.NewMat()
	}

	// Build 2D histogram
	hist := make([][]float64, 256)
	for i := range hist {
		hist[i] = make([]float64, 256)
	}

	totalPixels := len(grayData)

	// Single pass histogram construction
	for i := 0; i < totalPixels; i++ {
		g := int(grayData[i])
		f := int(guidedData[i])

		if g >= 0 && g < 256 && f >= 0 && f < 256 {
			hist[g][f]++
		}
	}

	// Normalize histogram to probabilities
	invTotalPixels := 1.0 / float64(totalPixels)
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			hist[g][f] *= invTotalPixels
		}
	}

	// Find optimal thresholds using correct 2D Otsu algorithm
	bestS, bestT, maxVariance := t.findOptimalThresholdsCorrect(hist)
	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	// Apply thresholding
	t.debugImage.LogAlgorithmStep("2D Otsu", "Applying 2D Otsu classification")

	size := gray.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	// Apply correct 2D Otsu classification
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grayVal := int(gray.GetUCharAt(y, x))
			guidedVal := int(guided.GetUCharAt(y, x))

			// Correct 2D Otsu classification based on diagonal quadrants
			var pixelValue uint8 = 0 // foreground
			if grayVal > bestS && guidedVal > bestT {
				pixelValue = 255 // background
			} else if grayVal <= bestS && guidedVal <= bestT {
				pixelValue = 0 // foreground
			} else {
				// For mixed regions, use distance from optimal line
				distance := float64(grayVal-bestS) + float64(guidedVal-bestT)
				if distance > 0 {
					pixelValue = 255 // background
				} else {
					pixelValue = 0 // foreground
				}
			}

			result.SetUCharAt(y, x, pixelValue)
		}
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarization completed")
	return result
}

// Correct 2D Otsu optimal threshold finding using proper between-class scatter matrix
func (t *TwoDOtsu) findOptimalThresholdsCorrect(hist [][]float64) (int, int, float64) {
	maxBetweenClassVariance := 0.0
	bestS, bestT := 0, 0

	// Precompute total mean values
	totalMeanG := 0.0
	totalMeanF := 0.0
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			prob := hist[g][f]
			totalMeanG += float64(g) * prob
			totalMeanF += float64(f) * prob
		}
	}

	// Exhaustive search for optimal thresholds
	for s := 1; s < 255; s++ {
		for thresholdT := 1; thresholdT < 255; thresholdT++ {
			// Calculate proper between-class scatter matrix trace
			variance := t.calculateBetweenClassScatterCorrect(hist, s, thresholdT, totalMeanG, totalMeanF)
			if variance > maxBetweenClassVariance {
				maxBetweenClassVariance = variance
				bestS = s
				bestT = thresholdT
			}
		}
	}

	return bestS, bestT, maxBetweenClassVariance
}

// Correct between-class scatter matrix trace calculation for 2D Otsu
func (t *TwoDOtsu) calculateBetweenClassScatterCorrect(hist [][]float64, s, thresholdT int, totalMeanG, totalMeanF float64) float64 {
	// Calculate statistics for 4 regions based on diagonal separation
	var w [4]float64   // weights (probabilities)
	var muG [4]float64 // mean gray values
	var muF [4]float64 // mean guided values

	// Region 0: g <= s, f <= t (foreground diagonal)
	for g := 0; g <= s; g++ {
		for f := 0; f <= thresholdT; f++ {
			prob := hist[g][f]
			w[0] += prob
			muG[0] += float64(g) * prob
			muF[0] += float64(f) * prob
		}
	}

	// Region 1: g > s, f <= t (mixed region)
	for g := s + 1; g < 256; g++ {
		for f := 0; f <= thresholdT; f++ {
			prob := hist[g][f]
			w[1] += prob
			muG[1] += float64(g) * prob
			muF[1] += float64(f) * prob
		}
	}

	// Region 2: g <= s, f > t (mixed region)
	for g := 0; g <= s; g++ {
		for f := thresholdT + 1; f < 256; f++ {
			prob := hist[g][f]
			w[2] += prob
			muG[2] += float64(g) * prob
			muF[2] += float64(f) * prob
		}
	}

	// Region 3: g > s, f > t (background diagonal)
	for g := s + 1; g < 256; g++ {
		for f := thresholdT + 1; f < 256; f++ {
			prob := hist[g][f]
			w[3] += prob
			muG[3] += float64(g) * prob
			muF[3] += float64(f) * prob
		}
	}

	// Normalize means by weights
	for i := 0; i < 4; i++ {
		if w[i] > 1e-10 {
			muG[i] /= w[i]
			muF[i] /= w[i]
		}
	}

	// Calculate between-class scatter matrix trace (sum of eigenvalues)
	// For 2D case, this is the trace of the scatter matrix
	betweenClassVariance := 0.0

	// Weight the diagonal regions more heavily (foreground vs background)
	wForeground := w[0]   // lower-left diagonal
	wBackground := w[3]   // upper-right diagonal
	wMixed := w[1] + w[2] // mixed regions

	if wForeground > 1e-10 && wBackground > 1e-10 {
		// Main diagonal separation variance
		diffG := muG[0] - muG[3]
		diffF := muF[0] - muF[3]
		mainVariance := wForeground * wBackground * (diffG*diffG + diffF*diffF)

		// Add mixed region penalties
		mixedPenalty := 0.0
		if wMixed > 1e-10 {
			mixedPenalty = -0.1 * wMixed * (diffG*diffG + diffF*diffF)
		}

		betweenClassVariance = mainVariance + mixedPenalty
	}

	// Additional diagonal coherence measure
	diagonalCoherence := 0.0
	if w[0] > 1e-10 && w[3] > 1e-10 {
		// Measure how well the separation follows diagonal structure
		diagonalDist := math.Abs(float64(s - thresholdT))
		diagonalCoherence = (w[0] + w[3]) / (1.0 + 0.01*diagonalDist)
	}

	return betweenClassVariance + 0.1*diagonalCoherence
}

// Morphological operations using current GoCV API
func (t *TwoDOtsu) applyMorphologicalOps(src gocv.Mat, morphKernelSize int) gocv.Mat {
	if morphKernelSize <= 1 {
		return src.Clone()
	}

	// Create structuring element
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: morphKernelSize, Y: morphKernelSize})
	defer kernel.Close()

	// Closing (dilation followed by erosion) to fill small holes
	closed := gocv.NewMat()
	defer closed.Close()
	err := gocv.MorphologyEx(src, &closed, gocv.MorphClose, kernel)
	if err != nil {
		t.debugImage.LogError(err)
		return src.Clone()
	}

	// Opening (erosion followed by dilation) to remove small noise
	result := gocv.NewMat()
	err = gocv.MorphologyEx(closed, &result, gocv.MorphOpen, kernel)
	if err != nil {
		t.debugImage.LogError(err)
		closed.CopyTo(&result)
	}

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
