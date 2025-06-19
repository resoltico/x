package main

import (
	"fmt"
	"math"
	"sync"

	"gocv.io/x/gocv"
)

// SafeMat provides safe OpenCV Mat management
type SafeMat struct {
	mat    gocv.Mat
	closed bool
	mutex  sync.Mutex
}

func NewSafeMat() *SafeMat {
	return &SafeMat{
		mat:    gocv.NewMat(),
		closed: false,
	}
}

func (s *SafeMat) Close() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if !s.closed && !s.mat.Empty() {
		s.mat.Close()
	}
	s.closed = true
}

func (s *SafeMat) Mat() gocv.Mat {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return gocv.NewMat() // Return empty Mat if closed
	}
	return s.mat
}

func (s *SafeMat) IsEmpty() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.closed || s.mat.Empty()
}

func (s *SafeMat) IsClosed() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.closed
}

// FIXED: Safe replacement that recreates SafeMat instead of replacing content
func (s *SafeMat) Replace(newMat gocv.Mat) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Close old mat if it exists and not already closed
	if !s.closed && !s.mat.Empty() {
		s.mat.Close()
	}

	s.mat = newMat
	s.closed = false // Reset closed state
	return nil
}

type ImagePipeline struct {
	originalImage   *SafeMat
	processedImage  *SafeMat
	previewImage    *SafeMat
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
	}

	// Initialize with safe Mat wrappers
	pipeline.originalImage = NewSafeMat()
	pipeline.processedImage = NewSafeMat()
	pipeline.previewImage = NewSafeMat()

	return pipeline
}

func (p *ImagePipeline) HasImage() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.initialized && !p.originalImage.IsEmpty()
}

// FIXED: Completely recreate SafeMats instead of trying to reuse closed ones
func (p *ImagePipeline) SetOriginalImage(img gocv.Mat) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.debugPipeline.LogSetOriginalStart()

	// Validate input
	if img.Empty() {
		return fmt.Errorf("input image is empty")
	}

	// FIXED: Always create new SafeMats for clean state
	p.debugPipeline.LogSetOriginalStep("creating new SafeMats")

	// Close existing SafeMats
	if p.initialized {
		p.debugPipeline.LogSetOriginalStep("closing existing SafeMats")
		p.originalImage.Close()
		p.processedImage.Close()
		p.previewImage.Close()
	}

	// Create completely new SafeMats
	p.originalImage = NewSafeMat()
	p.processedImage = NewSafeMat()
	p.previewImage = NewSafeMat()

	// Create clones
	p.debugPipeline.LogSetOriginalStep("cloning images safely")

	originalClone := img.Clone()
	if originalClone.Empty() {
		return fmt.Errorf("failed to clone original image")
	}

	processedClone := img.Clone()
	if processedClone.Empty() {
		originalClone.Close()
		return fmt.Errorf("failed to clone processed image")
	}

	previewClone := img.Clone()
	if previewClone.Empty() {
		originalClone.Close()
		processedClone.Close()
		return fmt.Errorf("failed to clone preview image")
	}

	// Replace content in the new SafeMats
	if err := p.originalImage.Replace(originalClone); err != nil {
		originalClone.Close()
		processedClone.Close()
		previewClone.Close()
		return fmt.Errorf("failed to set original image: %w", err)
	}

	if err := p.processedImage.Replace(processedClone); err != nil {
		processedClone.Close()
		previewClone.Close()
		return fmt.Errorf("failed to set processed image: %w", err)
	}

	if err := p.previewImage.Replace(previewClone); err != nil {
		previewClone.Close()
		return fmt.Errorf("failed to set preview image: %w", err)
	}

	p.initialized = true
	p.debugPipeline.LogImageStats("original", p.originalImage.Mat())

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
	if p.processedImage.IsEmpty() {
		p.debugPipeline.LogGetProcessedImage("processed image empty, returning original")
		return p.originalImage.Mat()
	}
	return p.processedImage.Mat()
}

