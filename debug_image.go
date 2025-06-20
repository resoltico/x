package main

import (
	"log"
	"runtime"

	"gocv.io/x/gocv"
)

type DebugImage struct {
	enabled bool
}

func NewDebugImage(config *DebugConfig) *DebugImage {
	enabled := false
	if config != nil {
		enabled = config.Image
	}
	return &DebugImage{
		enabled: enabled,
	}
}

func (d *DebugImage) Log(message string) {
	if !d.enabled {
		return
	}
	log.Println("[IMAGE DEBUG]", message)
}

func (d *DebugImage) LogError(err error) {
	if !d.enabled || err == nil {
		return
	}
	log.Println("[IMAGE ERROR]", err)
}

func (d *DebugImage) LogMatInfo(name string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() {
		return
	}

	size := mat.Size()
	log.Printf("[IMAGE DEBUG] Mat '%s': %dx%d, channels=%d, type=%d, elemSize=%d",
		name, size[1], size[0], mat.Channels(), int(mat.Type()), mat.ElemSize())
}

func (d *DebugImage) LogMatPixelSamples(name string, mat gocv.Mat, numSamples int) {
	if !d.enabled || mat.Empty() {
		return
	}

	data := mat.ToBytes()
	if len(data) == 0 {
		d.Log("WARNING: Mat '" + name + "' has no data")
		return
	}

	// Sample pixels from different regions
	size := mat.Size()
	width, height := size[1], size[0]
	channels := mat.Channels()

	sampleIndices := []int{
		0,                      // Top-left
		(width / 2) * channels, // Top-center
		(width - 1) * channels, // Top-right
		(height/2)*width*channels + (width/2)*channels, // Center
		len(data) - channels,                           // Bottom-right
	}

	log.Printf("[IMAGE DEBUG] Mat '%s' pixel samples:", name)
	for i, idx := range sampleIndices {
		if idx < len(data) {
			if channels == 1 {
				log.Printf("[IMAGE DEBUG]   Sample %d: %d", i, data[idx])
			} else if channels == 3 && idx+2 < len(data) {
				log.Printf("[IMAGE DEBUG]   Sample %d: [%d,%d,%d]", i, data[idx], data[idx+1], data[idx+2])
			}
		}
	}
}

func (d *DebugImage) LogPixelDistribution(name string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() {
		return
	}

	data := mat.ToBytes()
	if len(data) == 0 {
		return
	}

	// Count pixel value distribution
	distribution := make(map[uint8]int)
	sampleSize := 10000
	if len(data) < sampleSize {
		sampleSize = len(data)
	}

	for i := 0; i < sampleSize; i++ {
		distribution[data[i]]++
	}

	log.Printf("[IMAGE DEBUG] Mat '%s' pixel distribution (first %d pixels):", name, sampleSize)
	for val, count := range distribution {
		percentage := float64(count) / float64(sampleSize) * 100
		log.Printf("[IMAGE DEBUG]   Value %d: %d pixels (%.1f%%)", val, count, percentage)
	}
}

func (d *DebugImage) LogPixelDistributionDetailed(name string, mat gocv.Mat, regions int) {
	if !d.enabled || mat.Empty() {
		return
	}

	data := mat.ToBytes()
	if len(data) == 0 {
		return
	}

	size := mat.Size()
	width, height := size[1], size[0]
	channels := mat.Channels()

	log.Printf("[IMAGE DEBUG] DETAILED pixel analysis for '%s':", name)
	log.Printf("[IMAGE DEBUG]   Dimensions: %dx%d, channels=%d, total pixels=%d", width, height, channels, len(data))

	// Sample from multiple regions to detect spatial distribution issues
	regionSize := len(data) / regions
	for r := 0; r < regions; r++ {
		start := r * regionSize
		end := start + min(regionSize, 1000) // Sample max 1000 pixels per region
		if start >= len(data) {
			break
		}
		if end > len(data) {
			end = len(data)
		}

		distribution := make(map[uint8]int)
		for i := start; i < end; i++ {
			distribution[data[i]]++
		}

		log.Printf("[IMAGE DEBUG]   Region %d (bytes %d-%d):", r, start, end-1)
		for val, count := range distribution {
			percentage := float64(count) / float64(end-start) * 100
			log.Printf("[IMAGE DEBUG]     Value %d: %d pixels (%.1f%%)", val, count, percentage)
		}
	}

	// Check specific coordinates for spatial verification
	d.LogPixelAtCoordinates(name, mat, [][]int{
		{0, 0},                          // Top-left corner
		{width / 2, height / 2},         // Center
		{width - 1, height - 1},         // Bottom-right corner
		{width / 4, height / 4},         // Quarter point
		{3 * width / 4, 3 * height / 4}, // Three-quarter point
	})
}

