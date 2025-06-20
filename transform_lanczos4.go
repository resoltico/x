package main

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

// Lanczos4Transform implements Lanczos4 interpolation with FIXED memory management and numerical stability
type Lanczos4Transform struct {
	debugImage   *DebugImage
	scaleFactor  float64
	targetDPI    float64
	originalDPI  float64
	useIterative bool

	// Callback for parameter changes
	onParameterChanged func()
}

// NewLanczos4Transform creates a new Lanczos4 transformation with validated parameters
func NewLanczos4Transform(config *DebugConfig) *Lanczos4Transform {
	return &Lanczos4Transform{
		debugImage:   NewDebugImage(config),
		scaleFactor:  2.0,
		targetDPI:    300.0,
		originalDPI:  150.0,
		useIterative: false,
	}
}

func (l *Lanczos4Transform) Name() string {
	return "Lanczos4 Scaling"
}

func (l *Lanczos4Transform) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"scaleFactor":  l.scaleFactor,
		"targetDPI":    l.targetDPI,
		"originalDPI":  l.originalDPI,
		"useIterative": l.useIterative,
	}
}

func (l *Lanczos4Transform) SetParameters(params map[string]interface{}) {
	if scale, ok := params["scaleFactor"].(float64); ok {
		// FIXED: Validate scale factor to prevent extreme values
		if scale >= 0.1 && scale <= 10.0 {
			l.scaleFactor = scale
		}
	}
	if target, ok := params["targetDPI"].(float64); ok {
		// FIXED: Validate DPI range
		if target >= 72 && target <= 2400 {
			l.targetDPI = target
		}
	}
	if original, ok := params["originalDPI"].(float64); ok {
		// FIXED: Validate DPI range
		if original >= 72 && original <= 2400 {
			l.originalDPI = original
		}
	}
	if iterative, ok := params["useIterative"].(bool); ok {
		l.useIterative = iterative
	}
}

func (l *Lanczos4Transform) Apply(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", "Starting full resolution scaling")
	return l.applyLanczos4Fixed(src, l.scaleFactor)
}

func (l *Lanczos4Transform) ApplyPreview(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4 Preview Fixed", "Starting preview scaling")

	// For preview, use a reasonable scale factor to maintain responsiveness
	previewScale := l.scaleFactor
	if l.scaleFactor > 3.0 {
		previewScale = 3.0
	}

	result := l.applyLanczos4Fixed(src, previewScale)
	l.debugImage.LogAlgorithmStep("Lanczos4 Preview Fixed", "Preview scaling completed")
	return result
}

func (l *Lanczos4Transform) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	l.onParameterChanged = onParameterChanged
	return l.createParameterUI()
}

func (l *Lanczos4Transform) Close() {
	// No resources to cleanup
}

