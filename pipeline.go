package main

import (
	"fmt"
	"image"
	"math"
	"sync"
	"sync/atomic"

	"gocv.io/x/gocv"
)

type ImagePipeline struct {
	// Atomic access patterns for thread safety
	originalImage   gocv.Mat
	processedImage  gocv.Mat
	previewImage    gocv.Mat
	transformations []Transformation
	debugPipeline   *DebugPipeline
	initialized     int32 // atomic flag
	mutex           sync.RWMutex

	processingMutex sync.Mutex
}

func NewImagePipeline(config *DebugConfig) *ImagePipeline {
	debugPipeline := NewDebugPipeline(config)

	pipeline := &ImagePipeline{
		transformations: make([]Transformation, 0),
		debugPipeline:   debugPipeline,
		originalImage:   gocv.NewMat(),
		processedImage:  gocv.NewMat(),
		previewImage:    gocv.NewMat(),
	}

	return pipeline
}

func (p *ImagePipeline) HasImage() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return atomic.LoadInt32(&p.initialized) == 1 && !p.originalImage.Empty()
}

func (p *ImagePipeline) SetOriginalImage(img gocv.Mat) (err error) {
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
	if atomic.LoadInt32(&p.initialized) == 1 {
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

	atomic.StoreInt32(&p.initialized, 1)
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

func (p *ImagePipeline) ProcessImage() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.debugPipeline.Log("ProcessImage: Force reprocessing full resolution image")
	return p.processImageUnsafe()
}

func (p *ImagePipeline) ForcePreviewRegeneration() error {
	p.debugPipeline.Log("ForcePreviewRegeneration: Regenerating preview with current parameters")

	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.HasImageUnsafe() {
		return fmt.Errorf("no image loaded")
	}

	return p.processPreviewUnsafe()
}

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
	return atomic.LoadInt32(&p.initialized) == 1 && !p.originalImage.Empty()
}

