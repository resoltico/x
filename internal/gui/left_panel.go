// internal/gui/left_panel.go
// Perfect UI Left Panel: Controls and Parameters (250px wide)
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/algorithms"
	"advanced-image-processing/internal/core"
)

type LeftPanel struct {
	pipeline      *core.EnhancedPipeline
	regionManager *core.RegionManager
	imageData     *core.ImageData
	logger        *slog.Logger

	container *fyne.Container

	// Mode selection
	modeToggle *widget.RadioGroup

	// Algorithm selection
	algorithmSelect *widget.Select
	addAlgorithmBtn *widget.Button

	// Parameters section
	parametersContainer *container.Scroll
	currentParams       *fyne.Container

	// State
	enabled bool

	// Callbacks
	onModeChanged func(bool)
}

func NewLeftPanel(pipeline *core.EnhancedPipeline, regionManager *core.RegionManager,
	imageData *core.ImageData, logger *slog.Logger) *LeftPanel {

	panel := &LeftPanel{
		pipeline:      pipeline,
		regionManager: regionManager,
		imageData:     imageData,
		logger:        logger,
		enabled:       false,
	}

	panel.initializeUI()
	return panel
}

func (lp *LeftPanel) initializeUI() {
	// Mode selection section
	modeCard := lp.createModeSection()

	// Algorithm selection section
	algorithmCard := lp.createAlgorithmSection()

	// Parameters section
	parametersCard := lp.createParametersSection()

	// Main container with proper spacing
	content := container.NewVBox(
		modeCard,
		algorithmCard,
		parametersCard,
	)

	// Create scroll container and set fixed width to 250px as per specification
	scroll := container.NewScroll(content)
	lp.container = container.NewBorder(nil, nil, nil, nil, scroll)
	lp.container.Resize(fyne.NewSize(250, 850)) // 850px available height (900 - 50 toolbar)

	lp.Disable()
}

func (lp *LeftPanel) createModeSection() *widget.Card {
	lp.modeToggle = widget.NewRadioGroup(
		[]string{"Sequential Mode", "Layer Mode"},
		func(value string) {
			isLayerMode := (value == "Layer Mode")
			lp.pipeline.SetProcessingMode(isLayerMode)
			if lp.onModeChanged != nil {
				lp.onModeChanged(isLayerMode)
			}
		},
	)
	lp.modeToggle.SetSelected("Sequential Mode")
	lp.modeToggle.Horizontal = false

	return widget.NewCard("MODE", "", container.NewVBox(
		lp.modeToggle,
	))
}

func (lp *LeftPanel) createAlgorithmSection() *widget.Card {
	// Algorithm dropdown
	lp.algorithmSelect = widget.NewSelect(lp.getAlgorithmOptions(), func(selected string) {
		lp.updateParametersForAlgorithm(selected)
	})
	lp.algorithmSelect.PlaceHolder = "Select algorithm..."
	lp.algorithmSelect.Resize(fyne.NewSize(200, 30))

	// Add algorithm button
	lp.addAlgorithmBtn = widget.NewButtonWithIcon("Add Algorithm", nil, lp.addAlgorithm)
	lp.addAlgorithmBtn.Importance = widget.HighImportance
	lp.addAlgorithmBtn.Resize(fyne.NewSize(120, 30))

	return widget.NewCard("ALGORITHM", "", container.NewVBox(
		lp.algorithmSelect,
		lp.addAlgorithmBtn,
	))
}

func (lp *LeftPanel) createParametersSection() *widget.Card {
	lp.currentParams = container.NewVBox(
		widget.NewLabel("Select an algorithm to edit its parameters"),
	)
	lp.parametersContainer = container.NewScroll(lp.currentParams)

	return widget.NewCard("PARAMETERS", "", lp.parametersContainer)
}

func (lp *LeftPanel) getAlgorithmOptions() []string {
	categories := algorithms.GetAlgorithmsByCategory()
	var options []string

	// Add category headers and algorithms
	for category, algs := range categories {
		options = append(options, fmt.Sprintf("--- %s ---", category))
		options = append(options, algs...)
	}

	return options
}

func (lp *LeftPanel) updateParametersForAlgorithm(algorithmName string) {
	// Clear existing parameters
	lp.currentParams.RemoveAll()

	// Skip category headers
	if algorithmName == "" || algorithmName[:3] == "---" {
		lp.currentParams.Add(widget.NewLabel("Select an algorithm to edit its parameters"))
		lp.currentParams.Refresh()
		return
	}

	// Get algorithm and create parameter widgets
	algorithm, exists := algorithms.Get(algorithmName)
	if !exists {
		lp.currentParams.Add(widget.NewLabel("Algorithm not found"))
		lp.currentParams.Refresh()
		return
	}

	// Create parameter controls
	params := algorithm.GetDefaultParams()
	paramInfo := algorithm.GetParameterInfo()

	for _, param := range paramInfo {
		lp.createParameterWidget(param, params)
	}

	lp.currentParams.Refresh()
}

