package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"gocv.io/x/gocv"
)

type MatInfo struct {
	Name      string
	Pointer   unsafe.Pointer
	CreatedAt time.Time
	Stack     string
	Closed    bool
}

type DebugMemory struct {
	enabled bool
	mats    map[unsafe.Pointer]*MatInfo
	mutex   sync.RWMutex
}

func NewDebugMemory() *DebugMemory {
	return &DebugMemory{
		enabled: true,
		mats:    make(map[unsafe.Pointer]*MatInfo),
	}
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

func (d *DebugMemory) LogMatCreationWithPointer(name string, mat gocv.Mat) {
	if !d.enabled {
		return
	}

	ptr := mat.Ptr()
	if ptr == nil {
		d.Log(fmt.Sprintf("WARNING: Mat '%s' created with null pointer", name))
		return
	}

	// Get stack trace
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	d.mutex.Lock()
	d.mats[unsafe.Pointer(ptr)] = &MatInfo{
		Name:      name,
		Pointer:   unsafe.Pointer(ptr),
		CreatedAt: time.Now(),
		Stack:     stack,
		Closed:    false,
	}
	d.mutex.Unlock()

	log.Printf("[MEMORY DEBUG] Mat '%s' created with pointer %p", name, ptr)
}

func (d *DebugMemory) LogMatCleanup(name string) {
	if d.enabled {
		log.Printf("[MEMORY DEBUG] Mat '%s' cleaned up", name)
	}
}

func (d *DebugMemory) LogMatCleanupWithPointer(name string, mat gocv.Mat) {
	if !d.enabled {
		return
	}

	ptr := mat.Ptr()
	if ptr == nil {
		d.Log(fmt.Sprintf("WARNING: Attempting to cleanup Mat '%s' with null pointer", name))
		return
	}

	d.mutex.Lock()
	if info, exists := d.mats[unsafe.Pointer(ptr)]; exists {
		info.Closed = true
		log.Printf("[MEMORY DEBUG] Mat '%s' cleaned up (pointer %p, age: %v)", name, ptr, time.Since(info.CreatedAt))
	} else {
		log.Printf("[MEMORY DEBUG] WARNING: Mat '%s' cleanup attempted but not tracked (pointer %p)", name, ptr)
	}
	d.mutex.Unlock()
}

func (d *DebugMemory) ValidateMatPointer(name string, mat gocv.Mat) bool {
	if !d.enabled {
		return true
	}

	ptr := mat.Ptr()
	if ptr == nil {
		log.Printf("[MEMORY DEBUG] ERROR: Mat '%s' has null pointer!", name)
		return false
	}

	d.mutex.RLock()
	info, exists := d.mats[unsafe.Pointer(ptr)]
	d.mutex.RUnlock()

	if !exists {
		log.Printf("[MEMORY DEBUG] WARNING: Mat '%s' pointer %p not tracked", name, ptr)
		return false
	}

	if info.Closed {
		log.Printf("[MEMORY DEBUG] ERROR: Mat '%s' pointer %p was already closed! Created at %v",
			name, ptr, info.CreatedAt)
		log.Printf("[MEMORY DEBUG] Original creation stack:\n%s", info.Stack)
		return false
	}

	return true
}

func (d *DebugMemory) LogMemorySummary() {
	if !d.enabled {
		return
	}

	d.mutex.RLock()
	total := len(d.mats)
	active := 0
	closed := 0

	for _, info := range d.mats {
		if info.Closed {
			closed++
		} else {
			active++
		}
	}
	d.mutex.RUnlock()

	log.Printf("[MEMORY DEBUG] Summary: %d total Mats tracked, %d active, %d closed",
		total, active, closed)
}

func (d *DebugMemory) LogActiveMatDetails() {
	if !d.enabled {
		return
	}

	d.mutex.RLock()
	defer d.mutex.RUnlock()

	log.Printf("[MEMORY DEBUG] Active Mat details:")
	for ptr, info := range d.mats {
		if !info.Closed {
			age := time.Since(info.CreatedAt)
			log.Printf("[MEMORY DEBUG]   '%s' (ptr=%p, age=%v)", info.Name, ptr, age)
		}
	}
}

func (d *DebugMemory) IsEnabled() bool {
	return d.enabled
}

func (d *DebugMemory) Enable() {
	d.enabled = true
	d.Log("Memory debugging enabled with pointer tracking")
}

func (d *DebugMemory) Disable() {
	d.enabled = false
}
