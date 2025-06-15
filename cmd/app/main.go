// Advanced Image Processing Application
// Author: Ervins Strauhmanis
// License: MIT

package main

import (
	"flag"
	"os"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"advanced-image-processing/internal/gui"
	"advanced-image-processing/internal/utils"
)

func main() {
	// Parse command line flags
	debugMode := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	// Initialize logger
	logger := utils.InitLogger(*debugMode)
	logger.Info("Starting Advanced Image Processing Application")

	// Create Fyne application
	myApp := app.NewWithID("com.strauhmanis.advanced-image-processing")
	myApp.SetIcon(theme.DocumentIcon())
	myApp.Settings().SetTheme(theme.DefaultTheme())

	// Create main window
	mainWindow := gui.NewMainWindow(myApp, logger, *debugMode)
	mainWindow.ShowAndRun()

	logger.Info("Application shutting down")
	os.Exit(0)
}