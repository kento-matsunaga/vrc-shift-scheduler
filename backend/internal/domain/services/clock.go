package services

import "time"

// Clock is an interface for getting the current time.
// Domain layer should receive time as a parameter, but Application layer
// uses this interface for dependency injection.
type Clock interface {
	Now() time.Time
}
