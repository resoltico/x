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
// FIXED: Proper algorithm implementation and memory management
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
	// No resources to cleanup - GoCV MatProfile handles tracking
}

func (t *TwoDOtsu) applyWithScale(src gocv.Mat, scale float64) gocv.Mat {
	// Panic recovery for OpenCV operations
	defer func() {
		if r := recover(); r != nil {
			t.debugImage.LogError(fmt.Errorf("panic in 2D Otsu: %v", r))
		}
	}()

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

	// Scale input if needed - GoCV handles memory tracking
	var workingImage gocv.Mat
	if scale != 1.0 {
		newWidth := int(float64(src.Cols()) * scale)
		newHeight := int(float64(src.Rows()) * scale)

		// Input validation
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
	t.debugImage.LogColorConversion("Input", "Grayscale")
	var grayscale gocv.Mat

	if workingImage.Channels() > 1 {
		grayscale = gocv.NewMat()
		gocv.CvtColor(workingImage, &grayscale, gocv.ColorBGRToGray)
		defer grayscale.Close()
	} else {
		grayscale = workingImage.Clone()
		defer grayscale.Close()
	}

	// Validate grayscale conversion
	if grayscale.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Grayscale conversion failed")
		return gocv.NewMat()
	}

	t.debugImage.LogMatInfo("grayscale", grayscale)
	t.debugImage.LogHistogramAnalysis("input_grayscale", grayscale)

	// Apply guided filter for smoother guided image
	guided := t.applyGuidedFilter(grayscale, windowRadius, epsilon)
	defer guided.Close()

	// Validate guided filter result
	if guided.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Guided filter failed, using original")
		guided = grayscale.Clone()
		defer guided.Close()
	}

	t.debugImage.LogHistogramAnalysis("guided_filter", guided)

	// Apply 2D Otsu thresholding
	binaryResult := t.apply2DOtsu(grayscale, guided)
	defer func() {
		if scale != 1.0 && !binaryResult.Empty() {
			binaryResult.Close()
		}
	}()

	// Validate binarization result
	if binaryResult.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Binarization failed")
		return gocv.NewMat()
	}

	// Post-processing with morphological operations
	t.debugImage.LogAlgorithmStep("2D Otsu", "Postprocessing")
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

	// Final validation
	t.debugImage.LogMatDataValidation("final_result", result)
	t.debugImage.LogMatInfo("final_binary", result)
	t.debugImage.LogPixelDistributionDetailed("final_output", result, 5)
	t.debugImage.LogAlgorithmStep("2D Otsu", "Completed")

	return result
}

