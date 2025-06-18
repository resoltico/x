package main

import (
	"fmt"
	"image"
	"strconv"

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
}

func NewTwoDOtsu() *TwoDOtsu {
	return &TwoDOtsu{
		windowRadius:    5,
		epsilon:         0.02,
		morphKernelSize: 3,
		debugImage:      NewDebugImage(),
	}
}

func (t *TwoDOtsu) Name() string {
	return "2D Otsu"
}

func (t *TwoDOtsu) Apply(input gocv.Mat) gocv.Mat {
	if t.debugImage.IsEnabled() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "Starting binarization")
	}

	// Step 1: Preprocessing
	gray := gocv.NewMat()
	defer gray.Close()

	if input.Channels() == 3 {
		gocv.CvtColor(input, &gray, gocv.ColorBGRToGray)
		if t.debugImage.IsEnabled() {
			t.debugImage.LogColorConversion("BGR", "Grayscale")
		}
	} else {
		gray = input.Clone()
	}
	if t.debugImage.IsEnabled() {
		t.debugImage.LogMatInfo("grayscale", gray)
	}

	// Apply bilateral filter as approximation for guided filter
	guided := gocv.NewMat()
	defer guided.Close()

	gocv.BilateralFilter(gray, &guided, -1, 80, 80)
	if t.debugImage.IsEnabled() {
		t.debugImage.LogFilter("BilateralFilter", "radius=-1", "sigmaColor=80", "sigmaSpace=80")
	}

	// Step 2: Construct 2D Histogram
	if t.debugImage.IsEnabled() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "Constructing 2D histogram")
	}
	hist := make([][]float64, 256)
	for i := range hist {
		hist[i] = make([]float64, 256)
	}

	rows, cols := gray.Rows(), gray.Cols()
	N := float64(rows * cols)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			intensity := int(gray.GetUCharAt(i, j))
			avg := int(guided.GetUCharAt(i, j))
			hist[intensity][avg]++
		}
	}

	// Normalize histogram to probabilities
	for i := 0; i < 256; i++ {
		for j := 0; j < 256; j++ {
			hist[i][j] /= N
		}
	}

	// Step 3-6: Find Optimal Thresholds
	if t.debugImage.IsEnabled() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "Finding optimal thresholds")
	}
	maxVariance := 0.0
	bestS, bestT := 128, 128 // Default thresholds

	for s := 0; s < 256; s++ {
		for t_ := 0; t_ < 256; t_++ {
			// Compute probabilities
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

			// Compute means
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

			mu2I, mu2M := 0.0, 0.0
			for i := 0; i <= s; i++ {
				for j := 0; j < 256; j++ {
					mu2I += float64(i) * hist[i][j]
					mu2M += float64(j) * hist[i][j]
				}
			}
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

	if t.debugImage.IsEnabled() {
		t.debugImage.LogOptimalThresholds(bestS, bestT, maxVariance)
	}

	// Step 7: Binarize Image
	if t.debugImage.IsEnabled() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "Binarizing image")
	}
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
	if t.debugImage.IsEnabled() {
		t.debugImage.LogAlgorithmStep("2D Otsu", "Postprocessing")
	}
	kernel := gocv.GetStructuringElement(gocv.MorphRect,
		image.Pt(t.morphKernelSize, t.morphKernelSize))
	defer kernel.Close()

	processed := gocv.NewMat()
	gocv.MorphologyEx(binary, &processed, gocv.MorphClose, kernel)
	binary.Close()
	if t.debugImage.IsEnabled() {
		t.debugImage.LogMorphology("Close", t.morphKernelSize)
	}

	final := gocv.NewMat()
	gocv.MorphologyEx(processed, &final, gocv.MorphOpen, kernel)
	processed.Close()
	if t.debugImage.IsEnabled() {
		t.debugImage.LogMorphology("Open", t.morphKernelSize)
		t.debugImage.LogMatInfo("final_binary", final)
		t.debugImage.LogAlgorithmStep("2D Otsu", "Completed")
	}

	return final
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

func (t *TwoDOtsu) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"windowRadius":    t.windowRadius,
		"epsilon":         t.epsilon,
		"morphKernelSize": t.morphKernelSize,
	}
}

func (t *TwoDOtsu) SetParameters(params map[string]interface{}) {
	if value, ok := params["windowRadius"].(int); ok {
		t.windowRadius = value
	}
	if value, ok := params["epsilon"].(float64); ok {
		t.epsilon = value
	}
	if value, ok := params["morphKernelSize"].(int); ok {
		t.morphKernelSize = value
	}
}
