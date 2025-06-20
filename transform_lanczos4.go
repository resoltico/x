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

// Lanczos4Transform implements Lanczos4 interpolation with FIXED memory management using standard GoCV APIs
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

	// Apply optional pre-filtering using standard GoCV Gaussian blur
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
		// Standard Lanczos4 scaling using GoCV Resize
		result = gocv.NewMat()
		gocv.Resize(filtered, &result, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLanczos4)
	}

	// FIXED: Validate result
	if result.Empty() {
		l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", "ERROR: Scaling operation failed")
		return gocv.NewMat()
	}

	// Apply post-processing using standard GoCV blur for artifact reduction
	final := l.applyPostFilterFixed(result)
	if !result.Empty() {
		result.Close()
	}

	l.debugImage.LogMatInfo("final_result", final)
	l.debugImage.LogAlgorithmStep("Lanczos4 Fixed", "Scaling completed successfully")

	return final
}

// FIXED: Use standard GoCV GaussianBlur for pre-filtering
func (l *Lanczos4Transform) applyPreFilterFixed(src gocv.Mat) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 PreFilter Fixed", "Applying Gaussian blur to reduce ringing")

	// FIXED: Validate kernel size based on image dimensions
	kernelSize := 3
	if src.Cols() < 10 || src.Rows() < 10 {
		// For very small images, skip pre-filtering
		l.debugImage.LogFilter("PreFilter", "Skipping for small image")
		return src.Clone()
	}

	// Use standard GoCV GaussianBlur
	blurred := gocv.NewMat()
	gocv.GaussianBlur(src, &blurred, image.Point{X: kernelSize, Y: kernelSize}, 0.5, 0.5, gocv.BorderDefault)

	l.debugImage.LogFilter("GaussianBlur Fixed", "kernel=3x3", "sigma=0.5")
	return blurred
}

// FIXED: Use standard GoCV blur instead of custom guided filter for post-processing
func (l *Lanczos4Transform) applyPostFilterFixed(src gocv.Mat) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 PostFilter Fixed", "Applying blur for artifact reduction")

	// FIXED: Use simple blur instead of complex guided filter for post-processing
	filtered := gocv.NewMat()
	gocv.Blur(src, &filtered, image.Point{X: 3, Y: 3})

	l.debugImage.LogFilter("Blur Fixed", "kernel=3x3")
	return filtered
}

// FIXED: Proper memory management for iterative scaling using standard GoCV
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

		// Use standard GoCV Resize with Lanczos4
		gocv.Resize(current, &temp, image.Point{X: nextWidth, Y: nextHeight}, 0, 0, gocv.InterpolationLanczos4)

		// FIXED: Proper cleanup of previous iteration
		current.Close()
		current = temp
		currentWidth, currentHeight = nextWidth, nextHeight
	}

	// Final resize to exact target dimensions using standard GoCV
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