func (p *ImagePipeline) GetPreviewImage() gocv.Mat {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() {
		p.debugPipeline.LogGetProcessedImage("preview not initialized, returning empty Mat")
		return gocv.NewMat()
	}
	if p.previewImage.IsEmpty() {
		p.debugPipeline.LogGetProcessedImage("preview image empty, returning original")
		return p.originalImage.Mat()
	}
	return p.previewImage.Mat()
}

func (p *ImagePipeline) ProcessPreview() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.processPreviewUnsafe()
}

// Thread-unsafe helper methods (must be called with mutex held)
func (p *ImagePipeline) HasImageUnsafe() bool {
	return p.initialized && !p.originalImage.IsEmpty()
}

func (p *ImagePipeline) processImageUnsafe() error {
	p.debugPipeline.LogProcessStart()
	if !p.HasImageUnsafe() {
		p.debugPipeline.LogProcessEarlyReturn("not initialized")
		return fmt.Errorf("pipeline not initialized")
	}
	if p.originalImage.IsEmpty() {
		p.debugPipeline.LogProcessEarlyReturn("original image is empty")
		return fmt.Errorf("original image is empty")
	}

	p.debugPipeline.StartTimer("processImage")
	defer func() {
		p.debugPipeline.EndTimer("processImage")
		if p.initialized && !p.processedImage.IsEmpty() {
			originalMat := p.originalImage.Mat()
			processedMat := p.processedImage.Mat()
			p.debugPipeline.LogPipelineStats(originalMat.Size(), processedMat.Size(), len(p.transformations))
		}
		p.debugPipeline.LogMemoryUsage()
	}()

	p.debugPipeline.LogProcessStep("creating working copy")
	originalMat := p.originalImage.Mat()
	currentResult := originalMat.Clone()
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

	if err := p.processedImage.Replace(currentResult); err != nil {
		currentResult.Close()
		return fmt.Errorf("failed to update processed image: %w", err)
	}

	p.debugPipeline.LogProcessComplete()
	return nil
}

func (p *ImagePipeline) processPreviewUnsafe() error {
	p.debugPipeline.LogProcessStart()
	if !p.HasImageUnsafe() {
		p.debugPipeline.LogProcessEarlyReturn("preview not initialized")
		return fmt.Errorf("pipeline not initialized")
	}
	if p.originalImage.IsEmpty() {
		p.debugPipeline.LogProcessEarlyReturn("original image is empty for preview")
		return fmt.Errorf("original image is empty")
	}

	p.debugPipeline.StartTimer("processPreview")
	defer func() {
		p.debugPipeline.EndTimer("processPreview")
		if p.initialized && !p.previewImage.IsEmpty() {
			originalMat := p.originalImage.Mat()
			previewMat := p.previewImage.Mat()
			p.debugPipeline.LogPipelineStats(originalMat.Size(), previewMat.Size(), len(p.transformations))
		}
	}()

	originalMat := p.originalImage.Mat()
	currentResult := originalMat.Clone()
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

	if err := p.previewImage.Replace(currentResult); err != nil {
		currentResult.Close()
		return fmt.Errorf("failed to update preview image: %w", err)
	}

	return nil
}

func (p *ImagePipeline) CalculatePSNR() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.IsEmpty() || p.processedImage.IsEmpty() {
		return 0.0
	}

	originalMat := p.originalImage.Mat()
	processedMat := p.processedImage.Mat()

	orig := originalMat.Clone()
	defer orig.Close()
	proc := processedMat.Clone()
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

	if !p.HasImageUnsafe() || p.originalImage.IsEmpty() || p.processedImage.IsEmpty() {
		return 0.0
	}

	originalMat := p.originalImage.Mat()
	processedMat := p.processedImage.Mat()

	orig := originalMat.Clone()
	defer orig.Close()
	proc := processedMat.Clone()
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
		p.originalImage.Close()
		p.processedImage.Close()
		p.previewImage.Close()
		for _, transform := range p.transformations {
			transform.Close()
		}
		p.debugMemory.LogMemorySummary()
	}
}
