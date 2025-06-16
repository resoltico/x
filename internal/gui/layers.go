// Layer management UI panel with selection integration
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/algorithms"
	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/layers"
)

// LayerPanel manages layer-based processing UI
type LayerPanel struct {
	pipeline      *core.EnhancedPipeline
	regionManager *core.RegionManager
	logger        *slog.Logger

	vbox            *fyne.Container
	modeSwitch      *widget.Check
	algorithmSelect *widget.Select
	regionSelect    *widget.Select
	layerList       *widget.List
	paramContainer  *fyne.Container
	addButton       *widget.Button

	enabled       bool
	currentLayers []*layers.Layer
	selectedLayer int
	paramWidgets  map[string]fyne.CanvasObject
}

func NewLayerPanel(pipeline *core.EnhancedPipeline, regionManager *core.RegionManager, logger *slog.Logger) *LayerPanel {
	panel := &LayerPanel{
		pipeline:      pipeline,
		regionManager: regionManager,
		logger:        logger,
		selectedLayer: -1,
		paramWidgets:  make(map[string]fyne.CanvasObject),
	}

	panel.initializeUI()
	return panel
}

func (lp *LayerPanel) initializeUI() {
	// Mode switcher
	lp.modeSwitch = widget.NewCheck("Use Layer Mode", func(checked bool) {
		lp.pipeline.SetProcessingMode(checked)
		lp.updateUIState(checked)
	})

	// Algorithm selection
	categories := algorithms.GetAlgorithmsByCategory()
	var algorithmOptions []string
	for category, algs := range categories {
		for _, alg := range algs {
			algorithmOptions = append(algorithmOptions, fmt.Sprintf("%s - %s", category, alg))
		}
	}

	lp.algorithmSelect = widget.NewSelect(algorithmOptions, nil)
	lp.algorithmSelect.PlaceHolder = "Select algorithm..."

	// Region selection
	lp.regionSelect = widget.NewSelect([]string{"Global (no region)"}, nil)
	lp.regionSelect.SetSelected("Global (no region)")

	// Add layer button
	lp.addButton = widget.NewButton("Add Layer", func() {
		lp.addLayer()
	})

	// Layer list
	lp.layerList = widget.NewList(
		func() int {
			return len(lp.currentLayers)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", nil),
				widget.NewLabel("Layer"),
				widget.NewButton("X", nil),
			)
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			if id >= len(lp.currentLayers) {
				return
			}

			layer := lp.currentLayers[id]
			hbox := item.(*fyne.Container)

			check := hbox.Objects[0].(*widget.Check)
			label := hbox.Objects[1].(*widget.Label)
			deleteBtn := hbox.Objects[2].(*widget.Button)

			check.SetChecked(layer.Enabled)
			check.OnChanged = func(checked bool) {
				lp.toggleLayer(id, checked)
			}

			label.SetText(fmt.Sprintf("%s (%s)", layer.Name, layer.Algorithm))

			deleteBtn.OnTapped = func() {
				lp.deleteLayer(id)
			}
		},
	)

	lp.layerList.OnSelected = func(id widget.ListItemID) {
		lp.selectedLayer = id
		lp.updateParameterPanel()
	}

	// Parameter container
	lp.paramContainer = container.NewVBox()

	// Main container
	content := container.NewVBox(
		widget.NewCard("Processing Mode", "", lp.modeSwitch),
		widget.NewSeparator(),
		widget.NewCard("Add Layer", "",
			container.NewVBox(
				widget.NewLabel("Algorithm:"),
				lp.algorithmSelect,
				widget.NewLabel("Apply to region:"),
				lp.regionSelect,
				lp.addButton,
			),
		),
		widget.NewSeparator(),
		widget.NewCard("Layers", "", lp.layerList),
		widget.NewSeparator(),
		widget.NewCard("Layer Parameters", "", lp.paramContainer),
	)

	lp.vbox = container.NewVBox(content)
	lp.updateUIState(false)
}

func (lp *LayerPanel) updateUIState(layerMode bool) {
	if layerMode {
		lp.algorithmSelect.Enable()
		lp.regionSelect.Enable()
		lp.addButton.Enable()
		// Note: widget.List doesn't have Enable/Disable methods
		// The list will be functional when layer mode is enabled
	} else {
		lp.algorithmSelect.Disable()
		lp.regionSelect.Disable()
		lp.addButton.Disable()
		// Note: widget.List doesn't have Enable/Disable methods
		// The list will not be functional when layer mode is disabled
	}
}