func (l *Lanczos4Transform) applyLanczos4Fixed(src gocv.Mat, scale float64) gocv.Mat {
	// FIXED: Input validation
	if src.Empty() {
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", "ERROR: Input matrix is empty")
		return gocv.NewMat()
	}

	// FIXED: Validate scale factor to prevent extreme scaling
	if scale <= 0 || math.IsInf(scale, 0) || math.IsNaN(scale) {
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", fmt.Sprintf("ERROR: Invalid scale factor: %.3f", scale))
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", fmt.Sprintf("Input: %dx%d, scale: %.2f", src.Cols(), src.Rows(), scale))

	// Convert to grayscale if multi-channel for compatibility with 2D Otsu
	working := gocv.NewMat()
	defer working.Close()

	if src.Channels() > 1 {
		l.debugImage.LogColorConversion("BGR", "Grayscale")
		gocv.CvtColor(src, &working, gocv.ColorBGRToGray)
	} else {
		working = src.Clone()
	}

	l.debugImage.LogMatInfo("input_working", working)

	// FIXED: Validate dimensions before processing
	if working.Cols() <= 0 || working.Rows() <= 0 {
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", "ERROR: Invalid input dimensions")
		return gocv.NewMat()
	}

	// Apply optional pre-filtering to reduce ringing artifacts
	filtered := l.applyPreFilterFixed(working)
	defer filtered.Close()

	// FIXED: Calculate target dimensions with bounds checking
	newWidth := int(math.Round(float64(filtered.Cols()) * scale))
	newHeight := int(math.Round(float64(filtered.Rows()) * scale))

	// FIXED: Validate output dimensions
	if newWidth <= 0 || newHeight <= 0 {
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", fmt.Sprintf("ERROR: Invalid target dimensions: %dx%d", newWidth, newHeight))
		return gocv.NewMat()
	}

	// FIXED: Prevent excessive memory usage
	maxDimension := 32768
	if newWidth > maxDimension || newHeight > maxDimension {
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", fmt.Sprintf("ERROR: Target dimensions too large: %dx%d (max: %d)", newWidth, newHeight, maxDimension))
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", fmt.Sprintf("Target dimensions: %dx%d", newWidth, newHeight))

	var result gocv.Mat

	if l.useIterative && scale < 0.5 {
		// Use iterative downscaling for large reductions
		result = l.iterativeLanczos4Fixed(filtered, newWidth, newHeight)
	} else {
		// Standard Lanczos4 scaling
		result = gocv.NewMat()
		gocv.Resize(filtered, &result, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLanczos4)
	}

	// FIXED: Validate result
	if result.Empty() {
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", "ERROR: Scaling operation failed")
		return gocv.NewMat()
	}

	// Apply post-processing guided filter for artifact reduction
	final := l.applyPostFilterFixed(result)
	if !result.Empty() {
		result.Close()
	}

	l.debugImage.LogMatInfo("final_result", final)
	l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", "Scaling completed successfully")

	return final
}

func (l *Lanczos4Transform) applyPreFilterFixed(src gocv.Mat) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 PreFilter Fixed", "Applying Gaussian blur to reduce ringing")

	// Light Gaussian blur to reduce potential ringing artifacts
	blurred := gocv.NewMat()

	// FIXED: Validate kernel size based on image dimensions
	kernelSize := 3
	if src.Cols() < 10 || src.Rows() < 10 {
		// For very small images, skip pre-filtering
		l.debugImage.LogFilter("PreFilter", "Skipping for small image")
		return src.Clone()
	}

	gocv.GaussianBlur(src, &blurred, image.Point{X: kernelSize, Y: kernelSize}, 0.5, 0.5, gocv.BorderDefault)

	l.debugImage.LogFilter("GaussianBlur Fixed", "kernel=3x3", "sigma=0.5")
	return blurred
}

func (l *Lanczos4Transform) applyPostFilterFixed(src gocv.Mat) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 PostFilter Fixed", "Applying guided filter for artifact reduction")

	// FIXED: Implement safe guided filter with proper bounds checking
	filtered := l.applySimpleGuidedFilterFixed(src, src, 3, 0.01)

	l.debugImage.LogFilter("GuidedFilter Fixed", "radius=3", "epsilon=0.01")
	return filtered
}

