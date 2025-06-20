package main

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"
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
}

func NewImageRestorationUI(window fyne.Window, config *DebugConfig) *ImageRestorationUI {
	ui := &ImageRestorationUI{
		window:      window,
		pipeline:    NewImagePipeline(config),
		debugGUI:    NewDebugGUI(config),
		debugRender: NewDebugRender(config),
	}
	return ui
}

func (ui *ImageRestorationUI) BuildUI() fyne.CanvasObject {
	toolbar := ui.createToolbar()
	leftPanel := ui.createLeftPanel()
	centerPanel := ui.createCenterPanel()
	rightPanel := ui.createRightPanel()

	mainContainer := container.NewBorder(
		toolbar,
		nil,
		leftPanel,
		rightPanel,
		centerPanel,
	)

	return mainContainer
}

func (ui *ImageRestorationUI) createToolbar() fyne.CanvasObject {
	openBtn := widget.NewButtonWithIcon("OPEN IMAGE", theme.FolderOpenIcon(), ui.openImage)
	openBtn.Importance = widget.HighImportance

	saveBtn := widget.NewButtonWithIcon("SAVE IMAGE", theme.DocumentSaveIcon(), ui.saveImage)
	saveBtn.Importance = widget.HighImportance

	resetBtn := widget.NewButtonWithIcon("Reset", theme.ViewRefreshIcon(), ui.resetTransformations)
	resetBtn.Importance = widget.HighImportance

	leftSection := container.NewHBox(openBtn, saveBtn, resetBtn)
	toolbar := container.NewBorder(nil, nil, leftSection, nil, nil)
	toolbarCard := container.NewPadded(toolbar)
	toolbarCard.Resize(fyne.NewSize(0, 50))

	return toolbarCard
}

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

func (ui *ImageRestorationUI) createCenterPanel() fyne.CanvasObject {
	ui.originalImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ui.originalImage.FillMode = canvas.ImageFillContain
	ui.originalImage.ScaleMode = canvas.ImageScaleSmooth
	ui.originalImage.Resize(fyne.NewSize(500, 400))

	ui.previewImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ui.previewImage.FillMode = canvas.ImageFillContain
	ui.previewImage.ScaleMode = canvas.ImageScaleSmooth
	ui.previewImage.Resize(fyne.NewSize(500, 400))

	ui.originalScroll = container.NewScroll(ui.originalImage)
	ui.originalScroll.Resize(fyne.NewSize(500, 400))

	ui.previewScroll = container.NewScroll(ui.previewImage)
	ui.previewScroll.Resize(fyne.NewSize(500, 400))

	makeHeader := func(text string) fyne.CanvasObject {
		bg := canvas.NewRectangle(&color.RGBA{R: 233, G: 208, B: 255, A: 255})
		bg.SetMinSize(fyne.NewSize(0, 24))
		lbl := canvas.NewText(text, color.Black)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		return container.NewMax(bg, container.NewCenter(lbl))
	}

	originalContainer := container.NewBorder(
		makeHeader("ORIGINAL"), nil, nil, nil,
		ui.originalScroll,
	)
	previewContainer := container.NewBorder(
		makeHeader("PREVIEW"), nil, nil, nil,
		ui.previewScroll,
	)

	imagesSplit := container.NewHSplit(originalContainer, previewContainer)
	imagesSplit.SetOffset(0.5)

	ui.transformationsList = widget.NewList(
		func() int { return len(ui.pipeline.transformations) },
		func() fyne.CanvasObject {
			return container.NewBorder(
				nil, nil, nil,
				widget.NewButtonWithIcon("", theme.ContentRemoveIcon(), nil),
				widget.NewLabel("Transformation"),
			)
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			borderContainer := obj.(*fyne.Container)
			label := borderContainer.Objects[0].(*widget.Label)
			removeBtn := borderContainer.Objects[1].(*widget.Button)

			if id < len(ui.pipeline.transformations) {
				label.SetText(ui.pipeline.transformations[id].Name())
				removeBtn.OnTapped = func() {
					ui.removeTransformation(id)
				}
			}
		},
	)
	ui.transformationsList.OnSelected = ui.onAppliedTransformationSelected

	transformationsListContainer := container.NewBorder(
		makeHeader("ACTIVE TRANSFORMATIONS"), nil, nil, nil,
		ui.transformationsList,
	)
	ui.parametersContainer = container.NewBorder(
		makeHeader("PARAMETERS"), nil, nil, nil,
		widget.NewLabel("Select a Transformation"),
	)

	bottomSplit := container.NewHSplit(transformationsListContainer, ui.parametersContainer)
	bottomSplit.SetOffset(0.5)

	centerPanel := container.NewVSplit(imagesSplit, bottomSplit)
	centerPanel.SetOffset(0.6)

	return centerPanel
}

