package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (ui *ImageRestorationUI) createLeftPanel() fyne.CanvasObject {
	transformations := []string{"2D Otsu", "Lanczos4 Scaling"}

	ui.availableTransformationsList = widget.NewList(
		func() int { return len(transformations) },
		func() fyne.CanvasObject { return widget.NewLabel("Transformation") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(transformations[id])
		},
	)
	ui.availableTransformationsList.OnSelected = ui.onTransformationSelected

	scrollableList := container.NewVScroll(ui.availableTransformationsList)
	scrollableList.SetMinSize(fyne.NewSize(200, 200))

	headerBg := canvas.NewRectangle(&color.RGBA{R: 233, G: 208, B: 255, A: 255})
	headerBg.SetMinSize(fyne.NewSize(200, 24))

	headerLabel := canvas.NewText("TRANSFORMATIONS", color.Black)
	headerLabel.TextStyle = fyne.TextStyle{Bold: true}

	header := container.NewMax(headerBg, container.NewCenter(headerLabel))

	content := container.NewVBox(header, scrollableList)

	leftPanel := container.NewVBox(content)
	leftPanel.Resize(fyne.NewSize(200, 0))

	return leftPanel
}
