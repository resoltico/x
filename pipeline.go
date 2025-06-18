package main

import (
	"fmt"
	"math"

	"gocv.io/x/gocv"
)

type ImagePipeline struct {
	originalImage   gocv.Mat
	processedImage  gocv.Mat
	transformations []Transformation
	debugPipeline   *DebugPipeline
	debugMemory     *DebugMemory
	initialized     bool
}

func NewImagePipeline() *ImagePipeline {
	pipeline := &ImagePipeline{
		transformations: make([]Transformation, 0),
		debugPipeline:   NewDebugPipeline(),
		debugMemory:     NewDebugMemory(),
		initialized:     false,
	}
	// Initialize with empty Mats to prevent segfaults
	pipeline.originalImage = gocv.NewMat()
	pipeline.processedImage = gocv.NewMat()
	pipeline.debugMemory.LogMatCreation("originalImage")
	pipeline.debugMemory.LogMatCreation("processedImage")
	return pipeline
}

func (p *ImagePipeline) SetOriginalImage(img gocv.Mat) {
	p.debugPipeline.LogSetOriginalStart()

	// Close existing images first
	if p.initialized && !p.originalImage.Empty() {
		p.debugPipeline.LogSetOriginalStep("closing existing original image")
		p.originalImage.Close()
		p.debugMemory.LogMatCleanup("originalImage")
	}
	if p.initialized && !p.processedImage.Empty() {
		p.debugPipeline.LogSetOriginalStep("closing existing processed image")
		p.processedImage.Close()
		p.debugMemory.LogMatCleanup("processedImage")
	}

	// Clear existing transformations when loading a new image
	p.debugPipeline.LogSetOriginalStep("clearing existing transformations")
	p.transformations = make([]Transformation, 0)

	// Set up new image
	p.originalImage = img.Clone()
	p.processedImage = p.originalImage.Clone()
	p.initialized = true
	p.debugPipeline.LogImageStats("original", p.originalImage)

	// Process the image with no transformations
	p.processImage()
}

func (p *ImagePipeline) AddTransformation(transformation Transformation) {
	p.transformations = append(p.transformations, transformation)
	p.processImage()
}

func (p *ImagePipeline) RemoveTransformation(index int) {
	if index >= 0 && index < len(p.transformations) {
		p.transformations = append(p.transformations[:index], p.transformations[index+1:]...)
		p.processImage()
	}
}

func (p *ImagePipeline) ClearTransformations() {
	p.transformations = make([]Transformation, 0)
	p.processImage()
}

func (p *ImagePipeline) GetProcessedImage() gocv.Mat {
	if !p.initialized {
		p.debugPipeline.LogGetProcessedImage("not initialized, returning empty Mat")
		return gocv.NewMat()
	}
	if p.processedImage.Empty() {
		p.debugPipeline.LogGetProcessedImage("processed image empty, returning original")
		return p.originalImage
	}
	return p.processedImage
}

func (p *ImagePipeline) processImage() {
	p.debugPipeline.LogProcessStart()
	if !p.initialized {
		p.debugPipeline.LogProcessEarlyReturn("not initialized")
		return
	}
	if p.originalImage.Empty() {
		p.debugPipeline.LogProcessEarlyReturn("original image is empty")
		return
	}

	p.debugPipeline.StartTimer("processImage")
	defer func() {
		p.debugPipeline.EndTimer("processImage")
		if p.initialized && !p.processedImage.Empty() {
			p.debugPipeline.LogPipelineStats(p.originalImage.Size(), p.processedImage.Size(), len(p.transformations))
		}
		p.debugPipeline.LogMemoryUsage()
	}()

	// Start with original image
	p.debugPipeline.LogProcessStep("checking if processedImage needs reset")
	if p.initialized && !p.processedImage.Empty() {
		p.debugPipeline.LogProcessStep("closing existing processedImage")
		p.processedImage.Close()
		p.debugMemory.LogMatCleanup("processedImage")
	}
	p.debugPipeline.LogProcessStep("cloning original image")
	p.processedImage = p.originalImage.Clone()

	// Apply all transformations sequentially
	p.debugPipeline.LogTransformationCount(len(p.transformations))
	for i, transformation := range p.transformations {
		p.debugPipeline.StartTimer(fmt.Sprintf("transformation_%d_%s", i, transformation.Name()))

		before := p.processedImage.Clone()
		result := transformation.Apply(p.processedImage)
		duration := p.debugPipeline.EndTimer(fmt.Sprintf("transformation_%d_%s", i, transformation.Name()))

		p.debugPipeline.LogTransformationApplied(transformation.Name(), before, result, duration)

		p.processedImage.Close()
		p.processedImage = result
		before.Close()
	}
	p.debugPipeline.LogProcessComplete()
}

func (p *ImagePipeline) CalculatePSNR() float64 {
	if !p.initialized || p.originalImage.Empty() || p.processedImage.Empty() {
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
	if !p.initialized || p.originalImage.Empty() || p.processedImage.Empty() {
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
	if p.initialized {
		if !p.originalImage.Empty() {
			p.originalImage.Close()
			p.debugMemory.LogMatCleanup("originalImage final")
		}
		if !p.processedImage.Empty() {
			p.processedImage.Close()
			p.debugMemory.LogMatCleanup("processedImage final")
		}
	}
}