// FIXED: Proper memory management for guided filter
func (t *TwoDOtsu) applyGuidedFilter(src gocv.Mat, windowRadius int, epsilon float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("GuidedFilter", "Starting guided filter implementation")

	// Validate input
	if src.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: Input is empty")
		return gocv.NewMat()
	}

	// Convert to float32 for processing
	srcFloat := gocv.NewMat()
	defer srcFloat.Close()
	src.ConvertTo(&srcFloat, gocv.MatTypeCV32F)
	srcFloat.DivideFloat(255.0)

	// Parameters
	kernelSize := 2*windowRadius + 1

	// Mean filters - FIXED: Proper cleanup
	meanI := gocv.NewMat()
	defer meanI.Close()
	gocv.BoxFilter(srcFloat, &meanI, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanII := gocv.NewMat()
	defer meanII.Close()
	srcSquared := gocv.NewMat()
	defer srcSquared.Close() // FIXED: Added missing defer
	gocv.Multiply(srcFloat, srcFloat, &srcSquared)
	gocv.BoxFilter(srcSquared, &meanII, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Variance calculation - FIXED: Proper cleanup
	varI := gocv.NewMat()
	defer varI.Close()
	meanISquared := gocv.NewMat()
	defer meanISquared.Close() // FIXED: Added missing defer
	gocv.Multiply(meanI, meanI, &meanISquared)
	gocv.Subtract(meanII, meanISquared, &varI)

	// Calculate coefficients - FIXED: Proper cleanup
	a := gocv.NewMat()
	defer a.Close()
	denominator := gocv.NewMat()
	defer denominator.Close() // FIXED: Added missing defer

	varI.CopyTo(&denominator)
	denominator.AddFloat(float32(epsilon))
	gocv.Divide(varI, denominator, &a)

	b := gocv.NewMat()
	defer b.Close()
	temp := gocv.NewMat()
	defer temp.Close() // FIXED: Added missing defer
	gocv.Multiply(a, meanI, &temp)
	gocv.Subtract(meanI, temp, &b)

	// Smooth coefficients - FIXED: Proper cleanup
	meanA := gocv.NewMat()
	defer meanA.Close()
	gocv.BoxFilter(a, &meanA, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanB := gocv.NewMat()
	defer meanB.Close()
	gocv.BoxFilter(b, &meanB, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Final result - FIXED: Proper cleanup
	resultFloat := gocv.NewMat()
	defer resultFloat.Close()
	temp1 := gocv.NewMat()
	defer temp1.Close() // FIXED: Added missing defer
	gocv.Multiply(meanA, srcFloat, &temp1)
	gocv.Add(temp1, meanB, &resultFloat)

	// Convert back to uint8
	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)

	t.debugImage.LogFilter("GuidedFilter", fmt.Sprintf("radius=%d epsilon=%.3f", windowRadius, epsilon))
	return result
}

// FIXED: Correct 2D Otsu algorithm implementation
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

	// Find optimal thresholds using correct 2D Otsu
	t.debugImage.LogAlgorithmStep("2D Otsu", "Finding optimal thresholds")
	bestS, bestT, maxVariance := t.findOptimalThresholds(hist, len(grayData))

	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	// Apply thresholding with FIXED classification logic
	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarizing image with correct 2D Otsu logic")

	size := gray.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	foregroundCount := 0
	backgroundCount := 0

	// Process pixel by pixel using FIXED 2D Otsu classification
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grayVal := int(gray.GetUCharAt(y, x))
			guidedVal := int(guided.GetUCharAt(y, x))

			// FIXED: Correct 2D Otsu classification logic
			isBackground := t.classifyPixelCorrect(grayVal, guidedVal, bestS, bestT)

			if isBackground {
				result.SetUCharAt(y, x, 255) // Background (white)
				backgroundCount++
			} else {
				result.SetUCharAt(y, x, 0) // Foreground (black)
				foregroundCount++
			}
		}
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Binarization result: %d foreground, %d background pixels", foregroundCount, backgroundCount))

	// Validate the result with enhanced debugging
	t.debugImage.LogMatDataValidation("binary_result", result)
	t.debugImage.LogBinarizationResult("gray+guided", "binary", gray, result, bestS, bestT)

	return result
}

// FIXED: Correct 2D Otsu classification with proper mathematical foundation
func (t *TwoDOtsu) classifyPixelCorrect(grayVal, guidedVal, bestS, bestT int) bool {
	// FIXED: Proper 2D Otsu classification based on statistical regions
	// The 2D histogram is divided into 4 quadrants by thresholds (bestS, bestT)
	//
	// Quadrant classification:
	// Q1: g <= bestS, f <= bestT (typically foreground/object)
	// Q2: g > bestS,  f <= bestT (edge/transition region)
	// Q3: g <= bestS, f > bestT  (edge/transition region)
	// Q4: g > bestS,  f > bestT  (typically background)

	if grayVal <= bestS && guidedVal <= bestT {
		// Q1: Low intensity, low guided value -> likely foreground/text
		return false
	} else if grayVal > bestS && guidedVal > bestT {
		// Q4: High intensity, high guided value -> likely background
		return true
	} else {
		// Q2 or Q3: Transition regions
		// Use statistical distance to closest well-defined region

		// Distance to foreground center (Q1 center)
		q1CenterG := bestS / 2
		q1CenterF := bestT / 2
		distToForeground := math.Sqrt(float64((grayVal-q1CenterG)*(grayVal-q1CenterG) +
			(guidedVal-q1CenterF)*(guidedVal-q1CenterF)))

		// Distance to background center (Q4 center)
		q4CenterG := (bestS + 255) / 2
		q4CenterF := (bestT + 255) / 2
		distToBackground := math.Sqrt(float64((grayVal-q4CenterG)*(grayVal-q4CenterG) +
			(guidedVal-q4CenterF)*(guidedVal-q4CenterF)))

		// Classify based on shorter distance
		return distToBackground < distToForeground
	}
}

