// internal/gui/left_panel.go
// Perfect UI Left Panel: Layers and Parameters with integrated debug system
package gui

import (
	"fmt"
	"log/slog"
	"time"

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

	// Main container
	content := container.NewVBox(
		lp.layersCard,
		lp.parametersCard,
	)

	scroll := container.NewScroll(content)
	lp.container = container.NewBorder(nil, nil, nil, nil, scroll)
	lp.container.Resize(fyne.NewSize(300, 950))

	lp.Disable()
}

func (lp *LeftPanel) createLayersSection() {
	// Layers list
	lp.layersList = widget.NewList(
		func() int {
			return len(lp.layersData)
		},
		func() fyne.CanvasObject {
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
				rightBox := rightContainer.Objects[0].(*fyne.Container)

				visibilityBtn := rightBox.Objects[0].(*widget.Button)
				deleteBtn := rightBox.Objects[1].(*widget.Button)

				if layer.Enabled {
					visibilityBtn.SetIcon(theme.VisibilityIcon())
				} else {
					visibilityBtn.SetIcon(theme.VisibilityOffIcon())
				}

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

	lp.layersList.Resize(fyne.NewSize(280, 200))

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
	lp.parametersContainer.Resize(fyne.NewSize(280, 500))

	lp.parametersCard = widget.NewCard("PARAMETERS", "", lp.parametersContainer)
}

func (lp *LeftPanel) getAlgorithmOptions() []string {
	categories := algorithms.GetAlgorithmsByCategory()
	var options []string

	for category, algs := range categories {
		options = append(options, fmt.Sprintf("--- %s ---", category))
		options = append(options, algs...)
	}

	return options
}

func (lp *LeftPanel) addLayer() {
	selected := lp.algorithmSelect.Selected
	start := time.Now()

	// DEBUG: Log attempt
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogLayerAddAttempt(selected, nil)
	}

	lp.logger.Info("LAYER: Add layer requested", "selected_algorithm", selected)

	if selected == "" || len(selected) > 3 && selected[:3] == "---" {
		lp.logger.Warn("LAYER: Invalid algorithm selection", "selected", selected)

		// DEBUG: Log failed attempt
		if GlobalGUIDebugger != nil {
			duration := time.Since(start)
			GlobalGUIDebugger.LogLayerAddComplete("", false, fmt.Errorf("invalid algorithm selection: %s", selected))
			GlobalGUIDebugger.LogLayerOperation("add", "", selected, nil, false, fmt.Errorf("invalid selection"), duration)
		}

		lp.showSelectionDialog()
		return
	}

	algorithm, exists := algorithms.Get(selected)
	if !exists {
		lp.logger.Error("LAYER: Algorithm not found", "algorithm", selected)

		// DEBUG: Log failed attempt
		if GlobalGUIDebugger != nil {
			duration := time.Since(start)
			err := fmt.Errorf("algorithm not found: %s", selected)
			GlobalGUIDebugger.LogLayerAddComplete("", false, err)
			GlobalGUIDebugger.LogLayerOperation("add", "", selected, nil, false, err, duration)
		}
		return
	}

	lp.logger.Debug("LAYER: Algorithm found, getting default params", "algorithm", selected)
	params := algorithm.GetDefaultParams()
	lp.logger.Debug("LAYER: Default params", "params", params)

	// Set pipeline to use layer mode
	lp.pipeline.SetProcessingMode(true)
	lp.logger.Info("LAYER: Set pipeline to layer mode")

	// DEBUG: Log processing mode change
	if GlobalGUIDebugger != nil {
		layerCount := len(lp.layersData)
		GlobalGUIDebugger.LogProcessingEvent("mode_change", "layer", layerCount, map[string]interface{}{
			"reason":    "adding_layer",
			"algorithm": selected,
		})
	}

	layerID, err := lp.pipeline.AddLayer(selected, selected, params, "")
	if err != nil {
		lp.logger.Error("LAYER: Failed to add layer to pipeline", "error", err, "algorithm", selected)

		// DEBUG: Log failed attempt
		if GlobalGUIDebugger != nil {
			duration := time.Since(start)
			GlobalGUIDebugger.LogLayerAddComplete("", false, err)
			GlobalGUIDebugger.LogLayerOperation("add", "", selected, params, false, err, duration)
		}
		return
	}

	lp.logger.Info("LAYER: Layer added to pipeline successfully", "layer_id", layerID, "algorithm", selected)

	// DEBUG: Log successful addition
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogLayerAddComplete(layerID, true, nil)
		GlobalGUIDebugger.LogLayerOperation("add", layerID, selected, params, true, nil, duration)
	}

	lp.refreshLayersList()
	lp.algorithmSelect.ClearSelected()
	lp.selectLayer(layerID)

	// CRITICAL: Trigger preview processing immediately
	lp.logger.Info("LAYER: Triggering immediate preview update after adding layer")

	// DEBUG: Log preview trigger
	if GlobalGUIDebugger != nil {
		layerCount := len(lp.layersData) + 1 // +1 because we just added one
		GlobalGUIDebugger.LogPreviewTrigger("layer_added", layerCount)
	}

	lp.TriggerPreviewUpdate()

	lp.logger.Info("LAYER: Add layer completed", "layer_id", layerID, "algorithm", selected)
}

