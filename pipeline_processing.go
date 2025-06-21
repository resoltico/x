package main

import (
	"fmt"
	"sync/atomic"
)

func (p *ImagePipeline) ProcessImage() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.debugPipeline.Log("ProcessImage: Force reprocessing full resolution image")
	return p.processImageUnsafe()
}

func (p *ImagePipeline) ProcessPreview() error {
	p.processingMutex.Lock()
	defer p.processingMutex.Unlock()

	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.processPreviewUnsafe()
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

	p.debugPipeline.LogProcessStep("creating new processed image")
	newProcessed := p.originalImage.Clone()
	if newProcessed.Empty() {
		return fmt.Errorf("failed to clone original for processing")
	}

	defer func() {
		if err != nil && !newProcessed.Empty() {
			newProcessed.Close()
		}
	}()

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

		if !before.Empty() {
			before.Close()
		}

		if result.Empty() {
			return fmt.Errorf("transformation %s returned empty result", transformation.Name())
		}

		if !newProcessed.Empty() {
			newProcessed.Close()
		}
		newProcessed = result
	}

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

	newPreview := p.originalImage.Clone()
	if newPreview.Empty() {
		return fmt.Errorf("failed to clone original for preview")
	}

	defer func() {
		if err != nil && !newPreview.Empty() {
			newPreview.Close()
		}
	}()

	for i, transformation := range p.transformations {
		if transformation == nil {
			return fmt.Errorf("preview transformation %d is nil", i)
		}

		timerName := fmt.Sprintf("preview_transformation_%d_%s", i, transformation.Name())
		p.debugPipeline.StartTimer(timerName)

		before := newPreview.Clone()
		result := transformation.ApplyPreview(newPreview)
		duration := p.debugPipeline.EndTimer(timerName)

		p.debugPipeline.LogTransformationApplied(transformation.Name()+" (preview)", before, result, duration)

		if !before.Empty() {
			before.Close()
		}

		if result.Empty() {
			return fmt.Errorf("preview transformation %s returned empty result", transformation.Name())
		}

		if !newPreview.Empty() {
			newPreview.Close()
		}
		newPreview = result
	}

	if !p.previewImage.Empty() {
		p.previewImage.Close()
	}
	p.previewImage = newPreview

	return nil
}