// FIXED: Correct between-class variance calculation for 2D Otsu
func (t *TwoDOtsu) findOptimalThresholds(hist [256][256]int, totalPixels int) (int, int, float64) {
	maxVariance := 0.0
	bestS, bestT := 0, 0

	// Exhaustive search for optimal thresholds
	for s := 1; s < 255; s++ {
		for threshold := 1; threshold < 255; threshold++ {
			variance := t.calculateBetweenClassVariance2D(hist, s, threshold, totalPixels)
			if variance > maxVariance {
				maxVariance = variance
				bestS = s
				bestT = threshold
			}
		}
	}

	return bestS, bestT, maxVariance
}

// FIXED: Proper 2D between-class variance calculation
func (t *TwoDOtsu) calculateBetweenClassVariance2D(hist [256][256]int, s, threshold, totalPixels int) float64 {
	// Calculate class statistics for 4 regions
	var w [4]int       // weights (pixel counts)
	var muG [4]float64 // mean gray values
	var muF [4]float64 // mean guided values

	// Region 0: g <= s, f <= t (foreground)
	for g := 0; g <= s; g++ {
		for f := 0; f <= threshold; f++ {
			count := hist[g][f]
			w[0] += count
			muG[0] += float64(g * count)
			muF[0] += float64(f * count)
		}
	}

	// Region 1: g > s, f <= t (edge region 1)
	for g := s + 1; g < 256; g++ {
		for f := 0; f <= threshold; f++ {
			count := hist[g][f]
			w[1] += count
			muG[1] += float64(g * count)
			muF[1] += float64(f * count)
		}
	}

	// Region 2: g <= s, f > t (edge region 2)
	for g := 0; g <= s; g++ {
		for f := threshold + 1; f < 256; f++ {
			count := hist[g][f]
			w[2] += count
			muG[2] += float64(g * count)
			muF[2] += float64(f * count)
		}
	}

	// Region 3: g > s, f > t (background)
	for g := s + 1; g < 256; g++ {
		for f := threshold + 1; f < 256; f++ {
			count := hist[g][f]
			w[3] += count
			muG[3] += float64(g * count)
			muF[3] += float64(f * count)
		}
	}

	// Normalize means and calculate probabilities
	var p [4]float64
	for i := 0; i < 4; i++ {
		if w[i] > 0 {
			muG[i] /= float64(w[i])
			muF[i] /= float64(w[i])
			p[i] = float64(w[i]) / float64(totalPixels)
		}
	}

	// Calculate overall means
	muGTotal := 0.0
	muFTotal := 0.0
	for i := 0; i < 4; i++ {
		muGTotal += p[i] * muG[i]
		muFTotal += p[i] * muF[i]
	}

	// Calculate between-class variance (trace of between-class scatter matrix)
	betweenClassVariance := 0.0
	for i := 0; i < 4; i++ {
		if p[i] > 0 {
			diffG := muG[i] - muGTotal
			diffF := muF[i] - muFTotal
			// Trace of scatter matrix = sum of squared distances
			betweenClassVariance += p[i] * (diffG*diffG + diffF*diffF)
		}
	}

	return betweenClassVariance
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
				// FIXED: Use goroutine to prevent UI thread blocking
				go func() {
					t.onParameterChanged()
				}()
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
				// FIXED: Use goroutine to prevent UI thread blocking
				go func() {
					t.onParameterChanged()
				}()
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
				// FIXED: Use goroutine to prevent UI thread blocking
				go func() {
					t.onParameterChanged()
				}()
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
