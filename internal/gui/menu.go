// Author: Ervins Strauhmanis
// License: MIT

package gui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"

	"advanced-image-processing/internal/image_processing"
	"advanced-image-processing/internal/models"
	"advanced-image-processing/internal/presets"
)

// Menu handles menu bar and file operations
type Menu struct {
	loader    *image_processing.ImageLoader
	imageData *models.ImageData
	pipeline  *image_processing.Pipeline
	presetMgr *presets.Manager
	logger    *logrus.Logger

	// Callbacks
	onImageLoaded  func()
	onPresetLoaded func()
}

// NewMenu creates a new menu component
func NewMenu(loader *image_processing.ImageLoader, imageData *models.ImageData,
	pipeline *image_processing.Pipeline, presetMgr *presets.Manager, logger *logrus.Logger) *Menu {

	return &Menu{
		loader:    loader,
		imageData: imageData,
		pipeline:  pipeline,
		presetMgr: presetMgr,
		logger:    logger,
	}
}

// GetMainMenu returns the main menu bar
func (m *Menu) GetMainMenu() *fyne.MainMenu {
	return fyne.NewMainMenu(
		m.createFileMenu(),
		m.createPresetsMenu(),
		m.createHelpMenu(),
	)
}

// createFileMenu creates the File menu
func (m *Menu) createFileMenu() *fyne.Menu {
	openItem := fyne.NewMenuItem("Open Image...", func() {
		m.openImageDialog()
	})
	openItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyO, Modifier: fyne.KeyModifierShortcutDefault}

	saveItem := fyne.NewMenuItem("Save Image...", func() {
		m.saveImageDialog()
	})
	saveItem.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierShortcutDefault}

	return fyne.NewMenu("File",
		openItem,
		fyne.NewMenuItemSeparator(),
		saveItem,
	)
}

// createPresetsMenu creates the Presets menu
func (m *Menu) createPresetsMenu() *fyne.Menu {
	savePresetItem := fyne.NewMenuItem("Save Preset...", func() {
		m.savePresetDialog()
	})

	loadPresetItem := fyne.NewMenuItem("Load Preset...", func() {
		m.loadPresetDialog()
	})

	managePresetsItem := fyne.NewMenuItem("Manage Presets...", func() {
		m.managePresetsDialog()
	})

	return fyne.NewMenu("Presets",
		savePresetItem,
		loadPresetItem,
		fyne.NewMenuItemSeparator(),
		managePresetsItem,
	)
}

// createHelpMenu creates the Help menu
func (m *Menu) createHelpMenu() *fyne.Menu {
	aboutItem := fyne.NewMenuItem("About", func() {
		m.showAboutDialog()
	})

	return fyne.NewMenu("Help",
		aboutItem,
	)
}

// openImageDialog shows the open image dialog
func (m *Menu) openImageDialog() {
	fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			m.showErrorDialog("File Error", err)
			return
		}
		if reader == nil {
			return
		}
		defer reader.Close()

		// Load the image
		filepath := reader.URI().Path()
		m.logger.WithField("filepath", filepath).Info("Attempting to load image")

		mat, err := m.loader.LoadImage(filepath)
		if err != nil {
			m.showErrorDialog("Failed to load image", err)
			return
		}

		// Ensure we have a valid Mat before proceeding
		if mat.Empty() {
			mat.Close()
			m.showErrorDialog("Invalid image", fmt.Errorf("loaded image is empty: %s", filepath))
			return
		}

		// Validate the image before using it
		if err := m.loader.ValidateImage(mat); err != nil {
			mat.Close()
			m.showErrorDialog("Invalid image", err)
			return
		}

		// Log successful loading before setting the data
		m.logger.WithFields(logrus.Fields{
			"filepath": filepath,
			"width":    mat.Cols(),
			"height":   mat.Rows(),
			"channels": mat.Channels(),
		}).Info("Image loaded and validated successfully")

		// Set the image data - this is where the crash was occurring
		// The ImageData.SetOriginal method now has proper validation
		m.imageData.SetOriginal(mat)

		// Close the original Mat since SetOriginal clones it
		mat.Close()

		// Clear any existing transformations and reset to original
		m.pipeline.ClearSequence()

		// Trigger callback
		if m.onImageLoaded != nil {
			m.onImageLoaded()
		}

		m.logger.WithField("filepath", filepath).Info("Image loading completed successfully")
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	// Set file filters for image files
	imageFilter := storage.NewExtensionFileFilter([]string{".jpg", ".jpeg", ".png", ".tiff", ".tif", ".bmp", ".JPG", ".JPEG", ".PNG", ".TIFF", ".TIF", ".BMP"})
	fileDialog.SetFilter(imageFilter)
	fileDialog.Show()
}

