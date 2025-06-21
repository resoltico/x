package main

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

func (l *Lanczos4Transform) applyPreFilter(src gocv.Mat) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 PreFilter", "Applying adaptive Gaussian blur")

	minDim := min(src.Cols(), src.Rows())
	var kernelSize int
	if minDim < 100 {
		l.debugImage.LogFilter("PreFilter", "Skipping for small image")
		return src.Clone()
	} else if minDim < 500 {
		kernelSize = 3
	} else if minDim < 1000 {
		kernelSize = 5
	} else {
		kernelSize = 7
	}

	if kernelSize%2 == 0 {
		kernelSize++
	}

	blurred := gocv.NewMat()
	sigma := float64(kernelSize) / 6.0
	err := gocv.GaussianBlur(src, &blurred, image.Point{X: kernelSize, Y: kernelSize}, sigma, sigma, gocv.BorderDefault)
	if err != nil {
		l.debugImage.LogError(err)
		blurred.Close()
		return src.Clone()
	}

	l.debugImage.LogFilter("GaussianBlur", fmt.Sprintf("kernel=%dx%d", kernelSize, kernelSize), fmt.Sprintf("sigma=%.2f", sigma))
	return blurred
}
