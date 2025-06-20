package main

import (
	"fmt"
	"math"
	"sync"

	"gocv.io/x/gocv"
)

type ImagePipeline struct {
	// Use proper atomic access patterns for thread safety
	originalImage   gocv.Mat
	processedImage  gocv.Mat
	previewImage    gocv.Mat
	transformations []Transformation
	debugPipeline   *DebugPipeline
	initialized     bool
	mutex           sync.RWMutex // Protect concurrent access

	// Remove excessive rate limiting for preview processing
	processingMutex sync.Mutex // Separate mutex for processing operations
}

func NewImagePipeline(config *DebugConfig) *ImagePipeline {
	debugPipeline := NewDebugPipeline(config)

	pipeline := &ImagePipeline{
		transformations: make([]Transformation, 0),
		debugPipeline:   debugPipeline,
		initialized:     false,
		// Initialize with empty Mats using proper GoCV patterns
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
	// Panic recovery with proper error handling
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in SetOriginalImage: %v", r)
			p.debugPipeline.Log(fmt.Sprintf("PANIC RECOVERED: %v", r))
		}
	}()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.debugPipeline.LogSetOriginalStart()

	// Input validation
	if img.Empty() {
		return fmt.Errorf("input image is empty")
	}

	// Validate image dimensions
	if img.Cols() <= 0 || img.Rows() <= 0 || img.Cols() > 65536 || img.Rows() > 65536 {
		return fmt.Errorf("invalid image dimensions: %dx%d", img.Cols(), img.Rows())
	}

	// Clean up existing resources properly
	if p.initialized {
		p.debugPipeline.LogSetOriginalStep("cleaning up existing resources")
		p.cleanupResourcesUnsafe()
	}

	// Atomic resource setup with proper error handling
	p.debugPipeline.LogSetOriginalStep("cloning original image")
	p.originalImage = img.Clone()
	if p.originalImage.Empty() {
		return fmt.Errorf("failed to clone original image")
	}

	p.debugPipeline.LogSetOriginalStep("creating processed image")
	p.processedImage = p.originalImage.Clone()
	if p.processedImage.Empty() {
		p.originalImage.Close()
		return fmt.Errorf("failed to clone processed image")
	}

	p.debugPipeline.LogSetOriginalStep("creating preview image")
	p.previewImage = p.originalImage.Clone()
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

	// Validate transformation
	if transformation == nil {
		return fmt.Errorf("transformation is nil")
	}

	p.transformations = append(p.transformations, transformation)

	if err := p.processImageUnsafe(); err != nil {
		// Remove the failed transformation
		p.transformations = p.transformations[:len(p.transformations)-1]
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
		// Proper cleanup of transformation resources
		if p.transformations[index] != nil {
			p.transformations[index].Close()
		}
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
		if transform != nil {
			transform.Close()
		}
	}
	p.transformations = make([]Transformation, 0)

	if p.HasImageUnsafe() {
		p.processImageUnsafe()
		p.processPreviewUnsafe()
	}
}

// Add public ProcessImage method to force full reprocessing
func (p *ImagePipeline) ProcessImage() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.debugPipeline.Log("ProcessImage: Force reprocessing full resolution image")
	return p.processImageUnsafe()
}

// Add method to force preview regeneration with current parameters
func (p *ImagePipeline) ForcePreviewRegeneration() error {
	p.debugPipeline.Log("ForcePreviewRegeneration: Regenerating preview with current parameters")

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.HasImageUnsafe() {
		return fmt.Errorf("no image loaded")
	}

	// Always regenerate preview from scratch with current parameters
	return p.processPreviewUnsafe()
}

// Add public method to reprocess preview with current parameters
func (p *ImagePipeline) ReprocessPreview() error {
	p.debugPipeline.Log("ReprocessPreview: Force reprocessing preview with current parameters")

	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.processPreviewUnsafe()
}

// Return thread-safe clones with proper error handling
func (p *ImagePipeline) GetProcessedImage() gocv.Mat {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() {
		p.debugPipeline.LogGetProcessedImage("not initialized, returning empty Mat")
		return gocv.NewMat()
	}
	if p.processedImage.Empty() {
		p.debugPipeline.LogGetProcessedImage("processed image empty, returning original clone")
		return p.originalImage.Clone()
	}
	// Always return clones to prevent race conditions
	return p.processedImage.Clone()
}

// Return thread-safe clones with proper error handling
func (p *ImagePipeline) GetPreviewImage() gocv.Mat {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() {
		p.debugPipeline.LogGetProcessedImage("preview not initialized, returning empty Mat")
		return gocv.NewMat()
	}
	if p.previewImage.Empty() {
		p.debugPipeline.LogGetProcessedImage("preview image empty, returning original clone")
		return p.originalImage.Clone()
	}
	// Always return clones to prevent race conditions
	return p.previewImage.Clone()
}

// Remove debouncing and use separate mutex for processing
func (p *ImagePipeline) ProcessPreview() error {
	// Use separate mutex for processing to prevent deadlocks
	p.processingMutex.Lock()
	defer p.processingMutex.Unlock()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.processPreviewUnsafe()
}

