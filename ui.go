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
}

func NewImageRestorationUI(window fyne.Window) *ImageRestorationUI {
	ui := &ImageRestorationUI{
		window:   window,
		pipeline: NewImagePipeline(),
		debugGUI: NewDebugGUI(),
	}
	return ui
}

func (ui *ImageRestorationUI) BuildUI() fyne.CanvasObject {
	// Create main layout
	toolbar := ui.createToolbar()
	leftPanel := ui.createLeftPanel()
	centerPanel := ui.createCenterPanel()
	rightPanel := ui.createRightPanel()

	// Main container
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
	leftPanel.Resize(fyne.NewSize(300, 0))

	return leftPanel
}

func (ui *ImageRestorationUI) createCenterPanel() fyne.CanvasObject {
	// Image display area
	ui.originalImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ui.originalImage.FillMode = canvas.ImageFillContain
	ui.originalImage.ScaleMode = canvas.ImageScaleSmooth

	ui.previewImage = canvas.NewImageFromImage(image.NewRGBA(image.Rect(0, 0, 1, 1)))
	ui.previewImage.FillMode = canvas.ImageFillContain
	ui.previewImage.ScaleMode = canvas.ImageScaleSmooth

	ui.originalScroll = container.NewScroll(ui.originalImage)
	ui.previewScroll = container.NewScroll(ui.previewImage)

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

	bottomSplit := container.NewHSplit(transformationsListContainer, ui.parametersContainer)
	bottomSplit.SetOffset(0.5)

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

	// Quality metrics
	ui.psnrLabel = widget.NewLabel("PSNR: will appear here during processing")
	ui.psnrProgress = widget.NewProgressBar()
	ui.psnrProgress.Hide()

	ui.ssimLabel = widget.NewLabel("SSIM: will appear here during processing")
	ui.ssimProgress = widget.NewProgressBar()
	ui.ssimProgress.Hide()

	qualityContent := container.NewVBox(
		ui.psnrLabel,
		ui.psnrProgress,
		ui.ssimLabel,
		ui.ssimProgress,
	)

	qualityCard := widget.NewCard("", "QUALITY METRICS", qualityContent)

	rightPanel := container.NewVBox(imageInfoCard, qualityCard)
	rightPanel.Resize(fyne.NewSize(350, 0))

	return rightPanel
}

func (ui *ImageRestorationUI) openImage() {
	ui.debugGUI.LogButtonClick("OPEN IMAGE")
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}
		defer reader.Close()

		ui.debugGUI.LogFileOperation("open", reader.URI().Name())

		// Load image using OpenCV
		mat := gocv.IMRead(reader.URI().Path(), gocv.IMReadColor)
		if mat.Empty() {
			err := fmt.Errorf("failed to load image")
			ui.debugGUI.LogError(err)
			dialog.ShowError(err, ui.window)
			return
		}

		size := mat.Size()
		ui.debugGUI.LogImageInfo(size[1], size[0], mat.Channels())
		ui.pipeline.SetOriginalImage(mat)
		ui.updateUI()
		ui.updateWindowTitle(reader.URI().Name())
		ui.debugGUI.Log("Image loaded successfully")
	}, ui.window)
}

func (ui *ImageRestorationUI) saveImage() {
	ui.debugGUI.LogButtonClick("SAVE IMAGE")
	if !ui.pipeline.initialized || ui.pipeline.originalImage.Empty() {
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

		// Check file extension and add .png if missing
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

		processedImage := ui.pipeline.GetProcessedImage()
		hasImage := !processedImage.Empty()
		ui.debugGUI.LogSaveOperation(filename, filepath.Ext(filename), hasImage)

		if !hasImage {
			ui.debugGUI.LogSaveResult(filename, false, "no processed image available")
			dialog.ShowError(fmt.Errorf("no processed image available"), ui.window)
			return
		}

		success := gocv.IMWrite(filePath, processedImage)
		if !success {
			err := fmt.Errorf("failed to write image to %s", filePath)
			ui.debugGUI.LogSaveResult(filename, false, err.Error())
			dialog.ShowError(err, ui.window)
		} else {
			ui.debugGUI.LogSaveResult(filename, true, "")
			ui.debugGUI.Log("Image saved successfully")
		}
	}, ui.window)
}

