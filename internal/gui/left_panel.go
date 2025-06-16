// internal/gui/left_panel.go
// Perfect UI Left Panel: Control Hub (300px wide)
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/algorithms"
	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"
)

type LeftPanel struct {
	pipeline      *core.EnhancedPipeline
	regionManager *core.RegionManager
	imageData     *core.ImageData
	loader        *io.ImageLoader
	logger        *slog.Logger

	container *fyne.Container

	// Processing Mode Toggle (40px height)
	modeToggle *widget.RadioGroup

	// Main Section (910px height, split into List 650px + Parameters 260px)
	mainSection     *fyne.Container
	listContainer   *fyne.Container
	paramsContainer *container.Scroll

	// Layer Mode Components
	layerList   *widget.List
	layerData   []*LayerItem
	addLayerBtn *widget.Button

	// Sequential Mode Components
	sequenceList *widget.List
	sequenceData []*SequenceItem
	addAlgBtn    *widget.Button

	// Parameters Area
	currentParams *fyne.Container
	selectedIndex int

	// State
	isLayerMode bool
	enabled     bool

	// Callbacks
	onModeChanged func(bool)
	onImageLoaded func(string)
	onImageSaved  func(string)
	onReset       func()
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

func NewLeftPanel(pipeline *core.EnhancedPipeline, regionManager *core.RegionManager,
	imageData *core.ImageData, loader *io.ImageLoader, logger *slog.Logger) *LeftPanel {

	panel := &LeftPanel{
		pipeline:      pipeline,
		regionManager: regionManager,
		imageData:     imageData,
		loader:        loader,
		logger:        logger,
		selectedIndex: -1,
		isLayerMode:   false,
		layerData:     make([]*LayerItem, 0),
		sequenceData:  make([]*SequenceItem, 0),
	}

	panel.initializeUI()
	return panel
}

func (lp *LeftPanel) initializeUI() {
	// Processing Mode Toggle - 40px height, centered
	lp.modeToggle = widget.NewRadioGroup([]string{"Sequential Mode", "Layer Mode"}, lp.onModeToggleChanged)
	lp.modeToggle.Horizontal = false

	modeContainer := container.NewVBox(
		widget.NewLabel("Mode"),
		lp.modeToggle,
	)

	// Main Section initialization
	lp.listContainer = container.NewVBox()
	lp.currentParams = container.NewVBox(widget.NewLabel("Select an item above to edit parameters"))
	lp.paramsContainer = container.NewScroll(lp.currentParams)

	// Create main section split: List Area (650px) + Parameters Area (260px)
	lp.mainSection = container.NewVBox(
		lp.listContainer,
		widget.NewSeparator(),
		container.NewBorder(
			widget.NewLabel("Parameters"),
			nil, nil, nil,
			lp.paramsContainer,
		),
	)

	// File operations (at top)
	openBtn := widget.NewButtonWithIcon("Open Image", theme.FolderOpenIcon(), lp.openImage)
	saveBtn := widget.NewButtonWithIcon("Save Image", theme.DocumentSaveIcon(), lp.saveImage)
	resetBtn := widget.NewButtonWithIcon("Reset", theme.ViewRefreshIcon(), func() {
		if lp.onReset != nil {
			lp.onReset()
		}
	})

	fileOps := container.NewVBox(
		container.NewGridWithColumns(2, openBtn, saveBtn),
		resetBtn,
	)

	// Main container following Perfect UI spec: Mode toggle + Main section
	content := container.NewVBox(
		fileOps,
		widget.NewSeparator(),
		modeContainer,
		widget.NewSeparator(),
		lp.mainSection,
	)

	// Set fixed width to 300px as per specification
	lp.container = container.NewBorder(nil, nil, nil, nil, content)
	lp.container.Resize(fyne.NewSize(300, 1000))

	// Initialize with Sequential Mode
	lp.modeToggle.SetSelected("Sequential Mode")
	lp.updateMainSection()
	lp.Disable()
}

func (lp *LeftPanel) onModeToggleChanged(value string) {
	lp.isLayerMode = (value == "Layer Mode")
	lp.updateMainSection()
	if lp.onModeChanged != nil {
		lp.onModeChanged(lp.isLayerMode)
	}
}

func (lp *LeftPanel) updateMainSection() {
	lp.listContainer.RemoveAll()

	if lp.isLayerMode {
		lp.setupLayerMode()
	} else {
		lp.setupSequentialMode()
	}

	lp.listContainer.Refresh()
}

func (lp *LeftPanel) setupLayerMode() {
	// Layer Stack List
	lp.layerList = widget.NewList(
		func() int { return len(lp.layerData) },
		lp.createLayerTemplate,
		lp.updateLayerItem,
	)
	lp.layerList.OnSelected = func(id widget.ListItemID) {
		lp.selectedIndex = id
		lp.updateParametersArea()
	}

	// Add Layer Button
	lp.addLayerBtn = widget.NewButtonWithIcon("+ New Layer", theme.ContentAddIcon(), lp.showAddLayerDialog)
	lp.addLayerBtn.Importance = widget.HighImportance

	layerContainer := container.NewVBox(
		widget.NewLabel("Layer Stack"),
		container.NewScroll(lp.layerList),
		lp.addLayerBtn,
	)

	lp.listContainer.Add(layerContainer)
}

func (lp *LeftPanel) setupSequentialMode() {
	// Algorithm Sequence List
	lp.sequenceList = widget.NewList(
		func() int { return len(lp.sequenceData) },
		lp.createSequenceTemplate,
		lp.updateSequenceItem,
	)
	lp.sequenceList.OnSelected = func(id widget.ListItemID) {
		lp.selectedIndex = id
		lp.updateParametersArea()
	}

	// Add Algorithm Button
	lp.addAlgBtn = widget.NewButtonWithIcon("+ Add Algorithm", theme.ContentAddIcon(), lp.showAddAlgorithmDialog)
	lp.addAlgBtn.Importance = widget.HighImportance

	sequenceContainer := container.NewVBox(
		widget.NewLabel("Algorithm Sequence"),
		container.NewScroll(lp.sequenceList),
		lp.addAlgBtn,
	)

	lp.listContainer.Add(sequenceContainer)
}

func (lp *LeftPanel) createLayerTemplate() fyne.CanvasObject {
	// Layer Entry Template (70px collapsed, 140px expanded)
	nameEntry := widget.NewEntry()
	nameEntry.Resize(fyne.NewSize(90, 20))

	visToggle := widget.NewCheck("", nil)

	blendSelect := widget.NewSelect([]string{"Normal", "Multiply", "Overlay"}, nil)
	blendSelect.Resize(fyne.NewSize(80, 20))

	algorithmLabel := widget.NewLabel("Algorithm")

	expandBtn := widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), nil)
	expandBtn.Resize(fyne.NewSize(15, 15))

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	deleteBtn.Resize(fyne.NewSize(15, 15))
	deleteBtn.Importance = widget.DangerImportance

	topRow := container.NewHBox(
		nameEntry,
		visToggle,
		expandBtn,
		deleteBtn,
	)

	bottomRow := container.NewHBox(
		algorithmLabel,
		blendSelect,
	)

	return container.NewVBox(topRow, bottomRow)
}

