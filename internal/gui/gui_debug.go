// internal/gui/gui_debug.go
// GUI-specific debugging - separated from core pipeline debugging
package gui

import (
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
)

// GUIDebugger provides GUI-specific debugging (UI interactions, display operations)
type GUIDebugger struct {
	logger    *slog.Logger
	enabled   bool
	startTime time.Time

	// GUI Component states
	toolbarState     map[string]interface{}
	leftPanelState   map[string]interface{}
	centerPanelState map[string]interface{}
	rightPanelState  map[string]interface{}

	// GUI-specific operations
	uiInteractions []UIInteraction
	displayOps     []DisplayOperation

	// Error tracking
	buildErrors   []string
	runtimeErrors []string

	// Performance tracking
	renderTimes []time.Duration
	layoutTimes []time.Duration
}

// UIInteraction tracks user interface interactions
type UIInteraction struct {
	Timestamp time.Time
	Component string // "Toolbar", "LeftPanel", "CenterPanel", "RightPanel"
	Action    string // "button_click", "slider_change", "dropdown_select", etc.
	Data      map[string]interface{}
	Duration  time.Duration
}

// DisplayOperation tracks image display operations
type DisplayOperation struct {
	Timestamp time.Time
	Operation string // "update_original", "update_preview", "view_change", "zoom_change"
	Success   bool
	Details   map[string]interface{}
	Duration  time.Duration
}

func NewGUIDebugger(logger *slog.Logger) *GUIDebugger {
	return &GUIDebugger{
		logger:           logger,
		enabled:          true,
		startTime:        time.Now(),
		toolbarState:     make(map[string]interface{}),
		leftPanelState:   make(map[string]interface{}),
		centerPanelState: make(map[string]interface{}),
		rightPanelState:  make(map[string]interface{}),
		uiInteractions:   make([]UIInteraction, 0),
		displayOps:       make([]DisplayOperation, 0),
		buildErrors:      make([]string, 0),
		runtimeErrors:    make([]string, 0),
		renderTimes:      make([]time.Duration, 0),
		layoutTimes:      make([]time.Duration, 0),
	}
}

// UI Interaction logging
func (d *GUIDebugger) LogUIInteraction(component, action string, data map[string]interface{}) {
	if !d.enabled {
		return
	}

	var duration time.Duration
	if startTime, ok := data["start_time"].(time.Time); ok {
		duration = time.Since(startTime)
		delete(data, "start_time") // Remove from data to avoid logging it
	}

	interaction := UIInteraction{
		Timestamp: time.Now(),
		Component: component,
		Action:    action,
		Data:      data,
		Duration:  duration,
	}

	d.uiInteractions = append(d.uiInteractions, interaction)

	// Update component state
	switch component {
	case "Toolbar":
		d.updateToolbarState(action, data)
	case "LeftPanel":
		d.updateLeftPanelState(action, data)
	case "CenterPanel":
		d.updateCenterPanelState(action, data)
	case "RightPanel":
		d.updateRightPanelState(action, data)
	}

	d.logger.Debug("GUI UI Interaction",
		"component", component,
		"action", action,
		"data", data,
		"duration_ms", duration.Milliseconds())
}

// Display operation logging
func (d *GUIDebugger) LogDisplayOperation(operation string, success bool, details map[string]interface{}) {
	if !d.enabled {
		return
	}

	var duration time.Duration
	if startTime, ok := details["start_time"].(time.Time); ok {
		duration = time.Since(startTime)
		delete(details, "start_time")
	}

	displayOp := DisplayOperation{
		Timestamp: time.Now(),
		Operation: operation,
		Success:   success,
		Details:   details,
		Duration:  duration,
	}

	d.displayOps = append(d.displayOps, displayOp)

	level := slog.LevelDebug
	if !success {
		level = slog.LevelError
	}

	d.logger.Log(nil, level, "GUI Display Operation",
		"operation", operation,
		"success", success,
		"details", details,
		"duration_ms", duration.Milliseconds())
}

// Component state updates
func (d *GUIDebugger) updateToolbarState(action string, data map[string]interface{}) {
	switch action {
	case "zoom_set":
		d.toolbarState["current_zoom"] = data["new_zoom"]
	case "view_set":
		d.toolbarState["current_view"] = data["new_view"]
	case "processing_buttons_enabled", "processing_buttons_disabled":
		d.toolbarState["save_enabled"] = data["save_enabled"]
		d.toolbarState["reset_enabled"] = data["reset_enabled"]
	}
	d.toolbarState["last_action"] = action
	d.toolbarState["last_update"] = time.Now()
}

func (d *GUIDebugger) updateLeftPanelState(action string, data map[string]interface{}) {
	switch action {
	case "layer_added":
		if count, ok := d.leftPanelState["layer_count"].(int); ok {
			d.leftPanelState["layer_count"] = count + 1
		} else {
			d.leftPanelState["layer_count"] = 1
		}
	case "layer_deleted":
		if count, ok := d.leftPanelState["layer_count"].(int); ok && count > 0 {
			d.leftPanelState["layer_count"] = count - 1
		}
	case "layer_selected":
		d.leftPanelState["selected_layer"] = data["layer_id"]
	}
	d.leftPanelState["last_action"] = action
	d.leftPanelState["last_update"] = time.Now()
}

