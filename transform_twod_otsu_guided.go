package main

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

func (t *TwoDOtsu) applyGuidedFilter(src gocv.Mat, windowRadius int, epsilon float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("GuidedFilter", "Starting guided filter with covariance")
	t.debugPerf.StartOperation("GuidedFilter_Complete", fmt.Sprintf("radius=%d,eps=%.3f", windowRadius, epsilon))
	defer t.debugPerf.EndOperation("GuidedFilter_Complete")

	if src.Empty() {
		return gocv.NewMat()
	}

	if epsilon <= 0 {
		epsilon = 0.001
	}

	t.debugPerf.StartOperation("GuidedFilter_Conversion", "float_conversion")
	srcFloat := gocv.NewMat()
	defer srcFloat.Close()
	src.ConvertTo(&srcFloat, gocv.MatTypeCV32F)
	srcFloat.DivideFloat(255.0)
	t.debugPerf.LogMatrixOperation("ConvertToFloat", src, srcFloat)
	t.debugPerf.EndOperation("GuidedFilter_Conversion")

	kernelSize := 2*windowRadius + 1
	t.debugPerf.LogStep("GuidedFilter_Complete", "Kernel setup", fmt.Sprintf("kernel_size=%dx%d", kernelSize, kernelSize))

	t.debugPerf.StartOperation("GuidedFilter_MeanI", "blur_operation")
	meanI := gocv.NewMat()
	defer meanI.Close()
	err := gocv.Blur(srcFloat, &meanI, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_MeanI")
		return gocv.NewMat()
	}
	t.debugPerf.LogMatrixOperation("MeanI", srcFloat, meanI)
	t.debugPerf.EndOperation("GuidedFilter_MeanI")

	t.debugPerf.StartOperation("GuidedFilter_Correlation", "correlation_calculation")
	correlation := gocv.NewMat()
	defer correlation.Close()
	err = gocv.Multiply(srcFloat, srcFloat, &correlation)
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_Correlation")
		return gocv.NewMat()
	}

	meanCorr := gocv.NewMat()
	defer meanCorr.Close()
	err = gocv.Blur(correlation, &meanCorr, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_Correlation")
		return gocv.NewMat()
	}
	t.debugPerf.LogMatrixOperation("MeanCorrelation", correlation, meanCorr)
	t.debugPerf.EndOperation("GuidedFilter_Correlation")

	t.debugPerf.StartOperation("GuidedFilter_Variance", "variance_calculation")
	meanISquared := gocv.NewMat()
	defer meanISquared.Close()
	err = gocv.Multiply(meanI, meanI, &meanISquared)
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_Variance")
		return gocv.NewMat()
	}

	varI := gocv.NewMat()
	defer varI.Close()
	err = gocv.Subtract(meanCorr, meanISquared, &varI)
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_Variance")
		return gocv.NewMat()
	}
	t.debugPerf.LogMatrixOperation("Variance", meanCorr, varI)
	t.debugPerf.EndOperation("GuidedFilter_Variance")

	t.debugPerf.StartOperation("GuidedFilter_Coefficients", "a_b_coefficient_calculation")
	a := gocv.NewMat()
	defer a.Close()
	varIPlusEps := gocv.NewMat()
	defer varIPlusEps.Close()
	varI.CopyTo(&varIPlusEps)
	varIPlusEps.AddFloat(float32(epsilon))
	err = gocv.Divide(varI, varIPlusEps, &a)
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_Coefficients")
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
		t.debugPerf.EndOperation("GuidedFilter_Coefficients")
		return gocv.NewMat()
	}

	err = gocv.Multiply(meanI, oneMinusA, &b)
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_Coefficients")
		return gocv.NewMat()
	}
	t.debugPerf.LogMatrixOperation("Coefficients", a, b)
	t.debugPerf.EndOperation("GuidedFilter_Coefficients")

	t.debugPerf.StartOperation("GuidedFilter_MeanCoeff", "coefficient_smoothing")
	meanA := gocv.NewMat()
	defer meanA.Close()
	err = gocv.Blur(a, &meanA, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_MeanCoeff")
		return gocv.NewMat()
	}

	meanB := gocv.NewMat()
	defer meanB.Close()
	err = gocv.Blur(b, &meanB, image.Point{X: kernelSize, Y: kernelSize})
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_MeanCoeff")
		return gocv.NewMat()
	}
	t.debugPerf.LogMatrixOperation("MeanCoefficients", meanA, meanB)
	t.debugPerf.EndOperation("GuidedFilter_MeanCoeff")

	t.debugPerf.StartOperation("GuidedFilter_FinalCalc", "final_result_calculation")
	resultFloat := gocv.NewMat()
	defer resultFloat.Close()
	temp := gocv.NewMat()
	defer temp.Close()
	err = gocv.Multiply(meanA, srcFloat, &temp)
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_FinalCalc")
		return gocv.NewMat()
	}

	err = gocv.Add(temp, meanB, &resultFloat)
	if err != nil {
		t.debugImage.LogError(err)
		t.debugPerf.EndOperation("GuidedFilter_FinalCalc")
		return gocv.NewMat()
	}

	result := gocv.NewMat()
	resultFloat.MultiplyFloat(255.0)
	resultFloat.ConvertTo(&result, gocv.MatTypeCV8U)
	t.debugPerf.LogMatrixOperation("FinalResult", resultFloat, result)
	t.debugPerf.EndOperation("GuidedFilter_FinalCalc")

	t.debugImage.LogFilter("GuidedFilter", fmt.Sprintf("radius=%d epsilon=%.3f", windowRadius, epsilon))
	t.debugPerf.LogStep("GuidedFilter_Complete", "Filter completed successfully", fmt.Sprintf("output_size=%dx%d", result.Cols(), result.Rows()))
	return result
}
