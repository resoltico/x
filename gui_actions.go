package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
)

func (ui *ImageRestorationUI) openImage() {
	ui.debugGUI.LogButtonClick("OPEN IMAGE")

	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		ui.debugGUI.LogFileOperation("open", reader.URI().Name())

		go func() {
			mat := gocv.IMRead(reader.URI().Path(), gocv.IMReadColor)
			defer mat.Close()

			if mat.Empty() {
				err := fmt.Errorf("failed to load image")
				ui.debugGUI.LogError(err)
				fyne.Do(func() {
					dialog.ShowError(err, ui.window)
				})
				return
			}

			size := mat.Size()
			ui.debugGUI.LogImageInfo(size[1], size[0], mat.Channels())

			ui.pipeline.ClearTransformations()

			clonedMat := mat.Clone()
			err = ui.pipeline.SetOriginalImage(clonedMat)
			if err != nil {
				clonedMat.Close()
				ui.debugGUI.LogError(err)
				fyne.Do(func() {
					dialog.ShowError(err, ui.window)
				})
				return
			}

			fyne.Do(func() {
				ui.updateUI()
				ui.updateWindowTitle(reader.URI().Name())

				ui.parametersContainer.Objects[0] = widget.NewLabel("Select a Transformation")
				ui.parametersContainer.Refresh()
				ui.transformationsList.UnselectAll()
			})

			ui.debugGUI.Log("Image loaded successfully")
		}()
	}, ui.window)
}

func (ui *ImageRestorationUI) saveImage() {
	ui.debugGUI.LogButtonClick("SAVE IMAGE")

	if !ui.pipeline.HasImage() {
		dialog.ShowInformation("No Image", "Please load an image first", ui.window)
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			ui.debugGUI.LogError(err)
			return
		}
		defer writer.Close()

		filename := writer.URI().Name()
		filePath := writer.URI().Path()

		ui.debugGUI.LogFileOperation("save", filename)

		ext := strings.ToLower(filepath.Ext(filename))
		ui.debugGUI.LogFileExtensionCheck(filename, ext, ext != "")

		if ext == "" {
			filePath = filePath + ".png"
			filename = filename + ".png"
			ui.debugGUI.Log("Added .png extension to filename")
		} else if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".tiff" && ext != ".tif" {
			filePath = strings.TrimSuffix(filePath, ext) + ".png"
			filename = strings.TrimSuffix(filename, ext) + ".png"
		}

		go func() {
			ui.debugGUI.Log("Forcing full reprocessing before save to ensure latest parameters")
			err := ui.pipeline.ProcessImage()
			if err != nil {
				ui.debugGUI.LogError(fmt.Errorf("failed to reprocess image before save: %w", err))
				fyne.Do(func() {
					dialog.ShowError(err, ui.window)
				})
				return
			}

			processedImage := ui.pipeline.GetProcessedImage()
			defer processedImage.Close()

			hasImage := !processedImage.Empty()
			ui.debugGUI.LogSaveOperation(filename, filepath.Ext(filename), hasImage)

			if !hasImage {
				ui.debugGUI.LogSaveResult(filename, false, "no processed image available")
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("no processed image available"), ui.window)
				})
				return
			}

			success := gocv.IMWrite(filePath, processedImage)
			if !success {
				err := fmt.Errorf("failed to write image to %s", filePath)
				ui.debugGUI.LogSaveResult(filename, false, err.Error())
				fyne.Do(func() {
					dialog.ShowError(err, ui.window)
				})
			} else {
				ui.debugGUI.LogSaveResult(filename, true, "")
				ui.debugGUI.Log("Image saved successfully")
			}
		}()
	}, ui.window)
}

func (ui *ImageRestorationUI) resetTransformations() {
	ui.debugGUI.LogButtonClick("Reset")

	ui.pipeline.ClearTransformations()
	fyne.Do(func() {
		ui.updateUI()
		ui.parametersContainer.Objects[0] = widget.NewLabel("Select a Transformation")
		ui.parametersContainer.Refresh()
		ui.transformationsList.UnselectAll()
	})
}
