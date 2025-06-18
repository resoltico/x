package main

import (
	"fmt"
	"image"
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

func NewImageRestorationUI(window fyne.Window) *ImageRestorationUI {
	ui := &ImageRestorationUI{
		window:      window,
		pipeline:    NewImagePipeline(),
		debugGUI:    NewDebugGUI(),
		debugRender: NewDebugRender(),
	}
	return ui
}

func (ui *ImageRestorationUI) BuildUI() fyne.CanvasObject {
	// Create main layout
	toolbar := ui.createToolbar()
	leftPanel := ui.createLeftPanel()
	centerPanel := ui.createCenterPanel()
	rightPanel := ui.createRightPanel()

	// Main container with fixed structure
	mainContainer := container.NewBorder(
		toolbar, // top
		nil,     // bottom
		leftPanel,
		rightPanel,
		centerPanel,
	)

	return mainContainer
}

func (ui *ImageRestorationUI) createToolbar() fyne.CanvasObject {
	// File operations
	openBtn := widget.NewButtonWithIcon("OPEN IMAGE", theme.FolderOpenIcon(), ui.openImage)
	openBtn.Importance = widget.HighImportance

	saveBtn := widget.NewButtonWithIcon("SAVE IMAGE", theme.DocumentSaveIcon(), ui.saveImage)
	saveBtn.Importance = widget.HighImportance

	resetBtn := widget.NewButtonWithIcon("Reset", theme.ViewRefreshIcon(), ui.resetTransformations)
	resetBtn.Importance = widget.HighImportance

	leftSection := container.NewHBox(openBtn, saveBtn, resetBtn)

	toolbar := container.NewBorder(
		nil, nil,
		leftSection,
		nil,
		nil,
	)

	toolbarCard := container.NewPadded(toolbar)
	// Fixed toolbar height
	toolbarCard.Resize(fyne.NewSize(0, 50))

	return toolbarCard
}

func (ui *ImageRestorationUI) createLeftPanel() fyne.CanvasObject {
	// Transformations list
	transformations := []string{"2D Otsu"}

	ui.availableTransformationsList = widget.NewList(
		func() int { return len(transformations) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Transformation")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(transformations[id])
		},
	)

	ui.availableTransformationsList.OnSelected = ui.onTransformationSelected

	transformationsCard := container.NewBorder(
		widget.NewCard("", "TRANSFORMATIONS", ui.availableTransformationsList),
		nil, nil, nil,
	)

	leftPanel := container.NewVBox(transformationsCard)
	// Fixed left panel width
	leftPanel.Resize(fyne.NewSize(300, 0))

	return leftPanel
}

func (ui *ImageRestorationUI) createCenterPanel() fyne.CanvasObject {
	// Image display area with fixed constraints
	ui.originalImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ui.originalImage.FillMode = canvas.ImageFillContain
	ui.originalImage.ScaleMode = canvas.ImageScaleSmooth
	// Fixed image canvas size - prevents expansion
	ui.originalImage.Resize(fyne.NewSize(500, 400))

	ui.previewImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ui.previewImage.FillMode = canvas.ImageFillContain
	ui.previewImage.ScaleMode = canvas.ImageScaleSmooth
	// Fixed image canvas size - prevents expansion
	ui.previewImage.Resize(fyne.NewSize(500, 400))

	// Scroll containers with size locks
	ui.originalScroll = container.NewScroll(ui.originalImage)
	ui.originalScroll.Resize(fyne.NewSize(500, 400))

	ui.previewScroll = container.NewScroll(ui.previewImage)
	ui.previewScroll.Resize(fyne.NewSize(500, 400))

	originalContainer := container.NewBorder(
		widget.NewCard("", "Original", nil),
		nil, nil, nil,
		ui.originalScroll,
	)

	previewContainer := container.NewBorder(
		widget.NewCard("", "Preview", nil),
		nil, nil, nil,
		ui.previewScroll,
	)

	// Fixed split layout
	imagesSplit := container.NewHSplit(originalContainer, previewContainer)
	imagesSplit.SetOffset(0.5)

	// Transformations list and parameters
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
		widget.NewCard("", "Transformations", nil),
		nil, nil, nil,
		ui.transformationsList,
	)

	ui.parametersContainer = container.NewBorder(
		widget.NewCard("", "Parameters", nil),
		nil, nil, nil,
		widget.NewLabel("Select a Transformation"),
	)

	// Fixed split layout
	bottomSplit := container.NewHSplit(transformationsListContainer, ui.parametersContainer)
	bottomSplit.SetOffset(0.5)

	// Fixed split layout
	centerPanel := container.NewVSplit(imagesSplit, bottomSplit)
	centerPanel.SetOffset(0.6)

	return centerPanel
}

