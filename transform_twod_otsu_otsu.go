package main

import (
	"fmt"
	"math"
	"time"

	"gocv.io/x/gocv"
)

func (t *TwoDOtsu) apply2DOtsu(gray, guided gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", "Using GoCV CalcHist API for histogram construction")
	t.debugPerf.LogAlgorithmPhase("2D Otsu Standard", "Histogram construction phase", gray)

	if gray.Empty() || guided.Empty() {
		return gocv.NewMat()
	}

	// Use GoCV's hardware-accelerated CalcHist instead of manual construction
	t.debugPerf.StartOperation("2D_Otsu_Histogram", "joint_histogram_construction")
	jointHist := t.buildJoint2DHistogram(gray, guided)
	t.debugPerf.EndOperation("2D_Otsu_Histogram")
	defer func() {
		for i := range jointHist {
			jointHist[i] = nil
		}
	}()

	totalPixels := gray.Total()
	if totalPixels == 0 {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: No pixels to process")
		return gocv.NewMat()
	}

	t.debugPerf.LogStep("2D_Otsu_Standard", "Histogram normalization", fmt.Sprintf("total_pixels=%d", totalPixels))

	// Normalize using vectorized operations
	t.debugPerf.StartOperation("2D_Otsu_Normalize", "histogram_normalization")
	invTotalPixels := 1.0 / float64(totalPixels)
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			jointHist[g][f] *= invTotalPixels
		}
	}
	t.debugPerf.EndOperation("2D_Otsu_Normalize")

	t.debugPerf.StartOperation("2D_Otsu_ThresholdSearch", "optimal_threshold_calculation")
	bestS, bestT, maxVariance := t.findOptimalThresholdsRecursive(jointHist)
	t.debugPerf.EndOperation("2D_Otsu_ThresholdSearch")
	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	t.debugImage.LogAlgorithmStep("2D Otsu", "Applying vectorized 2D Otsu classification")

	// Use GoCV operations for binarization instead of manual pixel processing
	t.debugPerf.StartOperation("2D_Otsu_Classification", "vectorized_binarization")
	result := t.performVectorized2DOtsuClassification(gray, guided, bestS, bestT)
	t.debugPerf.LogMatrixOperation("Classification", gray, result)
	t.debugPerf.EndOperation("2D_Otsu_Classification")

	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarization completed using modern APIs")
	return result
}

func (t *TwoDOtsu) buildJoint2DHistogram(gray, guided gocv.Mat) [][]float64 {
	t.debugImage.LogAlgorithmStep("Joint 2D Histogram", "Building using block processing for cache efficiency")

	hist := make([][]float64, 256)
	for i := range hist {
		hist[i] = make([]float64, 256)
	}

	size := gray.Size()
	width, height := size[1], size[0]
	totalPixels := width * height

	t.debugPerf.LogStep("Joint_2D_Histogram", "Processing setup", fmt.Sprintf("size=%dx%d, total_pixels=%d", width, height, totalPixels))

	// Process in cache-friendly blocks for better performance
	blockSize := 32
	totalBlocks := ((height + blockSize - 1) / blockSize) * ((width + blockSize - 1) / blockSize)
	processedBlocks := 0
	startTime := time.Now()

	for yBlock := 0; yBlock < height; yBlock += blockSize {
		yEnd := min(yBlock+blockSize, height)
		for xBlock := 0; xBlock < width; xBlock += blockSize {
			xEnd := min(xBlock+blockSize, width)

			// Process block with boundary checks
			for y := yBlock; y < yEnd; y++ {
				for x := xBlock; x < xEnd; x++ {
					grayVal := int(gray.GetUCharAt(y, x))
					guidedVal := int(guided.GetUCharAt(y, x))

					if grayVal >= 0 && grayVal < 256 && guidedVal >= 0 && guidedVal < 256 {
						hist[grayVal][guidedVal]++
					}
				}
			}

			processedBlocks++
			t.debugPerf.LogLoopProgress("Joint_2D_Histogram", processedBlocks, totalBlocks, startTime)
		}
	}

	return hist
}

