package main

import (
	"log"

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
	// Log initial MatProfile count (only when built with -tags matprofile)
	log.Printf("Initial MatProfile count: %d", gocv.MatProfile.Count())

	myApp := app.NewWithID("com.imagerestoration.suite")
	myWindow := myApp.NewWindow("Image Restoration Suite")
	myWindow.Resize(fyne.NewSize(1600, 900))

	ui := NewImageRestorationUI(myWindow, &debugConfig)
	myWindow.SetContent(ui.BuildUI())

	// Log final MatProfile count on close
	myWindow.SetOnClosed(func() {
		log.Printf("Final MatProfile count: %d", gocv.MatProfile.Count())
		if gocv.MatProfile.Count() > 0 {
			log.Println("WARNING: Memory leaks detected! Check MatProfile for details.")
		}
	})

	myWindow.ShowAndRun()
}
