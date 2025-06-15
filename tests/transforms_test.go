// Author: Ervins Strauhmanis
// License: MIT

package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gocv.io/x/gocv"

	"advanced-image-processing/internal/transforms"
	"advanced-image-processing/internal/transforms/binarization"
	"advanced-image-processing/internal/transforms/color_manipulation"
	"advanced-image-processing/internal/transforms/morphology"
	"advanced-image-processing/internal/transforms/noise_reduction"
)

// createTestImage creates a test image for testing
func createTestImage() gocv.Mat {
	// Create a simple 100x100 grayscale test image
	img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC1)

	// Fill with gradient pattern
	for i := 0; i < img.Rows(); i++ {
		for j := 0; j < img.Cols(); j++ {
			value := uint8((i + j) % 256)
			img.SetUCharAt(i, j, value)
		}
	}

	return img
}

// createColorTestImage creates a color test image
func createColorTestImage() gocv.Mat {
	// Create a simple 100x100 color test image
	img := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8UC3)

	// Fill with color pattern
	for i := 0; i < img.Rows(); i++ {
		for j := 0; j < img.Cols(); j++ {
			b := uint8(i % 256)
			g := uint8(j % 256)
			r := uint8((i + j) % 256)
			img.SetVecbAt(i, j, gocv.Vecb{b, g, r})
		}
	}

	return img
}

func TestOtsuTransform(t *testing.T) {
	transform := binarization.NewOtsuTransform()

	t.Run("Basic functionality", func(t *testing.T) {
		input := createTestImage()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
		assert.Equal(t, input.Rows(), result.Rows())
		assert.Equal(t, input.Cols(), result.Cols())
	})

	t.Run("Color image conversion", func(t *testing.T) {
		input := createColorTestImage()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
		assert.Equal(t, 1, result.Channels()) // Should be grayscale
	})

	t.Run("Empty image handling", func(t *testing.T) {
		input := gocv.NewMat()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		assert.Error(t, err)
		assert.True(t, result.Empty())
	})

	t.Run("Parameter validation", func(t *testing.T) {
		// Test valid parameters
		validParams := map[string]interface{}{
			"max_value": 200.0,
		}
		err := transform.Validate(validParams)
		assert.NoError(t, err)

		// Test invalid parameters
		invalidParams := map[string]interface{}{
			"max_value": 300.0, // Out of range
		}
		err = transform.Validate(invalidParams)
		assert.Error(t, err)
	})

	t.Run("Interface compliance", func(t *testing.T) {
		assert.Equal(t, "Otsu Binarization", transform.GetName())
		assert.NotEmpty(t, transform.GetDescription())

		params := transform.GetDefaultParams()
		assert.Contains(t, params, "max_value")
	})
}

func TestNiblackTransform(t *testing.T) {
	transform := binarization.NewNiblackTransform()

	t.Run("Basic functionality", func(t *testing.T) {
		input := createTestImage()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
		assert.Equal(t, input.Rows(), result.Rows())
		assert.Equal(t, input.Cols(), result.Cols())
	})

	t.Run("Parameter validation", func(t *testing.T) {
		// Test block size adjustment (even to odd)
		params := map[string]interface{}{
			"block_size": 10.0, // Even number
		}

		input := createTestImage()
		defer input.Close()

		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
	})
}

func TestSauvolaTransform(t *testing.T) {
	transform := binarization.NewSauvolaTransform()

	t.Run("Basic functionality", func(t *testing.T) {
		input := createTestImage()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
	})

	t.Run("Default parameters", func(t *testing.T) {
		params := transform.GetDefaultParams()
		assert.Contains(t, params, "max_value")
		assert.Contains(t, params, "block_size")
		assert.Contains(t, params, "c")

		assert.Equal(t, 255.0, params["max_value"])
		assert.Equal(t, 15.0, params["block_size"])
		assert.Equal(t, 10.0, params["c"])
	})
}

func TestErosionTransform(t *testing.T) {
	transform := morphology.NewErosionTransform()

	t.Run("Basic functionality", func(t *testing.T) {
		input := createTestImage()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
		assert.Equal(t, input.Rows(), result.Rows())
		assert.Equal(t, input.Cols(), result.Cols())
	})

	t.Run("Multiple iterations", func(t *testing.T) {
		input := createTestImage()
		defer input.Close()

		params := map[string]interface{}{
			"kernel_size": 3.0,
			"iterations":  3.0,
		}

		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
	})

	t.Run("Parameter validation", func(t *testing.T) {
		// Test valid parameters
		validParams := map[string]interface{}{
			"kernel_size": 5.0,
			"iterations":  2.0,
		}
		err := transform.Validate(validParams)
		assert.NoError(t, err)

		// Test invalid kernel size
		invalidParams := map[string]interface{}{
			"kernel_size": 50.0, // Too large
		}
		err = transform.Validate(invalidParams)
		assert.Error(t, err)
	})
}

func TestGaussianTransform(t *testing.T) {
	transform := noise_reduction.NewGaussianTransform()

	t.Run("Basic functionality", func(t *testing.T) {
		input := createTestImage()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
		assert.Equal(t, input.Rows(), result.Rows())
		assert.Equal(t, input.Cols(), result.Cols())
	})

	t.Run("Color image processing", func(t *testing.T) {
		input := createColorTestImage()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
		assert.Equal(t, 3, result.Channels()) // Should maintain color channels
	})
}

func TestGrayscaleTransform(t *testing.T) {
	transform := color_manipulation.NewGrayscaleTransform()

	t.Run("Color to grayscale conversion", func(t *testing.T) {
		input := createColorTestImage()
		defer input.Close()

		params := transform.GetDefaultParams()
		result, err := transform.Apply(input, params)
		defer result.Close()

		require.NoError(t, err)
		assert.False(t, result.Empty())
		assert.Equal(t, 1, result.Channels()) // Should be grayscale
		assert.Equal(t, input.Rows(), result.Rows())
		assert.Equal(t, input.Cols(), result.Cols())
	})
}

// Test transform registry
func TestTransformRegistry(t *testing.T) {
	registry := transforms.NewTransformRegistry()

	// Register test transforms
	registry.Register("otsu", binarization.NewOtsuTransform())
	registry.Register("gaussian", noise_reduction.NewGaussianTransform())

	t.Run("Get registered transform", func(t *testing.T) {
		transform, exists := registry.Get("otsu")
		assert.True(t, exists)
		assert.NotNil(t, transform)
		assert.Equal(t, "Otsu Binarization", transform.GetName())
	})

	t.Run("Get non-existent transform", func(t *testing.T) {
		_, exists := registry.Get("nonexistent")
		assert.False(t, exists)
	})

	t.Run("Get all transforms", func(t *testing.T) {
		all := registry.GetAll()
		assert.Len(t, all, 2)
		assert.Contains(t, all, "otsu")
		assert.Contains(t, all, "gaussian")
	})
}