func (ui *ImageRestorationUI) createRightPanel() fyne.CanvasObject {
	// Image information
	ui.imageInfoLabel = widget.NewRichText(&widget.TextSegment{
		Text:  "No image loaded",
		Style: widget.RichTextStyle{},
	})

	imageInfoCard := widget.NewCard("", "IMAGE INFORMATION", ui.imageInfoLabel)

	// Quality metrics with fixed text length
	ui.psnrLabel = widget.NewLabel("PSNR: 33.14 dB") // Fixed length placeholder
	ui.psnrProgress = widget.NewProgressBar()
	ui.psnrProgress.Resize(fyne.NewSize(300, 20))

	ui.ssimLabel = widget.NewLabel("SSIM: 0.9674") // Fixed length placeholder
	ui.ssimProgress = widget.NewProgressBar()
	ui.ssimProgress.Resize(fyne.NewSize(300, 20))

	qualityContent := container.NewVBox(
		ui.psnrLabel,
		ui.psnrProgress,
		ui.ssimLabel,
		ui.ssimProgress,
	)
	// Fixed height for quality content
	qualityContent.Resize(fyne.NewSize(0, 120))

	qualityCard := widget.NewCard("", "QUALITY METRICS", qualityContent)
	qualityCard.Resize(fyne.NewSize(340, 150)) // Fixed card size

	rightPanel := container.NewVBox(imageInfoCard, qualityCard)
	// Absolute fixed right panel width
	rightPanel.Resize(fyne.NewSize(340, 0))

	return rightPanel
}

func (ui *ImageRestorationUI) openImage() {
	if ui.debugGUI.IsEnabled() {
		ui.debugGUI.LogButtonClick("OPEN IMAGE")
	}

	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogFileOperation("open", reader.URI().Name())
		}

		// Load image using OpenCV
		mat := gocv.IMRead(reader.URI().Path(), gocv.IMReadColor)
		if mat.Empty() {
			err := fmt.Errorf("failed to load image")
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.LogError(err)
			}
			dialog.ShowError(err, ui.window)
			return
		}

		size := mat.Size()
		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogImageInfo(size[1], size[0], mat.Channels())
		}

		ui.pipeline.SetOriginalImage(mat)
		ui.updateUI()
		ui.updateWindowTitle(reader.URI().Name())

		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.Log("Image loaded successfully")
		}
	}, ui.window)
}

func (ui *ImageRestorationUI) saveImage() {
	if ui.debugGUI.IsEnabled() {
		ui.debugGUI.LogButtonClick("SAVE IMAGE")
	}

	if !ui.pipeline.initialized || ui.pipeline.originalImage.Empty() {
		dialog.ShowInformation("No Image", "Please load an image first", ui.window)
		return
	}

	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil || writer == nil {
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.LogError(err)
			}
			return
		}
		defer writer.Close()

		filename := writer.URI().Name()
		filePath := writer.URI().Path()

		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogFileOperation("save", filename)
		}

		// Check file extension and add .png if missing
		ext := strings.ToLower(filepath.Ext(filename))
		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogFileExtensionCheck(filename, ext, ext != "")
		}

		if ext == "" {
			filePath = filePath + ".png"
			filename = filename + ".png"
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.Log("Added .png extension to filename")
			}
		} else if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".tiff" && ext != ".tif" {
			filePath = strings.TrimSuffix(filePath, ext) + ".png"
			filename = strings.TrimSuffix(filename, ext) + ".png"
		}

		processedImage := ui.pipeline.GetProcessedImage()
		hasImage := !processedImage.Empty()
		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogSaveOperation(filename, filepath.Ext(filename), hasImage)
		}

		if !hasImage {
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.LogSaveResult(filename, false, "no processed image available")
			}
			dialog.ShowError(fmt.Errorf("no processed image available"), ui.window)
			return
		}

		success := gocv.IMWrite(filePath, processedImage)
		if !success {
			err := fmt.Errorf("failed to write image to %s", filePath)
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.LogSaveResult(filename, false, err.Error())
			}
			dialog.ShowError(err, ui.window)
		} else {
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.LogSaveResult(filename, true, "")
				ui.debugGUI.Log("Image saved successfully")
			}
		}
	}, ui.window)
}

