// Real-time processing pipeline with preview support
package core

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"fyne.io/fyne/v2"
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

// ProcessingPipeline manages real-time image processing
type ProcessingPipeline struct {
	mu            sync.RWMutex
	imageData     *ImageData
	regionManager *RegionManager
	metricsEval   *metrics.Evaluator
	logger        *slog.Logger

	steps []ProcessingStep

	processing bool
	cancel     context.CancelFunc

	// Real-time callbacks
	onPreviewUpdate func(preview gocv.Mat, metrics map[string]float64)
	onError         func(error)

	// Real-time processing
	previewTimer *time.Timer
	previewDelay time.Duration
	realtimeMode bool
}

func NewProcessingPipeline(imageData *ImageData, regionManager *RegionManager, logger *slog.Logger) *ProcessingPipeline {
	return &ProcessingPipeline{
		imageData:     imageData,
		regionManager: regionManager,
		metricsEval:   metrics.NewEvaluator(),
		logger:        logger,
		steps:         make([]ProcessingStep, 0),
		previewDelay:  200 * time.Millisecond, // Fast real-time updates
		realtimeMode:  true,
	}
}

func (pp *ProcessingPipeline) SetCallbacks(
	onPreviewUpdate func(gocv.Mat, map[string]float64),
	onError func(error),
) {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.onPreviewUpdate = onPreviewUpdate
	pp.onError = onError
}

func (pp *ProcessingPipeline) SetRealtimeMode(enabled bool) {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.realtimeMode = enabled
}

func (pp *ProcessingPipeline) AddStep(algorithm string, parameters map[string]interface{}) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if !algorithms.IsValidAlgorithm(algorithm) {
		return fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	if err := algorithms.ValidateParameters(algorithm, parameters); err != nil {
		return fmt.Errorf("invalid parameters for %s: %w", algorithm, err)
	}

	step := ProcessingStep{
		Algorithm:  algorithm,
		Parameters: parameters,
		Enabled:    true,
	}

	pp.steps = append(pp.steps, step)
	pp.logger.Debug("Added processing step", "algorithm", algorithm, "parameters", parameters)

	// Trigger real-time preview
	if pp.realtimeMode {
		pp.triggerPreviewProcessing()
	}

	return nil
}

func (pp *ProcessingPipeline) RemoveStep(index int) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if index < 0 || index >= len(pp.steps) {
		return fmt.Errorf("invalid step index: %d", index)
	}

	pp.steps = append(pp.steps[:index], pp.steps[index+1:]...)
	pp.logger.Debug("Removed processing step", "index", index)

	// Trigger real-time preview
	if pp.realtimeMode {
		pp.triggerPreviewProcessing()
	}

	return nil
}

func (pp *ProcessingPipeline) UpdateStep(index int, parameters map[string]interface{}) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if index < 0 || index >= len(pp.steps) {
		return fmt.Errorf("invalid step index: %d", index)
	}

	step := &pp.steps[index]

	if err := algorithms.ValidateParameters(step.Algorithm, parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	step.Parameters = parameters
	pp.logger.Debug("Updated processing step", "index", index, "algorithm", step.Algorithm, "parameters", parameters)

	// Trigger real-time preview
	if pp.realtimeMode {
		pp.triggerPreviewProcessing()
	}

	return nil
}

func (pp *ProcessingPipeline) ToggleStep(index int, enabled bool) error {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if index < 0 || index >= len(pp.steps) {
		return fmt.Errorf("invalid step index: %d", index)
	}

	pp.steps[index].Enabled = enabled
	pp.logger.Debug("Toggled processing step", "index", index, "enabled", enabled)

	// Trigger real-time preview
	if pp.realtimeMode {
		pp.triggerPreviewProcessing()
	}

	return nil
}

func (pp *ProcessingPipeline) GetSteps() []ProcessingStep {
	pp.mu.RLock()
	defer pp.mu.RUnlock()

	steps := make([]ProcessingStep, len(pp.steps))
	copy(steps, pp.steps)
	return steps
}

func (pp *ProcessingPipeline) ClearSteps() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	pp.steps = make([]ProcessingStep, 0)
	pp.logger.Debug("Cleared all processing steps")

	// Reset to original image
	if pp.imageData.HasImage() {
		pp.imageData.ResetToOriginal()

		// Show original preview
		if pp.onPreviewUpdate != nil {
			original := pp.imageData.GetPreview()
			fyne.Do(func() {
				pp.onPreviewUpdate(original, nil)
			})
		}
	}
}

