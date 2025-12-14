package clock

import "time"

// Clock is an interface for getting the current time
// App層で使用し、Domain層には now を引数で渡す
type Clock interface {
	Now() time.Time
}

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
