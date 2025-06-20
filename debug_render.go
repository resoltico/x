package main

import (
	"fmt"
	"image"
	"log"
	"runtime"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"gocv.io/x/gocv"
)

type DebugRender struct {
	enabled bool
}

func NewDebugRender(config *DebugConfig) *DebugRender {
	enabled := false
	if config != nil {
		enabled = config.Render
	}
	return &DebugRender{
		enabled: enabled,
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
