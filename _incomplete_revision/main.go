package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// DebugConfig controls which debug modules are enabled
type DebugConfig struct {
	GUI      bool
	Image    bool
	Memory   bool
	Pipeline bool
	Render   bool
}

// Global debug configuration - centralized control
var debugConfig = DebugConfig{
	GUI:      true,  // Enable GUI debugging
	Image:    false, // Disable image debugging for performance
	Memory:   true,  // Enable memory debugging
	Pipeline: true,  // Enable pipeline debugging
	Render:   false, // Disable render debugging for performance
}

func main() {
	myApp := app.NewWithID("com.imagerestoration.suite")
	myWindow := myApp.NewWindow("Image Restoration Suite")
	myWindow.Resize(fyne.NewSize(1600, 900))

	ui := NewImageRestorationUI(myWindow, &debugConfig)
	myWindow.SetContent(ui.BuildUI())

	myWindow.ShowAndRun()
}