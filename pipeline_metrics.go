package main

import (
	"image"
	"math"

	"gocv.io/x/gocv"
)

func (p *ImagePipeline) CalculatePSNR() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.Empty() || p.processedImage.Empty() {
		return 0.0
	}

	orig := p.originalImage.Clone()
	defer orig.Close()
	proc := p.processedImage.Clone()
	defer proc.Close()

	if orig.Rows() != proc.Rows() || orig.Cols() != proc.Cols() {
		return 0.0
	}

	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			temp := gocv.NewMat()
			defer temp.Close()
			err := gocv.CvtColor(proc, &temp, gocv.ColorGrayToBGR)
			if err != nil {
				return 0.0
			}
			proc.Close()
			proc = temp.Clone()
		} else if orig.Type() == gocv.MatTypeCV8U && proc.Channels() == 3 {
			temp := gocv.NewMat()
			defer temp.Close()
			err := gocv.CvtColor(orig, &temp, gocv.ColorBGRToGray)
			if err != nil {
				return 0.0
			}
			orig.Close()
			orig = temp.Clone()
		}
	}

	origFloat := gocv.NewMat()
	defer origFloat.Close()
	procFloat := gocv.NewMat()
	defer procFloat.Close()

	orig.ConvertTo(&origFloat, gocv.MatTypeCV64F)
	proc.ConvertTo(&procFloat, gocv.MatTypeCV64F)

	diff := gocv.NewMat()
	defer diff.Close()

	err := gocv.Subtract(origFloat, procFloat, &diff)
	if err != nil {
		return 0.0
	}

	normValue := gocv.Norm(diff, gocv.NormL2)
	normValueSquared := normValue * normValue
	totalPixels := float64(orig.Total())

	if totalPixels == 0 {
		return 0.0
	}

	mse := normValueSquared / totalPixels

	if mse == 0 {
		return 100.0
	}
	if mse < 1e-15 {
		return 100.0
	}

	if math.IsInf(mse, 0) || math.IsNaN(mse) {
		return 0.0
	}

	maxI := 255.0
	psnr := 20*math.Log10(maxI) - 10*math.Log10(mse)

	if math.IsInf(psnr, 0) || math.IsNaN(psnr) {
		return 100.0
	}
	if psnr > 100 {
		return 100.0
	}
	if psnr < 0 {
		return 0.0
	}

	return psnr
}