func (ui *ImageRestorationUI) createRightPanel() fyne.CanvasObject {
	ui.imageInfoLabel = widget.NewRichText(&widget.TextSegment{
		Text:  "No image loaded",
		Style: widget.RichTextStyle{},
	})

	ui.psnrLabel = widget.NewLabel("PSNR: 33.14 dB")
	ui.psnrProgress = widget.NewProgressBar()
	ui.psnrProgress.Resize(fyne.NewSize(300, 20))

	ui.ssimLabel = widget.NewLabel("SSIM: 0.9674")
	ui.ssimProgress = widget.NewProgressBar()
	ui.ssimProgress.Resize(fyne.NewSize(300, 20))

	qualityContent := container.NewVBox(
		ui.psnrLabel,
		ui.psnrProgress,
		ui.ssimLabel,
		ui.ssimProgress,
	)
	qualityContent.Resize(fyne.NewSize(0, 120))

	makeHeader := func(text string) fyne.CanvasObject {
		bg := canvas.NewRectangle(&color.RGBA{R: 233, G: 208, B: 255, A: 255})
		bg.SetMinSize(fyne.NewSize(0, 24))
		lbl := canvas.NewText(text, color.Black)
		lbl.TextStyle = fyne.TextStyle{Bold: true}
		return container.NewMax(bg, container.NewCenter(lbl))
	}

	imageInfoContainer := container.NewBorder(
		makeHeader("IMAGE INFORMATION"),
		nil, nil, nil,
		ui.imageInfoLabel,
	)

	qualityContainer := container.NewBorder(
		makeHeader("QUALITY METRICS"),
		nil, nil, nil,
		qualityContent,
	)
	qualityContainer.Resize(fyne.NewSize(340, 150))

	rightPanel := container.NewVBox(imageInfoContainer, qualityContainer)
	rightPanel.Resize(fyne.NewSize(340, 0))

	return rightPanel
}

func (ui *ImageRestorationUI) openImage() {
	ui.debugGUI.LogButtonClick("OPEN IMAGE")

	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer func() {
			if err := reader.Close(); err != nil {
				ui.debugGUI.LogError(err)
			}
		}()

		ui.debugGUI.LogFileOperation("open", reader.URI().Name())

		// Use proper error boundaries
		go func() {
			defer func() {
				if r := recover(); r != nil {
					ui.debugGUI.Log(fmt.Sprintf("Panic in openImage: %v", r))
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("error loading image: %v", r), ui.window)
					})
				}
			}()

			mat := gocv.IMRead(reader.URI().Path(), gocv.IMReadColor)
			defer func() {
				if !mat.Empty() {
					mat.Close()
				}
			}()

			if mat.Empty() {
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("failed to load image"), ui.window)
				})
				return
			}

			size := mat.Size()
			ui.debugGUI.LogImageInfo(size[1], size[0], mat.Channels())

			ui.pipeline.ClearTransformations()

			if err := ui.pipeline.SetOriginalImage(mat); err != nil {
				ui.debugGUI.LogError(err)
				fyne.Do(func() {
					dialog.ShowError(err, ui.window)
				})
				return
			}

			fyne.Do(func() {
				ui.updateUI()
				ui.updateWindowTitle(reader.URI().Name())
				ui.parametersContainer.Objects[1] = widget.NewLabel("Select a Transformation")
				ui.parametersContainer.Refresh()
				ui.transformationsList.UnselectAll()
				ui.availableTransformationsList.UnselectAll()
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
		defer func() {
			if err := writer.Close(); err != nil {
				ui.debugGUI.LogError(err)
			}
		}()

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
			defer func() {
				if r := recover(); r != nil {
					ui.debugGUI.Log(fmt.Sprintf("Panic in saveImage: %v", r))
					fyne.Do(func() {
						dialog.ShowError(fmt.Errorf("error saving image: %v", r), ui.window)
					})
				}
			}()

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
		ui.parametersContainer.Objects[1] = widget.NewLabel("Select a Transformation")
		ui.parametersContainer.Refresh()
		ui.transformationsList.UnselectAll()
	})
}

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
		defer func() {
			if r := recover(); r != nil {
				ui.debugGUI.Log(fmt.Sprintf("Panic in transformation: %v", r))
				fyne.Do(func() {
					dialog.ShowError(fmt.Errorf("error applying transformation: %v", r), ui.window)
				})
			}
		}()

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
		})

		ui.debugGUI.LogListUnselect("available transformations")
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
		defer func() {
			if r := recover(); r != nil {
				ui.debugGUI.Log(fmt.Sprintf("Panic in removeTransformation: %v", r))
			}
		}()

		err := ui.pipeline.RemoveTransformation(id)
		if err != nil {
			ui.debugGUI.LogError(err)
			return
		}

		fyne.Do(func() {
			ui.transformationsList.UnselectAll()
			ui.parametersContainer.Objects[1] = widget.NewLabel("Select a Transformation")
			ui.parametersContainer.Refresh()
			ui.updateUI()
		})
	}()
}

func (ui *ImageRestorationUI) showTransformationParameters(transformation Transformation) {
	parametersWidget := transformation.GetParametersWidget(ui.onParameterChanged)
	fyne.Do(func() {
		ui.parametersContainer.Objects[1] = parametersWidget
		ui.parametersContainer.Refresh()
	})
}

