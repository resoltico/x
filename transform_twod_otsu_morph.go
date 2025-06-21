package main

import (
	"image"

	"gocv.io/x/gocv"
)

func (t *TwoDOtsu) applyMorphologicalOps(src gocv.Mat, morphKernelSize int) gocv.Mat {
	if morphKernelSize <= 1 {
		return src.Clone()
	}

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: morphKernelSize, Y: morphKernelSize})
	defer kernel.Close()

	closed := gocv.NewMat()
	defer closed.Close()
	err := gocv.MorphologyEx(src, &closed, gocv.MorphClose, kernel)
	if err != nil {
		t.debugImage.LogError(err)
		return src.Clone()
	}

	result := gocv.NewMat()
	err = gocv.MorphologyEx(closed, &result, gocv.MorphOpen, kernel)
	if err != nil {
		t.debugImage.LogError(err)
		result.Close()
		return closed.Clone()
	}

	return result
}
