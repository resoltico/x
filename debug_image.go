package main

import (
	"log"

	"gocv.io/x/gocv"
)

type DebugImage struct {
	enabled bool
}

func NewDebugImage(config *DebugConfig) *DebugImage {
	enabled := false
	if config != nil {
		enabled = config.Image
	}
	return &DebugImage{
		enabled: enabled,
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
	log.Printf("[IMAGE DEBUG] Mat '%s': %dx%d, %d channels, type %d, size %d bytes",
		name, size[1], size[0], mat.Channels(), int(mat.Type()), mat.Total()*mat.ElemSize())
}

func (d *DebugImage) LogImageLoad(filename string, width, height, channels int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Loaded image '%s': %dx%d, %d channels", filename, width, height, channels)
}

func (d *DebugImage) LogImageSave(filename string, width, height int, success bool) {
	if !d.enabled {
		return
	}
	if success {
		log.Printf("[IMAGE DEBUG] Saved image '%s': %dx%d", filename, width, height)
	} else {
		log.Printf("[IMAGE DEBUG] Failed to save image '%s'", filename)
	}
}

func (d *DebugImage) LogColorConversion(from, to string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Color conversion: %s -> %s", from, to)
}

func (d *DebugImage) LogFilter(filterName, params, details string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Applied filter '%s': %s (%s)", filterName, params, details)
}

func (d *DebugImage) LogResize(name string, oldWidth, oldHeight, newWidth, newHeight int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Resized '%s': %dx%d -> %dx%d", name, oldWidth, oldHeight, newWidth, newHeight)
}

func (d *DebugImage) LogThreshold(name string, threshold float64, method string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Threshold '%s': %.2f using %s", name, threshold, method)
}

func (d *DebugImage) LogMorphology(operation string, kernelSize int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Morphology operation: %s with kernel size %d", operation, kernelSize)
}

func (d *DebugImage) LogAlgorithmStep(algorithm, step string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] %s: %s", algorithm, step)
}

func (d *DebugImage) LogPixelStats(name string, min, max, mean, stddev float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Pixel stats '%s': min=%.2f, max=%.2f, mean=%.2f, stddev=%.2f",
		name, min, max, mean, stddev)
}

func (d *DebugImage) LogHistogram(name string, bins []int) {
	if !d.enabled || len(bins) == 0 {
		return
	}

	// Log first few and last few histogram values
	if len(bins) <= 10 {
		log.Printf("[IMAGE DEBUG] Histogram '%s': %v", name, bins)
	} else {
		log.Printf("[IMAGE DEBUG] Histogram '%s': [%d,%d,%d...%d,%d,%d] (total %d bins)",
			name, bins[0], bins[1], bins[2], bins[len(bins)-3], bins[len(bins)-2], bins[len(bins)-1], len(bins))
	}
}

func (d *DebugImage) LogImageDimensions(name string, width, height int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Image dimensions '%s': %dx%d", name, width, height)
}

func (d *DebugImage) LogImageType(name string, matType gocv.MatType) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Image type '%s': %d", name, int(matType))
}

func (d *DebugImage) LogConvolution(kernelName string, kernelSize int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Applied convolution: %s (size %dx%d)", kernelName, kernelSize, kernelSize)
}

func (d *DebugImage) LogROI(name string, x, y, width, height int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] ROI '%s': (%d,%d) %dx%d", name, x, y, width, height)
}

func (d *DebugImage) LogTransformation(name string, inputSize, outputSize []int) {
	if !d.enabled || len(inputSize) < 2 || len(outputSize) < 2 {
		return
	}
	log.Printf("[IMAGE DEBUG] Transformation '%s': %dx%d -> %dx%d",
		name, inputSize[1], inputSize[0], outputSize[1], outputSize[0])
}

func (d *DebugImage) LogQualityMetric(metricName string, value float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Quality metric %s: %.4f", metricName, value)
}

func (d *DebugImage) LogChannelSplit(originalChannels int, splitChannels int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Channel split: %d channels -> %d separate channels", originalChannels, splitChannels)
}

func (d *DebugImage) LogChannelMerge(inputChannels int, outputChannels int) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Channel merge: %d channels -> %d channels", inputChannels, outputChannels)
}

func (d *DebugImage) LogWarning(warning string) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE WARNING] %s", warning)
}

func (d *DebugImage) LogPerformance(operation string, durationMs float64) {
	if !d.enabled {
		return
	}
	log.Printf("[IMAGE DEBUG] Performance '%s': %.2f ms", operation, durationMs)
}

func (d *DebugImage) LogImageProperties(name string, continuous bool, empty bool, size []int) {
	if !d.enabled {
		return
	}
	if len(size) >= 2 {
		log.Printf("[IMAGE DEBUG] Image properties '%s': %dx%d, continuous=%t, empty=%t",
			name, size[1], size[0], continuous, empty)
	}
}

func (d *DebugImage) IsEnabled() bool {
	return d.enabled
}

func (d *DebugImage) Enable() {
	d.enabled = true
	d.Log("Image debugging enabled")
}

func (d *DebugImage) Disable() {
	d.enabled = false
}