// saveImageDialog shows the save image dialog
func (m *Menu) saveImageDialog() {
	if !m.imageData.HasImage() {
		m.showInfoDialog("No Image", "Please load an image first.")
		return
	}

	fileDialog := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			m.showErrorDialog("File Error", err)
			return
		}
		if writer == nil {
			return
		}
		defer writer.Close()

		// Get processed image, or original if no processing applied
		var imageToSave = m.imageData.GetProcessed()
		defer imageToSave.Close()

		if imageToSave.Empty() {
			imageToSave.Close()
			imageToSave = m.imageData.GetOriginal()
			defer imageToSave.Close()
		}

		if imageToSave.Empty() {
			m.showErrorDialog("Save Error", fmt.Errorf("no image to save"))
			return
		}

		// Save the image
		filepath := writer.URI().Path()
		if err := m.loader.SaveImage(imageToSave, filepath); err != nil {
			m.showErrorDialog("Failed to save image", err)
			return
		}

		m.logger.WithField("filepath", filepath).Info("Image saved successfully")
	}, fyne.CurrentApp().Driver().AllWindows()[0])

	fileDialog.SetFileName("processed_image.png")
	fileDialog.Show()
}

// savePresetDialog shows the save preset dialog
func (m *Menu) savePresetDialog() {
	sequence := m.pipeline.GetTransformationSequence()
	if sequence.Length() == 0 {
		m.showInfoDialog("No Transformations", "Please add some transformations before saving a preset.")
		return
	}

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter preset name...")

	descEntry := widget.NewMultiLineEntry()
	descEntry.SetPlaceHolder("Enter preset description (optional)...")
	descEntry.Resize(fyne.NewSize(300, 100))

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Name:", Widget: nameEntry},
			{Text: "Description:", Widget: descEntry},
		},
	}

	confirmDialog := dialog.NewCustomConfirm("Save Preset", "Save", "Cancel",
		form,
		func(save bool) {
			if !save {
				return
			}

			name := nameEntry.Text
			if name == "" {
				m.showErrorDialog("Invalid Name", fmt.Errorf("preset name cannot be empty"))
				return
			}

			if m.presetMgr.PresetExists(name) {
				m.showErrorDialog("Preset Exists", fmt.Errorf("a preset with this name already exists"))
				return
			}

			description := descEntry.Text
			preset, err := m.presetMgr.CreatePresetFromSequence(name, description, sequence)
			if err != nil {
				m.showErrorDialog("Failed to save preset", err)
				return
			}

			m.logger.WithField("preset", preset.Name).Info("Preset saved successfully")
		},
		fyne.CurrentApp().Driver().AllWindows()[0])

	confirmDialog.Show()
}

