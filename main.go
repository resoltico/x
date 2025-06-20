package main

import (
	"log"
	"net/http"
	_ "net/http/pprof" // FIXED: Enable pprof profiling server

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"gocv.io/x/gocv"
)

// DebugConfig controls which debug modules are enabled
type DebugConfig struct {
	GUI      bool
	Image    bool
	Pipeline bool
	Render   bool
}

// Global debug configuration - centralized control
var debugConfig = DebugConfig{
	GUI:      true, // Toggle GUI debugging
	Image:    true, // Toggle image debugging
	Pipeline: true, // Toggle pipeline debugging
	Render:   true, // Toggle render debugging
}

func main() {
	// FIXED: Start pprof server in development mode for memory profiling
	go func() {
		log.Println("Starting pprof server on :6060")
		log.Println("Memory profiler available at: http://localhost:6060/debug/pprof/")
		log.Println("Mat-specific profiling at: http://localhost:6060/debug/pprof/gocv.io/x/gocv.Mat")

		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Printf("Failed to start pprof server: %v", err)
		}
	}()

	// Log initial MatProfile count (only when built with -tags matprofile)
	initialCount := gocv.MatProfile.Count()
	log.Printf("Initial MatProfile count: %d", initialCount)

	if initialCount > 0 {
		log.Printf("WARNING: Non-zero initial MatProfile count detected!")
	}

	myApp := app.NewWithID("com.imagerestoration.suite")
	myWindow := myApp.NewWindow("Image Restoration Suite")
	myWindow.Resize(fyne.NewSize(1600, 900))

	ui := NewImageRestorationUI(myWindow, &debugConfig)
	myWindow.SetContent(ui.BuildUI())

	// FIXED: Enhanced cleanup and memory leak detection on close
	myWindow.SetOnClosed(func() {
		log.Printf("Application closing - performing cleanup...")

		// Allow time for cleanup operations
		if ui.pipeline != nil {
			ui.pipeline.Close()
		}

		// Force garbage collection before final count
		log.Printf("Running garbage collection...")
		// Note: Go GC won't clean up OpenCV Mats, but let's be thorough
		// runtime.GC() // Commented out as it won't help with OpenCV memory

		finalCount := gocv.MatProfile.Count()
		log.Printf("Final MatProfile count: %d", finalCount)

		if finalCount > 0 {
			log.Printf("WARNING: Memory leaks detected! %d Mat(s) not properly closed.", finalCount)
			log.Printf("Check MatProfile for details at: http://localhost:6060/debug/pprof/gocv.io/x/gocv.Mat")
			log.Printf("This indicates missing defer mat.Close() calls in the code.")
		} else {
			log.Printf("SUCCESS: No memory leaks detected - all Mats properly closed.")
		}

		// Report memory usage change
		memoryChange := finalCount - initialCount
		if memoryChange > 0 {
			log.Printf("MEMORY LEAK: %d Mat(s) were created but not cleaned up during session", memoryChange)
		} else if memoryChange < 0 {
			log.Printf("UNUSUAL: More Mats were cleaned up than created (initial: %d, final: %d)", initialCount, finalCount)
		} else {
			log.Printf("CLEAN EXIT: Mat allocation/deallocation balanced")
		}
	})

	log.Printf("Starting Image Restoration Suite...")
	log.Printf("Debug configuration: GUI=%t, Image=%t, Pipeline=%t, Render=%t",
		debugConfig.GUI, debugConfig.Image, debugConfig.Pipeline, debugConfig.Render)

	myWindow.ShowAndRun()
}
