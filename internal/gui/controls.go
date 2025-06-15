// Author: Ervins Strauhmanis
// License: MIT

package gui

import (
	"fmt"
	"strconv"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/sirupsen/logrus"

	"advanced-image-processing/internal/image_processing"
	"advanced-image-processing/internal/transforms"
)

// Controls manages the parameter adjustment interface
type Controls struct {
	mu                sync.RWMutex
	registry          *transforms.TransformRegistry
	pipeline          *image_processing.Pipeline
	logger            *logrus.Logger
	
	// GUI components
	container         *fyne.Container
	transformList     *widget.List
	paramContainer    *fyne.Container
	addButton         *widget.Button
	clearButton       *widget.Button
	
	// Current state
	selectedIndex     int
	enabled           bool
	
	// Callbacks
	onTransformationChanged func()
}

// NewControls creates a new controls component
func NewControls(registry *transforms.TransformRegistry, pipeline *image_processing.Pipeline, logger *logrus.Logger) *Controls {
	c := &Controls{
		registry:      registry,
		pipeline:      pipeline,
		logger:        logger,
		selectedIndex: -1,
		enabled:       false,
	}
	
	c.initializeComponents()
	c.setupLayout()
	
	return c
}

// initializeComponents initializes the control components
func (c *Controls) initializeComponents() {
	// Create transformation list
	c.transformList = widget.NewList(
		func() int {
			sequence := c.pipeline.GetTransformationSequence()
			return sequence.Length()
		},
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewCheck("", nil),
				widget.NewLabel("Transform"),
				widget.NewButton("Remove", nil),
			)
		},
		func(id int, obj fyne.CanvasObject) {
			c.updateListItem(id, obj)
		},
	)
	
	c.transformList.OnSelected = func(id int) {
		c.selectTransformation(id)
	}
	
	// Create buttons
	c.clearButton = widget.NewButton("Clear All", func() {
		c.clearAllTransformations()
	})
	
	// Create parameter container
	c.paramContainer = container.NewVBox(
		widget.NewLabel("Select a transformation to edit parameters"),
	)
	
	// Initially disable controls
	c.setEnabled(false)
}

// setupLayout creates the controls layout
func (c *Controls) setupLayout() {
	// Transformation list section
	listSection := container.NewBorder(
		widget.NewLabel("Transformations:"),
		c.clearButton,
		nil,
		nil,
		c.transformList,
	)
	
	// Parameter section
	paramSection := container.NewBorder(
		widget.NewLabel("Parameters:"),
		nil,
		nil,
		nil,
		container.NewScroll(c.paramContainer),
	)
	
	// Combine in vertical split
	c.container = container.NewVSplit(
		listSection,
		paramSection,
	)
	c.container.SetOffset(0.4) // 40% for list, 60% for parameters
}

// updateListItem updates a single item in the transformation list
func (c *Controls) updateListItem(id int, obj fyne.CanvasObject) {
	container := obj.(*container.Horizontal)
	check := container.Objects[0].(*widget.Check)
	label := container.Objects[1].(*widget.Label)
	button := container.Objects[2].(*widget.Button)
	
	sequence := c.pipeline.GetTransformationSequence()
	steps := sequence.GetSteps()
	
	if id >= len(steps) {
		return
	}
	
	step := steps[id]
	
	// Update checkbox
	check.SetChecked(step.Enabled)
	check.OnChanged = func(enabled bool) {
		c.toggleTransformation(id, enabled)
	}
	
	// Update label
	if transform, exists := c.registry.Get(step.Type); exists {
		label.SetText(transform.GetName())
	} else {
		label.SetText(step.Type)
	}
	
	// Update remove button
	button.OnTapped = func() {
		c.removeTransformation(id)
	}
}

// selectTransformation selects a transformation for parameter editing
func (c *Controls) selectTransformation(index int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.selectedIndex = index
	c.updateParameterControls()
}

