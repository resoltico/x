package main

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

func (t *TwoDOtsu) applyGuidedFilter(src gocv.Mat, windowRadius int, epsilon float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("GuidedFilter", "Starting guided filter with covariance")

	if src.Empty() {
		return gocv.NewMat()
	}

	if epsilon <= 0 {
		epsilon = 0.001
	}

	srcFloat := gocv.NewMat()
	defer srcFloat.Close()
	src.ConvertTo(&srcFloat, gocv.MatTypeCV32F)
	srcFloat.DivideFloat(255.0)

	kernelSize := 2*windowRadius + 1

	meanI := gocv.NewMat()
	defer meanI.Close()
	err := gocv.Blur(srcFloat, &meanI, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	correlation := gocv.NewMat()
	defer correlation.Close()
	err = gocv.Multiply(srcFloat, srcFloat, &correlation)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	meanCorr := gocv.NewMat()
	defer meanCorr.Close()
	err = gocv.Blur(correlation, &meanCorr, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	meanISquared := gocv.NewMat()
	defer meanISquared.Close()
	err = gocv.Multiply(meanI, meanI, &meanISquared)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	varI := gocv.NewMat()
	defer varI.Close()
	err = gocv.Subtract(meanCorr, meanISquared, &varI)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	a := gocv.NewMat()
	defer a.Close()
	varIPlusEps := gocv.NewMat()
	defer varIPlusEps.Close()
	varI.CopyTo(&varIPlusEps)
	varIPlusEps.AddFloat(float32(epsilon))
	err = gocv.Divide(varI, varIPlusEps, &a)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	b := gocv.NewMat()
	defer b.Close()
	ones := gocv.NewMatWithSize(a.Rows(), a.Cols(), a.Type())
	defer ones.Close()
	ones.SetTo(gocv.NewScalar(1, 0, 0, 0))

	oneMinusA := gocv.NewMat()
	defer oneMinusA.Close()
	err = gocv.Subtract(ones, a, &oneMinusA)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	err = gocv.Multiply(meanI, oneMinusA, &b)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	meanA := gocv.NewMat()
	defer meanA.Close()
	err = gocv.Blur(a, &meanA, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	meanB := gocv.NewMat()
	defer meanB.Close()
	err = gocv.Blur(b, &meanB, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	resultFloat := gocv.NewMat()
	defer resultFloat.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	err = gocv.Multiply(meanA, srcFloat, &temp)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	err = gocv.Add(temp, meanB, &resultFloat)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)

	t.debugImage.LogFilter("GuidedFilter", fmt.Sprintf("radius=%d epsilon=%.3f", windowRadius, epsilon))
	return result
}
