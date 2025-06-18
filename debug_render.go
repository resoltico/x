package main

import (
	"fmt"
	"image"
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type DebugRender struct {
	enabled bool
}

func NewDebugRender() *DebugRender {
	return &DebugRender{
		enabled: true, // Set to true to enable terminal debug output for rendering operations
	}
}

func (d *DebugRender) Log(message string) {
	if !d.enabled {
		return
	}
	log.Println("[RENDER DEBUG]", message)
}

func (d *DebugRender) LogError(err error) {
	if !d.enabled || err == nil {
		return
	}
	log.Println("[RENDER ERROR]", err)
}

func (d *DebugRender) LogCanvasObjectDetails(name string, obj fyne.CanvasObject) {
	if !d.enabled || obj == nil {
		return
	}

	pos := obj.Position()
	size := obj.Size()
	minSize := obj.MinSize()
	visible := obj.Visible()

	log.Printf("[RENDER DEBUG] CanvasObject '%s': pos=(%.1f,%.1f), size=(%.1fx%.1f), minSize=(%.1fx%.1f), visible=%t",
		name, pos.X, pos.Y, size.Width, size.Height, minSize.Width, minSize.Height, visible)
}

func (d *DebugRender) LogImageDetails(name string, img *canvas.Image) {
	if !d.enabled || img == nil {
		return
	}

	pos := img.Position()
	size := img.Size()
	minSize := img.MinSize()
	visible := img.Visible()
	fillMode := img.FillMode
	scaleMode := img.ScaleMode
	translucency := img.Translucency

	log.Printf("[RENDER DEBUG] Image '%s': pos=(%.1f,%.1f), size=(%.1fx%.1f), minSize=(%.1fx%.1f), visible=%t",
		name, pos.X, pos.Y, size.Width, size.Height, minSize.Width, minSize.Height, visible)
	log.Printf("[RENDER DEBUG] Image '%s': fillMode=%d, scaleMode=%d, translucency=%.3f",
		name, int(fillMode), int(scaleMode), translucency)

	if img.Image != nil {
		bounds := img.Image.Bounds()
		log.Printf("[RENDER DEBUG] Image '%s': image bounds=(%d,%d,%d,%d), size=%dx%d",
			name, bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y,
			bounds.Dx(), bounds.Dy())
	}
}

func (d *DebugRender) LogScrollDetails(name string, scroll *container.Scroll) {
	if !d.enabled || scroll == nil {
		return
	}

	pos := scroll.Position()
	size := scroll.Size()
	minSize := scroll.MinSize()
	visible := scroll.Visible()
	offset := scroll.Offset

	log.Printf("[RENDER DEBUG] Scroll '%s': pos=(%.1f,%.1f), size=(%.1fx%.1f), minSize=(%.1fx%.1f), visible=%t",
		name, pos.X, pos.Y, size.Width, size.Height, minSize.Width, minSize.Height, visible)
	log.Printf("[RENDER DEBUG] Scroll '%s': offset=(%.1f,%.1f)",
		name, offset.X, offset.Y)

	if scroll.Content != nil {
		contentPos := scroll.Content.Position()
		contentSize := scroll.Content.Size()
		contentMinSize := scroll.Content.MinSize()
		log.Printf("[RENDER DEBUG] Scroll '%s' content: pos=(%.1f,%.1f), size=(%.1fx%.1f), minSize=(%.1fx%.1f)",
			name, contentPos.X, contentPos.Y, contentSize.Width, contentSize.Height,
			contentMinSize.Width, contentMinSize.Height)
	}
}

func (d *DebugRender) LogContainerDetails(name string, cont *fyne.Container) {
	if !d.enabled || cont == nil {
		return
	}

	pos := cont.Position()
	size := cont.Size()
	minSize := cont.MinSize()
	visible := cont.Visible()

	log.Printf("[RENDER DEBUG] Container '%s': pos=(%.1f,%.1f), size=(%.1fx%.1f), minSize=(%.1fx%.1f), visible=%t",
		name, pos.X, pos.Y, size.Width, size.Height, minSize.Width, minSize.Height, visible)
	log.Printf("[RENDER DEBUG] Container '%s': objects count=%d", name, len(cont.Objects))

	if cont.Layout != nil {
		layoutMinSize := cont.Layout.MinSize(cont.Objects)
		log.Printf("[RENDER DEBUG] Container '%s': layout minSize=(%.1fx%.1f)",
			name, layoutMinSize.Width, layoutMinSize.Height)
	}
}

func (d *DebugRender) LogLayoutEvent(name string, event string, beforeSize, afterSize fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Layout event '%s' on '%s': before=(%.1fx%.1f), after=(%.1fx%.1f)",
		event, name, beforeSize.Width, beforeSize.Height, afterSize.Width, afterSize.Height)
}

