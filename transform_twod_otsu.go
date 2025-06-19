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

// TwoDOtsu implements 2D Otsu thresholding with guided filtering
type TwoDOtsu struct {
	ThreadSafeTransformation
	debugImage *DebugImage

	// Thread-safe parameters
	paramMutex      sync.RWMutex
	windowRadius    int
	epsilon         float64
	morphKernelSize int

	// Callback for parameter changes
	onParameterChanged func()
}

// NewTwoDOtsu creates a new 2D Otsu transformation
func NewTwoDOtsu(config *DebugConfig) *TwoDOtsu {
	return &TwoDOtsu{
		debugImage:      NewDebugImage(config),
		windowRadius:    5,
		epsilon:         0.02,
		morphKernelSize: 1,
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
		t.windowRadius = radius
	}
	if eps, ok := params["epsilon"].(float64); ok {
		t.epsilon = eps
	}
	if kernel, ok := params["morphKernelSize"].(int); ok {
		t.morphKernelSize = kernel
	}
}

func (t *TwoDOtsu) Apply(src gocv.Mat) gocv.Mat {
	return t.applyWithScale(src, 1.0)
}

func (t *TwoDOtsu) ApplyPreview(src gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Starting preview application")
	result := t.applyWithScale(src, 0.5)
	t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Preview processing completed")
	return result
}

func (t *TwoDOtsu) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	t.onParameterChanged = onParameterChanged
	return t.createParameterUI()
}

func (t *TwoDOtsu) Close() {
	// No resources to cleanup
}

// FIXED: Safe memory management without defer overwrite patterns
func (t *TwoDOtsu) applyWithScale(src gocv.Mat, scale float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Starting binarization (scale: %.2f)", scale))

	// Validate input
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

	// SAFE: Create working image without defer overwrite risk
	var workingImage gocv.Mat
	var workingImageOwned bool = false

	if scale != 1.0 {
		newWidth := int(float64(src.Cols()) * scale)
		newHeight := int(float64(src.Rows()) * scale)

		// Input validation
		if newWidth <= 0 || newHeight <= 0 || newWidth > 16384 || newHeight > 16384 {
			t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("ERROR: Invalid scaled dimensions: %dx%d", newWidth, newHeight))
			return gocv.NewMat()
		}

		workingImage = gocv.NewMat()
		workingImageOwned = true
		gocv.Resize(src, &workingImage, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)
		t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Scaled to %dx%d", newWidth, newHeight))
	} else {
		workingImage = src.Clone()
		workingImageOwned = true
	}

	// SAFE: Explicit cleanup function for working image
	cleanupWorkingImage := func() {
		if workingImageOwned && !workingImage.Empty() {
			workingImage.Close()
			workingImageOwned = false
		}
	}

	// Convert to grayscale if needed - SAFE: explicit ownership management
	t.debugImage.LogColorConversion("Input", "Grayscale")
	var grayscale gocv.Mat
	var grayscaleOwned bool = false

	if workingImage.Channels() > 1 {
		grayscale = gocv.NewMat()
		grayscaleOwned = true
		gocv.CvtColor(workingImage, &grayscale, gocv.ColorBGRToGray)
	} else {
		grayscale = workingImage.Clone()
		grayscaleOwned = true
	}

	// SAFE: Cleanup function for grayscale
	cleanupGrayscale := func() {
		if grayscaleOwned && !grayscale.Empty() {
			grayscale.Close()
			grayscaleOwned = false
		}
	}

	// Validate grayscale conversion
	if grayscale.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Grayscale conversion failed")
		cleanupWorkingImage()
		cleanupGrayscale()
		return gocv.NewMat()
	}

	t.debugImage.LogMatInfo("grayscale", grayscale)
	t.debugImage.LogHistogramAnalysis("input_grayscale", grayscale)

	// Apply guided filter for smoother guided image
	guided := t.applyGuidedFilter(grayscale, windowRadius, epsilon)
	guidedOwned := true

	// SAFE: Cleanup function for guided
	cleanupGuided := func() {
		if guidedOwned && !guided.Empty() {
			guided.Close()
			guidedOwned = false
		}
	}

	// Validate guided filter result
	if guided.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Guided filter failed, using original")
		cleanupGuided()
		guided = grayscale.Clone()
		guidedOwned = true
	}

	t.debugImage.LogHistogramAnalysis("guided_filter", guided)

	// Apply 2D Otsu thresholding
	binaryResult := t.apply2DOtsu(grayscale, guided)
	binaryOwned := true

	// SAFE: Cleanup function for binary result
	cleanupBinary := func() {
		if binaryOwned && !binaryResult.Empty() {
			binaryResult.Close()
			binaryOwned = false
		}
	}

	// Validate binarization result
	if binaryResult.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Binarization failed")
		cleanupWorkingImage()
		cleanupGrayscale()
		cleanupGuided()
		cleanupBinary()
		return gocv.NewMat()
	}

	// Post-processing with morphological operations
	t.debugImage.LogAlgorithmStep("2D Otsu", "Postprocessing")
	processed := t.applyMorphologicalOps(binaryResult, morphKernelSize)
	processedOwned := true

	// SAFE: Cleanup function for processed
	cleanupProcessed := func() {
		if processedOwned && !processed.Empty() {
			processed.Close()
			processedOwned = false
		}
	}

	// SAFE: Handle scaling back if needed
	var finalResult gocv.Mat
	if scale != 1.0 {
		finalResult = gocv.NewMat()
		gocv.Resize(processed, &finalResult, image.Point{X: src.Cols(), Y: src.Rows()}, 0, 0, gocv.InterpolationLinear)
		t.debugImage.LogAlgorithmStep("2D Otsu", "Scaled back to original size")

		// Clean up intermediate results since we have a new final result
		cleanupProcessed()
	} else {
		// Transfer ownership to final result
		finalResult = processed
		processedOwned = false // Don't clean up - ownership transferred
	}

	// SAFE: Clean up all intermediate results
	cleanupWorkingImage()
	cleanupGrayscale()
	cleanupGuided()
	cleanupBinary()
	cleanupProcessed()

	// Final validation
	t.debugImage.LogMatDataValidation("final_result", finalResult)
	t.debugImage.LogMatInfo("final_binary", finalResult)
	t.debugImage.LogPixelDistributionDetailed("final_output", finalResult, 5)
	t.debugImage.LogAlgorithmStep("2D Otsu", "Completed")

	return finalResult
}