// loadPresetDialog shows the load preset dialog
func (m *Menu) loadPresetDialog() {
	presetNames := m.presetMgr.GetPresetNames()
	if len(presetNames) == 0 {
		m.showInfoDialog("No Presets", "No presets available. Create some presets first.")
		return
	}

	presetList := widget.NewList(
		func() int { return len(presetNames) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Preset")
		},
		func(id int, obj fyne.CanvasObject) {
			if id < len(presetNames) {
				obj.(*widget.Label).SetText(presetNames[id])
			}
		},
	)

	var selectedPreset string
	presetList.OnSelected = func(id int) {
		if id < len(presetNames) {
			selectedPreset = presetNames[id]
		}
	}

	confirmDialog := dialog.NewCustomConfirm("Load Preset", "Load", "Cancel",
		presetList,
		func(load bool) {
			if !load || selectedPreset == "" {
				return
			}

			sequence, err := m.presetMgr.ApplyPreset(selectedPreset)
			if err != nil {
				m.showErrorDialog("Failed to load preset", err)
				return
			}

			m.pipeline.LoadSequence(sequence)

			// Trigger callback
			if m.onPresetLoaded != nil {
				m.onPresetLoaded()
			}

			m.logger.WithField("preset", selectedPreset).Info("Preset loaded successfully")
		},
		fyne.CurrentApp().Driver().AllWindows()[0])

	confirmDialog.Resize(fyne.NewSize(400, 300))
	confirmDialog.Show()
}

// managePresetsDialog shows the preset management dialog
func (m *Menu) managePresetsDialog() {
	presets := m.presetMgr.ListPresets()

	presetList := widget.NewList(
		func() int { return len(presets) },
		func() fyne.CanvasObject {
			return widget.NewLabel("Preset")
		},
		func(id int, obj fyne.CanvasObject) {
			if id < len(presets) {
				preset := presets[id]
				obj.(*widget.Label).SetText(fmt.Sprintf("%s - %s", preset.Name, preset.GetSummary()))
			}
		},
	)

	var selectedIndex int = -1
	presetList.OnSelected = func(id int) {
		selectedIndex = id
	}

	deleteButton := widget.NewButton("Delete Selected", func() {
		if selectedIndex >= 0 && selectedIndex < len(presets) {
			presetName := presets[selectedIndex].Name
			confirm := dialog.NewConfirm("Delete Preset",
				fmt.Sprintf("Are you sure you want to delete preset '%s'?", presetName),
				func(confirmed bool) {
					if confirmed {
						if err := m.presetMgr.DeletePreset(presetName); err != nil {
							m.showErrorDialog("Failed to delete preset", err)
						} else {
							// Refresh the list
							presets = m.presetMgr.ListPresets()
							presetList.Refresh()
							selectedIndex = -1
						}
					}
				},
				fyne.CurrentApp().Driver().AllWindows()[0])
			confirm.Show()
		}
	})

	content := container.NewBorder(
		nil,
		deleteButton,
		nil,
		nil,
		presetList,
	)

	customDialog := dialog.NewCustom("Manage Presets", "Close", content, fyne.CurrentApp().Driver().AllWindows()[0])
	customDialog.Resize(fyne.NewSize(500, 400))
	customDialog.Show()
}

// showAboutDialog shows the about dialog
func (m *Menu) showAboutDialog() {
	content := widget.NewRichTextFromMarkdown(`
# Advanced Image Processing

**Version:** 1.0  
**Author:** Ervins Strauhmanis  
**License:** MIT

A powerful image processing application for historical illustrations, engravings, and document scans.

## Features
- Multiple binarization algorithms (Otsu, Niblack, Sauvola)
- Morphological operations (Erosion, Dilation)
- Noise reduction filters
- Real-time preview
- Preset management
- Extensible architecture

Built with Go, Fyne, and OpenCV.
`)

	aboutDialog := dialog.NewCustom("About", "Close", content, fyne.CurrentApp().Driver().AllWindows()[0])
	aboutDialog.Resize(fyne.NewSize(400, 500))
	aboutDialog.Show()
}

// showErrorDialog shows an error dialog
func (m *Menu) showErrorDialog(title string, err error) {
	dialog.ShowError(err, fyne.CurrentApp().Driver().AllWindows()[0])
	m.logger.WithError(err).Error(title)
}

// showInfoDialog shows an info dialog
func (m *Menu) showInfoDialog(title, message string) {
	dialog.ShowInformation(title, message, fyne.CurrentApp().Driver().AllWindows()[0])
}

// SetCallbacks sets callback functions
func (m *Menu) SetCallbacks(onImageLoaded, onPresetLoaded func()) {
	m.onImageLoaded = onImageLoaded
	m.onPresetLoaded = onPresetLoaded
}