func (d *DebugRender) LogRenderingPipeline(name string, stage string) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Rendering pipeline '%s': %s", name, stage)
}

func (d *DebugRender) LogMemoryUsage() {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)

	log.Printf("[RENDER DEBUG] Memory - Alloc: %.2f MB, TotalAlloc: %.2f MB, Sys: %.2f MB, NumGC: %d",
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		m.NumGC)
}

func (d *DebugRender) LogImageChangeImpact(name string, beforeChannels, afterChannels int, beforeImg, afterImg *canvas.Image) {
	if !d.enabled {
		return
	}

	log.Printf("[RENDER DEBUG] Image change impact '%s': channels %d->%d", name, beforeChannels, afterChannels)

	if beforeImg != nil && afterImg != nil {
		beforeSize := beforeImg.Size()
		afterSize := afterImg.Size()
		beforeMinSize := beforeImg.MinSize()
		afterMinSize := afterImg.MinSize()

		log.Printf("[RENDER DEBUG] Image change '%s': size (%.1fx%.1f)->(%.1fx%.1f), minSize (%.1fx%.1f)->(%.1fx%.1f)",
			name, beforeSize.Width, beforeSize.Height, afterSize.Width, afterSize.Height,
			beforeMinSize.Width, beforeMinSize.Height, afterMinSize.Width, afterMinSize.Height)
	}
}

func (d *DebugRender) LogViewportExpansion(name string, expectedViewport, actualViewport fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Viewport expansion detected '%s': expected=(%.1fx%.1f), actual=(%.1fx%.1f)",
		name, expectedViewport.Width, expectedViewport.Height, actualViewport.Width, actualViewport.Height)
}

func (d *DebugRender) LogDeepInspection(name string, obj fyne.CanvasObject) {
	if !d.enabled || obj == nil {
		return
	}

	d.LogCanvasObjectDetails(name, obj)

	switch v := obj.(type) {
	case *canvas.Image:
		d.LogImageDetails(name, v)
	case *container.Scroll:
		d.LogScrollDetails(name, v)
	case *fyne.Container:
		d.LogContainerDetails(name, v)
	}
}

