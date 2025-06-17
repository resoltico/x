package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	myApp := app.NewWithID("com.imagerestoration.suite")
	myWindow := myApp.NewWindow("Image Restoration Suite")
	myWindow.Resize(fyne.NewSize(1600, 900))

	ui := NewImageRestorationUI(myWindow)
	myWindow.SetContent(ui.BuildUI())

	myWindow.ShowAndRun()
}
