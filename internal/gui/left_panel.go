// internal/gui/left_panel.go
// Perfect UI Left Panel: Layers and Parameters (300px wide)
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/algorithms"
	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/layers"
)

type LeftPanel struct {
	pipeline      *core.EnhancedPipeline
	regionManager *core.RegionManager
	imageData     *core.ImageData
	logger        *slog.Logger

	container *fyne.Container

	// Layers section
	layersCard      *widget.Card
	layersList      *widget.List
	layersData      []*layers.Layer
	addLayerBtn     *widget.Button
	algorithmSelect *widget.Select

	// Parameters section
	parametersCard      *widget.Card
	parametersContainer *container.Scroll
	currentParams       *fyne.Container
	selectedLayerID     string

	// State
	enabled bool

	// Callbacks
	onParameterChanged func(string, map[string]interface{})
}

func NewLeftPanel(pipeline *core.EnhancedPipeline, regionManager *core.RegionManager,
	imageData *core.ImageData, logger *slog.Logger) *LeftPanel {

	panel := &LeftPanel{
		pipeline:      pipeline,
		regionManager: regionManager,
		imageData:     imageData,
		logger:        logger,
		enabled:       false,
		layersData:    make([]*layers.Layer, 0),
	}

	panel.initializeUI()
	return panel
}

func (lp *LeftPanel) initializeUI() {
	// Layers section
	lp.createLayersSection()

	// Parameters section
	lp.createParametersSection()

	// Main container with proper spacing and fixed width (300px)
	content := container.NewVBox(
		lp.layersCard,
		lp.parametersCard,
	)

	// Create scroll container
	scroll := container.NewScroll(content)
	lp.container = container.NewBorder(nil, nil, nil, nil, scroll)
	lp.container.Resize(fyne.NewSize(300, 950)) // 950px available height (1000 - 50 toolbar)

	lp.Disable()
}

func (lp *LeftPanel) createLayersSection() {
	// Layers list
	lp.layersList = widget.NewList(
		func() int {
			return len(lp.layersData)
		},
		func() fyne.CanvasObject {
			// Layer entry template: Name | Visibility Toggle | Delete Button
			nameLabel := widget.NewLabel("Layer Name")
			nameLabel.Resize(fyne.NewSize(150, 30))

			visibilityBtn := widget.NewButtonWithIcon("", theme.VisibilityIcon(), func() {})
			visibilityBtn.Resize(fyne.NewSize(30, 30))

			deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {})
			deleteBtn.Resize(fyne.NewSize(30, 30))
			deleteBtn.Importance = widget.DangerImportance

			return container.NewHBox(
				nameLabel,
				container.NewBorder(nil, nil, nil, container.NewHBox(visibilityBtn, deleteBtn), nil),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id < len(lp.layersData) {
				layer := lp.layersData[id]
				hbox := item.(*fyne.Container)

				nameLabel := hbox.Objects[0].(*widget.Label)
				nameLabel.SetText(layer.Name)

				rightContainer := hbox.Objects[1].(*fyne.Container)
				rightBox := rightContainer.Objects[0].(*fyne.Container) // Right side is actually in Objects[0] for border containers

				visibilityBtn := rightBox.Objects[0].(*widget.Button)
				deleteBtn := rightBox.Objects[1].(*widget.Button)

				// Update visibility icon based on layer state
				if layer.Enabled {
					visibilityBtn.SetIcon(theme.VisibilityIcon())
				} else {
					visibilityBtn.SetIcon(theme.VisibilityOffIcon())
				}

				// Set button callbacks
				visibilityBtn.OnTapped = func() {
					lp.toggleLayerVisibility(layer.ID)
				}

				deleteBtn.OnTapped = func() {
					lp.deleteLayer(layer.ID)
				}
			}
		},
	)

	lp.layersList.OnSelected = func(id widget.ListItemID) {
		if id < len(lp.layersData) {
			lp.selectLayer(lp.layersData[id].ID)
		}
	}

	lp.layersList.Resize(fyne.NewSize(280, 200)) // Reduced height to leave more space

	// Algorithm selection dropdown
	lp.algorithmSelect = widget.NewSelect(lp.getAlgorithmOptions(), nil)
	lp.algorithmSelect.PlaceHolder = "Select algorithm..."
	lp.algorithmSelect.Resize(fyne.NewSize(280, 30))

	// Add Layer button
	lp.addLayerBtn = widget.NewButtonWithIcon("+ Add Layer", theme.ContentAddIcon(), lp.addLayer)
	lp.addLayerBtn.Importance = widget.HighImportance
	lp.addLayerBtn.Resize(fyne.NewSize(120, 30))

	layersContent := container.NewVBox(
		lp.layersList,
		lp.algorithmSelect,
		lp.addLayerBtn,
	)

	lp.layersCard = widget.NewCard("LAYERS", "", layersContent)
}