func (ui *ImageRestorationUI) resetTransformations() {
	if ui.debugGUI.IsEnabled() {
		ui.debugGUI.LogButtonClick("Reset")
	}

	ui.pipeline.ClearTransformations()
	ui.updateUI()
	ui.parametersContainer.Objects[0] = widget.NewLabel("Select a Transformation")
	ui.parametersContainer.Refresh()
}

func (ui *ImageRestorationUI) onTransformationSelected(id widget.ListItemID) {
	if ui.debugGUI.IsEnabled() {
		ui.debugGUI.LogListSelection("available transformations", int(id), "2D Otsu")
	}

	switch id {
	case 0: // 2D Otsu
		transformation := NewTwoDOtsu()
		ui.pipeline.AddTransformation(transformation)

		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogTransformation(transformation.Name(), transformation.GetParameters())
			ui.debugGUI.LogTransformationApplication(transformation.Name(), true)
		}

		ui.updateUI()

		// Clear the selection so it can be clicked again
		ui.availableTransformationsList.UnselectAll()
		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogListUnselect("available transformations")
		}
	}
}

func (ui *ImageRestorationUI) onAppliedTransformationSelected(id widget.ListItemID) {
	if id < len(ui.pipeline.transformations) {
		transformation := ui.pipeline.transformations[id]
		ui.showTransformationParameters(transformation)
	}
}

func (ui *ImageRestorationUI) removeTransformation(id int) {
	ui.pipeline.RemoveTransformation(id)
	ui.updateUI()
}

func (ui *ImageRestorationUI) showTransformationParameters(transformation Transformation) {
	parametersWidget := transformation.GetParametersWidget(ui.onParameterChanged)
	ui.parametersContainer.Objects[0] = parametersWidget
	ui.parametersContainer.Refresh()
}

func (ui *ImageRestorationUI) onParameterChanged() {
	ui.updateUI()
}

func (ui *ImageRestorationUI) updateUI() {
	ui.updateImageDisplay()
	ui.updateImageInfo()
	ui.updateQualityMetrics()
	// Selective refresh - only refresh list content, not layout
	ui.transformationsList.Refresh()
}

func (ui *ImageRestorationUI) updateImageDisplay() {
	if ui.debugGUI.IsEnabled() {
		ui.debugGUI.LogUIEvent("updateImageDisplay called")
	}

	if ui.pipeline.initialized && !ui.pipeline.originalImage.Empty() {
		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogUIEvent("updateImageDisplay: converting original image")
		}

		// Convert original image
		originalImg, err := ui.pipeline.originalImage.ToImage()
		if err != nil {
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.LogImageConversion("original", false, err.Error())
			}
			return
		}
		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogImageConversion("original", true, "")
		}

		// Convert processed image and handle format consistency
		processedMat := ui.pipeline.GetProcessedImage()
		if processedMat.Empty() {
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.LogUIEvent("updateImageDisplay: processed image is empty")
			}
			return
		}

		var processedImg image.Image
		originalChannels := ui.pipeline.originalImage.Channels()
		processedChannels := processedMat.Channels()

		if originalChannels != processedChannels {
			if ui.debugGUI.IsEnabled() {
				ui.debugGUI.LogImageFormatChange("processed", originalChannels, processedChannels)
			}
			if ui.debugRender.IsEnabled() {
				ui.debugRender.Log("FORCING CONSISTENT FORMAT: Converting grayscale to RGBA at Go image level")
			}

			// Convert grayscale back to RGBA at the Go image level
			processedColor := gocv.NewMat()
			defer processedColor.Close()
			gocv.CvtColor(processedMat, &processedColor, gocv.ColorGrayToBGR)

			var err error
			processedImg, err = processedColor.ToImage()
			if err != nil {
				if ui.debugGUI.IsEnabled() {
					ui.debugGUI.LogImageConversion("processed_rgba_converted", false, err.Error())
				}
				return
			}
			if ui.debugRender.IsEnabled() {
				ui.debugRender.Log("SUCCESS: Processed image converted to RGBA format for consistent GPU texture handling")
			}
		} else {
			var err error
			processedImg, err = processedMat.ToImage()
			if err != nil {
				if ui.debugGUI.IsEnabled() {
					ui.debugGUI.LogImageConversion("processed", false, err.Error())
				}
				return
			}
		}
		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogImageConversion("processed", true, "")
		}

		// Update only image content, not canvas properties
		ui.originalImage.Image = originalImg
		ui.previewImage.Image = processedImg

		// Selective refresh - only refresh image content
		ui.originalImage.Refresh()
		ui.previewImage.Refresh()

		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogCanvasRefresh("originalImage")
			ui.debugGUI.LogCanvasRefresh("previewImage")
			ui.debugGUI.LogUIEvent("updateImageDisplay: completed successfully")
		}
	}
}

