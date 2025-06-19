package main

import (
	"fmt"
	"math"
	"sync"

	"gocv.io/x/gocv"
)

type ManagedMat struct {
	mat gocv.Mat
	id  uint64
}

func NewManagedMat(name string, debugMemory *DebugMemory) *ManagedMat {
	mat := gocv.NewMat()
	id := debugMemory.LogMatCreation(name)
	return &ManagedMat{
		mat: mat,
		id:  id,
	}
}

func (m *ManagedMat) Close(name string, debugMemory *DebugMemory) {
	// Check if mat is valid before calling Empty()
	if m != nil && m.id != 0 {
		debugMemory.LogMatCleanup(name, m.id)
		// Only close if mat is not already closed
		if !m.mat.Empty() {
			m.mat.Close()
		}
		m.id = 0 // Mark as closed
	}
}

func (m *ManagedMat) Mat() gocv.Mat {
	return m.mat
}

func (m *ManagedMat) IsEmpty() bool {
	if m == nil || m.id == 0 {
		return true
	}
	return m.mat.Empty()
}

type ImagePipeline struct {
	originalImage   *ManagedMat
	processedImage  *ManagedMat
	previewImage    *ManagedMat
	transformations []Transformation
	debugPipeline   *DebugPipeline
	debugMemory     *DebugMemory
	initialized     bool
	mutex           sync.RWMutex // Protect concurrent access
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

	// Initialize with empty managed Mats
	pipeline.originalImage = NewManagedMat("originalImage", debugMemory)
	pipeline.processedImage = NewManagedMat("processedImage", debugMemory)
	pipeline.previewImage = NewManagedMat("previewImage", debugMemory)

	return pipeline
}

func (p *ImagePipeline) HasImage() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.initialized && !p.originalImage.IsEmpty()
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

	// Set up new image - avoid defer overwrite pattern
	p.debugPipeline.LogSetOriginalStep("cloning original image")
	originalClone := img.Clone()
	if originalClone.Empty() {
		return fmt.Errorf("failed to clone original image")
	}

	// Close old mat and assign new one
	p.originalImage.Close("originalImage", p.debugMemory)
	p.originalImage = &ManagedMat{
		mat: originalClone,
		id:  p.debugMemory.LogMatCreation("originalImage"),
	}

	// Create processed and preview images
	p.debugPipeline.LogSetOriginalStep("creating processed image")
	processedClone := originalClone.Clone()
	if processedClone.Empty() {
		return fmt.Errorf("failed to clone processed image")
	}

	p.processedImage.Close("processedImage", p.debugMemory)
	p.processedImage = &ManagedMat{
		mat: processedClone,
		id:  p.debugMemory.LogMatCreation("processedImage"),
	}

	p.debugPipeline.LogSetOriginalStep("creating preview image")
	previewClone := originalClone.Clone()
	if previewClone.Empty() {
		return fmt.Errorf("failed to clone preview image")
	}

	p.previewImage.Close("previewImage", p.debugMemory)
	p.previewImage = &ManagedMat{
		mat: previewClone,
		id:  p.debugMemory.LogMatCreation("previewImage"),
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

	// Create new processed image - avoid defer overwrite pattern
	p.debugPipeline.LogProcessStep("creating new processed image")
	originalMat := p.originalImage.Mat()
	newProcessed := originalMat.Clone()
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

		// Clean up intermediate result
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

	// Replace processed image
	p.processedImage.Close("processedImage", p.debugMemory)
	p.processedImage = &ManagedMat{
		mat: newProcessed,
		id:  p.debugMemory.LogMatCreation("processedImage"),
	}

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

	// Create new preview image - avoid defer overwrite pattern
	originalMat := p.originalImage.Mat()
	newPreview := originalMat.Clone()
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

		// Clean up intermediate result
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

	// Replace preview image
	p.previewImage.Close("previewImage", p.debugMemory)
	p.previewImage = &ManagedMat{
		mat: newPreview,
		id:  p.debugMemory.LogMatCreation("previewImage"),
	}

	return nil
}

