package main

import (
	"log"
	"runtime"

	"fyne.io/fyne/v2"
)

type DebugGUI struct {
	enabled bool
}

func NewDebugGUI(config *DebugConfig) *DebugGUI {
	enabled := false
	if config != nil {
		enabled = config.GUI
	}
	return &DebugGUI{
		enabled: enabled,
	}
}

func (d *DebugGUI) Log(message string) {
	if !d.enabled {
		return
	}
	log.Println("[GUI DEBUG]", message)
}

func (d *DebugGUI) LogError(err error) {
	if !d.enabled || err == nil {
		return
	}
	log.Println("[GUI ERROR]", err)
}

func (d *DebugGUI) LogImageInfo(width, height, channels int) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Image Info - Size: %dx%d, Channels: %d", width, height, channels)
}

func (d *DebugGUI) LogTransformation(name string, params map[string]interface{}) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Applied Transformation: %s with params: %+v", name, params)
}

func (d *DebugGUI) LogUIEvent(event string, details ...interface{}) {
	if !d.enabled {
		return
	}
	if len(details) > 0 {
		log.Printf("[GUI DEBUG] UI Event: %s - %+v", event, details)
	} else {
		log.Printf("[GUI DEBUG] UI Event: %s", event)
	}
}

func (d *DebugGUI) LogButtonClick(buttonName string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Button clicked: %s", buttonName)
}

func (d *DebugGUI) LogSliderChange(sliderName string, oldValue, newValue float64) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Slider '%s' changed from %.3f to %.3f", sliderName, oldValue, newValue)
}

func (d *DebugGUI) LogListSelection(listName string, itemID int, itemName string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] List '%s' selection: ID=%d, Name='%s'", listName, itemID, itemName)
}

func (d *DebugGUI) LogListUnselect(listName string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] List '%s' unselected/cleared", listName)
}

func (d *DebugGUI) LogListInteraction(listName, action string, details interface{}) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] List '%s' %s: %v", listName, action, details)
}

func (d *DebugGUI) LogTransformationApplication(transformationName string, success bool) {
	if !d.enabled {
		return
	}
	if success {
		log.Printf("[GUI DEBUG] Transformation '%s' applied successfully", transformationName)
	} else {
		log.Printf("[GUI DEBUG] Transformation '%s' application failed", transformationName)
	}
}

func (d *DebugGUI) LogFileOperation(operation, filename string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] File operation: %s - %s", operation, filename)
}

func (d *DebugGUI) LogUIRefresh(componentName string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] UI component refreshed: %s", componentName)
}

func (d *DebugGUI) LogMemoryUsage() {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)

	log.Printf("[GUI DEBUG] Memory - Alloc: %.2f MB, TotalAlloc: %.2f MB, Sys: %.2f MB, NumGC: %d",
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		m.NumGC)
}

func (d *DebugGUI) LogWindowResize(width, height float32) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Window resized to %.0fx%.0f", width, height)
}

func (d *DebugGUI) LogParameterUpdate(transformationName, paramName string, oldValue, newValue interface{}) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Parameter updated in %s: %s changed from %v to %v",
		transformationName, paramName, oldValue, newValue)
}

func (d *DebugGUI) LogImageDisplay(imageName string, width, height int, hasImage bool) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Image display '%s': %dx%d, hasImage=%t", imageName, width, height, hasImage)
}

func (d *DebugGUI) LogCanvasRefresh(canvasName string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Canvas refreshed: %s", canvasName)
}

func (d *DebugGUI) LogImageConversion(imageName string, success bool, errorMsg string) {
	if !d.enabled {
		return
	}
	if success {
		log.Printf("[GUI DEBUG] Image conversion successful: %s", imageName)
	} else {
		log.Printf("[GUI DEBUG] Image conversion failed: %s - %s", imageName, errorMsg)
	}
}

func (d *DebugGUI) LogContainerRefresh(containerName string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Container refreshed: %s", containerName)
}

func (d *DebugGUI) LogLayoutIssue(componentName string, hasSize bool, width, height float32) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Layout Issue - %s: hasSize=%t, size=%.0fx%.0f", componentName, hasSize, width, height)
}

func (d *DebugGUI) LogUIStructure(description string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] UI Structure: %s", description)
}

func (d *DebugGUI) LogImageCanvasResize(canvasName string, width, height float32) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Canvas '%s' resized to %.0fx%.0f", canvasName, width, height)
}

