// internal/core/pipeline.go
// Enhanced pipeline with proper debug integration - no compilation errors
package core

import (
	"context"
	"fmt"
	"image"
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

	// Callbacks - THREAD SAFE: only accept Go image, not GoCV Mat
	onPreviewUpdate func(preview image.Image, metrics map[string]float64)
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

	oldMode := ep.useLayerMode
	ep.useLayerMode = useLayerMode
	ep.logger.Info("PIPELINE: Processing mode changed",
		"old_mode", map[bool]string{true: "layer", false: "sequential"}[oldMode],
		"new_mode", map[bool]string{true: "layer", false: "sequential"}[useLayerMode])

	// DEBUG: Log mode change
	if GlobalPipelineDebugger != nil {
		layerCount := len(ep.layerStack.GetLayers())
		GlobalPipelineDebugger.LogModeChange(
			map[bool]string{true: "layer", false: "sequential"}[oldMode],
			map[bool]string{true: "layer", false: "sequential"}[useLayerMode],
			layerCount,
		)
	}

	if ep.realtimeMode {
		ep.logger.Debug("PIPELINE: Triggering preview processing due to mode change")
		ep.triggerPreviewProcessing()
	}
}

// AddLayer adds a processing layer (layer mode)
func (ep *EnhancedPipeline) AddLayer(name, algorithm string, params map[string]interface{}, regionID string) (string, error) {
	start := time.Now()
	ep.logger.Info("PIPELINE: Adding layer", "name", name, "algorithm", algorithm, "region_id", regionID)

	if !algorithms.IsValidAlgorithm(algorithm) {
		err := fmt.Errorf("unknown algorithm: %s", algorithm)
		ep.logger.Error("PIPELINE: Invalid algorithm", "algorithm", algorithm)

		// DEBUG: Log failed addition
		if GlobalPipelineDebugger != nil {
			duration := time.Since(start)
			GlobalPipelineDebugger.LogLayerAddition("", algorithm, params, false, duration, err)
		}
		return "", err
	}

	if err := algorithms.ValidateParameters(algorithm, params); err != nil {
		ep.logger.Error("PIPELINE: Invalid parameters", "algorithm", algorithm, "error", err)

		// DEBUG: Log failed addition
		if GlobalPipelineDebugger != nil {
			duration := time.Since(start)
			GlobalPipelineDebugger.LogLayerAddition("", algorithm, params, false, duration, err)
		}
		return "", fmt.Errorf("invalid parameters: %w", err)
	}

	layerID := ep.layerStack.AddLayer(name, algorithm, params, regionID)
	ep.logger.Info("PIPELINE: Layer added successfully", "layer_id", layerID, "algorithm", algorithm)

	// DEBUG: Log successful addition
	if GlobalPipelineDebugger != nil {
		duration := time.Since(start)
		GlobalPipelineDebugger.LogLayerAddition(layerID, algorithm, params, true, duration, nil)
	}

	// Trigger processing immediately when layer is added
	if ep.realtimeMode && ep.useLayerMode {
		ep.logger.Info("PIPELINE: Triggering immediate preview processing after adding layer")

		// DEBUG: Log preview trigger
		if GlobalPipelineDebugger != nil {
			layerCount := len(ep.layerStack.GetLayers())
			GlobalPipelineDebugger.LogPreviewTrigger("layer_added_to_pipeline", ep.imageData.HasImage(), layerCount)
		}

		ep.triggerPreviewProcessing()
	} else {
		ep.logger.Debug("PIPELINE: Not triggering preview", "realtime", ep.realtimeMode, "layer_mode", ep.useLayerMode)

		// DEBUG: Log why preview wasn't triggered
		if GlobalPipelineDebugger != nil {
			GlobalPipelineDebugger.LogEvent("preview_not_triggered",
				map[bool]string{true: "layer", false: "sequential"}[ep.useLayerMode],
				map[string]interface{}{
					"realtime_mode": ep.realtimeMode,
					"layer_mode":    ep.useLayerMode,
					"reason":        "mode_conditions_not_met",
				})
		}
	}

	return layerID, nil
}

// AddStep adds sequential processing step (sequential mode)
func (ep *EnhancedPipeline) AddStep(algorithm string, parameters map[string]interface{}) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()

	ep.logger.Info("PIPELINE: Adding sequential step", "algorithm", algorithm)

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
	ep.logger.Info("PIPELINE: Sequential step added", "algorithm", algorithm)

	if ep.realtimeMode && !ep.useLayerMode {
		ep.triggerPreviewProcessing()
	}

	return nil
}