// triggerPreviewProcessing triggers debounced real-time preview processing
func (pp *ProcessingPipeline) triggerPreviewProcessing() {
	if !pp.imageData.HasImage() {
		return
	}

	// Cancel existing timer
	if pp.previewTimer != nil {
		pp.previewTimer.Stop()
	}

	// Start new timer for debounced processing
	pp.previewTimer = time.AfterFunc(pp.previewDelay, func() {
		pp.processPreview()
	})
}

// processPreview processes the preview image in real-time
func (pp *ProcessingPipeline) processPreview() {
	pp.mu.Lock()
	if pp.processing {
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

		// Get preview image for fast processing
		preview := pp.imageData.GetPreview()
		defer preview.Close()

		if preview.Empty() {
			if pp.onError != nil {
				fyne.Do(func() {
					pp.onError(fmt.Errorf("no preview image available"))
				})
			}
			return
		}

		// Process preview with current steps
		result, processMetrics := pp.processImage(ctx, preview)
		if result.Empty() {
			return // Processing was cancelled or failed
		}

		// Calculate final metrics
		finalMetrics := pp.calculatePreviewMetrics(preview, result)

		// Merge metrics
		allMetrics := make(map[string]float64)
		for k, v := range processMetrics {
			allMetrics[k] = v
		}
		for k, v := range finalMetrics {
			allMetrics[k] = v
		}

		// Report preview update
		if pp.onPreviewUpdate != nil {
			resultCopy := result.Clone()
			fyne.Do(func() {
				pp.onPreviewUpdate(resultCopy, allMetrics)
			})
		}

		result.Close()
		pp.logger.Debug("Preview processing completed")
	}()
}

// processImage processes an image through the pipeline
func (pp *ProcessingPipeline) processImage(ctx context.Context, input gocv.Mat) (gocv.Mat, map[string]float64) {
	current := input.Clone()
	processMetrics := make(map[string]float64)

	steps := pp.GetSteps()
	for _, step := range steps {
		select {
		case <-ctx.Done():
			current.Close()
			return gocv.NewMat(), nil
		default:
		}

		if !step.Enabled {
			continue
		}

		// Apply algorithm
		result, err := algorithms.Apply(step.Algorithm, current, step.Parameters)
		if err != nil {
			pp.logger.Error("Algorithm failed", "algorithm", step.Algorithm, "error", err)
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

// calculatePreviewMetrics calculates metrics for preview
func (pp *ProcessingPipeline) calculatePreviewMetrics(original, processed gocv.Mat) map[string]float64 {
	finalMetrics := make(map[string]float64)

	// Calculate PSNR
	if psnr, err := pp.metricsEval.CalculatePSNR(original, processed); err == nil {
		finalMetrics["psnr"] = psnr
	}

	// Calculate SSIM
	if ssim, err := pp.metricsEval.CalculateSSIM(original, processed); err == nil {
		finalMetrics["ssim"] = ssim
	}

	return finalMetrics
}

// ProcessFullResolution processes the full resolution image (for saving)
func (pp *ProcessingPipeline) ProcessFullResolution() (gocv.Mat, error) {
	if !pp.imageData.HasImage() {
		return gocv.NewMat(), fmt.Errorf("no image loaded")
	}

	original := pp.imageData.GetOriginal()
	defer original.Close()

	ctx := context.Background()
	result, _ := pp.processImage(ctx, original)

	if result.Empty() {
		return gocv.NewMat(), fmt.Errorf("processing failed")
	}

	// Update the processed image in imageData
	pp.imageData.SetProcessed(result)

	return result.Clone(), nil
}

func (pp *ProcessingPipeline) IsProcessing() bool {
	pp.mu.RLock()
	defer pp.mu.RUnlock()
	return pp.processing
}

func (pp *ProcessingPipeline) Stop() {
	pp.mu.Lock()
	defer pp.mu.Unlock()

	if pp.cancel != nil {
		pp.cancel()
	}
	if pp.previewTimer != nil {
		pp.previewTimer.Stop()
	}
}
