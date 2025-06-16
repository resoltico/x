// Enhanced Properties Panel with algorithm selection
package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"

	"advanced-image-processing/internal/algorithms"
	"advanced-image-processing/internal/core"
)

// EnhancedPropertiesPanel provides algorithm selection and parameter adjustment
type EnhancedPropertiesPanel struct {
	pipeline *core.ProcessingPipeline
	logger   *logrus.Logger

	vbox            *fyne.Container
	algorithmSelect *widget.Select
	paramContainer  *fyne.Container
	progressBar     *widget.ProgressBar
	statusLabel     *widget.Label
	enabled         bool

	currentAlgorithm string
	paramWidgets     map[string]fyne.CanvasObject
}

// NewEnhancedPropertiesPanel creates a new enhanced properties panel
func NewEnhancedPropertiesPanel(pipeline *core.ProcessingPipeline, logger *logrus.Logger) *EnhancedPropertiesPanel {
	panel := &EnhancedPropertiesPanel{
		pipeline:     pipeline,
		logger:       logger,
		enabled:      false,
		paramWidgets: make(map[string]fyne.CanvasObject),
	}

	panel.initializeUI()
	return panel
}

func (pp *EnhancedPropertiesPanel) initializeUI() {
	// Algorithm selection dropdown
	categories := algorithms.GetAlgorithmsByCategory()
	var algorithmOptions []string

	for category, algs := range categories {
		for _, alg := range algs {
			algorithmOptions = append(algorithmOptions, fmt.Sprintf("%s - %s", category, alg))
		}
	}

	pp.algorithmSelect = widget.NewSelect(algorithmOptions, pp.onAlgorithmSelected)
	pp.algorithmSelect.PlaceHolder = "Select an algorithm..."
	pp.algorithmSelect.Disable()

	// Parameter container
	pp.paramContainer = container.NewVBox()

	// Progress indicators
	pp.progressBar = widget.NewProgressBar()
	pp.progressBar.Hide()

	pp.statusLabel = widget.NewLabel("")
	pp.statusLabel.Hide()

	// Main container
	content := container.NewVBox(
		widget.NewLabel("Algorithm Selection"),
		pp.algorithmSelect,
		widget.NewSeparator(),
		widget.NewLabel("Parameters"),
		pp.paramContainer,
		widget.NewSeparator(),
		pp.progressBar,
		pp.statusLabel,
	)

	pp.vbox = container.NewVBox(
		widget.NewCard("Algorithm Properties", "", content),
	)
}

func (pp *EnhancedPropertiesPanel) onAlgorithmSelected(selected string) {
	if selected == "" {
		return
	}

	// Extract algorithm name from "Category - algorithm" format
	var algorithmName string
	categories := algorithms.GetAlgorithmsByCategory()

	for _, algs := range categories {
		for _, alg := range algs {
			if selected == fmt.Sprintf("%s - %s", pp.getCategoryForAlgorithm(alg), alg) {
				algorithmName = alg
				break
			}
		}
	}

	if algorithmName == "" {
		return
	}

	pp.currentAlgorithm = algorithmName
	pp.createParameterWidgets(algorithmName)
}

func (pp *EnhancedPropertiesPanel) getCategoryForAlgorithm(algorithm string) string {
	categories := algorithms.GetAlgorithmsByCategory()

	for category, algs := range categories {
		for _, alg := range algs {
			if alg == algorithm {
				return category
			}
		}
	}
	return "Unknown"
}

func (pp *EnhancedPropertiesPanel) createParameterWidgets(algorithmName string) {
	algorithm, exists := algorithms.Get(algorithmName)
	if !exists {
		pp.logger.WithField("algorithm", algorithmName).Error("Algorithm not found")
		return
	}

	// Clear existing parameters
	pp.paramContainer.RemoveAll()
	pp.paramWidgets = make(map[string]fyne.CanvasObject)

	// Get parameter info
	paramInfo, ok := algorithm.(algorithms.AlgorithmInfo)
	if !ok {
		pp.paramContainer.Add(widget.NewLabel("No parameters available"))
		return
	}

	parameters := paramInfo.GetParameterInfo()
	if len(parameters) == 0 {
		pp.paramContainer.Add(widget.NewLabel("No parameters available"))
		return
	}

	// Create widgets for each parameter
	for _, param := range parameters {
		pp.createParameterWidget(param)
	}

	// Add apply button
	applyBtn := widget.NewButton("Apply Algorithm", func() {
		pp.applyAlgorithm()
	})
	pp.paramContainer.Add(applyBtn)
}