// updateParameterControls updates the parameter editing interface
func (c *Controls) updateParameterControls() {
	sequence := c.pipeline.GetTransformationSequence()
	steps := sequence.GetSteps()
	
	if c.selectedIndex < 0 || c.selectedIndex >= len(steps) {
		c.paramContainer.Objects = []fyne.CanvasObject{
			widget.NewLabel("Select a transformation to edit parameters"),
		}
		c.paramContainer.Refresh()
		return
	}
	
	step := steps[c.selectedIndex]
	transform, exists := c.registry.Get(step.Type)
	if !exists {
		c.paramContainer.Objects = []fyne.CanvasObject{
			widget.NewLabel("Unknown transformation type"),
		}
		c.paramContainer.Refresh()
		return
	}
	
	// Get parameter info
	var paramInfo []transforms.ParameterInfo
	if infoProvider, ok := transform.(transforms.TransformInfo); ok {
		paramInfo = infoProvider.GetParameterInfo()
	}
	
	// Create parameter controls
	controls := make([]fyne.CanvasObject, 0)
	
	// Add transformation description
	controls = append(controls, widget.NewRichTextFromMarkdown(fmt.Sprintf("**%s**\n\n%s", 
		transform.GetName(), transform.GetDescription())))
	
	// Add parameter controls
	for _, info := range paramInfo {
		control := c.createParameterControl(info, step.Parameters)
		if control != nil {
			controls = append(controls, control)
		}
	}
	
	c.paramContainer.Objects = controls
	c.paramContainer.Refresh()
}

// createParameterControl creates a control widget for a parameter
func (c *Controls) createParameterControl(info transforms.ParameterInfo, currentParams map[string]interface{}) fyne.CanvasObject {
	var control fyne.CanvasObject
	
	// Get current value
	currentValue := info.Default
	if val, exists := currentParams[info.Name]; exists {
		currentValue = val
	}
	
	switch info.Type {
	case "int":
		min := int(info.Min.(float64))
		max := int(info.Max.(float64))
		current := int(currentValue.(float64))
		
		slider := widget.NewSlider(float64(min), float64(max))
		slider.SetValue(float64(current))
		slider.Step = 1
		
		label := widget.NewLabel(fmt.Sprintf("%s: %d", info.Name, current))
		
		slider.OnChanged = func(value float64) {
			intValue := int(value)
			label.SetText(fmt.Sprintf("%s: %d", info.Name, intValue))
			c.updateParameter(info.Name, float64(intValue))
		}
		
		control = container.NewVBox(
			label,
			slider,
			widget.NewLabel(info.Description),
		)
		
	case "float":
		min := info.Min.(float64)
		max := info.Max.(float64)
		current := currentValue.(float64)
		
		slider := widget.NewSlider(min, max)
		slider.SetValue(current)
		slider.Step = (max - min) / 100
		
		label := widget.NewLabel(fmt.Sprintf("%s: %.2f", info.Name, current))
		
		slider.OnChanged = func(value float64) {
			label.SetText(fmt.Sprintf("%s: %.2f", info.Name, value))
			c.updateParameter(info.Name, value)
		}
		
		control = container.NewVBox(
			label,
			slider,
			widget.NewLabel(info.Description),
		)
		
	case "bool":
		current := currentValue.(bool)
		
		check := widget.NewCheck(info.Name, func(checked bool) {
			c.updateParameter(info.Name, checked)
		})
		check.SetChecked(current)
		
		control = container.NewVBox(
			check,
			widget.NewLabel(info.Description),
		)
		
	case "enum":
		current := currentValue.(string)
		options := info.Options
		
		dropdown := widget.NewSelect(options, func(selected string) {
			c.updateParameter(info.Name, selected)
		})
		dropdown.SetSelected(current)
		
		control = container.NewVBox(
			widget.NewLabel(info.Name),
			dropdown,
			widget.NewLabel(info.Description),
		)
		
	default:
		// String or unknown type - use text entry
		entry := widget.NewEntry()
		entry.SetText(fmt.Sprintf("%v", currentValue))
		entry.OnChanged = func(text string) {
			// Try to convert to appropriate type
			if val, err := strconv.ParseFloat(text, 64); err == nil {
				c.updateParameter(info.Name, val)
			} else {
				c.updateParameter(info.Name, text)
			}
		}
		
		control = container.NewVBox(
			widget.NewLabel(info.Name),
			entry,
			widget.NewLabel(info.Description),
		)
	}
	
	return control
}