func (d *DebugImage) LogPixelAtCoordinates(name string, mat gocv.Mat, coords [][]int) {
	if !d.enabled || mat.Empty() {
		return
	}

	size := mat.Size()
	width, height := size[1], size[0]
	channels := mat.Channels()

	log.Printf("[IMAGE DEBUG] Pixel values at specific coordinates for '%s':", name)
	for _, coord := range coords {
		x, y := coord[0], coord[1]
		if x >= 0 && x < width && y >= 0 && y < height {
			// Calculate byte offset
			offset := y*width*channels + x*channels
			data := mat.ToBytes()
			if offset < len(data) {
				if channels == 1 {
					log.Printf("[IMAGE DEBUG]   At (%d,%d): %d", x, y, data[offset])
				} else if channels == 3 && offset+2 < len(data) {
					log.Printf("[IMAGE DEBUG]   At (%d,%d): [%d,%d,%d]", x, y, data[offset], data[offset+1], data[offset+2])
				}
			}
		}
	}
}

func (d *DebugImage) LogColorConversion(from, to string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Color conversion: %s -> %s", from, to)
}

func (d *DebugImage) LogFilter(filterName string, params ...interface{}) {
	if !d.enabled {
		return
	}
	if len(params) > 0 {
		log.Printf("[IMAGE DEBUG] Applied filter: %s with params: %+v", filterName, params)
	} else {
		log.Printf("[IMAGE DEBUG] Applied filter: %s", filterName)
	}
}

func (d *DebugImage) LogThreshold(method string, threshold1, threshold2 float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Threshold method: %s, thresholds: %.2f, %.2f", method, threshold1, threshold2)
}

func (d *DebugImage) LogMorphology(operation string, kernelSize int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Morphological operation: %s, kernel size: %dx%d", operation, kernelSize, kernelSize)
}

func (d *DebugImage) LogHistogram(name string, bins int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Histogram calculated for '%s': %d bins", name, bins)
}

func (d *DebugImage) LogOptimalThresholds(s, t int, variance float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Optimal thresholds found: s=%d, t=%d, variance=%.6f", s, t, variance)
}

func (d *DebugImage) LogPixelValues(name string, x, y int, values ...interface{}) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Pixel values at (%d,%d) in '%s': %+v", x, y, name, values)
}

func (d *DebugImage) LogImageLoad(filename string, success bool) {
	if !d.enabled {
		return
	}
	if success {
		log.Printf("[IMAGE DEBUG] Successfully loaded image: %s", filename)
	} else {
		log.Printf("[IMAGE DEBUG] Failed to load image: %s", filename)
	}
}

func (d *DebugImage) LogImageSave(filename string, success bool) {
	if !d.enabled {
		return
	}
	if success {
		log.Printf("[IMAGE DEBUG] Successfully saved image: %s", filename)
	} else {
		log.Printf("[IMAGE DEBUG] Failed to save image: %s", filename)
	}
}

func (d *DebugImage) LogQualityMetrics(psnr, ssim float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Quality metrics - PSNR: %.2f dB, SSIM: %.4f", psnr, ssim)
}

func (d *DebugImage) LogMemoryUsage() {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)

	log.Printf("[IMAGE DEBUG] Memory - Alloc: %.2f MB, TotalAlloc: %.2f MB, Sys: %.2f MB, NumGC: %d",
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		m.NumGC)
}

