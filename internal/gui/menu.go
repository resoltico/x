// Menu handler for application actions
package gui

import (
	"fmt"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	"advanced-image-processing/internal/core"
	"advanced-image-processing/internal/io"
)

// MenuHandler handles menu actions
type MenuHandler struct {
	window    fyne.Window
	imageData *core.ImageData
	loader    *io.ImageLoader
	logger    *slog.Logger

	onImageLoaded func(string)
	onImageSaved  func(string)
}

func NewMenuHandler(window fyne.Window, imageData *core.ImageData, loader *io.ImageLoader, logger *slog.Logger) *MenuHandler {
	return &MenuHandler{
		window:    window,
		imageData: imageData,
		loader:    loader,
		logger:    logger,
	}
}

func (mh *MenuHandler) GetMainMenu() *fyne.MainMenu {
	// File menu
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Image...", mh.openImage),
		fyne.NewMenuItem("Save Image...", mh.saveImage),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Exit", func() {
			mh.window.Close()
		}),
	)

	// Edit menu
	editMenu := fyne.NewMenu("Edit",
		fyne.NewMenuItem("Clear Selection", func() {
			// TODO: Clear selection
		}),
		fyne.NewMenuItem("Reset to Original", func() {
			if mh.imageData.HasImage() {
				mh.imageData.ResetToOriginal()
				mh.logger.Info("Reset to original image")
				if mh.onImageLoaded != nil {
					filepath := mh.imageData.GetFilepath()
					mh.onImageLoaded(filepath)
				}
			}
		}),
	)

	// Help menu
	helpMenu := fyne.NewMenu("Help",
		fyne.NewMenuItem("About", mh.showAbout),
	)

	return fyne.NewMainMenu(fileMenu, editMenu, helpMenu)
}

func (mh *MenuHandler) openImage() {
	mh.logger.Info("Opening file dialog for image selection")

	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			mh.showError("File Dialog Error", err)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		uri := reader.URI()
		filepath := uri.Path()

		mh.logger.Info("Loading selected image", "filepath", filepath)

		mat, err := mh.loader.LoadImage(filepath)
		if err != nil {
			mh.showError("Failed to Load Image", err)
			return
		}
		defer mat.Close()

		if err := core.ValidateImage(mat); err != nil {
			mh.showError("Invalid Image", err)
			return
		}

		if err := mh.imageData.SetOriginal(mat, filepath); err != nil {
			mh.showError("Failed to Set Image", err)
			return
		}

		mh.logger.Info("Image loaded successfully", "filepath", filepath)

		if mh.onImageLoaded != nil {
			mh.onImageLoaded(filepath)
		}

	}, mh.window)

	imageFilter := storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()
}

func (mh *MenuHandler) saveImage() {
	if !mh.imageData.HasImage() {
		mh.showError("No Image", fmt.Errorf("no image loaded to save"))
		return
	}

	mh.logger.Info("Opening file dialog for image saving")

	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			mh.showError("File Dialog Error", err)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		uri := writer.URI()
		filepath := uri.Path()

		mh.logger.Info("Saving image", "filepath", filepath)

		processed := mh.imageData.GetProcessed()
		defer processed.Close()

		if processed.Empty() {
			processed = mh.imageData.GetOriginal()
		}

		if err := mh.loader.SaveImage(processed, filepath); err != nil {
			mh.showError("Failed to Save Image", err)
			return
		}

		mh.logger.Info("Image saved successfully", "filepath", filepath)

		if mh.onImageSaved != nil {
			mh.onImageSaved(filepath)
		}

	}, mh.window)

	fileDialog.SetFileName("processed_image.png")
	imageFilter := storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".tiff", ".tif", ".bmp"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()
}

func (mh *MenuHandler) showAbout() {
	content := container.NewVBox(
		widget.NewLabel("Advanced Image Processing v2.0"),
		widget.NewSeparator(),
		widget.NewLabel("Optimized with real-time preview"),
		widget.NewLabel("Professional-grade image processing application"),
		widget.NewLabel("for historical documents with ROI selection,"),
		widget.NewLabel("optimized algorithms using standard APIs,"),
		widget.NewLabel("and real-time quality metrics."),
		widget.NewSeparator(),
		widget.NewLabel("Author: Ervins Strauhmanis"),
		widget.NewLabel("Built with Go, Fyne v2.6, and OpenCV 4.11"),
		widget.NewSeparator(),
		widget.NewLabel("License: MIT"),
	)

	aboutDialog := dialog.NewCustom("About", "Close", content, mh.window)
	aboutDialog.Resize(fyne.NewSize(400, 300))
	aboutDialog.Show()
}

func (mh *MenuHandler) showError(title string, err error) {
	mh.logger.Error(title, "error", err)
	dialog.ShowError(err, mh.window)
}

func (mh *MenuHandler) SetCallbacks(onImageLoaded, onImageSaved func(string)) {
	mh.onImageLoaded = onImageLoaded
	mh.onImageSaved = onImageSaved
}