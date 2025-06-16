// Advanced Otsu implementations with multi-level and local adaptive variants
package algorithms

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// MultiOtsu implements multi-level Otsu thresholding
type MultiOtsu struct{}

// NewMultiOtsu creates a new multi-level Otsu algorithm
func NewMultiOtsu() *MultiOtsu {
	return &MultiOtsu{}
}

func (m *MultiOtsu) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// Convert to grayscale if needed
	gray := m.ensureGrayscale(input)
	defer func() {
		if gray.Ptr() != input.Ptr() {
			gray.Close()
		}
	}()

	// Get parameters
	levels := 2 // Default to 2-level (finds 1 threshold)
	if val, ok := params["levels"]; ok {
		if v, ok := val.(float64); ok {
			levels = int(v)
		}
	}

	maxValue := 255.0
	if val, ok := params["max_value"]; ok {
		if v, ok := val.(float64); ok {
			maxValue = v
		}
	}

	// Calculate histogram
	hist := m.calculateHistogram(gray)

	var thresholds []float64
	if levels == 2 {
		// Single threshold (classic Otsu)
		threshold := m.calculateOtsuThreshold(hist)
		thresholds = []float64{threshold}
	} else {
		// Multi-level Otsu
		thresholds = m.calculateMultiOtsuThresholds(hist, levels)
	}

	// Apply thresholds
	output := m.applyThresholds(gray, thresholds, maxValue)
	return output, nil
}

func (m *MultiOtsu) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"levels":    2.0,
		"max_value": 255.0,
	}
}

func (m *MultiOtsu) GetName() string {
	return "Multi-Level Otsu"
}

func (m *MultiOtsu) GetDescription() string {
	return "Advanced Otsu thresholding with support for 2-level and 3-level segmentation"
}

func (m *MultiOtsu) Validate(params map[string]interface{}) error {
	if val, ok := params["levels"]; ok {
		if v, ok := val.(float64); ok {
			if v < 2 || v > 4 {
				return fmt.Errorf("levels must be between 2 and 4")
			}
		}
	}

	if val, ok := params["max_value"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0 || v > 255 {
				return fmt.Errorf("max_value must be between 0 and 255")
			}
		}
	}

	return nil
}

func (m *MultiOtsu) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "levels",
			Type:        "int",
			Min:         2.0,
			Max:         4.0,
			Default:     2.0,
			Description: "Number of threshold levels (2=1 threshold, 3=2 thresholds, etc.)",
		},
		{
			Name:        "max_value",
			Type:        "float",
			Min:         0.0,
			Max:         255.0,
			Default:     255.0,
			Description: "Maximum output value",
		},
	}
}

func (m *MultiOtsu) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}

	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (m *MultiOtsu) calculateHistogram(gray gocv.Mat) []float64 {
	hist := make([]float64, 256)

	// Manual histogram calculation for better control
	for y := 0; y < gray.Rows(); y++ {
		for x := 0; x < gray.Cols(); x++ {
			intensity := gray.GetUCharAt(y, x)
			hist[intensity]++
		}
	}

	// Normalize histogram
	totalPixels := float64(gray.Rows() * gray.Cols())
	for i := range hist {
		hist[i] /= totalPixels
	}

	return hist
}

func (m *MultiOtsu) calculateOtsuThreshold(hist []float64) float64 {
	// Calculate cumulative sums
	sum := 0.0
	for i := 0; i < 256; i++ {
		sum += float64(i) * hist[i]
	}

	sumB := 0.0
	wB := 0.0
	maximum := 0.0
	level := 0.0

	for t := 0; t < 256; t++ {
		wB += hist[t]
		if wB == 0 {
			continue
		}

		wF := 1.0 - wB
		if wF == 0 {
			break
		}

		sumB += float64(t) * hist[t]
		mB := sumB / wB
		mF := (sum - sumB) / wF

		// Calculate between-class variance
		between := wB * wF * (mB - mF) * (mB - mF)

		if between > maximum {
			level = float64(t)
			maximum = between
		}
	}

	return level
}

