// Processing pipeline with ROI support and metrics integration
package core

import (
	"context"
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"github.com/sirupsen/logrus"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/algorithms"
	"advanced-image-processing/internal/metrics"
)

// ProcessingStep represents a single processing step
type ProcessingStep struct {
	Algorithm  string                 `json:"algorithm"`
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
}

// ProcessingPipeline manages the image processing workflow
type ProcessingPipeline struct {
	mu            sync.RWMutex
	imageData     *ImageData
	regionManager *RegionManager
	metricsEval   *metrics.Evaluator
	logger        *logrus.Logger

	// Processing steps
	steps []ProcessingStep

	// Processing state
	processing bool
	cancel     context.CancelFunc

	// Callbacks
	onProgress func(step int, total int, stepName string)
	onComplete func(result gocv.Mat, metrics map[string]float64)
	onError    func(error)

	// Debouncing
	debounceTimer *time.Timer
	debounceDelay time.Duration
}

// NewProcessingPipeline creates a new processing pipeline
func NewProcessingPipeline(imageData *ImageData, regionManager *RegionManager, logger *logrus.Logger) *ProcessingPipeline {
	return &ProcessingPipeline{
		imageData:     imageData,
		regionManager: regionManager,
		metricsEval:   metrics.NewEvaluator(),
		logger:        logger,
		steps:         make([]ProcessingStep, 0),
		debounceDelay: 300 * time.Millisecond,
	}
}

// SetCallbacks sets the callback functions
func (pp *ProcessingPipeline) SetCallbacks(
	onProgress func(int, int, string),
	onComplete func(gocv.Mat, map[string]float64),
	onError func(error),
) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.onProgress = onProgress
	pp.onComplete = onComplete
	pp.onError = onError
}

// AddStep adds a processing step
func (pp *ProcessingPipeline) AddStep(algorithm string, parameters map[string]interface{}) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	// Validate algorithm exists
	if !algorithms.IsValidAlgorithm(algorithm) {
		return fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	// Validate parameters
	if err := algorithms.ValidateParameters(algorithm, parameters); err != nil {
		return fmt.Errorf("invalid parameters for %s: %w", algorithm, err)
	}

	step := ProcessingStep{
		Algorithm:  algorithm,
		Parameters: parameters,
		Enabled:    true,
	}

	pp.steps = append(pp.steps, step)
	pp.logger.WithFields(logrus.Fields{
		"algorithm":  algorithm,
		"parameters": parameters,
	}).Debug("Added processing step")

	// Trigger processing
	pp.triggerProcessing()

	return nil
}

// RemoveStep removes a processing step
func (pp *ProcessingPipeline) RemoveStep(index int) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if index < 0 || index >= len(pp.steps) {
		return fmt.Errorf("invalid step index: %d", index)
	}

	pp.steps = append(pp.steps[:index], pp.steps[index+1:]...)
	pp.logger.WithField("index", index).Debug("Removed processing step")

	// Trigger processing
	pp.triggerProcessing()

	return nil
}

// UpdateStep updates a processing step
func (pp *ProcessingPipeline) UpdateStep(index int, parameters map[string]interface{}) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if index < 0 || index >= len(pp.steps) {
		return fmt.Errorf("invalid step index: %d", index)
	}

	step := &pp.steps[index]

	// Validate parameters
	if err := algorithms.ValidateParameters(step.Algorithm, parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	step.Parameters = parameters
	pp.logger.WithFields(logrus.Fields{
		"index":      index,
		"algorithm":  step.Algorithm,
		"parameters": parameters,
	}).Debug("Updated processing step")

	// Trigger processing
	pp.triggerProcessing()

	return nil
}

// ToggleStep enables or disables a processing step
func (pp *ProcessingPipeline) ToggleStep(index int, enabled bool) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if index < 0 || index >= len(pp.steps) {
		return fmt.Errorf("invalid step index: %d", index)
	}

	pp.steps[index].Enabled = enabled
	pp.logger.WithFields(logrus.Fields{
		"index":   index,
		"enabled": enabled,
	}).Debug("Toggled processing step")

	// Trigger processing
	pp.triggerProcessing()

	return nil
}

// GetSteps returns a copy of all processing steps
func (pp *ProcessingPipeline) GetSteps() []ProcessingStep {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	steps := make([]ProcessingStep, len(pp.steps))
	copy(steps, pp.steps)
	return steps
}

// ClearSteps removes all processing steps
func (pp *ProcessingPipeline) ClearSteps() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.steps = make([]ProcessingStep, 0)
	pp.logger.Debug("Cleared all processing steps")

	// Reset to original image
	if pp.imageData.HasImage() {
		pp.imageData.ResetToOriginal()
		
		// Notify completion with original image
		if pp.onComplete != nil {
			original := pp.imageData.GetOriginal()
			fyne.Do(func() {
				pp.onComplete(original, nil)
			})
		}
	}
}

// triggerProcessing triggers debounced processing
func (pp *ProcessingPipeline) triggerProcessing() {
	if !pp.imageData.HasImage() {
		return
	}

	// Cancel existing timer
	if pp.debounceTimer != nil {
		pp.debounceTimer.Stop()
	}

	// Start new timer
	pp.debounceTimer = time.AfterFunc(pp.debounceDelay, func() {
		pp.processImage()
	})
}

