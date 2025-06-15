// Comprehensive metrics system for image quality assessment
package metrics

import (
	"gocv.io/x/gocv"
)

// Metric defines the interface for quality metrics
type Metric interface {
	// Calculate computes the metric value
	Calculate(original, processed gocv.Mat) (float64, error)
	
	// GetName returns the metric name
	GetName() string
	
	// GetDescription returns the metric description
	GetDescription() string
	
	// GetRange returns the value range (min, max)
	GetRange() (float64, float64)
	
	// IsHigherBetter returns true if higher values indicate better quality
	IsHigherBetter() bool
}

// Evaluator manages and calculates multiple metrics
type Evaluator struct {
	metrics map[string]Metric
}

// NewEvaluator creates a new metrics evaluator
func NewEvaluator() *Evaluator {
	e := &Evaluator{
		metrics: make(map[string]Metric),
	}
	
	// Register all available metrics
	e.RegisterDefaultMetrics()
	
	return e
}

// RegisterDefaultMetrics registers all default metrics
func (e *Evaluator) RegisterDefaultMetrics() {
	e.Register("psnr", NewPSNR())
	e.Register("ssim", NewSSIM())
	e.Register("f_measure", NewFMeasure())
	e.Register("mse", NewMSE())
	e.Register("contrast_ratio", NewContrastRatio())
	e.Register("sharpness", NewSharpness())
}

// Register registers a metric
func (e *Evaluator) Register(name string, metric Metric) {
	e.metrics[name] = metric
}

// Calculate calculates a specific metric
func (e *Evaluator) Calculate(name string, original, processed gocv.Mat) (float64, error) {
	metric, exists := e.metrics[name]
	if !exists {
		return 0, fmt.Errorf("metric not found: %s", name)
	}
	
	return metric.Calculate(original, processed)
}

// CalculateAll calculates all registered metrics
func (e *Evaluator) CalculateAll(original, processed gocv.Mat) map[string]float64 {
	results := make(map[string]float64)
	
	for name, metric := range e.metrics {
		if value, err := metric.Calculate(original, processed); err == nil {
			results[name] = value
		}
	}
	
	return results
}

// CalculatePSNR calculates PSNR between two images
func (e *Evaluator) CalculatePSNR(original, processed gocv.Mat) (float64, error) {
	return e.Calculate("psnr", original, processed)
}

// CalculateSSIM calculates SSIM between two images
func (e *Evaluator) CalculateSSIM(original, processed gocv.Mat) (float64, error) {
	return e.Calculate("ssim", original, processed)
}

// CalculateFMeasure calculates F-measure for binarized images
func (e *Evaluator) CalculateFMeasure(original, processed gocv.Mat) (float64, error) {
	return e.Calculate("f_measure", original, processed)
}

// EvaluateStep calculates metrics for a processing step
func (e *Evaluator) EvaluateStep(before, after gocv.Mat, stepName string) map[string]float64 {
	metrics := make(map[string]float64)
	
	// Calculate basic metrics
	if psnr, err := e.CalculatePSNR(before, after); err == nil {
		metrics["psnr"] = psnr
	}
	
	if ssim, err := e.CalculateSSIM(before, after); err == nil {
		metrics["ssim"] = ssim
	}
	
	// Add step-specific metrics based on algorithm type
	switch stepName {
	case "otsu_multi", "otsu_local", "niblack_true", "sauvola_true", "wolf_jolion", "nick":
		// Binarization metrics
		if fMeasure, err := e.CalculateFMeasure(before, after); err == nil {
			metrics["f_measure"] = fMeasure
		}
		
	case "gaussian", "median", "bilateral":
		// Noise reduction metrics
		if contrast, err := e.Calculate("contrast_ratio", before, after); err == nil {
			metrics["contrast_preservation"] = contrast
		}
		
	case "erosion", "dilation", "opening", "closing":
		// Morphological metrics
		if sharpness, err := e.Calculate("sharpness", before, after); err == nil {
			metrics["edge_preservation"] = sharpness
		}
	}
	
	return metrics
}

// GetMetricInfo returns information about all metrics
func (e *Evaluator) GetMetricInfo() map[string]MetricInfo {
	info := make(map[string]MetricInfo)
	
	for name, metric := range e.metrics {
		min, max := metric.GetRange()
		info[name] = MetricInfo{
			Name:          metric.GetName(),
			Description:   metric.GetDescription(),
			Range:         [2]float64{min, max},
			HigherBetter:  metric.IsHigherBetter(),
		}
	}
	
	return info
}

