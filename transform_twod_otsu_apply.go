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

	t.debugPerf.StartOperation("2D_Otsu_Complete", fmt.Sprintf("scale=%.2f", scale))
	defer t.debugPerf.EndOperation("2D_Otsu_Complete")

	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Starting binarization (scale: %.2f)", scale))
	t.debugPerf.LogAlgorithmPhase("2D Otsu", "Initialization", src)

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

	t.debugPerf.LogStep("2D_Otsu_Complete", "Parameters loaded", fmt.Sprintf("radius=%d, epsilon=%.3f, regions=%d", windowRadius, epsilon, adaptiveRegions))

	var workingImage gocv.Mat
	if scale != 1.0 {
		t.debugPerf.StartOperation("2D_Otsu_Scaling", "input_resize")
		newWidth := int(float64(src.Cols()) * scale)
		newHeight := int(float64(src.Rows()) * scale)

		if newWidth <= 0 || newHeight <= 0 || newWidth > 16384 || newHeight > 16384 {
			t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("ERROR: Invalid scaled dimensions: %dx%d", newWidth, newHeight))
			t.debugPerf.EndOperation("2D_Otsu_Scaling")
			return gocv.NewMat()
		}

		workingImage = gocv.NewMat()
		err := gocv.Resize(src, &workingImage, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLinear)
		if err != nil {
			t.debugImage.LogError(err)
			workingImage.Close()
			t.debugPerf.EndOperation("2D_Otsu_Scaling")
			return gocv.NewMat()
		}
		t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Scaled to %dx%d", newWidth, newHeight))
		t.debugPerf.LogMatrixOperation("Resize", src, workingImage)
		t.debugPerf.EndOperation("2D_Otsu_Scaling")
	} else {
		workingImage = src.Clone()
	}
	defer workingImage.Close()

	t.debugPerf.StartOperation("2D_Otsu_Grayscale", "color_conversion")
	var grayscale gocv.Mat
	if workingImage.Channels() > 1 {
		grayscale = gocv.NewMat()
		err := gocv.CvtColor(workingImage, &grayscale, gocv.ColorBGRToGray)
		if err != nil {
			t.debugImage.LogError(err)
			grayscale.Close()
			t.debugPerf.EndOperation("2D_Otsu_Grayscale")
			return gocv.NewMat()
		}
	} else {
		grayscale = workingImage.Clone()
	}
	defer grayscale.Close()
	t.debugPerf.LogMatrixOperation("CvtColor", workingImage, grayscale)
	t.debugPerf.EndOperation("2D_Otsu_Grayscale")

	if grayscale.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Grayscale conversion failed")
		return gocv.NewMat()
	}

	t.debugImage.LogMatInfo("grayscale", grayscale)

	// Apply noise reduction for historical images if enabled
	var denoisedGray gocv.Mat
	if noiseReduction {
		t.debugPerf.StartOperation("2D_Otsu_NoiseReduction", "bilateral_median_filter")
		denoisedGray = t.applyHistoricalNoiseReduction(grayscale)
		t.debugPerf.LogMatrixOperation("NoiseReduction", grayscale, denoisedGray)
		t.debugPerf.EndOperation("2D_Otsu_NoiseReduction")
		defer denoisedGray.Close()
	} else {
		denoisedGray = grayscale.Clone()
		defer denoisedGray.Close()
	}

	// Choose between adaptive regional processing or global processing
	var binaryResult gocv.Mat
	if adaptiveRegions > 1 {
		t.debugPerf.StartOperation("2D_Otsu_AdaptiveRegional", fmt.Sprintf("regions=%d", adaptiveRegions))
		t.debugPerf.LogStep("2D_Otsu_AdaptiveRegional", "Starting regional processing", fmt.Sprintf("regions=%d", adaptiveRegions))
		binaryResult = t.applyAdaptiveRegional2DOtsu(denoisedGray, adaptiveRegions, windowRadius, epsilon, useIntegralImage)
		t.debugPerf.LogMatrixOperation("AdaptiveRegional", denoisedGray, binaryResult)
		t.debugPerf.EndOperation("2D_Otsu_AdaptiveRegional")
	} else {
		t.debugPerf.StartOperation("2D_Otsu_Global", "single_region")

		t.debugPerf.StartOperation("2D_Otsu_GuidedFilter", "edge_preserving_smooth")
		guided := t.applyGuidedFilter(denoisedGray, windowRadius, epsilon)
		t.debugPerf.LogMatrixOperation("GuidedFilter", denoisedGray, guided)
		t.debugPerf.EndOperation("2D_Otsu_GuidedFilter")
		defer guided.Close()

		if guided.Empty() {
			t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Guided filter failed")
			guided = denoisedGray.Clone()
		}

		if useIntegralImage {
			t.debugPerf.StartOperation("2D_Otsu_Integral", "optimized_algorithm")
			size := denoisedGray.Size()
			t.debugPerf.LogHistogramOperation("2D_Otsu_Integral", size, 256*256)
			binaryResult = t.apply2DOtsuWithIntegralImage(denoisedGray, guided)
			t.debugPerf.LogMatrixOperation("IntegralOtsu", denoisedGray, binaryResult)
			t.debugPerf.EndOperation("2D_Otsu_Integral")
		} else {
			t.debugPerf.StartOperation("2D_Otsu_Standard", "standard_algorithm")
			size := denoisedGray.Size()
			t.debugPerf.LogHistogramOperation("2D_Otsu_Standard", size, 256*256)
			binaryResult = t.apply2DOtsu(denoisedGray, guided)
			t.debugPerf.LogMatrixOperation("StandardOtsu", denoisedGray, binaryResult)
			t.debugPerf.EndOperation("2D_Otsu_Standard")
		}
		t.debugPerf.EndOperation("2D_Otsu_Global")
	}

	if binaryResult.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Binarization failed")
		return gocv.NewMat()
	}

	t.debugPerf.StartOperation("2D_Otsu_Morphology", "cleanup_operations")
	processed := t.applyMorphologicalOps(binaryResult, morphKernelSize)
	t.debugPerf.LogMatrixOperation("Morphology", binaryResult, processed)
	t.debugPerf.EndOperation("2D_Otsu_Morphology")
	binaryResult.Close()

	if processed.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Morphological operations failed")
		return gocv.NewMat()
	}

	var result gocv.Mat
	if scale != 1.0 {
		t.debugPerf.StartOperation("2D_Otsu_FinalResize", "restore_original_size")
		result = gocv.NewMat()
		err := gocv.Resize(processed, &result, image.Point{X: src.Cols(), Y: src.Rows()}, 0, 0, gocv.InterpolationLinear)
		processed.Close()
		if err != nil {
			t.debugImage.LogError(err)
			result.Close()
			t.debugPerf.EndOperation("2D_Otsu_FinalResize")
			return gocv.NewMat()
		}
		t.debugImage.LogAlgorithmStep("2D Otsu", "Scaled back to original size")
		t.debugPerf.LogMatrixOperation("FinalResize", processed, result)
		t.debugPerf.EndOperation("2D_Otsu_FinalResize")
	} else {
		result = processed
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", "Completed successfully")
	t.debugPerf.LogStep("2D_Otsu_Complete", "Algorithm completed", fmt.Sprintf("output=%dx%d", result.Cols(), result.Rows()))
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
	t.debugPerf.LogStep("2D_Otsu_AdaptiveRegional", "Region setup", fmt.Sprintf("total_regions=%d", regions*regions))

	size := src.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	regionWidth := width / regions
	regionHeight := height / regions

	totalRegions := regions * regions
	processedRegions := 0

	for i := 0; i < regions; i++ {
		for j := 0; j < regions; j++ {
			regionName := fmt.Sprintf("Region_%d_%d", i, j)
			t.debugPerf.StartOperation(regionName, fmt.Sprintf("processing_region_%d_of_%d", processedRegions+1, totalRegions))

			x1 := i * regionWidth
			y1 := j * regionHeight
			x2 := min((i+1)*regionWidth, width)
			y2 := min((j+1)*regionHeight, height)

			if x2 <= x1 || y2 <= y1 {
				t.debugPerf.EndOperation(regionName)
				continue
			}

			t.debugPerf.LogStep(regionName, "ROI extraction", fmt.Sprintf("rect=(%d,%d,%d,%d)", x1, y1, x2-x1, y2-y1))

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

			processedRegions++
			t.debugPerf.LogStep("2D_Otsu_AdaptiveRegional", "Region completed", fmt.Sprintf("progress=%d/%d", processedRegions, totalRegions))
			t.debugPerf.EndOperation(regionName)
		}
	}

	t.debugImage.LogAlgorithmStep("Adaptive Regional 2D Otsu", "Regional processing completed")
	return result
}
