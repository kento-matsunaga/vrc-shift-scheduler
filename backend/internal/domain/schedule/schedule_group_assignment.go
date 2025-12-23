package schedule

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ScheduleGroupAssignment represents the association between a date schedule and a member group
// This is used to restrict which members can respond to a date schedule
type ScheduleGroupAssignment struct {
	scheduleID common.ScheduleID
	groupID    common.MemberGroupID
	createdAt  time.Time
}

// NewScheduleGroupAssignment creates a new ScheduleGroupAssignment
func NewScheduleGroupAssignment(
	now time.Time,
	scheduleID common.ScheduleID,
	groupID common.MemberGroupID,
) (*ScheduleGroupAssignment, error) {
	assignment := &ScheduleGroupAssignment{
		scheduleID: scheduleID,
		groupID:    groupID,
		createdAt:  now,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

// ReconstructScheduleGroupAssignment reconstructs from persistence
func ReconstructScheduleGroupAssignment(
	scheduleID common.ScheduleID,
	groupID common.MemberGroupID,
	createdAt time.Time,
) (*ScheduleGroupAssignment, error) {
	assignment := &ScheduleGroupAssignment{
		scheduleID: scheduleID,
		groupID:    groupID,
		createdAt:  createdAt,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (a *ScheduleGroupAssignment) validate() error {
	if err := a.scheduleID.Validate(); err != nil {
		return common.NewValidationError("schedule_id is required", err)
	}
	if err := a.groupID.Validate(); err != nil {
		return common.NewValidationError("group_id is required", err)
	}
	return nil
}

// Getters

func (a *ScheduleGroupAssignment) ScheduleID() common.ScheduleID {
	return a.scheduleID
}

func (a *ScheduleGroupAssignment) GroupID() common.MemberGroupID {
	return a.groupID
}

func (a *ScheduleGroupAssignment) CreatedAt() time.Time {
	return a.createdAt
}
