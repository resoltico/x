// Updated Properties Panel with improved UX and styling
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

// EnhancedPropertiesPanel provides algorithm selection for sequential mode
type EnhancedPropertiesPanel struct {
	pipeline *core.EnhancedPipeline
	logger   *slog.Logger

	vbox            *fyne.Container
	algorithmSelect *widget.Select
	paramContainer  *container.Scroll
	enabled         bool

	currentAlgorithm string
	paramWidgets     map[string]fyne.CanvasObject
}

func NewEnhancedPropertiesPanel(pipeline *core.EnhancedPipeline, logger *slog.Logger) *EnhancedPropertiesPanel {
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
	// Algorithm selection dropdown with better organization
	categories := algorithms.GetAlgorithmsByCategory()
	var algorithmOptions []string

	for category, algs := range categories {
		for _, alg := range algs {
			algorithmOptions = append(algorithmOptions, fmt.Sprintf("%s → %s", category, alg))
		}
	}

	pp.algorithmSelect = widget.NewSelect(algorithmOptions, pp.onAlgorithmSelected)
	pp.algorithmSelect.PlaceHolder = "Choose an algorithm..."
	pp.algorithmSelect.Disable()

	// Parameter container with scrolling
	paramContent := container.NewVBox()
	pp.paramContainer = container.NewVScroll(paramContent)
	pp.paramContainer.SetMinSize(fyne.NewSize(300, 200))

	// Mode explanation
	modeCard := widget.NewCard("📝 Sequential Processing", "",
		container.NewVBox(
			widget.NewLabel("Sequential mode processes the entire image step by step."),
			widget.NewLabel("Each algorithm is applied to the full image in order."),
			widget.NewSeparator(),
			widget.NewLabel("💡 Enable Layer Mode for region-specific processing."),
		))

	// Algorithm selection card
	algorithmCard := widget.NewCard("🔧 Algorithm Selection", "",
		container.NewVBox(
			widget.NewLabel("Choose an algorithm to add to the processing pipeline:"),
			pp.algorithmSelect,
		))

	// Parameters card
	parametersCard := widget.NewCard("⚙️ Real-time Parameters", "",
		container.NewVBox(
			widget.NewLabel("Adjust parameters in real-time:"),
			pp.paramContainer,
		))

	// Main container
	pp.vbox = container.NewVBox(
		modeCard,
		widget.NewSeparator(),
		algorithmCard,
		widget.NewSeparator(),
		parametersCard,
	)
}

