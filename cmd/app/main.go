// Advanced Image Processing Application
// Author: Ervins Strauhmanis
// License: MIT
// Version: 2.0.0 - Optimized with Standard APIs

package main

import (
	"flag"
	"log/slog"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"advanced-image-processing/internal/gui"
)

const (
	AppName    = "Advanced Image Processing"
	AppID      = "com.strauhmanis.advanced-image-processing"
	AppVersion = "2.0.0"
)

func main() {
	debugMode := flag.Bool("debug", false, "Enable debug mode with verbose logging")
	flag.Parse()

	// Initialize logger
	logger := initLogger(*debugMode)
	logger.Info("Starting Advanced Image Processing Application",
		"version", AppVersion,
		"debug_mode", *debugMode)

	// Create Fyne application
	myApp := app.NewWithID(AppID)
	myApp.SetIcon(theme.DocumentIcon())
	myApp.Settings().SetTheme(theme.DefaultTheme())

	// Create and show main application window
	mainApp := gui.NewApplication(myApp, logger, *debugMode)
	mainApp.ShowAndRun()

	logger.Info("Application shutting down gracefully")
	os.Exit(0)
}

func initLogger(debugMode bool) *slog.Logger {
	var handler slog.Handler

	if debugMode {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	return slog.New(handler)
}