// SetCallbacks sets preview update and error callbacks
func (ep *EnhancedPipeline) SetCallbacks(
	onPreviewUpdate func(image.Image, map[string]float64),
	onError func(error),
) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.onPreviewUpdate = onPreviewUpdate
	ep.onError = onError
	ep.logger.Debug("PIPELINE: Callbacks set")
}

// triggerPreviewProcessing starts debounced preview processing
func (ep *EnhancedPipeline) triggerPreviewProcessing() {
	if !ep.imageData.HasImage() {
		ep.logger.Debug("PIPELINE: No image available for preview processing")

		// DEBUG: Log no image
		if GlobalPipelineDebugger != nil {
			GlobalPipelineDebugger.LogPreviewTrigger("no_image", false, 0)
		}
		return
	}

	if ep.previewTimer != nil {
		ep.previewTimer.Stop()
	}

	ep.logger.Debug("PIPELINE: Scheduling preview processing", "delay_ms", ep.previewDelay.Milliseconds())

	// DEBUG: Log preview scheduling
	if GlobalPipelineDebugger != nil {
		layerCount := len(ep.layerStack.GetLayers())
		GlobalPipelineDebugger.LogEvent("preview_scheduled",
			map[bool]string{true: "layer", false: "sequential"}[ep.useLayerMode],
			map[string]interface{}{
				"delay_ms":    ep.previewDelay.Milliseconds(),
				"layer_count": layerCount,
			})
	}

	ep.previewTimer = time.AfterFunc(ep.previewDelay, func() {
		ep.processPreview()
	})
}