// FIXED: Safe guided filter implementation without defer overwrite patterns
func (t *TwoDOtsu) applyGuidedFilter(src gocv.Mat, windowRadius int, epsilon float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("GuidedFilter", "Starting guided filter implementation")
	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Input: %dx%d, channels=%d", src.Cols(), src.Rows(), src.Channels()))

	// Validate input
	if src.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Input is empty")
		return gocv.NewMat()
	}

	// SAFE: Convert to float32 for processing
	srcFloat := gocv.NewMat()
	src.ConvertTo(&srcFloat, gocv.MatTypeCV32F)
	srcFloat.DivideFloat(255.0)

	// SAFE: Cleanup function for srcFloat
	cleanupSrcFloat := func() {
		if !srcFloat.Empty() {
			srcFloat.Close()
		}
	}

	// Validate conversion
	if srcFloat.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Float conversion failed")
		cleanupSrcFloat()
		return src.Clone()
	}

	// Parameters
	kernelSize := 2*windowRadius + 1
	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Using box filter with kernel %dx%d", kernelSize, kernelSize))

	// SAFE: Calculate mean of input image I
	meanI := gocv.NewMat()
	gocv.BoxFilter(srcFloat, &meanI, -1, image.Point{X: kernelSize, Y: kernelSize})

	// SAFE: Cleanup function for meanI
	cleanupMeanI := func() {
		if !meanI.Empty() {
			meanI.Close()
		}
	}

	// Validate mean filter
	if meanI.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Mean filter failed")
		cleanupSrcFloat()
		cleanupMeanI()
		return src.Clone()
	}

	// SAFE: Calculate mean of I²
	srcSquared := gocv.NewMat()
	gocv.Multiply(srcFloat, srcFloat, &srcSquared)

	// SAFE: Cleanup function for srcSquared
	cleanupSrcSquared := func() {
		if !srcSquared.Empty() {
			srcSquared.Close()
		}
	}

	meanII := gocv.NewMat()
	gocv.BoxFilter(srcSquared, &meanII, -1, image.Point{X: kernelSize, Y: kernelSize})

	// SAFE: Cleanup function for meanII
	cleanupMeanII := func() {
		if !meanII.Empty() {
			meanII.Close()
		}
	}

	// Validate multiplication and mean filter
	if srcSquared.Empty() || meanII.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Squared mean calculation failed")
		cleanupSrcFloat()
		cleanupMeanI()
		cleanupSrcSquared()
		cleanupMeanII()
		return src.Clone()
	}

	// FIXED: Calculate variance properly: Var(I) = E[I²] - (E[I])²
	varI := gocv.NewMat()
	meanISquared := gocv.NewMat()
	gocv.Multiply(meanI, meanI, &meanISquared)
	gocv.Subtract(meanII, meanISquared, &varI)

	// SAFE: Cleanup functions for variance calculation
	cleanupVarI := func() {
		if !varI.Empty() {
			varI.Close()
		}
	}
	cleanupMeanISquared := func() {
		if !meanISquared.Empty() {
			meanISquared.Close()
		}
	}

	// Validate variance calculation
	if varI.Empty() || meanISquared.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Variance calculation failed")
		cleanupSrcFloat()
		cleanupMeanI()
		cleanupSrcSquared()
		cleanupMeanII()
		cleanupVarI()
		cleanupMeanISquared()
		return src.Clone()
	}

	// FIXED: For self-guided filter where guide = input, covariance = variance
	a := gocv.NewMat()
	denominator := gocv.NewMat()
	varI.CopyTo(&denominator)
	denominator.AddFloat(float32(epsilon))
	gocv.Divide(varI, denominator, &a)

	// SAFE: Cleanup functions for coefficients
	cleanupA := func() {
		if !a.Empty() {
			a.Close()
		}
	}
	cleanupDenominator := func() {
		if !denominator.Empty() {
			denominator.Close()
		}
	}

	// Validate division
	if a.Empty() || denominator.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Coefficient a calculation failed")
		cleanupSrcFloat()
		cleanupMeanI()
		cleanupSrcSquared()
		cleanupMeanII()
		cleanupVarI()
		cleanupMeanISquared()
		cleanupA()
		cleanupDenominator()
		return src.Clone()
	}

	// Calculate b = mean_p - a * mean_I (for self-guided: mean_p = mean_I)
	b := gocv.NewMat()
	temp := gocv.NewMat()
	gocv.Multiply(a, meanI, &temp)
	gocv.Subtract(meanI, temp, &b)

	// SAFE: Cleanup functions for b calculation
	cleanupB := func() {
		if !b.Empty() {
			b.Close()
		}
	}
	cleanupTemp := func() {
		if !temp.Empty() {
			temp.Close()
		}
	}

	// Validate b calculation
	if b.Empty() || temp.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Coefficient b calculation failed")
		cleanupSrcFloat()
		cleanupMeanI()
		cleanupSrcSquared()
		cleanupMeanII()
		cleanupVarI()
		cleanupMeanISquared()
		cleanupA()
		cleanupDenominator()
		cleanupB()
		cleanupTemp()
		return src.Clone()
	}

	// Smooth coefficients
	meanA := gocv.NewMat()
	gocv.BoxFilter(a, &meanA, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanB := gocv.NewMat()
	gocv.BoxFilter(b, &meanB, -1, image.Point{X: kernelSize, Y: kernelSize})

	// SAFE: Cleanup functions for smoothed coefficients
	cleanupMeanA := func() {
		if !meanA.Empty() {
			meanA.Close()
		}
	}
	cleanupMeanB := func() {
		if !meanB.Empty() {
			meanB.Close()
		}
	}

	// Validate coefficient smoothing
	if meanA.Empty() || meanB.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Coefficient smoothing failed")
		cleanupSrcFloat()
		cleanupMeanI()
		cleanupSrcSquared()
		cleanupMeanII()
		cleanupVarI()
		cleanupMeanISquared()
		cleanupA()
		cleanupDenominator()
		cleanupB()
		cleanupTemp()
		cleanupMeanA()
		cleanupMeanB()
		return src.Clone()
	}

	// Final result: q = mean_a * I + mean_b
	resultFloat := gocv.NewMat()
	temp1 := gocv.NewMat()
	gocv.Multiply(meanA, srcFloat, &temp1)
	gocv.Add(temp1, meanB, &resultFloat)

	// SAFE: Cleanup functions for final calculation
	cleanupResultFloat := func() {
		if !resultFloat.Empty() {
			resultFloat.Close()
		}
	}
	cleanupTemp1 := func() {
		if !temp1.Empty() {
			temp1.Close()
		}
	}

	// Validate final calculations
	if resultFloat.Empty() || temp1.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Final result calculation failed")
		// Cleanup all intermediate results
		cleanupSrcFloat()
		cleanupMeanI()
		cleanupSrcSquared()
		cleanupMeanII()
		cleanupVarI()
		cleanupMeanISquared()
		cleanupA()
		cleanupDenominator()
		cleanupB()
		cleanupTemp()
		cleanupMeanA()
		cleanupMeanB()
		cleanupResultFloat()
		cleanupTemp1()
		return src.Clone()
	}

	// Convert back to uint8
	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)

	// SAFE: Clean up all intermediate results before returning
	cleanupSrcFloat()
	cleanupMeanI()
	cleanupSrcSquared()
	cleanupMeanII()
	cleanupVarI()
	cleanupMeanISquared()
	cleanupA()
	cleanupDenominator()
	cleanupB()
	cleanupTemp()
	cleanupMeanA()
	cleanupMeanB()
	cleanupResultFloat()
	cleanupTemp1()

	// Final validation
	if result.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Final conversion failed")
		return src.Clone()
	}

	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Result: %dx%d, channels=%d", result.Cols(), result.Rows(), result.Channels()))
	t.debugImage.LogFilter("GuidedFilter", fmt.Sprintf("radius=%d epsilon=%.3f", windowRadius, epsilon))

	return result
}

