// internal/gui/right_panel.go
// Perfect UI Right Panel: Status and Metrics (250px wide)
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

	// Status section
	statusCard      *widget.Card
	fileLabel       *widget.Label
	stateLabel      *widget.Label
	lastActionLabel *widget.Label

	// Image information section
	imageInfoCard *widget.Card
	pathLabel     *widget.Label
	sizeLabel     *widget.Label
	channelsLabel *widget.Label

	// Quality metrics section
	metricsCard      *widget.Card
	metricsContainer *fyne.Container
	currentMetrics   map[string]float64

	// Tools section (future)
	toolsCard *widget.Card
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
	// Status section
	rp.createStatusSection()

	// Image information section
	rp.createImageInfoSection()

	// Quality metrics section
	rp.createMetricsSection()

	// Tools section (future features)
	rp.createToolsSection()

	// Main container with sections
	content := container.NewVBox(
		rp.statusCard,
		rp.imageInfoCard,
		rp.metricsCard,
		rp.toolsCard,
	)

	// Create scroll container and set fixed width to 250px as per specification
	scroll := container.NewScroll(content)
	rp.container = container.NewBorder(nil, nil, nil, nil, scroll)
	rp.container.Resize(fyne.NewSize(250, 850)) // 850px available height (900 - 50 toolbar)
}

func (rp *RightPanel) createStatusSection() {
	rp.fileLabel = widget.NewLabel("No image loaded")
	rp.stateLabel = widget.NewLabel("Ready")
	rp.lastActionLabel = widget.NewLabel("Application started")

	statusContent := container.NewVBox(
		container.NewHBox(widget.NewLabel("File:"), rp.fileLabel),
		container.NewHBox(widget.NewLabel("State:"), rp.stateLabel),
		container.NewHBox(widget.NewLabel("Last:"), rp.lastActionLabel),
	)

	rp.statusCard = widget.NewCard("STATUS", "", statusContent)
}

func (rp *RightPanel) createImageInfoSection() {
	rp.pathLabel = widget.NewLabel("No path")
	rp.sizeLabel = widget.NewLabel("No size")
	rp.channelsLabel = widget.NewLabel("No channels")

	imageInfoContent := container.NewVBox(
		container.NewHBox(widget.NewLabel("Path:"), rp.pathLabel),
		container.NewHBox(widget.NewLabel("Size:"), rp.sizeLabel),
		container.NewHBox(widget.NewLabel("Channels:"), rp.channelsLabel),
	)

	rp.imageInfoCard = widget.NewCard("IMAGE INFORMATION", "", imageInfoContent)
}

func (rp *RightPanel) createMetricsSection() {
	rp.metricsContainer = container.NewVBox(
		widget.NewLabel("Real-time quality metrics will appear here during processing."),
	)

	rp.metricsCard = widget.NewCard("QUALITY METRICS", "", rp.metricsContainer)
}

func (rp *RightPanel) createToolsSection() {
	toolsContent := container.NewVBox(
		// Histogram placeholder
		container.NewHBox(
			widget.NewIcon(theme.InfoIcon()),
			widget.NewLabel("Histogram"),
		),
		widget.NewLabel("Image histogram will appear here when implemented."),
		widget.NewSeparator(),

		// Color Picker placeholder
		container.NewHBox(
			widget.NewIcon(theme.ColorPaletteIcon()),
			widget.NewLabel("Color Picker"),
		),
		widget.NewLabel("Color picking tool will be available here."),
		widget.NewSeparator(),

		// Analysis Tools placeholder
		container.NewHBox(
			widget.NewIcon(theme.ComputerIcon()),
			widget.NewLabel("Analysis Tools"),
		),
		widget.NewLabel("Additional analysis features will be added here."),
	)

	rp.toolsCard = widget.NewCard("TOOLS", "", toolsContent)
}

func (rp *RightPanel) UpdateMetrics(metrics map[string]float64) {
	rp.currentMetrics = metrics
	rp.refreshMetricsDisplay()
}

