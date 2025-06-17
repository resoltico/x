// internal/core/pipeline_debug.go
// Pipeline-specific debugging and performance monitoring
package core

import (
	"fmt"
	"log/slog"
	"time"
)

// PipelineDebugger handles all pipeline-related debugging
type PipelineDebugger struct {
	logger  *slog.Logger
	enabled bool

	// Operation tracking
	operations []PipelineOperation
	events     []ProcessingEvent

	// Performance metrics
	processingTimes   []time.Duration
	conversionTimes   []time.Duration
	layerProcessTimes []time.Duration
	metricsCalcTimes  []time.Duration

	// State tracking
	currentMode         string
	currentLayerCount   int
	lastProcessingStart time.Time
}

// PipelineOperation tracks individual pipeline operations
type PipelineOperation struct {
	Timestamp time.Time
	Operation string // "add_layer", "process_preview", "process_full", "calculate_metrics"
	Success   bool
	Duration  time.Duration
	Details   map[string]interface{}
	Error     string
}

// ProcessingEvent tracks processing flow events
type ProcessingEvent struct {
	Timestamp time.Time
	Event     string // "trigger", "start", "layer_processing", "conversion", "complete", "error"
	Mode      string // "layer", "sequential", "none"
	Details   map[string]interface{}
}

func NewPipelineDebugger(logger *slog.Logger) *PipelineDebugger {
	return &PipelineDebugger{
		logger:            logger,
		enabled:           true,
		operations:        make([]PipelineOperation, 0),
		events:            make([]ProcessingEvent, 0),
		processingTimes:   make([]time.Duration, 0),
		conversionTimes:   make([]time.Duration, 0),
		layerProcessTimes: make([]time.Duration, 0),
		metricsCalcTimes:  make([]time.Duration, 0),
		currentMode:       "none",
	}
}

// Operation logging
func (pd *PipelineDebugger) LogOperation(operation string, success bool, duration time.Duration, details map[string]interface{}, err error) {
	if !pd.enabled {
		return
	}

	errorStr := ""
	if err != nil {
		errorStr = err.Error()
	}

	op := PipelineOperation{
		Timestamp: time.Now(),
		Operation: operation,
		Success:   success,
		Duration:  duration,
		Details:   details,
		Error:     errorStr,
	}

	pd.operations = append(pd.operations, op)

	// Store performance metrics by category
	switch operation {
	case "process_preview", "process_full":
		pd.processingTimes = append(pd.processingTimes, duration)
	case "mat_conversion":
		pd.conversionTimes = append(pd.conversionTimes, duration)
	case "layer_processing":
		pd.layerProcessTimes = append(pd.layerProcessTimes, duration)
	case "calculate_metrics":
		pd.metricsCalcTimes = append(pd.metricsCalcTimes, duration)
	}

	level := slog.LevelInfo
	if !success {
		level = slog.LevelError
	}

	pd.logger.Log(nil, level, "PIPELINE Debug",
		"operation", operation,
		"success", success,
		"duration_ms", duration.Milliseconds(),
		"details", details,
		"error", errorStr)
}

// Event logging
func (pd *PipelineDebugger) LogEvent(event, mode string, details map[string]interface{}) {
	if !pd.enabled {
		return
	}

	evt := ProcessingEvent{
		Timestamp: time.Now(),
		Event:     event,
		Mode:      mode,
		Details:   details,
	}

	pd.events = append(pd.events, evt)
	pd.currentMode = mode

	pd.logger.Info("PIPELINE Event",
		"event", event,
		"mode", mode,
		"details", details)
}

// Specific pipeline operation logging
func (pd *PipelineDebugger) LogLayerAddition(layerID, algorithm string, params map[string]interface{}, success bool, duration time.Duration, err error) {
	details := map[string]interface{}{
		"layer_id":  layerID,
		"algorithm": algorithm,
		"params":    params,
	}
	pd.LogOperation("add_layer", success, duration, details, err)
}

func (pd *PipelineDebugger) LogModeChange(oldMode, newMode string, layerCount int) {
	pd.currentLayerCount = layerCount
	details := map[string]interface{}{
		"old_mode":    oldMode,
		"new_mode":    newMode,
		"layer_count": layerCount,
	}
	pd.LogEvent("mode_change", newMode, details)
}

func (pd *PipelineDebugger) LogProcessingStart(mode string, layerCount int, inputSize string) {
	pd.lastProcessingStart = time.Now()
	pd.currentLayerCount = layerCount
	details := map[string]interface{}{
		"layer_count": layerCount,
		"input_size":  inputSize,
	}
	pd.LogEvent("processing_start", mode, details)
}

func (pd *PipelineDebugger) LogProcessingComplete(mode string, success bool, outputSize string, metrics map[string]float64) {
	duration := time.Since(pd.lastProcessingStart)
	details := map[string]interface{}{
		"output_size": outputSize,
		"metrics":     metrics,
	}

	var err error
	if !success {
		err = fmt.Errorf("processing failed")
	}

	pd.LogOperation("process_preview", success, duration, details, err)
	pd.LogEvent("processing_complete", mode, details)
}

