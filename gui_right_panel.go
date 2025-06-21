package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func (ui *ImageRestorationUI) createRightPanel() fyne.CanvasObject {
	ui.imageInfoLabel = widget.NewRichText(&widget.TextSegment{
		Text:  "No image loaded",
		Style: widget.RichTextStyle{},
	})

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
