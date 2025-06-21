package main

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

	updateMutex       sync.Mutex
	lastUpdateTime    time.Time
	processingUpdate  bool
	parameterDebounce time.Duration
}

func NewImageRestorationUI(window fyne.Window, config *DebugConfig) *ImageRestorationUI {
	ui := &ImageRestorationUI{
		window:            window,
		pipeline:          NewImagePipeline(config),
		debugGUI:          NewDebugGUI(config),
		debugRender:       NewDebugRender(config),
		parameterDebounce: 200 * time.Millisecond,
	}
	return ui
}

func (ui *ImageRestorationUI) BuildUI() fyne.CanvasObject {
	// Create main layout
	toolbar := ui.createToolbar()
	leftPanel := ui.createLeftPanel()
	centerPanel := ui.createCenterPanel()
	rightPanel := ui.createRightPanel()

	// Main container with structure
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
	// Update transformation list
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

	// Header background with color and height
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
	// Image display area with constraints
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

	// Header styling function
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
	// Image information
	ui.imageInfoLabel = widget.NewRichText(&widget.TextSegment{
		Text:  "No image loaded",
		Style: widget.RichTextStyle{},
	})

	// Quality metrics
	ui.psnrLabel = widget.NewLabel("PSNR: --")
	ui.psnrProgress = widget.NewProgressBar()
	ui.psnrProgress.Resize(fyne.NewSize(300, 20))

	ui.ssimLabel = widget.NewLabel("SSIM: --")
	ui.ssimProgress = widget.NewProgressBar()
	ui.ssimProgress.Resize(fyne.NewSize(300, 20))

	qualityContent := container.NewVBox(
		ui.psnrLabel,
		ui.psnrProgress,
		ui.ssimLabel,
		ui.ssimProgress,
	)
	qualityContent.Resize(fyne.NewSize(0, 120))

	// Header builder
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
		defer reader.Close()

		ui.debugGUI.LogFileOperation("open", reader.URI().Name())

		// Run in goroutine to prevent UI blocking
		go func() {
			// Load image using OpenCV with proper cleanup
			mat := gocv.IMRead(reader.URI().Path(), gocv.IMReadColor)
			defer mat.Close() // Always close the loaded Mat

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

			// Clean up existing transformations before setting new image
			ui.pipeline.ClearTransformations()

			// Clone the Mat before passing to pipeline since we defer Close above
			clonedMat := mat.Clone()
			err = ui.pipeline.SetOriginalImage(clonedMat)
			if err != nil {
				clonedMat.Close() // Clean up cloned Mat on error
				ui.debugGUI.LogError(err)
				// Show error in UI thread
				fyne.Do(func() {
					dialog.ShowError(err, ui.window)
				})
				return
			}

			// Update UI in main thread
			fyne.Do(func() {
				ui.updateUI()
				ui.updateWindowTitle(reader.URI().Name())

				// Reset parameters panel and clear list selections when new image is loaded
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

		// Validate and normalize file extension
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

		// Run in goroutine to prevent UI blocking
		go func() {
			// Force reprocessing of the full image to ensure latest parameters are applied
			ui.debugGUI.Log("Forcing full reprocessing before save to ensure latest parameters")
			err := ui.pipeline.ProcessImage()
			if err != nil {
				ui.debugGUI.LogError(fmt.Errorf("failed to reprocess image before save: %w", err))
				fyne.Do(func() {
					dialog.ShowError(err, ui.window)
				})
				return
			}

			// Get thread-safe clone to prevent race conditions
			processedImage := ui.pipeline.GetProcessedImage()
			defer processedImage.Close() // Always close the cloned image

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

	// Check if image is loaded before allowing transformation selection
	if !ui.pipeline.HasImage() {
		ui.debugGUI.Log("Cannot apply transformation: no image loaded")
		dialog.ShowInformation("No Image", "Please load an image before applying transformations", ui.window)
		ui.availableTransformationsList.UnselectAll()
		return
	}

	// Run in goroutine to prevent UI blocking
	go func() {
		var transformation Transformation
		switch id {
		case 0: // 2D Otsu
			transformation = NewTwoDOtsu(&debugConfig)
		case 1: // Lanczos4 Scaling
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
			// Clear the selection so it can be clicked again
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
	// Run in goroutine to prevent UI blocking
	go func() {
		err := ui.pipeline.RemoveTransformation(id)
		if err != nil {
			ui.debugGUI.LogError(err)
			return
		}

		// Clear selection since the list has changed
		fyne.Do(func() {
			ui.transformationsList.UnselectAll()
			ui.parametersContainer.Objects[0] = widget.NewLabel("Select a Transformation")
			ui.parametersContainer.Refresh()
			ui.updateUI()
		})
	}()
}

func (ui *ImageRestorationUI) showTransformationParameters(transformation Transformation) {
	// Pass ui.onParameterChanged as the callback
	parametersWidget := transformation.GetParametersWidget(ui.onParameterChanged)
	fyne.Do(func() {
		ui.parametersContainer.Objects[0] = parametersWidget
		ui.parametersContainer.Refresh()
	})
}

func (ui *ImageRestorationUI) onParameterChanged() {
	ui.debugGUI.LogUIEvent("onParameterChanged called - triggering immediate preview update")

	// Prevent concurrent updates with simple mutex
	ui.updateMutex.Lock()
	if ui.processingUpdate {
		ui.updateMutex.Unlock()
		ui.debugGUI.LogUIEvent("onParameterChanged: update already in progress, skipping")
		return
	}
	ui.processingUpdate = true
	ui.updateMutex.Unlock()

	// Minimal debouncing - only prevent rapid successive calls
	now := time.Now()
	if now.Sub(ui.lastUpdateTime) < ui.parameterDebounce {
		ui.updateMutex.Lock()
		ui.processingUpdate = false
		ui.updateMutex.Unlock()
		ui.debugGUI.LogUIEvent("onParameterChanged: rate limited, scheduling delayed update")

		// Schedule a single delayed update
		go func() {
			time.Sleep(ui.parameterDebounce - now.Sub(ui.lastUpdateTime))
			ui.onParameterChanged() // Retry after delay
		}()
		return
	}

	ui.lastUpdateTime = now

	// Force immediate preview regeneration
	go func() {
		defer func() {
			ui.updateMutex.Lock()
			ui.processingUpdate = false
			ui.updateMutex.Unlock()
		}()

		ui.debugGUI.LogUIEvent("onParameterChanged: starting preview regeneration")

		if !ui.pipeline.HasImage() {
			ui.debugGUI.LogUIEvent("onParameterChanged: no image loaded")
			return
		}

		// Force preview regeneration by calling the pipeline method
		// that triggers processing with current parameters
		err := ui.pipeline.ForcePreviewRegeneration()
		if err != nil {
			ui.debugGUI.LogError(fmt.Errorf("failed to regenerate preview: %w", err))
			return
		}

		ui.debugGUI.LogUIEvent("onParameterChanged: preview regenerated successfully")

		// Update UI on main thread
		fyne.Do(func() {
			ui.updateImageDisplay()
			ui.updateQualityMetrics()
		})
	}()
}

// Add thread safety to UI updates
func (ui *ImageRestorationUI) updateUI() {
	ui.updateImageDisplay()
	ui.updateImageInfo()
	ui.updateQualityMetrics()
	// Selective refresh - only refresh list content, not layout
	ui.transformationsList.Refresh()
}

// Image display with error handling and thread safety
func (ui *ImageRestorationUI) updateImageDisplay() {
	ui.debugGUI.LogUIEvent("updateImageDisplay called")

	if ui.pipeline.HasImage() && !ui.pipeline.originalImage.Empty() {
		ui.debugGUI.LogUIEvent("updateImageDisplay: converting original image")

		// Get thread-safe clones to prevent race conditions
		originalMat := ui.pipeline.originalImage.Clone()
		defer originalMat.Close()

		originalImg, err := originalMat.ToImage()
		if err != nil {
			ui.debugGUI.LogImageConversion("original", false, err.Error())
			return
		}
		ui.debugGUI.LogImageConversion("original", true, "")
		ui.debugRender.LogImageProperties("original", originalImg)

		// Get thread-safe preview clone
		previewMat := ui.pipeline.GetPreviewImage()
		defer previewMat.Close() // Clean up cloned Mat

		if previewMat.Empty() {
			ui.debugGUI.LogUIEvent("updateImageDisplay: preview image is empty")
			return
		}

		var previewImg image.Image
		originalChannels := originalMat.Channels()
		previewChannels := previewMat.Channels()

		if originalChannels != previewChannels {
			ui.debugGUI.LogImageFormatChange("preview", originalChannels, previewChannels)

			if previewChannels == 1 && originalChannels == 3 {
				// Use GoCV API for conversion
				ui.debugRender.Log("Converting grayscale to color using GoCV API")

				previewColor := gocv.NewMat()
				defer previewColor.Close()
				err := gocv.CvtColor(previewMat, &previewColor, gocv.ColorGrayToBGR)
				if err != nil {
					ui.debugGUI.LogImageConversion("preview_cvt", false, err.Error())
					// Fallback: Manual conversion
					size := previewMat.Size()
					width, height := size[1], size[0]
					bounds := image.Rect(0, 0, width, height)
					manualImg := image.NewRGBA(bounds)

					for y := 0; y < height; y++ {
						for x := 0; x < width; x++ {
							grayVal := previewMat.GetUCharAt(y, x)
							manualImg.Set(x, y, color.RGBA{R: grayVal, G: grayVal, B: grayVal, A: 255})
						}
					}
					previewImg = manualImg
				} else {
					var convErr error
					previewImg, convErr = previewColor.ToImage()
					if convErr != nil {
						ui.debugGUI.LogImageConversion("preview_final", false, convErr.Error())
						return
					}
					ui.debugRender.Log("SUCCESS: GoCV conversion completed")
				}
			} else {
				// Use standard ToImage for other cases
				var err error
				previewImg, err = previewMat.ToImage()
				if err != nil {
					ui.debugGUI.LogImageConversion("preview", false, err.Error())
					return
				}
			}
		} else {
			var err error
			previewImg, err = previewMat.ToImage()
			if err != nil {
				ui.debugGUI.LogImageConversion("preview", false, err.Error())
				return
			}
		}

		ui.debugGUI.LogImageConversion("preview", true, "")
		ui.debugRender.LogImageProperties("preview", previewImg)

		// Validate images before setting
		if originalImg != nil && previewImg != nil {
			// Update only image content, not canvas properties
			ui.originalImage.Image = originalImg
			ui.previewImage.Image = previewImg

			// Selective refresh - only refresh image content
			ui.originalImage.Refresh()
			ui.previewImage.Refresh()

			ui.debugGUI.LogCanvasRefresh("originalImage")
			ui.debugGUI.LogCanvasRefresh("previewImage")
			ui.debugGUI.LogUIEvent("updateImageDisplay: completed successfully")
		} else {
			ui.debugGUI.LogUIEvent("updateImageDisplay: ERROR - one or both images are nil")
		}
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
		// Run quality calculations in goroutine to prevent UI blocking
		go func() {
			// Calculate metrics with error handling
			psnr := ui.pipeline.CalculatePSNR()
			ssim := ui.pipeline.CalculateSSIM()

			ui.debugGUI.LogQualityMetricsUpdate(psnr, ssim, true)

			// Update UI in main thread
			fyne.Do(func() {
				// Format with bounds checking
				if psnr >= 0 && psnr <= 100 {
					ui.psnrLabel.SetText(fmt.Sprintf("PSNR: %.2f dB", psnr))
					ui.psnrProgress.SetValue(psnr / 100.0) // Normalize to 0-1 range
				} else {
					ui.psnrLabel.SetText("PSNR: --")
					ui.psnrProgress.SetValue(0)
				}

				if ssim >= 0 && ssim <= 1 {
					ui.ssimLabel.SetText(fmt.Sprintf("SSIM: %.4f", ssim))
					ui.ssimProgress.SetValue(ssim) // SSIM is already 0-1 range
				} else {
					ui.ssimLabel.SetText("SSIM: --")
					ui.ssimProgress.SetValue(0)
				}
			})
		}()
	} else {
		ui.debugGUI.LogQualityMetricsUpdate(0, 0, false)

		ui.psnrLabel.SetText("PSNR: --")
		ui.psnrProgress.SetValue(0)
		ui.ssimLabel.SetText("SSIM: --")
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