func (pp *EnhancedPropertiesPanel) onAlgorithmSelected(selected string) {
	if selected == "" {
		return
	}

	// Extract algorithm name from "Category → algorithm" format
	var algorithmName string
	categories := algorithms.GetAlgorithmsByCategory()

	for _, algs := range categories {
		for _, alg := range algs {
			if selected == fmt.Sprintf("%s → %s", pp.getCategoryForAlgorithm(alg), alg) {
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

	// Automatically add algorithm with default parameters to sequential pipeline
	algorithm, exists := algorithms.Get(algorithmName)
	if exists {
		params := algorithm.GetDefaultParams()
		if err := pp.pipeline.AddStep(algorithmName, params); err != nil {
			pp.logger.Error("Failed to add algorithm step", "error", err)
		} else {
			pp.logger.Info("Added algorithm to sequential pipeline", "algorithm", algorithmName)
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
	if pp.paramContainer.Content != nil {
		if paramContent, ok := pp.paramContainer.Content.(*fyne.Container); ok {
			paramContent.RemoveAll()
		}
	}
	pp.paramWidgets = make(map[string]fyne.CanvasObject)

	// Get parameter info
	paramInfo := algorithm.GetParameterInfo()
	if len(paramInfo) == 0 {
		pp.addToParamContainer(widget.NewLabel("ℹ️ No configurable parameters for this algorithm"))
		return
	}

	// Add algorithm info
	pp.addToParamContainer(widget.NewLabel(fmt.Sprintf("🔧 %s", algorithm.GetName())))
	pp.addToParamContainer(widget.NewLabel(fmt.Sprintf("📝 %s", algorithm.GetDescription())))
	pp.addToParamContainer(widget.NewSeparator())

	// Create widgets for each parameter
	for _, param := range paramInfo {
		pp.createParameterWidget(param)
	}

	// Add action buttons
	pp.addToParamContainer(widget.NewSeparator())

	removeBtn := widget.NewButton("🗑️ Remove Last Algorithm", func() {
		steps := pp.pipeline.GetSteps()
		if len(steps) > 0 {
			pp.logger.Debug("Remove last algorithm requested")
		}

		// Clear the UI
		pp.algorithmSelect.SetSelected("")
		if pp.paramContainer.Content != nil {
			if paramContent, ok := pp.paramContainer.Content.(*fyne.Container); ok {
				paramContent.RemoveAll()
			}
		}
		pp.currentAlgorithm = ""
	})
	removeBtn.Importance = widget.LowImportance

	clearBtn := widget.NewButton("🗑️ Clear All Steps", func() {
		pp.pipeline.ClearAll()
		pp.algorithmSelect.SetSelected("")
		if pp.paramContainer.Content != nil {
			if paramContent, ok := pp.paramContainer.Content.(*fyne.Container); ok {
				paramContent.RemoveAll()
			}
		}
		pp.currentAlgorithm = ""
	})
	clearBtn.Importance = widget.DangerImportance

	pp.addToParamContainer(container.NewHBox(removeBtn, clearBtn))
}

func (pp *EnhancedPropertiesPanel) addToParamContainer(obj fyne.CanvasObject) {
	if pp.paramContainer.Content != nil {
		if paramContent, ok := pp.paramContainer.Content.(*fyne.Container); ok {
			paramContent.Add(obj)
		}
	}
}

func (pp *EnhancedPropertiesPanel) createParameterWidget(param algorithms.ParameterInfo) {
	label := widget.NewLabel(fmt.Sprintf("🔧 %s:", param.Name))
	pp.addToParamContainer(label)

	var paramWidget fyne.CanvasObject

	switch param.Type {
	case "int":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		slider.SetValue(param.Default.(float64))
		slider.Step = 1
		valueLabel := widget.NewLabel(fmt.Sprintf("%.0f", param.Default.(float64)))

		// Real-time parameter update for sequential mode
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.0f", value))
			pp.updateAlgorithmParameter(param.Name, value)
		}
		paramWidget = container.NewVBox(
			container.NewHBox(slider, valueLabel),
			widget.NewLabel(fmt.Sprintf("📝 %s", param.Description)),
		)

	case "float":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		slider.SetValue(param.Default.(float64))
		slider.Step = 0.1
		valueLabel := widget.NewLabel(fmt.Sprintf("%.2f", param.Default.(float64)))

		// Real-time parameter update for sequential mode
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.2f", value))
			pp.updateAlgorithmParameter(param.Name, value)
		}
		paramWidget = container.NewVBox(
			container.NewHBox(slider, valueLabel),
			widget.NewLabel(fmt.Sprintf("📝 %s", param.Description)),
		)

	case "bool":
		check := widget.NewCheck("", func(checked bool) {
			pp.updateAlgorithmParameter(param.Name, checked)
		})
		if defaultVal, ok := param.Default.(bool); ok {
			check.SetChecked(defaultVal)
		}
		paramWidget = container.NewVBox(
			check,
			widget.NewLabel(fmt.Sprintf("📝 %s", param.Description)),
		)

	case "enum":
		selectWidget := widget.NewSelect(param.Options, func(selected string) {
			pp.updateAlgorithmParameter(param.Name, selected)
		})
		if defaultVal, ok := param.Default.(string); ok {
			selectWidget.SetSelected(defaultVal)
		}
		paramWidget = container.NewVBox(
			selectWidget,
			widget.NewLabel(fmt.Sprintf("📝 %s", param.Description)),
		)

	default:
		paramWidget = widget.NewLabel("❌ Unsupported parameter type")
	}

	pp.paramWidgets[param.Name] = paramWidget
	pp.addToParamContainer(paramWidget)
	pp.addToParamContainer(widget.NewSeparator())
}

func (pp *EnhancedPropertiesPanel) updateAlgorithmParameter(paramName string, value interface{}) {
	if pp.currentAlgorithm == "" {
		return
	}

	// For sequential mode, we would need to extend the pipeline to support parameter updates
	// For now, this is a placeholder that shows the concept
	pp.logger.Debug("Parameter update requested",
		"algorithm", pp.currentAlgorithm,
		"param", paramName,
		"value", value)
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