func (t *TwoDOtsu) apply2DOtsu(gray, guided gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", "Constructing 2D histogram")

	// Validate inputs
	if gray.Empty() || guided.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Input matrices are empty")
		return gocv.NewMat()
	}

	grayData := gray.ToBytes()
	guidedData := guided.ToBytes()

	// Validate data extraction
	if len(grayData) == 0 || len(guidedData) == 0 {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: No data extracted from matrices")
		return gocv.NewMat()
	}

	// Ensure data lengths match
	if len(grayData) != len(guidedData) {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Data length mismatch")
		return gocv.NewMat()
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Gray dimensions: %dx%d", gray.Cols(), gray.Rows()))
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Guided dimensions: %dx%d", guided.Cols(), guided.Rows()))
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Processing %d pixels", len(grayData)))

	// Validate input data integrity
	t.debugImage.LogMatDataValidation("gray_input", gray)
	t.debugImage.LogMatDataValidation("guided_input", guided)

	// Build 2D histogram with bounds checking
	var hist [256][256]int
	for i := 0; i < len(grayData); i++ {
		g := int(grayData[i])
		f := int(guidedData[i])

		// Add bounds checking
		if g < 0 {
			g = 0
		} else if g > 255 {
			g = 255
		}
		if f < 0 {
			f = 0
		} else if f > 255 {
			f = 255
		}

		hist[g][f]++
	}

	// Find optimal thresholds using 2D Otsu
	t.debugImage.LogAlgorithmStep("2D Otsu", "Finding optimal thresholds")
	bestS, bestT, maxVariance := t.findOptimalThresholds(hist, len(grayData))

	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	// Create debug output before binarization
	t.debugImage.LogPixelDistributionDetailed("gray_before_binarization", gray, 3)
	t.debugImage.LogPixelDistributionDetailed("guided_before_binarization", guided, 3)

	// FIXED: Apply thresholding with CORRECTED region-based classification
	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarizing image with corrected 2D Otsu logic")

	size := gray.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	foregroundCount := 0
	backgroundCount := 0

	// FIXED: Process pixel by pixel using CORRECTED 2D Otsu classification
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grayVal := int(gray.GetUCharAt(y, x))
			guidedVal := int(guided.GetUCharAt(y, x))

			// CORRECTED 2D Otsu classification logic
			isForeground := t.classifyPixel(grayVal, guidedVal, bestS, bestT)

			if isForeground {
				result.SetUCharAt(y, x, 0) // Foreground (black) - text/objects
				foregroundCount++
			} else {
				result.SetUCharAt(y, x, 255) // Background (white) - paper/background
				backgroundCount++
			}
		}
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Binarization result: %d foreground, %d background pixels", foregroundCount, backgroundCount))

	// Validate the result with enhanced debugging
	t.debugImage.LogMatDataValidation("binary_result", result)
	t.debugImage.LogBinarizationResult("gray+guided", "binary", gray, result, bestS, bestT)

	return result
}

