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

type TwoDOtsu struct {
	ThreadSafeTransformation
	debugImage *DebugImage

	paramMutex      sync.RWMutex
	windowRadius    int
	epsilon         float64
	morphKernelSize int

	onParameterChanged func()
}

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
	return t.applyWithScale(src, 0.5)
}

func (t *TwoDOtsu) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	t.onParameterChanged = onParameterChanged
	return t.createParameterUI()
}

func (t *TwoDOtsu) Close() {
	// No resources to cleanup
}

func (t *TwoDOtsu) applyWithScale(src gocv.Mat, scale float64) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	t.paramMutex.RLock()
	windowRadius := t.windowRadius
	epsilon := t.epsilon
	morphKernelSize := t.morphKernelSize
	t.paramMutex.RUnlock()

	// Scale image if needed
	working := gocv.NewMat()
	defer working.Close()

	if scale != 1.0 {
		newWidth := int(float64(src.Cols()) * scale)
		newHeight := int(float64(src.Rows()) * scale)
		gocv.Resize(src, &working, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)
	} else {
		working = src.Clone()
	}

	// Convert to grayscale
	grayscale := gocv.NewMat()
	defer grayscale.Close()

	if working.Channels() > 1 {
		gocv.CvtColor(working, &grayscale, gocv.ColorBGRToGray)
	} else {
		grayscale = working.Clone()
	}

	// Apply guided filter
	guided := t.applyGuidedFilter(grayscale, windowRadius, epsilon)
	defer guided.Close()

	// Apply 2D Otsu
	binary := t.apply2DOtsu(grayscale, guided)
	defer binary.Close()

	// Morphological operations
	processed := t.applyMorphology(binary, morphKernelSize)
	defer processed.Close()

	// Scale back if needed
	if scale != 1.0 {
		result := gocv.NewMat()
		gocv.Resize(processed, &result, image.Point{X: src.Cols(), Y: src.Rows()}, 0, 0, gocv.InterpolationLinear)
		return result
	}

	return processed.Clone()
}

func (t *TwoDOtsu) applyGuidedFilter(src gocv.Mat, radius int, eps float64) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	// Convert to float
	srcFloat := gocv.NewMat()
	defer srcFloat.Close()
	src.ConvertTo(&srcFloat, gocv.MatTypeCV32F)
	srcFloat.DivideFloat(255.0)

	kernelSize := 2*radius + 1

	// Mean of input
	meanI := gocv.NewMat()
	defer meanI.Close()
	gocv.BoxFilter(srcFloat, &meanI, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Mean of input squared
	srcSq := gocv.NewMat()
	defer srcSq.Close()
	gocv.Multiply(srcFloat, srcFloat, &srcSq)

	meanII := gocv.NewMat()
	defer meanII.Close()
	gocv.BoxFilter(srcSq, &meanII, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Variance: Var(I) = E[I*I] - E[I]*E[I]
	varI := gocv.NewMat()
	defer varI.Close()
	meanISq := gocv.NewMat()
	defer meanISq.Close()
	gocv.Multiply(meanI, meanI, &meanISq)
	gocv.Subtract(meanII, meanISq, &varI)

	// a = var(I) / (var(I) + eps)
	a := gocv.NewMat()
	defer a.Close()
	denominator := gocv.NewMat()
	defer denominator.Close()
	varI.CopyTo(&denominator)
	denominator.AddFloat(float32(eps))
	gocv.Divide(varI, denominator, &a)

	// b = mean(I) * (1 - a)
	b := gocv.NewMat()
	defer b.Close()
	ones := gocv.NewMatWithSizeFromScalar(gocv.NewScalar(1, 0, 0, 0), a.Rows(), a.Cols(), gocv.MatTypeCV32F)
	defer ones.Close()
	oneMinusA := gocv.NewMat()
	defer oneMinusA.Close()
	gocv.Subtract(ones, a, &oneMinusA)
	gocv.Multiply(meanI, oneMinusA, &b)

	// Smooth coefficients
	meanA := gocv.NewMat()
	defer meanA.Close()
	gocv.BoxFilter(a, &meanA, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanB := gocv.NewMat()
	defer meanB.Close()
	gocv.BoxFilter(b, &meanB, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Result: q = mean_a * I + mean_b
	result := gocv.NewMat()
	defer result.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	gocv.Multiply(meanA, srcFloat, &temp)
	gocv.Add(temp, meanB, &result)

	// Convert back to uint8
	final := gocv.NewMat()
	result.MultiplyFloat(255.0)
	result.ConvertTo(&final, gocv.MatTypeCV8U)

	return final
}

func (t *TwoDOtsu) apply2DOtsu(gray, guided gocv.Mat) gocv.Mat {
	if gray.Empty() || guided.Empty() {
		return gocv.NewMat()
	}

	grayData := gray.ToBytes()
	guidedData := guided.ToBytes()

	if len(grayData) != len(guidedData) {
		return gocv.NewMat()
	}

	// Build 2D histogram
	var hist [256][256]float64
	totalPixels := len(grayData)

	for i := 0; i < totalPixels; i++ {
		g := int(grayData[i])
		f := int(guidedData[i])
		hist[g][f]++
	}

	// Normalize histogram
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			hist[g][f] /= float64(totalPixels)
		}
	}

	// Find optimal thresholds
	bestS, bestT := t.findOptimalThresholds2D(hist)

	// Apply thresholding
	size := gray.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grayVal := int(gray.GetUCharAt(y, x))
			guidedVal := int(guided.GetUCharAt(y, x))

			if grayVal <= bestS && guidedVal <= bestT {
				result.SetUCharAt(y, x, 0) // Foreground (dark)
			} else {
				result.SetUCharAt(y, x, 255) // Background (light)
			}
		}
	}

	return result
}

