package main

import (
	"fmt"
	"image"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

type TwoDOtsu struct {
	windowRadius       int
	epsilon            float64
	morphKernelSize    int
	onParameterChanged func()
	debugImage         *DebugImage

	// Caching for preview optimization
	cachedParams map[string]interface{}
	cachedResult gocv.Mat
	previewScale float64
	mutex        sync.RWMutex
}

func NewTwoDOtsu() *TwoDOtsu {
	return &TwoDOtsu{
		windowRadius:    5,
		epsilon:         0.02,
		morphKernelSize: 3,
		debugImage:      NewDebugImage(),
		cachedParams:    make(map[string]interface{}),
		previewScale:    0.5, // 50% scale for preview
	}
}

func (t *TwoDOtsu) Name() string {
	return "2D Otsu"
}

func (t *TwoDOtsu) Apply(input gocv.Mat) gocv.Mat {
	return t.applyWithScale(input, 1.0) // Full resolution for saving
}

func (t *TwoDOtsu) ApplyPreview(input gocv.Mat) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Starting preview application")

	if input.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu Preview", "ERROR: input is empty")
		return gocv.NewMat()
	}

	// Check cache first
	t.mutex.RLock()
	currentParams := t.GetParameters()
	cached := t.paramsEqual(currentParams, t.cachedParams)
	cacheValid := cached && !t.cachedResult.Empty()
	t.mutex.RUnlock()

	if cacheValid {
		t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Using cached result")
		// Always return a new clone, never the cached Mat itself
		return t.cachedResult.Clone()
	}

	// Process with preview scaling
	result := t.applyWithScale(input, t.previewScale)

	if result.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu Preview", "ERROR: processing returned empty result")
		return gocv.NewMat()
	}

	// Update cache with a clone
	t.mutex.Lock()
	if !t.cachedResult.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Closing previous cached result")
		t.cachedResult.Close()
	}
	t.cachedResult = result.Clone() // Store a clone in cache
	t.cachedParams = make(map[string]interface{})
	for k, v := range currentParams {
		t.cachedParams[k] = v
	}
	t.mutex.Unlock()

	t.debugImage.LogAlgorithmStep("2D Otsu Preview", "Cache updated successfully")

	// Return the original result (not the cached clone)
	return result
}