func (lp *LeftPanel) deleteLayer(layerID string) {
	lp.logger.Info("LAYER: Delete layer requested", "layer_id", layerID)

	// DEBUG: Log delete attempt
	if GlobalGUIDebugger != nil {
		// Find layer info for logging
		var algorithm string
		for _, layer := range lp.layersData {
			if layer.ID == layerID {
				algorithm = layer.Algorithm
				break
			}
		}
		GlobalGUIDebugger.LogLayerOperation("delete_attempt", layerID, algorithm, nil, true, nil, 0)
	}

	confirmDialog := dialog.NewConfirm("Delete Layer",
		"Are you sure you want to delete this layer?",
		func(confirmed bool) {
			if confirmed {
				start := time.Now()
				lp.logger.Info("LAYER: Layer deletion confirmed", "layer_id", layerID)

				// Find layer info before deletion
				var algorithm string
				var params map[string]interface{}
				for _, layer := range lp.layersData {
					if layer.ID == layerID {
						algorithm = layer.Algorithm
						params = layer.Parameters
						break
					}
				}

				lp.removeLayerFromPipeline(layerID)
				lp.refreshLayersList()
				if lp.selectedLayerID == layerID {
					lp.clearParameters()
					lp.selectedLayerID = ""
				}

				// DEBUG: Log successful deletion
				if GlobalGUIDebugger != nil {
					duration := time.Since(start)
					GlobalGUIDebugger.LogLayerOperation("delete", layerID, algorithm, params, true, nil, duration)
					GlobalGUIDebugger.LogPreviewTrigger("layer_deleted", len(lp.layersData))
				}

				lp.TriggerPreviewUpdate()
			} else {
				lp.logger.Debug("LAYER: Layer deletion cancelled", "layer_id", layerID)

				// DEBUG: Log cancellation
				if GlobalGUIDebugger != nil {
					GlobalGUIDebugger.LogLayerOperation("delete_cancelled", layerID, "", nil, false, fmt.Errorf("user cancelled"), 0)
				}
			}
		},
		fyne.CurrentApp().Driver().AllWindows()[0])
	confirmDialog.Show()
}

func (lp *LeftPanel) toggleLayerVisibility(layerID string) {
	start := time.Now()
	lp.logger.Info("LAYER: Toggle visibility requested", "layer_id", layerID)

	var algorithm string
	var oldState, newState bool
	layers := lp.pipeline.GetLayers()
	for _, layer := range layers {
		if layer.ID == layerID {
			algorithm = layer.Algorithm
			oldState = layer.Enabled
			layer.Enabled = !layer.Enabled
			newState = layer.Enabled
			lp.logger.Info("LAYER: Visibility toggled", "layer_id", layerID, "old_state", oldState, "new_state", newState)
			break
		}
	}

	// DEBUG: Log visibility toggle
	if GlobalGUIDebugger != nil {
		duration := time.Since(start)
		GlobalGUIDebugger.LogLayerOperation("toggle_visibility", layerID, algorithm, map[string]interface{}{
			"old_state": oldState,
			"new_state": newState,
		}, true, nil, duration)
		GlobalGUIDebugger.LogPreviewTrigger("layer_visibility_toggled", len(lp.layersData))
	}

	lp.refreshLayersList()
	lp.TriggerPreviewUpdate()
}

func (lp *LeftPanel) removeLayerFromPipeline(layerID string) {
	lp.logger.Info("LAYER: Removing layer from pipeline", "layer_id", layerID)
	// TODO: Add RemoveLayer method to pipeline
	lp.refreshLayersList()
	lp.TriggerPreviewUpdate()
}

func (lp *LeftPanel) selectLayer(layerID string) {
	lp.logger.Debug("LAYER: Layer selected", "layer_id", layerID)

	// DEBUG: Log layer selection
	if GlobalGUIDebugger != nil {
		var algorithm string
		for _, layer := range lp.layersData {
			if layer.ID == layerID {
				algorithm = layer.Algorithm
				break
			}
		}
		GlobalGUIDebugger.LogLayerOperation("select", layerID, algorithm, nil, true, nil, 0)
	}

	lp.selectedLayerID = layerID
	lp.updateParametersForLayer(layerID)
}

