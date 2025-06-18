package main

import (
	"fmt"
	"image"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

// TwoDOtsu implements 2D Otsu thresholding with guided filtering
type TwoDOtsu struct {
	debugImage      *DebugImage
	windowRadius    int
	epsilon         float64
	morphKernelSize int

	// Cache for preview optimization
	cacheValid   bool
	cachedParams map[string]interface{}
	cachedResult gocv.Mat

	// Callback for parameter changes
	onParameterChanged func()
}

// NewTwoDOtsu creates a new 2D Otsu transformation
func NewTwoDOtsu() *TwoDOtsu {
	return &TwoDOtsu{
		debugImage:      NewDebugImage(),
		windowRadius:    5,
		epsilon:         0.02,
		morphKernelSize: 3,
		cacheValid:      false,
		cachedParams:    make(map[string]interface{}),
	}
}

func (t *TwoDOtsu) Name() string {
	return "2D Otsu"
}

func (t *TwoDOtsu) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"windowRadius":    t.windowRadius,
		"epsilon":         t.epsilon,
		"morphKernelSize": t.morphKernelSize,
	}
}

func (t *TwoDOtsu) SetParameters(params map[string]interface{}) {
	if radius, ok := params["windowRadius"].(int); ok {
		t.windowRadius = radius
	}
	if eps, ok := params["epsilon"].(float64); ok {
		t.epsilon = eps
	}
	if kernel, ok := params["morphKernelSize"].(int); ok {
		t.morphKernelSize = kernel
	}
	t.invalidateCache()
}

func (t *TwoDOtsu) Apply(src gocv.Mat) gocv.Mat {
	return t.applyWithScale(src, 1.0)
}

func (t *TwoDOtsu) ApplyPreview(src gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Starting preview application")

	// Check cache validity
	currentParams := t.GetParameters()
	paramsEqual := t.compareParams(currentParams)

	t.debugImage.LogAlgorithmStep("2D Otsu Preview", fmt.Sprintf("Cache check: valid=%t, params equal=%t", t.cacheValid, paramsEqual))

	if t.cacheValid && paramsEqual && !t.cachedResult.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Cache hit - returning cached result")
		result := gocv.NewMat()
		t.cachedResult.CopyTo(&result)
		return result
	}

	t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Cache miss - processing new result")

	// Process with reduced resolution for speed
	result := t.applyWithScale(src, 0.5)

	// Update cache
	if !t.cachedResult.Empty() {
		t.cachedResult.Close()
	}
	t.cachedResult = gocv.NewMat()
	result.CopyTo(&t.cachedResult)
	t.cacheValid = true
	t.cachedParams = make(map[string]interface{})
	for k, v := range currentParams {
		t.cachedParams[k] = v
	}

	t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Cache updated successfully")
	return result
}

func (t *TwoDOtsu) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	t.onParameterChanged = onParameterChanged
	return t.createParameterUI()
}

func (t *TwoDOtsu) Close() {
	if !t.cachedResult.Empty() {
		t.cachedResult.Close()
	}
}

func (t *TwoDOtsu) compareParams(current map[string]interface{}) bool {
	if len(current) != len(t.cachedParams) {
		return false
	}

	for k, v := range current {
		if cachedV, exists := t.cachedParams[k]; !exists || cachedV != v {
			return false
		}
	}
	return true
}

func (t *TwoDOtsu) applyWithScale(src gocv.Mat, scale float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Starting binarization (scale: %.2f)", scale))

	// Scale input if needed
	var workingImage gocv.Mat
	if scale != 1.0 {
		workingImage = gocv.NewMat()
		newWidth := int(float64(src.Cols()) * scale)
		newHeight := int(float64(src.Rows()) * scale)
		gocv.Resize(src, &workingImage, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)
		t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Scaled to %dx%d", newWidth, newHeight))
	} else {
		workingImage = src.Clone()
	}
	defer workingImage.Close()

	// Convert to grayscale
	t.debugImage.LogColorConversion("BGR", "Grayscale")
	grayscale := gocv.NewMat()
	defer grayscale.Close()
	gocv.CvtColor(workingImage, &grayscale, gocv.ColorBGRToGray)
	t.debugImage.LogMatInfo("grayscale", grayscale)

	// Apply guided filter for smoother guided image
	guided := t.applyGuidedFilter(grayscale)
	defer guided.Close()

	// Apply 2D Otsu thresholding
	binaryResult := t.apply2DOtsu(grayscale, guided)
	defer binaryResult.Close()

	// Post-processing with morphological operations
	t.debugImage.LogAlgorithmStep("2D Otsu", "Postprocessing")
	processed := t.applyMorphologicalOps(binaryResult)

	// Scale back to original size if needed
	var result gocv.Mat
	if scale != 1.0 {
		result = gocv.NewMat()
		gocv.Resize(processed, &result, image.Point{X: src.Cols(), Y: src.Rows()}, 0, 0, gocv.InterpolationNearest)
		processed.Close()
		t.debugImage.LogAlgorithmStep("2D Otsu", "Scaled back to original size")
	} else {
		result = processed
	}

	t.debugImage.LogMatInfo("final_binary", result)
	t.debugImage.LogAlgorithmStep("2D Otsu", "Completed")

	return result
}

