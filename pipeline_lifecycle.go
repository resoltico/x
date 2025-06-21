package main

import (
	"fmt"
	"sync/atomic"

	"gocv.io/x/gocv"
)

func (p *ImagePipeline) HasImage() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return atomic.LoadInt32(&p.initialized) == 1 && !p.originalImage.Empty()
}

func (p *ImagePipeline) HasImageUnsafe() bool {
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

	if img.Empty() {
		return fmt.Errorf("input image is empty")
	}

	if img.Cols() <= 0 || img.Rows() <= 0 || img.Cols() > 65536 || img.Rows() > 65536 {
		return fmt.Errorf("invalid image dimensions: %dx%d", img.Cols(), img.Rows())
	}

	if atomic.LoadInt32(&p.initialized) == 1 {
		p.debugPipeline.LogSetOriginalStep("cleaning up existing resources")
		p.cleanupResourcesUnsafe()
	}

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