func (ui *ImageRestorationUI) onParameterChanged() {
	ui.debugGUI.LogUIEvent("onParameterChanged called")

	go func() {
		defer func() {
			if r := recover(); r != nil {
				ui.debugGUI.Log(fmt.Sprintf("Panic in onParameterChanged: %v", r))
			}
		}()

		if ui.pipeline.HasImage() {
			err := ui.pipeline.ProcessPreview()
			if err != nil {
				ui.debugGUI.LogError(err)
				return
			}
		}

		fyne.Do(func() {
			ui.updateUI()
		})
	}()
}

func (ui *ImageRestorationUI) updateUI() {
	ui.updateImageDisplay()
	ui.updateImageInfo()
	ui.updateQualityMetrics()
	ui.transformationsList.Refresh()
}

func (ui *ImageRestorationUI) updateImageDisplay() {
	defer func() {
		if r := recover(); r != nil {
			ui.debugGUI.Log(fmt.Sprintf("Panic in updateImageDisplay: %v", r))
		}
	}()

	ui.debugGUI.LogUIEvent("updateImageDisplay called")

	if ui.pipeline.HasImage() && !ui.pipeline.originalImage.Empty() {
		ui.debugGUI.LogUIEvent("updateImageDisplay: converting original image")

		originalImg, err := ui.pipeline.originalImage.ToImage()
		if err != nil {
			ui.debugGUI.LogImageConversion("original", false, err.Error())
			return
		}
		ui.debugGUI.LogImageConversion("original", true, "")

		previewMat := ui.pipeline.GetPreviewImage()
		defer previewMat.Close()

		if previewMat.Empty() {
			ui.debugGUI.LogUIEvent("updateImageDisplay: preview image is empty")
			return
		}

		var previewImg image.Image
		originalChannels := ui.pipeline.originalImage.Channels()
		previewChannels := previewMat.Channels()

		if originalChannels != previewChannels {
			ui.debugGUI.LogImageFormatChange("preview", originalChannels, previewChannels)

			if previewChannels == 1 && originalChannels == 3 {
				previewColor := gocv.NewMat()
				defer previewColor.Close()
				gocv.CvtColor(previewMat, &previewColor, gocv.ColorGrayToBGR)

				previewImg, err = previewColor.ToImage()
				if err != nil {
					ui.debugGUI.LogImageConversion("preview", false, err.Error())
					return
				}
			} else {
				previewImg, err = previewMat.ToImage()
				if err != nil {
					ui.debugGUI.LogImageConversion("preview", false, err.Error())
					return
				}
			}
		} else {
			previewImg, err = previewMat.ToImage()
			if err != nil {
				ui.debugGUI.LogImageConversion("preview", false, err.Error())
				return
			}
		}

		ui.debugGUI.LogImageConversion("preview", true, "")

		ui.originalImage.Image = originalImg
		ui.previewImage.Image = previewImg

		ui.originalImage.Refresh()
		ui.previewImage.Refresh()

		ui.debugGUI.LogCanvasRefresh("originalImage")
		ui.debugGUI.LogCanvasRefresh("previewImage")
		ui.debugGUI.LogUIEvent("updateImageDisplay: completed successfully")
	}
}

func (ui *ImageRestorationUI) updateImageInfo() {
	if ui.pipeline.HasImage() && !ui.pipeline.originalImage.Empty() {
		size := ui.pipeline.originalImage.Size()
		channels := ui.pipeline.originalImage.Channels()

		info := fmt.Sprintf("Size: %dx%d\nChannels: %d", size[1], size[0], channels)
		ui.imageInfoLabel.ParseMarkdown(info)
	}
}

func (ui *ImageRestorationUI) updateQualityMetrics() {
	if len(ui.pipeline.transformations) > 0 {
		psnr := ui.pipeline.CalculatePSNR()
		ssim := ui.pipeline.CalculateSSIM()

		ui.debugGUI.LogQualityMetricsUpdate(psnr, ssim, true)

		ui.psnrLabel.SetText(fmt.Sprintf("PSNR: %.2f dB", psnr))
		ui.psnrProgress.SetValue(psnr / 50.0)

		ui.ssimLabel.SetText(fmt.Sprintf("SSIM: %.4f", ssim))
		ui.ssimProgress.SetValue(ssim)
	} else {
		ui.debugGUI.LogQualityMetricsUpdate(0, 0, false)

		ui.psnrLabel.SetText("PSNR: 33.14 dB")
		ui.psnrProgress.SetValue(0)
		ui.ssimLabel.SetText("SSIM: 0.9674")
		ui.ssimProgress.SetValue(0)
	}
}

func (ui *ImageRestorationUI) updateWindowTitle(filename string) {
	if filename != "" {
		ui.window.SetTitle(fmt.Sprintf("Image Restoration Suite - %s", filename))
	} else {
		ui.window.SetTitle("Image Restoration Suite")
	}
}
