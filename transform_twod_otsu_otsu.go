package main

import (
	"math"

	"gocv.io/x/gocv"
)

func (t *TwoDOtsu) apply2DOtsu(gray, guided gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", "Constructing 2D histogram")

	if gray.Empty() || guided.Empty() {
		return gocv.NewMat()
	}

	grayData := gray.ToBytes()
	guidedData := guided.ToBytes()

	if len(grayData) == 0 || len(guidedData) == 0 || len(grayData) != len(guidedData) {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: Invalid data")
		return gocv.NewMat()
	}

	hist := make([][]float64, 256)
	for i := range hist {
		hist[i] = make([]float64, 256)
	}

	totalPixels := len(grayData)

	for i := 0; i < totalPixels; i++ {
		g := int(grayData[i])
		f := int(guidedData[i])

		if g >= 0 && g < 256 && f >= 0 && f < 256 {
			hist[g][f]++
		}
	}

	invTotalPixels := 1.0 / float64(totalPixels)
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			hist[g][f] *= invTotalPixels
		}
	}

	bestS, bestT, maxVariance := t.findOptimalThresholds(hist)
	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	t.debugImage.LogAlgorithmStep("2D Otsu", "Applying 2D Otsu classification")

	size := gray.Size()
	width, height := size[1], size[0]
	result := gocv.NewMatWithSize(height, width, gocv.MatTypeCV8U)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grayVal := int(gray.GetUCharAt(y, x))
			guidedVal := int(guided.GetUCharAt(y, x))

			var pixelValue uint8 = 0
			if grayVal > bestS && guidedVal > bestT {
				pixelValue = 255
			} else if grayVal <= bestS && guidedVal <= bestT {
				pixelValue = 0
			} else {
				distance := float64(grayVal-bestS) + float64(guidedVal-bestT)
				if distance > 0 {
					pixelValue = 255
				} else {
					pixelValue = 0
				}
			}

			result.SetUCharAt(y, x, pixelValue)
		}
	}

	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarization completed")
	return result
}

func (t *TwoDOtsu) findOptimalThresholds(hist [][]float64) (int, int, float64) {
	maxBetweenClassVariance := 0.0
	bestS, bestT := 0, 0

	totalMeanG := 0.0
	totalMeanF := 0.0
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			prob := hist[g][f]
			totalMeanG += float64(g) * prob
			totalMeanF += float64(f) * prob
		}
	}

	for s := 1; s < 255; s++ {
		for thresholdT := 1; thresholdT < 255; thresholdT++ {
			variance := t.calculateBetweenClassScatter(hist, s, thresholdT, totalMeanG, totalMeanF)
			if variance > maxBetweenClassVariance {
				maxBetweenClassVariance = variance
				bestS = s
				bestT = thresholdT
			}
		}
	}

	return bestS, bestT, maxBetweenClassVariance
}

func (t *TwoDOtsu) calculateBetweenClassScatter(hist [][]float64, s, thresholdT int, totalMeanG, totalMeanF float64) float64 {
	var w [4]float64
	var muG [4]float64
	var muF [4]float64

	for g := 0; g <= s; g++ {
		for f := 0; f <= thresholdT; f++ {
			prob := hist[g][f]
			w[0] += prob
			muG[0] += float64(g) * prob
			muF[0] += float64(f) * prob
		}
	}

	for g := s + 1; g < 256; g++ {
		for f := 0; f <= thresholdT; f++ {
			prob := hist[g][f]
			w[1] += prob
			muG[1] += float64(g) * prob
			muF[1] += float64(f) * prob
		}
	}

	for g := 0; g <= s; g++ {
		for f := thresholdT + 1; f < 256; f++ {
			prob := hist[g][f]
			w[2] += prob
			muG[2] += float64(g) * prob
			muF[2] += float64(f) * prob
		}
	}

	for g := s + 1; g < 256; g++ {
		for f := thresholdT + 1; f < 256; f++ {
			prob := hist[g][f]
			w[3] += prob
			muG[3] += float64(g) * prob
			muF[3] += float64(f) * prob
		}
	}

	for i := 0; i < 4; i++ {
		if w[i] > 1e-10 {
			muG[i] /= w[i]
			muF[i] /= w[i]
		}
	}

	betweenClassVariance := 0.0

	wForeground := w[0]
	wBackground := w[3]
	wMixed := w[1] + w[2]

	if wForeground > 1e-10 && wBackground > 1e-10 {
		diffG := muG[0] - muG[3]
		diffF := muF[0] - muF[3]
		mainVariance := wForeground * wBackground * (diffG*diffG + diffF*diffF)

		mixedPenalty := 0.0
		if wMixed > 1e-10 {
			mixedPenalty = -0.1 * wMixed * (diffG*diffG + diffF*diffF)
		}

		betweenClassVariance = mainVariance + mixedPenalty
	}

	diagonalCoherence := 0.0
	if w[0] > 1e-10 && w[3] > 1e-10 {
		diagonalDist := math.Abs(float64(s - thresholdT))
		diagonalCoherence = (w[0] + w[3]) / (1.0 + 0.01*diagonalDist)
	}

	return betweenClassVariance + 0.1*diagonalCoherence
}
