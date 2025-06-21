package main

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"
)

func (l *Lanczos4Transform) iterativeLanczos4(src gocv.Mat, targetWidth, targetHeight int) gocv.Mat {
	if src.Empty() || targetWidth <= 0 || targetHeight <= 0 {
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", "Starting iterative downscaling")

	current := src.Clone()
	defer current.Close()
	currentWidth, currentHeight := current.Cols(), current.Rows()

	step := 0
	maxSteps := 15
	scalingFactor := 0.6

	for (currentWidth > targetWidth*2 || currentHeight > targetHeight*2) && step < maxSteps {
		step++

		nextWidth := int(math.Max(float64(currentWidth)*scalingFactor, float64(targetWidth)))
		nextHeight := int(math.Max(float64(currentHeight)*scalingFactor, float64(targetHeight)))

		if nextWidth >= currentWidth || nextHeight >= currentHeight {
			break
		}

		l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", fmt.Sprintf("Step %d: %dx%d -> %dx%d",
			step, currentWidth, currentHeight, nextWidth, nextHeight))

		temp := gocv.NewMat()

		interpolation := gocv.InterpolationArea
		if nextWidth > currentWidth || nextHeight > currentHeight {
			interpolation = gocv.InterpolationLanczos4
		}

		err := gocv.Resize(current, &temp, image.Point{X: nextWidth, Y: nextHeight}, 0, 0, interpolation)
		if err != nil {
			l.debugImage.LogError(err)
			temp.Close()
			break
		}

		current.Close()
		current = temp
		currentWidth, currentHeight = nextWidth, nextHeight
	}

	scaled := gocv.NewMat()
	err := gocv.Resize(current, &scaled, image.Point{X: targetWidth, Y: targetHeight}, 0, 0, gocv.InterpolationLanczos4)
	if err != nil {
		l.debugImage.LogError(err)
		scaled.Close()
		return gocv.NewMat()
	}

	l.debugImage.LogAlgorithmStep("Lanczos4 Iterative", fmt.Sprintf("Completed in %d steps", step+1))
	return scaled
}
