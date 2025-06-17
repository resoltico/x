package main

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type DebugFyne struct {
	enabled bool
}

func NewDebugFyne() *DebugFyne {
	return &DebugFyne{
		enabled: true, // Set to true to enable Fyne-specific debugging
	}
}

func (d *DebugFyne) Log(message string) {
	if !d.enabled {
		return
	}
	log.Println("[FYNE DEBUG]", message)
}

func (d *DebugFyne) LogCanvasImageProperties(name string, img *canvas.Image) {
	if !d.enabled || img == nil {
		return
	}

	size := img.Size()
	position := img.Position()
	visible := img.Visible()
	hasImage := img.Image != nil

	log.Printf("[FYNE DEBUG] Canvas Image '%s':", name)
	log.Printf("[FYNE DEBUG]   Size: %.0fx%.0f", size.Width, size.Height)
	log.Printf("[FYNE DEBUG]   Position: (%.0f,%.0f)", position.X, position.Y)
	log.Printf("[FYNE DEBUG]   Visible: %t", visible)
	log.Printf("[FYNE DEBUG]   Has Image: %t", hasImage)
	log.Printf("[FYNE DEBUG]   Fill Mode: %d", img.FillMode)
	log.Printf("[FYNE DEBUG]   Scale Mode: %d", img.ScaleMode)

	if hasImage {
		bounds := img.Image.Bounds()
		log.Printf("[FYNE DEBUG]   Image Bounds: %dx%d", bounds.Dx(), bounds.Dy())
	}
}

func (d *DebugFyne) LogContainerProperties(name string, cont *fyne.Container) {
	if !d.enabled || cont == nil {
		return
	}

	size := cont.Size()
	position := cont.Position()
	visible := cont.Visible()
	objectCount := len(cont.Objects)

	log.Printf("[FYNE DEBUG] Container '%s':", name)
	log.Printf("[FYNE DEBUG]   Size: %.0fx%.0f", size.Width, size.Height)
	log.Printf("[FYNE DEBUG]   Position: (%.0f,%.0f)", position.X, position.Y)
	log.Printf("[FYNE DEBUG]   Visible: %t", visible)
	log.Printf("[FYNE DEBUG]   Object Count: %d", objectCount)

	if cont.Layout != nil {
		log.Printf("[FYNE DEBUG]   Has Layout: true")
	} else {
		log.Printf("[FYNE DEBUG]   Has Layout: false")
	}
}

func (d *DebugFyne) LogWindowProperties(name string, window fyne.Window) {
	if !d.enabled || window == nil {
		return
	}

	size := window.Canvas().Size()
	content := window.Content()

	log.Printf("[FYNE DEBUG] Window '%s':", name)
	log.Printf("[FYNE DEBUG]   Canvas Size: %.0fx%.0f", size.Width, size.Height)
	log.Printf("[FYNE DEBUG]   Has Content: %t", content != nil)

	if content != nil {
		contentSize := content.Size()
		log.Printf("[FYNE DEBUG]   Content Size: %.0fx%.0f", contentSize.Width, contentSize.Height)
	}
}

func (d *DebugFyne) LogObjectHierarchy(name string, obj fyne.CanvasObject, depth int) {
	if !d.enabled || obj == nil {
		return
	}

	indent := ""
	for i := 0; i < depth; i++ {
		indent += "  "
	}

	size := obj.Size()
	position := obj.Position()
	visible := obj.Visible()

	log.Printf("[FYNE DEBUG] %s%s Object:", indent, name)
	log.Printf("[FYNE DEBUG] %s  Type: %T", indent, obj)
	log.Printf("[FYNE DEBUG] %s  Size: %.0fx%.0f", indent, size.Width, size.Height)
	log.Printf("[FYNE DEBUG] %s  Position: (%.0f,%.0f)", indent, position.X, position.Y)
	log.Printf("[FYNE DEBUG] %s  Visible: %t", indent, visible)

	// If it's a container, log its children
	if cont, ok := obj.(*fyne.Container); ok {
		log.Printf("[FYNE DEBUG] %s  Children: %d", indent, len(cont.Objects))
		for i, child := range cont.Objects {
			d.LogObjectHierarchy(fmt.Sprintf("Child[%d]", i), child, depth+1)
		}
	}
}

func (d *DebugFyne) LogImageDisplayIssue(imageName string, img *canvas.Image) {
	if !d.enabled {
		return
	}

	log.Printf("[FYNE DEBUG] Image Display Issue Analysis for '%s':", imageName)

	if img == nil {
		log.Printf("[FYNE DEBUG]   ERROR: Image canvas is nil")
		return
	}

	if img.Image == nil {
		log.Printf("[FYNE DEBUG]   ERROR: No image data set")
		return
	}

	if !img.Visible() {
		log.Printf("[FYNE DEBUG]   ERROR: Image is not visible")
	}

	size := img.Size()
	if size.Width == 0 || size.Height == 0 {
		log.Printf("[FYNE DEBUG]   ERROR: Image size is zero: %.0fx%.0f", size.Width, size.Height)
	}

	bounds := img.Image.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		log.Printf("[FYNE DEBUG]   ERROR: Image data has zero bounds: %dx%d", bounds.Dx(), bounds.Dy())
	}

	log.Printf("[FYNE DEBUG]   Canvas size: %.0fx%.0f", size.Width, size.Height)
	log.Printf("[FYNE DEBUG]   Image bounds: %dx%d", bounds.Dx(), bounds.Dy())
	log.Printf("[FYNE DEBUG]   Fill mode: %d", img.FillMode)
	log.Printf("[FYNE DEBUG]   Scale mode: %d", img.ScaleMode)
}

func (d *DebugFyne) IsEnabled() bool {
	return d.enabled
}

func (d *DebugFyne) Enable() {
	d.enabled = true
	d.Log("Fyne debugging enabled")
}

func (d *DebugFyne) Disable() {
	d.enabled = false
}
