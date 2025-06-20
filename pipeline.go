package main

import (
	"fmt"
	"math"
	"sync"

	"gocv.io/x/gocv"
)

type ImagePipeline struct {
	originalImage   gocv.Mat
	processedImage  gocv.Mat
	previewImage    gocv.Mat
	transformations []Transformation
	debugPipeline   *DebugPipeline
	debugMemory     *DebugMemory
	initialized     bool
	mutex           sync.RWMutex
}

func NewImagePipeline(config *DebugConfig) *ImagePipeline {
	debugPipeline := NewDebugPipeline(config)
	debugMemory := NewDebugMemory(config)

	pipeline := &ImagePipeline{
		transformations: make([]Transformation, 0),
		debugPipeline:   debugPipeline,
		debugMemory:     debugMemory,
		initialized:     false,
		originalImage:   gocv.NewMat(),
		processedImage:  gocv.NewMat(),
		previewImage:    gocv.NewMat(),
	}

	return pipeline
}

func (p *ImagePipeline) HasImage() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.initialized && !p.originalImage.Empty()
}

func (p *ImagePipeline) SetOriginalImage(img gocv.Mat) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.debugPipeline.LogSetOriginalStart()

	// Validate input
	if img.Empty() {
		return fmt.Errorf("input image is empty")
	}

	p.debugPipeline.LogSetOriginalStep("closing existing images")

	// Close existing images if they exist
	if p.initialized {
		if !p.originalImage.Empty() {
			p.originalImage.Close()
		}
		if !p.processedImage.Empty() {
			p.processedImage.Close()
		}
		if !p.previewImage.Empty() {
			p.previewImage.Close()
		}
	}

	// Create new images
	p.debugPipeline.LogSetOriginalStep("cloning images safely")

	p.originalImage = img.Clone()
	if p.originalImage.Empty() {
		return fmt.Errorf("failed to clone original image")
	}

	p.processedImage = img.Clone()
	if p.processedImage.Empty() {
		p.originalImage.Close()
		return fmt.Errorf("failed to clone processed image")
	}

	p.previewImage = img.Clone()
	if p.previewImage.Empty() {
		p.originalImage.Close()
		p.processedImage.Close()
		return fmt.Errorf("failed to clone preview image")
	}

	p.initialized = true
	p.debugPipeline.LogImageStats("original", p.originalImage)

	// Process images only if we have transformations
	if len(p.transformations) > 0 {
		if err := p.processImageUnsafe(); err != nil {
			return fmt.Errorf("failed to process image: %w", err)
		}
		if err := p.processPreviewUnsafe(); err != nil {
			return fmt.Errorf("failed to process preview: %w", err)
		}
	}

	return nil
}

func (p *ImagePipeline) AddTransformation(transformation Transformation) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.HasImageUnsafe() {
		p.debugPipeline.Log("Cannot add transformation: no image loaded")
		return fmt.Errorf("no image loaded")
	}

	p.transformations = append(p.transformations, transformation)

	if err := p.processImageUnsafe(); err != nil {
		return fmt.Errorf("failed to process image after adding transformation: %w", err)
	}
	if err := p.processPreviewUnsafe(); err != nil {
		return fmt.Errorf("failed to process preview after adding transformation: %w", err)
	}

	return nil
}

func (p *ImagePipeline) RemoveTransformation(index int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if index >= 0 && index < len(p.transformations) {
		p.transformations[index].Close()
		p.transformations = append(p.transformations[:index], p.transformations[index+1:]...)

		if p.HasImageUnsafe() {
			if err := p.processImageUnsafe(); err != nil {
				return fmt.Errorf("failed to process image after removing transformation: %w", err)
			}
			if err := p.processPreviewUnsafe(); err != nil {
				return fmt.Errorf("failed to process preview after removing transformation: %w", err)
			}
		}
	}
	return nil
}

func (p *ImagePipeline) ClearTransformations() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.debugPipeline.Log("Clearing all transformations")
	for _, transform := range p.transformations {
		transform.Close()
	}
	p.transformations = make([]Transformation, 0)

	if p.HasImageUnsafe() {
		p.processImageUnsafe()
		p.processPreviewUnsafe()
	}
}

