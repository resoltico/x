// Author: Ervins Strauhmanis
// License: MIT

package utils

import (
	"sync"
	"time"
)

// Debouncer provides debouncing functionality for function calls
type Debouncer struct {
	mu    sync.Mutex
	timer *time.Timer
	delay time.Duration
}

// NewDebouncer creates a new debouncer with the specified delay
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay: delay,
	}
}

// Debounce delays the execution of the given function
// If called multiple times within the delay period, only the last call will execute
func (d *Debouncer) Debounce(fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.delay, fn)
}

// Cancel cancels any pending debounced function execution
func (d *Debouncer) Cancel() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
}