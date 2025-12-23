package event

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// EventGroupAssignment represents the association between an event and a member group
// This is used to restrict which members can participate in an event
type EventGroupAssignment struct {
	eventID   common.EventID
	groupID   common.MemberGroupID
	createdAt time.Time
}

// NewEventGroupAssignment creates a new EventGroupAssignment
func NewEventGroupAssignment(
	now time.Time,
	eventID common.EventID,
	groupID common.MemberGroupID,
) (*EventGroupAssignment, error) {
	assignment := &EventGroupAssignment{
		eventID:   eventID,
		groupID:   groupID,
		createdAt: now,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

// ReconstructEventGroupAssignment reconstructs an EventGroupAssignment from persistence
func ReconstructEventGroupAssignment(
	eventID common.EventID,
	groupID common.MemberGroupID,
	createdAt time.Time,
) (*EventGroupAssignment, error) {
	assignment := &EventGroupAssignment{
		eventID:   eventID,
		groupID:   groupID,
		createdAt: createdAt,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (a *EventGroupAssignment) validate() error {
	if err := a.eventID.Validate(); err != nil {
		return common.NewValidationError("event_id is required", err)
	}
	if err := a.groupID.Validate(); err != nil {
		return common.NewValidationError("group_id is required", err)
	}
	return nil
}

// Getters

func (a *EventGroupAssignment) EventID() common.EventID {
	return a.eventID
}

func (a *EventGroupAssignment) GroupID() common.MemberGroupID {
	return a.groupID
}

func (a *EventGroupAssignment) CreatedAt() time.Time {
	return a.createdAt
}

// EventRoleGroupAssignment represents the association between an event and a role group
// This is used to restrict which roles can participate in an event
type EventRoleGroupAssignment struct {
	eventID     common.EventID
	roleGroupID common.RoleGroupID
	createdAt   time.Time
}

// NewEventRoleGroupAssignment creates a new EventRoleGroupAssignment
func NewEventRoleGroupAssignment(
	now time.Time,
	eventID common.EventID,
	roleGroupID common.RoleGroupID,
) (*EventRoleGroupAssignment, error) {
	assignment := &EventRoleGroupAssignment{
		eventID:     eventID,
		roleGroupID: roleGroupID,
		createdAt:   now,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

// ReconstructEventRoleGroupAssignment reconstructs from persistence
func ReconstructEventRoleGroupAssignment(
	eventID common.EventID,
	roleGroupID common.RoleGroupID,
	createdAt time.Time,
) (*EventRoleGroupAssignment, error) {
	assignment := &EventRoleGroupAssignment{
		eventID:     eventID,
		roleGroupID: roleGroupID,
		createdAt:   createdAt,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (a *EventRoleGroupAssignment) validate() error {
	if err := a.eventID.Validate(); err != nil {
		return common.NewValidationError("event_id is required", err)
	}
	if err := a.roleGroupID.Validate(); err != nil {
		return common.NewValidationError("role_group_id is required", err)
	}
	return nil
}

// Getters

func (a *EventRoleGroupAssignment) EventID() common.EventID {
	return a.eventID
}

func (a *EventRoleGroupAssignment) RoleGroupID() common.RoleGroupID {
	return a.roleGroupID
}

func (a *EventRoleGroupAssignment) CreatedAt() time.Time {
	return a.createdAt
}