// FIXED: Correct 2D Otsu classification with proper region-based logic for document binarization
func (t *TwoDOtsu) classifyPixel(grayVal, guidedVal, bestS, bestT int) bool {
	// CORRECTED: 2D Otsu classification for document binarization
	// For documents: dark text on light background
	// Region 1: g <= bestS, f <= bestT (Dark objects - TEXT/FOREGROUND)
	// Region 2: g > bestS, f <= bestT (Edge/transition regions)
	// Region 3: g <= bestS, f > bestT (Edge/transition regions)
	// Region 4: g > bestS, f > bestT (Light background - BACKGROUND)

	if grayVal <= bestS && guidedVal <= bestT {
		return true // Region 1: Dark objects (text) - FOREGROUND
	} else if grayVal > bestS && guidedVal > bestT {
		return false // Region 4: Light background - BACKGROUND
	} else {
		// Transition regions (2 & 3) - use distance-based classification
		// Calculate distance to foreground center vs background center
		foregroundCenterG := bestS / 2
		foregroundCenterF := bestT / 2
		backgroundCenterG := (bestS + 255) / 2
		backgroundCenterF := (bestT + 255) / 2

		distToForeground := math.Sqrt(float64((grayVal-foregroundCenterG)*(grayVal-foregroundCenterG) +
			(guidedVal-foregroundCenterF)*(guidedVal-foregroundCenterF)))
		distToBackground := math.Sqrt(float64((grayVal-backgroundCenterG)*(grayVal-backgroundCenterG) +
			(guidedVal-backgroundCenterF)*(guidedVal-backgroundCenterF)))

		return distToForeground < distToBackground // Closer to foreground = foreground
	}
}

