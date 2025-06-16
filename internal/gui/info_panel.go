// internal/gui/info_panel.go
// Modern info panel with metrics and additional tools
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// InfoPanel provides the right panel with metrics and analysis tools
type InfoPanel struct {
	logger *slog.Logger

	container *fyne.Container

	// Quality metrics section
	metricsCard    *widget.Card
	metricsContent *fyne.Container
	currentMetrics map[string]float64

	// Additional tools section
	toolsCard    *widget.Card
	toolsContent *fyne.Container

	// Histogram (placeholder for future)
	histogramCard    *widget.Card
	histogramContent *fyne.Container
}

func NewInfoPanel(logger *slog.Logger) *InfoPanel {
	panel := &InfoPanel{
		logger:         logger,
		currentMetrics: make(map[string]float64),
	}

	panel.initializeUI()
	return panel
}

func (ip *InfoPanel) initializeUI() {
	// Quality Metrics section
	ip.metricsContent = container.NewVBox(
		widget.NewLabel("Real-time quality metrics will appear here during processing."),
	)
	ip.metricsCard = widget.NewCard("📊 Quality Metrics", "", ip.metricsContent)

	// Additional Tools section
	ip.toolsContent = container.NewVBox(
		widget.NewLabel("🎨 Color Picker"),
		widget.NewButton("Sample Color", func() {
			// TODO: Implement color picker
		}),
		widget.NewSeparator(),
		widget.NewLabel("📏 Measurements"),
		widget.NewButton("Measure Distance", func() {
			// TODO: Implement measurement tool
		}),
		widget.NewSeparator(),
		widget.NewLabel("💾 Export Options"),
		widget.NewButton("Export Regions", func() {
			// TODO: Implement region export
		}),
	)
	ip.toolsCard = widget.NewCard("🛠️ Additional Tools", "", ip.toolsContent)

	// Histogram section (placeholder)
	ip.histogramContent = container.NewVBox(
		widget.NewLabel("📈 Histogram"),
		widget.NewLabel("Image histogram will appear here when an image is loaded."),
		widget.NewSeparator(),
		widget.NewCheck("Show RGB channels", nil),
		widget.NewCheck("Show intensity", nil),
	)
	ip.histogramCard = widget.NewCard("📊 Analysis", "", ip.histogramContent)

	// Main container with fixed width and scrollable content
	scrollContent := container.NewVBox(
		ip.metricsCard,
		widget.NewSeparator(),
		ip.toolsCard,
		widget.NewSeparator(),
		ip.histogramCard,
	)

	scroll := container.NewScroll(scrollContent)
	scroll.SetMinSize(fyne.NewSize(300, 950))

	ip.container = container.NewBorder(nil, nil, nil, nil, scroll)
}

func (ip *InfoPanel) GetContainer() fyne.CanvasObject {
	return ip.container
}

func (ip *InfoPanel) UpdateMetrics(metrics map[string]float64) {
	ip.currentMetrics = metrics
	ip.refreshMetricsDisplay()
}

func (ip *InfoPanel) refreshMetricsDisplay() {
	ip.metricsContent.RemoveAll()

	if len(ip.currentMetrics) == 0 {
		ip.metricsContent.Add(widget.NewLabel("🔄 Processing..."))
		ip.metricsContent.Refresh()
		return
	}

	// Create modern metric displays with color coding
	for name, value := range ip.currentMetrics {
		metricWidget := ip.createMetricWidget(name, value)
		ip.metricsContent.Add(metricWidget)
	}

	ip.metricsContent.Refresh()
}

func (ip *InfoPanel) createMetricWidget(name string, value float64) fyne.CanvasObject {
	var displayText string
	var qualityText string
	var qualityColor fyne.Resource

	switch name {
	case "psnr":
		displayText = fmt.Sprintf("📡 PSNR: %.2f dB", value)
		if value > 40 {
			qualityText = "Excellent"
			qualityColor = theme.ConfirmIcon()
		} else if value > 30 {
			qualityText = "Good"
			qualityColor = theme.InfoIcon()
		} else if value > 20 {
			qualityText = "Fair"
			qualityColor = theme.WarningIcon()
		} else {
			qualityText = "Poor"
			qualityColor = theme.ErrorIcon()
		}

	case "ssim":
		displayText = fmt.Sprintf("📈 SSIM: %.3f", value)
		if value > 0.95 {
			qualityText = "Excellent"
			qualityColor = theme.ConfirmIcon()
		} else if value > 0.8 {
			qualityText = "Good"
			qualityColor = theme.InfoIcon()
		} else if value > 0.6 {
			qualityText = "Fair"
			qualityColor = theme.WarningIcon()
		} else {
			qualityText = "Poor"
			qualityColor = theme.ErrorIcon()
		}

	case "mse":
		displayText = fmt.Sprintf("📊 MSE: %.2f", value)
		if value < 100 {
			qualityText = "Excellent"
			qualityColor = theme.ConfirmIcon()
		} else if value < 500 {
			qualityText = "Good"
			qualityColor = theme.InfoIcon()
		} else if value < 1000 {
			qualityText = "Fair"
			qualityColor = theme.WarningIcon()
		} else {
			qualityText = "Poor"
			qualityColor = theme.ErrorIcon()
		}

	default:
		displayText = fmt.Sprintf("📈 %s: %.3f", name, value)
		qualityText = ""
		qualityColor = theme.InfoIcon()
	}

	// Create modern metric card
	metricLabel := widget.NewLabel(displayText)
	
	if qualityText != "" {
		qualityIcon := widget.NewIcon(qualityColor)
		qualityLabel := widget.NewLabel(qualityText)
		
		qualityRow := container.NewHBox(
			qualityIcon,
			qualityLabel,
		)
		
		return container.NewVBox(
			metricLabel,
			qualityRow,
			widget.NewSeparator(),
		)
	}

	return container.NewVBox(
		metricLabel,
		widget.NewSeparator(),
	)
}

