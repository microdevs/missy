package server

import (
	"time"
)

type Timer struct {
	StartTime time.Time
}

func NewTimer() *Timer {
	return &Timer{StartTime: time.Now()}
}

func (t *Timer) Uptime() string {
	return time.Now().Sub(t.StartTime).String()
}

func (t *Timer) durationMillis() float64 {
	return float64(time.Since(t.StartTime) / time.Millisecond)
}
