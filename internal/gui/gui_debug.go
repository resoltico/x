package gui

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
)

// GUIDebugger provides comprehensive debugging for GUI components
type GUIDebugger struct {
	logger    *slog.Logger
	enabled   bool
	startTime time.Time

	// Component states
	toolbarState map[string]interface{}
	panelStates  map[string]map[string]interface{}
	imageState   map[string]interface{}

	// Error tracking
	buildErrors   []string
	runtimeErrors []string

	// Performance tracking
	renderTimes []time.Duration
	layoutTimes []time.Duration
}

func NewGUIDebugger(logger *slog.Logger) *GUIDebugger {
	return &GUIDebugger{
		logger:        logger,
		enabled:       true, // Always enabled - no need to remember GUI_DEBUG=true
		startTime:     time.Now(),
		toolbarState:  make(map[string]interface{}),
		panelStates:   make(map[string]map[string]interface{}),
		imageState:    make(map[string]interface{}),
		buildErrors:   []string{},
		runtimeErrors: []string{},
		renderTimes:   []time.Duration{},
		layoutTimes:   []time.Duration{},
	}
}

// Build-time debugging
func (d *GUIDebugger) LogBuildError(component, method, expected, got string) {
	if !d.enabled {
		return
	}

	errorMsg := fmt.Sprintf("BUILD ERROR [%s]: Method '%s' - Expected: %s, Got: %s",
		component, method, expected, got)
	d.buildErrors = append(d.buildErrors, errorMsg)
	d.logger.Error("GUI Build Error",
		"component", component,
		"method", method,
		"expected", expected,
		"got", got)
}

func (d *GUIDebugger) LogMissingCallback(component, callback string) {
	if !d.enabled {
		return
	}

	errorMsg := fmt.Sprintf("MISSING CALLBACK [%s]: %s not found", component, callback)
	d.buildErrors = append(d.buildErrors, errorMsg)
	d.logger.Error("GUI Missing Callback",
		"component", component,
		"callback", callback)
}

func (d *GUIDebugger) LogConstructorError(component string, expectedArgs, gotArgs int) {
	if !d.enabled {
		return
	}

	errorMsg := fmt.Sprintf("CONSTRUCTOR ERROR [%s]: Expected %d args, got %d",
		component, expectedArgs, gotArgs)
	d.buildErrors = append(d.buildErrors, errorMsg)
	d.logger.Error("GUI Constructor Error",
		"component", component,
		"expected_args", expectedArgs,
		"got_args", gotArgs)
}

// Runtime debugging
func (d *GUIDebugger) LogPanelState(panelName string, state map[string]interface{}) {
	if !d.enabled {
		return
	}

	if d.panelStates[panelName] == nil {
		d.panelStates[panelName] = make(map[string]interface{})
	}

	for key, value := range state {
		d.panelStates[panelName][key] = value
	}

	d.logger.Debug("GUI Panel State Updated",
		"panel", panelName,
		"state", state)
}

func (d *GUIDebugger) LogImageOperation(operation string, success bool, details map[string]interface{}) {
	if !d.enabled {
		return
	}

	d.imageState[operation] = map[string]interface{}{
		"success":   success,
		"timestamp": time.Now(),
		"details":   details,
	}

	level := slog.LevelDebug
	if !success {
		level = slog.LevelError
	}

	d.logger.Log(nil, level, "GUI Image Operation",
		"operation", operation,
		"success", success,
		"details", details)
}

func (d *GUIDebugger) LogUIInteraction(component, action string, data map[string]interface{}) {
	if !d.enabled {
		return
	}

	d.logger.Debug("GUI UI Interaction",
		"component", component,
		"action", action,
		"data", data,
		"timestamp", time.Now())
}

func (d *GUIDebugger) LogPerformance(operation string, duration time.Duration) {
	if !d.enabled {
		return
	}

	switch operation {
	case "render":
		d.renderTimes = append(d.renderTimes, duration)
	case "layout":
		d.layoutTimes = append(d.layoutTimes, duration)
	}

	d.logger.Debug("GUI Performance",
		"operation", operation,
		"duration_ms", duration.Milliseconds())
}

func (d *GUIDebugger) LogRuntimeError(component, error string) {
	if !d.enabled {
		return
	}

	errorMsg := fmt.Sprintf("RUNTIME ERROR [%s]: %s", component, error)
	d.runtimeErrors = append(d.runtimeErrors, errorMsg)
	d.logger.Error("GUI Runtime Error",
		"component", component,
		"error", error,
		"stack", getStackTrace())
}

