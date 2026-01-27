package schedule

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// CandidateDate represents a candidate date entity
type CandidateDate struct {
	candidateID   common.CandidateID
	scheduleID    common.ScheduleID
	candidateDate time.Time  // date part
	startTime     *time.Time // time part (optional)
	endTime       *time.Time // time part (optional)
	displayOrder  int
	createdAt     time.Time
}

// NewCandidateDate creates a new CandidateDate entity
func NewCandidateDate(
	now time.Time,
	scheduleID common.ScheduleID,
	candidateDate time.Time,
	startTime *time.Time,
	endTime *time.Time,
	displayOrder int,
) (*CandidateDate, error) {
	candidate := &CandidateDate{
		candidateID:   common.NewCandidateID(),
		scheduleID:    scheduleID,
		candidateDate: candidateDate,
		startTime:     startTime,
		endTime:       endTime,
		displayOrder:  displayOrder,
		createdAt:     now,
	}

	if err := candidate.validate(); err != nil {
		return nil, err
	}

	return candidate, nil
}

// ReconstructCandidateDate reconstructs a CandidateDate entity from persistence
func ReconstructCandidateDate(
	candidateID common.CandidateID,
	scheduleID common.ScheduleID,
	candidateDate time.Time,
	startTime *time.Time,
	endTime *time.Time,
	displayOrder int,
	createdAt time.Time,
) (*CandidateDate, error) {
	candidate := &CandidateDate{
		candidateID:   candidateID,
		scheduleID:    scheduleID,
		candidateDate: candidateDate,
		startTime:     startTime,
		endTime:       endTime,
		displayOrder:  displayOrder,
		createdAt:     createdAt,
	}

	if err := candidate.validate(); err != nil {
		return nil, err
	}

	return candidate, nil
}

func (c *CandidateDate) validate() error {
	// CandidateID の必須性チェック
	if err := c.candidateID.Validate(); err != nil {
		return common.NewValidationError("candidate_id is required", err)
	}

	// ScheduleID の必須性チェック
	if err := c.scheduleID.Validate(); err != nil {
		return common.NewValidationError("schedule_id is required", err)
	}

	return nil
}

// Getters

func (c *CandidateDate) CandidateID() common.CandidateID {
	return c.candidateID
}

func (c *CandidateDate) ScheduleID() common.ScheduleID {
	return c.scheduleID
}

func (c *CandidateDate) CandidateDateValue() time.Time {
	return c.candidateDate
}

func (c *CandidateDate) StartTime() *time.Time {
	return c.startTime
}

func (c *CandidateDate) EndTime() *time.Time {
	return c.endTime
}

func (c *CandidateDate) DisplayOrder() int {
	return c.displayOrder
}

func (c *CandidateDate) CreatedAt() time.Time {
	return c.createdAt
}
