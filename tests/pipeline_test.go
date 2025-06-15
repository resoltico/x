// Author: Ervins Strauhmanis
// License: MIT

package tests

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocv.io/x/gocv"
	"github.com/sirupsen/logrus"

	"advanced-image-processing/internal/image_processing"
	"advanced-image-processing/internal/models"
	"advanced-image-processing/internal/transforms"
	"advanced-image-processing/internal/transforms/binarization"
	"advanced-image-processing/internal/transforms/morphology"
	"advanced-image-processing/internal/transforms/noise_reduction"
)

func setupPipelineTest() (*transforms.TransformRegistry, *models.ImageData, *image_processing.Pipeline) {
	// Create registry and register transforms
	registry := transforms.NewTransformRegistry()
	registry.Register("otsu", binarization.NewOtsuTransform())
	registry.Register("gaussian", noise_reduction.NewGaussianTransform())
	registry.Register("erosion", morphology.NewErosionTransform())
	
	// Create image data and load test image
	imageData := models.NewImageData()
	testImage := createTestImage()
	imageData.SetOriginal(testImage)
	testImage.Close()
	
	// Create logger
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests
	
	// Create pipeline
	pipeline := image_processing.NewPipeline(registry, imageData, logger)
	
	return registry, imageData, pipeline
}

func TestPipelineBasicFunctionality(t *testing.T) {
	_, imageData, pipeline := setupPipelineTest()
	defer imageData.Clear()
	
	t.Run("Add single transformation", func(t *testing.T) {
		params := map[string]interface{}{
			"max_value": 255.0,
		}
		
		err := pipeline.AddTransformation("otsu", params)
		assert.NoError(t, err)
		
		sequence := pipeline.GetTransformationSequence()
		assert.Equal(t, 1, sequence.Length())
		
		steps := sequence.GetSteps()
		assert.Equal(t, "otsu", steps[0].Type)
		assert.True(t, steps[0].Enabled)
	})
	
	t.Run("Add multiple transformations", func(t *testing.T) {
		pipeline.ClearSequence()
		
		// Add Gaussian blur
		err := pipeline.AddTransformation("gaussian", map[string]interface{}{
			"kernel_size": 5.0,
			"sigma_x":     1.0,
			"sigma_y":     1.0,
		})
		assert.NoError(t, err)
		
		// Add Otsu binarization
		err = pipeline.AddTransformation("otsu", map[string]interface{}{
			"max_value": 255.0,
		})
		assert.NoError(t, err)
		
		sequence := pipeline.GetTransformationSequence()
		assert.Equal(t, 2, sequence.Length())
	})
	
	t.Run("Remove transformation", func(t *testing.T) {
		initialCount := pipeline.GetTransformationSequence().Length()
		
		err := pipeline.RemoveTransformation(0)
		assert.NoError(t, err)
		
		newCount := pipeline.GetTransformationSequence().Length()
		assert.Equal(t, initialCount-1, newCount)
	})
	
	t.Run("Update transformation parameters", func(t *testing.T) {
		pipeline.ClearSequence()
		
		// Add a transformation
		err := pipeline.AddTransformation("gaussian", map[string]interface{}{
			"kernel_size": 5.0,
			"sigma_x":     1.0,
			"sigma_y":     1.0,
		})
		assert.NoError(t, err)
		
		// Update parameters
		newParams := map[string]interface{}{
			"kernel_size": 7.0,
			"sigma_x":     2.0,
			"sigma_y":     2.0,
		}
		
		err = pipeline.UpdateTransformation(0, newParams)
		assert.NoError(t, err)
		
		// Verify update
		steps := pipeline.GetTransformationSequence().GetSteps()
		assert.Equal(t, 7.0, steps[0].Parameters["kernel_size"])
		assert.Equal(t, 2.0, steps[0].Parameters["sigma_x"])
	})
}

