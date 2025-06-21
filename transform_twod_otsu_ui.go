package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (t *TwoDOtsu) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	t.onParameterChanged = onParameterChanged
	return t.createParameterUI()
}

func (t *TwoDOtsu) createParameterUI() *fyne.Container {
	radiusLabel := widget.NewLabel("Window Radius (1-20):")
	radiusEntry := widget.NewEntry()

	t.paramMutex.RLock()
	radiusEntry.SetText(fmt.Sprintf("%d", t.windowRadius))
	t.paramMutex.RUnlock()

	radiusEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value >= 1 && value <= 20 {
			t.paramMutex.Lock()
			oldValue := t.windowRadius
			t.windowRadius = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Window radius changed: %d -> %d", oldValue, value))
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid window radius: %s (must be 1-20)", text))
		}
	}

	epsilonLabel := widget.NewLabel("Epsilon (0.001-1.0):")
	epsilonEntry := widget.NewEntry()

	t.paramMutex.RLock()
	epsilonEntry.SetText(fmt.Sprintf("%.3f", t.epsilon))
	t.paramMutex.RUnlock()

	epsilonEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value > 0.001 && value <= 1.0 {
			t.paramMutex.Lock()
			oldValue := t.epsilon
			t.epsilon = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Epsilon changed: %.3f -> %.3f", oldValue, value))
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid epsilon: %s (must be 0.001-1.0)", text))
		}
	}

	kernelLabel := widget.NewLabel("Morphological Kernel Size (1-15, odd):")
	kernelEntry := widget.NewEntry()

	t.paramMutex.RLock()
	kernelEntry.SetText(fmt.Sprintf("%d", t.morphKernelSize))
	t.paramMutex.RUnlock()

	kernelEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value >= 1 && value <= 15 && value%2 == 1 {
			t.paramMutex.Lock()
			oldValue := t.morphKernelSize
			t.morphKernelSize = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Morphological kernel size changed: %d -> %d", oldValue, value))
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid kernel size: %s (must be 1-15 and odd)", text))
		}
	}

	noiseReductionCheck := widget.NewCheck("Enable Historical Noise Reduction", func(checked bool) {
		t.paramMutex.Lock()
		oldValue := t.noiseReduction
		t.noiseReduction = checked
		t.paramMutex.Unlock()

		t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Noise reduction changed: %t -> %t", oldValue, checked))
		if t.onParameterChanged != nil {
			t.onParameterChanged()
		}
	})

	t.paramMutex.RLock()
	noiseReductionCheck.SetChecked(t.noiseReduction)
	t.paramMutex.RUnlock()

	integralImageCheck := widget.NewCheck("Use Integral Image Acceleration", func(checked bool) {
		t.paramMutex.Lock()
		oldValue := t.useIntegralImage
		t.useIntegralImage = checked
		t.paramMutex.Unlock()

		t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Integral image acceleration changed: %t -> %t", oldValue, checked))
		if t.onParameterChanged != nil {
			t.onParameterChanged()
		}
	})

	t.paramMutex.RLock()
	integralImageCheck.SetChecked(t.useIntegralImage)
	t.paramMutex.RUnlock()

	adaptiveRegionsLabel := widget.NewLabel("Adaptive Regions (1-8):")
	adaptiveRegionsEntry := widget.NewEntry()

	t.paramMutex.RLock()
	adaptiveRegionsEntry.SetText(fmt.Sprintf("%d", t.adaptiveRegions))
	t.paramMutex.RUnlock()

	adaptiveRegionsEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value >= 1 && value <= 8 {
			t.paramMutex.Lock()
			oldValue := t.adaptiveRegions
			t.adaptiveRegions = value
			t.paramMutex.Unlock()

			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Adaptive regions changed: %d -> %d", oldValue, value))
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		} else {
			t.debugImage.LogAlgorithmStep("2D Otsu Parameters", fmt.Sprintf("Invalid adaptive regions: %s (must be 1-8)", text))
		}
	}

	return container.NewVBox(
		radiusLabel, radiusEntry,
		epsilonLabel, epsilonEntry,
		kernelLabel, kernelEntry,
		noiseReductionCheck,
		integralImageCheck,
		adaptiveRegionsLabel, adaptiveRegionsEntry,
	)
}