func (d *DebugImage) LogImageProperties(name string, width, height, channels int, dataType string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Image '%s' properties: %dx%d, %d channels, type: %s",
		name, width, height, channels, dataType)
}

func (d *DebugImage) LogAlgorithmStep(algorithm, step string, details ...interface{}) {
	if !d.enabled {
		return
	}
	if len(details) > 0 {
		log.Printf("[IMAGE DEBUG] %s - %s: %+v", algorithm, step, details)
	} else {
		log.Printf("[IMAGE DEBUG] %s - %s", algorithm, step)
	}
}

func (d *DebugImage) LogProcessingTime(operation string, milliseconds float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Operation '%s' completed in %.2f ms", operation, milliseconds)
}

func (d *DebugImage) LogThresholdAnalysis(name string, mat gocv.Mat, s, t int) {
	if !d.enabled || mat.Empty() {
		return
	}

	log.Printf("[IMAGE DEBUG] Threshold analysis for '%s' with s=%d, t=%d:", name, s, t)

	data := mat.ToBytes()
	if len(data) == 0 {
		log.Printf("[IMAGE DEBUG] ERROR: No data in Mat")
		return
	}

	// Count pixels in different threshold regions
	counts := map[string]int{
		"below_s": 0,
		"s_to_t":  0,
		"above_t": 0,
		"at_s":    0,
		"at_t":    0,
	}

	for _, val := range data[:min(10000, len(data))] {
		v := int(val)
		if v < s {
			counts["below_s"]++
		} else if v > t {
			counts["above_t"]++
		} else {
			counts["s_to_t"]++
		}
		if v == s {
			counts["at_s"]++
		}
		if v == t {
			counts["at_t"]++
		}
	}

	total := min(10000, len(data))
	log.Printf("[IMAGE DEBUG]   Below s(%d): %d (%.1f%%)", s, counts["below_s"], float64(counts["below_s"])/float64(total)*100)
	log.Printf("[IMAGE DEBUG]   Between s-t: %d (%.1f%%)", counts["s_to_t"], float64(counts["s_to_t"])/float64(total)*100)
	log.Printf("[IMAGE DEBUG]   Above t(%d): %d (%.1f%%)", t, counts["above_t"], float64(counts["above_t"])/float64(total)*100)
	log.Printf("[IMAGE DEBUG]   Exactly s: %d, Exactly t: %d", counts["at_s"], counts["at_t"])
}

func (d *DebugImage) LogBinarizationResult(inputName, outputName string, inputMat, outputMat gocv.Mat, s, t int) {
	if !d.enabled {
		return
	}

	log.Printf("[IMAGE DEBUG] Binarization result: %s -> %s", inputName, outputName)

	if !inputMat.Empty() {
		d.LogPixelDistribution(inputName+"_input", inputMat)
	}

	if !outputMat.Empty() {
		// Detailed analysis with regional sampling
		d.LogPixelDistributionDetailed(outputName+"_output", outputMat, 5)

		// Check if result is all black or all white
		data := outputMat.ToBytes()
		if len(data) > 0 {
			allBlack := true
			allWhite := true
			mixedCount := 0
			blackCount := 0
			whiteCount := 0

			// Sample more thoroughly
			sampleSize := min(10000, len(data))
			for i := 0; i < sampleSize; i++ {
				val := data[i]
				if val == 0 {
					blackCount++
				} else if val == 255 {
					whiteCount++
				} else {
					mixedCount++
					allBlack = false
					allWhite = false
				}
				if val != 0 {
					allBlack = false
				}
				if val != 255 {
					allWhite = false
				}
			}

			log.Printf("[IMAGE DEBUG] Sample analysis (%d pixels): Black=%d (%.1f%%), White=%d (%.1f%%), Mixed=%d (%.1f%%)",
				sampleSize, blackCount, float64(blackCount)/float64(sampleSize)*100,
				whiteCount, float64(whiteCount)/float64(sampleSize)*100,
				mixedCount, float64(mixedCount)/float64(sampleSize)*100)

			if allBlack {
				log.Printf("[IMAGE DEBUG] WARNING: Output image is ALL BLACK")
				d.LogThresholdAnalysis(inputName, inputMat, s, t)
			} else if allWhite {
				log.Printf("[IMAGE DEBUG] WARNING: Output image is ALL WHITE")
				d.LogThresholdAnalysis(inputName, inputMat, s, t)
			} else {
				log.Printf("[IMAGE DEBUG] Output has mixed values (good)")
			}
		}
	}
}