// Current build errors from compilation
func (d *GUIDebugger) LogCurrentBuildIssues() {
	if !d.enabled {
		return
	}

	d.logger.Error("=== CURRENT BUILD ISSUES ===")

	// Core constructor issues
	d.LogConstructorError("NewEnhancedPipeline", 3, 1)
	d.LogConstructorError("NewRegionManager", 0, 1)

	// Missing toolbar callbacks
	d.LogMissingCallback("Toolbar", "onOpenImage")
	d.LogMissingCallback("Toolbar", "onSaveImage")
	d.LogMissingCallback("Toolbar", "onViewModeChanged")

	// Missing ImageData methods
	d.LogBuildError("ImageData", "LoadFromFile", "LoadFromFile(string) error", "method not found")
	d.LogBuildError("ImageData", "GetDimensions", "GetDimensions() (int, int, int)", "method not found")

	// Dialog API issues
	d.LogBuildError("dialog.NewFileSave", "callback signature", "func(fyne.URIWriteCloser, error)", "func(fyne.URIWriteCloser)")
	d.LogBuildError("saveDialog.SetFilter", "parameter type", "storage.FileFilter", "[]string")
}

// Status reporting
func (d *GUIDebugger) PrintStatus() {
	if !d.enabled {
		return
	}

	fmt.Println("\n=== GUI DEBUG STATUS ===")
	fmt.Printf("Runtime: %v\n", time.Since(d.startTime))
	fmt.Printf("Build Errors: %d\n", len(d.buildErrors))
	fmt.Printf("Runtime Errors: %d\n", len(d.runtimeErrors))

	if len(d.buildErrors) > 0 {
		fmt.Println("\nBuild Errors:")
		for _, err := range d.buildErrors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if len(d.runtimeErrors) > 0 {
		fmt.Println("\nRuntime Errors:")
		for _, err := range d.runtimeErrors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if len(d.renderTimes) > 0 {
		avg := averageDuration(d.renderTimes)
		fmt.Printf("\nAverage Render Time: %v\n", avg)
	}

	if len(d.layoutTimes) > 0 {
		avg := averageDuration(d.layoutTimes)
		fmt.Printf("Average Layout Time: %v\n", avg)
	}
}

func (d *GUIDebugger) DumpComponentStates() {
	if !d.enabled {
		return
	}

	fmt.Println("\n=== COMPONENT STATES ===")

	fmt.Println("Toolbar State:")
	for key, value := range d.toolbarState {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("Panel States:")
	for panel, state := range d.panelStates {
		fmt.Printf("  %s:\n", panel)
		for key, value := range state {
			fmt.Printf("    %s: %v\n", key, value)
		}
	}

	fmt.Println("Image State:")
	for key, value := range d.imageState {
		fmt.Printf("  %s: %v\n", key, value)
	}
}

// Helper functions
func getStackTrace() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

func averageDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}

// Convenience methods for common debugging scenarios
func (d *GUIDebugger) DebugWindowResize(oldSize, newSize fyne.Size) {
	if !d.enabled {
		return
	}

	d.LogUIInteraction("Window", "resize", map[string]interface{}{
		"old_width":  oldSize.Width,
		"old_height": oldSize.Height,
		"new_width":  newSize.Width,
		"new_height": newSize.Height,
	})
}

func (d *GUIDebugger) DebugImageLoad(filepath string, success bool, width, height, channels int) {
	if !d.enabled {
		return
	}

	d.LogImageOperation("load", success, map[string]interface{}{
		"filepath": filepath,
		"width":    width,
		"height":   height,
		"channels": channels,
	})
}

func (d *GUIDebugger) DebugLayerOperation(operation, layerID, algorithm string) {
	if !d.enabled {
		return
	}

	d.LogUIInteraction("LayerManager", operation, map[string]interface{}{
		"layer_id":  layerID,
		"algorithm": algorithm,
	})
}

// Global debugger instance
var GlobalGUIDebugger *GUIDebugger

func InitGUIDebugger(logger *slog.Logger) {
	GlobalGUIDebugger = NewGUIDebugger(logger)
	if GlobalGUIDebugger.enabled {
		GlobalGUIDebugger.LogCurrentBuildIssues()
	}
}
