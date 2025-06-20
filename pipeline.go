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
	initialized     bool
	mutex           sync.RWMutex // Protect concurrent access
}

func NewImagePipeline(config *DebugConfig) *ImagePipeline {
	debugPipeline := NewDebugPipeline(config)

	pipeline := &ImagePipeline{
		transformations: make([]Transformation, 0),
		debugPipeline:   debugPipeline,
		initialized:     false,
		// Initialize with empty Mats - GoCV handles memory tracking via MatProfile
		originalImage:  gocv.NewMat(),
		processedImage: gocv.NewMat(),
		previewImage:   gocv.NewMat(),
	}

	return pipeline
}

func (p *ImagePipeline) HasImage() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.initialized && !p.originalImage.Empty()
}

func (p *ImagePipeline) SetOriginalImage(img gocv.Mat) (err error) {
	// Panic recovery for OpenCV operations
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in SetOriginalImage: %v", r)
			p.debugPipeline.Log(fmt.Sprintf("PANIC RECOVERED: %v", r))
		}
	}()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.debugPipeline.LogSetOriginalStart()

	// Clean up existing resources
	if p.initialized {
		p.debugPipeline.LogSetOriginalStep("cleaning up existing resources")
		p.cleanupResourcesUnsafe()
	}

	// Validate input
	if img.Empty() {
		return fmt.Errorf("input image is empty")
	}

	// Set up new images using simple assignment - GoCV MatProfile tracks memory
	p.debugPipeline.LogSetOriginalStep("cloning original image")
	p.originalImage = img.Clone()
	if p.originalImage.Empty() {
		return fmt.Errorf("failed to clone original image")
	}

	p.debugPipeline.LogSetOriginalStep("creating processed image")
	p.processedImage = p.originalImage.Clone()
	if p.processedImage.Empty() {
		return fmt.Errorf("failed to clone processed image")
	}

	p.debugPipeline.LogSetOriginalStep("creating preview image")
	p.previewImage = p.originalImage.Clone()
	if p.previewImage.Empty() {
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
		return p.originalImage
	}
	return p.processedImage
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
		return p.originalImage
	}
	return p.previewImage
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

func (p *ImagePipeline) processImageUnsafe() (err error) {
	// Panic recovery for OpenCV operations
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in processImage: %v", r)
			p.debugPipeline.Log(fmt.Sprintf("PANIC RECOVERED: %v", r))
		}
	}()

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

	// Create new processed image using GoCV's memory management
	p.debugPipeline.LogProcessStep("creating new processed image")
	newProcessed := p.originalImage.Clone()
	if newProcessed.Empty() {
		return fmt.Errorf("failed to clone original for processing")
	}

	// Apply all transformations sequentially
	p.debugPipeline.LogTransformationCount(len(p.transformations))
	for i, transformation := range p.transformations {
		timerName := fmt.Sprintf("transformation_%d_%s", i, transformation.Name())
		p.debugPipeline.StartTimer(timerName)

		before := newProcessed.Clone()
		result := transformation.Apply(newProcessed)
		duration := p.debugPipeline.EndTimer(timerName)

		p.debugPipeline.LogTransformationApplied(transformation.Name(), before, result, duration)

		// Clean up intermediate results - GoCV handles memory tracking
		if !newProcessed.Empty() {
			newProcessed.Close()
		}
		newProcessed = result

		if !before.Empty() {
			before.Close()
		}

		// Validate result
		if newProcessed.Empty() {
			return fmt.Errorf("transformation %s returned empty result", transformation.Name())
		}
	}

	// Replace processed image - clean up old one first
	if !p.processedImage.Empty() {
		p.processedImage.Close()
	}
	p.processedImage = newProcessed

	p.debugPipeline.LogProcessComplete()
	return nil
}

func (p *ImagePipeline) processPreviewUnsafe() (err error) {
	// Panic recovery for OpenCV operations
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in processPreview: %v", r)
			p.debugPipeline.Log(fmt.Sprintf("PANIC RECOVERED: %v", r))
		}
	}()

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

	// Create new preview image using GoCV's memory management
	newPreview := p.originalImage.Clone()
	if newPreview.Empty() {
		return fmt.Errorf("failed to clone original for preview")
	}

	// Apply all transformations sequentially using preview method
	for i, transformation := range p.transformations {
		timerName := fmt.Sprintf("preview_transformation_%d_%s", i, transformation.Name())
		p.debugPipeline.StartTimer(timerName)

		before := newPreview.Clone()
		result := transformation.ApplyPreview(newPreview)
		duration := p.debugPipeline.EndTimer(timerName)

		p.debugPipeline.LogTransformationApplied(transformation.Name()+" (preview)", before, result, duration)

		// Clean up intermediate results
		if !newPreview.Empty() {
			newPreview.Close()
		}
		newPreview = result

		if !before.Empty() {
			before.Close()
		}

		// Validate result
		if newPreview.Empty() {
			return fmt.Errorf("preview transformation %s returned empty result", transformation.Name())
		}
	}

	// Replace preview image - clean up old one first
	if !p.previewImage.Empty() {
		p.previewImage.Close()
	}
	p.previewImage = newPreview

	return nil
}

