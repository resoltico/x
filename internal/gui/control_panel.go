// internal/gui/control_panel.go
// Fixed control panel with proper container sizing
package gui

import (
	"fmt"
	"log/slog"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/algorithms"
	"advanced-image-processing/internal/core"
)

type ControlPanel struct {
	pipeline      *core.EnhancedPipeline
	regionManager *core.RegionManager
	logger        *slog.Logger

	container *fyne.Container

	// Mode control
	modeToggle *widget.RadioGroup

	// Dynamic main section
	mainSection *fyne.Container

	// Layer mode components
	layerStack  *widget.List
	layerData   []*LayerItem
	addLayerBtn *widget.Button

	// Sequential mode components
	algorithmSequence *widget.List
	sequenceData      []*SequenceItem
	addAlgorithmBtn   *widget.Button

	// Parameters area
	parametersArea *container.Scroll
	currentParams  *fyne.Container

	// State
	isLayerMode   bool
	selectedIndex int
	enabled       bool

	// Callbacks
	onModeChanged      func(bool)
	onSelectionChanged func()
}

type LayerItem struct {
	ID         string
	Name       string
	Algorithm  string
	RegionID   string
	BlendMode  string
	Visible    bool
	Expanded   bool
	Parameters map[string]interface{}
}

type SequenceItem struct {
	Algorithm  string
	Parameters map[string]interface{}
	Expanded   bool
}

func NewControlPanel(pipeline *core.EnhancedPipeline, regionManager *core.RegionManager, logger *slog.Logger) *ControlPanel {
	panel := &ControlPanel{
		pipeline:      pipeline,
		regionManager: regionManager,
		logger:        logger,
		selectedIndex: -1,
		isLayerMode:   false,
		layerData:     make([]*LayerItem, 0),
		sequenceData:  make([]*SequenceItem, 0),
	}

	panel.initializeUI()
	return panel
}

func (cp *ControlPanel) initializeUI() {
	// Initialize main section
	cp.mainSection = container.NewVBox()

	// Processing Mode Toggle - make it more compact
	cp.modeToggle = widget.NewRadioGroup([]string{"Sequential Mode", "Layer Mode"}, nil)
	cp.modeToggle.Horizontal = false

	modeCard := widget.NewCard("Processing Mode", "", cp.modeToggle)

	// Parameters area with reasonable height
	cp.currentParams = container.NewVBox(
		widget.NewLabel("Select an item above to edit parameters"),
	)
	cp.parametersArea = container.NewScroll(cp.currentParams)

	parametersCard := widget.NewCard("Parameters", "", cp.parametersArea)

	// Main content in VBox
	content := container.NewVBox(
		modeCard,
		widget.NewSeparator(),
		cp.mainSection,
		widget.NewSeparator(),
		parametersCard,
	)

	// Create scroll container and wrap it
	scroll := container.NewScroll(content)
	cp.container = container.NewBorder(nil, nil, nil, nil, scroll)

	// Update main section first time
	cp.updateMainSection()

	// Set callback after initialization
	cp.modeToggle.OnChanged = func(value string) {
		cp.isLayerMode = (value == "Layer Mode")
		cp.updateMainSection()
		if cp.onModeChanged != nil {
			cp.onModeChanged(cp.isLayerMode)
		}
	}
	cp.modeToggle.SetSelected("Sequential Mode")

	cp.Disable()
}

func (cp *ControlPanel) updateMainSection() {
	if cp.mainSection == nil {
		cp.mainSection = container.NewVBox()
		return
	}

	cp.mainSection.RemoveAll()

	if cp.isLayerMode {
		cp.setupLayerMode()
	} else {
		cp.setupSequentialMode()
	}

	cp.mainSection.Refresh()
}

func (cp *ControlPanel) setupLayerMode() {
	// Layer stack list with fixed height
	cp.layerStack = widget.NewList(
		func() int { return len(cp.layerData) },
		cp.createLayerTemplate,
		cp.updateLayerItem,
	)
	cp.layerStack.OnSelected = func(id widget.ListItemID) {
		cp.selectedIndex = id
		cp.updateParametersArea()
	}

	stackScroll := container.NewScroll(cp.layerStack)
	stackScroll.SetMinSize(fyne.NewSize(280, 300))

	cp.addLayerBtn = widget.NewButtonWithIcon("Add Layer", theme.ContentAddIcon(), cp.showAddLayerDialog)
	cp.addLayerBtn.Importance = widget.HighImportance

	layerCard := widget.NewCard("Layer Stack", "",
		container.NewVBox(
			widget.NewLabel("Layers process from top to bottom:"),
			stackScroll,
			cp.addLayerBtn,
		))

	cp.mainSection.Add(layerCard)
}

