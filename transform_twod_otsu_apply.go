package main

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

func (t *TwoDOtsu) applyWithScale(src gocv.Mat, scale float64) gocv.Mat {
	defer func() {
		if r := recover(); r != nil {
			t.debugImage.LogError(fmt.Errorf("panic in 2D Otsu: %v", r))
		}
	}()

	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Starting binarization (scale: %.2f)", scale))

	if src.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Input matrix is empty")
		return gocv.NewMat()
	}

	t.paramMutex.RLock()
	windowRadius := t.windowRadius
	epsilon := t.epsilon
	morphKernelSize := t.morphKernelSize
	t.paramMutex.RUnlock()

	var workingImage gocv.Mat
	if scale != 1.0 {
		newWidth := int(float64(src.Cols()) * scale)
		newHeight := int(float64(src.Rows()) * scale)

		if newWidth <= 0 || newHeight <= 0 || newWidth > 16384 || newHeight > 16384 {
			t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("ERROR: Invalid scaled dimensions: %dx%d", newWidth, newHeight))
			return gocv.NewMat()
		}

		workingImage = gocv.NewMat()
		err := gocv.Resize(src, &workingImage, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)
		if err != nil {
			t.debugImage.LogError(err)
			workingImage.Close()
			return gocv.NewMat()
		}
		t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Scaled to %dx%d", newWidth, newHeight))
	} else {
		workingImage = src.Clone()
	}
	defer workingImage.Close()

	var grayscale gocv.Mat
	if workingImage.Channels() > 1 {
		grayscale = gocv.NewMat()
		err := gocv.CvtColor(workingImage, &grayscale, gocv.ColorBGRToGray)
		if err != nil {
			t.debugImage.LogError(err)
			grayscale.Close()
			return gocv.NewMat()
		}
	} else {
		grayscale = workingImage.Clone()
	}
	defer grayscale.Close()

	if grayscale.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Grayscale conversion failed")
		return gocv.NewMat()
	}

	t.debugImage.LogMatInfo("grayscale", grayscale)

	guided := t.applyGuidedFilter(grayscale, windowRadius, epsilon)
	defer guided.Close()

	if guided.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Guided filter failed")
		guided = grayscale.Clone()
	}

	binaryResult := t.apply2DOtsu(grayscale, guided)
	if binaryResult.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Binarization failed")
		return gocv.NewMat()
	}

	processed := t.applyMorphologicalOps(binaryResult, morphKernelSize)
	binaryResult.Close()

	if processed.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Morphological operations failed")
		return gocv.NewMat()
	}

	var result gocv.Mat
	if scale != 1.0 {
		result = gocv.NewMat()
		err := gocv.Resize(processed, &result, image.Point{X: src.Cols(), Y: src.Rows()}, 0, 0, gocv.InterpolationLinear)
		processed.Close()
		if err != nil {
			t.debugImage.LogError(err)
			result.Close()
			return gocv.NewMat()
		}
		t.debugImage.LogAlgorithmStep("2D Otsu", "Scaled back to original size")
	} else {
		result = processed
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", "Completed successfully")
	return result
}
