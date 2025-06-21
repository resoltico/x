package main

import "gocv.io/x/gocv"

type Lanczos4Transform struct {
	debugImage   *DebugImage
	scaleFactor  float64
	targetDPI    float64
	originalDPI  float64
	useIterative bool

	onParameterChanged func()
}

func NewLanczos4Transform(config *DebugConfig) *Lanczos4Transform {
	return &Lanczos4Transform{
		debugImage:   NewDebugImage(config),
		scaleFactor:  2.0,
		targetDPI:    300.0,
		originalDPI:  150.0,
		useIterative: false,
	}
}

func (l *Lanczos4Transform) Name() string {
	return "Lanczos4 Scaling"
}

func (l *Lanczos4Transform) Close() {
	// No resources to cleanup
}

func (l *Lanczos4Transform) Apply(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4", "Starting full resolution scaling")
	return l.applyLanczos4(src, l.scaleFactor)
}

func (l *Lanczos4Transform) ApplyPreview(src gocv.Mat) gocv.Mat {
	l.debugImage.LogAlgorithmStep("Lanczos4 Preview", "Starting preview scaling")

	previewScale := l.scaleFactor
	if l.scaleFactor > 3.0 {
		previewScale = 3.0
	}

	result := l.applyLanczos4(src, previewScale)
	l.debugImage.LogAlgorithmStep("Lanczos4 Preview", "Preview scaling completed")
	return result
}