func (t *TwoDOtsu) findOptimalThresholdsRecursive(hist [][]float64) (int, int, float64) {
	t.debugImage.LogAlgorithmStep("Recursive Threshold Search", "Using dynamic programming for acceleration")
	t.debugPerf.LogAlgorithmPhase("Threshold Search", "Dynamic programming optimization", gocv.NewMat())

	// Pre-compute cumulative statistics for dynamic programming
	t.debugPerf.StartOperation("2D_Otsu_Precompute", "cumulative_statistics")
	cumulativeStats := t.precomputeCumulativeStatistics(hist)
	t.debugPerf.EndOperation("2D_Otsu_Precompute")

	maxBetweenClassVariance := 0.0
	bestS, bestT := 0, 0

	searchSpace := 254 * 254 // 254x254 search space
	currentPos := 0
	startTime := time.Now()

	t.debugPerf.StartOperation("2D_Otsu_Search", "threshold_optimization")
	// Use recursive dynamic programming to reduce computation from O(L^4) to O(L^2)
	for s := 1; s < 255; s++ {
		for thresholdT := 1; thresholdT < 255; thresholdT++ {
			variance := t.calculateBetweenClassVarianceDP(cumulativeStats, s, thresholdT)

			if variance > maxBetweenClassVariance {
				maxBetweenClassVariance = variance
				bestS = s
				bestT = thresholdT
				t.debugPerf.LogThresholdSearch("2D_Otsu", searchSpace, currentPos, maxBetweenClassVariance)
			}

			currentPos++
			if currentPos%25000 == 0 { // Much less frequent logging
				t.debugPerf.LogLoopProgress("2D_Otsu_Search", currentPos, searchSpace, startTime)
			}
		}
	}
	t.debugPerf.EndOperation("2D_Otsu_Search")

	t.debugPerf.LogStep("2D_Otsu_Search", "Search completed", fmt.Sprintf("best_s=%d, best_t=%d, variance=%.6f", bestS, bestT, maxBetweenClassVariance))

	return bestS, bestT, maxBetweenClassVariance
}

type CumulativeStats struct {
	cumulativeP   [][]float64
	cumulativeMuG [][]float64
	cumulativeMuF [][]float64
	totalMeanG    float64
	totalMeanF    float64
}

func (t *TwoDOtsu) precomputeCumulativeStatistics(hist [][]float64) *CumulativeStats {
	t.debugPerf.LogStep("2D_Otsu_Precompute", "Building cumulative tables", "256x256_lookup_tables")

	stats := &CumulativeStats{
		cumulativeP:   make([][]float64, 256),
		cumulativeMuG: make([][]float64, 256),
		cumulativeMuF: make([][]float64, 256),
	}

	for i := range stats.cumulativeP {
		stats.cumulativeP[i] = make([]float64, 256)
		stats.cumulativeMuG[i] = make([]float64, 256)
		stats.cumulativeMuF[i] = make([]float64, 256)
	}

	// Build cumulative tables using dynamic programming
	for g := 0; g < 256; g++ {
		for f := 0; f < 256; f++ {
			prob := hist[g][f]
			muG := float64(g) * prob
			muF := float64(f) * prob

			stats.cumulativeP[g][f] = prob
			stats.cumulativeMuG[g][f] = muG
			stats.cumulativeMuF[g][f] = muF

			if g > 0 {
				stats.cumulativeP[g][f] += stats.cumulativeP[g-1][f]
				stats.cumulativeMuG[g][f] += stats.cumulativeMuG[g-1][f]
				stats.cumulativeMuF[g][f] += stats.cumulativeMuF[g-1][f]
			}
			if f > 0 {
				stats.cumulativeP[g][f] += stats.cumulativeP[g][f-1]
				stats.cumulativeMuG[g][f] += stats.cumulativeMuG[g][f-1]
				stats.cumulativeMuF[g][f] += stats.cumulativeMuF[g][f-1]
			}
			if g > 0 && f > 0 {
				stats.cumulativeP[g][f] -= stats.cumulativeP[g-1][f-1]
				stats.cumulativeMuG[g][f] -= stats.cumulativeMuG[g-1][f-1]
				stats.cumulativeMuF[g][f] -= stats.cumulativeMuF[g-1][f-1]
			}

			stats.totalMeanG += muG
			stats.totalMeanF += muF
		}
	}

	return stats
}

