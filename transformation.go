package main

import (
	"fyne.io/fyne/v2"
	"gocv.io/x/gocv"
)

type Transformation interface {
	Name() string
	Apply(input gocv.Mat) gocv.Mat           // Full resolution for saving
	ApplyPreview(input gocv.Mat) gocv.Mat    // Optimized for real-time preview
	GetParametersWidget(onParameterChanged func()) fyne.CanvasObject
	GetParameters() map[string]interface{}
	SetParameters(params map[string]interface{})
	Close() // For cleanup of resources
}