func (d *DebugGUI) LogImageCanvasProperties(canvasName string, canvasWidth, canvasHeight float32, imageWidth, imageHeight int) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Canvas '%s': image=%dx%d", canvasName, imageWidth, imageHeight)
}

func (d *DebugGUI) LogCanvasSizeIssue(canvasName string, expectedSize, actualSize fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Canvas size issue '%s': expected=%.0fx%.0f, actual=%.0fx%.0f",
		canvasName, expectedSize.Width, expectedSize.Height, actualSize.Width, actualSize.Height)
}

func (d *DebugGUI) LogImageFormatChange(imageName string, fromChannels, toChannels int) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Image format change '%s': %d channels -> %d channels",
		imageName, fromChannels, toChannels)
}

func (d *DebugGUI) LogCanvasResize(canvasName string, width, height float32, reason string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Canvas '%s' resize to %.0fx%.0f - reason: %s", canvasName, width, height, reason)
}

func (d *DebugGUI) LogScrollContainerSize(containerName string, contentWidth, contentHeight, viewWidth, viewHeight float32) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Scroll container '%s': content=%.0fx%.0f, view=%.0fx%.0f",
		containerName, contentWidth, contentHeight, viewWidth, viewHeight)
}

func (d *DebugGUI) LogImageMinSize(imageName string, minWidth, minHeight float32) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Image '%s' MinSize: %.0fx%.0f", imageName, minWidth, minHeight)
}

func (d *DebugGUI) LogLayoutRefresh(componentName string, beforeSize, afterSize fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Layout refresh '%s': before=%.0fx%.0f, after=%.0fx%.0f",
		componentName, beforeSize.Width, beforeSize.Height, afterSize.Width, afterSize.Height)
}

func (d *DebugGUI) LogSaveOperation(filename, extension string, hasProcessedImage bool) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Save operation: file='%s', ext='%s', hasImage=%t", filename, extension, hasProcessedImage)
}

func (d *DebugGUI) LogSaveResult(filename string, success bool, errorMsg string) {
	if !d.enabled {
		return
	}
	if success {
		log.Printf("[GUI DEBUG] Save successful: %s", filename)
	} else {
		log.Printf("[GUI DEBUG] Save failed: %s - %s", filename, errorMsg)
	}
}

func (d *DebugGUI) LogFileExtensionCheck(filename, detectedExt string, isValid bool) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] File extension check: '%s' -> '%s', valid=%t", filename, detectedExt, isValid)
}

func (d *DebugGUI) LogUIRefreshTrigger(component string, reason string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] UI refresh triggered: %s (%s)", component, reason)
}

func (d *DebugGUI) LogLayoutPositions(componentName string, pos fyne.Position, size fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Layout '%s': pos=(%.1f,%.1f), size=(%.1fx%.1f)",
		componentName, pos.X, pos.Y, size.Width, size.Height)
}

func (d *DebugGUI) LogTextSizeChange(componentName, oldText, newText string, oldSize, newSize fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Text size change '%s': '%s'->'%s', size (%.1fx%.1f)->(%.1fx%.1f)",
		componentName, oldText, newText, oldSize.Width, oldSize.Height, newSize.Width, newSize.Height)
}

func (d *DebugGUI) LogProgressBarChange(componentName string, oldValue, newValue float64) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Progress bar '%s': %.3f -> %.3f", componentName, oldValue, newValue)
}

func (d *DebugGUI) LogQualityMetricsUpdate(psnr, ssim float64, hasTransformations bool) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Quality metrics update: PSNR=%.2f, SSIM=%.4f, hasTransformations=%t",
		psnr, ssim, hasTransformations)
}

func (d *DebugGUI) LogPanelSizes(leftWidth, centerWidth, rightWidth, totalHeight float32) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Panel sizes: left=%.0f, center=%.0f, right=%.0f, height=%.0f",
		leftWidth, centerWidth, rightWidth, totalHeight)
}

func (d *DebugGUI) LogThreadSafetyViolation(operation string, details string) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Thread safety warning: %s - %s", operation, details)
}

func (d *DebugGUI) IsEnabled() bool {
	return d.enabled
}

func (d *DebugGUI) Enable() {
	d.enabled = true
	log.Println("[GUI DEBUG] GUI debugging enabled - output to terminal only")
}

func (d *DebugGUI) Disable() {
	d.enabled = false
}