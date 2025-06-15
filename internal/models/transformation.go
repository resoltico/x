// Author: Ervins Strauhmanis
// License: MIT

package models

import (
	"encoding/json"
	"sync"
)

// TransformationStep represents a single transformation with its parameters
type TransformationStep struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
}

// TransformationSequence represents a sequence of transformations
type TransformationSequence struct {
	mu    sync.RWMutex
	steps []TransformationStep
}

// NewTransformationSequence creates a new transformation sequence
func NewTransformationSequence() *TransformationSequence {
	return &TransformationSequence{
		steps: make([]TransformationStep, 0),
	}
}

// AddStep adds a transformation step to the sequence
func (ts *TransformationSequence) AddStep(transformType string, params map[string]interface{}) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	step := TransformationStep{
		Type:       transformType,
		Parameters: params,
		Enabled:    true,
	}
	ts.steps = append(ts.steps, step)
}

// RemoveStep removes a transformation step at the given index
func (ts *TransformationSequence) RemoveStep(index int) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if index >= 0 && index < len(ts.steps) {
		ts.steps = append(ts.steps[:index], ts.steps[index+1:]...)
	}
}

// GetSteps returns a copy of all transformation steps
func (ts *TransformationSequence) GetSteps() []TransformationStep {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	steps := make([]TransformationStep, len(ts.steps))
	copy(steps, ts.steps)
	return steps
}

// UpdateStep updates a transformation step at the given index
func (ts *TransformationSequence) UpdateStep(index int, params map[string]interface{}) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if index >= 0 && index < len(ts.steps) {
		ts.steps[index].Parameters = params
	}
}

// ToggleStep enables or disables a transformation step
func (ts *TransformationSequence) ToggleStep(index int, enabled bool) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if index >= 0 && index < len(ts.steps) {
		ts.steps[index].Enabled = enabled
	}
}

// Clear removes all transformation steps
func (ts *TransformationSequence) Clear() {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	ts.steps = make([]TransformationStep, 0)
}

// ToJSON serializes the transformation sequence to JSON
func (ts *TransformationSequence) ToJSON() ([]byte, error) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return json.MarshalIndent(ts.steps, "", "  ")
}

// FromJSON deserializes the transformation sequence from JSON
func (ts *TransformationSequence) FromJSON(data []byte) error {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	return json.Unmarshal(data, &ts.steps)
}

// Length returns the number of transformation steps
func (ts *TransformationSequence) Length() int {
	ts.mu.RLock()
	defer ts.mu.RUnlock()

	return len(ts.steps)
}