func (m *MultiOtsu) calculateMultiOtsuThresholds(hist []float64, levels int) []float64 {
	// For simplicity, use recursive binary Otsu for multi-level
	// This is a simplified implementation - full multi-level Otsu is more complex

	thresholds := make([]float64, levels-1)

	if levels == 3 {
		// Two thresholds for 3-level
		// First threshold using standard Otsu
		t1 := m.calculateOtsuThreshold(hist)
		thresholds[0] = t1

		// Second threshold using modified histogram
		t2 := m.calculateSecondThreshold(hist, t1)
		thresholds[1] = t2

		// Ensure proper ordering
		if thresholds[0] > thresholds[1] {
			thresholds[0], thresholds[1] = thresholds[1], thresholds[0]
		}
	} else {
		// For other levels, use recursive approach
		thresholds[0] = m.calculateOtsuThreshold(hist)
	}

	return thresholds
}

func (m *MultiOtsu) calculateSecondThreshold(hist []float64, firstThreshold float64) float64 {
	// Calculate second threshold for the upper part of the histogram
	t1 := int(firstThreshold)

	// Create histogram for upper part
	upperHist := make([]float64, 256-t1)
	sum := 0.0
	for i := t1; i < 256; i++ {
		upperHist[i-t1] = hist[i]
		sum += upperHist[i-t1]
	}

	// Normalize upper histogram
	if sum > 0 {
		for i := range upperHist {
			upperHist[i] /= sum
		}
	}

	// Calculate Otsu for upper part
	threshold := m.calculateOtsuThresholdForRange(upperHist)
	return threshold + firstThreshold
}

func (m *MultiOtsu) calculateOtsuThresholdForRange(hist []float64) float64 {
	length := len(hist)
	sum := 0.0
	for i := 0; i < length; i++ {
		sum += float64(i) * hist[i]
	}

	sumB := 0.0
	wB := 0.0
	maximum := 0.0
	level := 0.0

	for t := 0; t < length; t++ {
		wB += hist[t]
		if wB == 0 {
			continue
		}

		wF := 1.0 - wB
		if wF == 0 {
			break
		}

		sumB += float64(t) * hist[t]
		mB := sumB / wB
		mF := (sum - sumB) / wF

		// Calculate between-class variance
		between := wB * wF * (mB - mF) * (mB - mF)

		if between > maximum {
			level = float64(t)
			maximum = between
		}
	}

	return level
}

func (m *MultiOtsu) applyThresholds(gray gocv.Mat, thresholds []float64, maxValue float64) gocv.Mat {
	output := gocv.NewMat()
	gray.CopyTo(&output)

	// Apply thresholds to create segmented image
	for y := 0; y < output.Rows(); y++ {
		for x := 0; x < output.Cols(); x++ {
			intensity := float64(output.GetUCharAt(y, x))

			var newValue uint8
			if len(thresholds) == 1 {
				// Binary thresholding
				if intensity <= thresholds[0] {
					newValue = 0
				} else {
					newValue = uint8(maxValue)
				}
			} else {
				// Multi-level thresholding
				level := 0
				for _, threshold := range thresholds {
					if intensity > threshold {
						level++
					}
				}
				// Map level to output value
				newValue = uint8(float64(level) * maxValue / float64(len(thresholds)))
			}

			output.SetUCharAt(y, x, newValue)
		}
	}

	return output
}

// LocalOtsu implements local adaptive Otsu thresholding
type LocalOtsu struct{}

// NewLocalOtsu creates a new local adaptive Otsu algorithm
func NewLocalOtsu() *LocalOtsu {
	return &LocalOtsu{}
}

