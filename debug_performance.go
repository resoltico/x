package main

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

type DebugPerformance struct {
	enabled              bool
	operationStack       []OperationEntry
	stackMutex           sync.RWMutex
	hangDetectionEnabled bool
	hangThreshold        time.Duration
	activeOperations     map[string]*ActiveOperation
	activeOpMutex        sync.RWMutex
	goroutineTracker     *GoroutineTracker

	// Threshold search tracking
	lastLoggedVariance map[string]float64
	lastLogTime        map[string]time.Time
	searchMutex        sync.RWMutex
}

type OperationEntry struct {
	Name      string
	StartTime time.Time
	ThreadID  uint64
	MatCount  int
	MemAlloc  uint64
	Context   string
}

type ActiveOperation struct {
	Name      string
	StartTime time.Time
	Context   context.Context
	Cancel    context.CancelFunc
	ThreadID  uint64
	MatCount  int
}

type GoroutineTracker struct {
	initialCount int
	lastCheck    time.Time
	checkMutex   sync.Mutex
}

func NewDebugPerformance(config *DebugConfig) *DebugPerformance {
	enabled := false
	if config != nil {
		enabled = config.Pipeline || config.Image
	}

	return &DebugPerformance{
		enabled:              enabled,
		operationStack:       make([]OperationEntry, 0),
		hangDetectionEnabled: true,
		hangThreshold:        30 * time.Second,
		activeOperations:     make(map[string]*ActiveOperation),
		goroutineTracker:     &GoroutineTracker{initialCount: runtime.NumGoroutine()},
		lastLoggedVariance:   make(map[string]float64),
		lastLogTime:          make(map[string]time.Time),
	}
}

func (d *DebugPerformance) StartOperation(name, contextStr string) context.Context {
	if !d.enabled {
		return context.Background()
	}

	d.stackMutex.Lock()
	defer d.stackMutex.Unlock()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	entry := OperationEntry{
		Name:      name,
		StartTime: time.Now(),
		ThreadID:  uint64(getGoroutineID()),
		MatCount:  gocv.MatProfile.Count(),
		MemAlloc:  m.Alloc,
		Context:   contextStr,
	}

	d.operationStack = append(d.operationStack, entry)

	log.Printf("[PERF DEBUG] START: %s (ctx: %s) | Thread: %d | Mats: %d | Mem: %.2f MB | Stack depth: %d",
		name, contextStr, entry.ThreadID, entry.MatCount, float64(entry.MemAlloc)/1024/1024, len(d.operationStack))

	if d.hangDetectionEnabled {
		ctx, cancel := context.WithTimeout(context.Background(), d.hangThreshold)

		d.activeOpMutex.Lock()
		d.activeOperations[name] = &ActiveOperation{
			Name:      name,
			StartTime: entry.StartTime,
			Context:   ctx,
			Cancel:    cancel,
			ThreadID:  entry.ThreadID,
			MatCount:  entry.MatCount,
		}
		d.activeOpMutex.Unlock()

		go d.monitorOperation(name, ctx)
	}

	return context.Background()
}

func (d *DebugPerformance) EndOperation(name string) {
	if !d.enabled {
		return
	}

	d.stackMutex.Lock()
	defer d.stackMutex.Unlock()

	if len(d.operationStack) == 0 {
		log.Printf("[PERF DEBUG] WARNING: EndOperation called for '%s' but stack is empty", name)
		return
	}

	lastIdx := len(d.operationStack) - 1
	entry := d.operationStack[lastIdx]

	if entry.Name != name {
		log.Printf("[PERF DEBUG] WARNING: EndOperation mismatch! Expected '%s', got '%s'", name, entry.Name)
		log.Printf("[PERF DEBUG] Current stack:")
		for i, op := range d.operationStack {
			log.Printf("[PERF DEBUG]   [%d] %s (started %v ago)", i, op.Name, time.Since(op.StartTime))
		}
	}

	duration := time.Since(entry.StartTime)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	currentMatCount := gocv.MatProfile.Count()
	matDelta := currentMatCount - entry.MatCount
	memDelta := int64(m.Alloc) - int64(entry.MemAlloc)

	log.Printf("[PERF DEBUG] END: %s | Duration: %v | Mat delta: %+d | Mem delta: %+.2f MB | Final Mats: %d",
		name, duration, matDelta, float64(memDelta)/1024/1024, currentMatCount)

	if duration > 5*time.Second {
		log.Printf("[PERF DEBUG] SLOW OPERATION WARNING: %s took %v", name, duration)
		d.logDetailedSystemState(name)
	}

	if matDelta > 0 {
		log.Printf("[PERF DEBUG] POTENTIAL LEAK: %s created %d Mat(s) that weren't cleaned up", name, matDelta)
	}

	d.operationStack = d.operationStack[:lastIdx]

	if d.hangDetectionEnabled {
		d.activeOpMutex.Lock()
		if activeOp, exists := d.activeOperations[name]; exists {
			activeOp.Cancel()
			delete(d.activeOperations, name)
		}
		d.activeOpMutex.Unlock()
	}
}