func (t *TwoDOtsu) applyGuidedFilter(src gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("GuidedFilter", "Starting guided filter implementation")
	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Input: %dx%d, channels=%d", src.Cols(), src.Rows(), src.Channels()))

	// Convert to float32 for processing
	srcFloat := gocv.NewMat()
	defer srcFloat.Close()
	src.ConvertTo(&srcFloat, gocv.MatTypeCV32F)
	srcFloat.DivideFloat(255.0)

	// Parameters
	kernelSize := 2*t.windowRadius + 1
	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Using box filter with kernel %dx%d", kernelSize, kernelSize))

	// Mean filters
	meanI := gocv.NewMat()
	defer meanI.Close()
	gocv.BoxFilter(srcFloat, &meanI, -1, image.Point{X: kernelSize, Y: kernelSize}, image.Point{X: -1, Y: -1}, true, gocv.BorderReflect101)

	meanII := gocv.NewMat()
	defer meanII.Close()
	srcSquared := gocv.NewMat()
	defer srcSquared.Close()
	gocv.Multiply(srcFloat, srcFloat, &srcSquared, 1.0, -1)
	gocv.BoxFilter(srcSquared, &meanII, -1, image.Point{X: kernelSize, Y: kernelSize}, image.Point{X: -1, Y: -1}, true, gocv.BorderReflect101)

	// Variance calculation
	varI := gocv.NewMat()
	defer varI.Close()
	meanISquared := gocv.NewMat()
	defer meanISquared.Close()
	gocv.Multiply(meanI, meanI, &meanISquared, 1.0, -1)
	gocv.Subtract(meanII, meanISquared, &varI)

	// Calculate coefficients
	a := gocv.NewMat()
	defer a.Close()
	denominator := gocv.NewMat()
	defer denominator.Close()
	gocv.AddFloat(varI, t.epsilon, &denominator)
	gocv.Divide(varI, denominator, &a, 1.0, -1)

	b := gocv.NewMat()
	defer b.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	gocv.Multiply(a, meanI, &temp, 1.0, -1)
	gocv.Subtract(meanI, temp, &b)

	// Smooth coefficients
	meanA := gocv.NewMat()
	defer meanA.Close()
	gocv.BoxFilter(a, &meanA, -1, image.Point{X: kernelSize, Y: kernelSize}, image.Point{X: -1, Y: -1}, true, gocv.BorderReflect101)

	meanB := gocv.NewMat()
	defer meanB.Close()
	gocv.BoxFilter(b, &meanB, -1, image.Point{X: kernelSize, Y: kernelSize}, image.Point{X: -1, Y: -1}, true, gocv.BorderReflect101)

	// Final result
	resultFloat := gocv.NewMat()
	defer resultFloat.Close()
	temp1 := gocv.NewMat()
	defer temp1.Close()
	gocv.Multiply(meanA, srcFloat, &temp1, 1.0, -1)
	gocv.Add(temp1, meanB, &resultFloat)

	// Convert back to uint8
	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)

	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Result: %dx%d, channels=%d", result.Cols(), result.Rows(), result.Channels()))
	t.debugImage.LogFilter("GuidedFilter", fmt.Sprintf("radius=%d epsilon=%.3f", t.windowRadius, t.epsilon))

	return result
}

func (t *TwoDOtsu) apply2DOtsu(gray, guided gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", "Constructing 2D histogram")

	grayData := gray.ToBytes()
	guidedData := guided.ToBytes()

	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Gray dimensions: %dx%d", gray.Cols(), gray.Rows()))
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Guided dimensions: %dx%d", guided.Cols(), guided.Rows()))
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Processing %d pixels", len(grayData)))

	// Build 2D histogram
	var hist [256][256]int
	for i := 0; i < len(grayData); i++ {
		g := int(grayData[i])
		f := int(guidedData[i])
		hist[g][f]++
	}

	// Find optimal thresholds using 2D Otsu
	t.debugImage.LogAlgorithmStep("2D Otsu", "Finding optimal thresholds")
	bestS, bestT, maxVariance := t.findOptimalThresholds(hist, len(grayData))

	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	// Apply thresholding
	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarizing image")
	result := gocv.NewMat()
	result.Create(gray.Rows(), gray.Cols(), gocv.MatTypeCV8U)

	resultData := result.ToBytes()
	for i := 0; i < len(grayData); i++ {
		g := int(grayData[i])
		f := int(guidedData[i])
		if g <= bestS && f <= bestT {
			resultData[i] = 0 // Background
		} else {
			resultData[i] = 255 // Foreground
		}
	}

	return result
}