func (p *ImagePipeline) GetProcessedImage() gocv.Mat {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() {
		p.debugPipeline.LogGetProcessedImage("not initialized, returning empty Mat")
		return gocv.NewMat()
	}
	if p.processedImage.Empty() {
		p.debugPipeline.LogGetProcessedImage("processed image empty, returning original")
		return p.originalImage.Clone()
	}
	return p.processedImage.Clone()
}

func (p *ImagePipeline) GetPreviewImage() gocv.Mat {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() {
		p.debugPipeline.LogGetProcessedImage("preview not initialized, returning empty Mat")
		return gocv.NewMat()
	}
	if p.previewImage.Empty() {
		p.debugPipeline.LogGetProcessedImage("preview image empty, returning original")
		return p.originalImage.Clone()
	}
	return p.previewImage.Clone()
}

func (p *ImagePipeline) ProcessPreview() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.processPreviewUnsafe()
}

// Thread-unsafe helper methods (must be called with mutex held)
func (p *ImagePipeline) HasImageUnsafe() bool {
	return p.initialized && !p.originalImage.Empty()
}

func (p *ImagePipeline) processImageUnsafe() error {
	p.debugPipeline.LogProcessStart()
	if !p.HasImageUnsafe() {
		p.debugPipeline.LogProcessEarlyReturn("not initialized")
		return fmt.Errorf("pipeline not initialized")
	}
	if p.originalImage.Empty() {
		p.debugPipeline.LogProcessEarlyReturn("original image is empty")
		return fmt.Errorf("original image is empty")
	}

	p.debugPipeline.StartTimer("processImage")
	defer func() {
		p.debugPipeline.EndTimer("processImage")
		if p.initialized && !p.processedImage.Empty() {
			p.debugPipeline.LogPipelineStats(p.originalImage.Size(), p.processedImage.Size(), len(p.transformations))
		}
		p.debugPipeline.LogMemoryUsage()
	}()

	p.debugPipeline.LogProcessStep("creating working copy")
	currentResult := p.originalImage.Clone()
	if currentResult.Empty() {
		return fmt.Errorf("failed to clone original for processing")
	}

	p.debugPipeline.LogTransformationCount(len(p.transformations))
	for i, transformation := range p.transformations {
		timerName := fmt.Sprintf("transformation_%d_%s", i, transformation.Name())
		p.debugPipeline.StartTimer(timerName)

		beforeTransform := currentResult.Clone()
		transformedResult := transformation.Apply(currentResult)
		duration := p.debugPipeline.EndTimer(timerName)

		p.debugPipeline.LogTransformationApplied(transformation.Name(), beforeTransform, transformedResult, duration)

		if !currentResult.Empty() {
			currentResult.Close()
		}
		if !beforeTransform.Empty() {
			beforeTransform.Close()
		}

		currentResult = transformedResult

		if currentResult.Empty() {
			return fmt.Errorf("transformation %s returned empty result", transformation.Name())
		}
	}

	// Replace processed image
	if !p.processedImage.Empty() {
		p.processedImage.Close()
	}
	p.processedImage = currentResult

	p.debugPipeline.LogProcessComplete()
	return nil
}

func (p *ImagePipeline) processPreviewUnsafe() error {
	p.debugPipeline.LogProcessStart()
	if !p.HasImageUnsafe() {
		p.debugPipeline.LogProcessEarlyReturn("preview not initialized")
		return fmt.Errorf("pipeline not initialized")
	}
	if p.originalImage.Empty() {
		p.debugPipeline.LogProcessEarlyReturn("original image is empty for preview")
		return fmt.Errorf("original image is empty")
	}

	p.debugPipeline.StartTimer("processPreview")
	defer func() {
		p.debugPipeline.EndTimer("processPreview")
		if p.initialized && !p.previewImage.Empty() {
			p.debugPipeline.LogPipelineStats(p.originalImage.Size(), p.previewImage.Size(), len(p.transformations))
		}
	}()

	currentResult := p.originalImage.Clone()
	if currentResult.Empty() {
		return fmt.Errorf("failed to clone original for preview")
	}

	for i, transformation := range p.transformations {
		timerName := fmt.Sprintf("preview_transformation_%d_%s", i, transformation.Name())
		p.debugPipeline.StartTimer(timerName)

		beforeTransform := currentResult.Clone()
		transformedResult := transformation.ApplyPreview(currentResult)
		duration := p.debugPipeline.EndTimer(timerName)

		p.debugPipeline.LogTransformationApplied(transformation.Name()+" (preview)", beforeTransform, transformedResult, duration)

		if !currentResult.Empty() {
			currentResult.Close()
		}
		if !beforeTransform.Empty() {
			beforeTransform.Close()
		}

		currentResult = transformedResult

		if currentResult.Empty() {
			return fmt.Errorf("preview transformation %s returned empty result", transformation.Name())
		}
	}

	// Replace preview image
	if !p.previewImage.Empty() {
		p.previewImage.Close()
	}
	p.previewImage = currentResult

	return nil
}