// updateParameter updates a parameter value
func (c *Controls) updateParameter(paramName string, value interface{}) {
	if c.selectedIndex < 0 {
		return
	}
	
	sequence := c.pipeline.GetTransformationSequence()
	steps := sequence.GetSteps()
	
	if c.selectedIndex >= len(steps) {
		return
	}
	
	// Update parameter
	params := make(map[string]interface{})
	for k, v := range steps[c.selectedIndex].Parameters {
		params[k] = v
	}
	params[paramName] = value
	
	// Update pipeline
	if err := c.pipeline.UpdateTransformation(c.selectedIndex, params); err != nil {
		c.logger.WithError(err).Error("Failed to update transformation parameter")
		return
	}
	
	// Trigger callback
	if c.onTransformationChanged != nil {
		c.onTransformationChanged()
	}
	
	c.logger.WithFields(logrus.Fields{
		"index":     c.selectedIndex,
		"parameter": paramName,
		"value":     value,
	}).Debug("Updated parameter")
}

// removeTransformation removes a transformation from the pipeline
func (c *Controls) removeTransformation(index int) {
	if err := c.pipeline.RemoveTransformation(index); err != nil {
		c.logger.WithError(err).Error("Failed to remove transformation")
		return
	}
	
	// Clear selection if it was the removed item
	if c.selectedIndex == index {
		c.selectedIndex = -1
		c.updateParameterControls()
	} else if c.selectedIndex > index {
		c.selectedIndex-- // Adjust for removed item
	}
	
	c.transformList.Refresh()
	
	// Trigger callback
	if c.onTransformationChanged != nil {
		c.onTransformationChanged()
	}
	
	c.logger.WithField("index", index).Info("Removed transformation")
}

// toggleTransformation enables/disables a transformation
func (c *Controls) toggleTransformation(index int, enabled bool) {
	sequence := c.pipeline.GetTransformationSequence()
	sequence.ToggleStep(index, enabled)
	
	// Trigger reprocessing
	if c.onTransformationChanged != nil {
		c.onTransformationChanged()
	}
	
	c.logger.WithFields(logrus.Fields{
		"index":   index,
		"enabled": enabled,
	}).Debug("Toggled transformation")
}

// clearAllTransformations clears all transformations
func (c *Controls) clearAllTransformations() {
	c.pipeline.ClearSequence()
	c.selectedIndex = -1
	c.updateParameterControls()
	c.transformList.Refresh()
	
	// Trigger callback
	if c.onTransformationChanged != nil {
		c.onTransformationChanged()
	}
	
	c.logger.Info("Cleared all transformations")
}

// RefreshSequence refreshes the transformation list
func (c *Controls) RefreshSequence() {
	c.transformList.Refresh()
	c.updateParameterControls()
}

// Enable enables the controls
func (c *Controls) Enable() {
	c.setEnabled(true)
}

// Disable disables the controls
func (c *Controls) Disable() {
	c.setEnabled(false)
}

// setEnabled sets the enabled state of controls
func (c *Controls) setEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.enabled = enabled
	
	if enabled {
		c.clearButton.Enable()
	} else {
		c.clearButton.Disable()
	}
}

// SetCallbacks sets callback functions
func (c *Controls) SetCallbacks(onTransformationChanged func()) {
	c.onTransformationChanged = onTransformationChanged
}

// GetContainer returns the controls container
func (c *Controls) GetContainer() *fyne.Container {
	return c.container
}