func (lp *LeftPanel) updateLayerItem(id widget.ListItemID, item fyne.CanvasObject) {
	if id >= len(lp.layerData) {
		return
	}

	layer := lp.layerData[id]
	vbox := item.(*fyne.Container)

	topRow := vbox.Objects[0].(*fyne.Container)
	bottomRow := vbox.Objects[1].(*fyne.Container)

	nameEntry := topRow.Objects[0].(*widget.Entry)
	visToggle := topRow.Objects[1].(*widget.Check)
	expandBtn := topRow.Objects[2].(*widget.Button)
	deleteBtn := topRow.Objects[3].(*widget.Button)

	algorithmLabel := bottomRow.Objects[0].(*widget.Label)
	blendSelect := bottomRow.Objects[1].(*widget.Select)

	// Update values
	nameEntry.SetText(layer.Name)
	visToggle.SetChecked(layer.Visible)
	algorithmLabel.SetText(layer.Algorithm)
	blendSelect.SetSelected(layer.BlendMode)

	// Set callbacks
	nameEntry.OnChanged = func(text string) {
		layer.Name = text
	}

	visToggle.OnChanged = func(checked bool) {
		layer.Visible = checked
	}

	expandBtn.OnTapped = func() {
		layer.Expanded = !layer.Expanded
		lp.layerList.Refresh()
		if layer.Expanded {
			lp.selectedIndex = id
			lp.updateParametersArea()
		}
	}

	deleteBtn.OnTapped = func() {
		lp.deleteLayer(id)
	}

	blendSelect.OnChanged = func(value string) {
		layer.BlendMode = value
	}
}

