package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"gocv.io/x/gocv"
)

type DebugPipeline struct {
	enabled    bool
	timings    map[string]time.Time
	imageStats map[string]ImageStats
	operations []OperationLog
}

type ImageStats struct {
	Width     int
	Height    int
	Channels  int
	Type      gocv.MatType
	Size      int
	Timestamp time.Time
}

type OperationLog struct {
	Name      string
	Duration  time.Duration
	Before    ImageStats
	After     ImageStats
	Timestamp time.Time
}

func NewDebugPipeline() *DebugPipeline {
	return &DebugPipeline{
		enabled:    true, // Set to true to enable pipeline debugging
		timings:    make(map[string]time.Time),
		imageStats: make(map[string]ImageStats),
		operations: make([]OperationLog, 0),
	}
}

func (d *DebugPipeline) Enable() {
	d.enabled = true
	d.Log("Pipeline debugging enabled")
}

func (d *DebugPipeline) Disable() {
	d.enabled = false
}

func (d *DebugPipeline) Log(message string) {
	if !d.enabled {
		return
	}
	log.Println("[PIPELINE DEBUG]", message)
}

func (d *DebugPipeline) StartTimer(operation string) {
	if !d.enabled {
		return
	}
	d.timings[operation] = time.Now()
}

func (d *DebugPipeline) EndTimer(operation string) time.Duration {
	if !d.enabled {
		return 0
	}

	if startTime, exists := d.timings[operation]; exists {
		duration := time.Since(startTime)
		d.Log(fmt.Sprintf("Operation '%s' took %v", operation, duration))
		delete(d.timings, operation)
		return duration
	}

	return 0
}

func (d *DebugPipeline) LogImageStats(name string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() {
		return
	}

	size := mat.Size()
	stats := ImageStats{
		Width:     size[1],
		Height:    size[0],
		Channels:  mat.Channels(),
		Type:      mat.Type(),
		Size:      mat.Total() * mat.ElemSize(),
		Timestamp: time.Now(),
	}

	d.imageStats[name] = stats
	d.Log(fmt.Sprintf("Image '%s': %dx%d, %d channels, type %d, size %d bytes",
		name, stats.Width, stats.Height, stats.Channels, int(stats.Type), stats.Size))
}

func (d *DebugPipeline) LogTransformationApplied(transformationName string, input, output gocv.Mat, duration time.Duration) {
	if !d.enabled {
		return
	}

	var beforeStats, afterStats ImageStats

	if !input.Empty() {
		inputSize := input.Size()
		beforeStats = ImageStats{
			Width:    inputSize[1],
			Height:   inputSize[0],
			Channels: input.Channels(),
			Type:     input.Type(),
			Size:     input.Total() * input.ElemSize(),
		}
	}

	if !output.Empty() {
		outputSize := output.Size()
		afterStats = ImageStats{
			Width:    outputSize[1],
			Height:   outputSize[0],
			Channels: output.Channels(),
			Type:     output.Type(),
			Size:     output.Total() * output.ElemSize(),
		}
	}

	operation := OperationLog{
		Name:      transformationName,
		Duration:  duration,
		Before:    beforeStats,
		After:     afterStats,
		Timestamp: time.Now(),
	}

	d.operations = append(d.operations, operation)

	d.Log(fmt.Sprintf("Transformation '%s' applied in %v", transformationName, duration))
	d.Log(fmt.Sprintf("  Before: %dx%d, %d channels", beforeStats.Width, beforeStats.Height, beforeStats.Channels))
	d.Log(fmt.Sprintf("  After:  %dx%d, %d channels", afterStats.Width, afterStats.Height, afterStats.Channels))
}

func (d *DebugPipeline) LogMemoryUsage() {
	if !d.enabled {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	d.Log(fmt.Sprintf("Memory - Alloc: %.2f MB, TotalAlloc: %.2f MB, Sys: %.2f MB, NumGC: %d",
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		m.NumGC))
}

func (d *DebugPipeline) LogPipelineStats(originalSize, processedSize []int, numTransformations int) {
	if !d.enabled {
		return
	}

	d.Log(fmt.Sprintf("Pipeline Stats:"))
	d.Log(fmt.Sprintf("  Original size: %dx%d", originalSize[1], originalSize[0]))
	d.Log(fmt.Sprintf("  Processed size: %dx%d", processedSize[1], processedSize[0]))
	d.Log(fmt.Sprintf("  Number of transformations: %d", numTransformations))

	totalDuration := time.Duration(0)
	for _, op := range d.operations {
		totalDuration += op.Duration
	}
	d.Log(fmt.Sprintf("  Total processing time: %v", totalDuration))
}

func (d *DebugPipeline) GetOperationHistory() []OperationLog {
	return d.operations
}

func (d *DebugPipeline) ClearHistory() {
	if d.enabled {
		d.operations = make([]OperationLog, 0)
		d.imageStats = make(map[string]ImageStats)
		d.Log("Debug history cleared")
	}
}

func (d *DebugPipeline) LogMatrixProperties(name string, mat gocv.Mat) {
	if !d.enabled || mat.Empty() {
		return
	}

	size := mat.Size()
	d.Log(fmt.Sprintf("Matrix '%s' properties:", name))
	d.Log(fmt.Sprintf("  Dimensions: %dx%d", size[1], size[0]))
	d.Log(fmt.Sprintf("  Channels: %d", mat.Channels()))
	d.Log(fmt.Sprintf("  Type: %d", int(mat.Type())))
	d.Log(fmt.Sprintf("  Element size: %d bytes", mat.ElemSize()))
	d.Log(fmt.Sprintf("  Total elements: %d", mat.Total()))
	d.Log(fmt.Sprintf("  Total size: %d bytes", mat.Total()*mat.ElemSize()))
	d.Log(fmt.Sprintf("  Continuous: %t", mat.IsContinuous()))
}

func (d *DebugPipeline) IsEnabled() bool {
	return d.enabled
}
