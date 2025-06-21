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
	noiseReduction := t.noiseReduction
	useIntegralImage := t.useIntegralImage
	adaptiveRegions := t.adaptiveRegions
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

	// Apply noise reduction for historical images if enabled
	var denoisedGray gocv.Mat
	if noiseReduction {
		denoisedGray = t.applyHistoricalNoiseReduction(grayscale)
		defer denoisedGray.Close()
	} else {
		denoisedGray = grayscale.Clone()
		defer denoisedGray.Close()
	}

	// Choose between adaptive regional processing or global processing
	var binaryResult gocv.Mat
	if adaptiveRegions > 1 {
		binaryResult = t.applyAdaptiveRegional2DOtsu(denoisedGray, adaptiveRegions, windowRadius, epsilon, useIntegralImage)
	} else {
		guided := t.applyGuidedFilter(denoisedGray, windowRadius, epsilon)
		defer guided.Close()

		if guided.Empty() {
			t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Guided filter failed")
			guided = denoisedGray.Clone()
		}

		if useIntegralImage {
			binaryResult = t.apply2DOtsuWithIntegralImage(denoisedGray, guided)
		} else {
			binaryResult = t.apply2DOtsu(denoisedGray, guided)
		}
	}

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

func (t *TwoDOtsu) applyHistoricalNoiseReduction(src gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("Historical Noise Reduction", "Applying bilateral filter for salt-and-pepper noise")

	// Use bilateral filter to preserve edges while reducing noise
	bilateralFiltered := gocv.NewMat()
	err := gocv.BilateralFilter(src, &bilateralFiltered, 9, 75.0, 75.0)
	if err != nil {
		t.debugImage.LogError(err)
		bilateralFiltered.Close()
		return src.Clone()
	}

	// Apply median blur to further reduce salt-and-pepper noise
	medianFiltered := gocv.NewMat()
	err = gocv.MedianBlur(bilateralFiltered, &medianFiltered, 3)
	bilateralFiltered.Close()
	if err != nil {
		t.debugImage.LogError(err)
		medianFiltered.Close()
		return src.Clone()
	}

	t.debugImage.LogFilter("HistoricalNoiseReduction", "bilateral+median")
	return medianFiltered
}

func (t *TwoDOtsu) applyAdaptiveRegional2DOtsu(src gocv.Mat, regions int, windowRadius int, epsilon float64, useIntegralImage bool) gocv.Mat {
	t.debugImage.LogAlgorithmStep("Adaptive Regional 2D Otsu", fmt.Sprintf("Processing %d regions", regions))

	size := src.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	regionWidth := width / regions
	regionHeight := height / regions

	for i := 0; i < regions; i++ {
		for j := 0; j < regions; j++ {
			x1 := i * regionWidth
			y1 := j * regionHeight
			x2 := min((i+1)*regionWidth, width)
			y2 := min((j+1)*regionHeight, height)

			if x2 <= x1 || y2 <= y1 {
				continue
			}

			roi := src.Region(image.Rect(x1, y1, x2, y2))
			guided := t.applyGuidedFilter(roi, windowRadius, epsilon)

			var regionResult gocv.Mat
			if useIntegralImage {
				regionResult = t.apply2DOtsuWithIntegralImage(roi, guided)
			} else {
				regionResult = t.apply2DOtsu(roi, guided)
			}

			roi.Close()
			guided.Close()

			if !regionResult.Empty() {
				resultROI := result.Region(image.Rect(x1, y1, x2, y2))
				regionResult.CopyTo(&resultROI)
				resultROI.Close()
			}
			regionResult.Close()
		}
	}

	t.debugImage.LogAlgorithmStep("Adaptive Regional 2D Otsu", "Regional processing completed")
	return result
}
