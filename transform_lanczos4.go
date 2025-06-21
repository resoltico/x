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

// Lanczos4Transform implements Lanczos4 interpolation with memory management
type Lanczos4Transform struct {
	debugImage   *DebugImage
	scaleFactor  float64
	targetDPI    float64
	originalDPI  float64
	useIterative bool

	// Callback for parameter changes
	onParameterChanged func()
}

// NewLanczos4Transform creates new Lanczos4 transformation with validated parameters
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
		// Validate scale factor to prevent extreme values
		if scale >= 0.1 && scale <= 10.0 {
			l.scaleFactor = scale
		}
	}
	if target, ok := params["targetDPI"].(float64); ok {
		// Validate DPI range
		if target >= 72 && target <= 2400 {
			l.targetDPI = target
		}
	}
	if original, ok := params["originalDPI"].(float64); ok {
		// Validate DPI range
		if original >= 72 && original <= 2400 {
			l.originalDPI = original
		}
	}
	if iterative, ok := params["useIterative"].(bool); ok {
		l.useIterative = iterative
	}
}

func (l *Lanczos4Transform) Apply(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4", "Starting full resolution scaling")
	return l.applyLanczos4(src, l.scaleFactor)
}

func (l *Lanczos4Transform) ApplyPreview(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4 Preview", "Starting preview scaling")

	// For preview, use reasonable scale factor to maintain responsiveness
	previewScale := l.scaleFactor
	if l.scaleFactor > 3.0 {
		previewScale = 3.0
	}

	result := l.applyLanczos4(src, previewScale)
	l.debugImage.LogAlgorithmStep("Lanczos4 Preview", "Preview scaling completed")
	return result
}

func (l *Lanczos4Transform) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	l.onParameterChanged = onParameterChanged
	return l.createParameterUI()
}

func (l *Lanczos4Transform) Close() {
	// No resources to cleanup
}

func (l *Lanczos4Transform) applyLanczos4(src gocv.Mat, scale float64) gocv.Mat {
	// Input validation
	if src.Empty() {
		l.debugImage.LogAlgorithmStep("Lanczos4", "ERROR: Input matrix is empty")
		return gocv.NewMat()
	}

	// Validate scale factor to prevent extreme scaling
	if scale <= 0 || math.IsInf(scale, 0) || math.IsNaN(scale) {
		l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("ERROR: Invalid scale factor: %.3f", scale))
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Input: %dx%d, scale: %.2f", src.Cols(), src.Rows(), scale))

	// Convert to grayscale if multi-channel for compatibility with 2D Otsu
	working := gocv.NewMat()
	defer working.Close()

	if src.Channels() > 1 {
		l.debugImage.LogColorConversion("BGR", "Grayscale")
		err := gocv.CvtColor(src, &working, gocv.ColorBGRToGray)
		if err != nil {
			l.debugImage.LogError(err)
			return gocv.NewMat()
		}
	} else {
		working = src.Clone()
	}

	l.debugImage.LogMatInfo("input_working", working)

	// Validate dimensions before processing
	if working.Cols() <= 0 || working.Rows() <= 0 {
		l.debugImage.LogAlgorithmStep("Lanczos4", "ERROR: Invalid input dimensions")
		return gocv.NewMat()
	}

	// Apply pre-filtering using GoCV APIs with memory management
	filtered := l.applyPreFilter(working)
	defer filtered.Close()

	// Calculate target dimensions with bounds checking
	newWidth := int(math.Round(float64(filtered.Cols()) * scale))
	newHeight := int(math.Round(float64(filtered.Rows()) * scale))

	// Validate output dimensions
	if newWidth <= 0 || newHeight <= 0 {
		l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("ERROR: Invalid target dimensions: %dx%d", newWidth, newHeight))
		return gocv.NewMat()
	}

	// Prevent excessive memory usage
	maxDimension := 32768
	if newWidth > maxDimension || newHeight > maxDimension {
		l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("ERROR: Target dimensions too large: %dx%d (max: %d)", newWidth, newHeight, maxDimension))
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Target dimensions: %dx%d", newWidth, newHeight))

	var result gocv.Mat

	if l.useIterative && scale < 0.5 {
		// Use iterative downscaling for large reductions
		result = l.iterativeLanczos4(filtered, newWidth, newHeight)
	} else {
		// Direct Lanczos4 scaling using GoCV Resize
		result = gocv.NewMat()
		err := gocv.Resize(filtered, &result, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLanczos4)
		if err != nil {
			l.debugImage.LogError(err)
			result.Close()
			return gocv.NewMat()
		}
	}

	// Validate result
	if result.Empty() {
		l.debugImage.LogAlgorithmStep("Lanczos4", "ERROR: Scaling operation failed")
		return gocv.NewMat()
	}

	// Apply post-processing using GoCV blur with memory management
	final := l.applyPostFilter(result)
	result.Close() // Close intermediate result

	l.debugImage.LogMatInfo("final_result", final)
	l.debugImage.LogAlgorithmStep("Lanczos4", "Scaling completed successfully")

	return final
}

