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

// Lanczos4Transform implements Lanczos4 interpolation for high-quality image scaling
type Lanczos4Transform struct {
	debugImage   *DebugImage
	paramMutex   sync.RWMutex
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
	l.paramMutex.RLock()
	defer l.paramMutex.RUnlock()

	return map[string]interface{}{
		"scaleFactor":  l.scaleFactor,
		"targetDPI":    l.targetDPI,
		"originalDPI":  l.originalDPI,
		"useIterative": l.useIterative,
	}
}

func (l *Lanczos4Transform) SetParameters(params map[string]interface{}) {
	l.paramMutex.Lock()
	defer l.paramMutex.Unlock()

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
	l.paramMutex.RLock()
	scale := l.scaleFactor
	l.paramMutex.RUnlock()
	return l.applyLanczos4(src, scale)
}

func (l *Lanczos4Transform) ApplyPreview(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4 Preview", "Starting preview scaling")

	l.paramMutex.RLock()
	scale := l.scaleFactor
	l.paramMutex.RUnlock()

	// For preview, use a smaller scale factor to maintain responsiveness
	previewScale := scale
	if scale > 2.0 {
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

	// Calculate target dimensions
	newWidth := int(float64(working.Cols()) * scale)
	newHeight := int(float64(working.Rows()) * scale)

	l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Target dimensions: %dx%d", newWidth, newHeight))

	var result gocv.Mat

	l.paramMutex.RLock()
	useIterative := l.useIterative
	l.paramMutex.RUnlock()

	if useIterative && scale < 0.5 {
		// Use iterative downscaling for large reductions
		result = l.iterativeLanczos4(working, newWidth, newHeight)
	} else {
		// Standard Lanczos4 scaling without pre-filtering (which degrades quality)
		result = gocv.NewMat()
		gocv.Resize(working, &result, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLanczos4)
	}

	l.debugImage.LogMatInfo("final_result", result)
	l.debugImage.LogAlgorithmStep("Lanczos4", "Scaling completed")

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
	l.paramMutex.RLock()
	defer l.paramMutex.RUnlock()

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

	l.paramMutex.RLock()
	scaleEntry.SetText(fmt.Sprintf("%.2f", l.scaleFactor))
	l.paramMutex.RUnlock()

	scaleEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 0.1 && value <= 5.0 {
			l.paramMutex.Lock()
			oldValue := l.scaleFactor
			l.scaleFactor = value
			l.paramMutex.Unlock()

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

	l.paramMutex.RLock()
	targetDPIEntry.SetText(fmt.Sprintf("%.0f", l.targetDPI))
	l.paramMutex.RUnlock()

	targetDPIEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 72 && value <= 1200 {
			l.paramMutex.Lock()
			oldValue := l.targetDPI
			l.targetDPI = value
			// Recalculate scale factor based on DPI
			if l.originalDPI > 0 {
				l.scaleFactor = l.targetDPI / l.originalDPI
			}
			newScale := l.scaleFactor
			l.paramMutex.Unlock()

			scaleEntry.SetText(fmt.Sprintf("%.2f", newScale))
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

	l.paramMutex.RLock()
	originalDPIEntry.SetText(fmt.Sprintf("%.0f", l.originalDPI))
	l.paramMutex.RUnlock()

	originalDPIEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 72 && value <= 1200 {
			l.paramMutex.Lock()
			oldValue := l.originalDPI
			l.originalDPI = value
			// Recalculate scale factor based on DPI
			l.scaleFactor = l.targetDPI / l.originalDPI
			newScale := l.scaleFactor
			l.paramMutex.Unlock()

			scaleEntry.SetText(fmt.Sprintf("%.2f", newScale))
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
		l.paramMutex.Lock()
		oldValue := l.useIterative
		l.useIterative = checked
		l.paramMutex.Unlock()

		l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Iterative mode changed: %t -> %t", oldValue, checked))
		if l.onParameterChanged != nil {
			l.onParameterChanged()
		}
	})

	l.paramMutex.RLock()
	iterativeCheck.SetChecked(l.useIterative)
	l.paramMutex.RUnlock()

	// Calculate button
	calculateBtn := widget.NewButton("Calculate Scale from DPI", func() {
		l.paramMutex.Lock()
		l.scaleFactor = l.calculateScaleFactorUnsafe()
		newScale := l.scaleFactor
		l.paramMutex.Unlock()

		scaleEntry.SetText(fmt.Sprintf("%.2f", newScale))
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

func (l *Lanczos4Transform) calculateScaleFactorUnsafe() float64 {
	if l.originalDPI > 0 && l.targetDPI > 0 {
		calculated := l.targetDPI / l.originalDPI
		l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Calculated scale factor: %.3f (%.0f DPI -> %.0f DPI)",
			calculated, l.originalDPI, l.targetDPI))
		return calculated
	}
	return l.scaleFactor
}