func (lp *LeftPanel) createSequenceTemplate() fyne.CanvasObject {
	// Sequence Entry Template (50px collapsed, 110px expanded)
	algorithmLabel := widget.NewLabel("Algorithm")

	expandBtn := widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), nil)
	expandBtn.Resize(fyne.NewSize(15, 15))

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)
	deleteBtn.Resize(fyne.NewSize(15, 15))
	deleteBtn.Importance = widget.DangerImportance

	return container.NewHBox(
		algorithmLabel,
		expandBtn,
		deleteBtn,
	)
}

func (lp *LeftPanel) updateSequenceItem(id widget.ListItemID, item fyne.CanvasObject) {
	if id >= len(lp.sequenceData) {
		return
	}

	sequence := lp.sequenceData[id]
	hbox := item.(*fyne.Container)

	algorithmLabel := hbox.Objects[0].(*widget.Label)
	expandBtn := hbox.Objects[1].(*widget.Button)
	deleteBtn := hbox.Objects[2].(*widget.Button)

	algorithmLabel.SetText(sequence.Algorithm)

	expandBtn.OnTapped = func() {
		sequence.Expanded = !sequence.Expanded
		lp.sequenceList.Refresh()
		if sequence.Expanded {
			lp.selectedIndex = id
			lp.updateParametersArea()
		}
	}

	deleteBtn.OnTapped = func() {
		lp.deleteSequenceItem(id)
	}
}

func (lp *LeftPanel) showAddLayerDialog() {
	algorithmSelect := widget.NewSelect(lp.getAlgorithmOptions(), nil)
	algorithmSelect.PlaceHolder = "Choose algorithm..."

	regionSelect := widget.NewSelect(lp.getRegionOptions(), nil)
	regionSelect.SetSelected("Whole Image")

	nameEntry := widget.NewEntry()
	nameEntry.SetText(fmt.Sprintf("Layer %d", len(lp.layerData)+1))

	content := container.NewVBox(
		widget.NewLabel("Layer Name:"),
		nameEntry,
		widget.NewLabel("Algorithm:"),
		algorithmSelect,
		widget.NewLabel("Region:"),
		regionSelect,
	)

	dialog := widget.NewModalPopUp(content, fyne.CurrentApp().Driver().AllWindows()[0].Canvas())
	dialog.Resize(fyne.NewSize(400, 300))

	addBtn := widget.NewButton("Add", func() {
		if algorithmSelect.Selected != "" {
			lp.addLayer(nameEntry.Text, algorithmSelect.Selected, regionSelect.Selected)
			dialog.Hide()
		}
	})
	addBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		dialog.Hide()
	})

	buttons := container.NewHBox(addBtn, cancelBtn)
	content.Add(buttons)
	dialog.Show()
}