// Use GoCV GaussianBlur with adaptive kernel sizing and memory management
func (l *Lanczos4Transform) applyPreFilter(src gocv.Mat) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 PreFilter", "Applying adaptive Gaussian blur")

	// Adaptive kernel size based on image dimensions
	minDim := min(src.Cols(), src.Rows())
	var kernelSize int
	if minDim < 100 {
		// Skip pre-filtering for very small images
		l.debugImage.LogFilter("PreFilter", "Skipping for small image")
		return src.Clone()
	} else if minDim < 500 {
		kernelSize = 3
	} else if minDim < 1000 {
		kernelSize = 5
	} else {
		kernelSize = 7
	}

	// Ensure odd kernel size
	if kernelSize%2 == 0 {
		kernelSize++
	}

	// Use GoCV's GaussianBlur with adaptive sigma
	blurred := gocv.NewMat()
	sigma := float64(kernelSize) / 6.0 // Standard relationship between kernel size and sigma
	err := gocv.GaussianBlur(src, &blurred, image.Point{X: kernelSize, Y: kernelSize}, sigma, sigma, gocv.BorderDefault)
	if err != nil {
		l.debugImage.LogError(err)
		blurred.Close()
		return src.Clone()
	}

	l.debugImage.LogFilter("GaussianBlur", fmt.Sprintf("kernel=%dx%d", kernelSize, kernelSize), fmt.Sprintf("sigma=%.2f", sigma))
	return blurred
}

// Use GoCV's bilateral filter for better edge-preserving post-processing with memory management
func (l *Lanczos4Transform) applyPostFilter(src gocv.Mat) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 PostFilter", "Applying bilateral filter for artifact reduction")

	// Use bilateral filter for better edge preservation
	filtered := gocv.NewMat()

	// Adaptive filter strength based on image size
	minDim := min(src.Cols(), src.Rows())
	var d int
	var sigmaColor, sigmaSpace float64

	if minDim < 500 {
		d = 5
		sigmaColor = 50.0
		sigmaSpace = 50.0
	} else if minDim < 1000 {
		d = 7
		sigmaColor = 75.0
		sigmaSpace = 75.0
	} else {
		d = 9
		sigmaColor = 100.0
		sigmaSpace = 100.0
	}

	err := gocv.BilateralFilter(src, &filtered, d, sigmaColor, sigmaSpace)
	if err != nil {
		l.debugImage.LogError(err)
		filtered.Close()
		return src.Clone()
	}

	l.debugImage.LogFilter("BilateralFilter", fmt.Sprintf("d=%d sigmaColor=%.1f sigmaSpace=%.1f", d, sigmaColor, sigmaSpace))
	return filtered
}

