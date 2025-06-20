package main

import (
	"log"
	"runtime"
	"time"
)

type DebugMemory struct {
	enabled bool
}

func NewDebugMemory(config *DebugConfig) *DebugMemory {
	enabled := false
	if config != nil {
		enabled = config.Memory
	}
	return &DebugMemory{
		enabled: enabled,
	}
}

func (d *DebugMemory) Log(message string) {
	if !d.enabled {
		return
	}
	log.Println("[MEMORY DEBUG]", message)
}

func (d *DebugMemory) LogError(err error) {
	if !d.enabled || err == nil {
		return
	}
	log.Println("[MEMORY ERROR]", err)
}

func (d *DebugMemory) LogMemoryUsage() {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m)

	log.Printf("[MEMORY DEBUG] Memory - Alloc: %.2f MB, TotalAlloc: %.2f MB, Sys: %.2f MB, NumGC: %d",
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		m.NumGC)
}

func (d *DebugMemory) LogMemorySummary() {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	d.Log("=== Memory Summary ===")
	log.Printf("[MEMORY DEBUG] Final Memory Stats:")
	log.Printf("[MEMORY DEBUG]   Allocated: %.2f MB", float64(m.Alloc)/1024/1024)
	log.Printf("[MEMORY DEBUG]   Total Allocated: %.2f MB", float64(m.TotalAlloc)/1024/1024)
	log.Printf("[MEMORY DEBUG]   System Memory: %.2f MB", float64(m.Sys)/1024/1024)
	log.Printf("[MEMORY DEBUG]   Garbage Collections: %d", m.NumGC)
	log.Printf("[MEMORY DEBUG]   Last GC: %s", time.Unix(0, int64(m.LastGC)).Format(time.RFC3339))
	d.Log("=====================")
}

func (d *DebugMemory) LogAllocation(resource string, size int) {
	if !d.enabled {
		return
	}
	log.Printf("[MEMORY DEBUG] Allocated %s: %d bytes", resource, size)
}

func (d *DebugMemory) LogDeallocation(resource string) {
	if !d.enabled {
		return
	}
	log.Printf("[MEMORY DEBUG] Deallocated %s", resource)
}

func (d *DebugMemory) LogMatCreation(name string, width, height, channels int) {
	if !d.enabled {
		return
	}
	size := width * height * channels
	log.Printf("[MEMORY DEBUG] Mat created '%s': %dx%d, %d channels, ~%d bytes", 
		name, width, height, channels, size)
}

func (d *DebugMemory) LogMatDestruction(name string) {
	if !d.enabled {
		return
	}
	log.Printf("[MEMORY DEBUG] Mat destroyed '%s'", name)
}

func (d *DebugMemory) LogGC() {
	if !d.enabled {
		return
	}
	
	var before runtime.MemStats
	runtime.ReadMemStats(&before)
	
	runtime.GC()
	
	var after runtime.MemStats
	runtime.ReadMemStats(&after)
	
	freed := before.Alloc - after.Alloc
	log.Printf("[MEMORY DEBUG] GC triggered - freed %.2f MB", float64(freed)/1024/1024)
}

func (d *DebugMemory) IsEnabled() bool {
	return d.enabled
}

func (d *DebugMemory) Enable() {
	d.enabled = true
	d.Log("Memory debugging enabled")
}

func (d *DebugMemory) Disable() {
	d.enabled = false
}