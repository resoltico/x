package main

import (
	"fmt"
	"math"
)

func (l *Lanczos4Transform) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"scaleFactor":  l.scaleFactor,
		"targetDPI":    l.targetDPI,
		"originalDPI":  l.originalDPI,
		"useIterative": l.useIterative,
	}
}

func (l *Lanczos4Transform) SetParameters(params map[string]interface{}) {
	if scale, ok := params["scaleFactor"].(float64); ok {
		if scale >= 0.1 && scale <= 10.0 {
			l.scaleFactor = scale
		}
	}
	if target, ok := params["targetDPI"].(float64); ok {
		if target >= 72 && target <= 2400 {
			l.targetDPI = target
		}
	}
	if original, ok := params["originalDPI"].(float64); ok {
		if original >= 72 && original <= 2400 {
			l.originalDPI = original
		}
	}
	if iterative, ok := params["useIterative"].(bool); ok {
		l.useIterative = iterative
	}
}

func (l *Lanczos4Transform) calculateScaleFactor() float64 {
	if l.originalDPI > 0 && l.targetDPI > 0 && !math.IsInf(l.originalDPI, 0) && !math.IsInf(l.targetDPI, 0) {
		calculated := l.targetDPI / l.originalDPI

		if calculated > 0.01 && calculated < 100.0 {
			l.debugImage.LogAlgorithmStep("Lanczos4", fmt.Sprintf("Calculated scale factor: %.3f (%.0f DPI -> %.0f DPI)",
				calculated, l.originalDPI, l.targetDPI))
			return calculated
		}
	}
	return l.scaleFactor
}