func TestPipelineProcessing(t *testing.T) {
	_, imageData, pipeline := setupPipelineTest()
	defer imageData.Clear()
	
	t.Run("Processing with callbacks", func(t *testing.T) {
		var (
			progressCalled bool
			completeCalled bool
			errorCalled    bool
			resultMat      gocv.Mat
			mu             sync.Mutex
		)
		
		// Set up callbacks
		pipeline.SetCallbacks(
			func(step, total int, stepName string) {
				mu.Lock()
				progressCalled = true
				mu.Unlock()
			},
			func(result gocv.Mat) {
				mu.Lock()
				completeCalled = true
				resultMat = result.Clone()
				mu.Unlock()
			},
			func(err error) {
				mu.Lock()
				errorCalled = true
				mu.Unlock()
			},
		)
		
		// Add a transformation
		err := pipeline.AddTransformation("otsu", map[string]interface{}{
			"max_value": 255.0,
		})
		assert.NoError(t, err)
		
		// Wait for processing to complete
		time.Sleep(500 * time.Millisecond)
		
		mu.Lock()
		assert.True(t, progressCalled, "Progress callback should be called")
		assert.True(t, completeCalled, "Complete callback should be called")
		assert.False(t, errorCalled, "Error callback should not be called")
		assert.False(t, resultMat.Empty(), "Result should not be empty")
		mu.Unlock()
		
		if !resultMat.Empty() {
			resultMat.Close()
		}
	})
	
	t.Run("Processing with disabled transformation", func(t *testing.T) {
		pipeline.ClearSequence()
		
		// Add transformation
		err := pipeline.AddTransformation("otsu", map[string]interface{}{
			"max_value": 255.0,
		})
		assert.NoError(t, err)
		
		// Disable the transformation
		sequence := pipeline.GetTransformationSequence()
		sequence.ToggleStep(0, false)
		
		// Wait for processing
		time.Sleep(300 * time.Millisecond)
		
		// The result should be the original image (no processing applied)
		processed := imageData.GetProcessed()
		original := imageData.GetOriginal()
		defer processed.Close()
		defer original.Close()
		
		// Note: This is a simplified check - in practice you might compare image data
		assert.Equal(t, original.Rows(), processed.Rows())
		assert.Equal(t, original.Cols(), processed.Cols())
	})
}

