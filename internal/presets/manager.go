// Author: Ervins Strauhmanis
// License: MIT

package presets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"advanced-image-processing/internal/models"
)

// Manager handles preset operations (save, load, manage)
type Manager struct {
	logger     *logrus.Logger
	presetsDir string
	presets    map[string]*Preset
}

// NewManager creates a new preset manager
func NewManager(logger *logrus.Logger) *Manager {
	// Create presets directory in user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		logger.Error("Failed to get user home directory, using current directory")
		homeDir = "."
	}
	
	presetsDir := filepath.Join(homeDir, ".advanced-image-processing", "presets")
	
	// Create directory if it doesn't exist
	if err := os.MkdirAll(presetsDir, 0755); err != nil {
		logger.WithError(err).Error("Failed to create presets directory")
		presetsDir = "." // Fallback to current directory
	}

	manager := &Manager{
		logger:     logger,
		presetsDir: presetsDir,
		presets:    make(map[string]*Preset),
	}

	// Load existing presets
	manager.loadAllPresets()
	
	return manager
}

// SavePreset saves a preset to disk
func (m *Manager) SavePreset(preset *Preset) error {
	filename := m.sanitizeFilename(preset.Name) + ".json"
	filepath := filepath.Join(m.presetsDir, filename)

	data, err := preset.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to serialize preset: %w", err)
	}

	if err := ioutil.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write preset file: %w", err)
	}

	// Update in-memory cache
	m.presets[preset.Name] = preset

	m.logger.WithFields(logrus.Fields{
		"name":     preset.Name,
		"filepath": filepath,
	}).Info("Preset saved successfully")

	return nil
}

// LoadPreset loads a preset by name
func (m *Manager) LoadPreset(name string) (*Preset, error) {
	if preset, exists := m.presets[name]; exists {
		return preset, nil
	}

	filename := m.sanitizeFilename(name) + ".json"
	filepath := filepath.Join(m.presetsDir, filename)

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read preset file: %w", err)
	}

	preset := &Preset{}
	if err := preset.FromJSON(data); err != nil {
		return nil, fmt.Errorf("failed to parse preset: %w", err)
	}

	// Update in-memory cache
	m.presets[name] = preset

	m.logger.WithField("name", name).Info("Preset loaded successfully")

	return preset, nil
}

// DeletePreset deletes a preset
func (m *Manager) DeletePreset(name string) error {
	filename := m.sanitizeFilename(name) + ".json"
	filepath := filepath.Join(m.presetsDir, filename)

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete preset file: %w", err)
	}

	// Remove from in-memory cache
	delete(m.presets, name)

	m.logger.WithField("name", name).Info("Preset deleted successfully")

	return nil
}

// ListPresets returns a list of all available presets
func (m *Manager) ListPresets() []*Preset {
	presets := make([]*Preset, 0, len(m.presets))
	for _, preset := range m.presets {
		presets = append(presets, preset)
	}
	return presets
}

// GetPresetNames returns a list of preset names
func (m *Manager) GetPresetNames() []string {
	names := make([]string, 0, len(m.presets))
	for name := range m.presets {
		names = append(names, name)
	}
	return names
}

// PresetExists checks if a preset with the given name exists
func (m *Manager) PresetExists(name string) bool {
	_, exists := m.presets[name]
	return exists
}

// RenamePreset renames a preset
func (m *Manager) RenamePreset(oldName, newName string) error {
	preset, exists := m.presets[oldName]
	if !exists {
		return fmt.Errorf("preset not found: %s", oldName)
	}

	if m.PresetExists(newName) {
		return fmt.Errorf("preset already exists: %s", newName)
	}

	// Update preset name
	preset.Name = newName

	// Save with new name
	if err := m.SavePreset(preset); err != nil {
		return fmt.Errorf("failed to save renamed preset: %w", err)
	}

	// Delete old file
	if err := m.DeletePreset(oldName); err != nil {
		m.logger.WithError(err).Warn("Failed to delete old preset file")
	}

	m.logger.WithFields(logrus.Fields{
		"old_name": oldName,
		"new_name": newName,
	}).Info("Preset renamed successfully")

	return nil
}

// CreatePresetFromSequence creates a new preset from a transformation sequence
func (m *Manager) CreatePresetFromSequence(name, description string, sequence *models.TransformationSequence) (*Preset, error) {
	if m.PresetExists(name) {
		return nil, fmt.Errorf("preset already exists: %s", name)
	}

	preset := NewPreset(name, description, sequence)
	
	if err := m.SavePreset(preset); err != nil {
		return nil, fmt.Errorf("failed to save preset: %w", err)
	}

	return preset, nil
}

// ApplyPreset applies a preset to create a transformation sequence
func (m *Manager) ApplyPreset(name string) (*models.TransformationSequence, error) {
	preset, err := m.LoadPreset(name)
	if err != nil {
		return nil, fmt.Errorf("failed to load preset: %w", err)
	}

	sequence := preset.ToTransformationSequence()
	
	m.logger.WithField("name", name).Info("Preset applied successfully")
	
	return sequence, nil
}

// loadAllPresets loads all presets from the presets directory
func (m *Manager) loadAllPresets() {
	files, err := ioutil.ReadDir(m.presetsDir)
	if err != nil {
		m.logger.WithError(err).Warn("Failed to read presets directory")
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filepath := filepath.Join(m.presetsDir, file.Name())
			data, err := ioutil.ReadFile(filepath)
			if err != nil {
				m.logger.WithError(err).WithField("file", file.Name()).Warn("Failed to read preset file")
				continue
			}

			preset := &Preset{}
			if err := preset.FromJSON(data); err != nil {
				m.logger.WithError(err).WithField("file", file.Name()).Warn("Failed to parse preset file")
				continue
			}

			m.presets[preset.Name] = preset
		}
	}

	m.logger.WithField("count", len(m.presets)).Info("Loaded presets from disk")
}

// sanitizeFilename removes invalid characters from filename
func (m *Manager) sanitizeFilename(name string) string {
	// Replace invalid filename characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	
	// Trim spaces and dots
	result = strings.Trim(result, " .")
	
	// Ensure it's not empty
	if result == "" {
		result = "untitled"
	}
	
	return result
}

// GetPresetsDirectory returns the path to the presets directory
func (m *Manager) GetPresetsDirectory() string {
	return m.presetsDir
}