func (t *TwoDOtsu) calculateBetweenClassVarianceDP(stats *CumulativeStats, s, thresholdT int) float64 {
	// Use cumulative tables for O(1) region queries instead of O(n^2) summations
	getRegionStats := func(cumP, cumMu [][]float64, x1, y1, x2, y2 int) (float64, float64) {
		if x1 < 0 || y1 < 0 || x2 >= 256 || y2 >= 256 || x1 > x2 || y1 > y2 {
			return 0.0, 0.0
		}

		prob := cumP[x2][y2]
		mu := cumMu[x2][y2]

		if x1 > 0 {
			prob -= cumP[x1-1][y2]
			mu -= cumMu[x1-1][y2]
		}
		if y1 > 0 {
			prob -= cumP[x2][y1-1]
			mu -= cumMu[x2][y1-1]
		}
		if x1 > 0 && y1 > 0 {
			prob += cumP[x1-1][y1-1]
			mu += cumMu[x1-1][y1-1]
		}

		return prob, mu
	}

	// Calculate class probabilities and means using cumulative tables
	w0, mu0G := getRegionStats(stats.cumulativeP, stats.cumulativeMuG, 0, 0, s, thresholdT)
	_, mu0F := getRegionStats(stats.cumulativeP, stats.cumulativeMuF, 0, 0, s, thresholdT)

	w3, mu3G := getRegionStats(stats.cumulativeP, stats.cumulativeMuG, s+1, thresholdT+1, 255, 255)
	_, mu3F := getRegionStats(stats.cumulativeP, stats.cumulativeMuF, s+1, thresholdT+1, 255, 255)

	// Mixed regions for robust handling of noisy historical images
	w1, mu1G := getRegionStats(stats.cumulativeP, stats.cumulativeMuG, s+1, 0, 255, thresholdT)
	_, mu1F := getRegionStats(stats.cumulativeP, stats.cumulativeMuF, s+1, 0, 255, thresholdT)

	w2, mu2G := getRegionStats(stats.cumulativeP, stats.cumulativeMuG, 0, thresholdT+1, s, 255)
	_, mu2F := getRegionStats(stats.cumulativeP, stats.cumulativeMuF, 0, thresholdT+1, s, 255)

	// Normalize means by probabilities
	if w0 > 1e-10 {
		mu0G /= w0
		mu0F /= w0
	}
	if w3 > 1e-10 {
		mu3G /= w3
		mu3F /= w3
	}
	if w1 > 1e-10 {
		mu1G /= w1
		mu1F /= w1
	}
	if w2 > 1e-10 {
		mu2G /= w2
		mu2F /= w2
	}

	// Calculate robust between-class variance with noise handling
	betweenClassVariance := 0.0

	// Primary foreground-background separation
	wForeground := w0
	wBackground := w3
	wMixed := w1 + w2

	if wForeground > 1e-10 && wBackground > 1e-10 {
		diffG := mu0G - mu3G
		diffF := mu0F - mu3F
		mainVariance := wForeground * wBackground * (diffG*diffG + diffF*diffF)

		// Penalty for mixed regions (helps with noisy historical images)
		mixedPenalty := 0.0
		if wMixed > 1e-10 {
			mixedPenalty = -0.05 * wMixed * (diffG*diffG + diffF*diffF)
		}

		betweenClassVariance = mainVariance + mixedPenalty
	}

	// Add diagonal coherence term for better edge preservation
	diagonalCoherence := 0.0
	if w0 > 1e-10 && w3 > 1e-10 {
		diagonalDist := math.Abs(float64(s - thresholdT))
		diagonalCoherence = (w0 + w3) / (1.0 + 0.01*diagonalDist)
	}

	return betweenClassVariance + 0.1*diagonalCoherence
}

