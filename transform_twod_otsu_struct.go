package main

import (
	"sync"

	"gocv.io/x/gocv"
)

type TwoDOtsu struct {
	ThreadSafeTransformation
	debugImage *DebugImage

	paramMutex       sync.RWMutex
	windowRadius     int
	epsilon          float64
	morphKernelSize  int
	noiseReduction   bool
	useIntegralImage bool
	adaptiveRegions  int

	onParameterChanged func()
}

func NewTwoDOtsu(config *DebugConfig) *TwoDOtsu {
	return &TwoDOtsu{
		debugImage:       NewDebugImage(config),
		windowRadius:     5,
		epsilon:          0.02,
		morphKernelSize:  3,
		noiseReduction:   true,
		useIntegralImage: true,
		adaptiveRegions:  4, // Try reducing to 2 to lower processing complexity
	}
}

func (t *TwoDOtsu) Name() string {
	return "2D Otsu"
}

func (t *TwoDOtsu) Close() {
	// No resources to cleanup - GoCV MatProfile handles tracking
}

func (t *TwoDOtsu) Apply(src gocv.Mat) gocv.Mat {
	return t.applyWithScale(src, 1.0)
}

func (t *TwoDOtsu) ApplyPreview(src gocv.Mat) gocv.Mat {
	return t.applyWithScale(src, 0.5)
}
