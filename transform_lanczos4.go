package main

import (
	"fmt"
	"image"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

// Lanczos4Transform implements Lanczos4 interpolation for high-quality image scaling
type Lanczos4Transform struct {
	debugImage   *DebugImage
	scaleFactor  float64
	targetDPI    float64
	originalDPI  float64
	useIterative bool

	// Callback for parameter changes
	onParameterChanged func()
}

// NewLanczos4Transform creates a new Lanczos4 transformation
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
		l.scaleFactor = scale
	}
	if target, ok := params["targetDPI"].(float64); ok {
		l.targetDPI = target
	}
	if original, ok := params["originalDPI"].(float64); ok {
		l.originalDPI = original
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

	// For preview, use a smaller scale factor to maintain responsiveness
	previewScale := l.scaleFactor
	if l.scaleFactor > 2.0 {
		previewScale = 2.0
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
	l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Input: %dx%d, scale: %.2f", src.Cols(), src.Rows(), scale))

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

	// Apply optional pre-filtering to reduce ringing artifacts
	filtered := l.applyPreFilter(working)
	defer filtered.Close()

	// Calculate target dimensions
	newWidth := int(float64(filtered.Cols()) * scale)
	newHeight := int(float64(filtered.Rows()) * scale)

	l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Target dimensions: %dx%d", newWidth, newHeight))

	var result gocv.Mat

	if l.useIterative && scale < 0.5 {
		// Use iterative downscaling for large reductions
		result = l.iterativeLanczos4(filtered, newWidth, newHeight)
	} else {
		// Standard Lanczos4 scaling
		result = gocv.NewMat()
		gocv.Resize(filtered, &result, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLanczos4)
	}

	// Apply post-processing guided filter for artifact reduction
	final := l.applyPostFilter(result)
	result.Close()

	l.debugImage.LogMatInfo("final_result", final)
	l.debugImage.LogAlgorithmStep("Lanczos4", "Scaling completed")

	return final
}

func (l *Lanczos4Transform) applyPreFilter(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4 PreFilter", "Applying Gaussian blur to reduce ringing")

	// Light Gaussian blur to reduce potential ringing artifacts
	blurred := gocv.NewMat()
	gocv.GaussianBlur(src, &blurred, image.Point{X: 3, Y: 3}, 0.5, 0.5, gocv.BorderDefault)

	l.debugImage.LogFilter("GaussianBlur", "kernel=3x3", "sigma=0.5")
	return blurred
}

func (l *Lanczos4Transform) applyPostFilter(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4 PostFilter", "Applying guided filter for artifact reduction")

	// Simple guided filter implementation using box filter
	filtered := l.applySimpleGuidedFilter(src, src, 3, 0.01)

	l.debugImage.LogFilter("GuidedFilter", "radius=3", "epsilon=0.01")
	return filtered
}

func (l *Lanczos4Transform) applySimpleGuidedFilter(guide, input gocv.Mat, radius int, epsilon float64) gocv.Mat {
	// Convert to float32 for processing
	guideFloat := gocv.NewMat()
	defer guideFloat.Close()
	inputFloat := gocv.NewMat()
	defer inputFloat.Close()

	guide.ConvertTo(&guideFloat, gocv.MatTypeCV32F)
	input.ConvertTo(&inputFloat, gocv.MatTypeCV32F)

	guideFloat.DivideFloat(255.0)
	inputFloat.DivideFloat(255.0)

	kernelSize := 2*radius + 1

	// Mean filters
	meanI := gocv.NewMat()
	defer meanI.Close()
	gocv.BoxFilter(guideFloat, &meanI, -1, image.Point{X: kernelSize, Y: kernelSize})

	meanP := gocv.NewMat()
	defer meanP.Close()
	gocv.BoxFilter(inputFloat, &meanP, -1, image.Point{X: kernelSize, Y: kernelSize})

	// Correlation and variance
	corrIP := gocv.NewMat()
	defer corrIP.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	gocv.Multiply(guideFloat, inputFloat, &temp)
	gocv.BoxFilter(temp, &corrIP, -1, image.Point{X: kernelSize, Y: kernelSize})

	varI := gocv.NewMat()
	defer varI.Close()
	gocv.Multiply(guideFloat, guideFloat, &temp)
	gocv.BoxFilter(temp, &varI, -1, image.Point{X: kernelSize, Y: kernelSize})
	meanISquared := gocv.NewMat()
	defer meanISquared.Close()
	gocv.Multiply(meanI, meanI, &meanISquared)
	gocv.Subtract(varI, meanISquared, &varI)

	// Calculate coefficients
	a := gocv.NewMat()
	defer a.Close()
	covIP := gocv.NewMat()
	defer covIP.Close()
	gocv.Multiply(meanI, meanP, &temp)
	gocv.Subtract(corrIP, temp, &covIP)

	denominator := gocv.NewMat()
	defer denominator.Close()
	varI.CopyTo(&denominator)
	denominator.AddFloat(float32(epsilon))
	gocv.Divide(covIP, denominator, &a)

	b := gocv.NewMat()
	defer b.Close()
	gocv.Multiply(a, meanI, &temp)
	gocv.Subtract(meanP, temp, &b)

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
	gocv.Multiply(meanA, guideFloat, &temp)
	gocv.Add(temp, meanB, &resultFloat)

	// Convert back to uint8
	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)

	return result
}

