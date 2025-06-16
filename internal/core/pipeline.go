// internal/core/pipeline.go
// Fixed processing pipeline with proper preview handling
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
	"advanced-image-processing/internal/layers"
	"advanced-image-processing/internal/metrics"
)

// ProcessingStep represents a sequential processing step
type ProcessingStep struct {
	Algorithm  string
	Parameters map[string]interface{}
	Enabled    bool
}

// EnhancedPipeline combines sequential and layer-based processing
type EnhancedPipeline struct {
	mu            sync.RWMutex
	imageData     *ImageData
	regionManager *RegionManager
	layerStack    *layers.LayerStack
	metricsEval   *metrics.Evaluator
	logger        *slog.Logger

	// Sequential processing (existing)
	steps []ProcessingStep

	// Processing mode
	useLayerMode bool
	processing   bool
	cancel       context.CancelFunc

	// Callbacks
	onPreviewUpdate func(preview gocv.Mat, metrics map[string]float64)
	onError         func(error)

	// Real-time processing
	previewTimer *time.Timer
	previewDelay time.Duration
	realtimeMode bool
}

func NewEnhancedPipeline(imageData *ImageData, regionManager *RegionManager, logger *slog.Logger) *EnhancedPipeline {
	return &EnhancedPipeline{
		imageData:     imageData,
		regionManager: regionManager,
		layerStack:    layers.NewLayerStack(regionManager),
		metricsEval:   metrics.NewEvaluator(),
		logger:        logger,
		steps:         make([]ProcessingStep, 0),
		previewDelay:  200 * time.Millisecond,
		realtimeMode:  true,
		useLayerMode:  false,
	}
}

// SetProcessingMode switches between sequential and layer-based processing
func (ep *EnhancedPipeline) SetProcessingMode(useLayerMode bool) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.useLayerMode = useLayerMode
	ep.logger.Debug("Processing mode changed", "layer_mode", useLayerMode)

	if ep.realtimeMode {
		ep.triggerPreviewProcessing()
	}
}

// AddLayer adds a processing layer (layer mode)
func (ep *EnhancedPipeline) AddLayer(name, algorithm string, params map[string]interface{}, regionID string) (string, error) {
	if !algorithms.IsValidAlgorithm(algorithm) {
		return "", fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	if err := algorithms.ValidateParameters(algorithm, params); err != nil {
		return "", fmt.Errorf("invalid parameters: %w", err)
	}

	layerID := ep.layerStack.AddLayer(name, algorithm, params, regionID)
	ep.logger.Debug("Added processing layer", "layer_id", layerID, "algorithm", algorithm)

	// Trigger processing immediately when layer is added
	if ep.realtimeMode && ep.useLayerMode {
		ep.logger.Debug("Triggering preview processing after adding layer")
		ep.triggerPreviewProcessing()
	}

	return layerID, nil
}

// AddStep adds sequential processing step (sequential mode)
func (ep *EnhancedPipeline) AddStep(algorithm string, parameters map[string]interface{}) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if !algorithms.IsValidAlgorithm(algorithm) {
		return fmt.Errorf("unknown algorithm: %s", algorithm)
	}

	if err := algorithms.ValidateParameters(algorithm, parameters); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	step := ProcessingStep{
		Algorithm:  algorithm,
		Parameters: parameters,
		Enabled:    true,
	}

	ep.steps = append(ep.steps, step)
	ep.logger.Debug("Added processing step", "algorithm", algorithm)

	if ep.realtimeMode && !ep.useLayerMode {
		ep.triggerPreviewProcessing()
	}

	return nil
}

// SetCallbacks sets preview update and error callbacks
func (ep *EnhancedPipeline) SetCallbacks(
	onPreviewUpdate func(gocv.Mat, map[string]float64),
	onError func(error),
) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.onPreviewUpdate = onPreviewUpdate
	ep.onError = onError
}

// triggerPreviewProcessing starts debounced preview processing
func (ep *EnhancedPipeline) triggerPreviewProcessing() {
	if !ep.imageData.HasImage() {
		ep.logger.Debug("No image available for preview processing")
		return
	}

	if ep.previewTimer != nil {
		ep.previewTimer.Stop()
	}

	ep.logger.Debug("Scheduling preview processing", "delay_ms", ep.previewDelay.Milliseconds())
	ep.previewTimer = time.AfterFunc(ep.previewDelay, func() {
		ep.processPreview()
	})
}

