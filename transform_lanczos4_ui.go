package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (l *Lanczos4Transform) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	l.onParameterChanged = onParameterChanged
	return l.createParameterUI()
}

func (l *Lanczos4Transform) createParameterUI() *fyne.Container {
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

	targetDPILabel := widget.NewLabel("Target DPI (72-2400):")
	targetDPIEntry := widget.NewEntry()
	targetDPIEntry.SetText(fmt.Sprintf("%.0f", l.targetDPI))
	targetDPIEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 72 && value <= 2400 {
			oldValue := l.targetDPI
			l.targetDPI = value
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

	originalDPILabel := widget.NewLabel("Original DPI (72-2400):")
	originalDPIEntry := widget.NewEntry()
	originalDPIEntry.SetText(fmt.Sprintf("%.0f", l.originalDPI))
	originalDPIEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value >= 72 && value <= 2400 {
			oldValue := l.originalDPI
			l.originalDPI = value
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

	iterativeCheck := widget.NewCheck("Use Iterative Downscaling", func(checked bool) {
		oldValue := l.useIterative
		l.useIterative = checked
		l.debugImage.LogAlgorithmStep("Lanczos4 Parameters", fmt.Sprintf("Iterative mode changed: %t -> %t", oldValue, checked))
		if l.onParameterChanged != nil {
			l.onParameterChanged()
		}
	})
	iterativeCheck.SetChecked(l.useIterative)

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