func (l *Lanczos4Transform) iterativeLanczos4(src gocv.Mat, targetWidth, targetHeight int) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", "Starting iterative downscaling")

	current := src.Clone()
	currentWidth, currentHeight := current.Cols(), current.Rows()

	step := 0
	for currentWidth > targetWidth*2 || currentHeight > targetHeight*2 {
		step++
		temp := gocv.NewMat()
		nextWidth := currentWidth / 2
		nextHeight := currentHeight / 2

		l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", fmt.Sprintf("Step %d: %dx%d -> %dx%d",
			step, currentWidth, currentHeight, nextWidth, nextHeight))

		gocv.Resize(current, &temp, image.Point{X: nextWidth, Y: nextHeight}, 0, 0, gocv.InterpolationLanczos4)
		current.Close()
		current = temp
		currentWidth, currentHeight = nextWidth, nextHeight
	}

	// Final resize to exact target dimensions
	scaled := gocv.NewMat()
	gocv.Resize(current, &scaled, image.Point{X: targetWidth, Y: targetHeight}, 0, 0, gocv.InterpolationLanczos4)
	current.Close()

	l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", fmt.Sprintf("Completed in %d steps", step+1))
	return scaled
}

func (l *Lanczos4Transform) calculateScaleFactor() float64 {
	if l.originalDPI > 0 && l.targetDPI > 0 {
		calculated := l.targetDPI / l.originalDPI
		l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Calculated scale factor: %.3f (%.0f DPI -> %.0f DPI)",
			calculated, l.originalDPI, l.targetDPI))
		return calculated
	}
	return l.scaleFactor
}

func (l *Lanczos4Transform) createParameterUI() *fyne.Container {
	// Scale Factor parameter
	scaleLabel := widget.NewLabel("Scale Factor (0.1-5.0):")
	scaleEntry := widget.NewEntry()
	scaleEntry.SetText(fmt.Sprintf("%.2f", l.scaleFactor))
	scaleEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 0.1 && value <= 5.0 {
			oldValue := l.scaleFactor
			l.scaleFactor = value
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Scale factor changed: %.3f -> %.3f", oldValue, value))
			if l.onParameterChanged != nil {
				l.onParameterChanged()
			}
		} else {
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Invalid scale factor: %s", text))
		}
	}

	// Target DPI parameter
	targetDPILabel := widget.NewLabel("Target DPI (72-1200):")
	targetDPIEntry := widget.NewEntry()
	targetDPIEntry.SetText(fmt.Sprintf("%.0f", l.targetDPI))
	targetDPIEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 72 && value <= 1200 {
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
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Invalid target DPI: %s", text))
		}
	}

	// Original DPI parameter
	originalDPILabel := widget.NewLabel("Original DPI (72-1200):")
	originalDPIEntry := widget.NewEntry()
	originalDPIEntry.SetText(fmt.Sprintf("%.0f", l.originalDPI))
	originalDPIEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 72 && value <= 1200 {
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
			l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Invalid original DPI: %s", text))
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
