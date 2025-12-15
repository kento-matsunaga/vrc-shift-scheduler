package schedule

import (
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Status represents the status of a date schedule
type Status string

const (
	StatusOpen    Status = "open"    // 回答受付中
	StatusClosed  Status = "closed"  // 締切（確定なし）
	StatusDecided Status = "decided" // 日程確定
)

// NewStatus creates a new Status from a string
func NewStatus(status string) (Status, error) {
	s := Status(status)
	if err := s.Validate(); err != nil {
		return "", err
	}
	return s, nil
}

// Validate validates the status
func (s Status) Validate() error {
	switch s {
	case StatusOpen, StatusClosed, StatusDecided:
		return nil
	default:
		return common.NewValidationError(
			fmt.Sprintf("invalid status: must be 'open', 'closed', or 'decided', got: %s", s),
			nil,
		)
	}
}

func (s Status) String() string {
	return string(s)
}

// IsOpen returns true if the status is open
func (s Status) IsOpen() bool {
	return s == StatusOpen
}

// IsClosed returns true if the status is closed
func (s Status) IsClosed() bool {
	return s == StatusClosed
}

// IsDecided returns true if the status is decided
func (s Status) IsDecided() bool {
	return s == StatusDecided
}

// Availability represents the availability status for a candidate date
type Availability string

const (
	AvailabilityAvailable   Availability = "available"   // 参加可能
	AvailabilityUnavailable Availability = "unavailable" // 参加不可
	AvailabilityMaybe       Availability = "maybe"       // 未定・要相談
)

// NewAvailability creates a new Availability from a string
func NewAvailability(availability string) (Availability, error) {
	a := Availability(availability)
	if err := a.Validate(); err != nil {
		return "", err
	}
	return a, nil
}

// Validate validates the availability
func (a Availability) Validate() error {
	switch a {
	case AvailabilityAvailable, AvailabilityUnavailable, AvailabilityMaybe:
		return nil
	default:
		return common.NewValidationError(
			fmt.Sprintf("invalid availability: must be 'available', 'unavailable', or 'maybe', got: %s", a),
			nil,
		)
	}
}

func (a Availability) String() string {
	return string(a)
}
