// Advanced Image Processing Application - Complete Rewrite
// Author: Ervins Strauhmanis
// License: MIT
// Version: 2.0.0 - ROI + LAA + Full Metrics

package main

import (
	"flag"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"github.com/sirupsen/logrus"

	"advanced-image-processing/internal/gui"
)

const (
	AppName    = "Advanced Image Processing"
	AppID      = "com.strauhmanis.advanced-image-processing"
	AppVersion = "2.0.0"
)

func main() {
	// Parse command line flags
	debugMode := flag.Bool("debug", false, "Enable debug mode with verbose logging")
	flag.Parse()

	// Initialize logger
	logger := initLogger(*debugMode)
	logger.WithFields(logrus.Fields{
		"version":    AppVersion,
		"debug_mode": *debugMode,
	}).Info("Starting Advanced Image Processing Application")

	// Create Fyne application with updated metadata
	myApp := app.NewWithID(AppID)
	myApp.SetIcon(theme.DocumentIcon())

	// Note: In Fyne v2.6+, metadata is read-only and set from FyneApp.toml
	// The application metadata (name, version) is automatically loaded from FyneApp.toml

	// Use default theme (optimized for Fyne v2.6)
	myApp.Settings().SetTheme(theme.DefaultTheme())

	// Create and show main application window
	mainApp := gui.NewApplication(myApp, logger, *debugMode)
	mainApp.ShowAndRun()

	logger.Info("Application shutting down gracefully")
	os.Exit(0)
}

// initLogger initializes the logger with appropriate level
func initLogger(debugMode bool) *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	if debugMode {
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
			ForceColors:   true,
		})
		logger.Debug("Debug logging enabled")
	} else {
		logger.SetLevel(logrus.InfoLevel)
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	return logger
}
