// Author: Ervins Strauhmanis
// License: MIT

package presets

import (
	"encoding/json"
	"fmt"
	"time"

	"advanced-image-processing/internal/models"
)

// Preset represents a saved transformation sequence with metadata
type Preset struct {
	Name         string                         `json:"name"`
	Description  string                         `json:"description"`
	CreatedAt    time.Time                      `json:"created_at"`
	ModifiedAt   time.Time                      `json:"modified_at"`
	Version      string                         `json:"version"`
	Transformations []models.TransformationStep `json:"transformations"`
	Tags         []string                       `json:"tags"`
	Author       string                         `json:"author"`
}

// NewPreset creates a new preset with the given name and transformation sequence
func NewPreset(name, description string, sequence *models.TransformationSequence) *Preset {
	now := time.Now()
	return &Preset{
		Name:            name,
		Description:     description,
		CreatedAt:       now,
		ModifiedAt:      now,
		Version:         "1.0",
		Transformations: sequence.GetSteps(),
		Tags:            make([]string, 0),
		Author:          "Advanced Image Processing",
	}
}

// ToJSON serializes the preset to JSON
func (p *Preset) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// FromJSON deserializes a preset from JSON
func (p *Preset) FromJSON(data []byte) error {
	return json.Unmarshal(data, p)
}

// ToTransformationSequence converts the preset to a transformation sequence
func (p *Preset) ToTransformationSequence() *models.TransformationSequence {
	sequence := models.NewTransformationSequence()
	
	for _, step := range p.Transformations {
		sequence.AddStep(step.Type, step.Parameters)
		// Apply enabled state if it was stored
		if len(sequence.GetSteps()) > 0 {
			lastIndex := sequence.Length() - 1
			sequence.ToggleStep(lastIndex, step.Enabled)
		}
	}
	
	return sequence
}

// UpdateFromSequence updates the preset with a new transformation sequence
func (p *Preset) UpdateFromSequence(sequence *models.TransformationSequence) {
	p.Transformations = sequence.GetSteps()
	p.ModifiedAt = time.Now()
}

// AddTag adds a tag to the preset
func (p *Preset) AddTag(tag string) {
	// Check if tag already exists
	for _, existingTag := range p.Tags {
		if existingTag == tag {
			return
		}
	}
	p.Tags = append(p.Tags, tag)
	p.ModifiedAt = time.Now()
}

// RemoveTag removes a tag from the preset
func (p *Preset) RemoveTag(tag string) {
	for i, existingTag := range p.Tags {
		if existingTag == tag {
			p.Tags = append(p.Tags[:i], p.Tags[i+1:]...)
			p.ModifiedAt = time.Now()
			return
		}
	}
}

// HasTag checks if the preset has a specific tag
func (p *Preset) HasTag(tag string) bool {
	for _, existingTag := range p.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// Clone creates a deep copy of the preset
func (p *Preset) Clone() *Preset {
	clone := &Preset{
		Name:        p.Name + " (Copy)",
		Description: p.Description,
		CreatedAt:   time.Now(),
		ModifiedAt:  time.Now(),
		Version:     p.Version,
		Author:      p.Author,
		Tags:        make([]string, len(p.Tags)),
		Transformations: make([]models.TransformationStep, len(p.Transformations)),
	}
	
	copy(clone.Tags, p.Tags)
	copy(clone.Transformations, p.Transformations)
	
	return clone
}

// GetSummary returns a brief summary of the preset
func (p *Preset) GetSummary() string {
	if len(p.Transformations) == 0 {
		return "Empty preset"
	}
	
	if len(p.Transformations) == 1 {
		return "1 transformation"
	}
	
	return fmt.Sprintf("%d transformations", len(p.Transformations))
}