func (d *DebugRender) LogPixelAnalysis(stage string, origImg, prevImg *canvas.Image, origScroll, prevScroll *container.Scroll) {
	if !d.enabled {
		return
	}

	log.Printf("[RENDER DEBUG] === PIXEL ANALYSIS %s ===", stage)

	// Analyze original image viewport
	if origImg != nil && origScroll != nil {
		imgSize := origImg.Size()
		scrollSize := origScroll.Size()

		// Calculate actual viewport dimensions
		viewportWidth := scrollSize.Width
		viewportHeight := scrollSize.Height
		imageDisplayWidth := imgSize.Width
		imageDisplayHeight := imgSize.Height

		// Calculate fill ratio (how much of viewport is filled by image)
		fillRatioW := imageDisplayWidth / viewportWidth
		fillRatioH := imageDisplayHeight / viewportHeight

		log.Printf("[RENDER DEBUG] ORIGINAL viewport: scroll(%.1fx%.1f) img(%.1fx%.1f) fill(%.3fx%.3f)",
			viewportWidth, viewportHeight, imageDisplayWidth, imageDisplayHeight, fillRatioW, fillRatioH)

		// Check if image bounds exceed viewport
		if imageDisplayWidth > viewportWidth || imageDisplayHeight > viewportHeight {
			log.Printf("[RENDER DEBUG] ORIGINAL OVERFLOW detected: img exceeds viewport by (%.1f,%.1f)",
				imageDisplayWidth-viewportWidth, imageDisplayHeight-viewportHeight)
		}

		if origImg.Image != nil {
			imgBounds := origImg.Image.Bounds()
			log.Printf("[RENDER DEBUG] ORIGINAL source image: %dx%d, display scale: %.3fx%.3f",
				imgBounds.Dx(), imgBounds.Dy(),
				imageDisplayWidth/float32(imgBounds.Dx()),
				imageDisplayHeight/float32(imgBounds.Dy()))
		}
	}

	// Analyze preview image viewport
	if prevImg != nil && prevScroll != nil {
		imgSize := prevImg.Size()
		scrollSize := prevScroll.Size()

		// Calculate actual viewport dimensions
		viewportWidth := scrollSize.Width
		viewportHeight := scrollSize.Height
		imageDisplayWidth := imgSize.Width
		imageDisplayHeight := imgSize.Height

		// Calculate fill ratio
		fillRatioW := imageDisplayWidth / viewportWidth
		fillRatioH := imageDisplayHeight / viewportHeight

		log.Printf("[RENDER DEBUG] PREVIEW viewport: scroll(%.1fx%.1f) img(%.1fx%.1f) fill(%.3fx%.3f)",
			viewportWidth, viewportHeight, imageDisplayWidth, imageDisplayHeight, fillRatioW, fillRatioH)

		// Check if image bounds exceed viewport
		if imageDisplayWidth > viewportWidth || imageDisplayHeight > viewportHeight {
			log.Printf("[RENDER DEBUG] PREVIEW OVERFLOW detected: img exceeds viewport by (%.1f,%.1f)",
				imageDisplayWidth-viewportWidth, imageDisplayHeight-viewportHeight)
		}

		if prevImg.Image != nil {
			imgBounds := prevImg.Image.Bounds()
			imgChannels := "unknown"

			// Try to determine image format
			switch prevImg.Image.(type) {
			case *image.RGBA:
				imgChannels = "RGBA"
			case *image.Gray:
				imgChannels = "GRAY"
			case *image.NRGBA:
				imgChannels = "NRGBA"
			default:
				imgChannels = fmt.Sprintf("%T", prevImg.Image)
			}

			log.Printf("[RENDER DEBUG] PREVIEW source image: %dx%d (%s), display scale: %.3fx%.3f",
				imgBounds.Dx(), imgBounds.Dy(), imgChannels,
				imageDisplayWidth/float32(imgBounds.Dx()),
				imageDisplayHeight/float32(imgBounds.Dy()))
		}
	}

	// Compare viewports
	if origImg != nil && prevImg != nil && origScroll != nil && prevScroll != nil {
		origFillW := origImg.Size().Width / origScroll.Size().Width
		origFillH := origImg.Size().Height / origScroll.Size().Height
		prevFillW := prevImg.Size().Width / prevScroll.Size().Width
		prevFillH := prevImg.Size().Height / prevScroll.Size().Height

		fillDiffW := prevFillW - origFillW
		fillDiffH := prevFillH - origFillH

		if fillDiffW != 0 || fillDiffH != 0 {
			log.Printf("[RENDER DEBUG] VIEWPORT DIFFERENCE: preview fills %.3fx%.3f more than original",
				fillDiffW, fillDiffH)
		}
	}

	log.Printf("[RENDER DEBUG] === END PIXEL ANALYSIS %s ===", stage)
}

func (d *DebugRender) Enable() {
	d.enabled = true
	log.Println("[RENDER DEBUG] Render debugging enabled - output to terminal only")
}

func (d *DebugRender) Disable() {
	d.enabled = false
}
