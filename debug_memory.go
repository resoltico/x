package main

import (
	"log"
)

type DebugMemory struct {
	enabled bool
}

func NewDebugMemory() *DebugMemory {
	return &DebugMemory{enabled: true}
}

func (d *DebugMemory) Log(message string) {
	if d.enabled {
		log.Println("[MEMORY DEBUG]", message)
	}
}

func (d *DebugMemory) LogMatCreation(name string) {
	if d.enabled {
		log.Printf("[MEMORY DEBUG] Mat '%s' created", name)
	}
}

func (d *DebugMemory) LogMatCleanup(name string) {
	if d.enabled {
		log.Printf("[MEMORY DEBUG] Mat '%s' cleaned up", name)
	}
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