func (lp *LayerPanel) addLayer() {
	selected := lp.algorithmSelect.Selected
	if selected == "" {
		return
	}

	// Parse algorithm from "Category - algorithm" format
	var algorithmName string
	categories := algorithms.GetAlgorithmsByCategory()
	for _, algs := range categories {
		for _, alg := range algs {
			if selected == fmt.Sprintf("%s - %s", lp.getCategoryForAlgorithm(alg), alg) {
				algorithmName = alg
				break
			}
		}
	}

	if algorithmName == "" {
		return
	}

	// Get region ID
	var regionID string
	if lp.regionSelect.Selected != "Global (no region)" {
		regionID = lp.regionSelect.Selected
	}

	// Get default parameters
	algorithm, exists := algorithms.Get(algorithmName)
	if !exists {
		return
	}

	params := algorithm.GetDefaultParams()
	name := fmt.Sprintf("Layer %d", len(lp.currentLayers)+1)

	// Add layer to pipeline
	layerID, err := lp.pipeline.AddLayer(name, algorithmName, params, regionID)
	if err != nil {
		lp.logger.Error("Failed to add layer", "error", err)
		return
	}

	lp.logger.Debug("Added layer", "layer_id", layerID, "algorithm", algorithmName)
	lp.refreshLayers()
}

func (lp *LayerPanel) getCategoryForAlgorithm(algorithm string) string {
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

func (lp *LayerPanel) toggleLayer(index int, enabled bool) {
	// For now, layers don't have individual toggle - this would require extending layer system
	lp.logger.Debug("Layer toggle", "index", index, "enabled", enabled)
	lp.refreshLayers()
}

func (lp *LayerPanel) deleteLayer(index int) {
	// For now, layers don't have individual delete - this would require extending layer system
	lp.logger.Debug("Layer delete", "index", index)
	lp.refreshLayers()
}

func (lp *LayerPanel) refreshLayers() {
	lp.currentLayers = lp.pipeline.GetLayers()
	lp.layerList.Refresh()
	lp.updateRegionList()
}

func (lp *LayerPanel) updateRegionList() {
	regions := []string{"Global (no region)"}

	selections := lp.regionManager.GetAllSelections()
	for _, selection := range selections {
		regions = append(regions, selection.ID)
	}

	lp.regionSelect.Options = regions
	lp.regionSelect.Refresh()
}

func (lp *LayerPanel) updateParameterPanel() {
	lp.paramContainer.RemoveAll()

	if lp.selectedLayer < 0 || lp.selectedLayer >= len(lp.currentLayers) {
		lp.paramContainer.Add(widget.NewLabel("Select a layer to edit parameters"))
		return
	}

	layer := lp.currentLayers[lp.selectedLayer]
	algorithm, exists := algorithms.Get(layer.Algorithm)
	if !exists {
		lp.paramContainer.Add(widget.NewLabel("Algorithm not found"))
		return
	}

	// Show layer properties
	lp.paramContainer.Add(widget.NewLabel(fmt.Sprintf("Layer: %s", layer.Name)))
	lp.paramContainer.Add(widget.NewLabel(fmt.Sprintf("Algorithm: %s", layer.Algorithm)))
	if layer.RegionID != "" {
		lp.paramContainer.Add(widget.NewLabel(fmt.Sprintf("Region: %s", layer.RegionID)))
	}
	lp.paramContainer.Add(widget.NewSeparator())

	// Show algorithm parameters (read-only for now)
	paramInfo := algorithm.GetParameterInfo()
	for _, param := range paramInfo {
		value := layer.Parameters[param.Name]
		lp.paramContainer.Add(widget.NewLabel(fmt.Sprintf("%s: %v", param.Name, value)))
	}
}

func (lp *LayerPanel) GetContainer() fyne.CanvasObject {
	return lp.vbox
}

func (lp *LayerPanel) Enable() {
	lp.enabled = true
	lp.modeSwitch.Enable()
}

func (lp *LayerPanel) Disable() {
	lp.enabled = false
	lp.modeSwitch.Disable()
	lp.updateUIState(false)
}

func (lp *LayerPanel) Refresh() {
	lp.refreshLayers()
	lp.vbox.Refresh()
}

func (lp *LayerPanel) SetSelectionChangedCallback(callback func()) {
	// Update region list when selections change
	go func() {
		callback()
		lp.updateRegionList()
	}()
}