// Thread-unsafe helper methods (must be called with mutex held)
func (p *ImagePipeline) HasImageUnsafe() bool {
	return p.initialized && !p.originalImage.Empty()
}

func (p *ImagePipeline) processImageUnsafe() (err error) {
	// Panic recovery with proper cleanup
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

	// Create new processed image with validation
	p.debugPipeline.LogProcessStep("creating new processed image")
	newProcessed := p.originalImage.Clone()
	if newProcessed.Empty() {
		return fmt.Errorf("failed to clone original for processing")
	}

	// Apply all transformations sequentially with proper error handling
	p.debugPipeline.LogTransformationCount(len(p.transformations))
	for i, transformation := range p.transformations {
		if transformation == nil {
			newProcessed.Close()
			return fmt.Errorf("transformation %d is nil", i)
		}

		timerName := fmt.Sprintf("transformation_%d_%s", i, transformation.Name())
		p.debugPipeline.StartTimer(timerName)

		before := newProcessed.Clone()
		result := transformation.Apply(newProcessed)
		duration := p.debugPipeline.EndTimer(timerName)

		p.debugPipeline.LogTransformationApplied(transformation.Name(), before, result, duration)

		// Proper cleanup and validation
		if !newProcessed.Empty() {
			newProcessed.Close()
		}
		if !before.Empty() {
			before.Close()
		}

		// Validate transformation result
		if result.Empty() {
			return fmt.Errorf("transformation %s returned empty result", transformation.Name())
		}

		newProcessed = result
	}

	// Atomic replacement of processed image
	if !p.processedImage.Empty() {
		p.processedImage.Close()
	}
	p.processedImage = newProcessed

	p.debugPipeline.LogProcessComplete()
	return nil
}

func (p *ImagePipeline) processPreviewUnsafe() (err error) {
	// Panic recovery
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

	// Always create new preview image from original to ensure current parameters are used
	newPreview := p.originalImage.Clone()
	if newPreview.Empty() {
		return fmt.Errorf("failed to clone original for preview")
	}

	// Apply all transformations using their current parameters via ApplyPreview
	for i, transformation := range p.transformations {
		if transformation == nil {
			newPreview.Close()
			return fmt.Errorf("preview transformation %d is nil", i)
		}

		timerName := fmt.Sprintf("preview_transformation_%d_%s", i, transformation.Name())
		p.debugPipeline.StartTimer(timerName)

		before := newPreview.Clone()
		// Use ApplyPreview which reads current parameters
		result := transformation.ApplyPreview(newPreview)
		duration := p.debugPipeline.EndTimer(timerName)

		p.debugPipeline.LogTransformationApplied(transformation.Name()+" (preview)", before, result, duration)

		// Proper cleanup
		if !newPreview.Empty() {
			newPreview.Close()
		}
		if !before.Empty() {
			before.Close()
		}

		// Validate preview result
		if result.Empty() {
			return fmt.Errorf("preview transformation %s returned empty result", transformation.Name())
		}

		newPreview = result
	}

	// Atomic replacement of preview image
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

// Proper PSNR calculation with numerical stability
func (p *ImagePipeline) CalculatePSNR() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.Empty() || p.processedImage.Empty() {
		return 0.0
	}

	// Create thread-safe clones for calculation
	orig := p.originalImage.Clone()
	defer orig.Close()
	proc := p.processedImage.Clone()
	defer proc.Close()

	// Validate dimensions match
	if orig.Rows() != proc.Rows() || orig.Cols() != proc.Cols() {
		return 0.0
	}

	// Handle channel mismatch properly
	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			temp := gocv.NewMat()
			defer temp.Close()
			gocv.CvtColor(proc, &temp, gocv.ColorGrayToBGR)
			proc.Close()
			proc = temp.Clone()
		} else if orig.Type() == gocv.MatTypeCV8U && proc.Channels() == 3 {
			temp := gocv.NewMat()
			defer temp.Close()
			gocv.CvtColor(orig, &temp, gocv.ColorGrayToBGR)
			orig.Close()
			orig = temp.Clone()
		}
	}

	// Calculate MSE with proper numerical handling
	diff := gocv.NewMat()
	defer diff.Close()
	gocv.Subtract(orig, proc, &diff)

	diffSq := gocv.NewMat()
	defer diffSq.Close()
	gocv.Multiply(diff, diff, &diffSq)

	sumResult := diffSq.Sum()
	totalPixels := float64(orig.Total())

	// Handle potential division by zero
	if totalPixels == 0 {
		return 0.0
	}

	mse := sumResult.Val1 / totalPixels

	// Edge case handling
	if mse == 0 {
		return 100.0 // Perfect match - return large finite value instead of infinity
	}
	if mse < 1e-10 {
		return 100.0 // Very small differences
	}

	// Prevent NaN and infinite results
	if math.IsInf(mse, 0) || math.IsNaN(mse) {
		return 0.0
	}

	psnr := 20*math.Log10(255) - 10*math.Log10(mse)

	// Clamp to reasonable range and handle edge cases
	if math.IsInf(psnr, 0) || math.IsNaN(psnr) {
		return 100.0
	}
	if psnr > 100 {
		return 100.0
	}
	if psnr < 0 {
		return 0.0
	}

	return psnr
}

