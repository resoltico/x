package main

import (
	"log"
	"runtime"
)

type DebugGUI struct {
	enabled bool
}

func NewDebugGUI() *DebugGUI {
	return &DebugGUI{
		enabled: true, // Set to true to enable terminal debug output for GUI operations
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

func (d *DebugGUI) LogZoomChange(oldZoom, newZoom float64) {
	if !d.enabled {
		return
	}
	log.Printf("[GUI DEBUG] Zoom changed from %.1f%% to %.1f%%", oldZoom*100, newZoom*100)
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