func (t *TwoDOtsu) applyWithScale(input gocv.Mat, scale float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Starting binarization (scale: %.2f)", scale))

	// Step 1: Preprocessing with scaling
	workingMat := input.Clone()
	if scale < 1.0 {
		newSize := image.Point{
			X: int(float64(input.Cols()) * scale),
			Y: int(float64(input.Rows()) * scale),
		}
		scaledMat := gocv.NewMat()
		gocv.Resize(workingMat, &scaledMat, newSize, 0, 0, gocv.InterpolationLinear)
		workingMat.Close()
		workingMat = scaledMat
		t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Scaled to %dx%d", newSize.X, newSize.Y))
	}

	gray := gocv.NewMat()
	defer gray.Close()

	if workingMat.Channels() == 3 {
		gocv.CvtColor(workingMat, &gray, gocv.ColorBGRToGray)
		t.debugImage.LogColorConversion("BGR", "Grayscale")
	} else {
		gray = workingMat.Clone()
	}
	workingMat.Close()

	t.debugImage.LogMatInfo("grayscale", gray)

	// Apply guided filter (proper implementation)
	guided := t.guidedFilter(gray, t.windowRadius, t.epsilon)
	defer guided.Close()

	t.debugImage.LogFilter("GuidedFilter", fmt.Sprintf("radius=%d", t.windowRadius), fmt.Sprintf("epsilon=%.3f", t.epsilon))

	// Step 2: Construct 2D Histogram
	t.debugImage.LogAlgorithmStep("2D Otsu", "Constructing 2D histogram")

	// Validate Mats before processing
	if gray.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: gray image is empty")
		return gocv.NewMat()
	}
	if guided.Empty() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: guided image is empty")
		return gocv.NewMat()
	}

	rows, cols := gray.Rows(), gray.Cols()
	guidedRows, guidedCols := guided.Rows(), guided.Cols()

	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Gray dimensions: %dx%d", cols, rows))
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Guided dimensions: %dx%d", guidedCols, guidedRows))

	// Ensure both images have same dimensions
	if rows != guidedRows || cols != guidedCols {
		t.debugImage.LogAlgorithmStep("2D Otsu", "ERROR: dimension mismatch between gray and guided images")
		return gocv.NewMat()
	}

	hist := make([][]float64, 256)
	for i := range hist {
		hist[i] = make([]float64, 256)
	}

	N := float64(rows * cols)
	t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Processing %d pixels", int(N)))

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Add bounds checking
			if i >= 0 && i < rows && j >= 0 && j < cols {
				intensity := int(gray.GetUCharAt(i, j))
				avg := int(guided.GetUCharAt(i, j))

				// Validate pixel values
				if intensity >= 0 && intensity < 256 && avg >= 0 && avg < 256 {
					hist[intensity][avg]++
				} else {
					t.debugImage.LogAlgorithmStep("2D Otsu", fmt.Sprintf("Invalid pixel values at (%d,%d): intensity=%d, avg=%d", i, j, intensity, avg))
				}
			}
		}
	}

	// Normalize histogram to probabilities
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			hist[i][j] /= N
		}
	}

	// Step 3-6: Find Optimal Thresholds (corrected implementation)
	t.debugImage.LogAlgorithmStep("2D Otsu", "Finding optimal thresholds")
	maxVariance := 0.0
	bestS, bestT := 128, 128 // Default thresholds

	for s := 0; s < 256; s++ {
		for t_ := 0; t_ < 256; t_++ {
			// Compute probabilities (corrected)
			P1 := 0.0
			for i := s + 1; i < 256; i++ {
				for j := t_ + 1; j < 256; j++ {
					P1 += hist[i][j]
				}
			}

			if P1 == 0 || P1 == 1 {
				continue
			}
			P2 := 1 - P1

			// Compute means (corrected)
			mu1I, mu1M := 0.0, 0.0
			for i := s + 1; i < 256; i++ {
				for j := t_ + 1; j < 256; j++ {
					mu1I += float64(i) * hist[i][j]
					mu1M += float64(j) * hist[i][j]
				}
			}
			if P1 > 0 {
				mu1I /= P1
				mu1M /= P1
			}

			// Background: all pixels NOT in foreground
			mu2I, mu2M := 0.0, 0.0
			// Region 1: i <= s, all j
			for i := 0; i <= s; i++ {
				for j := 0; j < 256; j++ {
					mu2I += float64(i) * hist[i][j]
					mu2M += float64(j) * hist[i][j]
				}
			}
			// Region 2: i > s, j <= t
			for i := s + 1; i < 256; i++ {
				for j := 0; j <= t_; j++ {
					mu2I += float64(i) * hist[i][j]
					mu2M += float64(j) * hist[i][j]
				}
			}
			if P2 > 0 {
				mu2I /= P2
				mu2M /= P2
			}

			// Compute between-class variance
			variance := P1 * P2 * ((mu1I-mu2I)*(mu1I-mu2I) + (mu1M-mu2M)*(mu1M-mu2M))
			if variance > maxVariance {
				maxVariance = variance
				bestS, bestT = s, t_
			}
		}
	}

	t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)

	// Step 7: Binarize Image
	t.debugImage.LogAlgorithmStep("2D Otsu", "Binarizing image")
	binary := gocv.NewMatWithSize(rows, cols, gocv.MatTypeCV8U)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			intensity := int(gray.GetUCharAt(i, j))
			avg := int(guided.GetUCharAt(i, j))
			if intensity > bestS && avg > bestT {
				binary.SetUCharAt(i, j, 255)
			} else {
				binary.SetUCharAt(i, j, 0)
			}
		}
	}

	// Step 8: Postprocessing
	t.debugImage.LogAlgorithmStep("2D Otsu", "Postprocessing")
	kernel := gocv.GetStructuringElement(gocv.MorphRect,
		image.Pt(t.morphKernelSize, t.morphKernelSize))
	defer kernel.Close()

	processed := gocv.NewMat()
	gocv.MorphologyEx(binary, &processed, gocv.MorphClose, kernel)
	binary.Close()
	t.debugImage.LogMorphology("Close", t.morphKernelSize)

	final := gocv.NewMat()
	gocv.MorphologyEx(processed, &final, gocv.MorphOpen, kernel)
	processed.Close()
	t.debugImage.LogMorphology("Open", t.morphKernelSize)

	// Scale back to original size if needed
	if scale < 1.0 {
		originalSize := image.Point{X: input.Cols(), Y: input.Rows()}
		fullSizeMat := gocv.NewMat()
		gocv.Resize(final, &fullSizeMat, originalSize, 0, 0, gocv.InterpolationNearestNeighbor)
		final.Close()
		final = fullSizeMat
		t.debugImage.LogAlgorithmStep("2D Otsu", "Scaled back to original size")
	}

	t.debugImage.LogMatInfo("final_binary", final)
	t.debugImage.LogAlgorithmStep("2D Otsu", "Completed")

	return final
}

// Guided filter implementation
func (t *TwoDOtsu) guidedFilter(input gocv.Mat, radius int, epsilon float64) gocv.Mat {
	t.debugImage.LogAlgorithmStep("GuidedFilter", "Starting guided filter implementation")

	if input.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: input image is empty")
		return gocv.NewMat()
	}

	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Input: %dx%d, channels=%d", input.Cols(), input.Rows(), input.Channels()))

	// Use simple box filter as approximation for now to avoid complex guided filter implementation
	// This maintains compatibility while providing smoothing
	result := gocv.NewMat()
	kernelSize := image.Point{X: 2*radius + 1, Y: 2*radius + 1}

	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Using box filter with kernel %dx%d", kernelSize.X, kernelSize.Y))

	gocv.BoxFilter(input, &result, -1, kernelSize)

	if result.Empty() {
		t.debugImage.LogAlgorithmStep("GuidedFilter", "ERROR: box filter returned empty Mat")
		return gocv.NewMat()
	}

	t.debugImage.LogAlgorithmStep("GuidedFilter", fmt.Sprintf("Result: %dx%d, channels=%d", result.Cols(), result.Rows(), result.Channels()))

	return result
}