// MetricInfo provides metadata about a metric
type MetricInfo struct {
	Name         string
	Description  string
	Range        [2]float64 // [min, max]
	HigherBetter bool
}

// QualityReport contains comprehensive quality assessment
type QualityReport struct {
	OverallScore float64            `json:"overall_score"`
	Metrics      map[string]float64 `json:"metrics"`
	Analysis     QualityAnalysis    `json:"analysis"`
	Timestamp    string            `json:"timestamp"`
}

// QualityAnalysis provides interpretation of metrics
type QualityAnalysis struct {
	QualityLevel string   `json:"quality_level"` // "excellent", "good", "fair", "poor"
	Issues       []string `json:"issues"`
	Suggestions  []string `json:"suggestions"`
}

// GenerateReport generates a comprehensive quality report
func (e *Evaluator) GenerateReport(original, processed gocv.Mat) QualityReport {
	metrics := e.CalculateAll(original, processed)
	
	// Calculate overall score (weighted average)
	overallScore := e.calculateOverallScore(metrics)
	
	// Analyze quality
	analysis := e.analyzeQuality(metrics)
	
	return QualityReport{
		OverallScore: overallScore,
		Metrics:      metrics,
		Analysis:     analysis,
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
	}
}

// calculateOverallScore calculates a weighted overall quality score
func (e *Evaluator) calculateOverallScore(metrics map[string]float64) float64 {
	weights := map[string]float64{
		"psnr":          0.3,
		"ssim":          0.3,
		"f_measure":     0.2,
		"contrast_ratio": 0.1,
		"sharpness":     0.1,
	}
	
	totalWeight := 0.0
	weightedSum := 0.0
	
	for name, weight := range weights {
		if value, exists := metrics[name]; exists {
			// Normalize metrics to 0-1 range
			normalizedValue := e.normalizeMetric(name, value)
			weightedSum += normalizedValue * weight
			totalWeight += weight
		}
	}
	
	if totalWeight == 0 {
		return 0
	}
	
	return (weightedSum / totalWeight) * 100 // Return as percentage
}

// normalizeMetric normalizes a metric value to 0-1 range
func (e *Evaluator) normalizeMetric(name string, value float64) float64 {
	metric, exists := e.metrics[name]
	if !exists {
		return 0
	}
	
	min, max := metric.GetRange()
	
	// Clamp value to range
	if value < min {
		value = min
	}
	if value > max {
		value = max
	}
	
	// Normalize to 0-1
	if max == min {
		return 1.0
	}
	
	normalized := (value - min) / (max - min)
	
	// Invert if lower is better
	if !metric.IsHigherBetter() {
		normalized = 1.0 - normalized
	}
	
	return normalized
}

// analyzeQuality analyzes quality metrics and provides insights
func (e *Evaluator) analyzeQuality(metrics map[string]float64) QualityAnalysis {
	analysis := QualityAnalysis{
		Issues:      make([]string, 0),
		Suggestions: make([]string, 0),
	}
	
	// Determine overall quality level
	overallScore := e.calculateOverallScore(metrics)
	
	switch {
	case overallScore >= 90:
		analysis.QualityLevel = "excellent"
	case overallScore >= 75:
		analysis.QualityLevel = "good"
	case overallScore >= 60:
		analysis.QualityLevel = "fair"
	default:
		analysis.QualityLevel = "poor"
	}
	
	// Analyze specific metrics
	if psnr, exists := metrics["psnr"]; exists {
		if psnr < 20 {
			analysis.Issues = append(analysis.Issues, "Low PSNR indicates significant noise or distortion")
			analysis.Suggestions = append(analysis.Suggestions, "Consider noise reduction or different binarization parameters")
		}
	}
	
	if ssim, exists := metrics["ssim"]; exists {
		if ssim < 0.7 {
			analysis.Issues = append(analysis.Issues, "Low SSIM indicates poor structural similarity")
			analysis.Suggestions = append(analysis.Suggestions, "Adjust processing parameters to preserve image structure")
		}
	}
	
	if fMeasure, exists := metrics["f_measure"]; exists {
		if fMeasure < 0.8 {
			analysis.Issues = append(analysis.Issues, "Low F-measure indicates poor text/background separation")
			analysis.Suggestions = append(analysis.Suggestions, "Try different binarization algorithm or adjust window size")
		}
	}
	
	return analysis
}
