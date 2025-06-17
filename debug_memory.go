package main

import (
	"log"

	"gocv.io/x/gocv"
)

type DebugMemory struct {
	enabled bool
}

func NewDebugMemory() *DebugMemory {
	return &DebugMemory{enabled: true}
}

// SafeMatEmpty - Check pointer before calling Empty() to prevent segfault
func (d *DebugMemory) SafeMatEmpty(name string, mat gocv.Mat) bool {
	// Get the raw pointer as uintptr
	matPtr := mat.Ptr()

	// Check if pointer is zero (NULL)
	if matPtr == uintptr(0) {
		if d.enabled {
			log.Printf("[MEMORY DEBUG] Mat '%s' has NULL pointer - returning true (empty)", name)
		}
		return true
	}

	// Safe to call Empty() - pointer is valid
	if d.enabled {
		log.Printf("[MEMORY DEBUG] Mat '%s' has valid pointer - calling Empty()", name)
	}
	return mat.Empty()
}

func (d *DebugMemory) Log(message string) {
	if d.enabled {
		log.Println("[MEMORY DEBUG]", message)
	}
}

func (d *DebugMemory) IsEnabled() bool {
	return d.enabled
}