func (t *TwoDOtsu) findOptimalThresholds(hist [256][256]int, totalPixels int) (int, int, float64) {
	maxVariance := 0.0
	bestS, bestT := 0, 0

	for s := 0; s < 255; s++ {
		for threshold := 0; threshold < 255; threshold++ {
			variance := t.calculateBetweenClassVariance(hist, s, threshold, totalPixels)
			if variance > maxVariance {
				maxVariance = variance
				bestS = s
				bestT = threshold
			}
		}
	}

	return bestS, bestT, maxVariance
}

func (t *TwoDOtsu) calculateBetweenClassVariance(hist [256][256]int, s, threshold, totalPixels int) float64 {
	// Calculate class probabilities and means
	w0, w1 := 0, 0
	mu0G, mu0F, mu1G, mu1F := 0.0, 0.0, 0.0, 0.0

	// Class 0: g <= s, f <= threshold (foreground - dark text)
	for g := 0; g <= s; g++ {
		for f := 0; f <= threshold; f++ {
			count := hist[g][f]
			w0 += count
			mu0G += float64(g * count)
			mu0F += float64(f * count)
		}
	}

	// Class 1: remaining pixels (background - white paper)
	w1 = totalPixels - w0

	if w0 == 0 || w1 == 0 {
		return 0.0
	}

	mu0G /= float64(w0)
	mu0F /= float64(w0)

	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			if g > s || f > threshold {
				count := hist[g][f]
				mu1G += float64(g * count)
				mu1F += float64(f * count)
			}
		}
	}

	mu1G /= float64(w1)
	mu1F /= float64(w1)

	// Between-class variance
	p0 := float64(w0) / float64(totalPixels)
	p1 := float64(w1) / float64(totalPixels)

	diffG := mu1G - mu0G
	diffF := mu1F - mu0F

	variance := p0 * p1 * (diffG*diffG + diffF*diffF)
	return variance
}