func (cp *ControlPanel) setupSequentialMode() {
	// Algorithm sequence list
	cp.algorithmSequence = widget.NewList(
		func() int { return len(cp.sequenceData) },
		cp.createSequenceTemplate,
		cp.updateSequenceItem,
	)
	cp.algorithmSequence.OnSelected = func(id widget.ListItemID) {
		cp.selectedIndex = id
		cp.updateParametersArea()
	}

	sequenceScroll := container.NewScroll(cp.algorithmSequence)

	cp.addAlgorithmBtn = widget.NewButtonWithIcon("Add Algorithm", theme.ContentAddIcon(), cp.showAddAlgorithmDialog)
	cp.addAlgorithmBtn.Importance = widget.HighImportance

	sequenceCard := widget.NewCard("Algorithm Sequence", "",
		container.NewVBox(
			widget.NewLabel("Algorithms apply in order:"),
			sequenceScroll,
			cp.addAlgorithmBtn,
		))

	cp.mainSection.Add(sequenceCard)
}

func (cp *ControlPanel) createLayerTemplate() fyne.CanvasObject {
	nameEntry := widget.NewEntry()
	nameEntry.Resize(fyne.NewSize(120, 30))

	visToggle := widget.NewCheck("", nil)

	blendSelect := widget.NewSelect([]string{"Normal", "Multiply", "Overlay", "Screen"}, nil)
	blendSelect.Resize(fyne.NewSize(80, 30))

	algorithmLabel := widget.NewLabel("Algorithm")
	algorithmLabel.Resize(fyne.NewSize(120, 30))

	expandBtn := widget.NewButtonWithIcon("", theme.MenuIcon(), nil)
	expandBtn.Resize(fyne.NewSize(30, 30))

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	deleteBtn.Resize(fyne.NewSize(30, 30))
	deleteBtn.Importance = widget.DangerImportance

	topRow := container.NewHBox(
		visToggle,
		nameEntry,
		expandBtn,
		deleteBtn,
	)

	bottomRow := container.NewHBox(
		algorithmLabel,
		blendSelect,
	)

	return container.NewVBox(topRow, bottomRow)
}

func (cp *ControlPanel) updateLayerItem(id widget.ListItemID, item fyne.CanvasObject) {
	if id >= len(cp.layerData) {
		return
	}

	layer := cp.layerData[id]
	vbox := item.(*fyne.Container)
	topRow := vbox.Objects[0].(*fyne.Container)
	bottomRow := vbox.Objects[1].(*fyne.Container)

	visToggle := topRow.Objects[0].(*widget.Check)
	nameEntry := topRow.Objects[1].(*widget.Entry)
	expandBtn := topRow.Objects[2].(*widget.Button)
	deleteBtn := topRow.Objects[3].(*widget.Button)

	visToggle.SetChecked(layer.Visible)
	nameEntry.SetText(layer.Name)

	if layer.Expanded {
		expandBtn.SetIcon(theme.MenuExpandIcon())
	} else {
		expandBtn.SetIcon(theme.MenuIcon())
	}

	algorithmLabel := bottomRow.Objects[0].(*widget.Label)
	blendSelect := bottomRow.Objects[1].(*widget.Select)

	algorithmLabel.SetText(layer.Algorithm)
	blendSelect.SetSelected(layer.BlendMode)

	visToggle.OnChanged = func(checked bool) {
		layer.Visible = checked
		cp.refreshLayers()
	}

	nameEntry.OnChanged = func(text string) {
		layer.Name = text
	}

	expandBtn.OnTapped = func() {
		layer.Expanded = !layer.Expanded
		cp.layerStack.Refresh()
		if layer.Expanded {
			cp.selectedIndex = id
			cp.updateParametersArea()
		}
	}

	deleteBtn.OnTapped = func() {
		cp.deleteLayer(id)
	}

	blendSelect.OnChanged = func(value string) {
		layer.BlendMode = value
		cp.refreshLayers()
	}
}