func (d *GUIDebugger) updateCenterPanelState(action string, data map[string]interface{}) {
	switch action {
	case "view_mode_change":
		d.centerPanelState["current_view_mode"] = data["new_mode"]
	case "zoom_change":
		d.centerPanelState["current_zoom"] = data["zoom"]
	case "update_original_success":
		d.centerPanelState["has_original"] = true
		d.centerPanelState["original_bounds"] = data["image_bounds"]
	case "update_preview_success":
		d.centerPanelState["has_preview"] = true
		d.centerPanelState["preview_bounds"] = data["image_bounds"]
	case "center_panel_reset":
		d.centerPanelState["has_original"] = false
		d.centerPanelState["has_preview"] = false
	}
	d.centerPanelState["last_action"] = action
	d.centerPanelState["last_update"] = time.Now()
}

func (d *GUIDebugger) updateRightPanelState(action string, data map[string]interface{}) {
	switch action {
	case "metrics_updated":
		d.rightPanelState["psnr"] = data["psnr"]
		d.rightPanelState["ssim"] = data["ssim"]
	case "status_updated":
		d.rightPanelState["status"] = data["status"]
	}
	d.rightPanelState["last_action"] = action
	d.rightPanelState["last_update"] = time.Now()
}

// Image operation convenience method (for backward compatibility)
func (d *GUIDebugger) LogImageOperation(operation string, success bool, details map[string]interface{}) {
	d.LogDisplayOperation(operation, success, details)
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

// Left Panel specific debug methods
func (d *GUIDebugger) LogLayerAddAttempt(algorithm string, params map[string]interface{}) {
	d.LogUIInteraction("LeftPanel", "layer_add_attempt", map[string]interface{}{
		"algorithm": algorithm,
		"params":    params,
	})
}

func (d *GUIDebugger) LogLayerAddComplete(layerID string, success bool, err error) {
	data := map[string]interface{}{
		"layer_id": layerID,
		"success":  success,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	d.LogUIInteraction("LeftPanel", "layer_add_complete", data)
}

func (d *GUIDebugger) LogLayerOperation(operation, layerID, algorithm string, params map[string]interface{}, success bool, err error, duration time.Duration) {
	data := map[string]interface{}{
		"operation": operation,
		"layer_id":  layerID,
		"algorithm": algorithm,
		"params":    params,
		"success":   success,
		"duration":  duration,
	}
	if err != nil {
		data["error"] = err.Error()
	}
	d.LogUIInteraction("LeftPanel", "layer_operation", data)
}

func (d *GUIDebugger) LogProcessingEvent(event, mode string, layerCount int, details map[string]interface{}) {
	data := map[string]interface{}{
		"event":       event,
		"mode":        mode,
		"layer_count": layerCount,
		"details":     details,
	}
	d.LogUIInteraction("LeftPanel", "processing_event", data)
}

func (d *GUIDebugger) LogPreviewTrigger(reason string, layerCount int) {
	d.LogUIInteraction("LeftPanel", "preview_trigger", map[string]interface{}{
		"reason":      reason,
		"layer_count": layerCount,
	})
}

// Status reporting
func (d *GUIDebugger) PrintStatus() {
	if !d.enabled {
		return
	}

	fmt.Println("\n=== GUI DEBUG STATUS ===")
	fmt.Printf("Runtime: %v\n", time.Since(d.startTime))
	fmt.Printf("UI Interactions: %d\n", len(d.uiInteractions))
	fmt.Printf("Display Operations: %d\n", len(d.displayOps))
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

	// Recent UI interactions
	fmt.Println("\nRecent UI Interactions:")
	recentCount := 5
	if len(d.uiInteractions) < recentCount {
		recentCount = len(d.uiInteractions)
	}

	for i := len(d.uiInteractions) - recentCount; i < len(d.uiInteractions); i++ {
		interaction := d.uiInteractions[i]
		fmt.Printf("  [%s] %s.%s (%v)\n",
			interaction.Timestamp.Format("15:04:05.000"),
			interaction.Component,
			interaction.Action,
			interaction.Duration)
	}

	// Recent display operations
	fmt.Println("\nRecent Display Operations:")
	recentDisplayCount := 5
	if len(d.displayOps) < recentDisplayCount {
		recentDisplayCount = len(d.displayOps)
	}

	for i := len(d.displayOps) - recentDisplayCount; i < len(d.displayOps); i++ {
		op := d.displayOps[i]
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
}

func (d *GUIDebugger) DumpComponentStates() {
	if !d.enabled {
		return
	}

	fmt.Println("\n=== GUI COMPONENT STATES ===")

	fmt.Println("Toolbar State:")
	for key, value := range d.toolbarState {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("Left Panel State:")
	for key, value := range d.leftPanelState {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("Center Panel State:")
	for key, value := range d.centerPanelState {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("Right Panel State:")
	for key, value := range d.rightPanelState {
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

	d.LogDisplayOperation("load", success, map[string]interface{}{
		"filepath": filepath,
		"width":    width,
		"height":   height,
		"channels": channels,
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