func (lp *LeftPanel) createParametersSection() {
	lp.currentParams = container.NewVBox(
		widget.NewLabel("Select a layer to edit parameters"),
	)
	lp.parametersContainer = container.NewScroll(lp.currentParams)
	lp.parametersContainer.Resize(fyne.NewSize(280, 500)) // Increased height for parameters

	lp.parametersCard = widget.NewCard("PARAMETERS", "", lp.parametersContainer)
}

func (lp *LeftPanel) getAlgorithmOptions() []string {
	categories := algorithms.GetAlgorithmsByCategory()
	var options []string

	// Add algorithms organized by category
	for category, algs := range categories {
		options = append(options, fmt.Sprintf("--- %s ---", category))
		options = append(options, algs...)
	}

	return options
}

func (lp *LeftPanel) addLayer() {
	selected := lp.algorithmSelect.Selected
	if selected == "" || len(selected) > 3 && selected[:3] == "---" {
		lp.showSelectionDialog()
		return
	}

	algorithm, exists := algorithms.Get(selected)
	if !exists {
		lp.logger.Error("Algorithm not found", "algorithm", selected)
		return
	}

	params := algorithm.GetDefaultParams()
	layerID, err := lp.pipeline.AddLayer(selected, selected, params, "")
	if err != nil {
		lp.logger.Error("Failed to add layer", "error", err)
		return
	}

	lp.refreshLayersList()
	lp.algorithmSelect.ClearSelected()
	lp.selectLayer(layerID)

	lp.logger.Info("Layer added", "layer_id", layerID, "algorithm", selected)
}

func (lp *LeftPanel) deleteLayer(layerID string) {
	confirmDialog := dialog.NewConfirm("Delete Layer",
		"Are you sure you want to delete this layer?",
		func(confirmed bool) {
			if confirmed {
				lp.removeLayerFromPipeline(layerID)
				lp.refreshLayersList()
				if lp.selectedLayerID == layerID {
					lp.clearParameters()
					lp.selectedLayerID = ""
				}
			}
		},
		fyne.CurrentApp().Driver().AllWindows()[0])
	confirmDialog.Show()
}

func (lp *LeftPanel) toggleLayerVisibility(layerID string) {
	// Find and toggle layer in pipeline
	layers := lp.pipeline.GetLayers()
	for _, layer := range layers {
		if layer.ID == layerID {
			layer.Enabled = !layer.Enabled
			break
		}
	}
	lp.refreshLayersList()
	lp.TriggerPreviewUpdate()
}

func (lp *LeftPanel) removeLayerFromPipeline(layerID string) {
	// This would require adding a RemoveLayer method to the pipeline
	// For now, we'll just refresh the display
	lp.refreshLayersList()
	lp.TriggerPreviewUpdate()
}

func (lp *LeftPanel) selectLayer(layerID string) {
	lp.selectedLayerID = layerID
	lp.updateParametersForLayer(layerID)
}

func (lp *LeftPanel) refreshLayersList() {
	lp.layersData = lp.pipeline.GetLayers()
	lp.layersList.Refresh()
}

