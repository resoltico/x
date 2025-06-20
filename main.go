package main

import (
	"log"
	"net/http"
	_ "net/http/pprof" // Enable pprof profiling server
	"runtime"
	"time"

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
	// pprof server startup with error handling
	go func() {
		log.Println("Starting pprof server on :6060")
		log.Println("Memory profiler available at: http://localhost:6060/debug/pprof/")
		log.Println("Mat-specific profiling at: http://localhost:6060/debug/pprof/gocv.io/x/gocv.Mat")

		server := &http.Server{
			Addr:         "localhost:6060",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Failed to start pprof server: %v", err)
		}
	}()

	// Initial MatProfile logging
	initialCount := gocv.MatProfile.Count()
	log.Printf("=== MEMORY TRACKING INITIALIZED ===")
	log.Printf("Initial MatProfile count: %d", initialCount)

	if initialCount > 0 {
		log.Printf("WARNING: Non-zero initial MatProfile count detected!")
		log.Printf("This may indicate leftover Mats from previous sessions")
	}

	// Log Go runtime information
	log.Printf("Go version: %s", runtime.Version())
	log.Printf("GOMAXPROCS: %d", runtime.GOMAXPROCS(0))

	myApp := app.NewWithID("com.imagerestoration.suite")
	myWindow := myApp.NewWindow("Image Restoration Suite")
	myWindow.Resize(fyne.NewSize(1600, 900))

	// Create UI with error handling
	ui := NewImageRestorationUI(myWindow, &debugConfig)
	if ui == nil {
		log.Fatal("Failed to create UI")
	}

	content := ui.BuildUI()
	if content == nil {
		log.Fatal("Failed to build UI")
	}

	myWindow.SetContent(content)

	// Cleanup and memory leak detection with proper Mat cleanup
	myWindow.SetOnClosed(func() {
		log.Printf("=== APPLICATION SHUTDOWN INITIATED ===")
		log.Printf("Performing cleanup...")

		// Force multiple garbage collections before cleanup to ensure all go objects are collected
		for i := 0; i < 3; i++ {
			runtime.GC()
			time.Sleep(10 * time.Millisecond)
		}

		startCleanupTime := time.Now()

		// Pipeline cleanup with proper Mat resource management
		if ui != nil && ui.pipeline != nil {
			log.Printf("Closing image pipeline...")
			ui.pipeline.Close()
			log.Printf("Pipeline closed successfully")
		}

		// Allow more time for all goroutines to complete cleanup
		log.Printf("Waiting for cleanup operations to complete...")
		time.Sleep(200 * time.Millisecond)

		// Force final garbage collection multiple times
		log.Printf("Running final garbage collection...")
		for i := 0; i < 5; i++ {
			runtime.GC()
			time.Sleep(20 * time.Millisecond)
		}

		cleanupDuration := time.Since(startCleanupTime)
		log.Printf("Cleanup completed in %v", cleanupDuration)

		// Memory analysis
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)

		finalCount := gocv.MatProfile.Count()
		memoryChange := finalCount - initialCount

		log.Printf("=== FINAL MEMORY ANALYSIS ===")
		log.Printf("Initial MatProfile count: %d", initialCount)
		log.Printf("Final MatProfile count: %d", finalCount)
		log.Printf("Net change: %+d Mats", memoryChange)

		// Memory statistics
		log.Printf("Go Memory Stats:")
		log.Printf("  Allocated: %.2f MB", float64(memStats.Alloc)/1024/1024)
		log.Printf("  Total Allocated: %.2f MB", float64(memStats.TotalAlloc)/1024/1024)
		log.Printf("  System Memory: %.2f MB", float64(memStats.Sys)/1024/1024)
		log.Printf("  GC Runs: %d", memStats.NumGC)
		log.Printf("  Last GC: %v ago", time.Since(time.Unix(0, int64(memStats.LastGC))))

		// Memory leak detection and reporting
		if finalCount > 0 {
			log.Printf("⚠️  WARNING: MEMORY LEAKS DETECTED!")
			log.Printf("   %d Mat(s) were not properly closed", finalCount)
			log.Printf("   This indicates missing defer mat.Close() calls")
			log.Printf("   Check MatProfile for details: http://localhost:6060/debug/pprof/gocv.io/x/gocv.Mat")

			if memoryChange > 0 {
				log.Printf("   🔥 ACTIVE LEAK: %d new Mat(s) created during session", memoryChange)
			} else if memoryChange < 0 {
				log.Printf("   ℹ️  Note: More Mats cleaned than created (initial=%d, final=%d)", initialCount, finalCount)
			}
		} else {
			log.Printf("✅ SUCCESS: No memory leaks detected")
			log.Printf("   All OpenCV Mats properly closed")

			if memoryChange == 0 {
				log.Printf("   🎯 PERFECT: Mat allocation/deallocation perfectly balanced")
			} else if memoryChange < 0 {
				log.Printf("   🧹 CLEANUP: %d pre-existing Mat(s) were cleaned up", -memoryChange)
			}
		}

		// Performance summary
		log.Printf("=== SESSION SUMMARY ===")
		if ui != nil && ui.pipeline != nil && ui.pipeline.debugPipeline != nil {
			operations := ui.pipeline.debugPipeline.GetOperationHistory()
			if len(operations) > 0 {
				totalDuration := time.Duration(0)
				for _, op := range operations {
					totalDuration += op.Duration
				}
				log.Printf("Total operations: %d", len(operations))
				log.Printf("Total processing time: %v", totalDuration)
				log.Printf("Average operation time: %v", totalDuration/time.Duration(len(operations)))
			}
		}

		log.Printf("Application shutdown complete.")
		log.Printf("=====================================")
	})

	// Startup logging
	log.Printf("=== STARTING IMAGE RESTORATION SUITE ===")
	log.Printf("Debug configuration:")
	log.Printf("  GUI debugging: %t", debugConfig.GUI)
	log.Printf("  Image debugging: %t", debugConfig.Image)
	log.Printf("  Pipeline debugging: %t", debugConfig.Pipeline)
	log.Printf("  Render debugging: %t", debugConfig.Render)
	log.Printf("========================================")

	myWindow.ShowAndRun()
}