func (t *TwoDOtsu) findOptimalThresholds2D(hist [256][256]float64) (int, int) {
	maxVariance := 0.0
	bestS, bestT := 128, 128

	for s := 1; s < 255; s++ {
		for threshold := 1; threshold < 255; threshold++ {
			variance := t.calculateBetweenClassVariance2D(hist, s, threshold)
			if variance > maxVariance {
				maxVariance = variance
				bestS = s
				bestT = threshold
			}
		}
	}

	return bestS, bestT
}

func (t *TwoDOtsu) calculateBetweenClassVariance2D(hist [256][256]float64, s, threshold int) float64 {
	// Class probabilities and means
	var w0, w1 float64
	var μ0g, μ0f, μ1g, μ1f float64

	// Class 0: g <= s, f <= threshold
	for g := 0; g <= s; g++ {
		for f := 0; f <= threshold; f++ {
			prob := hist[g][f]
			w0 += prob
			μ0g += float64(g) * prob
			μ0f += float64(f) * prob
		}
	}

	// Class 1: rest
	w1 = 1.0 - w0

	if w0 == 0 || w1 == 0 {
		return 0.0
	}

	// Normalize means
	μ0g /= w0
	μ0f /= w0

	// Calculate means for class 1
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			if g > s || f > threshold {
				prob := hist[g][f]
				μ1g += float64(g) * prob
				μ1f += float64(f) * prob
			}
		}
	}

	μ1g /= w1
	μ1f /= w1

	// Between-class variance in 2D
	diffG := μ1g - μ0g
	diffF := μ1f - μ0f

	return w0 * w1 * (diffG*diffG + diffF*diffF)
}

func (t *TwoDOtsu) applyMorphology(src gocv.Mat, kernelSize int) gocv.Mat {
	if kernelSize <= 1 {
		return src.Clone()
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: kernelSize, Y: kernelSize})
	defer kernel.Close()

	// Closing operation
	closed := gocv.NewMat()
	defer closed.Close()
	gocv.MorphologyEx(src, &closed, gocv.MorphClose, kernel)

	// Opening operation
	result := gocv.NewMat()
	gocv.MorphologyEx(closed, &result, gocv.MorphOpen, kernel)

	return result
}

func (t *TwoDOtsu) createParameterUI() *fyne.Container {
	// Window Radius
	radiusLabel := widget.NewLabel("Window Radius (1-20):")
	radiusEntry := widget.NewEntry()
	t.paramMutex.RLock()
	radiusEntry.SetText(fmt.Sprintf("%d", t.windowRadius))
	t.paramMutex.RUnlock()

	radiusEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value > 0 && value <= 20 {
			t.paramMutex.Lock()
			t.windowRadius = value
			t.paramMutex.Unlock()
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		}
	}

	// Epsilon
	epsilonLabel := widget.NewLabel("Epsilon (0.001-1.0):")
	epsilonEntry := widget.NewEntry()
	t.paramMutex.RLock()
	epsilonEntry.SetText(fmt.Sprintf("%.3f", t.epsilon))
	t.paramMutex.RUnlock()

	epsilonEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value > 0 && value <= 1.0 {
			t.paramMutex.Lock()
			t.epsilon = value
			t.paramMutex.Unlock()
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		}
	}

	// Morphological Kernel Size
	kernelLabel := widget.NewLabel("Morphological Kernel Size (1-15, odd):")
	kernelEntry := widget.NewEntry()
	t.paramMutex.RLock()
	kernelEntry.SetText(fmt.Sprintf("%d", t.morphKernelSize))
	t.paramMutex.RUnlock()

	kernelEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value > 0 && value <= 15 && value%2 == 1 {
			t.paramMutex.Lock()
			t.morphKernelSize = value
			t.paramMutex.Unlock()
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		}
	}

	return container.NewVBox(
		radiusLabel, radiusEntry,
		epsilonLabel, epsilonEntry,
		kernelLabel, kernelEntry,
	)
}
