package service

import (
	"time"
)

// Timer type
type Timer struct {
	StartTime time.Time
}

// NewTimer returns a new Timer instance
func NewTimer() *Timer {
	return &Timer{StartTime: time.Now()}
}

// Uptime tells the difference between Now and start time
func (t *Timer) Uptime() string {
	return time.Now().Sub(t.StartTime).String()
}

// durationMillis is a helper to return Time in milliseconds
func (t *Timer) durationMillis() float64 {
	return float64(time.Since(t.StartTime) / time.Millisecond)
}