func (lp *LeftPanel) updateParametersForLayer(layerID string) {
	// Clear existing parameters
	lp.currentParams.RemoveAll()

	if layerID == "" {
		lp.currentParams.Add(widget.NewLabel("Select a layer to edit parameters"))
		lp.currentParams.Refresh()
		return
	}

	// Find the layer
	var targetLayer *layers.Layer
	for _, layer := range lp.layersData {
		if layer.ID == layerID {
			targetLayer = layer
			break
		}
	}

	if targetLayer == nil {
		lp.currentParams.Add(widget.NewLabel("Layer not found"))
		lp.currentParams.Refresh()
		return
	}

	// Get algorithm and create parameter widgets
	algorithm, exists := algorithms.Get(targetLayer.Algorithm)
	if !exists {
		lp.currentParams.Add(widget.NewLabel("Algorithm not found"))
		lp.currentParams.Refresh()
		return
	}

	// Create parameter controls
	paramInfo := algorithm.GetParameterInfo()

	for _, param := range paramInfo {
		lp.createParameterWidget(param, targetLayer.Parameters, layerID)
	}

	lp.currentParams.Refresh()
}

func (lp *LeftPanel) createParameterWidget(param algorithms.ParameterInfo, params map[string]interface{}, layerID string) {
	// Parameter label
	label := widget.NewLabelWithStyle(param.Name+":", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
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
			lp.onParameterChange(layerID, params)
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
			lp.onParameterChange(layerID, params)
		}

		paramContainer := container.NewVBox(
			container.NewHBox(slider, valueLabel),
		)
		lp.currentParams.Add(paramContainer)

	case "bool":
		check := widget.NewCheck("", func(checked bool) {
			params[param.Name] = checked
			lp.onParameterChange(layerID, params)
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
			lp.onParameterChange(layerID, params)
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
		descLabel.TextStyle = fyne.TextStyle{Italic: true}
		lp.currentParams.Add(descLabel)
	}

	// Add separator
	lp.currentParams.Add(widget.NewSeparator())
}

func (lp *LeftPanel) onParameterChange(layerID string, params map[string]interface{}) {
	if lp.onParameterChanged != nil {
		lp.onParameterChanged(layerID, params)
	}
	lp.TriggerPreviewUpdate()
}

func (lp *LeftPanel) TriggerPreviewUpdate() {
	// Trigger real-time preview update via pipeline
	// The pipeline handles debounced processing automatically
}

func (lp *LeftPanel) showSelectionDialog() {
	content := widget.NewLabel("Please select an algorithm from the dropdown first.")
	dialog := dialog.NewInformation("No Algorithm Selected", content.Text, fyne.CurrentApp().Driver().AllWindows()[0])
	dialog.Show()
}

func (lp *LeftPanel) clearParameters() {
	lp.currentParams.RemoveAll()
	lp.currentParams.Add(widget.NewLabel("Select a layer to edit parameters"))
	lp.currentParams.Refresh()
}

func (lp *LeftPanel) GetContainer() fyne.CanvasObject {
	return lp.container
}

func (lp *LeftPanel) EnableProcessing() {
	lp.enabled = true
	lp.addLayerBtn.Enable()
	lp.algorithmSelect.Enable()
}

func (lp *LeftPanel) Disable() {
	lp.enabled = false
	lp.addLayerBtn.Disable()
	lp.algorithmSelect.Disable()
}

func (lp *LeftPanel) Reset() {
	lp.layersData = make([]*layers.Layer, 0)
	lp.layersList.Refresh()
	lp.algorithmSelect.ClearSelected()
	lp.selectedLayerID = ""
	lp.clearParameters()
}

func (lp *LeftPanel) UpdateSelectionState(hasSelection bool) {
	// Update UI based on selection state for future ROI features
}

func (lp *LeftPanel) SetCallbacks(onParameterChanged func(string, map[string]interface{})) {
	lp.onParameterChanged = onParameterChanged
}