func (pd *PipelineDebugger) LogLayerProcessing(layerCount int, inputSize, outputSize string, success bool, duration time.Duration, err error) {
	details := map[string]interface{}{
		"layer_count": layerCount,
		"input_size":  inputSize,
		"output_size": outputSize,
	}
	pd.LogOperation("layer_processing", success, duration, details, err)
}

func (pd *PipelineDebugger) LogMatConversion(inputSize, outputSize string, success bool, duration time.Duration, err error) {
	details := map[string]interface{}{
		"input_size":  inputSize,
		"output_size": outputSize,
	}
	pd.LogOperation("mat_conversion", success, duration, details, err)
}

func (pd *PipelineDebugger) LogMetricsCalculation(psnr, ssim float64, duration time.Duration, err error) {
	success := err == nil
	details := map[string]interface{}{
		"psnr": psnr,
		"ssim": ssim,
	}
	pd.LogOperation("calculate_metrics", success, duration, details, err)
}

func (pd *PipelineDebugger) LogPreviewTrigger(reason string, hasImage bool, layerCount int) {
	details := map[string]interface{}{
		"reason":      reason,
		"has_image":   hasImage,
		"layer_count": layerCount,
	}
	pd.LogEvent("preview_trigger", pd.currentMode, details)
}

func (pd *PipelineDebugger) LogError(component, operation, error string) {
	details := map[string]interface{}{
		"component": component,
		"operation": operation,
		"error":     error,
	}
	pd.LogEvent("error", pd.currentMode, details)
	pd.logger.Error("PIPELINE Error", "component", component, "operation", operation, "error", error)
}

// Status and reporting
func (pd *PipelineDebugger) PrintStatus() {
	if !pd.enabled {
		return
	}

	fmt.Println("\n=== PIPELINE DEBUG STATUS ===")
	fmt.Printf("Current Mode: %s\n", pd.currentMode)
	fmt.Printf("Current Layer Count: %d\n", pd.currentLayerCount)
	fmt.Printf("Total Operations: %d\n", len(pd.operations))
	fmt.Printf("Total Events: %d\n", len(pd.events))

	// Performance summary
	if len(pd.processingTimes) > 0 {
		avg := pd.averageDuration(pd.processingTimes)
		fmt.Printf("Average Processing Time: %v\n", avg)
	}

	if len(pd.conversionTimes) > 0 {
		avg := pd.averageDuration(pd.conversionTimes)
		fmt.Printf("Average Conversion Time: %v\n", avg)
	}

	// Recent operations
	fmt.Println("\nRecent Operations:")
	recentCount := 5
	if len(pd.operations) < recentCount {
		recentCount = len(pd.operations)
	}

	for i := len(pd.operations) - recentCount; i < len(pd.operations); i++ {
		op := pd.operations[i]
		status := "SUCCESS"
		if !op.Success {
			status = "FAILED"
		}
		fmt.Printf("  [%s] %s - %s (%v)\n",
			op.Timestamp.Format("15:04:05.000"),
			op.Operation,
			status,
			op.Duration)
	}

	// Recent events
	fmt.Println("\nRecent Events:")
	recentEventCount := 5
	if len(pd.events) < recentEventCount {
		recentEventCount = len(pd.events)
	}

	for i := len(pd.events) - recentEventCount; i < len(pd.events); i++ {
		event := pd.events[i]
		fmt.Printf("  [%s] %s (%s mode)\n",
			event.Timestamp.Format("15:04:05.000"),
			event.Event,
			event.Mode)
	}
}

func (pd *PipelineDebugger) GetStats() map[string]interface{} {
	if !pd.enabled {
		return nil
	}

	stats := map[string]interface{}{
		"current_mode":        pd.currentMode,
		"current_layer_count": pd.currentLayerCount,
		"total_operations":    len(pd.operations),
		"total_events":        len(pd.events),
	}

	// Success rates
	successCount := 0
	for _, op := range pd.operations {
		if op.Success {
			successCount++
		}
	}
	if len(pd.operations) > 0 {
		stats["success_rate"] = float64(successCount) / float64(len(pd.operations))
	}

	// Performance averages
	if len(pd.processingTimes) > 0 {
		stats["avg_processing_time"] = pd.averageDuration(pd.processingTimes)
	}
	if len(pd.conversionTimes) > 0 {
		stats["avg_conversion_time"] = pd.averageDuration(pd.conversionTimes)
	}

	return stats
}

func (pd *PipelineDebugger) averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}

// Global pipeline debugger instance
var GlobalPipelineDebugger *PipelineDebugger

func InitPipelineDebugger(logger *slog.Logger) {
	GlobalPipelineDebugger = NewPipelineDebugger(logger)
}
