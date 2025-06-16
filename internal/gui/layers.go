// Layer management UI panel with improved sizing and scrolling
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

// LayerPanel manages layer-based processing UI with improved UX
type LayerPanel struct {
	pipeline      *core.EnhancedPipeline
	regionManager *core.RegionManager
	logger        *slog.Logger

	vbox            *fyne.Container
	modeSwitch      *widget.Check
	algorithmSelect *widget.Select
	regionSelect    *widget.Select
	layerScrollArea *container.Scroll
	layerList       *widget.List
	paramContainer  *fyne.Container
	addButton       *widget.Button
	clearAllButton  *widget.Button

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
	// Mode switcher with better styling
	lp.modeSwitch = widget.NewCheck("🎨 Enable Layer Processing Mode", func(checked bool) {
		lp.pipeline.SetProcessingMode(checked)
		lp.updateUIState(checked)
	})

	// Algorithm selection with better organization
	categories := algorithms.GetAlgorithmsByCategory()
	var algorithmOptions []string
	for category, algs := range categories {
		for _, alg := range algs {
			algorithmOptions = append(algorithmOptions, fmt.Sprintf("%s → %s", category, alg))
		}
	}

	lp.algorithmSelect = widget.NewSelect(algorithmOptions, nil)
	lp.algorithmSelect.PlaceHolder = "Choose an algorithm..."

	// Region selection with better labeling
	lp.regionSelect = widget.NewSelect([]string{"🌐 Global (entire image)"}, nil)
	lp.regionSelect.SetSelected("🌐 Global (entire image)")

	// Action buttons with better styling
	lp.addButton = widget.NewButton("➕ Add Layer", func() {
		lp.addLayer()
	})
	lp.addButton.Importance = widget.HighImportance

	lp.clearAllButton = widget.NewButton("🗑️ Clear All Layers", func() {
		lp.clearAllLayers()
	})
	lp.clearAllButton.Importance = widget.LowImportance

	// Layer list with proper sizing and scrolling
	lp.layerList = widget.NewList(
		func() int {
			return len(lp.currentLayers)
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", nil),
				widget.NewLabel("Layer Name"),
				widget.NewButton("❌", nil),
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

			layerText := fmt.Sprintf("🎯 %s\n📝 %s", layer.Name, layer.Algorithm)
			if layer.RegionID != "" {
				layerText += fmt.Sprintf("\n🎯 Region: %s", layer.RegionID)
			}
			label.SetText(layerText)

			deleteBtn.OnTapped = func() {
				lp.deleteLayer(id)
			}
		},
	)

	// Create scrollable container for layer list with minimum height
	lp.layerScrollArea = container.NewVScroll(lp.layerList)
	lp.layerScrollArea.SetMinSize(fyne.NewSize(300, 200))

	lp.layerList.OnSelected = func(id widget.ListItemID) {
		lp.selectedLayer = id
		lp.updateParameterPanel()
	}

	// Parameter container for layer editing
	lp.paramContainer = container.NewVBox()

	// Main layout with proper spacing and organization
	modeCard := widget.NewCard("🔧 Processing Mode", "",
		container.NewVBox(
			lp.modeSwitch,
			widget.NewLabel("Layer mode allows region-specific processing with blending."),
		))

	addLayerCard := widget.NewCard("➕ Add New Layer", "",
		container.NewVBox(
			widget.NewLabel("Algorithm:"),
			lp.algorithmSelect,
			widget.NewSeparator(),
			widget.NewLabel("Apply to region:"),
			lp.regionSelect,
			widget.NewSeparator(),
			container.NewHBox(lp.addButton, lp.clearAllButton),
		),
	)

	layerListCard := widget.NewCard("📋 Layer Stack", "",
		container.NewVBox(
			widget.NewLabel("Layers are processed from top to bottom:"),
			lp.layerScrollArea,
		))

	paramCard := widget.NewCard("⚙️ Layer Properties", "",
		container.NewVScroll(lp.paramContainer))

	// Create main container with proper proportions
	lp.vbox = container.NewVBox(
		modeCard,
		widget.NewSeparator(),
		addLayerCard,
		widget.NewSeparator(),
		layerListCard,
		widget.NewSeparator(),
		paramCard,
	)

	lp.updateUIState(false)
}

func (lp *LayerPanel) updateUIState(layerMode bool) {
	if layerMode {
		lp.algorithmSelect.Enable()
		lp.regionSelect.Enable()
		lp.addButton.Enable()
		lp.clearAllButton.Enable()
	} else {
		lp.algorithmSelect.Disable()
		lp.regionSelect.Disable()
		lp.addButton.Disable()
		lp.clearAllButton.Disable()
	}
}

func (lp *LayerPanel) addLayer() {
	selected := lp.algorithmSelect.Selected
	if selected == "" {
		return
	}

	// Parse algorithm from "Category → algorithm" format
	var algorithmName string
	categories := algorithms.GetAlgorithmsByCategory()
	for _, algs := range categories {
		for _, alg := range algs {
			if selected == fmt.Sprintf("%s → %s", lp.getCategoryForAlgorithm(alg), alg) {
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
	if lp.regionSelect.Selected != "🌐 Global (entire image)" {
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

	// Clear selection after adding
	lp.algorithmSelect.SetSelected("")
}

func (lp *LayerPanel) clearAllLayers() {
	lp.pipeline.ClearAll()
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
	regions := []string{"🌐 Global (entire image)"}

	selections := lp.regionManager.GetAllSelections()
	for _, selection := range selections {
		var icon string
		switch selection.Type {
		case core.SelectionRectangle:
			icon = "📐"
		case core.SelectionFreehand:
			icon = "✏️"
		default:
			icon = "🎯"
		}
		regions = append(regions, fmt.Sprintf("%s %s", icon, selection.ID))
	}

	lp.regionSelect.Options = regions
	lp.regionSelect.Refresh()
}

func (lp *LayerPanel) updateParameterPanel() {
	lp.paramContainer.RemoveAll()

	if lp.selectedLayer < 0 || lp.selectedLayer >= len(lp.currentLayers) {
		lp.paramContainer.Add(widget.NewLabel("👆 Select a layer above to edit its properties"))
		return
	}

	layer := lp.currentLayers[lp.selectedLayer]
	algorithm, exists := algorithms.Get(layer.Algorithm)
	if !exists {
		lp.paramContainer.Add(widget.NewLabel("❌ Algorithm not found"))
		return
	}

	// Show layer properties with better formatting
	lp.paramContainer.Add(widget.NewLabel(fmt.Sprintf("🎯 Layer: %s", layer.Name)))
	lp.paramContainer.Add(widget.NewLabel(fmt.Sprintf("📝 Algorithm: %s", layer.Algorithm)))
	if layer.RegionID != "" {
		lp.paramContainer.Add(widget.NewLabel(fmt.Sprintf("🎯 Region: %s", layer.RegionID)))
	}
	lp.paramContainer.Add(widget.NewSeparator())

	// Show algorithm parameters (read-only for now)
	paramInfo := algorithm.GetParameterInfo()
	if len(paramInfo) == 0 {
		lp.paramContainer.Add(widget.NewLabel("ℹ️ No configurable parameters"))
	} else {
		lp.paramContainer.Add(widget.NewLabel("⚙️ Current Parameters:"))
		for _, param := range paramInfo {
			value := layer.Parameters[param.Name]
			lp.paramContainer.Add(widget.NewLabel(fmt.Sprintf("• %s: %v", param.Name, value)))
		}
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