func (t *TwoDOtsu) findOptimalThresholds(hist [256][256]int, totalPixels int) (int, int, float64) {
	maxVariance := 0.0
	bestS, bestT := 0, 0

	for s := 0; s < 255; s++ {
		for t := 0; t < 255; t++ {
			variance := t.calculateBetweenClassVariance(hist, s, t, totalPixels)
			if variance > maxVariance {
				maxVariance = variance
				bestS = s
				bestT = t
			}
		}
	}

	return bestS, bestT, maxVariance
}

func (t *TwoDOtsu) calculateBetweenClassVariance(hist [256][256]int, s, t, totalPixels int) float64 {
	// Calculate class probabilities and means
	w0, w1 := 0, 0
	mu0G, mu0F, mu1G, mu1F := 0.0, 0.0, 0.0, 0.0

	// Class 0: g <= s, f <= t
	for g := 0; g <= s; g++ {
		for f := 0; f <= t; f++ {
			count := hist[g][f]
			w0 += count
			mu0G += float64(g * count)
			mu0F += float64(f * count)
		}
	}

	// Class 1: remaining pixels
	w1 = totalPixels - w0

	if w0 == 0 || w1 == 0 {
		return 0.0
	}

	mu0G /= float64(w0)
	mu0F /= float64(w0)

	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			if g > s || f > t {
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

func (t *TwoDOtsu) applyMorphologicalOps(src gocv.Mat) gocv.Mat {
	if t.morphKernelSize <= 1 {
		t.debugImage.LogMorphology("Close", t.morphKernelSize)
		t.debugImage.LogMorphology("Open", t.morphKernelSize)
		return src.Clone()
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: t.morphKernelSize, Y: t.morphKernelSize})
	defer kernel.Close()

	// Closing (dilation followed by erosion)
	t.debugImage.LogMorphology("Close", t.morphKernelSize)
	closed := gocv.NewMat()
	defer closed.Close()
	gocv.MorphologyEx(src, &closed, gocv.MorphClose, kernel, image.Point{X: -1, Y: -1}, 1, gocv.BorderConstant, color.RGBA{0, 0, 0, 0})

	// Opening (erosion followed by dilation)
	t.debugImage.LogMorphology("Open", t.morphKernelSize)
	result := gocv.NewMat()
	gocv.MorphologyEx(closed, &result, gocv.MorphOpen, kernel, image.Point{X: -1, Y: -1}, 1, gocv.BorderConstant, color.RGBA{0, 0, 0, 0})

	return result
}

func (t *TwoDOtsu) createParameterUI() *fyne.Container {
	// Window Radius parameter
	radiusLabel := widget.NewLabel("Window Radius (1-20):")
	radiusEntry := widget.NewEntry()
	radiusEntry.SetText(fmt.Sprintf("%d", t.windowRadius))
	radiusEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value > 0 && value <= 20 {
			oldValue := t.windowRadius
			t.windowRadius = value
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Window radius changed: %d -> %d", oldValue, value))
			t.invalidateCache()
			if t.onParameterChanged != nil {
				t.debugImage.LogAlgorithmStep("2D Otsu Parameters", "Calling onParameterChanged callback")
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid window radius value: %s", text))
		}
	}

	// Epsilon parameter
	epsilonLabel := widget.NewLabel("Epsilon (0.001-1.0):")
	epsilonEntry := widget.NewEntry()
	epsilonEntry.SetText(fmt.Sprintf("%.3f", t.epsilon))
	epsilonEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value > 0 && value <= 1.0 {
			oldValue := t.epsilon
			t.epsilon = value
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Epsilon changed: %.3f -> %.3f", oldValue, value))
			t.invalidateCache()
			if t.onParameterChanged != nil {
				t.debugImage.LogAlgorithmStep("2D Otsu Parameters", "Calling onParameterChanged callback")
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid epsilon value: %s", text))
		}
	}

	// Morphological Kernel Size parameter
	kernelLabel := widget.NewLabel("Morphological Kernel Size (1-15, odd):")
	kernelEntry := widget.NewEntry()
	kernelEntry.SetText(fmt.Sprintf("%d", t.morphKernelSize))
	kernelEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value > 0 && value <= 15 && value%2 == 1 {
			oldValue := t.morphKernelSize
			t.morphKernelSize = value
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Morphological kernel size changed: %d -> %d", oldValue, value))
			t.invalidateCache()
			if t.onParameterChanged != nil {
				t.debugImage.LogAlgorithmStep("2D Otsu Parameters", "Calling onParameterChanged callback")
				t.onParameterChanged()
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

func (t *TwoDOtsu) invalidateCache() {
	t.cacheValid = false
}