func (d *DebugPerformance) LogStep(operation, step string, details ...interface{}) {
	if !d.enabled {
		return
	}

	timestamp := time.Now().Format("15:04:05.000")
	if len(details) > 0 {
		log.Printf("[PERF DEBUG] %s | %s - %s: %v", timestamp, operation, step, details)
	} else {
		log.Printf("[PERF DEBUG] %s | %s - %s", timestamp, operation, step)
	}
}

func (d *DebugPerformance) LogLoopProgress(operation string, current, total int, startTime time.Time) {
	if !d.enabled {
		return
	}

	// Only log at major milestones: 10%, 25%, 50%, 75%, 90%, 100%
	percentage := float64(current) / float64(total) * 100
	shouldLog := current == total || // Always log completion
		(current > 0 && (int(percentage) == 10 && int(float64(current-1)/float64(total)*100) < 10 ||
			int(percentage) == 25 && int(float64(current-1)/float64(total)*100) < 25 ||
			int(percentage) == 50 && int(float64(current-1)/float64(total)*100) < 50 ||
			int(percentage) == 75 && int(float64(current-1)/float64(total)*100) < 75 ||
			int(percentage) == 90 && int(float64(current-1)/float64(total)*100) < 90))

	if shouldLog {
		elapsed := time.Since(startTime)
		var eta time.Duration
		if current > 0 && current < total {
			avgTimePerItem := elapsed / time.Duration(current)
			remaining := total - current
			eta = avgTimePerItem * time.Duration(remaining)
		}

		log.Printf("[PERF DEBUG] %s PROGRESS: %d/%d (%.0f%%) | Elapsed: %v | ETA: %v",
			operation, current, total, percentage, elapsed, eta)
	}
}

func (d *DebugPerformance) LogMatrixOperation(operation string, input, output gocv.Mat) {
	if !d.enabled {
		return
	}

	var inputInfo, outputInfo string

	if !input.Empty() {
		inSize := input.Size()
		inputInfo = fmt.Sprintf("%dx%d/%dch", inSize[1], inSize[0], input.Channels())
	} else {
		inputInfo = "empty"
	}

	if !output.Empty() {
		outSize := output.Size()
		outputInfo = fmt.Sprintf("%dx%d/%dch", outSize[1], outSize[0], output.Channels())
	} else {
		outputInfo = "empty"
	}

	log.Printf("[PERF DEBUG] MATRIX OP: %s | Input: %s | Output: %s | Current Mats: %d",
		operation, inputInfo, outputInfo, gocv.MatProfile.Count())
}

func (d *DebugPerformance) LogHangDetection(operation string) {
	if !d.enabled {
		return
	}

	log.Printf("[PERF DEBUG] HANG DETECTED: %s has been running for >%v", operation, d.hangThreshold)
	log.Printf("[PERF DEBUG] HANG ANALYSIS:")

	d.logDetailedSystemState(operation)
	d.logCurrentStack()
	d.logGoroutineState()
}

func (d *DebugPerformance) monitorOperation(name string, ctx context.Context) {
	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			d.LogHangDetection(name)
		}
	}
}

func (d *DebugPerformance) logDetailedSystemState(operation string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	log.Printf("[PERF DEBUG] SYSTEM STATE for %s:", operation)
	log.Printf("[PERF DEBUG]   Go Memory:")
	log.Printf("[PERF DEBUG]     Alloc: %.2f MB", float64(m.Alloc)/1024/1024)
	log.Printf("[PERF DEBUG]     TotalAlloc: %.2f MB", float64(m.TotalAlloc)/1024/1024)
	log.Printf("[PERF DEBUG]     Sys: %.2f MB", float64(m.Sys)/1024/1024)
	log.Printf("[PERF DEBUG]     NumGC: %d", m.NumGC)
	log.Printf("[PERF DEBUG]     LastGC: %v ago", time.Since(time.Unix(0, int64(m.LastGC))))

	log.Printf("[PERF DEBUG]   OpenCV:")
	log.Printf("[PERF DEBUG]     MatProfile Count: %d", gocv.MatProfile.Count())

	log.Printf("[PERF DEBUG]   Runtime:")
	log.Printf("[PERF DEBUG]     Goroutines: %d", runtime.NumGoroutine())
	log.Printf("[PERF DEBUG]     GOMAXPROCS: %d", runtime.GOMAXPROCS(0))
	log.Printf("[PERF DEBUG]     CPUs: %d", runtime.NumCPU())
}