func (p *ImagePipeline) cleanupResourcesUnsafe() {
	if !p.originalImage.Empty() {
		p.debugPipeline.LogResourceCleanup("originalImage", true)
		p.originalImage.Close()
	}

	if !p.processedImage.Empty() {
		p.debugPipeline.LogResourceCleanup("processedImage", true)
		p.processedImage.Close()
	}

	if !p.previewImage.Empty() {
		p.debugPipeline.LogResourceCleanup("previewImage", true)
		p.previewImage.Close()
	}
}

func (p *ImagePipeline) CalculatePSNR() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.Empty() || p.processedImage.Empty() {
		return 0.0
	}

	// Convert to same type if needed
	orig := p.originalImage.Clone()
	defer orig.Close()
	proc := p.processedImage.Clone()
	defer proc.Close()

	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			gocv.CvtColor(proc, &proc, gocv.ColorGrayToBGR)
		}
	}

	// Calculate MSE
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

	// Simple SSIM approximation using correlation coefficient
	orig := p.originalImage.Clone()
	defer orig.Close()
	proc := p.processedImage.Clone()
	defer proc.Close()

	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			gocv.CvtColor(proc, &proc, gocv.ColorGrayToBGR)
		}
	}

	// Convert to float for calculations
	origF := gocv.NewMat()
	defer origF.Close()
	procF := gocv.NewMat()
	defer procF.Close()
	orig.ConvertTo(&origF, gocv.MatTypeCV32F)
	proc.ConvertTo(&procF, gocv.MatTypeCV32F)

	// Calculate means
	origMean := origF.Mean()
	procMean := procF.Mean()

	// Calculate standard deviations and covariance
	origVar := gocv.NewMat()
	defer origVar.Close()
	procVar := gocv.NewMat()
	defer procVar.Close()
	covar := gocv.NewMat()
	defer covar.Close()

	origSub := gocv.NewMat()
	defer origSub.Close()
	procSub := gocv.NewMat()
	defer procSub.Close()

	// Create scalar mats for subtraction
	origMeanMat := gocv.NewMatFromScalar(origMean, origF.Type())
	defer origMeanMat.Close()
	procMeanMat := gocv.NewMatFromScalar(procMean, procF.Type())
	defer procMeanMat.Close()

	gocv.Subtract(origF, origMeanMat, &origSub)
	gocv.Subtract(procF, procMeanMat, &procSub)

	gocv.Multiply(origSub, origSub, &origVar)
	gocv.Multiply(procSub, procSub, &procVar)
	gocv.Multiply(origSub, procSub, &covar)

	// Calculate variance and covariance
	origVarSum := origVar.Sum()
	procVarSum := procVar.Sum()
	covarSum := covar.Sum()

	origStd := math.Sqrt(origVarSum.Val1 / float64(orig.Total()))
	procStd := math.Sqrt(procVarSum.Val1 / float64(proc.Total()))
	covariance := covarSum.Val1 / float64(orig.Total())

	// SSIM constants
	c1 := math.Pow(0.01*255, 2)
	c2 := math.Pow(0.03*255, 2)

	// Calculate SSIM
	numerator := (2*origMean.Val1*procMean.Val1 + c1) * (2*covariance + c2)
	denominator := (math.Pow(origMean.Val1, 2) + math.Pow(procMean.Val1, 2) + c1) * (math.Pow(origStd, 2) + math.Pow(procStd, 2) + c2)

	ssim := numerator / denominator

	if ssim > 1.0 {
		ssim = 1.0
	}
	if ssim < 0.0 {
		ssim = 0.0
	}

	return ssim
}

func (p *ImagePipeline) Close() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.initialized {
		p.cleanupResourcesUnsafe()
		for _, transform := range p.transformations {
			transform.Close()
		}
	}
}
