// internal/gui/right_panel.go
// Perfect UI Right Panel: Feedback and Insights (300px wide)
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type RightPanel struct {
	logger *slog.Logger

	container *fyne.Container

	// Status Section (80px height)
	statusContainer *fyne.Container
	fileLabel       *widget.Label
	stateLabel      *widget.Label
	lastActionLabel *widget.Label

	// Quality Metrics Section (180px height)
	metricsContainer *fyne.Container
	currentMetrics   map[string]float64

	// Tools Section (650px height - future-proofed)
	toolsContainer *fyne.Container
	
	// Additional sections for image info
	imageInfoContainer *fyne.Container
}

func NewRightPanel(logger *slog.Logger) *RightPanel {
	panel := &RightPanel{
		logger:         logger,
		currentMetrics: make(map[string]float64),
	}

	panel.initializeUI()
	return panel
}

func (rp *RightPanel) initializeUI() {
	// Status Section (80px height)
	rp.fileLabel = widget.NewLabel("No file loaded")
	rp.stateLabel = widget.NewLabel("Ready")
	rp.lastActionLabel = widget.NewLabel("Application started")

	rp.statusContainer = container.NewVBox(
		widget.NewCard("Status", "", container.NewVBox(
			container.NewHBox(widget.NewLabel("File:"), rp.fileLabel),
			container.NewHBox(widget.NewLabel("State:"), rp.stateLabel),
			container.NewHBox(widget.NewLabel("Last:"), rp.lastActionLabel),
		)),
	)

	// Quality Metrics Section (180px height)
	rp.metricsContainer = container.NewVBox(
		widget.NewCard("Quality Metrics", "", container.NewVBox(
			widget.NewLabel("Real-time quality metrics will appear here during processing."),
		)),
	)

	// Image Information Container
	rp.imageInfoContainer = container.NewVBox()

	// Tools Section (650px height - placeholder for future features)
	rp.toolsContainer = container.NewVBox(
		widget.NewCard("Tools (Future)", "", container.NewVBox(
			widget.NewLabel("🔧 Histogram"),
			widget.NewLabel("Image histogram will appear here when implemented."),
			widget.NewSeparator(),
			widget.NewLabel("🎨 Color Picker"),
			widget.NewLabel("Color picking tool will be available here."),
			widget.NewSeparator(),
			widget.NewLabel("📊 Analysis Tools"),
			widget.NewLabel("Additional analysis features will be added here."),
		)),
	)

	// Main container with sections
	content := container.NewVBox(
		rp.statusContainer,
		widget.NewSeparator(),
		rp.imageInfoContainer,
		widget.NewSeparator(),
		rp.metricsContainer,
		widget.NewSeparator(),
		rp.toolsContainer,
	)

	// Create scroll container and set fixed width to 300px as per specification
	scroll := container.NewScroll(content)
	rp.container = container.NewBorder(nil, nil, nil, nil, scroll)
	rp.container.Resize(fyne.NewSize(300, 1000))
}

func (rp *RightPanel) GetContainer() fyne.CanvasObject {
	return rp.container
}

func (rp *RightPanel) UpdateMetrics(metrics map[string]float64) {
	rp.currentMetrics = metrics
	rp.refreshMetricsDisplay()
}

func (rp *RightPanel) refreshMetricsDisplay() {
	// Clear existing metrics content
	metricsContent := container.NewVBox()

	if len(rp.currentMetrics) == 0 {
		metricsContent.Add(widget.NewLabel("Processing..."))
	} else {
		// PSNR Metric
		if psnr, exists := rp.currentMetrics["psnr"]; exists {
			psnrWidget := rp.createPSNRWidget(psnr)
			metricsContent.Add(psnrWidget)
		}

		// SSIM Metric
		if ssim, exists := rp.currentMetrics["ssim"]; exists {
			ssimWidget := rp.createSSIMWidget(ssim)
			metricsContent.Add(ssimWidget)
		}

		// MSE Metric
		if mse, exists := rp.currentMetrics["mse"]; exists {
			mseWidget := rp.createMSEWidget(mse)
			metricsContent.Add(mseWidget)
		}
	}

	// Update metrics container
	rp.metricsContainer.RemoveAll()
	rp.metricsContainer.Add(widget.NewCard("Quality Metrics", "", metricsContent))
	rp.metricsContainer.Refresh()
}

func (rp *RightPanel) createPSNRWidget(psnr float64) fyne.CanvasObject {
	valueLabel := widget.NewLabel(fmt.Sprintf("PSNR: %.2f dB", psnr))
	
	// Quality assessment with color-coded bar
	var qualityText string
	var barColor fyne.Resource
	var progress float32

	if psnr > 40 {
		qualityText = "Excellent"
		barColor = theme.ConfirmIcon()
		progress = 1.0
	} else if psnr > 30 {
		qualityText = "Good" 
		barColor = theme.InfoIcon()
		progress = 0.75
	} else if psnr > 20 {
		qualityText = "Fair"
		barColor = theme.WarningIcon()
		progress = 0.5
	} else {
		qualityText = "Poor"
		barColor = theme.ErrorIcon()
		progress = 0.25
	}

	qualityBar := widget.NewProgressBar()
	qualityBar.SetValue(float64(progress))

	qualityRow := container.NewHBox(
		widget.NewIcon(barColor),
		widget.NewLabel(qualityText),
	)

	return container.NewVBox(
		valueLabel,
		qualityBar,
		qualityRow,
		widget.NewSeparator(),
	)
}