// processPreview processes preview image in real-time
func (ep *EnhancedPipeline) processPreview() {
	start := time.Now()
	ep.logger.Info("PIPELINE: Starting preview processing")

	// DEBUG: Log processing start
	if GlobalPipelineDebugger != nil {
		layerCount := len(ep.layerStack.GetLayers())
		mode := map[bool]string{true: "layer", false: "sequential"}[ep.useLayerMode]
		GlobalPipelineDebugger.LogProcessingStart(mode, layerCount, "")
		ep.logger.Debug("PIPELINE: Processing start logged", "duration_since_start", time.Since(start))
	}

	ep.mu.Lock()
	if ep.processing {
		ep.logger.Debug("PIPELINE: Already processing, cancelling previous")
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
			ep.logger.Debug("PIPELINE: Preview processing goroutine ended")
		}()

		preview := ep.imageData.GetPreview()
		defer preview.Close()

		if preview.Empty() {
			ep.logger.Error("PIPELINE: Preview image is empty")

			// DEBUG: Log empty preview error
			if GlobalPipelineDebugger != nil {
				GlobalPipelineDebugger.LogProcessingComplete("error", false, "empty_preview", nil)
				GlobalPipelineDebugger.LogError("Pipeline", "processPreview", "empty preview image")
			}

			if ep.onError != nil {
				fyne.Do(func() {
					ep.onError(fmt.Errorf("no preview image available"))
				})
			}
			return
		}

		inputSize := fmt.Sprintf("%dx%d", preview.Cols(), preview.Rows())
		ep.logger.Debug("PIPELINE: Preview image obtained", "size", inputSize)

		// Update debug with input size
		if GlobalPipelineDebugger != nil {
			layerCount := len(ep.layerStack.GetLayers())
			mode := map[bool]string{true: "layer", false: "sequential"}[ep.useLayerMode]
			GlobalPipelineDebugger.LogProcessingStart(mode, layerCount, inputSize)
		}

		var result gocv.Mat
		var err error

		// Process based on current mode
		ep.mu.RLock()
		useLayerMode := ep.useLayerMode
		ep.mu.RUnlock()

		mode := map[bool]string{true: "layer", false: "sequential"}[useLayerMode]
		ep.logger.Info("PIPELINE: Processing preview", "mode", mode)

		if useLayerMode {
			ep.logger.Debug("PIPELINE: Using layer processing")

			layerStart := time.Now()
			result, err = ep.layerStack.ProcessLayers(preview)
			layerDuration := time.Since(layerStart)

			if err != nil {
				ep.logger.Error("PIPELINE: Layer processing failed", "error", err)

				// DEBUG: Log layer processing failure
				if GlobalPipelineDebugger != nil {
					layerCount := len(ep.layerStack.GetLayers())
					GlobalPipelineDebugger.LogLayerProcessing(layerCount, inputSize, "", false, layerDuration, err)
					GlobalPipelineDebugger.LogProcessingComplete("layer", false, "processing_error", nil)
				}
			} else {
				outputSize := fmt.Sprintf("%dx%d", result.Cols(), result.Rows())
				ep.logger.Info("PIPELINE: Layer processing completed successfully",
					"input_size", inputSize,
					"output_size", outputSize,
					"output_empty", result.Empty())

				// DEBUG: Log layer processing success
				if GlobalPipelineDebugger != nil {
					layerCount := len(ep.layerStack.GetLayers())
					GlobalPipelineDebugger.LogLayerProcessing(layerCount, inputSize, outputSize, true, layerDuration, nil)
				}
			}
		} else {
			ep.logger.Debug("PIPELINE: Using sequential processing")
			result, _ = ep.processSequential(ctx, preview)
			ep.logger.Info("PIPELINE: Sequential processing completed")
		}

		if err != nil {
			ep.logger.Error("PIPELINE: Processing failed", "error", err)
			if !result.Empty() {
				result.Close()
			}

			// DEBUG: Log processing failure
			if GlobalPipelineDebugger != nil {
				GlobalPipelineDebugger.LogProcessingComplete(mode, false, "processing_failed", nil)
			}

			if ep.onError != nil {
				fyne.Do(func() {
					ep.onError(fmt.Errorf("processing failed: %w", err))
				})
			}
			return
		}

		if result.Empty() {
			ep.logger.Error("PIPELINE: Processing returned empty result")
			result.Close()

			// DEBUG: Log empty result
			if GlobalPipelineDebugger != nil {
				GlobalPipelineDebugger.LogProcessingComplete(mode, false, "empty_result", nil)
			}

			if ep.onError != nil {
				fyne.Do(func() {
					ep.onError(fmt.Errorf("processing returned empty result"))
				})
			}
			return
		}

		ep.logger.Info("PIPELINE: Processing successful, calculating metrics")

		// Calculate metrics BEFORE converting to image
		metricsStart := time.Now()
		finalMetrics := ep.calculatePreviewMetrics(preview, result)
		metricsDuration := time.Since(metricsStart)

		ep.logger.Info("PIPELINE: Metrics calculated", "psnr", finalMetrics["psnr"], "ssim", finalMetrics["ssim"])

		// DEBUG: Log metrics calculation
		if GlobalPipelineDebugger != nil {
			psnr, _ := finalMetrics["psnr"]
			ssim, _ := finalMetrics["ssim"]
			GlobalPipelineDebugger.LogMetricsCalculation(psnr, ssim, metricsDuration, nil)
		}

		// Convert Mat to image.Image in THIS goroutine
		ep.logger.Debug("PIPELINE: Converting Mat to image for UI callback")
		conversionStart := time.Now()
		previewImage, err := result.ToImage()
		conversionDuration := time.Since(conversionStart)
		result.Close() // Close Mat immediately after conversion

		if err != nil {
			ep.logger.Error("PIPELINE: Failed to convert Mat to image", "error", err)

			// DEBUG: Log conversion failure
			if GlobalPipelineDebugger != nil {
				GlobalPipelineDebugger.LogMatConversion(inputSize, "", false, conversionDuration, err)
				GlobalPipelineDebugger.LogProcessingComplete(mode, false, "conversion_failed", nil)
			}

			if ep.onError != nil {
				fyne.Do(func() {
					ep.onError(fmt.Errorf("failed to convert preview: %w", err))
				})
			}
			return
		}

		outputSize := previewImage.Bounds().Size().String()
		ep.logger.Info("PIPELINE: Successfully converted to image", "bounds", previewImage.Bounds())

		// DEBUG: Log successful conversion
		if GlobalPipelineDebugger != nil {
			GlobalPipelineDebugger.LogMatConversion(inputSize, outputSize, true, conversionDuration, nil)
		}

		// DEBUG: Log successful completion
		if GlobalPipelineDebugger != nil {
			GlobalPipelineDebugger.LogProcessingComplete(mode, true, outputSize, finalMetrics)
		}

		// Update UI with Go image (thread-safe)
		ep.mu.RLock()
		callback := ep.onPreviewUpdate
		ep.mu.RUnlock()

		if callback != nil {
			ep.logger.Debug("PIPELINE: Calling UI callback with converted image")
			fyne.Do(func() {
				ep.logger.Info("PIPELINE: Executing preview update callback in UI thread")
				callback(previewImage, finalMetrics)
			})
		} else {
			ep.logger.Warn("PIPELINE: No preview update callback set")
		}

		ep.logger.Info("PIPELINE: Preview processing completed successfully")
	}()
}

