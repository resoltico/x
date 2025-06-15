// Algorithm system with pluggable implementations
package algorithms

import (
	"fmt"

	"gocv.io/x/gocv"
)

// Algorithm defines the interface for image processing algorithms
type Algorithm interface {
	// Apply applies the algorithm to the input image
	Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error)
	
	// GetDefaultParams returns default parameters
	GetDefaultParams() map[string]interface{}
	
	// GetName returns the algorithm name
	GetName() string
	
	// GetDescription returns the algorithm description
	GetDescription() string
	
	// Validate validates parameters
	Validate(params map[string]interface{}) error
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

// AlgorithmInfo provides metadata about algorithms
type AlgorithmInfo interface {
	GetParameterInfo() []ParameterInfo
}

// Algorithm registry
var algorithms = make(map[string]Algorithm)

// Register registers an algorithm
func Register(name string, algorithm Algorithm) {
	algorithms[name] = algorithm
}

// Get retrieves an algorithm by name
func Get(name string) (Algorithm, bool) {
	algorithm, exists := algorithms[name]
	return algorithm, exists
}

// Apply applies an algorithm by name
func Apply(name string, input gocv.Mat, params map[string]interface{}) (gocv.Mat, error) {
	algorithm, exists := algorithms[name]
	if !exists {
		return gocv.NewMat(), fmt.Errorf("algorithm not found: %s", name)
	}
	
	return algorithm.Apply(input, params)
}

// ValidateParameters validates parameters for an algorithm
func ValidateParameters(name string, params map[string]interface{}) error {
	algorithm, exists := algorithms[name]
	if !exists {
		return fmt.Errorf("algorithm not found: %s", name)
	}
	
	return algorithm.Validate(params)
}

// IsValidAlgorithm checks if an algorithm exists
func IsValidAlgorithm(name string) bool {
	_, exists := algorithms[name]
	return exists
}

// GetAllAlgorithms returns all registered algorithms
func GetAllAlgorithms() map[string]Algorithm {
	result := make(map[string]Algorithm)
	for name, algorithm := range algorithms {
		result[name] = algorithm
	}
	return result
}

// GetAlgorithmsByCategory returns algorithms grouped by category
func GetAlgorithmsByCategory() map[string][]string {
	return map[string][]string{
		"Binarization": {
			"otsu_multi",
			"otsu_local", 
			"niblack_true",
			"sauvola_true",
			"wolf_jolion",
			"nick",
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

// init registers all algorithms
func init() {
	// Register binarization algorithms
	Register("otsu_multi", NewMultiOtsu())
	Register("otsu_local", NewLocalOtsu())
	Register("niblack_true", NewTrueNiblack())
	Register("sauvola_true", NewTrueSauvola())
	Register("wolf_jolion", NewWolfJolion())
	Register("nick", NewNICK())
	
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
