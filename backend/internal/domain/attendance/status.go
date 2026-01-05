package attendance

import (
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Status represents the status of an attendance collection
type Status string

const (
	StatusOpen   Status = "open"   // 回答受付中
	StatusClosed Status = "closed" // 締切
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
	case StatusOpen, StatusClosed:
		return nil
	default:
		return common.NewValidationError(
			fmt.Sprintf("invalid status: must be 'open' or 'closed', got: %s", s),
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

// TargetType represents the target type of an attendance collection
type TargetType string

const (
	TargetTypeEvent      	TargetType = "event"       // イベント
	TargetTypeBusinessDay TargetType = "business_day" // 営業日
)

// NewTargetType creates a new TargetType from a string
func NewTargetType(targetType string) (TargetType, error) {
	t := TargetType(targetType)
	if err := t.Validate(); err != nil {
		return "", err
	}
	return t, nil
}

// Validate validates the target type
func (t TargetType) Validate() error {
	switch t {
	case TargetTypeEvent, TargetTypeBusinessDay:
		return nil
	default:
		return common.NewValidationError(
			fmt.Sprintf("invalid target_type: must be 'event' or 'business_day', got: %s", t),
			nil,
		)
	}
}

func (t TargetType) String() string {
	return string(t)
}

// ResponseType represents the type of an attendance response
type ResponseType string

const (
	ResponseTypeAttending  ResponseType = "attending"  // 出席
	ResponseTypeAbsent     ResponseType = "absent"     // 欠席
	ResponseTypeUndecided  ResponseType = "undecided"  // 未定
)

// NewResponseType creates a new ResponseType from a string
func NewResponseType(responseType string) (ResponseType, error) {
	r := ResponseType(responseType)
	if err := r.Validate(); err != nil {
		return "", err
	}
	return r, nil
}

// Validate validates the response type
func (r ResponseType) Validate() error {
	switch r {
	case ResponseTypeAttending, ResponseTypeAbsent, ResponseTypeUndecided:
		return nil
	default:
		return common.NewValidationError(
			fmt.Sprintf("invalid response: must be 'attending', 'absent', or 'undecided', got: %s", r),
			nil,
		)
	}
}

func (r ResponseType) String() string {
	return string(r)
}
