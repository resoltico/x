package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"advanced-image-processing/internal/gui"
)

func main() {
	// Parse command line flags
	var (
		debugMode = flag.Bool("debug", false, "Enable debug mode")
		version   = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	// Show version if requested
	if *version {
		fmt.Println("Image Restoration Suite v2.0")
		fmt.Println("Advanced image processing with Perfect UI Design")
		return
	}

	// Setup logger with appropriate level
	logLevel := slog.LevelInfo
	if *debugMode {
		logLevel = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	// Log startup information
	logger.Info("Starting Image Restoration Suite v2.0 with Perfect UI Design",
		"version", "2.0.0",
		"debug_mode", *debugMode)

	// Create and initialize the GUI application
	app := gui.NewApplication(logger) // Only 1 argument - the logger

	if err := app.Initialize(); err != nil {
		logger.Error("Failed to initialize application", "error", err)
		os.Exit(1)
	}

	// Run the application (this will block until the app is closed)
	app.Run()
}