func (rp *RightPanel) createSSIMWidget(ssim float64) fyne.CanvasObject {
	valueLabel := widget.NewLabel(fmt.Sprintf("SSIM: %.3f", ssim))
	
	// Quality assessment with color-coded bar
	var qualityText string
	var barColor fyne.Resource
	var progress float32

	if ssim > 0.95 {
		qualityText = "Excellent"
		barColor = theme.ConfirmIcon()
		progress = 1.0
	} else if ssim > 0.8 {
		qualityText = "Good"
		barColor = theme.InfoIcon()
		progress = 0.75
	} else if ssim > 0.6 {
		qualityText = "Fair"
		barColor = theme.WarningIcon()
		progress = 0.5
	} else {
		qualityText = "Poor"
		barColor = theme.ErrorIcon()
		progress = 0.25
	}

	qualityBar := widget.NewProgressBar()
	qualityBar.SetValue(float64(progress))

	qualityRow := container.NewHBox(
		widget.NewIcon(barColor),
		widget.NewLabel(qualityText),
	)

	return container.NewVBox(
		valueLabel,
		qualityBar,
		qualityRow,
		widget.NewSeparator(),
	)
}

func (rp *RightPanel) createMSEWidget(mse float64) fyne.CanvasObject {
	valueLabel := widget.NewLabel(fmt.Sprintf("MSE: %.2f", mse))
	
	// Quality assessment (lower MSE is better)
	var qualityText string
	var barColor fyne.Resource
	var progress float32

	if mse < 100 {
		qualityText = "Excellent"
		barColor = theme.ConfirmIcon()
		progress = 1.0
	} else if mse < 500 {
		qualityText = "Good"
		barColor = theme.InfoIcon()
		progress = 0.75
	} else if mse < 1000 {
		qualityText = "Fair"
		barColor = theme.WarningIcon()
		progress = 0.5
	} else {
		qualityText = "Poor"
		barColor = theme.ErrorIcon()
		progress = 0.25
	}

	qualityBar := widget.NewProgressBar()
	qualityBar.SetValue(float64(progress))

	qualityRow := container.NewHBox(
		widget.NewIcon(barColor),
		widget.NewLabel(qualityText),
	)

	return container.NewVBox(
		valueLabel,
		qualityBar,
		qualityRow,
		widget.NewSeparator(),
	)
}

func (rp *RightPanel) ShowImageInfo(filepath string, width, height, channels int) {
	// Truncate filepath if too long (>200px equivalent ~30 chars)
	displayPath := filepath
	if len(filepath) > 30 {
		displayPath = "..." + filepath[len(filepath)-27:]
	}

	imageInfoContent := container.NewVBox(
		widget.NewLabel(fmt.Sprintf("Path: %s", displayPath)),
		widget.NewLabel(fmt.Sprintf("Size: %dx%d", width, height)),
		widget.NewLabel(fmt.Sprintf("Channels: %d", channels)),
	)

	rp.imageInfoContainer.RemoveAll()
	rp.imageInfoContainer.Add(widget.NewCard("Image Information", "", imageInfoContent))
	rp.imageInfoContainer.Refresh()

	// Update status
	rp.fileLabel.SetText(displayPath)
	rp.stateLabel.SetText("Image Loaded")
	rp.lastActionLabel.SetText("Image loaded successfully")
}

func (rp *RightPanel) ShowMessage(message string) {
	rp.lastActionLabel.SetText(message)
}

func (rp *RightPanel) ShowError(err error) {
	rp.stateLabel.SetText("Error")
	rp.lastActionLabel.SetText(fmt.Sprintf("Error: %s", err.Error()))
}

func (rp *RightPanel) Clear() {
	rp.currentMetrics = make(map[string]float64)
	rp.refreshMetricsDisplay()
	
	rp.imageInfoContainer.RemoveAll()
	rp.imageInfoContainer.Refresh()

	rp.fileLabel.SetText("No file loaded")
	rp.stateLabel.SetText("Ready")
	rp.lastActionLabel.SetText("Application ready")
}

func (rp *RightPanel) UpdateStatus(file, state, lastAction string) {
	if file != "" {
		// Truncate if too long
		if len(file) > 30 {
			file = "..." + file[len(file)-27:]
		}
		rp.fileLabel.SetText(file)
	}
	
	if state != "" {
		rp.stateLabel.SetText(state)
	}
	
	if lastAction != "" {
		rp.lastActionLabel.SetText(lastAction)
	}
}

func (rp *RightPanel) SetProcessingState() {
	rp.stateLabel.SetText("Processing...")
}

func (rp *RightPanel) SetReadyState() {
	rp.stateLabel.SetText("Ready")
}

func (rp *RightPanel) AddLastAction(action string) {
	rp.lastActionLabel.SetText(action)
}