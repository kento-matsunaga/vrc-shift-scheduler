package attendance

import "errors"

var (
	// ErrCollectionClosed is returned when trying to respond to a closed collection
	ErrCollectionClosed = errors.New("collection is closed")

	// ErrDeadlinePassed is returned when the deadline has passed
	ErrDeadlinePassed = errors.New("deadline has passed")

	// ErrAlreadyClosed is returned when trying to close an already closed collection
	ErrAlreadyClosed = errors.New("collection is already closed")
)