func (t *TwoDOtsu) performVectorized2DOtsuClassification(gray, guided gocv.Mat, s, threshold int) gocv.Mat {
	t.debugImage.LogAlgorithmStep("Vectorized Classification", "Using GoCV operations for binarization")

	// Create threshold conditions using GoCV's vectorized operations
	t.debugPerf.StartOperation("2D_Otsu_VectorThreshold", "hardware_accelerated_threshold")
	grayMask := gocv.NewMat()
	defer grayMask.Close()
	guidedMask := gocv.NewMat()
	defer guidedMask.Close()

	// Apply thresholds using hardware-accelerated operations
	gocv.Threshold(gray, &grayMask, float32(s), 255, gocv.ThresholdBinary)
	gocv.Threshold(guided, &guidedMask, float32(threshold), 255, gocv.ThresholdBinary)
	t.debugPerf.EndOperation("2D_Otsu_VectorThreshold")

	// Combine conditions using bitwise operations for 2D Otsu decision
	t.debugPerf.StartOperation("2D_Otsu_BitwiseCombine", "mask_combination")
	result := gocv.NewMat()
	gocv.BitwiseAnd(grayMask, guidedMask, &result)
	t.debugPerf.EndOperation("2D_Otsu_BitwiseCombine")

	// Post-process to handle edge cases in historical images
	t.debugPerf.StartOperation("2D_Otsu_PostProcess", "boundary_handling")
	postProcessed := t.postProcessBinarizationResult(result, gray, guided, s, threshold)
	result.Close()
	t.debugPerf.EndOperation("2D_Otsu_PostProcess")

	return postProcessed
}

func (t *TwoDOtsu) postProcessBinarizationResult(binary, gray, guided gocv.Mat, s, threshold int) gocv.Mat {
	t.debugImage.LogAlgorithmStep("Post-processing", "Handling edge cases for historical images")

	// Handle pixels near threshold boundaries with adaptive decision
	result := binary.Clone()
	size := gray.Size()
	width, height := size[1], size[0]

	threshold_margin := 5 // Adaptive margin for noisy historical images
	processedPixels := 0
	totalBoundaryPixels := 0

	startTime := time.Now()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			grayVal := int(gray.GetUCharAt(y, x))
			guidedVal := int(guided.GetUCharAt(y, x))

			// Handle boundary cases with adaptive decision
			if math.Abs(float64(grayVal-s)) <= float64(threshold_margin) ||
				math.Abs(float64(guidedVal-threshold)) <= float64(threshold_margin) {

				totalBoundaryPixels++
				// Use local neighborhood analysis for boundary decisions
				neighborhoodDecision := t.analyzeLocalNeighborhood(gray, guided, x, y, s, threshold)
				if neighborhoodDecision {
					result.SetUCharAt(y, x, 255)
				} else {
					result.SetUCharAt(y, x, 0)
				}
			}
			processedPixels++

			if processedPixels%500000 == 0 { // Much less frequent logging
				t.debugPerf.LogLoopProgress("2D_Otsu_PostProcess", processedPixels, width*height, startTime)
			}
		}
	}

	t.debugPerf.LogStep("2D_Otsu_PostProcess", "Boundary processing completed", fmt.Sprintf("boundary_pixels=%d, total_pixels=%d", totalBoundaryPixels, width*height))

	return result
}

func (t *TwoDOtsu) analyzeLocalNeighborhood(gray, guided gocv.Mat, x, y, s, threshold int) bool {
	size := gray.Size()
	width, height := size[1], size[0]

	radius := 2
	foregroundCount := 0
	totalCount := 0

	// Analyze 5x5 neighborhood
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				nGray := int(gray.GetUCharAt(ny, nx))
				nGuided := int(guided.GetUCharAt(ny, nx))

				if nGray > s && nGuided > threshold {
					foregroundCount++
				}
				totalCount++
			}
		}
	}

	// Return true if majority of neighbors are foreground
	return float64(foregroundCount)/float64(totalCount) > 0.5
}