func (p *ImagePipeline) processImageUnsafe() (err error) {
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
		if atomic.LoadInt32(&p.initialized) == 1 && !p.processedImage.Empty() {
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

	// Ensure proper cleanup even on early returns
	defer func() {
		// Only close newProcessed if we haven't assigned it to p.processedImage
		if err != nil && !newProcessed.Empty() {
			newProcessed.Close()
		}
	}()

	// Apply all transformations sequentially with proper error handling
	p.debugPipeline.LogTransformationCount(len(p.transformations))
	for i, transformation := range p.transformations {
		if transformation == nil {
			return fmt.Errorf("transformation %d is nil", i)
		}

		timerName := fmt.Sprintf("transformation_%d_%s", i, transformation.Name())
		p.debugPipeline.StartTimer(timerName)

		before := newProcessed.Clone()
		result := transformation.Apply(newProcessed)
		duration := p.debugPipeline.EndTimer(timerName)

		p.debugPipeline.LogTransformationApplied(transformation.Name(), before, result, duration)

		// Proper cleanup and validation
		if !before.Empty() {
			before.Close()
		}

		// Validate transformation result
		if result.Empty() {
			return fmt.Errorf("transformation %s returned empty result", transformation.Name())
		}

		// Close the old newProcessed before replacing it
		if !newProcessed.Empty() {
			newProcessed.Close()
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
		if atomic.LoadInt32(&p.initialized) == 1 && !p.previewImage.Empty() {
			p.debugPipeline.LogPipelineStats(p.originalImage.Size(), p.previewImage.Size(), len(p.transformations))
		}
	}()

	// Always create new preview image from original to ensure current parameters are used
	newPreview := p.originalImage.Clone()
	if newPreview.Empty() {
		return fmt.Errorf("failed to clone original for preview")
	}

	// Ensure proper cleanup even on early returns
	defer func() {
		// Only close newPreview if we haven't assigned it to p.previewImage
		if err != nil && !newPreview.Empty() {
			newPreview.Close()
		}
	}()

	// Apply all transformations using their current parameters via ApplyPreview
	for i, transformation := range p.transformations {
		if transformation == nil {
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
		if !before.Empty() {
			before.Close()
		}

		// Validate preview result
		if result.Empty() {
			return fmt.Errorf("preview transformation %s returned empty result", transformation.Name())
		}

		// Close the old newPreview before replacing it
		if !newPreview.Empty() {
			newPreview.Close()
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
		p.originalImage = gocv.NewMat() // Initialize new empty Mat
	}

	if !p.processedImage.Empty() {
		p.debugPipeline.LogResourceCleanup("processedImage", true)
		p.processedImage.Close()
		p.processedImage = gocv.NewMat() // Initialize new empty Mat
	}

	if !p.previewImage.Empty() {
		p.debugPipeline.LogResourceCleanup("previewImage", true)
		p.previewImage.Close()
		p.previewImage = gocv.NewMat() // Initialize new empty Mat
	}
}

// PSNR calculation with mathematical correctness and numerical stability
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

	// Handle channel mismatch properly - convert to same format
	if orig.Type() != proc.Type() {
		if proc.Type() == gocv.MatTypeCV8U && orig.Channels() == 3 {
			// Convert grayscale processed to BGR
			temp := gocv.NewMat()
			defer temp.Close()
			err := gocv.CvtColor(proc, &temp, gocv.ColorGrayToBGR)
			if err != nil {
				return 0.0
			}
			proc.Close()
			proc = temp.Clone()
		} else if orig.Type() == gocv.MatTypeCV8U && proc.Channels() == 3 {
			// Convert BGR original to grayscale
			temp := gocv.NewMat()
			defer temp.Close()
			err := gocv.CvtColor(orig, &temp, gocv.ColorBGRToGray)
			if err != nil {
				return 0.0
			}
			orig.Close()
			orig = temp.Clone()
		}
	}

	// Convert to float64 for precise calculations
	origFloat := gocv.NewMat()
	defer origFloat.Close()
	procFloat := gocv.NewMat()
	defer procFloat.Close()

	orig.ConvertTo(&origFloat, gocv.MatTypeCV64F)
	proc.ConvertTo(&procFloat, gocv.MatTypeCV64F)

	// Calculate MSE using OpenCV norm function for better precision
	diff := gocv.NewMat()
	defer diff.Close()

	err := gocv.Subtract(origFloat, procFloat, &diff)
	if err != nil {
		return 0.0
	}

	// Use L2 norm squared for MSE calculation
	normValue := gocv.Norm(diff, gocv.NormL2)
	normValueSquared := normValue * normValue
	totalPixels := float64(orig.Total())

	// Handle potential division by zero
	if totalPixels == 0 {
		return 0.0
	}

	mse := normValueSquared / totalPixels

	// Handle edge cases
	if mse == 0 {
		return 100.0 // Perfect match - return high but finite value
	}
	if mse < 1e-15 {
		return 100.0 // Very small differences
	}

	// Prevent NaN and infinite results
	if math.IsInf(mse, 0) || math.IsNaN(mse) {
		return 0.0
	}

	// Standard PSNR formula: PSNR = 20 * log10(MAX_I / sqrt(MSE))
	// where MAX_I = 255 for 8-bit images
	maxI := 255.0
	psnr := 20*math.Log10(maxI) - 10*math.Log10(mse)

	// Clamp to reasonable range
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

// SSIM calculation with mathematical correctness and numerical stability
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

	// Convert to grayscale for SSIM calculation if needed
	if orig.Channels() > 1 {
		origGray := gocv.NewMat()
		defer origGray.Close()
		err := gocv.CvtColor(orig, &origGray, gocv.ColorBGRToGray)
		if err != nil {
			return 0.0
		}
		orig.Close()
		orig = origGray.Clone()
	}

	if proc.Channels() > 1 {
		procGray := gocv.NewMat()
		defer procGray.Close()
		err := gocv.CvtColor(proc, &procGray, gocv.ColorBGRToGray)
		if err != nil {
			return 0.0
		}
		proc.Close()
		proc = procGray.Clone()
	}

	// Convert to float64 for calculations with proper normalization
	origF := gocv.NewMat()
	defer origF.Close()
	procF := gocv.NewMat()
	defer procF.Close()

	orig.ConvertTo(&origF, gocv.MatTypeCV64F)
	proc.ConvertTo(&procF, gocv.MatTypeCV64F)

	// Scale to [0,1] range for mathematical correctness
	origF.DivideFloat(255.0)
	procF.DivideFloat(255.0)

	// SSIM constants with standard values
	c1 := 0.01 * 0.01 // (k1 * L)^2 where k1=0.01, L=1 for normalized images
	c2 := 0.03 * 0.03 // (k2 * L)^2 where k2=0.03, L=1 for normalized images

	// Calculate means using Gaussian kernel for better perceptual relevance
	kernel := gocv.GetGaussianKernel(11, 1.5, gocv.MatTypeCV64F)
	defer kernel.Close()

	mu1 := gocv.NewMat()
	defer mu1.Close()
	mu2 := gocv.NewMat()
	defer mu2.Close()

	err := gocv.Filter2D(origF, &mu1, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}
	err = gocv.Filter2D(procF, &mu2, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}

	// Calculate mu1*mu2, mu1^2, mu2^2
	mu1Mu2 := gocv.NewMat()
	defer mu1Mu2.Close()
	mu1Sq := gocv.NewMat()
	defer mu1Sq.Close()
	mu2Sq := gocv.NewMat()
	defer mu2Sq.Close()

	err = gocv.Multiply(mu1, mu2, &mu1Mu2)
	if err != nil {
		return 0.0
	}
	err = gocv.Multiply(mu1, mu1, &mu1Sq)
	if err != nil {
		return 0.0
	}
	err = gocv.Multiply(mu2, mu2, &mu2Sq)
	if err != nil {
		return 0.0
	}

	// Calculate sigma1^2, sigma2^2, sigma12
	origF2 := gocv.NewMat()
	defer origF2.Close()
	procF2 := gocv.NewMat()
	defer procF2.Close()
	origFProcF := gocv.NewMat()
	defer origFProcF.Close()

	err = gocv.Multiply(origF, origF, &origF2)
	if err != nil {
		return 0.0
	}
	err = gocv.Multiply(procF, procF, &procF2)
	if err != nil {
		return 0.0
	}
	err = gocv.Multiply(origF, procF, &origFProcF)
	if err != nil {
		return 0.0
	}

	sigma1Sq := gocv.NewMat()
	defer sigma1Sq.Close()
	sigma2Sq := gocv.NewMat()
	defer sigma2Sq.Close()
	sigma12 := gocv.NewMat()
	defer sigma12.Close()

	temp1 := gocv.NewMat()
	defer temp1.Close()
	temp2 := gocv.NewMat()
	defer temp2.Close()
	temp3 := gocv.NewMat()
	defer temp3.Close()

	err = gocv.Filter2D(origF2, &temp1, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}
	err = gocv.Subtract(temp1, mu1Sq, &sigma1Sq)
	if err != nil {
		return 0.0
	}

	err = gocv.Filter2D(procF2, &temp2, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}
	err = gocv.Subtract(temp2, mu2Sq, &sigma2Sq)
	if err != nil {
		return 0.0
	}

	err = gocv.Filter2D(origFProcF, &temp3, -1, kernel, image.Point{X: -1, Y: -1}, 0, gocv.BorderReflect101)
	if err != nil {
		return 0.0
	}
	err = gocv.Subtract(temp3, mu1Mu2, &sigma12)
	if err != nil {
		return 0.0
	}

	// Calculate SSIM map
	numerator1 := gocv.NewMat()
	defer numerator1.Close()
	numerator2 := gocv.NewMat()
	defer numerator2.Close()
	denominator1 := gocv.NewMat()
	defer denominator1.Close()
	denominator2 := gocv.NewMat()
	defer denominator2.Close()

	// Numerator: (2*mu1*mu2 + C1) * (2*sigma12 + C2)
	mu1Mu2Times2 := gocv.NewMat()
	defer mu1Mu2Times2.Close()
	sigma12Times2 := gocv.NewMat()
	defer sigma12Times2.Close()

	mu1Mu2.MultiplyFloat(2.0)
	mu1Mu2.CopyTo(&mu1Mu2Times2)
	numerator1.SetTo(gocv.NewScalar(c1, 0, 0, 0))
	err = gocv.Add(mu1Mu2Times2, numerator1, &numerator1)
	if err != nil {
		return 0.0
	}

	sigma12.MultiplyFloat(2.0)
	sigma12.CopyTo(&sigma12Times2)
	numerator2.SetTo(gocv.NewScalar(c2, 0, 0, 0))
	err = gocv.Add(sigma12Times2, numerator2, &numerator2)
	if err != nil {
		return 0.0
	}

	numerator := gocv.NewMat()
	defer numerator.Close()
	err = gocv.Multiply(numerator1, numerator2, &numerator)
	if err != nil {
		return 0.0
	}

	// Denominator: (mu1^2 + mu2^2 + C1) * (sigma1^2 + sigma2^2 + C2)
	err = gocv.Add(mu1Sq, mu2Sq, &denominator1)
	if err != nil {
		return 0.0
	}
	denominator1.AddFloat(c1)

	err = gocv.Add(sigma1Sq, sigma2Sq, &denominator2)
	if err != nil {
		return 0.0
	}
	denominator2.AddFloat(c2)

	denominator := gocv.NewMat()
	defer denominator.Close()
	err = gocv.Multiply(denominator1, denominator2, &denominator)
	if err != nil {
		return 0.0
	}

	// Calculate final SSIM
	ssimMap := gocv.NewMat()
	defer ssimMap.Close()
	err = gocv.Divide(numerator, denominator, &ssimMap)
	if err != nil {
		return 0.0
	}

	// Calculate mean SSIM value
	meanSSIM := ssimMap.Mean()
	ssim := meanSSIM.Val1

	// Bounds checking and numerical stability
	if math.IsInf(ssim, 0) || math.IsNaN(ssim) {
		return 0.0
	}

	// Clamp to valid range [0, 1]
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

	if atomic.LoadInt32(&p.initialized) == 1 {
		p.cleanupResourcesUnsafe()
		for _, transform := range p.transformations {
			if transform != nil {
				transform.Close()
			}
		}
		p.transformations = nil
		atomic.StoreInt32(&p.initialized, 0)
	}
}
