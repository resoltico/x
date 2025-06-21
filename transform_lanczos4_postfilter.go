package main

import (
	"fmt"

	"gocv.io/x/gocv"
)

func (l *Lanczos4Transform) applyPostFilter(src gocv.Mat) gocv.Mat {
	if src.Empty() {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 PostFilter", "Applying bilateral filter for artifact reduction")

	filtered := gocv.NewMat()

	minDim := min(src.Cols(), src.Rows())
	var d int
	var sigmaColor, sigmaSpace float64

	if minDim < 500 {
		d = 5
		sigmaColor = 50.0
		sigmaSpace = 50.0
	} else if minDim < 1000 {
		d = 7
		sigmaColor = 75.0
		sigmaSpace = 75.0
	} else {
		d = 9
		sigmaColor = 100.0
		sigmaSpace = 100.0
	}

	err := gocv.BilateralFilter(src, &filtered, d, sigmaColor, sigmaSpace)
	if err != nil {
		l.debugImage.LogError(err)
		filtered.Close()
		return src.Clone()
	}

	l.debugImage.LogFilter("BilateralFilter", fmt.Sprintf("d=%d sigmaColor=%.1f sigmaSpace=%.1f", d, sigmaColor, sigmaSpace))
	return filtered
}