func (cp *ControlPanel) createSequenceTemplate() fyne.CanvasObject {
	algorithmLabel := widget.NewLabel("Algorithm")
	algorithmLabel.Resize(fyne.NewSize(150, 30))

	expandBtn := widget.NewButtonWithIcon("", theme.MenuIcon(), nil)
	expandBtn.Resize(fyne.NewSize(30, 30))

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	deleteBtn.Resize(fyne.NewSize(30, 30))
	deleteBtn.Importance = widget.DangerImportance

	return container.NewHBox(
		algorithmLabel,
		expandBtn,
		deleteBtn,
	)
}

func (cp *ControlPanel) updateSequenceItem(id widget.ListItemID, item fyne.CanvasObject) {
	if id >= len(cp.sequenceData) {
		return
	}

	sequence := cp.sequenceData[id]
	hbox := item.(*fyne.Container)

	algorithmLabel := hbox.Objects[0].(*widget.Label)
	expandBtn := hbox.Objects[1].(*widget.Button)
	deleteBtn := hbox.Objects[2].(*widget.Button)

	algorithmLabel.SetText(sequence.Algorithm)

	if sequence.Expanded {
		expandBtn.SetIcon(theme.MenuExpandIcon())
	} else {
		expandBtn.SetIcon(theme.MenuIcon())
	}

	expandBtn.OnTapped = func() {
		sequence.Expanded = !sequence.Expanded
		cp.algorithmSequence.Refresh()
		if sequence.Expanded {
			cp.selectedIndex = id
			cp.updateParametersArea()
		}
	}

	deleteBtn.OnTapped = func() {
		cp.deleteSequenceItem(id)
	}
}

func (cp *ControlPanel) showAddLayerDialog() {
	categories := algorithms.GetAlgorithmsByCategory()
	var algorithmOptions []string
	for category, algs := range categories {
		for _, alg := range algs {
			algorithmOptions = append(algorithmOptions, fmt.Sprintf("%s → %s", category, alg))
		}
	}

	algorithmSelect := widget.NewSelect(algorithmOptions, nil)
	algorithmSelect.PlaceHolder = "Choose algorithm..."

	regionOptions := []string{"Global (entire image)"}
	selections := cp.regionManager.GetAllSelections()
	for _, selection := range selections {
		var icon string
		switch selection.Type {
		case core.SelectionRectangle:
			icon = "📐"
		case core.SelectionFreehand:
			icon = "✏️"
		}
		regionOptions = append(regionOptions, fmt.Sprintf("%s %s", icon, selection.ID))
	}

	regionSelect := widget.NewSelect(regionOptions, nil)
	regionSelect.SetSelected("Global (entire image)")

	nameEntry := widget.NewEntry()
	nameEntry.SetText(fmt.Sprintf("Layer %d", len(cp.layerData)+1))

	content := container.NewVBox(
		widget.NewLabel("Layer Name:"),
		nameEntry,
		widget.NewSeparator(),
		widget.NewLabel("Algorithm:"),
		algorithmSelect,
		widget.NewSeparator(),
		widget.NewLabel("Apply to Region:"),
		regionSelect,
	)

	dialog := widget.NewModalPopUp(content, fyne.CurrentApp().Driver().AllWindows()[0].Canvas())
	dialog.Resize(fyne.NewSize(400, 300))

	addBtn := widget.NewButton("Add Layer", func() {
		if algorithmSelect.Selected != "" {
			cp.addLayer(nameEntry.Text, algorithmSelect.Selected, regionSelect.Selected)
			dialog.Hide()
		}
	})
	addBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		dialog.Hide()
	})

	buttons := container.NewHBox(addBtn, cancelBtn)
	content.Add(widget.NewSeparator())
	content.Add(buttons)

	dialog.Show()
}

func (cp *ControlPanel) showAddAlgorithmDialog() {
	categories := algorithms.GetAlgorithmsByCategory()
	var algorithmOptions []string
	for category, algs := range categories {
		for _, alg := range algs {
			algorithmOptions = append(algorithmOptions, fmt.Sprintf("%s → %s", category, alg))
		}
	}

	algorithmSelect := widget.NewSelect(algorithmOptions, nil)
	algorithmSelect.PlaceHolder = "Choose algorithm..."

	content := container.NewVBox(
		widget.NewLabel("Select Algorithm:"),
		algorithmSelect,
	)

	dialog := widget.NewModalPopUp(content, fyne.CurrentApp().Driver().AllWindows()[0].Canvas())
	dialog.Resize(fyne.NewSize(350, 200))

	addBtn := widget.NewButton("Add Algorithm", func() {
		if algorithmSelect.Selected != "" {
			cp.addAlgorithm(algorithmSelect.Selected)
			dialog.Hide()
		}
	})
	addBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		dialog.Hide()
	})

	buttons := container.NewHBox(addBtn, cancelBtn)
	content.Add(widget.NewSeparator())
	content.Add(buttons)

	dialog.Show()
}