func TestPipelineErrorHandling(t *testing.T) {
	_, imageData, pipeline := setupPipelineTest()
	defer imageData.Clear()
	
	t.Run("Unknown transformation", func(t *testing.T) {
		err := pipeline.AddTransformation("unknown", map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown transformation type")
	})
	
	t.Run("Invalid parameters", func(t *testing.T) {
		// Try to add with invalid parameters
		err := pipeline.AddTransformation("otsu", map[string]interface{}{
			"max_value": 300.0, // Out of valid range
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid parameters")
	})
	
	t.Run("Remove invalid index", func(t *testing.T) {
		err := pipeline.RemoveTransformation(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transformation index")
	})
	
	t.Run("Update invalid index", func(t *testing.T) {
		err := pipeline.UpdateTransformation(999, map[string]interface{}{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid transformation index")
	})
}

func TestPipelineSequenceManagement(t *testing.T) {
	_, imageData, pipeline := setupPipelineTest()
	defer imageData.Clear()
	
	t.Run("Clear sequence", func(t *testing.T) {
		// Add some transformations
		pipeline.AddTransformation("gaussian", map[string]interface{}{
			"kernel_size": 5.0,
			"sigma_x":     1.0,
			"sigma_y":     1.0,
		})
		pipeline.AddTransformation("otsu", map[string]interface{}{
			"max_value": 255.0,
		})
		
		assert.Equal(t, 2, pipeline.GetTransformationSequence().Length())
		
		// Clear sequence
		pipeline.ClearSequence()
		
		assert.Equal(t, 0, pipeline.GetTransformationSequence().Length())
	})
	
	t.Run("Load sequence", func(t *testing.T) {
		// Create a new sequence
		newSequence := models.NewTransformationSequence()
		newSequence.AddStep("gaussian", map[string]interface{}{
			"kernel_size": 7.0,
			"sigma_x":     1.5,
			"sigma_y":     1.5,
		})
		newSequence.AddStep("erosion", map[string]interface{}{
			"kernel_size": 3.0,
			"iterations":  1.0,
		})
		
		// Load the sequence
		pipeline.LoadSequence(newSequence)
		
		// Verify
		loadedSequence := pipeline.GetTransformationSequence()
		assert.Equal(t, 2, loadedSequence.Length())
		
		steps := loadedSequence.GetSteps()
		assert.Equal(t, "gaussian", steps[0].Type)
		assert.Equal(t, "erosion", steps[1].Type)
	})
}

func TestPipelineProcessingState(t *testing.T) {
	_, imageData, pipeline := setupPipelineTest()
	defer imageData.Clear()
	
	t.Run("Processing state management", func(t *testing.T) {
		// Initially not processing
		assert.False(t, pipeline.IsProcessing())
		
		// Add transformation (triggers processing)
		pipeline.AddTransformation("otsu", map[string]interface{}{
			"max_value": 255.0,
		})
		
		// Might be processing briefly
		// Note: This test is timing-dependent and might be flaky
		// In a real scenario, you might want to use more sophisticated synchronization
		
		// Wait for processing to complete
		time.Sleep(100 * time.Millisecond)
		
		// Should not be processing anymore
		assert.False(t, pipeline.IsProcessing())
	})
	
	t.Run("Stop processing", func(t *testing.T) {
		// Add transformation
		pipeline.AddTransformation("gaussian", map[string]interface{}{
			"kernel_size": 5.0,
			"sigma_x":     1.0,
			"sigma_y":     1.0,
		})
		
		// Stop processing
		pipeline.Stop()
		
		// Should not be processing
		assert.False(t, pipeline.IsProcessing())
	})
}

// Test complex pipeline scenarios
func TestComplexPipelineScenarios(t *testing.T) {
	_, imageData, pipeline := setupPipelineTest()
	defer imageData.Clear()
	
	t.Run("Complex processing sequence", func(t *testing.T) {
		var (
			completionCount int
			mu              sync.Mutex
		)
		
		// Set up completion callback
		pipeline.SetCallbacks(
			nil, // progress
			func(result gocv.Mat) {
				mu.Lock()
				completionCount++
				mu.Unlock()
				result.Close()
			},
			nil, // error
		)
		
		// Add multiple transformations
		transformations := []struct {
			name   string
			params map[string]interface{}
		}{
			{"gaussian", map[string]interface{}{
				"kernel_size": 5.0,
				"sigma_x":     1.0,
				"sigma_y":     1.0,
			}},
			{"otsu", map[string]interface{}{
				"max_value": 255.0,
			}},
			{"erosion", map[string]interface{}{
				"kernel_size": 3.0,
				"iterations":  1.0,
			}},
		}
		
		for _, transform := range transformations {
			err := pipeline.AddTransformation(transform.name, transform.params)
			assert.NoError(t, err)
		}
		
		// Wait for all processing to complete
		time.Sleep(1 * time.Second)
		
		mu.Lock()
		// Should have completed processing (at least once)
		assert.True(t, completionCount > 0, "Should have completed processing")
		mu.Unlock()
		
		// Verify final sequence
		sequence := pipeline.GetTransformationSequence()
		assert.Equal(t, 3, sequence.Length())
		
		// Check processed image exists
		processed := imageData.GetProcessed()
		assert.False(t, processed.Empty())
		processed.Close()
	})
}

// Benchmark pipeline performance
func BenchmarkPipelineProcessing(b *testing.B) {
	_, imageData, pipeline := setupPipelineTest()
	defer imageData.Clear()
	
	// Set up a simple pipeline
	pipeline.AddTransformation("gaussian", map[string]interface{}{
		"kernel_size": 5.0,
		"sigma_x":     1.0,
		"sigma_y":     1.0,
	})
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Update parameters to trigger processing
		pipeline.UpdateTransformation(0, map[string]interface{}{
			"kernel_size": 5.0,
			"sigma_x":     float64(i%5 + 1),
			"sigma_y":     1.0,
		})
		
		// Wait for processing
		time.Sleep(10 * time.Millisecond)
	}
}