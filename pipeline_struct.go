package main

import (
	"sync"

	"gocv.io/x/gocv"
)

type ImagePipeline struct {
	originalImage   gocv.Mat
	processedImage  gocv.Mat
	previewImage    gocv.Mat
	transformations []Transformation
	debugPipeline   *DebugPipeline
	initialized     int32
	mutex           sync.RWMutex
	processingMutex sync.Mutex
}

func NewImagePipeline(config *DebugConfig) *ImagePipeline {
	return &ImagePipeline{
		transformations: make([]Transformation, 0),
		debugPipeline:   NewDebugPipeline(config),
		originalImage:   gocv.NewMat(),
		processedImage:  gocv.NewMat(),
		previewImage:    gocv.NewMat(),
	}
}
