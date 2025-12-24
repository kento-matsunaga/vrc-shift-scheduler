package clock

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// Compile-time interface compliance check
var (
	_ services.Clock = (*RealClock)(nil)
	_ services.Clock = (*FixedClock)(nil)
)

// RealClock is the production implementation that returns the actual current time
type RealClock struct{}

// NewRealClock creates a new RealClock
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current time
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// FixedClock is a test implementation that returns a fixed time
type FixedClock struct {
	FixedTime time.Time
}

// NewFixedClock creates a new FixedClock with the given time
func NewFixedClock(t time.Time) *FixedClock {
	return &FixedClock{FixedTime: t}
}

// Now returns the fixed time
func (c *FixedClock) Now() time.Time {
	return c.FixedTime
}