func (d *DebugImage) LogHistogramAnalysis(name string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() {
		return
	}

	data := mat.ToBytes()
	if len(data) == 0 {
		return
	}

	// Build histogram
	histogram := make([]int, 256)
	for _, val := range data {
		histogram[val]++
	}

	// Find min, max, and dominant values
	var minVal, maxVal uint8 = 255, 0
	maxCount := 0
	dominantVal := uint8(0)

	for i, count := range histogram {
		if count > 0 {
			if uint8(i) < minVal {
				minVal = uint8(i)
			}
			if uint8(i) > maxVal {
				maxVal = uint8(i)
			}
			if count > maxCount {
				maxCount = count
				dominantVal = uint8(i)
			}
		}
	}

	log.Printf("[IMAGE DEBUG] Histogram analysis for '%s':", name)
	log.Printf("[IMAGE DEBUG]   Min value: %d, Max value: %d, Range: %d", minVal, maxVal, maxVal-minVal)
	log.Printf("[IMAGE DEBUG]   Dominant value: %d (%d pixels, %.1f%%)", dominantVal, maxCount, float64(maxCount)/float64(len(data))*100)

	// Show distribution in ranges
	ranges := []struct{ start, end int }{
		{0, 63}, {64, 127}, {128, 191}, {192, 255},
	}
	for _, r := range ranges {
		count := 0
		for i := r.start; i <= r.end; i++ {
			count += histogram[i]
		}
		if count > 0 {
			log.Printf("[IMAGE DEBUG]   Range %d-%d: %d pixels (%.1f%%)", r.start, r.end, count, float64(count)/float64(len(data))*100)
		}
	}
}

func (d *DebugImage) LogMatDataValidation(name string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() {
		return
	}

	log.Printf("[IMAGE DEBUG] VALIDATION for Mat '%s':", name)

	// Basic properties
	size := mat.Size()
	width, height := size[1], size[0]
	channels := mat.Channels()
	dataType := mat.Type()

	log.Printf("[IMAGE DEBUG]   Properties: %dx%d, channels=%d, type=%d", width, height, channels, int(dataType))
	log.Printf("[IMAGE DEBUG]   Expected data size: %d bytes", width*height*channels)

	// Get raw data
	data := mat.ToBytes()
	actualSize := len(data)
	log.Printf("[IMAGE DEBUG]   Actual data size: %d bytes", actualSize)

	if actualSize == 0 {
		log.Printf("[IMAGE DEBUG]   ERROR: Mat contains no data!")
		return
	}

	// Check data integrity
	expectedSize := width * height * channels
	if actualSize != expectedSize {
		log.Printf("[IMAGE DEBUG]   WARNING: Size mismatch! Expected %d, got %d", expectedSize, actualSize)
	}

	// Sample data verification
	log.Printf("[IMAGE DEBUG]   First 10 bytes: %v", data[:min(10, len(data))])
	log.Printf("[IMAGE DEBUG]   Last 10 bytes: %v", data[max(0, len(data)-10):])

	// Check for pattern consistency
	uniqueValues := make(map[uint8]bool)
	for i := 0; i < min(1000, len(data)); i++ {
		uniqueValues[data[i]] = true
	}
	log.Printf("[IMAGE DEBUG]   Unique values in first 1000 bytes: %d", len(uniqueValues))
}

func (d *DebugImage) IsEnabled() bool {
	return d.enabled
}

func (d *DebugImage) Enable() {
	d.enabled = true
	log.Println("[IMAGE DEBUG] Image processing debugging enabled - output to terminal only")
}

func (d *DebugImage) Disable() {
	d.enabled = false
}