// Iterative scaling with better memory management
func (l *Lanczos4Transform) iterativeLanczos4(src gocv.Mat, targetWidth, targetHeight int) gocv.Mat {
	if src.Empty() || targetWidth <= 0 || targetHeight <= 0 {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", "Starting iterative downscaling")

	current := src.Clone()
	defer current.Close()
	currentWidth, currentHeight := current.Cols(), current.Rows()

	step := 0
	maxSteps := 15 // Limit max steps for better performance

	// Use more aggressive downscaling steps
	scalingFactor := 0.6 // Scale down by 40% each step instead of 50%

	for (currentWidth > targetWidth*2 || currentHeight > targetHeight*2) && step < maxSteps {
		step++

		// Calculate next dimensions more aggressively
		nextWidth := int(math.Max(float64(currentWidth)*scalingFactor, float64(targetWidth)))
		nextHeight := int(math.Max(float64(currentHeight)*scalingFactor, float64(targetHeight)))

		// Ensure we're making progress
		if nextWidth >= currentWidth || nextHeight >= currentHeight {
			break
		}

		l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", fmt.Sprintf("Step %d: %dx%d -> %dx%d",
			step, currentWidth, currentHeight, nextWidth, nextHeight))

		temp := gocv.NewMat()

		// Use area interpolation for downscaling (better quality than Lanczos for large reductions)
		interpolation := gocv.InterpolationArea
		if nextWidth > currentWidth || nextHeight > currentHeight {
			interpolation = gocv.InterpolationLanczos4 // Use Lanczos4 for upscaling
		}

		err := gocv.Resize(current, &temp, image.Point{X: nextWidth, Y: nextHeight}, 0, 0, interpolation)
		if err != nil {
			l.debugImage.LogError(err)
			temp.Close()
			break
		}

		// Cleanup of previous iteration
		current.Close()
		current = temp
		currentWidth, currentHeight = nextWidth, nextHeight
	}

	// Final resize to exact target dimensions using Lanczos4
	scaled := gocv.NewMat()
	err := gocv.Resize(current, &scaled, image.Point{X: targetWidth, Y: targetHeight}, 0, 0, gocv.InterpolationLanczos4)
	if err != nil {
		l.debugImage.LogError(err)
		scaled.Close()
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", fmt.Sprintf("Completed in %d steps", step+1))
	return scaled
}

func (l *Lanczos4Transform) calculateScaleFactor() float64 {
	// Validate DPI values to prevent division by zero
	if l.originalDPI > 0 && l.targetDPI > 0 && !math.IsInf(l.originalDPI, 0) && !math.IsInf(l.targetDPI, 0) {
		calculated := l.targetDPI / l.originalDPI

		// Clamp to reasonable range
		if calculated > 0.01 && calculated < 100.0 {
			l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Calculated scale factor: %.3f (%.0f DPI -> %.0f DPI)",
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
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Scale factor changed: %.3f -> %.3f", oldValue, value))
			if l.onParameterChanged != nil {
				l.onParameterChanged()
			}
		} else {
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Invalid scale factor: %s (must be 0.1-10.0)", text))
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
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Target DPI changed: %.0f -> %.0f", oldValue, value))
			if l.onParameterChanged != nil {
				l.onParameterChanged()
			}
		} else {
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Invalid target DPI: %s (must be 72-2400)", text))
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
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Original DPI changed: %.0f -> %.0f", oldValue, value))
			if l.onParameterChanged != nil {
				l.onParameterChanged()
			}
		} else {
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Invalid original DPI: %s (must be 72-2400)", text))
		}
	}

	// Iterative downscaling checkbox
	iterativeCheck := widget.NewCheck("Use Iterative Downscaling", func(checked bool) {
		oldValue := l.useIterative
		l.useIterative = checked
		l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Iterative mode changed: %t -> %t", oldValue, checked))
		if l.onParameterChanged != nil {
			l.onParameterChanged()
		}
	})
	iterativeCheck.SetChecked(l.useIterative)

	// Calculate button
	calculateBtn := widget.NewButton("Calculate Scale from DPI", func() {
		l.scaleFactor = l.calculateScaleFactor()
		scaleEntry.SetText(fmt.Sprintf("%.2f", l.scaleFactor))
		l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", "Scale factor recalculated from DPI values")
		if l.onParameterChanged != nil {
			l.onParameterChanged()
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
