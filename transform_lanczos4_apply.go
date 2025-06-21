package main

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"
)

func (l *Lanczos4Transform) applyLanczos4(src gocv.Mat, scale float64) gocv.Mat {
	if src.Empty() {
		l.debugImage.LogAlgorithmStep("Lanczos4", "ERROR: Input matrix is empty")
		return gocv.NewMat()
	}

	if scale <= 0 || math.IsInf(scale, 0) || math.IsNaN(scale) {
		l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("ERROR: Invalid scale factor: %.3f", scale))
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Input: %dx%d, scale: %.2f", src.Cols(), src.Rows(), scale))

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

	if working.Cols() <= 0 || working.Rows() <= 0 {
		l.debugImage.LogAlgorithmStep("Lanczos4", "ERROR: Invalid input dimensions")
		return gocv.NewMat()
	}

	filtered := l.applyPreFilter(working)
	defer filtered.Close()

	newWidth := int(math.Round(float64(filtered.Cols()) * scale))
	newHeight := int(math.Round(float64(filtered.Rows()) * scale))

	if newWidth <= 0 || newHeight <= 0 {
		l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("ERROR: Invalid target dimensions: %dx%d", newWidth, newHeight))
		return gocv.NewMat()
	}

	maxDimension := 32768
	if newWidth > maxDimension || newHeight > maxDimension {
		l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("ERROR: Target dimensions too large: %dx%d (max: %d)", newWidth, newHeight, maxDimension))
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Target dimensions: %dx%d", newWidth, newHeight))

	var result gocv.Mat

	if l.useIterative && scale < 0.5 {
		result = l.iterativeLanczos4(filtered, newWidth, newHeight)
	} else {
		result = gocv.NewMat()
		err := gocv.Resize(filtered, &result, image.Point{X: newWidth, Y: newHeight}, 0, 0, gocv.InterpolationLanczos4)
		if err != nil {
			l.debugImage.LogError(err)
			result.Close()
			return gocv.NewMat()
		}
	}

	if result.Empty() {
		l.debugImage.LogAlgorithmStep("Lanczos4", "ERROR: Scaling operation failed")
		return gocv.NewMat()
	}

	final := l.applyPostFilter(result)
	result.Close()

	l.debugImage.LogMatInfo("final_result", final)
	l.debugImage.LogAlgorithmStep("Lanczos4", "Scaling completed successfully")

	return final
}