// processPreview processes preview image in real-time
func (ep *EnhancedPipeline) processPreview() {
	ep.mu.Lock()
	if ep.processing {
		if ep.cancel != nil {
			ep.cancel()
		}
	}
	ep.processing = true
	ep.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	ep.mu.Lock()
	ep.cancel = cancel
	ep.mu.Unlock()

	go func() {
		defer func() {
			ep.mu.Lock()
			ep.processing = false
			ep.cancel = nil
			ep.mu.Unlock()
		}()

		ep.logger.Debug("Starting preview processing")

		preview := ep.imageData.GetPreview()
		defer preview.Close()

		if preview.Empty() {
			ep.logger.Error("Preview image is empty")
			if ep.onError != nil {
				fyne.Do(func() {
					ep.onError(fmt.Errorf("no preview image available"))
				})
			}
			return
		}

		var result gocv.Mat
		var err error

		// Process based on current mode
		ep.mu.RLock()
		useLayerMode := ep.useLayerMode
		ep.mu.RUnlock()

		ep.logger.Debug("Processing preview", "layer_mode", useLayerMode)

		if useLayerMode {
			result, err = ep.layerStack.ProcessLayers(preview)
			if err != nil {
				ep.logger.Error("Layer processing failed", "error", err)
			} else {
				ep.logger.Debug("Layer processing completed", "success", true)
			}
		} else {
			result, _ = ep.processSequential(ctx, preview)
			ep.logger.Debug("Sequential processing completed")
		}

		if err != nil {
			ep.logger.Error("Processing failed", "error", err)
			if !result.Empty() {
				result.Close()
			}
			if ep.onError != nil {
				fyne.Do(func() {
					ep.onError(fmt.Errorf("processing failed: %w", err))
				})
			}
			return
		}

		if result.Empty() {
			ep.logger.Error("Processing returned empty result")
			result.Close()
			if ep.onError != nil {
				fyne.Do(func() {
					ep.onError(fmt.Errorf("processing returned empty result"))
				})
			}
			return
		}

		// Calculate metrics
		finalMetrics := ep.calculatePreviewMetrics(preview, result)
		ep.logger.Debug("Metrics calculated", "psnr", finalMetrics["psnr"], "ssim", finalMetrics["ssim"])

		// Update UI immediately in the main thread
		ep.mu.RLock()
		callback := ep.onPreviewUpdate
		ep.mu.RUnlock()

		if callback != nil {
			fyne.Do(func() {
				ep.logger.Debug("Calling preview update callback in UI thread")
				// Pass the result directly - callback will handle Mat lifecycle
				callback(result, finalMetrics)
			})
		} else {
			result.Close()
		}

		result.Close()
		ep.logger.Debug("Preview processing completed successfully")
	}()
}

// processSequential applies sequential processing steps
func (ep *EnhancedPipeline) processSequential(ctx context.Context, input gocv.Mat) (gocv.Mat, map[string]float64) {
	current := input.Clone()
	processMetrics := make(map[string]float64)

	steps := ep.GetSteps()
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

		result, err := algorithms.Apply(step.Algorithm, current, step.Parameters)
		if err != nil {
			ep.logger.Error("Algorithm failed", "algorithm", step.Algorithm, "error", err)
			current.Close()
			return gocv.NewMat(), nil
		}

		stepMetrics := ep.metricsEval.EvaluateStep(current, result, step.Algorithm)
		for k, v := range stepMetrics {
			processMetrics[fmt.Sprintf("%s_%s", step.Algorithm, k)] = v
		}

		current.Close()
		current = result
	}

	return current, processMetrics
}

// calculatePreviewMetrics calculates quality metrics
func (ep *EnhancedPipeline) calculatePreviewMetrics(original, processed gocv.Mat) map[string]float64 {
	finalMetrics := make(map[string]float64)

	if psnr, err := ep.metricsEval.CalculatePSNR(original, processed); err == nil {
		finalMetrics["psnr"] = psnr
	}

	if ssim, err := ep.metricsEval.CalculateSSIM(original, processed); err == nil {
		finalMetrics["ssim"] = ssim
	}

	return finalMetrics
}

// ProcessFullResolution processes full resolution image
func (ep *EnhancedPipeline) ProcessFullResolution() (gocv.Mat, error) {
	if !ep.imageData.HasImage() {
		return gocv.NewMat(), fmt.Errorf("no image loaded")
	}

	original := ep.imageData.GetOriginal()
	defer original.Close()

	var result gocv.Mat
	var err error

	if ep.useLayerMode {
		result, err = ep.layerStack.ProcessLayers(original)
	} else {
		ctx := context.Background()
		result, _ = ep.processSequential(ctx, original)
	}

	if err != nil || result.Empty() {
		return gocv.NewMat(), fmt.Errorf("processing failed: %w", err)
	}

	ep.imageData.SetProcessed(result)
	return result.Clone(), nil
}

// GetSteps returns sequential processing steps
func (ep *EnhancedPipeline) GetSteps() []ProcessingStep {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	steps := make([]ProcessingStep, len(ep.steps))
	copy(steps, ep.steps)
	return steps
}

// GetLayers returns processing layers
func (ep *EnhancedPipeline) GetLayers() []*layers.Layer {
	return ep.layerStack.GetLayers()
}

// ClearAll clears both steps and layers
func (ep *EnhancedPipeline) ClearAll() {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.steps = make([]ProcessingStep, 0)
	ep.layerStack = layers.NewLayerStack(ep.regionManager)

	if ep.imageData.HasImage() {
		ep.imageData.ResetToOriginal()
		if ep.onPreviewUpdate != nil {
			original := ep.imageData.GetPreview()
			fyne.Do(func() {
				ep.onPreviewUpdate(original, nil)
			})
		}
	}
}

// SetRealtimeMode enables/disables real-time processing
func (ep *EnhancedPipeline) SetRealtimeMode(enabled bool) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.realtimeMode = enabled
}

// IsProcessing returns current processing state
func (ep *EnhancedPipeline) IsProcessing() bool {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return ep.processing
}

// Stop stops processing
func (ep *EnhancedPipeline) Stop() {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	if ep.cancel != nil {
		ep.cancel()
	}
	if ep.previewTimer != nil {
		ep.previewTimer.Stop()
	}
}
