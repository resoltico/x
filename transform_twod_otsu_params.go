package main

func (t *TwoDOtsu) GetParameters() map[string]interface{} {
	t.paramMutex.RLock()
	defer t.paramMutex.RUnlock()

	return map[string]interface{}{
		"windowRadius":    t.windowRadius,
		"epsilon":         t.epsilon,
		"morphKernelSize": t.morphKernelSize,
	}
}

func (t *TwoDOtsu) SetParameters(params map[string]interface{}) {
	t.paramMutex.Lock()
	defer t.paramMutex.Unlock()

	if radius, ok := params["windowRadius"].(int); ok {
		if radius >= 1 && radius <= 20 {
			t.windowRadius = radius
		}
	}
	if eps, ok := params["epsilon"].(float64); ok {
		if eps > 0.001 && eps <= 1.0 {
			t.epsilon = eps
		}
	}
	if kernel, ok := params["morphKernelSize"].(int); ok {
		if kernel >= 1 && kernel <= 15 && kernel%2 == 1 {
			t.morphKernelSize = kernel
		}
	}
}
