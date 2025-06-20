package main

import (
	"log"

	"fyne.io/fyne/v2"
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

func (d *DebugRender) LogCanvasRefresh(canvasName string) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Canvas refreshed: %s", canvasName)
}

func (d *DebugRender) LogImageRender(imageName string, width, height int) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Image rendered '%s': %dx%d", imageName, width, height)
}

func (d *DebugRender) LogWidgetResize(widgetName string, oldSize, newSize fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Widget '%s' resized: %.0fx%.0f -> %.0fx%.0f", 
		widgetName, oldSize.Width, oldSize.Height, newSize.Width, newSize.Height)
}

func (d *DebugRender) LogLayoutUpdate(containerName string, childCount int) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Layout updated '%s': %d children", containerName, childCount)
}

func (d *DebugRender) LogScrollUpdate(scrollName string, offset fyne.Position, contentSize fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Scroll updated '%s': offset(%.1f,%.1f), content %.0fx%.0f", 
		scrollName, offset.X, offset.Y, contentSize.Width, contentSize.Height)
}

func (d *DebugRender) LogImageScale(imageName string, originalSize, scaledSize fyne.Size, scale float32) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Image scaled '%s': %.0fx%.0f -> %.0fx%.0f (scale: %.2f)", 
		imageName, originalSize.Width, originalSize.Height, scaledSize.Width, scaledSize.Height, scale)
}

func (d *DebugRender) LogRenderTime(operation string, timeMs float64) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Render time '%s': %.2f ms", operation, timeMs)
}

func (d *DebugRender) LogUIRefresh(componentName string, reason string) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] UI refresh '%s': %s", componentName, reason)
}

func (d *DebugRender) LogFPSUpdate(fps float64) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] FPS: %.1f", fps)
}

func (d *DebugRender) LogTextureUpdate(textureName string, width, height int) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Texture updated '%s': %dx%d", textureName, width, height)
}

func (d *DebugRender) LogMemoryUsage(textureMemMB, vertexMemMB float64) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] GPU Memory: textures %.1f MB, vertices %.1f MB", textureMemMB, vertexMemMB)
}

func (d *DebugRender) LogWindowResize(width, height float32) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Window resized: %.0fx%.0f", width, height)
}

func (d *DebugRender) LogImageDisplayUpdate(imageName string, hasData bool, size fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Image display '%s': hasData=%t, size=%.0fx%.0f", 
		imageName, hasData, size.Width, size.Height)
}

func (d *DebugRender) LogCanvasObjectCount(containerName string, objectCount int) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Container '%s': %d objects", containerName, objectCount)
}

func (d *DebugRender) LogThemeChange(oldTheme, newTheme string) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Theme changed: %s -> %s", oldTheme, newTheme)
}

func (d *DebugRender) LogDrawCall(objectType string, position fyne.Position, size fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Draw call '%s': pos(%.1f,%.1f), size(%.1fx%.1f)", 
		objectType, position.X, position.Y, size.Width, size.Height)
}

func (d *DebugRender) LogClipping(objectName string, clipRect fyne.Size) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Clipping '%s': %.0fx%.0f", objectName, clipRect.Width, clipRect.Height)
}

func (d *DebugRender) LogAnimationFrame(animationName string, frame int, progress float64) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Animation '%s': frame %d, progress %.2f", animationName, frame, progress)
}

func (d *DebugRender) LogCursorChange(cursorType string) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Cursor changed: %s", cursorType)
}

func (d *DebugRender) LogImageConversionTime(format string, timeMs float64) {
	if !d.enabled {
		return
	}
	log.Printf("[RENDER DEBUG] Image conversion '%s': %.2f ms", format, timeMs)
}

func (d *DebugRender) IsEnabled() bool {
	return d.enabled
}

func (d *DebugRender) Enable() {
	d.enabled = true
	d.Log("Render debugging enabled")
}

func (d *DebugRender) Disable() {
	d.enabled = false
}