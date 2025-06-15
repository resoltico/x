// Author: Ervins Strauhmanis
// License: MIT

package image_processing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gocv.io/x/gocv"
	"github.com/sirupsen/logrus"

	"advanced-image-processing/internal/models"
	"advanced-image-processing/internal/transforms"
	"advanced-image-processing/internal/utils"
)

// Pipeline manages the transformation sequence and processing
type Pipeline struct {
	mu         sync.RWMutex
	registry   *transforms.TransformRegistry
	sequence   *models.TransformationSequence
	imageData  *models.ImageData
	logger     *logrus.Logger
	debouncer  *utils.Debouncer
	
	// Callbacks
	onProgress   func(step int, total int, stepName string)
	onComplete   func(result gocv.Mat)
	onError      func(error)
	
	// Processing state
	processing   bool
	cancel       context.CancelFunc
}

// NewPipeline creates a new transformation pipeline
func NewPipeline(registry *transforms.TransformRegistry, imageData *models.ImageData, logger *logrus.Logger) *Pipeline {
	return &Pipeline{
		registry:  registry,
		sequence:  models.NewTransformationSequence(),
		imageData: imageData,
		logger:    logger,
		debouncer: utils.NewDebouncer(300 * time.Millisecond), // 300ms debounce
	}
}

// SetCallbacks sets the callback functions for pipeline events
func (p *Pipeline) SetCallbacks(onProgress func(int, int, string), onComplete func(gocv.Mat), onError func(error)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	p.onProgress = onProgress
	p.onComplete = onComplete
	p.onError = onError
}

// AddTransformation adds a transformation to the pipeline
func (p *Pipeline) AddTransformation(transformType string, params map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Validate that the transformation exists
	transform, exists := p.registry.Get(transformType)
	if !exists {
		return fmt.Errorf("unknown transformation type: %s", transformType)
	}

	// Validate parameters
	if err := transform.Validate(params); err != nil {
		return fmt.Errorf("invalid parameters for %s: %w", transformType, err)
	}

	// Add to sequence
	p.sequence.AddStep(transformType, params)
	
	p.logger.WithFields(logrus.Fields{
		"transform": transformType,
		"params":    params,
	}).Debug("Added transformation to pipeline")

	// Trigger processing
	p.triggerProcessing()
	
	return nil
}

// RemoveTransformation removes a transformation from the pipeline
func (p *Pipeline) RemoveTransformation(index int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if index < 0 || index >= p.sequence.Length() {
		return fmt.Errorf("invalid transformation index: %d", index)
	}

	p.sequence.RemoveStep(index)
	p.logger.WithField("index", index).Debug("Removed transformation from pipeline")

	// Trigger processing
	p.triggerProcessing()
	
	return nil
}

// UpdateTransformation updates parameters of a transformation
func (p *Pipeline) UpdateTransformation(index int, params map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	steps := p.sequence.GetSteps()
	if index < 0 || index >= len(steps) {
		return fmt.Errorf("invalid transformation index: %d", index)
	}

	// Validate parameters
	transform, exists := p.registry.Get(steps[index].Type)
	if !exists {
		return fmt.Errorf("unknown transformation type: %s", steps[index].Type)
	}

	if err := transform.Validate(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	p.sequence.UpdateStep(index, params)
	p.logger.WithFields(logrus.Fields{
		"index":  index,
		"params": params,
	}).Debug("Updated transformation parameters")

	// Trigger processing
	p.triggerProcessing()
	
	return nil
}

// triggerProcessing triggers debounced processing
func (p *Pipeline) triggerProcessing() {
	if !p.imageData.HasImage() {
		return
	}

	p.debouncer.Debounce(func() {
		p.processImage()
	})
}

// processImage processes the image through the transformation pipeline
func (p *Pipeline) processImage() {
	p.mu.Lock()
	if p.processing {
		// Cancel previous processing
		if p.cancel != nil {
			p.cancel()
		}
	}
	p.processing = true
	p.mu.Unlock()

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	p.mu.Lock()
	p.cancel = cancel
	p.mu.Unlock()

	go func() {
		defer func() {
			p.mu.Lock()
			p.processing = false
			p.cancel = nil
			p.mu.Unlock()
		}()

		// Get original image
		original := p.imageData.GetOriginal()
		if original.Empty() {
			if p.onError != nil {
				p.onError(fmt.Errorf("no image loaded"))
			}
			return
		}
		defer original.Close()

		// Get transformation steps
		steps := p.sequence.GetSteps()
		
		// Start with original image
		current := original.Clone()
		defer current.Close()

		// Apply each transformation
		for i, step := range steps {
			select {
			case <-ctx.Done():
				return // Processing was cancelled
			default:
			}

			if !step.Enabled {
				continue
			}

			// Report progress
			if p.onProgress != nil {
				p.onProgress(i+1, len(steps), step.Type)
			}

			// Get transformation
			transform, exists := p.registry.Get(step.Type)
			if !exists {
				if p.onError != nil {
					p.onError(fmt.Errorf("unknown transformation: %s", step.Type))
				}
				return
			}

			// Apply transformation
			result, err := transform.Apply(current, step.Parameters)
			if err != nil {
				p.logger.WithFields(logrus.Fields{
					"transform": step.Type,
					"error":     err,
				}).Error("Transformation failed")
				
				if p.onError != nil {
					p.onError(fmt.Errorf("transformation %s failed: %w", step.Type, err))
				}
				return
			}

			// Replace current with result
			current.Close()
			current = result

			p.logger.WithField("transform", step.Type).Debug("Transformation applied successfully")
		}

		// Update processed image
		p.imageData.SetProcessed(current)

		// Report completion
		if p.onComplete != nil {
			result := current.Clone()
			p.onComplete(result)
		}

		p.logger.Debug("Pipeline processing completed")
	}()
}

// GetTransformationSequence returns the current transformation sequence
func (p *Pipeline) GetTransformationSequence() *models.TransformationSequence {
	return p.sequence
}

// LoadSequence loads a transformation sequence
func (p *Pipeline) LoadSequence(sequence *models.TransformationSequence) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sequence = sequence
	p.logger.Debug("Loaded transformation sequence")

	// Trigger processing
	p.triggerProcessing()
}

// ClearSequence clears all transformations
func (p *Pipeline) ClearSequence() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sequence.Clear()
	p.logger.Debug("Cleared transformation sequence")

	// Reset to original image
	if p.imageData.HasImage() {
		original := p.imageData.GetOriginal()
		p.imageData.SetProcessed(original)
		original.Close()

		if p.onComplete != nil {
			result := p.imageData.GetProcessed()
			p.onComplete(result)
			result.Close()
		}
	}
}

// IsProcessing returns true if the pipeline is currently processing
func (p *Pipeline) IsProcessing() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.processing
}

// Stop stops any ongoing processing
func (p *Pipeline) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}
	p.debouncer.Cancel()
}