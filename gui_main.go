package main

import (
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type ImageRestorationUI struct {
	window                       fyne.Window
	pipeline                     *ImagePipeline
	originalImage                *canvas.Image
	previewImage                 *canvas.Image
	originalScroll               *container.Scroll
	previewScroll                *container.Scroll
	transformationsList          *widget.List
	availableTransformationsList *widget.List
	parametersContainer          *fyne.Container
	imageInfoLabel               *widget.RichText
	psnrProgress                 *widget.ProgressBar
	ssimProgress                 *widget.ProgressBar
	psnrLabel                    *widget.Label
	ssimLabel                    *widget.Label
	debugGUI                     *DebugGUI
	debugRender                  *DebugRender

	updateMutex       sync.Mutex
	lastUpdateTime    time.Time
	processingUpdate  bool
	parameterDebounce time.Duration
}

func NewImageRestorationUI(window fyne.Window, config *DebugConfig) *ImageRestorationUI {
	return &ImageRestorationUI{
		window:            window,
		pipeline:          NewImagePipeline(config),
		debugGUI:          NewDebugGUI(config),
		debugRender:       NewDebugRender(config),
		parameterDebounce: 200 * time.Millisecond,
	}
}

func (ui *ImageRestorationUI) BuildUI() fyne.CanvasObject {
	toolbar := ui.createToolbar()
	leftPanel := ui.createLeftPanel()
	centerPanel := ui.createCenterPanel()
	rightPanel := ui.createRightPanel()

	return container.NewBorder(
		toolbar,
		nil,
		leftPanel,
		rightPanel,
		centerPanel,
	)
}