func (ui *ImageRestorationUI) resetTransformations() {
	ui.debugGUI.LogButtonClick("Reset")
	ui.pipeline.ClearTransformations()
	ui.updateUI()
	ui.parametersContainer.Objects[0] = widget.NewLabel("Select a Transformation")
	ui.parametersContainer.Refresh()
}

func (ui *ImageRestorationUI) onTransformationSelected(id widget.ListItemID) {
	ui.debugGUI.LogListSelection("available transformations", int(id), "2D Otsu")
	switch id {
	case 0: // 2D Otsu
		transformation := NewTwoDOtsu()
		ui.pipeline.AddTransformation(transformation)
		ui.debugGUI.LogTransformation(transformation.Name(), transformation.GetParameters())
		ui.debugGUI.LogTransformationApplication(transformation.Name(), true)
		ui.updateUI()

		// Clear the selection so it can be clicked again
		ui.availableTransformationsList.UnselectAll()
		ui.debugGUI.LogListUnselect("available transformations")
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
	ui.transformationsList.Refresh()
}

func (ui *ImageRestorationUI) updateImageDisplay() {
	ui.debugGUI.LogUIEvent("updateImageDisplay called")
	if ui.pipeline.initialized && !ui.pipeline.originalImage.Empty() {
		ui.debugGUI.LogUIEvent("updateImageDisplay: converting original image")

		// Convert original image
		originalImg, err := ui.pipeline.originalImage.ToImage()
		if err != nil {
			ui.debugGUI.LogImageConversion("original", false, err.Error())
			return
		}
		ui.debugGUI.LogImageConversion("original", true, "")

		// Convert processed image
		processedMat := ui.pipeline.GetProcessedImage()
		if processedMat.Empty() {
			ui.debugGUI.LogUIEvent("updateImageDisplay: processed image is empty")
			return
		}
		processedImg, err := processedMat.ToImage()
		if err != nil {
			ui.debugGUI.LogImageConversion("processed", false, err.Error())
			return
		}
		ui.debugGUI.LogImageConversion("processed", true, "")

		// Get actual image dimensions
		originalBounds := originalImg.Bounds()
		processedBounds := processedImg.Bounds()

		// Update images
		ui.originalImage.Image = originalImg
		ui.debugGUI.LogImageCanvasProperties("originalImage", 0, 0, originalBounds.Dx(), originalBounds.Dy())

		ui.previewImage.Image = processedImg
		ui.debugGUI.LogImageCanvasProperties("previewImage", 0, 0, processedBounds.Dx(), processedBounds.Dy())

		// Refresh images and scroll containers
		ui.originalImage.Refresh()
		ui.previewImage.Refresh()
		ui.originalScroll.Refresh()
		ui.previewScroll.Refresh()

		ui.debugGUI.LogCanvasRefresh("originalImage")
		ui.debugGUI.LogCanvasRefresh("previewImage")
		ui.debugGUI.LogUIEvent("updateImageDisplay: completed successfully")
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
	if len(ui.pipeline.transformations) > 0 {
		// Calculate PSNR and SSIM
		psnr := ui.pipeline.CalculatePSNR()
		ssim := ui.pipeline.CalculateSSIM()

		ui.psnrLabel.SetText(fmt.Sprintf("PSNR: %.2f dB", psnr))
		ui.psnrProgress.SetValue(psnr / 50.0) // Normalize to 0-1 range
		ui.psnrProgress.Show()

		ui.ssimLabel.SetText(fmt.Sprintf("SSIM: %.4f", ssim))
		ui.ssimProgress.SetValue(ssim) // SSIM is already 0-1 range
		ui.ssimProgress.Show()
	} else {
		ui.psnrLabel.SetText("PSNR: will appear here during processing")
		ui.psnrProgress.Hide()
		ui.ssimLabel.SetText("SSIM: will appear here during processing")
		ui.ssimProgress.Hide()
	}
}

func (ui *ImageRestorationUI) updateWindowTitle(filename string) {
	if filename != "" {
		ui.window.SetTitle(fmt.Sprintf("Image Restoration Suite - %s", filename))
	} else {
		ui.window.SetTitle("Image Restoration Suite")
	}
}
