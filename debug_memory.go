package main

import (
	"log"
	"runtime"
	"sync"
	"time"
)

type MatInfo struct {
	ID        uint64
	Name      string
	CreatedAt time.Time
	Stack     string
	Closed    bool
}

type DebugMemory struct {
	enabled bool
	mats    map[uint64]*MatInfo
	nextID  uint64
	mutex   sync.RWMutex
}

func NewDebugMemory(config *DebugConfig) *DebugMemory {
	enabled := false
	if config != nil {
		enabled = config.Memory
	}
	return &DebugMemory{
		enabled: enabled,
		mats:    make(map[uint64]*MatInfo),
		nextID:  1,
	}
}

func (d *DebugMemory) Log(message string) {
	if d.enabled {
		log.Println("[MEMORY DEBUG]", message)
	}
}

func (d *DebugMemory) LogMatCreation(name string) uint64 {
	if !d.enabled {
		return 0
	}

	// Get stack trace
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	stack := string(buf[:n])

	d.mutex.Lock()
	id := d.nextID
	d.nextID++
	d.mats[id] = &MatInfo{
		ID:        id,
		Name:      name,
		CreatedAt: time.Now(),
		Stack:     stack,
		Closed:    false,
	}
	d.mutex.Unlock()

	log.Printf("[MEMORY DEBUG] Mat '%s' created with ID %d", name, id)
	return id
}

func (d *DebugMemory) LogMatCleanup(name string, id uint64) {
	if !d.enabled {
		return
	}

	d.mutex.Lock()
	if info, exists := d.mats[id]; exists {
		info.Closed = true
		log.Printf("[MEMORY DEBUG] Mat '%s' cleaned up (ID %d, age: %v)", name, id, time.Since(info.CreatedAt))
	} else {
		log.Printf("[MEMORY DEBUG] WARNING: Mat '%s' cleanup attempted but not tracked (ID %d)", name, id)
	}
	d.mutex.Unlock()
}

func (d *DebugMemory) ValidateMatID(name string, id uint64) bool {
	if !d.enabled {
		return true
	}

	d.mutex.RLock()
	info, exists := d.mats[id]
	d.mutex.RUnlock()

	if !exists {
		log.Printf("[MEMORY DEBUG] WARNING: Mat '%s' ID %d not tracked", name, id)
		return false
	}

	if info.Closed {
		log.Printf("[MEMORY DEBUG] ERROR: Mat '%s' ID %d was already closed! Created at %v",
			name, id, info.CreatedAt)
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
	for id, info := range d.mats {
		if !info.Closed {
			age := time.Since(info.CreatedAt)
			log.Printf("[MEMORY DEBUG]   '%s' (ID=%d, age=%v)", info.Name, id, age)
		}
	}
}

func (d *DebugMemory) LogMemoryLeak(name string, id uint64) {
	if !d.enabled {
		return
	}

	d.mutex.RLock()
	info, exists := d.mats[id]
	d.mutex.RUnlock()

	if exists && !info.Closed {
		age := time.Since(info.CreatedAt)
		log.Printf("[MEMORY DEBUG] POTENTIAL LEAK: Mat '%s' (ID=%d) has been active for %v", name, id, age)
		log.Printf("[MEMORY DEBUG] Creation stack:\n%s", info.Stack)
	}
}

func (d *DebugMemory) LogDeferOverwritePattern(name string, originalID uint64, newMatInfo string) {
	if !d.enabled {
		return
	}
	log.Printf("[MEMORY DEBUG] DEFER OVERWRITE DETECTED: Mat '%s' (ID=%d) overwritten with %s - original will leak!",
		name, originalID, newMatInfo)
}

func (d *DebugMemory) IsEnabled() bool {
	return d.enabled
}

func (d *DebugMemory) Enable() {
	d.enabled = true
	d.Log("Memory debugging enabled with ID-based tracking")
}

func (d *DebugMemory) Disable() {
	d.enabled = false
}