func (p *ImagePipeline) CalculateSSIM() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.Empty() || p.processedImage.Empty() {
		return 0.0
	}

	orig := p.originalImage.Clone()
	defer orig.Close()
	proc := p.processedImage.Clone()
	defer proc.Close()

	if orig.Rows() != proc.Rows() || orig.Cols() != proc.Cols() {
		return 0.0
	}

	if orig.Channels() > 1 {
		origGray := gocv.NewMat()
		defer origGray.Close()
		err := gocv.CvtColor(orig, &origGray, gocv.ColorBGRToGray)
		if err != nil {
			return 0.0
		}
		orig.Close()
		orig = origGray.Clone()
	}

	if proc.Channels() > 1 {
		procGray := gocv.NewMat()
		defer procGray.Close()
		err := gocv.CvtColor(proc, &procGray, gocv.ColorBGRToGray)
		if err != nil {
			return 0.0
		}
		proc.Close()
		proc = procGray.Clone()
	}

	origF := gocv.NewMat()
	defer origF.Close()
	procF := gocv.NewMat()
	defer procF.Close()

	orig.ConvertTo(&origF, gocv.MatTypeCV64F)
	proc.ConvertTo(&procF, gocv.MatTypeCV64F)

	origF.DivideFloat(255.0)
	procF.DivideFloat(255.0)

	c1 := 0.01 * 0.01
	c2 := 0.03 * 0.03

	kernel := gocv.GetGaussianKernel(11, 1.5)
	defer kernel.Close()

	mu1 := gocv.NewMat()
	defer mu1.Close()
	mu2 := gocv.NewMat()
	defer mu2.Close()

	err := gocv.Filter2D(origF, &mu1, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}
	err = gocv.Filter2D(procF, &mu2, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}

	mu1Mu2 := gocv.NewMat()
	defer mu1Mu2.Close()
	mu1Sq := gocv.NewMat()
	defer mu1Sq.Close()
	mu2Sq := gocv.NewMat()
	defer mu2Sq.Close()

	err = gocv.Multiply(mu1, mu2, &mu1Mu2)
	if err != nil {
		return 0.0
	}
	err = gocv.Multiply(mu1, mu1, &mu1Sq)
	if err != nil {
		return 0.0
	}
	err = gocv.Multiply(mu2, mu2, &mu2Sq)
	if err != nil {
		return 0.0
	}

	origF2 := gocv.NewMat()
	defer origF2.Close()
	procF2 := gocv.NewMat()
	defer procF2.Close()
	origFProcF := gocv.NewMat()
	defer origFProcF.Close()

	err = gocv.Multiply(origF, origF, &origF2)
	if err != nil {
		return 0.0
	}
	err = gocv.Multiply(procF, procF, &procF2)
	if err != nil {
		return 0.0
	}
	err = gocv.Multiply(origF, procF, &origFProcF)
	if err != nil {
		return 0.0
	}

	sigma1Sq := gocv.NewMat()
	defer sigma1Sq.Close()
	sigma2Sq := gocv.NewMat()
	defer sigma2Sq.Close()
	sigma12 := gocv.NewMat()
	defer sigma12.Close()

	temp1 := gocv.NewMat()
	defer temp1.Close()
	temp2 := gocv.NewMat()
	defer temp2.Close()
	temp3 := gocv.NewMat()
	defer temp3.Close()

	err = gocv.Filter2D(origF2, &temp1, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}
	err = gocv.Subtract(temp1, mu1Sq, &sigma1Sq)
	if err != nil {
		return 0.0
	}

	err = gocv.Filter2D(procF2, &temp2, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}
	err = gocv.Subtract(temp2, mu2Sq, &sigma2Sq)
	if err != nil {
		return 0.0
	}

	err = gocv.Filter2D(origFProcF, &temp3, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}
	err = gocv.Subtract(temp3, mu1Mu2, &sigma12)
	if err != nil {
		return 0.0
	}

	numerator1 := gocv.NewMat()
	defer numerator1.Close()
	numerator2 := gocv.NewMat()
	defer numerator2.Close()
	denominator1 := gocv.NewMat()
	defer denominator1.Close()
	denominator2 := gocv.NewMat()
	defer denominator2.Close()

	mu1Mu2Times2 := gocv.NewMat()
	defer mu1Mu2Times2.Close()
	sigma12Times2 := gocv.NewMat()
	defer sigma12Times2.Close()

	mu1Mu2.MultiplyFloat(2.0)
	mu1Mu2.CopyTo(&mu1Mu2Times2)
	numerator1.SetTo(gocv.NewScalar(c1, 0, 0, 0))
	err = gocv.Add(mu1Mu2Times2, numerator1, &numerator1)
	if err != nil {
		return 0.0
	}

	sigma12.MultiplyFloat(2.0)
	sigma12.CopyTo(&sigma12Times2)
	numerator2.SetTo(gocv.NewScalar(c2, 0, 0, 0))
	err = gocv.Add(sigma12Times2, numerator2, &numerator2)
	if err != nil {
		return 0.0
	}

	numerator := gocv.NewMat()
	defer numerator.Close()
	err = gocv.Multiply(numerator1, numerator2, &numerator)
	if err != nil {
		return 0.0
	}

	err = gocv.Add(mu1Sq, mu2Sq, &denominator1)
	if err != nil {
		return 0.0
	}
	denominator1.AddFloat(float32(c1))

	err = gocv.Add(sigma1Sq, sigma2Sq, &denominator2)
	if err != nil {
		return 0.0
	}
	denominator2.AddFloat(float32(c2))

	denominator := gocv.NewMat()
	defer denominator.Close()
	err = gocv.Multiply(denominator1, denominator2, &denominator)
	if err != nil {
		return 0.0
	}

	ssimMap := gocv.NewMat()
	defer ssimMap.Close()
	err = gocv.Divide(numerator, denominator, &ssimMap)
	if err != nil {
		return 0.0
	}

	meanSSIM := ssimMap.Mean()
	ssim := meanSSIM.Val1

	if math.IsInf(ssim, 0) || math.IsNaN(ssim) {
		return 0.0
	}

	if ssim > 1.0 {
		ssim = 1.0
	}
	if ssim < 0.0 {
		ssim = 0.0
	}

	return ssim
}
