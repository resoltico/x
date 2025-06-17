package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type RightPanel struct {
	container *container.Scroll // This should be *container.Scroll, not *fyne.Container
	logger    *slog.Logger

	// Status section
	statusCard      *widget.Card
	stateLabel      *widget.Label
	lastActionLabel *widget.Label

	// Image information section
	imageInfoCard *widget.Card
	sizeLabel     *widget.Label
	channelsLabel *widget.Label

	// Quality metrics section
	qualityCard *widget.Card
	psnrLabel   *widget.Label
	ssimLabel   *widget.Label
	psnrBar     *widget.ProgressBar
	ssimBar     *widget.ProgressBar

	// Tools section
	toolsCard *widget.Card

	// Callbacks
	onWindowTitleChange func(title string)
}

func NewRightPanel(logger *slog.Logger) *RightPanel {
	rp := &RightPanel{
		logger: logger,
	}

	rp.createStatusSection()
	rp.createImageInfoSection()
	rp.createQualitySection()
	rp.createToolsSection()

	// Create main container with all sections
	mainContent := container.NewVBox(
		rp.statusCard,
		rp.imageInfoCard,
		rp.qualityCard,
		rp.toolsCard,
	)

	// Add scrolling for overflow content
	rp.container = container.NewScroll(mainContent)

	return rp
}

func (rp *RightPanel) createStatusSection() {
	rp.stateLabel = widget.NewLabel("Ready")
	rp.stateLabel.Importance = widget.MediumImportance

	rp.lastActionLabel = widget.NewLabel("Application started")
	rp.lastActionLabel.Wrapping = fyne.TextWrapWord

	statusContent := container.NewVBox(
		container.NewVBox(
			widget.NewLabelWithStyle("State:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			rp.stateLabel,
		),
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabelWithStyle("Last:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			rp.lastActionLabel,
		),
	)

	rp.statusCard = widget.NewCard("STATUS", "", statusContent)
}

func (rp *RightPanel) createImageInfoSection() {
	rp.sizeLabel = widget.NewLabel("No size")
	rp.channelsLabel = widget.NewLabel("No channels")

	imageInfoContent := container.NewVBox(
		container.NewVBox(
			widget.NewLabelWithStyle("Size:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			rp.sizeLabel,
		),
		widget.NewSeparator(),
		container.NewVBox(
			widget.NewLabelWithStyle("Channels:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			rp.channelsLabel,
		),
	)

	rp.imageInfoCard = widget.NewCard("IMAGE INFORMATION", "", imageInfoContent)
}

func (rp *RightPanel) createQualitySection() {
	rp.psnrLabel = widget.NewLabel("PSNR: --")
	rp.ssimLabel = widget.NewLabel("SSIM: --")

	rp.psnrBar = widget.NewProgressBar()
	rp.psnrBar.SetValue(0)

	rp.ssimBar = widget.NewProgressBar()
	rp.ssimBar.SetValue(0)

	qualityContent := container.NewVBox(
		rp.psnrLabel,
		rp.psnrBar,
		widget.NewSeparator(),
		rp.ssimLabel,
		rp.ssimBar,
		widget.NewSeparator(),
		widget.NewLabel("Real-time quality metrics will appear here during processing"),
	)

	rp.qualityCard = widget.NewCard("QUALITY METRICS", "", qualityContent)
}

func (rp *RightPanel) createToolsSection() {
	toolsContent := container.NewVBox(
		widget.NewLabel("Additional tools and options will appear here"),
	)

	rp.toolsCard = widget.NewCard("TOOLS", "", toolsContent)
}

func (rp *RightPanel) ShowImageInfo(filepath string, width, height, channels int) {
	rp.sizeLabel.SetText(fmt.Sprintf("%dx%d", width, height))
	rp.channelsLabel.SetText(fmt.Sprintf("%d", channels))

	// Update status
	rp.stateLabel.SetText("Image Loaded")
	rp.lastActionLabel.SetText("Image loaded successfully")

	// Extract filename and update window title
	filename := filepath
	lastSlash := -1
	for i := len(filename) - 1; i >= 0; i-- {
		if filename[i] == '/' || filename[i] == '\\' {
			lastSlash = i
			break
		}
	}
	if lastSlash >= 0 {
		filename = filename[lastSlash+1:]
	}

	newTitle := fmt.Sprintf("Image Restoration Suite - %s", filename)
	rp.logger.Debug("Updating window title", "new_title", newTitle)

	// Use the callback if it exists
	if rp.onWindowTitleChange != nil {
		rp.onWindowTitleChange(newTitle)
	}
}

func (rp *RightPanel) UpdateMetrics(psnr, ssim float64) {
	rp.psnrLabel.SetText(fmt.Sprintf("PSNR: %.2f dB", psnr))
	rp.ssimLabel.SetText(fmt.Sprintf("SSIM: %.4f", ssim))

	// Update progress bars (assuming PSNR 0-40 range, SSIM 0-1 range)
	psnrProgress := psnr / 40.0
	if psnrProgress > 1.0 {
		psnrProgress = 1.0
	}
	if psnrProgress < 0.0 {
		psnrProgress = 0.0
	}

	rp.psnrBar.SetValue(psnrProgress)
	rp.ssimBar.SetValue(ssim)

	rp.logger.Debug("Metrics updated", "psnr", psnr, "ssim", ssim)
}

func (rp *RightPanel) UpdateStatus(state, lastAction string) {
	rp.stateLabel.SetText(state)
	rp.lastActionLabel.SetText(lastAction)
}

func (rp *RightPanel) ShowError(message string) {
	rp.stateLabel.SetText("Error")
	rp.lastActionLabel.SetText(message)
	rp.logger.Error("Error displayed", "message", message)
}

func (rp *RightPanel) ShowMessage(message string) {
	rp.lastActionLabel.SetText(message)
	rp.logger.Debug("Message displayed", "message", message)
}

func (rp *RightPanel) Clear() {
	rp.sizeLabel.SetText("No size")
	rp.channelsLabel.SetText("No channels")
	rp.stateLabel.SetText("Ready")
	rp.lastActionLabel.SetText("Application started")
	rp.psnrLabel.SetText("PSNR: --")
	rp.ssimLabel.SetText("SSIM: --")
	rp.psnrBar.SetValue(0)
	rp.ssimBar.SetValue(0)
	rp.logger.Debug("Right panel cleared")
}

func (rp *RightPanel) SetWindowTitleChangeCallback(callback func(string)) {
	rp.onWindowTitleChange = callback
}

func (rp *RightPanel) GetContainer() *container.Scroll {
	return rp.container
}
