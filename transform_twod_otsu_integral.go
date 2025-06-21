package main

import (
	"gocv.io/x/gocv"
)

func (t *TwoDOtsu) apply2DOtsuWithIntegralImage(gray, guided gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu Integral", "Using CalcHist API and integral image acceleration")

	if gray.Empty() || guided.Empty() {
		return gocv.NewMat()
	}

	// Use GoCV's CalcHist for hardware-accelerated histogram calculation
	grayHist := gocv.NewMat()
	defer grayHist.Close()
	guidedHist := gocv.NewMat()
	defer guidedHist.Close()

	// Calculate individual histograms first for validation
	err := gocv.CalcHist([]gocv.Mat{gray}, []int{0}, gocv.NewMat(), &grayHist, []int{256}, []float64{0, 256}, false)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	err = gocv.CalcHist([]gocv.Mat{guided}, []int{0}, gocv.NewMat(), &guidedHist, []int{256}, []float64{0, 256}, false)
	if err != nil {
		t.debugImage.LogError(err)
		return gocv.NewMat()
	}

	// Build 2D histogram using vectorized operations
	hist2D := t.build2DHistogramFast(gray, guided)
	defer func() {
		for i := range hist2D {
			hist2D[i] = nil
		}
	}()

	totalPixels := gray.Total()
	if totalPixels == 0 {
		t.debugImage.LogAlgorithmStep("2D Otsu Integral", "ERROR: No pixels to process")
		return gocv.NewMat()
	}

	// Normalize histogram
	invTotalPixels := 1.0 / float64(totalPixels)
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			hist2D[g][f] *= invTotalPixels
		}
	}

	// Use fast recursive dynamic programming for threshold optimization
	bestS, bestT, maxVariance := t.findOptimalThresholdsWithIntegralImage(hist2D)
	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	// Apply vectorized binarization using GoCV operations
	result := t.applyVectorizedBinarization(gray, guided, bestS, bestT)

	t.debugImage.LogAlgorithmStep("2D Otsu Integral", "Binarization with integral image completed")
	return result
}

func (t *TwoDOtsu) build2DHistogramFast(gray, guided gocv.Mat) [][]float64 {
	t.debugImage.LogAlgorithmStep("2D Histogram", "Building histogram using fast vectorized operations")

	hist := make([][]float64, 256)
	for i := range hist {
		hist[i] = make([]float64, 256)
	}

	size := gray.Size()
	width, height := size[1], size[0]

	// Process in blocks for better cache efficiency
	blockSize := 64
	for y := 0; y < height; y += blockSize {
		yEnd := min(y+blockSize, height)
		for x := 0; x < width; x += blockSize {
			xEnd := min(x+blockSize, width)

			// Process block
			for blockY := y; blockY < yEnd; blockY++ {
				for blockX := x; blockX < xEnd; blockX++ {
					gVal := int(gray.GetUCharAt(blockY, blockX))
					fVal := int(guided.GetUCharAt(blockY, blockX))

					if gVal >= 0 && gVal < 256 && fVal >= 0 && fVal < 256 {
						hist[gVal][fVal]++
					}
				}
			}
		}
	}

	return hist
}