func (cp *ControlPanel) addLayer(name, algorithmSelection, regionSelection string) {
	var algorithmName string
	categories := algorithms.GetAlgorithmsByCategory()
	for _, algs := range categories {
		for _, alg := range algs {
			if algorithmSelection == fmt.Sprintf("%s → %s", cp.getCategoryForAlgorithm(alg), alg) {
				algorithmName = alg
				break
			}
		}
	}

	if algorithmName == "" {
		return
	}

	algorithm, exists := algorithms.Get(algorithmName)
	if !exists {
		return
	}

	params := algorithm.GetDefaultParams()

	var regionID string
	if regionSelection != "Global (entire image)" {
		parts := strings.Fields(regionSelection)
		if len(parts) > 1 {
			regionID = parts[1]
		}
	}

	layerID, err := cp.pipeline.AddLayer(name, algorithmName, params, regionID)
	if err != nil {
		cp.logger.Error("Failed to add layer", "error", err)
		return
	}

	layer := &LayerItem{
		ID:         layerID,
		Name:       name,
		Algorithm:  algorithmName,
		RegionID:   regionID,
		BlendMode:  "Normal",
		Visible:    true,
		Expanded:   false,
		Parameters: params,
	}

	cp.layerData = append(cp.layerData, layer)
	cp.layerStack.Refresh()
}

func (cp *ControlPanel) addAlgorithm(algorithmSelection string) {
	var algorithmName string
	categories := algorithms.GetAlgorithmsByCategory()
	for _, algs := range categories {
		for _, alg := range algs {
			if algorithmSelection == fmt.Sprintf("%s → %s", cp.getCategoryForAlgorithm(alg), alg) {
				algorithmName = alg
				break
			}
		}
	}

	if algorithmName == "" {
		return
	}

	algorithm, exists := algorithms.Get(algorithmName)
	if !exists {
		return
	}

	params := algorithm.GetDefaultParams()

	err := cp.pipeline.AddStep(algorithmName, params)
	if err != nil {
		cp.logger.Error("Failed to add algorithm", "error", err)
		return
	}

	sequence := &SequenceItem{
		Algorithm:  algorithmName,
		Parameters: params,
		Expanded:   false,
	}

	cp.sequenceData = append(cp.sequenceData, sequence)
	cp.algorithmSequence.Refresh()
}

func (cp *ControlPanel) deleteLayer(index int) {
	if index < 0 || index >= len(cp.layerData) {
		return
	}

	cp.layerData = append(cp.layerData[:index], cp.layerData[index+1:]...)
	cp.layerStack.Refresh()
	cp.selectedIndex = -1
	cp.updateParametersArea()
}

func (cp *ControlPanel) deleteSequenceItem(index int) {
	if index < 0 || index >= len(cp.sequenceData) {
		return
	}

	cp.sequenceData = append(cp.sequenceData[:index], cp.sequenceData[index+1:]...)
	cp.algorithmSequence.Refresh()
	cp.selectedIndex = -1
	cp.updateParametersArea()
}

func (cp *ControlPanel) updateParametersArea() {
	if cp.currentParams == nil {
		cp.currentParams = container.NewVBox()
	}

	cp.currentParams.RemoveAll()

	if cp.selectedIndex < 0 {
		cp.currentParams.Add(widget.NewLabel("Select an item above to edit parameters"))
		cp.currentParams.Refresh()
		return
	}

	var algorithm algorithms.Algorithm
	var params map[string]interface{}
	var exists bool

	if cp.isLayerMode {
		if cp.selectedIndex >= len(cp.layerData) {
			return
		}
		layer := cp.layerData[cp.selectedIndex]
		algorithm, exists = algorithms.Get(layer.Algorithm)
		params = layer.Parameters

		cp.currentParams.Add(widget.NewLabel(fmt.Sprintf("Layer: %s", layer.Name)))
		cp.currentParams.Add(widget.NewLabel(fmt.Sprintf("Algorithm: %s", layer.Algorithm)))
		if layer.RegionID != "" {
			cp.currentParams.Add(widget.NewLabel(fmt.Sprintf("Region: %s", layer.RegionID)))
		}
	} else {
		if cp.selectedIndex >= len(cp.sequenceData) {
			return
		}
		sequence := cp.sequenceData[cp.selectedIndex]
		algorithm, exists = algorithms.Get(sequence.Algorithm)
		params = sequence.Parameters

		cp.currentParams.Add(widget.NewLabel(fmt.Sprintf("Algorithm: %s", sequence.Algorithm)))
	}

	if !exists {
		cp.currentParams.Add(widget.NewLabel("Algorithm not found"))
		cp.currentParams.Refresh()
		return
	}

	cp.currentParams.Add(widget.NewSeparator())
	cp.currentParams.Add(widget.NewLabel("Parameters:"))

	paramInfo := algorithm.GetParameterInfo()
	for _, param := range paramInfo {
		cp.createParameterWidget(param, params)
	}

	cp.currentParams.Refresh()
}

