package main

import (
	"fmt"
	"image"
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"gocv.io/x/gocv"
)

type DebugRender struct {
	enabled bool
}

func NewDebugRender() *DebugRender {
	return &DebugRender{
		enabled: true, // Set to false to disable render debugging
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

func (d *DebugRender) LogMatToImageConversion(matName string, mat gocv.Mat, success bool, errorMsg string) {
	if !d.enabled {
		return
	}

	if !mat.Empty() {
		size := mat.Size()
		channels := mat.Channels()
		matType := mat.Type()
		log.Printf("[RENDER DEBUG] Mat '%s' conversion: %dx%d, channels=%d, type=%d",
			matName, size[1], size[0], channels, int(matType))

		// Enhanced debugging for binary images
		if channels == 1 {
			// Sample pixel values to understand the data distribution
			data := mat.ToBytes()
			if len(data) > 100 {
				log.Printf("[RENDER DEBUG] Mat '%s' sample pixels: [%d, %d, %d, %d, %d]",
					matName, data[0], data[1], data[2], data[3], data[4])

				// Check for all-black condition that's plaguing us
				allBlack := true
				allWhite := true
				mixedCount := 0

				sampleSize := min(1000, len(data))
				for i := 0; i < sampleSize; i++ {
					if data[i] != 0 {
						allBlack = false
					}
					if data[i] != 255 {
						allWhite = false
					}
					if data[i] != 0 && data[i] != 255 {
						mixedCount++
					}
				}

				log.Printf("[RENDER DEBUG] Mat '%s' sample analysis: allBlack=%t, allWhite=%t, mixed=%d",
					matName, allBlack, allWhite, mixedCount)
			}
		}
	}

	if success {
		log.Printf("[RENDER DEBUG] Mat '%s' conversion to image.Image: SUCCESS", matName)
	} else {
		log.Printf("[RENDER DEBUG] Mat '%s' conversion to image.Image: FAILED - %s", matName, errorMsg)
	}
}

func (d *DebugRender) LogImageProperties(imgName string, img image.Image) {
	if !d.enabled || img == nil {
		return
	}

	bounds := img.Bounds()
	log.Printf("[RENDER DEBUG] Image '%s' properties: bounds=(%d,%d,%d,%d), size=%dx%d",
		imgName, bounds.Min.X, bounds.Min.Y, bounds.Max.X, bounds.Max.Y, bounds.Dx(), bounds.Dy())

	// Check image format and sample pixel values
	switch typedImg := img.(type) {
	case *image.RGBA:
		log.Printf("[RENDER DEBUG] Image '%s' format: RGBA", imgName)
		if bounds.Dx() > 10 && bounds.Dy() > 10 {
			// Sample a few pixels to verify content
			samples := []image.Point{{5, 5}, {bounds.Dx() / 2, bounds.Dy() / 2}, {bounds.Dx() - 5, bounds.Dy() - 5}}
			for i, pt := range samples {
				rgba := typedImg.RGBAAt(pt.X, pt.Y)
				log.Printf("[RENDER DEBUG] Image '%s' RGBA sample %d at (%d,%d): (%d,%d,%d,%d)",
					imgName, i, pt.X, pt.Y, rgba.R, rgba.G, rgba.B, rgba.A)
			}
		}
	case *image.Gray:
		log.Printf("[RENDER DEBUG] Image '%s' format: GRAY", imgName)
		// Sample some pixel values for grayscale
		if bounds.Dx() > 10 && bounds.Dy() > 10 {
			samples := []image.Point{{5, 5}, {bounds.Dx() / 2, bounds.Dy() / 2}, {bounds.Dx() - 5, bounds.Dy() - 5}}
			for i, pt := range samples {
				gray := typedImg.GrayAt(pt.X, pt.Y)
				log.Printf("[RENDER DEBUG] Image '%s' Gray sample %d at (%d,%d): %d",
					imgName, i, pt.X, pt.Y, gray.Y)
			}
		}
	case *image.NRGBA:
		log.Printf("[RENDER DEBUG] Image '%s' format: NRGBA", imgName)
	default:
		log.Printf("[RENDER DEBUG] Image '%s' format: %T", imgName, img)
	}
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

		// Enhanced black pixel analysis for binary images
		d.LogImageContentAnalysis(name, img.Image)
	}
}

func (d *DebugRender) LogImageContentAnalysis(name string, img image.Image) {
	if !d.enabled || img == nil {
		return
	}

	bounds := img.Bounds()
	if bounds.Dx() < 10 || bounds.Dy() < 10 {
		return
	}

	log.Printf("[RENDER DEBUG] CONTENT ANALYSIS for '%s':", name)

	// Comprehensive pixel sampling
	samplePoints := []image.Point{
		{bounds.Min.X + 5, bounds.Min.Y + 5},                             // Top-left
		{bounds.Min.X + bounds.Dx()/4, bounds.Min.Y + bounds.Dy()/4},     // Quarter
		{bounds.Min.X + bounds.Dx()/2, bounds.Min.Y + bounds.Dy()/2},     // Center
		{bounds.Min.X + 3*bounds.Dx()/4, bounds.Min.Y + 3*bounds.Dy()/4}, // Three-quarter
		{bounds.Max.X - 5, bounds.Max.Y - 5},                             // Bottom-right
	}

	// Count pixel values to detect all-black/all-white issues
	colorCounts := make(map[string]int)
	totalSamples := 0

	for i, pt := range samplePoints {
		var pixelDesc string
		switch typedImg := img.(type) {
		case *image.RGBA:
			rgba := typedImg.RGBAAt(pt.X, pt.Y)
			pixelDesc = fmt.Sprintf("RGBA(%d,%d,%d,%d)", rgba.R, rgba.G, rgba.B, rgba.A)
			// Classify as black/white/other
			if rgba.R == 0 && rgba.G == 0 && rgba.B == 0 {
				colorCounts["black"]++
			} else if rgba.R == 255 && rgba.G == 255 && rgba.B == 255 {
				colorCounts["white"]++
			} else {
				colorCounts["other"]++
			}
		case *image.Gray:
			gray := typedImg.GrayAt(pt.X, pt.Y)
			pixelDesc = fmt.Sprintf("Gray(%d)", gray.Y)
			if gray.Y == 0 {
				colorCounts["black"]++
			} else if gray.Y == 255 {
				colorCounts["white"]++
			} else {
				colorCounts["other"]++
			}
		case *image.NRGBA:
			nrgba := typedImg.NRGBAAt(pt.X, pt.Y)
			pixelDesc = fmt.Sprintf("NRGBA(%d,%d,%d,%d)", nrgba.R, nrgba.G, nrgba.B, nrgba.A)
			if nrgba.R == 0 && nrgba.G == 0 && nrgba.B == 0 {
				colorCounts["black"]++
			} else if nrgba.R == 255 && nrgba.G == 255 && nrgba.B == 255 {
				colorCounts["white"]++
			} else {
				colorCounts["other"]++
			}
		default:
			rgba := img.At(pt.X, pt.Y)
			pixelDesc = fmt.Sprintf("%T(%v)", rgba, rgba)
			colorCounts["unknown"]++
		}
		log.Printf("[RENDER DEBUG]   Sample %d at (%d,%d): %s", i, pt.X, pt.Y, pixelDesc)
		totalSamples++
	}

	// Report distribution analysis
	log.Printf("[RENDER DEBUG] Color distribution in %d samples:", totalSamples)
	for color, count := range colorCounts {
		percentage := float64(count) / float64(totalSamples) * 100
		log.Printf("[RENDER DEBUG]   %s: %d (%.1f%%)", color, count, percentage)
	}

	// Detect problematic patterns
	if colorCounts["black"] == totalSamples {
		log.Printf("[RENDER DEBUG] *** CRITICAL: Image '%s' appears to be ALL BLACK ***", name)
	} else if colorCounts["white"] == totalSamples {
		log.Printf("[RENDER DEBUG] *** CRITICAL: Image '%s' appears to be ALL WHITE ***", name)
	} else if colorCounts["black"]+colorCounts["white"] == totalSamples {
		log.Printf("[RENDER DEBUG] Image '%s' appears to be properly binary", name)
	} else {
		log.Printf("[RENDER DEBUG] Image '%s' has mixed/grayscale content", name)
	}
}

func (d *DebugRender) LogBinaryImageConversionIssue(matName string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() {
		return
	}

	if mat.Channels() != 1 {
		return // Not a binary/grayscale image
	}

	log.Printf("[RENDER DEBUG] BINARY IMAGE ANALYSIS for '%s':", matName)

	size := mat.Size()
	log.Printf("[RENDER DEBUG]   Size: %dx%d, Channels: %d, Type: %d",
		size[1], size[0], mat.Channels(), int(mat.Type()))

	// Get raw bytes and analyze with enhanced spatial sampling
	data := mat.ToBytes()
	if len(data) == 0 {
		log.Printf("[RENDER DEBUG]   ERROR: No data in Mat")
		return
	}

	// Sample from different spatial regions
	width, height := size[1], size[0]
	regions := []struct {
		name   string
		startX int
		startY int
		endX   int
		endY   int
	}{
		{"top-left", 0, 0, width / 4, height / 4},
		{"top-right", 3 * width / 4, 0, width, height / 4},
		{"center", width / 4, height / 4, 3 * width / 4, 3 * height / 4},
		{"bottom-left", 0, 3 * height / 4, width / 4, height},
		{"bottom-right", 3 * width / 4, 3 * height / 4, width, height},
	}

	log.Printf("[RENDER DEBUG]   Spatial region analysis:")
	for _, region := range regions {
		valueCount := make(map[uint8]int)
		sampleCount := 0

		for y := region.startY; y < region.endY && y < height; y++ {
			for x := region.startX; x < region.endX && x < width; x++ {
				offset := y*width + x
				if offset < len(data) {
					valueCount[data[offset]]++
					sampleCount++
				}
				if sampleCount >= 100 { // Limit samples per region
					break
				}
			}
			if sampleCount >= 100 {
				break
			}
		}

		log.Printf("[RENDER DEBUG]     Region %s (%d samples):", region.name, sampleCount)
		for val, count := range valueCount {
			percentage := float64(count) / float64(sampleCount) * 100
			log.Printf("[RENDER DEBUG]       Value %d: %d (%.1f%%)", val, count, percentage)
		}
	}

	// Try direct conversion
	img, err := mat.ToImage()
	if err != nil {
		log.Printf("[RENDER DEBUG]   Direct ToImage() FAILED: %v", err)
	} else {
		log.Printf("[RENDER DEBUG]   Direct ToImage() SUCCESS")
		d.LogImageProperties(matName+"_direct", img)
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

func (d *DebugRender) LogConversionMethodComparison(matName string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() || mat.Channels() != 1 {
		return
	}

	log.Printf("[RENDER DEBUG] CONVERSION METHOD COMPARISON for '%s':", matName)

	// Method 1: Direct ToImage()
	img1, err1 := mat.ToImage()
	if err1 != nil {
		log.Printf("[RENDER DEBUG]   Method 1 (Direct ToImage): FAILED - %v", err1)
	} else {
		log.Printf("[RENDER DEBUG]   Method 1 (Direct ToImage): SUCCESS")
		d.LogImageContentAnalysis(matName+"_method1", img1)
	}

	// Method 2: Convert to 3-channel then ToImage()
	mat3Channel := gocv.NewMat()
	defer mat3Channel.Close()
	gocv.CvtColor(mat, &mat3Channel, gocv.ColorGrayToBGR)

	img2, err2 := mat3Channel.ToImage()
	if err2 != nil {
		log.Printf("[RENDER DEBUG]   Method 2 (Gray->BGR->ToImage): FAILED - %v", err2)
	} else {
		log.Printf("[RENDER DEBUG]   Method 2 (Gray->BGR->ToImage): SUCCESS")
		d.LogImageContentAnalysis(matName+"_method2", img2)
	}

	// Method 3: Manual pixel copying
	size := mat.Size()
	width, height := size[1], size[0]
	bounds := image.Rect(0, 0, width, height)
	img3 := image.NewGray(bounds)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			val := mat.GetUCharAt(y, x)
			img3.SetGray(x, y, image.Gray{Y: val})
		}
	}

	log.Printf("[RENDER DEBUG]   Method 3 (Manual pixel copy): SUCCESS")
	d.LogImageContentAnalysis(matName+"_method3", img3)
}

func (d *DebugRender) IsEnabled() bool {
	return d.enabled
}

func (d *DebugRender) Enable() {
	d.enabled = true
	log.Println("[RENDER DEBUG] Render debugging enabled - output to terminal only")
}

func (d *DebugRender) Disable() {
	d.enabled = false
}