func (rp *RightPanel) refreshMetricsDisplay() {
	// Clear existing metrics content
	rp.metricsContainer.RemoveAll()

	if len(rp.currentMetrics) == 0 {
		rp.metricsContainer.Add(widget.NewLabel("No metrics available"))
	} else {
		// PSNR Metric
		if psnr, exists := rp.currentMetrics["psnr"]; exists {
			psnrWidget := rp.createPSNRWidget(psnr)
			rp.metricsContainer.Add(psnrWidget)
		}

		// SSIM Metric
		if ssim, exists := rp.currentMetrics["ssim"]; exists {
			ssimWidget := rp.createSSIMWidget(ssim)
			rp.metricsContainer.Add(ssimWidget)
		}

		// MSE Metric (if available)
		if mse, exists := rp.currentMetrics["mse"]; exists {
			mseWidget := rp.createMSEWidget(mse)
			rp.metricsContainer.Add(mseWidget)
		}
	}

	rp.metricsContainer.Refresh()
}

func (rp *RightPanel) createPSNRWidget(psnr float64) fyne.CanvasObject {
	valueLabel := widget.NewLabel(fmt.Sprintf("PSNR: %.2f dB", psnr))

	// Quality assessment with progress bar
	var qualityText string
	var progress float64

	if psnr > 40 {
		qualityText = "Excellent"
		progress = 1.0
	} else if psnr > 30 {
		qualityText = "Good"
		progress = 0.75
	} else if psnr > 20 {
		qualityText = "Fair"
		progress = 0.5
	} else {
		qualityText = "Poor"
		progress = 0.25
	}

	progressBar := widget.NewProgressBar()
	progressBar.SetValue(progress)

	qualityLabel := widget.NewLabel(qualityText)

	return container.NewVBox(
		valueLabel,
		progressBar,
		qualityLabel,
		widget.NewSeparator(),
	)
}

func (rp *RightPanel) createSSIMWidget(ssim float64) fyne.CanvasObject {
	valueLabel := widget.NewLabel(fmt.Sprintf("SSIM: %.3f", ssim))

	// Quality assessment with progress bar
	var qualityText string
	var progress float64

	if ssim > 0.95 {
		qualityText = "Excellent"
		progress = 1.0
	} else if ssim > 0.8 {
		qualityText = "Good"
		progress = 0.75
	} else if ssim > 0.6 {
		qualityText = "Fair"
		progress = 0.5
	} else {
		qualityText = "Poor"
		progress = 0.25
	}

	progressBar := widget.NewProgressBar()
	progressBar.SetValue(progress)

	qualityLabel := widget.NewLabel(qualityText)

	return container.NewVBox(
		valueLabel,
		progressBar,
		qualityLabel,
		widget.NewSeparator(),
	)
}

func (rp *RightPanel) createMSEWidget(mse float64) fyne.CanvasObject {
	valueLabel := widget.NewLabel(fmt.Sprintf("MSE: %.2f", mse))

	// Quality assessment (lower MSE is better)
	var qualityText string
	var progress float64

	if mse < 100 {
		qualityText = "Excellent"
		progress = 1.0
	} else if mse < 500 {
		qualityText = "Good"
		progress = 0.75
	} else if mse < 1000 {
		qualityText = "Fair"
		progress = 0.5
	} else {
		qualityText = "Poor"
		progress = 0.25
	}

	progressBar := widget.NewProgressBar()
	progressBar.SetValue(progress)

	qualityLabel := widget.NewLabel(qualityText)

	return container.NewVBox(
		valueLabel,
		progressBar,
		qualityLabel,
		widget.NewSeparator(),
	)
}

func (rp *RightPanel) ShowImageInfo(filepath string, width, height, channels int) {
	// Truncate filepath if too long
	displayPath := filepath
	if len(filepath) > 25 {
		displayPath = "..." + filepath[len(filepath)-22:]
	}

	rp.pathLabel.SetText(displayPath)
	rp.sizeLabel.SetText(fmt.Sprintf("%dx%d", width, height))
	rp.channelsLabel.SetText(fmt.Sprintf("%d", channels))

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

	rp.pathLabel.SetText("No path")
	rp.sizeLabel.SetText("No size")
	rp.channelsLabel.SetText("No channels")

	rp.fileLabel.SetText("No image loaded")
	rp.stateLabel.SetText("Ready")
	rp.lastActionLabel.SetText("Application ready")
}

func (rp *RightPanel) GetContainer() fyne.CanvasObject {
	return rp.container
}