func (lp *LeftPanel) refreshLayersList() {
	lp.logger.Debug("LAYER: Refreshing layers list")
	lp.layersData = lp.pipeline.GetLayers()
	lp.logger.Debug("LAYER: Current layers count", "count", len(lp.layersData))
	for i, layer := range lp.layersData {
		lp.logger.Debug("LAYER: Layer in list", "index", i, "id", layer.ID, "name", layer.Name, "algorithm", layer.Algorithm, "enabled", layer.Enabled)
	}

	// DEBUG: Log layer list state
	if GlobalGUIDebugger != nil {
		layerInfo := make([]map[string]interface{}, len(lp.layersData))
		for i, layer := range lp.layersData {
			layerInfo[i] = map[string]interface{}{
				"id":        layer.ID,
				"name":      layer.Name,
				"algorithm": layer.Algorithm,
				"enabled":   layer.Enabled,
			}
		}
		GlobalGUIDebugger.LogProcessingEvent("layers_refreshed", "layer", len(lp.layersData), map[string]interface{}{
			"layers": layerInfo,
		})
	}

	lp.layersList.Refresh()
}

func (lp *LeftPanel) updateParametersForLayer(layerID string) {
	lp.logger.Debug("LAYER: Updating parameters for layer", "layer_id", layerID)

	lp.currentParams.RemoveAll()

	if layerID == "" {
		lp.currentParams.Add(widget.NewLabel("Select a layer to edit parameters"))
		lp.currentParams.Refresh()
		return
	}

	var targetLayer *layers.Layer
	for _, layer := range lp.layersData {
		if layer.ID == layerID {
			targetLayer = layer
			break
		}
	}

	if targetLayer == nil {
		lp.logger.Error("LAYER: Target layer not found", "layer_id", layerID)
		lp.currentParams.Add(widget.NewLabel("Layer not found"))
		lp.currentParams.Refresh()
		return
	}

	algorithm, exists := algorithms.Get(targetLayer.Algorithm)
	if !exists {
		lp.logger.Error("LAYER: Algorithm not found for layer", "algorithm", targetLayer.Algorithm, "layer_id", layerID)
		lp.currentParams.Add(widget.NewLabel("Algorithm not found"))
		lp.currentParams.Refresh()
		return
	}

	lp.logger.Debug("LAYER: Creating parameter controls", "layer_id", layerID, "algorithm", targetLayer.Algorithm)

	paramInfo := algorithm.GetParameterInfo()
	for _, param := range paramInfo {
		lp.createParameterWidget(param, targetLayer.Parameters, layerID)
	}

	lp.currentParams.Refresh()
	lp.logger.Debug("LAYER: Parameter controls created", "param_count", len(paramInfo))
}