func (d *DebugPerformance) logCurrentStack() {
	d.stackMutex.RLock()
	defer d.stackMutex.RUnlock()

	log.Printf("[PERF DEBUG] CURRENT OPERATION STACK (%d operations):", len(d.operationStack))
	for i, op := range d.operationStack {
		duration := time.Since(op.StartTime)
		log.Printf("[PERF DEBUG]   [%d] %s | Running: %v | Thread: %d | Mats: %d | Context: %s",
			i, op.Name, duration, op.ThreadID, op.MatCount, op.Context)
	}
}

func (d *DebugPerformance) logGoroutineState() {
	d.goroutineTracker.checkMutex.Lock()
	defer d.goroutineTracker.checkMutex.Unlock()

	currentCount := runtime.NumGoroutine()
	delta := currentCount - d.goroutineTracker.initialCount

	log.Printf("[PERF DEBUG] GOROUTINE STATE:")
	log.Printf("[PERF DEBUG]   Initial count: %d", d.goroutineTracker.initialCount)
	log.Printf("[PERF DEBUG]   Current count: %d", currentCount)
	log.Printf("[PERF DEBUG]   Delta: %+d", delta)

	if delta > 10 {
		log.Printf("[PERF DEBUG]   WARNING: Significant goroutine increase detected!")
	}
}

func (d *DebugPerformance) LogAlgorithmPhase(algorithm, phase string, input gocv.Mat) {
	if !d.enabled {
		return
	}

	var inputInfo string
	if !input.Empty() {
		size := input.Size()
		inputInfo = fmt.Sprintf("%dx%d/%dch", size[1], size[0], input.Channels())
	} else {
		inputInfo = "empty"
	}

	log.Printf("[PERF DEBUG] ALGORITHM PHASE: %s | %s | Input: %s | Mats: %d",
		algorithm, phase, inputInfo, gocv.MatProfile.Count())
}

func (d *DebugPerformance) LogHistogramOperation(operation string, size []int, bins int) {
	if !d.enabled {
		return
	}

	pixels := 1
	for _, dim := range size {
		pixels *= dim
	}

	log.Printf("[PERF DEBUG] HISTOGRAM: %s | Image: %dx%d (%d pixels) | Bins: %d | Complexity: %d",
		operation, size[1], size[0], pixels, bins, pixels*bins)
}

func (d *DebugPerformance) LogThresholdSearch(algorithm string, searchSpace, currentPos int, maxVariance float64) {
	if !d.enabled {
		return
	}

	d.searchMutex.Lock()
	defer d.searchMutex.Unlock()

	lastVariance := d.lastLoggedVariance[algorithm]
	lastTime := d.lastLogTime[algorithm]

	// Only log if variance doubled AND 10 seconds passed
	if maxVariance > lastVariance*2.0 && time.Since(lastTime) > 10*time.Second {
		percentage := float64(currentPos) / float64(searchSpace) * 100
		log.Printf("[PERF DEBUG] %s SEARCH: %.0f%% | Best: %.2f",
			algorithm, percentage, maxVariance)
		d.lastLoggedVariance[algorithm] = maxVariance
		d.lastLogTime[algorithm] = time.Now()
	}
}

func (d *DebugPerformance) LogResourceContention(resource string, waitTime time.Duration) {
	if !d.enabled {
		return
	}

	if waitTime > 100*time.Millisecond {
		log.Printf("[PERF DEBUG] RESOURCE CONTENTION: %s | Wait time: %v", resource, waitTime)
	}
}

func (d *DebugPerformance) EnableHangDetection(threshold time.Duration) {
	d.hangDetectionEnabled = true
	d.hangThreshold = threshold
	log.Printf("[PERF DEBUG] Hang detection enabled with threshold: %v", threshold)
}

func (d *DebugPerformance) DisableHangDetection() {
	d.hangDetectionEnabled = false
	log.Printf("[PERF DEBUG] Hang detection disabled")
}

func (d *DebugPerformance) GetActiveOperations() []string {
	d.activeOpMutex.RLock()
	defer d.activeOpMutex.RUnlock()

	operations := make([]string, 0, len(d.activeOperations))
	for name, op := range d.activeOperations {
		duration := time.Since(op.StartTime)
		operations = append(operations, fmt.Sprintf("%s (running %v)", name, duration))
	}
	return operations
}

func (d *DebugPerformance) IsEnabled() bool {
	return d.enabled
}

func (d *DebugPerformance) Enable() {
	d.enabled = true
	log.Println("[PERF DEBUG] Performance debugging enabled")
}

func (d *DebugPerformance) Disable() {
	d.enabled = false
}

// Helper function to get goroutine ID (simple implementation)
func getGoroutineID() int {
	return runtime.NumGoroutine() // Simplified - in production you might want actual goroutine ID
}