func (t *TwoDOtsu) applyMorphologicalOps(src gocv.Mat, morphKernelSize int) gocv.Mat {
	if morphKernelSize <= 1 {
		t.debugImage.LogMorphology("Close", morphKernelSize)
		t.debugImage.LogMorphology("Open", morphKernelSize)
		return src.Clone()
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: morphKernelSize, Y: morphKernelSize})
	defer kernel.Close()

	// Closing (dilation followed by erosion)
	t.debugImage.LogMorphology("Close", morphKernelSize)
	closed := gocv.NewMat()
	defer closed.Close()
	gocv.MorphologyEx(src, &closed, gocv.MorphClose, kernel)

	// Opening (erosion followed by dilation)
	t.debugImage.LogMorphology("Open", morphKernelSize)
	result := gocv.NewMat()
	gocv.MorphologyEx(closed, &result, gocv.MorphOpen, kernel)

	return result
}

func (t *TwoDOtsu) createParameterUI() *fyne.Container {
	// Window Radius parameter
	radiusLabel := widget.NewLabel("Window Radius (1-20):")
	radiusEntry := widget.NewEntry()

	t.paramMutex.RLock()
	radiusEntry.SetText(fmt.Sprintf("%d", t.windowRadius))
	t.paramMutex.RUnlock()

	radiusEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value > 0 && value <= 20 {
			t.paramMutex.Lock()
			oldValue := t.windowRadius
			t.windowRadius = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Window radius changed: %d -> %d", oldValue, value))
			if t.onParameterChanged != nil {
				t.debugImage.LogAlgorithmStep("2D Otsu Parameters", "Calling onParameterChanged callback")
				// FIXED: Use fyne.Do for thread safety
				fyne.Do(func() {
					t.onParameterChanged()
				})
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid window radius value: %s", text))
		}
	}

	// Epsilon parameter
	epsilonLabel := widget.NewLabel("Epsilon (0.001-1.0):")
	epsilonEntry := widget.NewEntry()

	t.paramMutex.RLock()
	epsilonEntry.SetText(fmt.Sprintf("%.3f", t.epsilon))
	t.paramMutex.RUnlock()

	epsilonEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value > 0 && value <= 1.0 {
			t.paramMutex.Lock()
			oldValue := t.epsilon
			t.epsilon = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Epsilon changed: %.3f -> %.3f", oldValue, value))
			if t.onParameterChanged != nil {
				t.debugImage.LogAlgorithmStep("2D Otsu Parameters", "Calling onParameterChanged callback")
				// FIXED: Use fyne.Do for thread safety
				fyne.Do(func() {
					t.onParameterChanged()
				})
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid epsilon value: %s", text))
		}
	}

	// Morphological Kernel Size parameter
	kernelLabel := widget.NewLabel("Morphological Kernel Size (1-15, odd):")
	kernelEntry := widget.NewEntry()

	t.paramMutex.RLock()
	kernelEntry.SetText(fmt.Sprintf("%d", t.morphKernelSize))
	t.paramMutex.RUnlock()

	kernelEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value > 0 && value <= 15 && value%2 == 1 {
			t.paramMutex.Lock()
			oldValue := t.morphKernelSize
			t.morphKernelSize = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Morphological kernel size changed: %d -> %d", oldValue, value))
			if t.onParameterChanged != nil {
				t.debugImage.LogAlgorithmStep("2D Otsu Parameters", "Calling onParameterChanged callback")
				// FIXED: Use fyne.Do for thread safety
				fyne.Do(func() {
					t.onParameterChanged()
				})
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid morphological kernel size: %s", text))
		}
	}

	return container.NewVBox(
		radiusLabel, radiusEntry,
		epsilonLabel, epsilonEntry,
		kernelLabel, kernelEntry,
	)
}