func (p *ImagePipeline) CalculatePSNR() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.Empty() || p.processedImage.Empty() {
		return 0.0
	}

	orig := p.originalImage.Clone()
	defer orig.Close()
	proc := p.processedImage.Clone()
	defer proc.Close()

	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			gocv.CvtColor(proc, &proc, gocv.ColorGrayToBGR)
		}
	}

	diff := gocv.NewMat()
	defer diff.Close()
	gocv.Subtract(orig, proc, &diff)

	diffSq := gocv.NewMat()
	defer diffSq.Close()
	gocv.Multiply(diff, diff, &diffSq)

	sumResult := diffSq.Sum()
	mse := sumResult.Val1 / float64(orig.Total())

	if mse == 0 {
		return math.Inf(1)
	}

	psnr := 20*math.Log10(255) - 10*math.Log10(mse)
	return psnr
}

func (p *ImagePipeline) CalculateSSIM() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.Empty() || p.processedImage.Empty() {
		return 0.0
	}

	orig := p.originalImage.Clone()
	defer orig.Close()
	proc := p.processedImage.Clone()
	defer proc.Close()

	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			gocv.CvtColor(proc, &proc, gocv.ColorGrayToBGR)
		} else if orig.Type() == gocv.MatTypeCV8U && proc.Channels() == 3 {
			gocv.CvtColor(orig, &orig, gocv.ColorGrayToBGR)
		}
	}

	origF := gocv.NewMat()
	defer origF.Close()
	procF := gocv.NewMat()
	defer procF.Close()

	orig.ConvertTo(&origF, gocv.MatTypeCV32F)
	proc.ConvertTo(&procF, gocv.MatTypeCV32F)

	K1 := 0.01
	K2 := 0.03
	L := 255.0
	C1 := (K1 * L) * (K1 * L)
	C2 := (K2 * L) * (K2 * L)

	origMean := origF.Mean()
	procMean := procF.Mean()

	μx := origMean.Val1
	μy := procMean.Val1

	origSquared := gocv.NewMat()
	defer origSquared.Close()
	procSquared := gocv.NewMat()
	defer procSquared.Close()

	gocv.Multiply(origF, origF, &origSquared)
	gocv.Multiply(procF, procF, &procSquared)

	origSquaredMean := origSquared.Mean()
	procSquaredMean := procSquared.Mean()

	σx2 := origSquaredMean.Val1 - μx*μx
	σy2 := procSquaredMean.Val1 - μy*μy

	crossProduct := gocv.NewMat()
	defer crossProduct.Close()

	gocv.Multiply(origF, procF, &crossProduct)
	crossMean := crossProduct.Mean()

	σxy := crossMean.Val1 - μx*μy

	numerator := (2*μx*μy + C1) * (2*σxy + C2)
	denominator := (μx*μx + μy*μy + C1) * (σx2 + σy2 + C2)

	var ssim float64
	if denominator == 0 {
		if μx == μy && σx2 == σy2 {
			ssim = 1.0
		} else {
			ssim = 0.0
		}
	} else {
		ssim = numerator / denominator
	}

	if ssim > 1.0 {
		ssim = 1.0
	} else if ssim < -1.0 {
		ssim = -1.0
	}

	return ssim
}

func (p *ImagePipeline) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.initialized {
		if !p.originalImage.Empty() {
			p.originalImage.Close()
		}
		if !p.processedImage.Empty() {
			p.processedImage.Close()
		}
		if !p.previewImage.Empty() {
			p.previewImage.Close()
		}
		for _, transform := range p.transformations {
			transform.Close()
		}
		p.debugMemory.LogMemorySummary()
	}
}