func (ui *ImageRestorationUI) updateImageInfo() {
	if ui.pipeline.initialized && !ui.pipeline.originalImage.Empty() {
		size := ui.pipeline.originalImage.Size()
		channels := ui.pipeline.originalImage.Channels()

		info := fmt.Sprintf("Size: %dx%d\nChannels: %d", size[1], size[0], channels)
		ui.imageInfoLabel.ParseMarkdown(info)
	}
}

func (ui *ImageRestorationUI) updateQualityMetrics() {
	if ui.debugGUI.IsEnabled() {
		// Log panel positions before update
		leftPos := ui.window.Content().(*fyne.Container).Objects[1].Position()
		leftSize := ui.window.Content().(*fyne.Container).Objects[1].Size()
		ui.debugGUI.LogLayoutPositions("leftPanel", leftPos, leftSize)

		rightPos := ui.window.Content().(*fyne.Container).Objects[3].Position()
		rightSize := ui.window.Content().(*fyne.Container).Objects[3].Size()
		ui.debugGUI.LogLayoutPositions("rightPanel", rightPos, rightSize)
	}

	if len(ui.pipeline.transformations) > 0 {
		// Calculate PSNR and SSIM
		psnr := ui.pipeline.CalculatePSNR()
		ssim := ui.pipeline.CalculateSSIM()

		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogQualityMetricsUpdate(psnr, ssim, true)
		}

		ui.psnrLabel.SetText(fmt.Sprintf("PSNR: %.2f dB", psnr))
		ui.psnrProgress.SetValue(psnr / 50.0) // Normalize to 0-1 range

		ui.ssimLabel.SetText(fmt.Sprintf("SSIM: %.4f", ssim))
		ui.ssimProgress.SetValue(ssim) // SSIM is already 0-1 range
	} else {
		if ui.debugGUI.IsEnabled() {
			ui.debugGUI.LogQualityMetricsUpdate(0, 0, false)
		}

		ui.psnrLabel.SetText("PSNR: 33.14 dB") // Keep same text length
		ui.psnrProgress.SetValue(0)
		ui.ssimLabel.SetText("SSIM: 0.9674") // Keep same text length
		ui.ssimProgress.SetValue(0)
	}

	if ui.debugGUI.IsEnabled() {
		// Log panel positions after update
		rightPos := ui.window.Content().(*fyne.Container).Objects[3].Position()
		rightSize := ui.window.Content().(*fyne.Container).Objects[3].Size()
		ui.debugGUI.LogLayoutPositions("rightPanel_after", rightPos, rightSize)
	}
}

func (ui *ImageRestorationUI) updateWindowTitle(filename string) {
	if filename != "" {
		ui.window.SetTitle(fmt.Sprintf("Image Restoration Suite - %s", filename))
	} else {
		ui.window.SetTitle("Image Restoration Suite")
	}
}
