package attendance

import "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"

var (
	// ErrCollectionClosed is returned when trying to respond to a closed collection
	ErrCollectionClosed = common.NewInvariantViolationError("collection is closed")

	// ErrDeadlinePassed is returned when the deadline has passed
	ErrDeadlinePassed = common.NewInvariantViolationError("deadline has passed")

	// ErrAlreadyClosed is returned when trying to close an already closed collection
	ErrAlreadyClosed = common.NewInvariantViolationError("collection is already closed")
)