func (pp *EnhancedPropertiesPanel) createParameterWidget(param algorithms.ParameterInfo) {
	label := widget.NewLabel(fmt.Sprintf("%s:", param.Name))

	var paramWidget fyne.CanvasObject

	switch param.Type {
	case "int":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		slider.SetValue(param.Default.(float64))
		slider.Step = 1
		valueLabel := widget.NewLabel(fmt.Sprintf("%.0f", param.Default.(float64)))
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.0f", value))
		}
		paramWidget = container.NewHBox(slider, valueLabel)

	case "float":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		slider.SetValue(param.Default.(float64))
		slider.Step = 0.1
		valueLabel := widget.NewLabel(fmt.Sprintf("%.2f", param.Default.(float64)))
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.2f", value))
		}
		paramWidget = container.NewHBox(slider, valueLabel)

	case "bool":
		check := widget.NewCheck("", nil)
		if defaultVal, ok := param.Default.(bool); ok {
			check.SetChecked(defaultVal)
		}
		paramWidget = check

	default:
		paramWidget = widget.NewLabel("Unsupported parameter type")
	}

	pp.paramWidgets[param.Name] = paramWidget

	paramBox := container.NewVBox(
		label,
		paramWidget,
		widget.NewLabel(param.Description),
		widget.NewSeparator(),
	)

	pp.paramContainer.Add(paramBox)
}

func (pp *EnhancedPropertiesPanel) applyAlgorithm() {
	if pp.currentAlgorithm == "" {
		return
	}

	// Collect parameter values
	params := make(map[string]interface{})

	algorithm, exists := algorithms.Get(pp.currentAlgorithm)
	if !exists {
		return
	}

	paramInfo, ok := algorithm.(algorithms.AlgorithmInfo)
	if ok {
		parameters := paramInfo.GetParameterInfo()

		for _, param := range parameters {
			paramWidget, exists := pp.paramWidgets[param.Name]
			if !exists {
				continue
			}

			switch param.Type {
			case "int", "float":
				if hbox, ok := paramWidget.(*fyne.Container); ok && len(hbox.Objects) > 0 {
					if slider, ok := hbox.Objects[0].(*widget.Slider); ok {
						params[param.Name] = slider.Value
					}
				}
			case "bool":
				if check, ok := paramWidget.(*widget.Check); ok {
					params[param.Name] = check.Checked
				}
			}
		}
	}

	// Add algorithm to pipeline
	err := pp.pipeline.AddStep(pp.currentAlgorithm, params)
	if err != nil {
		pp.logger.WithError(err).Error("Failed to add algorithm step")
	}
}

func (pp *EnhancedPropertiesPanel) GetContainer() fyne.CanvasObject {
	return pp.vbox
}

func (pp *EnhancedPropertiesPanel) Enable() {
	pp.enabled = true
	pp.algorithmSelect.Enable()
}

func (pp *EnhancedPropertiesPanel) Disable() {
	pp.enabled = false
	pp.algorithmSelect.Disable()
}

func (pp *EnhancedPropertiesPanel) UpdateProgress(step, total int, stepName string) {
	pp.progressBar.Show()
	pp.statusLabel.Show()

	if total > 0 {
		pp.progressBar.SetValue(float64(step) / float64(total))
	}
	pp.statusLabel.SetText(fmt.Sprintf("Step %d/%d: %s", step, total, stepName))

	pp.logger.WithFields(logrus.Fields{
		"step":      step,
		"total":     total,
		"step_name": stepName,
	}).Debug("Processing progress")
}

func (pp *EnhancedPropertiesPanel) ClearProgress() {
	pp.progressBar.Hide()
	pp.statusLabel.Hide()
}

func (pp *EnhancedPropertiesPanel) Refresh() {
	pp.vbox.Refresh()
}
