package schedule

import (
	"errors"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

var (
	// ErrScheduleClosed is returned when trying to respond to a closed schedule
	ErrScheduleClosed = errors.New("schedule is closed")

	// ErrDeadlinePassed is returned when the deadline has passed
	ErrDeadlinePassed = errors.New("deadline has passed")

	// ErrAlreadyClosed is returned when trying to close an already closed schedule
	ErrAlreadyClosed = errors.New("schedule is already closed")

	// ErrAlreadyDecided is returned when trying to decide an already decided schedule
	ErrAlreadyDecided = errors.New("schedule is already decided")

	// ErrCandidateNotFound is returned when the specified candidate is not found
	ErrCandidateNotFound = errors.New("candidate not found")

	// ErrAlreadyDeleted is returned when trying to delete an already deleted schedule
	ErrAlreadyDeleted = common.NewInvariantViolationError("schedule is already deleted")
)
