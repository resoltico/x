package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func (ui *ImageRestorationUI) onTransformationSelected(id widget.ListItemID) {
	var transformationName string
	switch id {
	case 0:
		transformationName = "2D Otsu"
	case 1:
		transformationName = "Lanczos4 Scaling"
	default:
		return
	}

	ui.debugGUI.LogListSelection("available transformations", int(id), transformationName)

	if !ui.pipeline.HasImage() {
		ui.debugGUI.Log("Cannot apply transformation: no image loaded")
		dialog.ShowInformation("No Image", "Please load an image before applying transformations", ui.window)
		ui.availableTransformationsList.UnselectAll()
		return
	}

	go func() {
		var transformation Transformation
		switch id {
		case 0:
			transformation = NewTwoDOtsu(&debugConfig)
		case 1:
			transformation = NewLanczos4Transform(&debugConfig)
		default:
			return
		}

		err := ui.pipeline.AddTransformation(transformation)
		if err != nil {
			ui.debugGUI.LogError(err)
			fyne.Do(func() {
				dialog.ShowError(err, ui.window)
			})
			return
		}

		ui.debugGUI.LogTransformation(transformation.Name(), transformation.GetParameters())
		ui.debugGUI.LogTransformationApplication(transformation.Name(), true)

		fyne.Do(func() {
			ui.updateUI()
			ui.availableTransformationsList.UnselectAll()
			ui.debugGUI.LogListUnselect("available transformations")
		})
	}()
}

func (ui *ImageRestorationUI) onAppliedTransformationSelected(id widget.ListItemID) {
	if id < len(ui.pipeline.transformations) {
		transformation := ui.pipeline.transformations[id]
		ui.showTransformationParameters(transformation)
	}
}

func (ui *ImageRestorationUI) removeTransformation(id int) {
	go func() {
		err := ui.pipeline.RemoveTransformation(id)
		if err != nil {
			ui.debugGUI.LogError(err)
			return
		}

		fyne.Do(func() {
			ui.transformationsList.UnselectAll()
			ui.parametersContainer.Objects[0] = widget.NewLabel("Select a Transformation")
			ui.parametersContainer.Refresh()
			ui.updateUI()
		})
	}()
}

func (ui *ImageRestorationUI) showTransformationParameters(transformation Transformation) {
	parametersWidget := transformation.GetParametersWidget(ui.onParameterChanged)
	fyne.Do(func() {
		ui.parametersContainer.Objects[0] = parametersWidget
		ui.parametersContainer.Refresh()
	})
}
