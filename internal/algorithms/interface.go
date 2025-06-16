// Optimized algorithm system using standard APIs
package algorithms

import (
	"fmt"

	"gocv.io/x/gocv"
)

// Algorithm defines the interface for image processing algorithms
type Algorithm interface {
	Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error)
	GetDefaultParams() map[string]interface{}
	GetName() string
	GetDescription() string
	Validate(params map[string]interface{}) error
	GetParameterInfo() []ParameterInfo
}

// ParameterInfo describes a parameter for UI generation
type ParameterInfo struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"` // "int", "float", "bool", "string", "enum"
	Min         interface{} `json:"min,omitempty"`
	Max         interface{} `json:"max,omitempty"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Options     []string    `json:"options,omitempty"` // For enum type
}

var algorithms = make(map[string]Algorithm)

func Register(name string, algorithm Algorithm) {
	algorithms[name] = algorithm
}

func Get(name string) (Algorithm, bool) {
	algorithm, exists := algorithms[name]
	return algorithm, exists
}

func Apply(name string, input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	algorithm, exists := algorithms[name]
	if !exists {
		return gocv.NewMat(), fmt.Errorf("algorithm not found: %s", name)
	}

	return algorithm.Apply(input, params)
}

func ValidateParameters(name string, params map[string]interface{}) error {
	algorithm, exists := algorithms[name]
	if !exists {
		return fmt.Errorf("algorithm not found: %s", name)
	}

	return algorithm.Validate(params)
}

func IsValidAlgorithm(name string) bool {
	_, exists := algorithms[name]
	return exists
}

func GetAllAlgorithms() map[string]Algorithm {
	result := make(map[string]Algorithm)
	for name, algorithm := range algorithms {
		result[name] = algorithm
	}
	return result
}

func GetAlgorithmsByCategory() map[string][]string {
	return map[string][]string{
		"Binarization": {
			"adaptive_threshold",
			"otsu_multi",
		},
		"Morphology": {
			"erosion",
			"dilation",
			"opening",
			"closing",
		},
		"Filters": {
			"gaussian",
			"median",
			"bilateral",
		},
	}
}

func init() {
	// Register optimized algorithms
	Register("adaptive_threshold", NewAdaptiveThreshold())
	Register("otsu_multi", NewMultiOtsu())

	// Register morphological algorithms
	Register("erosion", NewErosion())
	Register("dilation", NewDilation())
	Register("opening", NewOpening())
	Register("closing", NewClosing())

	// Register filter algorithms
	Register("gaussian", NewGaussianFilter())
	Register("median", NewMedianFilter())
	Register("bilateral", NewBilateralFilter())
}