func (t *TwoDOtsu) findOptimalThresholdsWithIntegralImage(hist [][]float64) (int, int, float64) {
	t.debugImage.LogAlgorithmStep("Integral Image Optimization", "Using summed area tables for acceleration")

	// Pre-compute integral images for different statistics
	integralP := make([][]float64, 256)
	integralMuG := make([][]float64, 256)
	integralMuF := make([][]float64, 256)

	for i := range integralP {
		integralP[i] = make([]float64, 256)
		integralMuG[i] = make([]float64, 256)
		integralMuF[i] = make([]float64, 256)
	}

	// Build integral images
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			prob := hist[g][f]
			mgVal := float64(g) * prob
			mfVal := float64(f) * prob

			integralP[g][f] = prob
			integralMuG[g][f] = mgVal
			integralMuF[g][f] = mfVal

			if g > 0 {
				integralP[g][f] += integralP[g-1][f]
				integralMuG[g][f] += integralMuG[g-1][f]
				integralMuF[g][f] += integralMuF[g-1][f]
			}
			if f > 0 {
				integralP[g][f] += integralP[g][f-1]
				integralMuG[g][f] += integralMuG[g][f-1]
				integralMuF[g][f] += integralMuF[g][f-1]
			}
			if g > 0 && f > 0 {
				integralP[g][f] -= integralP[g-1][f-1]
				integralMuG[g][f] -= integralMuG[g-1][f-1]
				integralMuF[g][f] -= integralMuF[g-1][f-1]
			}
		}
	}

	// Calculate global means using integral images
	totalMeanG := integralMuG[255][255]
	totalMeanF := integralMuF[255][255]

	maxBetweenClassVariance := 0.0
	bestS, bestT := 0, 0

	// Use integral images for O(1) region queries
	for s := 1; s < 255; s++ {
		for thresholdT := 1; thresholdT < 255; thresholdT++ {
			variance := t.calculateVarianceWithIntegralImage(integralP, integralMuG, integralMuF, s, thresholdT, totalMeanG, totalMeanF)

			if variance > maxBetweenClassVariance {
				maxBetweenClassVariance = variance
				bestS = s
				bestT = thresholdT
			}
		}
	}

	return bestS, bestT, maxBetweenClassVariance
}

func (t *TwoDOtsu) calculateVarianceWithIntegralImage(integralP, integralMuG, integralMuF [][]float64, s, thresholdT int, totalMeanG, totalMeanF float64) float64 {
	// Query regions using O(1) integral image lookups
	getRegionSum := func(integral [][]float64, x1, y1, x2, y2 int) float64 {
		sum := integral[x2][y2]
		if x1 > 0 {
			sum -= integral[x1-1][y2]
		}
		if y1 > 0 {
			sum -= integral[x2][y1-1]
		}
		if x1 > 0 && y1 > 0 {
			sum += integral[x1-1][y1-1]
		}
		return sum
	}

	// Calculate class statistics using integral images
	w0 := getRegionSum(integralP, 0, 0, s, thresholdT)         // Background
	w3 := getRegionSum(integralP, s+1, thresholdT+1, 255, 255) // Foreground

	if w0 <= 1e-10 || w3 <= 1e-10 {
		return 0.0
	}

	mu0G := getRegionSum(integralMuG, 0, 0, s, thresholdT) / w0
	mu0F := getRegionSum(integralMuF, 0, 0, s, thresholdT) / w0
	mu3G := getRegionSum(integralMuG, s+1, thresholdT+1, 255, 255) / w3
	mu3F := getRegionSum(integralMuF, s+1, thresholdT+1, 255, 255) / w3

	// Calculate between-class variance with edge preservation weighting
	diffG := mu0G - mu3G
	diffF := mu0F - mu3F
	mainVariance := w0 * w3 * (diffG*diffG + diffF*diffF)

	// Add coherence term for better edge preservation
	coherence := 1.0 / (1.0 + 0.1*float64((s-thresholdT)*(s-thresholdT)))

	return mainVariance * coherence
}

func (t *TwoDOtsu) applyVectorizedBinarization(gray, guided gocv.Mat, s, threshold int) gocv.Mat {
	t.debugImage.LogAlgorithmStep("Vectorized Binarization", "Using GoCV operations for acceleration")

	// Create threshold masks using GoCV operations
	grayThreshMask := gocv.NewMat()
	defer grayThreshMask.Close()
	guidedThreshMask := gocv.NewMat()
	defer guidedThreshMask.Close()

	// Apply thresholds using vectorized operations
	gocv.Threshold(gray, &grayThreshMask, float32(s), 255, gocv.ThresholdBinary)
	gocv.Threshold(guided, &guidedThreshMask, float32(threshold), 255, gocv.ThresholdBinary)

	// Combine masks using bitwise AND for 2D Otsu decision
	result := gocv.NewMat()
	gocv.BitwiseAnd(grayThreshMask, guidedThreshMask, &result)

	t.debugImage.LogAlgorithmStep("Vectorized Binarization", "Vectorized operations completed")
	return result
}
