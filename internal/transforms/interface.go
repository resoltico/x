// Author: Ervins Strauhmanis
// License: MIT

package transforms

import (
	"gocv.io/x/gocv"
)

// Transform defines the interface that all image transformations must implement
type Transform interface {
	// Apply applies the transformation to the input matrix with given parameters
	Apply(input gocv.Mat, params map[string]interface{}) (gocv.Mat, error)
	
	// GetDefaultParams returns the default parameters for this transformation
	GetDefaultParams() map[string]interface{}
	
	// GetName returns the name of this transformation
	GetName() string
	
	// GetDescription returns a description of what this transformation does
	GetDescription() string
	
	// Validate validates the parameters for this transformation
	Validate(params map[string]interface{}) error
}

// ParameterInfo provides metadata about a parameter
type ParameterInfo struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`        // "int", "float", "bool", "string", "enum"
	Min         interface{} `json:"min,omitempty"`
	Max         interface{} `json:"max,omitempty"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Options     []string    `json:"options,omitempty"` // For enum type
}

// TransformInfo provides metadata about a transformation
type TransformInfo interface {
	// GetParameterInfo returns information about all parameters
	GetParameterInfo() []ParameterInfo
}

// TransformRegistry manages all available transformations
type TransformRegistry struct {
	transforms map[string]Transform
}

// NewTransformRegistry creates a new transform registry
func NewTransformRegistry() *TransformRegistry {
	return &TransformRegistry{
		transforms: make(map[string]Transform),
	}
}

// Register registers a transformation
func (tr *TransformRegistry) Register(name string, transform Transform) {
	tr.transforms[name] = transform
}

// Get retrieves a transformation by name
func (tr *TransformRegistry) Get(name string) (Transform, bool) {
	transform, exists := tr.transforms[name]
	return transform, exists
}

// GetAll returns all registered transformations
func (tr *TransformRegistry) GetAll() map[string]Transform {
	result := make(map[string]Transform)
	for name, transform := range tr.transforms {
		result[name] = transform
	}
	return result
}

// GetByCategory returns transformations grouped by category
func (tr *TransformRegistry) GetByCategory() map[string][]string {
	categories := map[string][]string{
		"Binarization":       {"otsu", "niblack", "sauvola"},
		"Morphology":         {"erosion", "dilation", "opening", "closing"},
		"Noise Reduction":    {"gaussian", "median", "bilateral"},
		"Scaling":            {"bilinear", "bicubic", "lanczos"},
		"Color Manipulation": {"grayscale", "overlay"},
	}
	return categories
}