package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"gocv.io/x/gocv"
)

func (ui *ImageRestorationUI) onParameterChanged() {
	ui.debugGUI.LogUIEvent("onParameterChanged called - triggering immediate preview update")

	ui.updateMutex.Lock()
	if ui.processingUpdate {
		ui.updateMutex.Unlock()
		ui.debugGUI.LogUIEvent("onParameterChanged: update already in progress, skipping")
		return
	}
	ui.processingUpdate = true
	ui.updateMutex.Unlock()

	now := time.Now()
	if now.Sub(ui.lastUpdateTime) < ui.parameterDebounce {
		ui.updateMutex.Lock()
		ui.processingUpdate = false
		ui.updateMutex.Unlock()
		ui.debugGUI.LogUIEvent("onParameterChanged: rate limited, scheduling delayed update")

		go func() {
			time.Sleep(ui.parameterDebounce - now.Sub(ui.lastUpdateTime))
			ui.onParameterChanged()
		}()
		return
	}

	ui.lastUpdateTime = now

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

		err := ui.pipeline.ForcePreviewRegeneration()
		if err != nil {
			ui.debugGUI.LogError(fmt.Errorf("failed to regenerate preview: %w", err))
			return
		}

		ui.debugGUI.LogUIEvent("onParameterChanged: preview regenerated successfully")

		fyne.Do(func() {
			ui.updateImageDisplay()
			ui.updateQualityMetrics()
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
	ui.debugGUI.LogUIEvent("updateImageDisplay called")

	if ui.pipeline.HasImage() && !ui.pipeline.originalImage.Empty() {
		ui.debugGUI.LogUIEvent("updateImageDisplay: converting original image")

		originalMat := ui.pipeline.originalImage.Clone()
		defer originalMat.Close()

		originalImg, err := originalMat.ToImage()
		if err != nil {
			ui.debugGUI.LogImageConversion("original", false, err.Error())
			return
		}
		ui.debugGUI.LogImageConversion("original", true, "")
		ui.debugRender.LogImageProperties("original", originalImg)

		previewMat := ui.pipeline.GetPreviewImage()
		defer previewMat.Close()

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
				ui.debugRender.Log("Converting grayscale to color using GoCV API")

				previewColor := gocv.NewMat()
				defer previewColor.Close()
				err := gocv.CvtColor(previewMat, &previewColor, gocv.ColorGrayToBGR)
				if err != nil {
					ui.debugGUI.LogImageConversion("preview_cvt", false, err.Error())
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

		if originalImg != nil && previewImg != nil {
			ui.originalImage.Image = originalImg
			ui.previewImage.Image = previewImg

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
		go func() {
			psnr := ui.pipeline.CalculatePSNR()
			ssim := ui.pipeline.CalculateSSIM()

			ui.debugGUI.LogQualityMetricsUpdate(psnr, ssim, true)

			fyne.Do(func() {
				if psnr >= 0 && psnr <= 100 {
					ui.psnrLabel.SetText(fmt.Sprintf("PSNR: %.2f dB", psnr))
					ui.psnrProgress.SetValue(psnr / 100.0)
				} else {
					ui.psnrLabel.SetText("PSNR: --")
					ui.psnrProgress.SetValue(0)
				}

				if ssim >= 0 && ssim <= 1 {
					ui.ssimLabel.SetText(fmt.Sprintf("SSIM: %.4f", ssim))
					ui.ssimProgress.SetValue(ssim)
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