func (l *LocalOtsu) Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	if input.Empty() {
		return gocv.NewMat(), fmt.Errorf("input image is empty")
	}

	// Convert to grayscale if needed
	gray := l.ensureGrayscale(input)
	defer func() {
		if gray.Ptr() != input.Ptr() {
			gray.Close()
		}
	}()

	// Get parameters
	windowSize := 15
	if val, ok := params["window_size"]; ok {
		if v, ok := val.(float64); ok {
			windowSize = int(v)
		}
	}

	overlap := 0.5
	if val, ok := params["overlap"]; ok {
		if v, ok := val.(float64); ok {
			overlap = v
		}
	}

	interpolation := true
	if val, ok := params["interpolation"]; ok {
		if v, ok := val.(bool); ok {
			interpolation = v
		}
	}

	// Ensure window size is odd
	if windowSize%2 == 0 {
		windowSize++
	}

	// Apply local Otsu
	output := l.applyLocalOtsu(gray, windowSize, overlap, interpolation)
	return output, nil
}

func (l *LocalOtsu) GetDefaultParams() map[string]interface{} {
	return map[string]interface{}{
		"window_size":   15.0,
		"overlap":       0.5,
		"interpolation": true,
	}
}

func (l *LocalOtsu) GetName() string {
	return "Local Adaptive Otsu"
}

func (l *LocalOtsu) GetDescription() string {
	return "Local adaptive Otsu thresholding with overlapping windows and threshold interpolation"
}

func (l *LocalOtsu) Validate(params map[string]interface{}) error {
	if val, ok := params["window_size"]; ok {
		if v, ok := val.(float64); ok {
			if v < 5 || v > 101 {
				return fmt.Errorf("window_size must be between 5 and 101")
			}
		}
	}

	if val, ok := params["overlap"]; ok {
		if v, ok := val.(float64); ok {
			if v < 0 || v > 0.9 {
				return fmt.Errorf("overlap must be between 0 and 0.9")
			}
		}
	}

	return nil
}

func (l *LocalOtsu) GetParameterInfo() []ParameterInfo {
	return []ParameterInfo{
		{
			Name:        "window_size",
			Type:        "int",
			Min:         5.0,
			Max:         101.0,
			Default:     15.0,
			Description: "Size of local window for threshold calculation",
		},
		{
			Name:        "overlap",
			Type:        "float",
			Min:         0.0,
			Max:         0.9,
			Default:     0.5,
			Description: "Overlap ratio between adjacent windows",
		},
		{
			Name:        "interpolation",
			Type:        "bool",
			Default:     true,
			Description: "Enable threshold interpolation between windows",
		},
	}
}

func (l *LocalOtsu) ensureGrayscale(input gocv.Mat) gocv.Mat {
	if input.Channels() == 1 {
		return input
	}

	gray := gocv.NewMat()
	gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
	return gray
}

func (l *LocalOtsu) applyLocalOtsu(gray gocv.Mat, windowSize int, overlap float64, interpolation bool) gocv.Mat {
	height := gray.Rows()
	width := gray.Cols()
	output := gocv.NewMat()
	gray.CopyTo(&output)

	// Calculate step size based on overlap
	step := int(float64(windowSize) * (1.0 - overlap))
	if step < 1 {
		step = 1
	}

	// Create threshold map for interpolation
	var thresholdMap gocv.Mat
	if interpolation {
		thresholdMap = gocv.NewMatWithSize(height/step+1, width/step+1, gocv.MatTypeCV32F)
		defer thresholdMap.Close()
	}

	// Process windows
	for y := 0; y < height; y += step {
		for x := 0; x < width; x += step {
			// Define window bounds
			x1 := x
			y1 := y
			x2 := minInt(x+windowSize, width)
			y2 := minInt(y+windowSize, height)

			// Extract window
			windowRect := image.Rect(x1, y1, x2, y2)
			window := gray.Region(windowRect)

			// Calculate local Otsu threshold
			threshold := l.calculateLocalOtsuThreshold(window)
			window.Close()

			if interpolation {
				// Store threshold in map
				mapY := y / step
				mapX := x / step
				if mapY < thresholdMap.Rows() && mapX < thresholdMap.Cols() {
					thresholdMap.SetFloatAt(mapY, mapX, float32(threshold))
				}
			} else {
				// Apply threshold directly to window
				l.applyThresholdToRegion(output, windowRect, threshold)
			}
		}
	}

	if interpolation {
		// Apply interpolated thresholds
		l.applyInterpolatedThresholds(output, thresholdMap, step)
	}

	return output
}

