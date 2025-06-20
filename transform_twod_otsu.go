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
// Uses GoCV's MatProfile for memory management instead of custom tracking
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

	// Mean filters
	meanI := gocv.NewMat()
	defer meanI.Close()
	gocv.BoxFilter(srcFloat, &meanI, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanII := gocv.NewMat()
	defer meanII.Close()
	srcSquared := gocv.NewMat()
	defer srcSquared.Close()
	gocv.Multiply(srcFloat, srcFloat, &srcSquared)
	gocv.BoxFilter(srcSquared, &meanII, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Variance calculation
	varI := gocv.NewMat()
	defer varI.Close()
	meanISquared := gocv.NewMat()
	defer meanISquared.Close()
	gocv.Multiply(meanI, meanI, &meanISquared)
	gocv.Subtract(meanII, meanISquared, &varI)

	// Calculate coefficients
	a := gocv.NewMat()
	defer a.Close()
	denominator := gocv.NewMat()
	defer denominator.Close()

	varI.CopyTo(&denominator)
	denominator.AddFloat(float32(epsilon))
	gocv.Divide(varI, denominator, &a)

	b := gocv.NewMat()
	defer b.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	gocv.Multiply(a, meanI, &temp)
	gocv.Subtract(meanI, temp, &b)

	// Smooth coefficients
	meanA := gocv.NewMat()
	defer meanA.Close()
	gocv.BoxFilter(a, &meanA, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanB := gocv.NewMat()
	defer meanB.Close()
	gocv.BoxFilter(b, &meanB, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Final result
	resultFloat := gocv.NewMat()
	defer resultFloat.Close()
	temp1 := gocv.NewMat()
	defer temp1.Close()
	gocv.Multiply(meanA, srcFloat, &temp1)
	gocv.Add(temp1, meanB, &resultFloat)

	// Convert back to uint8
	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)

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

	// Apply thresholding with corrected region-based classification
	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarizing image with corrected 2D Otsu logic")

	size := gray.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	foregroundCount := 0
	backgroundCount := 0

	// Process pixel by pixel using corrected 2D Otsu classification
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grayVal := int(gray.GetUCharAt(y, x))
			guidedVal := int(guided.GetUCharAt(y, x))

			// Corrected 2D Otsu classification logic
			isBackground := t.classifyPixel(grayVal, guidedVal, bestS, bestT)

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

// Proper 2D Otsu classification with region-based logic
func (t *TwoDOtsu) classifyPixel(grayVal, guidedVal, bestS, bestT int) bool {
	// 2D Otsu defines 4 regions in the 2D histogram:
	// Region 1: g <= bestS, f <= bestT (Object/Foreground)
	// Region 2: g > bestS, f <= bestT (Transition)
	// Region 3: g <= bestS, f > bestT (Transition)
	// Region 4: g > bestS, f > bestT (Background)

	if grayVal <= bestS && guidedVal <= bestT {
		return false // Object region - foreground
	} else if grayVal > bestS && guidedVal > bestT {
		return true // Background region
	} else {
		// Transition regions - use distance metric for better classification
		objDistance := math.Sqrt(float64((grayVal-bestS/2)*(grayVal-bestS/2) +
			(guidedVal-bestT/2)*(guidedVal-bestT/2)))
		bgDistance := math.Sqrt(float64((grayVal-(bestS+255)/2)*(grayVal-(bestS+255)/2) +
			(guidedVal-(bestT+255)/2)*(guidedVal-(bestT+255)/2)))
		return bgDistance < objDistance
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
