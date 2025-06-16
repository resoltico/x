// Real-time Properties Panel with algorithm selection
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/algorithms"
	"advanced-image-processing/internal/core"
)

// EnhancedPropertiesPanel provides real-time algorithm selection and parameter adjustment
type EnhancedPropertiesPanel struct {
	pipeline *core.ProcessingPipeline
	logger   *slog.Logger

	vbox            *fyne.Container
	algorithmSelect *widget.Select
	paramContainer  *fyne.Container
	enabled         bool

	currentAlgorithm string
	paramWidgets     map[string]fyne.CanvasObject
}

func NewEnhancedPropertiesPanel(pipeline *core.ProcessingPipeline, logger *slog.Logger) *EnhancedPropertiesPanel {
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

	// Main container
	content := container.NewVBox(
		widget.NewLabel("Algorithm Selection"),
		pp.algorithmSelect,
		widget.NewSeparator(),
		widget.NewLabel("Parameters (Real-time)"),
		pp.paramContainer,
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

	// Automatically add algorithm with default parameters
	algorithm, exists := algorithms.Get(algorithmName)
	if exists {
		params := algorithm.GetDefaultParams()
		if err := pp.pipeline.AddStep(algorithmName, params); err != nil {
			pp.logger.Error("Failed to add algorithm step", "error", err)
		}
	}
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
		pp.logger.Error("Algorithm not found", "algorithm", algorithmName)
		return
	}

	// Clear existing parameters
	pp.paramContainer.RemoveAll()
	pp.paramWidgets = make(map[string]fyne.CanvasObject)

	// Get parameter info
	paramInfo := algorithm.GetParameterInfo()
	if len(paramInfo) == 0 {
		pp.paramContainer.Add(widget.NewLabel("No parameters available"))
		return
	}

	// Create widgets for each parameter
	for _, param := range paramInfo {
		pp.createParameterWidget(param)
	}

	// Add remove button
	removeBtn := widget.NewButton("Remove Algorithm", func() {
		// Find and remove the last step with this algorithm
		steps := pp.pipeline.GetSteps()
		for i := len(steps) - 1; i >= 0; i-- {
			if steps[i].Algorithm == algorithmName {
				pp.pipeline.RemoveStep(i)
				break
			}
		}

		// Clear the UI
		pp.algorithmSelect.SetSelected("")
		pp.paramContainer.RemoveAll()
		pp.currentAlgorithm = ""
	})
	pp.paramContainer.Add(removeBtn)
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

		// Real-time parameter update
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.0f", value))
			pp.updateAlgorithmParameter(param.Name, value)
		}
		paramWidget = container.NewHBox(slider, valueLabel)

	case "float":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		slider.SetValue(param.Default.(float64))
		slider.Step = 0.1
		valueLabel := widget.NewLabel(fmt.Sprintf("%.2f", param.Default.(float64)))

		// Real-time parameter update
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.2f", value))
			pp.updateAlgorithmParameter(param.Name, value)
		}
		paramWidget = container.NewHBox(slider, valueLabel)

	case "bool":
		check := widget.NewCheck("", func(checked bool) {
			pp.updateAlgorithmParameter(param.Name, checked)
		})
		if defaultVal, ok := param.Default.(bool); ok {
			check.SetChecked(defaultVal)
		}
		paramWidget = check

	case "enum":
		selectWidget := widget.NewSelect(param.Options, func(selected string) {
			pp.updateAlgorithmParameter(param.Name, selected)
		})
		if defaultVal, ok := param.Default.(string); ok {
			selectWidget.SetSelected(defaultVal)
		}
		paramWidget = selectWidget

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

func (pp *EnhancedPropertiesPanel) updateAlgorithmParameter(paramName string, value interface{}) {
	if pp.currentAlgorithm == "" {
		return
	}

	// Find the last step with current algorithm and update its parameters
	steps := pp.pipeline.GetSteps()
	for i := len(steps) - 1; i >= 0; i-- {
		if steps[i].Algorithm == pp.currentAlgorithm {
			// Get current parameters
			params := make(map[string]interface{})
			for k, v := range steps[i].Parameters {
				params[k] = v
			}

			// Update the changed parameter
			params[paramName] = value

			// Update the pipeline step
			if err := pp.pipeline.UpdateStep(i, params); err != nil {
				pp.logger.Error("Failed to update algorithm parameter", "error", err)
			}
			break
		}
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

func (pp *EnhancedPropertiesPanel) Refresh() {
	pp.vbox.Refresh()
}
