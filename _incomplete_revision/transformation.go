package main

import (
	"sync"

	"fyne.io/fyne/v2"
	"gocv.io/x/gocv"
)

type Transformation interface {
	Name() string
	Apply(input gocv.Mat) gocv.Mat        // Full resolution for saving
	ApplyPreview(input gocv.Mat) gocv.Mat // Optimized for real-time preview
	GetParametersWidget(onParameterChanged func()) fyne.CanvasObject
	GetParameters() map[string]interface{}
	SetParameters(params map[string]interface{})
	Close() // For cleanup of resources
}

// ThreadSafeTransformation provides base thread safety for transformations
type ThreadSafeTransformation struct {
	mutex sync.RWMutex
}

func (t *ThreadSafeTransformation) LockRead() {
	t.mutex.RLock()
}

func (t *ThreadSafeTransformation) UnlockRead() {
	t.mutex.RUnlock()
}

func (t *ThreadSafeTransformation) LockWrite() {
	t.mutex.Lock()
}

func (t *ThreadSafeTransformation) UnlockWrite() {
	t.mutex.Unlock()
}