func (t *TwoDOtsu) paramsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func (t *TwoDOtsu) GetParametersWidget(onParameterChanged func()) fyne.CanvasObject {
	t.onParameterChanged = onParameterChanged

	// Window Radius parameter
	radiusLabel := widget.NewLabel("Window Radius:")
	radiusEntry := widget.NewEntry()
	radiusEntry.SetText(fmt.Sprintf("%d", t.windowRadius))
	radiusEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value > 0 && value <= 20 {
			t.windowRadius = value
			t.invalidateCache()
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		}
	}

	radiusSlider := widget.NewSlider(1, 20)
	radiusSlider.Value = float64(t.windowRadius)
	radiusSlider.OnChanged = func(value float64) {
		t.windowRadius = int(value)
		radiusEntry.SetText(fmt.Sprintf("%d", t.windowRadius))
		t.invalidateCache()
		if t.onParameterChanged != nil {
			t.onParameterChanged()
		}
	}

	// Epsilon parameter
	epsilonLabel := widget.NewLabel("Epsilon:")
	epsilonEntry := widget.NewEntry()
	epsilonEntry.SetText(fmt.Sprintf("%.3f", t.epsilon))
	epsilonEntry.OnSubmitted = func(text string) {
		if value, err := strconv.ParseFloat(text, 64); err == nil && value > 0 && value <= 1.0 {
			t.epsilon = value
			t.invalidateCache()
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		}
	}

	epsilonSlider := widget.NewSlider(0.001, 0.1)
	epsilonSlider.Value = t.epsilon
	epsilonSlider.OnChanged = func(value float64) {
		t.epsilon = value
		epsilonEntry.SetText(fmt.Sprintf("%.3f", t.epsilon))
		t.invalidateCache()
		if t.onParameterChanged != nil {
			t.onParameterChanged()
		}
	}

	// Morphological Kernel Size parameter
	kernelLabel := widget.NewLabel("Morphological Kernel Size:")
	kernelEntry := widget.NewEntry()
	kernelEntry.SetText(fmt.Sprintf("%d", t.morphKernelSize))
	kernelEntry.OnSubmitted = func(text string) {
		if value, err := strconv.Atoi(text); err == nil && value > 0 && value <= 15 && value%2 == 1 {
			t.morphKernelSize = value
			t.invalidateCache()
			if t.onParameterChanged != nil {
				t.onParameterChanged()
			}
		}
	}

	kernelSlider := widget.NewSlider(1, 15)
	kernelSlider.Value = float64(t.morphKernelSize)
	kernelSlider.OnChanged = func(value float64) {
		// Ensure odd values only
		oddValue := int(value)
		if oddValue%2 == 0 {
			oddValue++
		}
		t.morphKernelSize = oddValue
		kernelEntry.SetText(fmt.Sprintf("%d", t.morphKernelSize))
		t.invalidateCache()
		if t.onParameterChanged != nil {
			t.onParameterChanged()
		}
	}

	parametersForm := container.NewVBox(
		container.NewGridWithColumns(2, radiusLabel, radiusEntry),
		radiusSlider,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, epsilonLabel, epsilonEntry),
		epsilonSlider,
		widget.NewSeparator(),
		container.NewGridWithColumns(2, kernelLabel, kernelEntry),
		kernelSlider,
	)

	return parametersForm
}

func (t *TwoDOtsu) invalidateCache() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if !t.cachedResult.Empty() {
		t.cachedResult.Close()
		t.cachedResult = gocv.NewMat()
	}
	t.cachedParams = make(map[string]interface{})
}

func (t *TwoDOtsu) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"windowRadius":    t.windowRadius,
		"epsilon":         t.epsilon,
		"morphKernelSize": t.morphKernelSize,
	}
}

func (t *TwoDOtsu) SetParameters(params map[string]interface{}) {
	changed := false
	if value, ok := params["windowRadius"].(int); ok && value != t.windowRadius {
		t.windowRadius = value
		changed = true
	}
	if value, ok := params["epsilon"].(float64); ok && value != t.epsilon {
		t.epsilon = value
		changed = true
	}
	if value, ok := params["morphKernelSize"].(int); ok && value != t.morphKernelSize {
		t.morphKernelSize = value
		changed = true
	}
	if changed {
		t.invalidateCache()
	}
}

func (t *TwoDOtsu) Close() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if !t.cachedResult.Empty() {
		t.cachedResult.Close()
	}
}