func (lp *LeftPanel) showAddAlgorithmDialog() {
	algorithmSelect := widget.NewSelect(lp.getAlgorithmOptions(), nil)
	algorithmSelect.PlaceHolder = "Choose algorithm..."

	content := container.NewVBox(
		widget.NewLabel("Select Algorithm:"),
		algorithmSelect,
	)

	dialog := widget.NewModalPopUp(content, fyne.CurrentApp().Driver().AllWindows()[0].Canvas())
	dialog.Resize(fyne.NewSize(350, 200))

	addBtn := widget.NewButton("Add", func() {
		if algorithmSelect.Selected != "" {
			lp.addAlgorithm(algorithmSelect.Selected)
			dialog.Hide()
		}
	})
	addBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		dialog.Hide()
	})

	buttons := container.NewHBox(addBtn, cancelBtn)
	content.Add(buttons)
	dialog.Show()
}

func (lp *LeftPanel) getAlgorithmOptions() []string {
	categories := algorithms.GetAlgorithmsByCategory()
	var options []string
	for _, algs := range categories {
		options = append(options, algs...)
	}
	return options
}

func (lp *LeftPanel) getRegionOptions() []string {
	options := []string{"Whole Image"}
	selections := lp.regionManager.GetAllSelections()
	for _, selection := range selections {
		options = append(options, selection.ID)
	}
	return options
}

func (lp *LeftPanel) addLayer(name, algorithm, region string) {
	alg, exists := algorithms.Get(algorithm)
	if !exists {
		return
	}

	params := alg.GetDefaultParams()
	var regionID string
	if region != "Whole Image" {
		regionID = region
	}

	layerID, err := lp.pipeline.AddLayer(name, algorithm, params, regionID)
	if err != nil {
		lp.logger.Error("Failed to add layer", "error", err)
		return
	}

	layer := &LayerItem{
		ID:         layerID,
		Name:       name,
		Algorithm:  algorithm,
		RegionID:   regionID,
		BlendMode:  "Normal",
		Visible:    true,
		Expanded:   false,
		Parameters: params,
	}

	lp.layerData = append(lp.layerData, layer)
	lp.layerList.Refresh()
}

func (lp *LeftPanel) addAlgorithm(algorithm string) {
	alg, exists := algorithms.Get(algorithm)
	if !exists {
		return
	}

	params := alg.GetDefaultParams()

	err := lp.pipeline.AddStep(algorithm, params)
	if err != nil {
		lp.logger.Error("Failed to add algorithm", "error", err)
		return
	}

	sequence := &SequenceItem{
		Algorithm:  algorithm,
		Parameters: params,
		Expanded:   false,
	}

	lp.sequenceData = append(lp.sequenceData, sequence)
	lp.sequenceList.Refresh()
}

func (lp *LeftPanel) deleteLayer(index int) {
	if index < 0 || index >= len(lp.layerData) {
		return
	}
	lp.layerData = append(lp.layerData[:index], lp.layerData[index+1:]...)
	lp.layerList.Refresh()
	lp.selectedIndex = -1
	lp.updateParametersArea()
}

func (lp *LeftPanel) deleteSequenceItem(index int) {
	if index < 0 || index >= len(lp.sequenceData) {
		return
	}
	lp.sequenceData = append(lp.sequenceData[:index], lp.sequenceData[index+1:]...)
	lp.sequenceList.Refresh()
	lp.selectedIndex = -1
	lp.updateParametersArea()
}

func (lp *LeftPanel) updateParametersArea() {
	lp.currentParams.RemoveAll()

	if lp.selectedIndex < 0 {
		lp.currentParams.Add(widget.NewLabel("Select an item above to edit parameters"))
		lp.currentParams.Refresh()
		return
	}

	var algorithm algorithms.Algorithm
	var params map[string]interface{}
	var exists bool

	if lp.isLayerMode {
		if lp.selectedIndex >= len(lp.layerData) {
			return
		}
		layer := lp.layerData[lp.selectedIndex]
		algorithm, exists = algorithms.Get(layer.Algorithm)
		params = layer.Parameters
	} else {
		if lp.selectedIndex >= len(lp.sequenceData) {
			return
		}
		sequence := lp.sequenceData[lp.selectedIndex]
		algorithm, exists = algorithms.Get(sequence.Algorithm)
		params = sequence.Parameters
	}

	if !exists {
		lp.currentParams.Add(widget.NewLabel("Algorithm not found"))
		lp.currentParams.Refresh()
		return
	}

	// Create parameter widgets
	paramInfo := algorithm.GetParameterInfo()
	for _, param := range paramInfo {
		lp.createParameterWidget(param, params)
	}

	lp.currentParams.Refresh()
}