// Proper SSIM calculation with numerical stability
func (p *ImagePipeline) CalculateSSIM() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.Empty() || p.processedImage.Empty() {
		return 0.0
	}

	// Create thread-safe clones for calculation
	orig := p.originalImage.Clone()
	defer orig.Close()
	proc := p.processedImage.Clone()
	defer proc.Close()

	// Validate dimensions match
	if orig.Rows() != proc.Rows() || orig.Cols() != proc.Cols() {
		return 0.0
	}

	// Handle channel mismatch properly
	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			temp := gocv.NewMat()
			defer temp.Close()
			gocv.CvtColor(proc, &temp, gocv.ColorGrayToBGR)
			proc.Close()
			proc = temp.Clone()
		} else if orig.Type() == gocv.MatTypeCV8U && proc.Channels() == 3 {
			temp := gocv.NewMat()
			defer temp.Close()
			gocv.CvtColor(orig, &temp, gocv.ColorGrayToBGR)
			orig.Close()
			orig = temp.Clone()
		}
	}

	// Convert to grayscale for SSIM calculation
	if orig.Channels() > 1 {
		origGray := gocv.NewMat()
		defer origGray.Close()
		gocv.CvtColor(orig, &origGray, gocv.ColorBGRToGray)
		orig.Close()
		orig = origGray.Clone()
	}

	if proc.Channels() > 1 {
		procGray := gocv.NewMat()
		defer procGray.Close()
		gocv.CvtColor(proc, &procGray, gocv.ColorBGRToGray)
		proc.Close()
		proc = procGray.Clone()
	}

	// Convert to float for calculations
	origF := gocv.NewMat()
	defer origF.Close()
	procF := gocv.NewMat()
	defer procF.Close()
	orig.ConvertTo(&origF, gocv.MatTypeCV32F)
	proc.ConvertTo(&procF, gocv.MatTypeCV32F)

	// SSIM constants for numerical stability
	c1 := math.Pow(0.01*255, 2)
	c2 := math.Pow(0.03*255, 2)

	// Calculate means with proper error handling
	origMean := origF.Mean()
	procMean := procF.Mean()

	// Validate mean calculations
	if math.IsInf(origMean.Val1, 0) || math.IsNaN(origMean.Val1) ||
		math.IsInf(procMean.Val1, 0) || math.IsNaN(procMean.Val1) {
		return 0.0
	}

	mu1 := origMean.Val1
	mu2 := procMean.Val1

	// Calculate variances and covariance using proper formulas
	origSub := gocv.NewMat()
	defer origSub.Close()
	procSub := gocv.NewMat()
	defer procSub.Close()

	origMeanMat := gocv.NewMatFromScalar(origMean, origF.Type())
	defer origMeanMat.Close()
	procMeanMat := gocv.NewMatFromScalar(procMean, procF.Type())
	defer procMeanMat.Close()

	gocv.Subtract(origF, origMeanMat, &origSub)
	gocv.Subtract(procF, procMeanMat, &procSub)

	// Calculate sigma1^2, sigma2^2, and sigma12
	sigma1Sq := gocv.NewMat()
	defer sigma1Sq.Close()
	sigma2Sq := gocv.NewMat()
	defer sigma2Sq.Close()
	sigma12 := gocv.NewMat()
	defer sigma12.Close()

	gocv.Multiply(origSub, origSub, &sigma1Sq)
	gocv.Multiply(procSub, procSub, &sigma2Sq)
	gocv.Multiply(origSub, procSub, &sigma12)

	sigma1SqVal := sigma1Sq.Mean().Val1
	sigma2SqVal := sigma2Sq.Mean().Val1
	sigma12Val := sigma12.Mean().Val1

	// Validate variance calculations
	if math.IsInf(sigma1SqVal, 0) || math.IsNaN(sigma1SqVal) ||
		math.IsInf(sigma2SqVal, 0) || math.IsNaN(sigma2SqVal) ||
		math.IsInf(sigma12Val, 0) || math.IsNaN(sigma12Val) {
		return 0.0
	}

	// Calculate SSIM using proper formula with numerical stability
	numerator := (2*mu1*mu2 + c1) * (2*sigma12Val + c2)
	denominator := (mu1*mu1 + mu2*mu2 + c1) * (sigma1SqVal + sigma2SqVal + c2)

	// Prevent division by zero
	if denominator == 0 || math.IsInf(denominator, 0) || math.IsNaN(denominator) {
		return 0.0
	}

	ssim := numerator / denominator

	// Bounds checking and numerical stability
	if math.IsInf(ssim, 0) || math.IsNaN(ssim) {
		return 0.0
	}

	// Clamp to valid range
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
			if transform != nil {
				transform.Close()
			}
		}
		p.transformations = nil
		p.initialized = false
	}
}