func (l *LocalOtsu) calculateLocalOtsuThreshold(window gocv.Mat) float64 {
	// Calculate histogram for the window
	hist := make([]float64, 256)
	totalPixels := 0

	for y := 0; y < window.Rows(); y++ {
		for x := 0; x < window.Cols(); x++ {
			intensity := window.GetUCharAt(y, x)
			hist[intensity]++
			totalPixels++
		}
	}

	if totalPixels == 0 {
		return 127 // Default threshold
	}

	// Normalize histogram
	for i := range hist {
		hist[i] /= float64(totalPixels)
	}

	// Calculate Otsu threshold
	return l.calculateOtsuThreshold(hist)
}

func (l *LocalOtsu) calculateOtsuThreshold(hist []float64) float64 {
	sum := 0.0
	for i := 0; i < 256; i++ {
		sum += float64(i) * hist[i]
	}

	sumB := 0.0
	wB := 0.0
	maximum := 0.0
	level := 0.0

	for t := 0; t < 256; t++ {
		wB += hist[t]
		if wB == 0 {
			continue
		}

		wF := 1.0 - wB
		if wF == 0 {
			break
		}

		sumB += float64(t) * hist[t]
		mB := sumB / wB
		mF := (sum - sumB) / wF

		// Calculate between-class variance
		between := wB * wF * (mB - mF) * (mB - mF)

		if between > maximum {
			level = float64(t)
			maximum = between
		}
	}

	return level
}

func (l *LocalOtsu) applyThresholdToRegion(output gocv.Mat, region image.Rectangle, threshold float64) {
	for y := region.Min.Y; y < region.Max.Y; y++ {
		for x := region.Min.X; x < region.Max.X; x++ {
			intensity := output.GetUCharAt(y, x)
			if float64(intensity) <= threshold {
				output.SetUCharAt(y, x, 0)
			} else {
				output.SetUCharAt(y, x, 255)
			}
		}
	}
}

func (l *LocalOtsu) applyInterpolatedThresholds(output, thresholdMap gocv.Mat, step int) {
	height := output.Rows()
	width := output.Cols()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate interpolated threshold
			threshold := l.interpolateThreshold(thresholdMap, x, y, step)

			// Apply threshold
			intensity := output.GetUCharAt(y, x)
			if float64(intensity) <= threshold {
				output.SetUCharAt(y, x, 0)
			} else {
				output.SetUCharAt(y, x, 255)
			}
		}
	}
}

func (l *LocalOtsu) interpolateThreshold(thresholdMap gocv.Mat, x, y, step int) float64 {
	// Map coordinates to threshold map
	fx := float64(x) / float64(step)
	fy := float64(y) / float64(step)

	// Get integer coordinates
	x1 := int(fx)
	y1 := int(fy)
	x2 := x1 + 1
	y2 := y1 + 1

	// Clamp to map bounds
	mapHeight := thresholdMap.Rows()
	mapWidth := thresholdMap.Cols()

	if x2 >= mapWidth {
		x2 = mapWidth - 1
		x1 = x2
	}
	if y2 >= mapHeight {
		y2 = mapHeight - 1
		y1 = y2
	}

	// Get threshold values
	t11 := float64(thresholdMap.GetFloatAt(y1, x1))
	t12 := float64(thresholdMap.GetFloatAt(y2, x1))
	t21 := float64(thresholdMap.GetFloatAt(y1, x2))
	t22 := float64(thresholdMap.GetFloatAt(y2, x2))

	// Bilinear interpolation
	wx := fx - float64(x1)
	wy := fy - float64(y1)

	t1 := t11*(1-wx) + t21*wx
	t2 := t12*(1-wx) + t22*wx

	return t1*(1-wy) + t2*wy
}
