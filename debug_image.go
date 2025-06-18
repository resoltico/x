package main

import (
	"log"
	"runtime"

	"gocv.io/x/gocv"
)

type DebugImage struct {
	enabled bool
}

func NewDebugImage() *DebugImage {
	return &DebugImage{
		enabled: true, // Set to false to disable debug output for production
	}
}

func (d *DebugImage) Log(message string) {
	if !d.enabled {
		return
	}
	log.Println("[IMAGE DEBUG]", message)
}

func (d *DebugImage) LogError(err error) {
	if !d.enabled || err == nil {
		return
	}
	log.Println("[IMAGE ERROR]", err)
}

func (d *DebugImage) LogMatInfo(name string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() {
		return
	}

	size := mat.Size()
	log.Printf("[IMAGE DEBUG] Mat '%s': %dx%d, channels=%d, type=%d, elemSize=%d",
		name, size[1], size[0], mat.Channels(), int(mat.Type()), mat.ElemSize())
}

func (d *DebugImage) LogColorConversion(from, to string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Color conversion: %s -> %s", from, to)
}

func (d *DebugImage) LogFilter(filterName string, params ...interface{}) {
	if !d.enabled {
		return
	}
	if len(params) > 0 {
		log.Printf("[IMAGE DEBUG] Applied filter: %s with params: %+v", filterName, params)
	} else {
		log.Printf("[IMAGE DEBUG] Applied filter: %s", filterName)
	}
}

func (d *DebugImage) LogThreshold(method string, threshold1, threshold2 float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Threshold method: %s, thresholds: %.2f, %.2f", method, threshold1, threshold2)
}

func (d *DebugImage) LogMorphology(operation string, kernelSize int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Morphological operation: %s, kernel size: %dx%d", operation, kernelSize, kernelSize)
}

func (d *DebugImage) LogHistogram(name string, bins int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Histogram calculated for '%s': %d bins", name, bins)
}

func (d *DebugImage) LogOptimalThresholds(s, t int, variance float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Optimal thresholds found: s=%d, t=%d, variance=%.6f", s, t, variance)
}

func (d *DebugImage) LogPixelValues(name string, x, y int, values ...interface{}) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Pixel values at (%d,%d) in '%s': %+v", x, y, name, values)
}

func (d *DebugImage) LogImageLoad(filename string, success bool) {
	if !d.enabled {
		return
	}
	if success {
		log.Printf("[IMAGE DEBUG] Successfully loaded image: %s", filename)
	} else {
		log.Printf("[IMAGE DEBUG] Failed to load image: %s", filename)
	}
}

func (d *DebugImage) LogImageSave(filename string, success bool) {
	if !d.enabled {
		return
	}
	if success {
		log.Printf("[IMAGE DEBUG] Successfully saved image: %s", filename)
	} else {
		log.Printf("[IMAGE DEBUG] Failed to save image: %s", filename)
	}
}

func (d *DebugImage) LogQualityMetrics(psnr, ssim float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Quality metrics - PSNR: %.2f dB, SSIM: %.4f", psnr, ssim)
}

func (d *DebugImage) LogMemoryUsage() {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)

	log.Printf("[IMAGE DEBUG] Memory - Alloc: %.2f MB, TotalAlloc: %.2f MB, Sys: %.2f MB, NumGC: %d",
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		m.NumGC)
}

func (d *DebugImage) LogImageProperties(name string, width, height, channels int, dataType string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Image '%s' properties: %dx%d, %d channels, type: %s",
		name, width, height, channels, dataType)
}

func (d *DebugImage) LogAlgorithmStep(algorithm, step string, details ...interface{}) {
	if !d.enabled {
		return
	}
	if len(details) > 0 {
		log.Printf("[IMAGE DEBUG] %s - %s: %+v", algorithm, step, details)
	} else {
		log.Printf("[IMAGE DEBUG] %s - %s", algorithm, step)
	}
}

func (d *DebugImage) LogProcessingTime(operation string, milliseconds float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Operation '%s' completed in %.2f ms", operation, milliseconds)
}

func (d *DebugImage) IsEnabled() bool {
	return d.enabled
}

func (d *DebugImage) Enable() {
	d.enabled = true
	log.Println("[IMAGE DEBUG] Image processing debugging enabled - output to terminal only")
}

func (d *DebugImage) Disable() {
	d.enabled = false
}