func (p *ImagePipeline) cleanupResourcesUnsafe() {
	if p.originalImage != nil && p.originalImage.id != 0 {
		p.debugPipeline.LogResourceCleanup("originalImage", true)
		p.originalImage.Close("originalImage", p.debugMemory)
	}

	if p.processedImage != nil && p.processedImage.id != 0 {
		p.debugPipeline.LogResourceCleanup("processedImage", true)
		p.processedImage.Close("processedImage", p.debugMemory)
	}

	if p.previewImage != nil && p.previewImage.id != 0 {
		p.debugPipeline.LogResourceCleanup("previewImage", true)
		p.previewImage.Close("previewImage", p.debugMemory)
	}
}

func (p *ImagePipeline) CalculatePSNR() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.IsEmpty() || p.processedImage.IsEmpty() {
		return 0.0
	}

	// Convert to same type if needed
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

// FIXED: Correct SSIM calculation implementation
func (p *ImagePipeline) CalculateSSIM() float64 {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	if !p.HasImageUnsafe() || p.originalImage.IsEmpty() || p.processedImage.IsEmpty() {
		return 0.0
	}

	originalMat := p.originalImage.Mat()
	processedMat := p.processedImage.Mat()

	// Clone matrices for processing
	orig := originalMat.Clone()
	defer orig.Close()
	proc := processedMat.Clone()
	defer proc.Close()

	// Ensure same type - convert processed to match original if needed
	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			gocv.CvtColor(proc, &proc, gocv.ColorGrayToBGR)
		} else if orig.Type() == gocv.MatTypeCV8U && proc.Channels() == 3 {
			gocv.CvtColor(orig, &orig, gocv.ColorGrayToBGR)
		}
	}

	// Convert to float32 for accurate calculations
	origF := gocv.NewMat()
	defer origF.Close()
	procF := gocv.NewMat()
	defer procF.Close()

	orig.ConvertTo(&origF, gocv.MatTypeCV32F)
	proc.ConvertTo(&procF, gocv.MatTypeCV32F)

	// SSIM constants (K1=0.01, K2=0.03, L=255 for 8-bit images)
	K1 := 0.01
	K2 := 0.03
	L := 255.0
	C1 := (K1 * L) * (K1 * L)
	C2 := (K2 * L) * (K2 * L)

	// Calculate means
	origMean := origF.Mean()
	procMean := procF.Mean()

	μx := origMean.Val1
	μy := procMean.Val1

	// FIXED: Calculate variances properly
	// σx² = E[X²] - (E[X])²
	origSquared := gocv.NewMat()
	defer origSquared.Close()
	procSquared := gocv.NewMat()
	defer procSquared.Close()

	gocv.Multiply(origF, origF, &origSquared)
	gocv.Multiply(procF, procF, &procSquared)

	origSquaredMean := origSquared.Mean()
	procSquaredMean := procSquared.Mean()

	σx2 := origSquaredMean.Val1 - μx*μx // Var(X) = E[X²] - (E[X])²
	σy2 := procSquaredMean.Val1 - μy*μy // Var(Y) = E[Y²] - (E[Y])²

	σx := math.Sqrt(σx2)
	σy := math.Sqrt(σy2)

	// FIXED: Calculate covariance properly
	// σxy = E[XY] - E[X]E[Y]
	crossProduct := gocv.NewMat()
	defer crossProduct.Close()

	gocv.Multiply(origF, procF, &crossProduct)
	crossMean := crossProduct.Mean()

	σxy := crossMean.Val1 - μx*μy // Cov(X,Y) = E[XY] - E[X]E[Y]

	// FIXED: Calculate SSIM using correct formula
	// SSIM = ((2μxμy + C1)(2σxy + C2)) / ((μx² + μy² + C1)(σx² + σy² + C2))
	numerator := (2*μx*μy + C1) * (2*σxy + C2)
	denominator := (μx*μx + μy*μy + C1) * (σx2 + σy2 + C2)

	var ssim float64
	if denominator == 0 {
		// Handle edge case where denominator is zero
		if μx == μy && σx == σy {
			ssim = 1.0 // Perfect similarity
		} else {
			ssim = 0.0
		}
	} else {
		ssim = numerator / denominator
	}

	// Clamp SSIM to valid range [-1, 1], though it should typically be [0, 1]
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
		p.cleanupResourcesUnsafe()
		for _, transform := range p.transformations {
			transform.Close()
		}
		p.debugMemory.LogMemorySummary()
	}
}