// FIXED: Proper memory management and numerical stability for guided filter
func (l *Lanczos4Transform) applySimpleGuidedFilterFixed(guide, input gocv.Mat, radius int, epsilon float64) gocv.Mat {
	if guide.Empty() || input.Empty() {
		return gocv.NewMat()
	}

	// FIXED: Validate epsilon to prevent division by zero
	if epsilon <= 0 || math.IsInf(epsilon, 0) || math.IsNaN(epsilon) {
		epsilon = 0.01
	}

	// FIXED: Validate radius
	if radius <= 0 {
		radius = 1
	}

	// Convert to float32 for processing
	guideFloat := gocv.NewMat()
	defer guideFloat.Close()
	inputFloat := gocv.NewMat()
	defer inputFloat.Close()

	guide.ConvertTo(&guideFloat, gocv.MatTypeCV32F)
	input.ConvertTo(&inputFloat, gocv.MatTypeCV32F)

	// FIXED: Normalize to [0,1] range to prevent overflow
	guideFloat.DivideFloat(255.0)
	inputFloat.DivideFloat(255.0)

	kernelSize := 2*radius + 1

	// FIXED: Validate kernel size against image dimensions
	if kernelSize > guideFloat.Cols() || kernelSize > guideFloat.Rows() {
		l.debugImage.LogFilter("GuidedFilter Fixed", "Kernel too large for image, returning input")
		return input.Clone()
	}

	// Mean filters with proper cleanup
	meanI := gocv.NewMat()
	defer meanI.Close()
	gocv.BoxFilter(guideFloat, &meanI, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanP := gocv.NewMat()
	defer meanP.Close()
	gocv.BoxFilter(inputFloat, &meanP, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Correlation and variance with proper cleanup
	corrIP := gocv.NewMat()
	defer corrIP.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	gocv.Multiply(guideFloat, inputFloat, &temp)
	gocv.BoxFilter(temp, &corrIP, -1, image.Point{X: kernelSize, Y: kernelSize})

	varI := gocv.NewMat()
	defer varI.Close()
	temp2 := gocv.NewMat()
	defer temp2.Close()
	gocv.Multiply(guideFloat, guideFloat, &temp2)
	gocv.BoxFilter(temp2, &varI, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanISquared := gocv.NewMat()
	defer meanISquared.Close()
	gocv.Multiply(meanI, meanI, &meanISquared)
	gocv.Subtract(varI, meanISquared, &varI)

	// Calculate coefficients with proper cleanup
	a := gocv.NewMat()
	defer a.Close()
	covIP := gocv.NewMat()
	defer covIP.Close()
	temp3 := gocv.NewMat()
	defer temp3.Close()
	gocv.Multiply(meanI, meanP, &temp3)
	gocv.Subtract(corrIP, temp3, &covIP)

	denominator := gocv.NewMat()
	defer denominator.Close()
	varI.CopyTo(&denominator)
	denominator.AddFloat(float32(epsilon))
	gocv.Divide(covIP, denominator, &a)

	b := gocv.NewMat()
	defer b.Close()
	temp4 := gocv.NewMat()
	defer temp4.Close()
	gocv.Multiply(a, meanI, &temp4)
	gocv.Subtract(meanP, temp4, &b)

	// Smooth coefficients with proper cleanup
	meanA := gocv.NewMat()
	defer meanA.Close()
	gocv.BoxFilter(a, &meanA, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanB := gocv.NewMat()
	defer meanB.Close()
	gocv.BoxFilter(b, &meanB, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Final result with proper cleanup
	resultFloat := gocv.NewMat()
	defer resultFloat.Close()
	temp5 := gocv.NewMat()
	defer temp5.Close()
	gocv.Multiply(meanA, guideFloat, &temp5)
	gocv.Add(temp5, meanB, &resultFloat)

	// FIXED: Clamp values to valid range before conversion
	resultFloat.MultiplyFloat(255.0)

	// Convert back to uint8 with proper range checking
	result := gocv.NewMat()
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)

	return result
}

// FIXED: Proper memory management for iterative scaling
func (l *Lanczos4Transform) iterativeLanczos4Fixed(src gocv.Mat, targetWidth, targetHeight int) gocv.Mat {
	if src.Empty() || targetWidth <= 0 || targetHeight <= 0 {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 Iterative Fixed", "Starting iterative downscaling")

	current := src.Clone()
	defer current.Close()
	currentWidth, currentHeight := current.Cols(), current.Rows()

	step := 0
	maxSteps := 20 // FIXED: Prevent infinite loops

	for (currentWidth > targetWidth*2 || currentHeight > targetHeight*2) && step < maxSteps {
		step++
		temp := gocv.NewMat()
		nextWidth := int(math.Max(float64(currentWidth)/2, float64(targetWidth)))
		nextHeight := int(math.Max(float64(currentHeight)/2, float64(targetHeight)))

		// FIXED: Ensure we're making progress
		if nextWidth >= currentWidth || nextHeight >= currentHeight {
			temp.Close()
			break
		}

		l.debugImage.LogAlgorithmStep("Lanczos4 Iterative Fixed", fmt.Sprintf("Step %d: %dx%d -> %dx%d",
			step, currentWidth, currentHeight, nextWidth, nextHeight))

		gocv.Resize(current, &temp, image.Point{X: nextWidth, Y: nextHeight}, 0, 0, gocv.InterpolationLanczos4)

		// FIXED: Proper cleanup of previous iteration
		current.Close()
		current = temp
		currentWidth, currentHeight = nextWidth, nextHeight
	}

	// Final resize to exact target dimensions
	scaled := gocv.NewMat()
	gocv.Resize(current, &scaled, image.Point{X: targetWidth, Y: targetHeight}, 0, 0, gocv.InterpolationLanczos4)

	l.debugImage.LogAlgorithmStep("Lanczos4 Iterative Fixed", fmt.Sprintf("Completed in %d steps", step+1))
	return scaled
}

func (l *Lanczos4Transform) calculateScaleFactor() float64 {
	// FIXED: Validate DPI values to prevent division by zero
	if l.originalDPI > 0 && l.targetDPI > 0 && !math.IsInf(l.originalDPI, 0) && !math.IsInf(l.targetDPI, 0) {
		calculated := l.targetDPI / l.originalDPI

		// FIXED: Clamp to reasonable range
		if calculated > 0.01 && calculated < 100.0 {
			l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", fmt.Sprintf("Calculated scale factor: %.3f (%.0f DPI -> %.0f DPI)",
				calculated, l.originalDPI, l.targetDPI))
			return calculated
		}
	}
	return l.scaleFactor
}

func (l *Lanczos4Transform) createParameterUI() *fyne.Container {
	// Scale Factor parameter with validation
	scaleLabel := widget.NewLabel("Scale Factor (0.1-10.0):")
	scaleEntry := widget.NewEntry()
	scaleEntry.SetText(fmt.Sprintf("%.2f", l.scaleFactor))
	scaleEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 0.1 && value <= 10.0 {
			oldValue := l.scaleFactor
			l.scaleFactor = value
			l.debugImage.LogAlgorithmStep("Lanczos4 Fixed Parameters", fmt.Sprintf("Scale factor changed: %.3f -> %.3f", oldValue, value))
			if l.onParameterChanged != nil {
				go l.onParameterChanged()
			}
		} else {
			l.debugImage.LogAlgorithmStep("Lanczos4 Fixed Parameters", fmt.Sprintf("Invalid scale factor: %s (must be 0.1-10.0)", text))
		}
	}

	// Target DPI parameter with validation
	targetDPILabel := widget.NewLabel("Target DPI (72-2400):")
	targetDPIEntry := widget.NewEntry()
	targetDPIEntry.SetText(fmt.Sprintf("%.0f", l.targetDPI))
	targetDPIEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 72 && value <= 2400 {
			oldValue := l.targetDPI
			l.targetDPI = value
			// Recalculate scale factor based on DPI
			if l.originalDPI > 0 {
				l.scaleFactor = l.calculateScaleFactor()
				scaleEntry.SetText(fmt.Sprintf("%.2f", l.scaleFactor))
			}
			l.debugImage.LogAlgorithmStep("Lanczos4 Fixed Parameters", fmt.Sprintf("Target DPI changed: %.0f -> %.0f", oldValue, value))
			if l.onParameterChanged != nil {
				go l.onParameterChanged()
			}
		} else {
			l.debugImage.LogAlgorithmStep("Lanczos4 Fixed Parameters", fmt.Sprintf("Invalid target DPI: %s (must be 72-2400)", text))
		}
	}

	// Original DPI parameter with validation
	originalDPILabel := widget.NewLabel("Original DPI (72-2400):")
	originalDPIEntry := widget.NewEntry()
	originalDPIEntry.SetText(fmt.Sprintf("%.0f", l.originalDPI))
	originalDPIEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 72 && value <= 2400 {
			oldValue := l.originalDPI
			l.originalDPI = value
			// Recalculate scale factor based on DPI
			l.scaleFactor = l.calculateScaleFactor()
			scaleEntry.SetText(fmt.Sprintf("%.2f", l.scaleFactor))
			l.debugImage.LogAlgorithmStep("Lanczos4 Fixed Parameters", fmt.Sprintf("Original DPI changed: %.0f -> %.0f", oldValue, value))
			if l.onParameterChanged != nil {
				go l.onParameterChanged()
			}
		} else {
			l.debugImage.LogAlgorithmStep("Lanczos4 Fixed Parameters", fmt.Sprintf("Invalid original DPI: %s (must be 72-2400)", text))
		}
	}

	// Iterative downscaling checkbox
	iterativeCheck := widget.NewCheck("Use Iterative Downscaling", func(checked bool) {
		oldValue := l.useIterative
		l.useIterative = checked
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed Parameters", fmt.Sprintf("Iterative mode changed: %t -> %t", oldValue, checked))
		if l.onParameterChanged != nil {
			go l.onParameterChanged()
		}
	})
	iterativeCheck.SetChecked(l.useIterative)

	// Calculate button
	calculateBtn := widget.NewButton("Calculate Scale from DPI", func() {
		l.scaleFactor = l.calculateScaleFactor()
		scaleEntry.SetText(fmt.Sprintf("%.2f", l.scaleFactor))
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed Parameters", "Scale factor recalculated from DPI values")
		if l.onParameterChanged != nil {
			go l.onParameterChanged()
		}
	})

	return container.NewVBox(
		scaleLabel, scaleEntry,
		targetDPILabel, targetDPIEntry,
		originalDPILabel, originalDPIEntry,
		iterativeCheck,
		calculateBtn,
	)
}