func (lp *LeftPanel) createParameterWidget(param algorithms.ParameterInfo, params map[string]interface{}) {
	lp.currentParams.Add(widget.NewLabel(param.Name + ":"))

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
		}

		lp.currentParams.Add(container.NewHBox(slider, valueLabel))

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
		}

		lp.currentParams.Add(container.NewHBox(slider, valueLabel))

	case "bool":
		check := widget.NewCheck("", func(checked bool) {
			params[param.Name] = checked
		})
		if val, ok := params[param.Name].(bool); ok {
			check.SetChecked(val)
		}
		lp.currentParams.Add(check)

	case "enum":
		selectWidget := widget.NewSelect(param.Options, func(selected string) {
			params[param.Name] = selected
		})
		if val, ok := params[param.Name].(string); ok {
			selectWidget.SetSelected(val)
		}
		lp.currentParams.Add(selectWidget)
	}

	lp.currentParams.Add(widget.NewLabel(param.Description))
}

func (lp *LeftPanel) openImage() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		filepath := reader.URI().Path()
		mat, err := lp.loader.LoadImage(filepath)
		if err != nil {
			lp.logger.Error("Failed to load image", "error", err)
			return
		}
		defer mat.Close()

		if err := lp.imageData.SetOriginal(mat, filepath); err != nil {
			lp.logger.Error("Failed to set image", "error", err)
			return
		}

		if lp.onImageLoaded != nil {
			lp.onImageLoaded(filepath)
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	imageFilter := storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()
}

func (lp *LeftPanel) saveImage() {
	if !lp.imageData.HasImage() {
		return
	}

	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			return
		}
		defer writer.Close()

		filepath := writer.URI().Path()
		processed := lp.imageData.GetProcessed()
		defer processed.Close()

		if processed.Empty() {
			processed = lp.imageData.GetOriginal()
		}

		if err := lp.loader.SaveImage(processed, filepath); err != nil {
			lp.logger.Error("Failed to save image", "error", err)
			return
		}

		if lp.onImageSaved != nil {
			lp.onImageSaved(filepath)
		}
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	imageFilter := storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".tiff", ".tif"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()
}

func (lp *LeftPanel) GetContainer() fyne.CanvasObject {
	return lp.container
}

func (lp *LeftPanel) EnableProcessing() {
	lp.enabled = true
	lp.modeToggle.Enable()
	if lp.addLayerBtn != nil {
		lp.addLayerBtn.Enable()
	}
	if lp.addAlgBtn != nil {
		lp.addAlgBtn.Enable()
	}
}

func (lp *LeftPanel) Disable() {
	lp.enabled = false
	lp.modeToggle.Disable()
	if lp.addLayerBtn != nil {
		lp.addLayerBtn.Disable()
	}
	if lp.addAlgBtn != nil {
		lp.addAlgBtn.Disable()
	}
}

func (lp *LeftPanel) Reset() {
	lp.layerData = make([]*LayerItem, 0)
	lp.sequenceData = make([]*SequenceItem, 0)
	lp.selectedIndex = -1

	if lp.layerList != nil {
		lp.layerList.Refresh()
	}
	if lp.sequenceList != nil {
		lp.sequenceList.Refresh()
	}
	lp.updateParametersArea()
}

func (lp *LeftPanel) UpdateSelectionState(hasSelection bool) {
	// Update UI based on selection state
}

func (lp *LeftPanel) SetCallbacks(
	onModeChanged func(bool),
	onImageLoaded func(string),
	onImageSaved func(string),
	onReset func(),
) {
	lp.onModeChanged = onModeChanged
	lp.onImageLoaded = onImageLoaded
	lp.onImageSaved = onImageSaved
	lp.onReset = onReset
}