func (cp *ControlPanel) createParameterWidget(param algorithms.ParameterInfo, params map[string]interface{}) {
	cp.currentParams.Add(widget.NewLabel(param.Name + ":"))

	switch param.Type {
	case "int":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		if val, ok := params[param.Name].(float64); ok {
			slider.SetValue(val)
		}
		slider.Step = 1

		valueLabel := widget.NewLabel(fmt.Sprintf("%.0f", slider.Value))
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.0f", value))
			params[param.Name] = value
			cp.refreshProcessing()
		}

		cp.currentParams.Add(container.NewHBox(slider, valueLabel))

	case "float":
		slider := widget.NewSlider(param.Min.(float64), param.Max.(float64))
		if val, ok := params[param.Name].(float64); ok {
			slider.SetValue(val)
		}
		slider.Step = 0.1

		valueLabel := widget.NewLabel(fmt.Sprintf("%.2f", slider.Value))
		slider.OnChanged = func(value float64) {
			valueLabel.SetText(fmt.Sprintf("%.2f", value))
			params[param.Name] = value
			cp.refreshProcessing()
		}

		cp.currentParams.Add(container.NewHBox(slider, valueLabel))

	case "bool":
		check := widget.NewCheck("", func(checked bool) {
			params[param.Name] = checked
			cp.refreshProcessing()
		})
		if val, ok := params[param.Name].(bool); ok {
			check.SetChecked(val)
		}
		cp.currentParams.Add(check)

	case "enum":
		selectWidget := widget.NewSelect(param.Options, func(selected string) {
			params[param.Name] = selected
			cp.refreshProcessing()
		})
		if val, ok := params[param.Name].(string); ok {
			selectWidget.SetSelected(val)
		}
		cp.currentParams.Add(selectWidget)
	}

	cp.currentParams.Add(widget.NewLabel(param.Description))
	cp.currentParams.Add(widget.NewSeparator())
}

func (cp *ControlPanel) getCategoryForAlgorithm(algorithm string) string {
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

func (cp *ControlPanel) refreshLayers() {
	// Trigger layer refresh in pipeline
}

func (cp *ControlPanel) refreshProcessing() {
	// Trigger processing update
}

func (cp *ControlPanel) GetContainer() fyne.CanvasObject {
	return cp.container
}

func (cp *ControlPanel) Enable() {
	cp.enabled = true
	cp.modeToggle.Enable()
	if cp.addLayerBtn != nil {
		cp.addLayerBtn.Enable()
	}
	if cp.addAlgorithmBtn != nil {
		cp.addAlgorithmBtn.Enable()
	}
}

func (cp *ControlPanel) Disable() {
	cp.enabled = false
	cp.modeToggle.Disable()
	if cp.addLayerBtn != nil {
		cp.addLayerBtn.Disable()
	}
	if cp.addAlgorithmBtn != nil {
		cp.addAlgorithmBtn.Disable()
	}
}

func (cp *ControlPanel) Reset() {
	cp.layerData = make([]*LayerItem, 0)
	cp.sequenceData = make([]*SequenceItem, 0)
	cp.selectedIndex = -1

	if cp.layerStack != nil {
		cp.layerStack.Refresh()
	}
	if cp.algorithmSequence != nil {
		cp.algorithmSequence.Refresh()
	}
	cp.updateParametersArea()
}

func (cp *ControlPanel) UpdateSelectionState(hasSelection bool) {
	if cp.onSelectionChanged != nil {
		cp.onSelectionChanged()
	}
}

func (cp *ControlPanel) SetCallbacks(onModeChanged func(bool), onSelectionChanged func()) {
	cp.onModeChanged = onModeChanged
	cp.onSelectionChanged = onSelectionChanged
}