func (ip *InfoPanel) Clear() {
	ip.currentMetrics = make(map[string]float64)
	ip.metricsContent.RemoveAll()
	ip.metricsContent.Add(widget.NewLabel("Real-time quality metrics will appear here during processing."))
	ip.metricsContent.Refresh()
}

func (ip *InfoPanel) ShowImageInfo(filepath string, width, height, channels int) {
	// Update tools section with image information
	ip.toolsContent.RemoveAll()
	
	// Image information
	ip.toolsContent.Add(widget.NewLabel("📄 Image Information"))
	ip.toolsContent.Add(widget.NewLabel(fmt.Sprintf("Path: %s", filepath)))
	ip.toolsContent.Add(widget.NewLabel(fmt.Sprintf("Size: %dx%d", width, height)))
	ip.toolsContent.Add(widget.NewLabel(fmt.Sprintf("Channels: %d", channels)))
	ip.toolsContent.Add(widget.NewSeparator())
	
	// Color picker tool
	ip.toolsContent.Add(widget.NewLabel("🎨 Color Picker"))
	colorPickerBtn := widget.NewButton("Sample Color", func() {
		// TODO: Implement color picker functionality
		ip.logger.Info("Color picker activated")
	})
	ip.toolsContent.Add(colorPickerBtn)
	ip.toolsContent.Add(widget.NewSeparator())
	
	// Measurement tools
	ip.toolsContent.Add(widget.NewLabel("📏 Measurements"))
	measureBtn := widget.NewButton("Measure Distance", func() {
		// TODO: Implement measurement tool
		ip.logger.Info("Measurement tool activated")
	})
	ip.toolsContent.Add(measureBtn)
	ip.toolsContent.Add(widget.NewSeparator())
	
	// Export options
	ip.toolsContent.Add(widget.NewLabel("💾 Export Options"))
	exportRegionsBtn := widget.NewButton("Export Regions", func() {
		// TODO: Implement region export
		ip.logger.Info("Region export requested")
	})
	exportMetricsBtn := widget.NewButton("Export Metrics", func() {
		// TODO: Implement metrics export
		ip.logger.Info("Metrics export requested")
	})
	ip.toolsContent.Add(exportRegionsBtn)
	ip.toolsContent.Add(exportMetricsBtn)
	
	ip.toolsContent.Refresh()
}

// StatusManager handles status messages and notifications
type StatusManager struct {
	widget    *widget.Card
	label     *widget.Label
	container *fyne.Container
}

func NewStatusManager() *StatusManager {
	manager := &StatusManager{}
	manager.initializeUI()
	return manager
}

func (sm *StatusManager) initializeUI() {
	sm.label = widget.NewLabel("Application ready")
	sm.container = container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		sm.label,
	)
	sm.widget = widget.NewCard("", "", sm.container)
}

func (sm *StatusManager) GetWidget() fyne.CanvasObject {
	return sm.widget
}

func (sm *StatusManager) ShowInfo(message string) {
	sm.updateStatus(message, theme.InfoIcon())
}

func (sm *StatusManager) ShowSuccess(message string) {
	sm.updateStatus(message, theme.ConfirmIcon())
}

func (sm *StatusManager) ShowWarning(message string) {
	sm.updateStatus(message, theme.WarningIcon())
}

func (sm *StatusManager) ShowError(err error) {
	sm.updateStatus(fmt.Sprintf("Error: %s", err.Error()), theme.ErrorIcon())
}

func (sm *StatusManager) updateStatus(message string, icon fyne.Resource) {
	sm.container.RemoveAll()
	sm.container.Add(widget.NewIcon(icon))
	sm.container.Add(widget.NewLabel(message))
	sm.container.Refresh()
}