func (lp *LeftPanel) createParameterWidget(param algorithms.ParameterInfo, params map[string]interface{}) {
	// Parameter label
	label := widget.NewLabel(param.Name + ":")
	lp.currentParams.Add(label)

	switch param.Type {
	case "int":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		if val, ok := params[param.Name].(float64); ok {
			slider.SetValue(val)
		} else if defaultVal, ok := param.Default.(float64); ok {
			slider.SetValue(defaultVal)
		}
		slider.Step = 1

		valueLabel := widget.NewLabel(fmt.Sprintf("%.0f", slider.Value))
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.0f", value))
			params[param.Name] = value
		}

		paramContainer := container.NewVBox(
			container.NewHBox(slider, valueLabel),
		)
		lp.currentParams.Add(paramContainer)

	case "float":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		if val, ok := params[param.Name].(float64); ok {
			slider.SetValue(val)
		} else if defaultVal, ok := param.Default.(float64); ok {
			slider.SetValue(defaultVal)
		}
		slider.Step = 0.1

		valueLabel := widget.NewLabel(fmt.Sprintf("%.2f", slider.Value))
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.2f", value))
			params[param.Name] = value
		}

		paramContainer := container.NewVBox(
			container.NewHBox(slider, valueLabel),
		)
		lp.currentParams.Add(paramContainer)

	case "bool":
		check := widget.NewCheck("", func(checked bool) {
			params[param.Name] = checked
		})
		if val, ok := params[param.Name].(bool); ok {
			check.SetChecked(val)
		} else if defaultVal, ok := param.Default.(bool); ok {
			check.SetChecked(defaultVal)
		}
		lp.currentParams.Add(check)

	case "enum":
		selectWidget := widget.NewSelect(param.Options, func(selected string) {
			params[param.Name] = selected
		})
		if val, ok := params[param.Name].(string); ok {
			selectWidget.SetSelected(val)
		} else if defaultVal, ok := param.Default.(string); ok {
			selectWidget.SetSelected(defaultVal)
		}
		lp.currentParams.Add(selectWidget)
	}

	// Add description as small text
	if param.Description != "" {
		descLabel := widget.NewLabel(param.Description)
		lp.currentParams.Add(descLabel)
	}

	// Add separator
	lp.currentParams.Add(widget.NewSeparator())
}

func (lp *LeftPanel) addAlgorithm() {
	selected := lp.algorithmSelect.Selected
	if selected == "" || selected[:3] == "---" {
		lp.showSelectionDialog()
		return
	}

	algorithm, exists := algorithms.Get(selected)
	if !exists {
		lp.logger.Error("Algorithm not found", "algorithm", selected)
		return
	}

	params := algorithm.GetDefaultParams()
	err := lp.pipeline.AddStep(selected, params)
	if err != nil {
		lp.logger.Error("Failed to add algorithm", "error", err)
		return
	}

	lp.logger.Info("Algorithm added", "algorithm", selected)
}

func (lp *LeftPanel) showSelectionDialog() {
	content := widget.NewLabel("Please select an algorithm from the dropdown first.")
	dialog := dialog.NewInformation("No Algorithm Selected", content.Text, fyne.CurrentApp().Driver().AllWindows()[0])
	dialog.Show()
}

func (lp *LeftPanel) GetContainer() fyne.CanvasObject {
	return lp.container
}

func (lp *LeftPanel) EnableProcessing() {
	lp.enabled = true
	lp.modeToggle.Enable()
	lp.algorithmSelect.Enable()
	lp.addAlgorithmBtn.Enable()
}

func (lp *LeftPanel) Disable() {
	lp.enabled = false
	lp.modeToggle.Disable()
	lp.algorithmSelect.Disable()
	lp.addAlgorithmBtn.Disable()
}

func (lp *LeftPanel) Reset() {
	lp.algorithmSelect.ClearSelected()
	lp.currentParams.RemoveAll()
	lp.currentParams.Add(widget.NewLabel("Select an algorithm to edit its parameters"))
	lp.currentParams.Refresh()
}

func (lp *LeftPanel) UpdateSelectionState(hasSelection bool) {
	// Update UI based on selection state - placeholder for future ROI features
}

func (lp *LeftPanel) SetCallbacks(onModeChanged func(bool)) {
	lp.onModeChanged = onModeChanged
}