// processImage processes the image through the pipeline
func (pp *ProcessingPipeline) processImage() {
	pp.mu.Lock()
	if pp.processing {
		// Cancel previous processing
		if pp.cancel != nil {
			pp.cancel()
		}
	}
	pp.processing = true
	pp.mu.Unlock()

	// Create cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	pp.mu.Lock()
	pp.cancel = cancel
	pp.mu.Unlock()

	go func() {
		defer func() {
			pp.mu.Lock()
			pp.processing = false
			pp.cancel = nil
			pp.mu.Unlock()
		}()

		// Get original image
		original := pp.imageData.GetOriginal()
		defer original.Close()

		if original.Empty() {
			if pp.onError != nil {
				fyne.Do(func() {
					pp.onError(fmt.Errorf("no image loaded"))
				})
			}
			return
		}

		var result gocv.Mat
		var processMetrics map[string]float64

		// Check if we have ROI selection
		if pp.regionManager.HasActiveSelection() {
			result, processMetrics = pp.processWithROI(ctx, original)
		} else {
			result, processMetrics = pp.processFullImage(ctx, original)
		}

		if result.Empty() {
			return // Processing was cancelled or failed
		}

		// Update processed image
		pp.imageData.SetProcessed(result)

		// Calculate final metrics
		finalMetrics := pp.calculateFinalMetrics(original, result)

		// Merge processing metrics with final metrics
		allMetrics := make(map[string]float64)
		for k, v := range processMetrics {
			allMetrics[k] = v
		}
		for k, v := range finalMetrics {
			allMetrics[k] = v
		}

		// Report completion
		if pp.onComplete != nil {
			resultCopy := result.Clone()
			fyne.Do(func() {
				pp.onComplete(resultCopy, allMetrics)
			})
		}

		result.Close()
		pp.logger.Debug("Pipeline processing completed")
	}()
}

// processFullImage processes the entire image
func (pp *ProcessingPipeline) processFullImage(ctx context.Context, original gocv.Mat) (gocv.Mat, map[string]float64) {
	current := original.Clone()
	processMetrics := make(map[string]float64)

	steps := pp.GetSteps()
	for i, step := range steps {
		select {
		case <-ctx.Done():
			current.Close()
			return gocv.NewMat(), nil
		default:
		}

		if !step.Enabled {
			continue
		}

		// Report progress
		if pp.onProgress != nil {
			fyne.Do(func() {
				pp.onProgress(i+1, len(steps), step.Algorithm)
			})
		}

		// Apply algorithm
		result, err := algorithms.Apply(step.Algorithm, current, step.Parameters)
		if err != nil {
			pp.logger.WithError(err).WithField("algorithm", step.Algorithm).Error("Algorithm failed")
			if pp.onError != nil {
				fyne.Do(func() {
					pp.onError(fmt.Errorf("algorithm %s failed: %w", step.Algorithm, err))
				})
			}
			current.Close()
			return gocv.NewMat(), nil
		}

		if result.Empty() {
			current.Close()
			if pp.onError != nil {
				fyne.Do(func() {
					pp.onError(fmt.Errorf("algorithm %s produced empty result", step.Algorithm))
				})
			}
			return gocv.NewMat(), nil
		}

		// Calculate step metrics
		stepMetrics := pp.metricsEval.EvaluateStep(current, result, step.Algorithm)
		for k, v := range stepMetrics {
			processMetrics[fmt.Sprintf("%s_%s", step.Algorithm, k)] = v
		}

		current.Close()
		current = result
	}

	return current, processMetrics
}

// processWithROI processes only the selected region
func (pp *ProcessingPipeline) processWithROI(ctx context.Context, original gocv.Mat) (gocv.Mat, map[string]float64) {
	// Create mask for active selection
	mask := pp.regionManager.CreateMask(original.Cols(), original.Rows())
	defer mask.Close()

	if mask.Empty() {
		return pp.processFullImage(ctx, original)
	}

	// Extract ROI region
	roi := gocv.NewMat()
	original.CopyToWithMask(&roi, mask)
	defer roi.Close()

	// Process the ROI
	processedROI, processMetrics := pp.processFullImage(ctx, roi)
	if processedROI.Empty() {
		return gocv.NewMat(), nil
	}
	defer processedROI.Close()

	// Combine processed ROI with original image
	result := original.Clone()
	processedROI.CopyToWithMask(&result, mask)

	return result, processMetrics
}

// calculateFinalMetrics calculates final quality metrics
func (pp *ProcessingPipeline) calculateFinalMetrics(original, processed gocv.Mat) map[string]float64 {
	finalMetrics := make(map[string]float64)

	// Calculate PSNR
	if psnr, err := pp.metricsEval.CalculatePSNR(original, processed); err == nil {
		finalMetrics["psnr"] = psnr
	}

	// Calculate SSIM
	if ssim, err := pp.metricsEval.CalculateSSIM(original, processed); err == nil {
		finalMetrics["ssim"] = ssim
	}

	// Calculate F-measure for binarized images
	if pp.isBinarized(processed) {
		if fMeasure, err := pp.metricsEval.CalculateFMeasure(original, processed); err == nil {
			finalMetrics["f_measure"] = fMeasure
		}
	}

	return finalMetrics
}

// isBinarized checks if an image is binary (only contains 0 and 255 values)
func (pp *ProcessingPipeline) isBinarized(mat gocv.Mat) bool {
	if mat.Empty() || mat.Channels() != 1 {
		return false
	}

	// Sample a few pixels to check if they're binary
	minVal, maxVal, _, _ := gocv.MinMaxLoc(mat)
	return minVal == 0.0 && maxVal == 255.0
}

// IsProcessing returns true if currently processing
func (pp *ProcessingPipeline) IsProcessing() bool {
	pp.mu.RLock()
	defer pp.mu.RUnlock()
	return pp.processing
}

// Stop stops any ongoing processing
func (pp *ProcessingPipeline) Stop() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if pp.cancel != nil {
		pp.cancel()
	}
	if pp.debounceTimer != nil {
		pp.debounceTimer.Stop()
	}
}