func (lp *LeftPanel) createParameterWidget(param algorithms.ParameterInfo, params map[string]interface{}, layerID string) {
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
			start := time.Now()
			valueLabel.SetText(fmt.Sprintf("%.0f", value))
			oldValue := params[param.Name]
			params[param.Name] = value

			lp.logger.Info("LAYER: Parameter changed", "layer_id", layerID, "param", param.Name, "old_value", oldValue, "new_value", value)

			// DEBUG: Log parameter change
			if GlobalGUIDebugger != nil {
				duration := time.Since(start)
				var algorithm string
				for _, layer := range lp.layersData {
					if layer.ID == layerID {
						algorithm = layer.Algorithm
						break
					}
				}
				GlobalGUIDebugger.LogLayerOperation("parameter_change", layerID, algorithm, map[string]interface{}{
					"parameter":  param.Name,
					"old_value":  oldValue,
					"new_value":  value,
					"param_type": param.Type,
				}, true, nil, duration)
				GlobalGUIDebugger.LogPreviewTrigger("parameter_changed", len(lp.layersData))
			}

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
			start := time.Now()
			valueLabel.SetText(fmt.Sprintf("%.2f", value))
			oldValue := params[param.Name]
			params[param.Name] = value

			lp.logger.Info("LAYER: Parameter changed", "layer_id", layerID, "param", param.Name, "old_value", oldValue, "new_value", value)

			// DEBUG: Log parameter change
			if GlobalGUIDebugger != nil {
				duration := time.Since(start)
				var algorithm string
				for _, layer := range lp.layersData {
					if layer.ID == layerID {
						algorithm = layer.Algorithm
						break
					}
				}
				GlobalGUIDebugger.LogLayerOperation("parameter_change", layerID, algorithm, map[string]interface{}{
					"parameter":  param.Name,
					"old_value":  oldValue,
					"new_value":  value,
					"param_type": param.Type,
				}, true, nil, duration)
				GlobalGUIDebugger.LogPreviewTrigger("parameter_changed", len(lp.layersData))
			}

			lp.onParameterChange(layerID, params)
		}

		paramContainer := container.NewVBox(
			container.NewHBox(slider, valueLabel),
		)
		lp.currentParams.Add(paramContainer)

	case "bool":
		check := widget.NewCheck("", func(checked bool) {
			start := time.Now()
			oldValue := params[param.Name]
			params[param.Name] = checked

			lp.logger.Info("LAYER: Parameter changed", "layer_id", layerID, "param", param.Name, "old_value", oldValue, "new_value", checked)

			// DEBUG: Log parameter change
			if GlobalGUIDebugger != nil {
				duration := time.Since(start)
				var algorithm string
				for _, layer := range lp.layersData {
					if layer.ID == layerID {
						algorithm = layer.Algorithm
						break
					}
				}
				GlobalGUIDebugger.LogLayerOperation("parameter_change", layerID, algorithm, map[string]interface{}{
					"parameter":  param.Name,
					"old_value":  oldValue,
					"new_value":  checked,
					"param_type": param.Type,
				}, true, nil, duration)
				GlobalGUIDebugger.LogPreviewTrigger("parameter_changed", len(lp.layersData))
			}

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
			start := time.Now()
			oldValue := params[param.Name]
			params[param.Name] = selected

			lp.logger.Info("LAYER: Parameter changed", "layer_id", layerID, "param", param.Name, "old_value", oldValue, "new_value", selected)

			// DEBUG: Log parameter change
			if GlobalGUIDebugger != nil {
				duration := time.Since(start)
				var algorithm string
				for _, layer := range lp.layersData {
					if layer.ID == layerID {
						algorithm = layer.Algorithm
						break
					}
				}
				GlobalGUIDebugger.LogLayerOperation("parameter_change", layerID, algorithm, map[string]interface{}{
					"parameter":  param.Name,
					"old_value":  oldValue,
					"new_value":  selected,
					"param_type": param.Type,
				}, true, nil, duration)
				GlobalGUIDebugger.LogPreviewTrigger("parameter_changed", len(lp.layersData))
			}

			lp.onParameterChange(layerID, params)
		})
		if val, ok := params[param.Name].(string); ok {
			selectWidget.SetSelected(val)
		} else if defaultVal, ok := param.Default.(string); ok {
			selectWidget.SetSelected(defaultVal)
		}
		lp.currentParams.Add(selectWidget)
	}

	if param.Description != "" {
		descLabel := widget.NewLabel(param.Description)
		descLabel.TextStyle = fyne.TextStyle{Italic: true}
		lp.currentParams.Add(descLabel)
	}

	lp.currentParams.Add(widget.NewSeparator())
}

func (lp *LeftPanel) onParameterChange(layerID string, params map[string]interface{}) {
	lp.logger.Info("LAYER: Parameter change triggering preview update", "layer_id", layerID)
	if lp.onParameterChanged != nil {
		lp.onParameterChanged(layerID, params)
	}
	lp.TriggerPreviewUpdate()
}

func (lp *LeftPanel) TriggerPreviewUpdate() {
	lp.logger.Info("LAYER: Triggering preview update via pipeline")

	// DEBUG: Log preview processing attempt
	if GlobalGUIDebugger != nil {
		layerCount := len(lp.layersData)
		GlobalGUIDebugger.LogProcessingEvent("trigger", "layer", layerCount, map[string]interface{}{
			"trigger_source": "left_panel",
		})
	}

	if !lp.pipeline.IsProcessing() {
		lp.logger.Debug("LAYER: Pipeline not processing, may need manual trigger")
	}
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
	lp.logger.Info("LAYER: Enabling processing controls")
	lp.enabled = true
	lp.addLayerBtn.Enable()
	lp.algorithmSelect.Enable()
}

func (lp *LeftPanel) Disable() {
	lp.logger.Debug("LAYER: Disabling processing controls")
	lp.enabled = false
	lp.addLayerBtn.Disable()
	lp.algorithmSelect.Disable()
}

func (lp *LeftPanel) Reset() {
	lp.logger.Info("LAYER: Resetting left panel")

	// DEBUG: Log reset
	if GlobalGUIDebugger != nil {
		GlobalGUIDebugger.LogProcessingEvent("reset", "none", 0, map[string]interface{}{
			"previous_layer_count": len(lp.layersData),
		})
	}

	lp.layersData = make([]*layers.Layer, 0)
	lp.layersList.Refresh()
	lp.algorithmSelect.ClearSelected()
	lp.selectedLayerID = ""
	lp.clearParameters()
}

func (lp *LeftPanel) UpdateSelectionState(hasSelection bool) {
	lp.logger.Debug("LAYER: Selection state updated", "has_selection", hasSelection)
}

func (lp *LeftPanel) SetCallbacks(onParameterChanged func(string, map[string]interface{})) {
	lp.onParameterChanged = onParameterChanged
}