// processSequential applies sequential processing steps
func (ep *EnhancedPipeline) processSequential(ctx context.Context, input gocv.Mat) (gocv.Mat, map[string]float64) {
	current := input.Clone()
	processMetrics := make(map[string]float64)

	steps := ep.GetSteps()
	ep.logger.Debug("PIPELINE: Processing sequential steps", "step_count", len(steps))

	for i, step := range steps {
		select {
		case <-ctx.Done():
			ep.logger.Debug("PIPELINE: Sequential processing cancelled")
			current.Close()
			return gocv.NewMat(), nil
		default:
		}

		if !step.Enabled {
			ep.logger.Debug("PIPELINE: Skipping disabled step", "step", i, "algorithm", step.Algorithm)
			continue
		}

		ep.logger.Debug("PIPELINE: Processing step", "step", i, "algorithm", step.Algorithm)

		result, err := algorithms.Apply(step.Algorithm, current, step.Parameters)
		if err != nil {
			ep.logger.Error("PIPELINE: Sequential step failed", "step", i, "algorithm", step.Algorithm, "error", err)
			current.Close()
			return gocv.NewMat(), nil
		}

		stepMetrics := ep.metricsEval.EvaluateStep(current, result, step.Algorithm)
		for k, v := range stepMetrics {
			processMetrics[fmt.Sprintf("%s_%s", step.Algorithm, k)] = v
		}

		current.Close()
		current = result
		ep.logger.Debug("PIPELINE: Step completed", "step", i, "algorithm", step.Algorithm)
	}

	return current, processMetrics
}

// calculatePreviewMetrics calculates quality metrics
func (ep *EnhancedPipeline) calculatePreviewMetrics(original, processed gocv.Mat) map[string]float64 {
	finalMetrics := make(map[string]float64)

	if original.Empty() || processed.Empty() {
		ep.logger.Error("PIPELINE: Cannot calculate metrics with empty Mats",
			"original_empty", original.Empty(),
			"processed_empty", processed.Empty())
		return finalMetrics
	}

	if psnr, err := ep.metricsEval.CalculatePSNR(original, processed); err == nil {
		finalMetrics["psnr"] = psnr
		ep.logger.Debug("PIPELINE: PSNR calculated", "value", psnr)
	} else {
		ep.logger.Error("PIPELINE: Failed to calculate PSNR", "error", err)
	}

	if ssim, err := ep.metricsEval.CalculateSSIM(original, processed); err == nil {
		finalMetrics["ssim"] = ssim
		ep.logger.Debug("PIPELINE: SSIM calculated", "value", ssim)
	} else {
		ep.logger.Error("PIPELINE: Failed to calculate SSIM", "error", err)
	}

	return finalMetrics
}

// ProcessFullResolution processes full resolution image
func (ep *EnhancedPipeline) ProcessFullResolution() (gocv.Mat, error) {
	ep.logger.Info("PIPELINE: Processing full resolution image")

	if !ep.imageData.HasImage() {
		return gocv.NewMat(), fmt.Errorf("no image loaded")
	}

	original := ep.imageData.GetOriginal()
	defer original.Close()

	var result gocv.Mat
	var err error

	if ep.useLayerMode {
		ep.logger.Debug("PIPELINE: Full resolution using layer mode")
		result, err = ep.layerStack.ProcessLayers(original)
	} else {
		ep.logger.Debug("PIPELINE: Full resolution using sequential mode")
		ctx := context.Background()
		result, _ = ep.processSequential(ctx, original)
	}

	if err != nil || result.Empty() {
		ep.logger.Error("PIPELINE: Full resolution processing failed", "error", err)
		return gocv.NewMat(), fmt.Errorf("processing failed: %w", err)
	}

	ep.imageData.SetProcessed(result)
	ep.logger.Info("PIPELINE: Full resolution processing completed")
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

	ep.logger.Info("PIPELINE: Clearing all processing steps and layers")

	// DEBUG: Log clear operation
	if GlobalPipelineDebugger != nil {
		GlobalPipelineDebugger.LogEvent("clear_all", "none", map[string]interface{}{
			"previous_layer_count": len(ep.layerStack.GetLayers()),
			"previous_step_count":  len(ep.steps),
		})
	}

	ep.steps = make([]ProcessingStep, 0)
	ep.layerStack = layers.NewLayerStack(ep.regionManager)

	if ep.imageData.HasImage() {
		ep.imageData.ResetToOriginal()
		if ep.onPreviewUpdate != nil {
			original := ep.imageData.GetPreview()
			if !original.Empty() {
				if img, err := original.ToImage(); err == nil {
					fyne.Do(func() {
						ep.onPreviewUpdate(img, nil)
					})
				}
			}
			original.Close()
		}
	}
}

// SetRealtimeMode enables/disables real-time processing
func (ep *EnhancedPipeline) SetRealtimeMode(enabled bool) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	ep.realtimeMode = enabled
	ep.logger.Debug("PIPELINE: Realtime mode changed", "enabled", enabled)
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

	ep.logger.Debug("PIPELINE: Stopping processing")
	if ep.cancel != nil {
		ep.cancel()
	}
	if ep.previewTimer != nil {
		ep.previewTimer.Stop()
	}
}
