package schedule

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// DateScheduleResponse represents a schedule response entity
type DateScheduleResponse struct {
	responseID   common.ResponseID
	tenantID     common.TenantID
	scheduleID   common.ScheduleID
	memberID     common.MemberID
	candidateID  common.CandidateID
	availability Availability
	note         string
	respondedAt  time.Time
	createdAt    time.Time
	updatedAt    time.Time
}

// NewDateScheduleResponse creates a new DateScheduleResponse entity
func NewDateScheduleResponse(
	now time.Time,
	scheduleID common.ScheduleID,
	tenantID common.TenantID,
	memberID common.MemberID,
	candidateID common.CandidateID,
	availability Availability,
	note string,
) (*DateScheduleResponse, error) {
	response := &DateScheduleResponse{
		responseID:   common.NewResponseID(),
		tenantID:     tenantID,
		scheduleID:   scheduleID,
		memberID:     memberID,
		candidateID:  candidateID,
		availability: availability,
		note:         note,
		respondedAt:  now,
		createdAt:    now,
		updatedAt:    now,
	}

	if err := response.validate(); err != nil {
		return nil, err
	}

	return response, nil
}

// ReconstructDateScheduleResponse reconstructs a DateScheduleResponse entity from persistence
func ReconstructDateScheduleResponse(
	responseID common.ResponseID,
	tenantID common.TenantID,
	scheduleID common.ScheduleID,
	memberID common.MemberID,
	candidateID common.CandidateID,
	availability Availability,
	note string,
	respondedAt time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (*DateScheduleResponse, error) {
	response := &DateScheduleResponse{
		responseID:   responseID,
		tenantID:     tenantID,
		scheduleID:   scheduleID,
		memberID:     memberID,
		candidateID:  candidateID,
		availability: availability,
		note:         note,
		respondedAt:  respondedAt,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}

	if err := response.validate(); err != nil {
		return nil, err
	}

	return response, nil
}

func (r *DateScheduleResponse) validate() error {
	if err := r.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}
	if err := r.scheduleID.Validate(); err != nil {
		return common.NewValidationError("schedule_id is required", err)
	}
	if err := r.memberID.Validate(); err != nil {
		return common.NewValidationError("member_id is required", err)
	}
	if err := r.candidateID.Validate(); err != nil {
		return common.NewValidationError("candidate_id is required", err)
	}
	if err := r.availability.Validate(); err != nil {
		return err
	}
	return nil
}

// Getters

func (r *DateScheduleResponse) ResponseID() common.ResponseID {
	return r.responseID
}

func (r *DateScheduleResponse) TenantID() common.TenantID {
	return r.tenantID
}

func (r *DateScheduleResponse) ScheduleID() common.ScheduleID {
	return r.scheduleID
}

func (r *DateScheduleResponse) MemberID() common.MemberID {
	return r.memberID
}

func (r *DateScheduleResponse) CandidateID() common.CandidateID {
	return r.candidateID
}

func (r *DateScheduleResponse) Availability() Availability {
	return r.availability
}

func (r *DateScheduleResponse) Note() string {
	return r.note
}

func (r *DateScheduleResponse) RespondedAt() time.Time {
	return r.respondedAt
}

func (r *DateScheduleResponse) CreatedAt() time.Time {
	return r.createdAt
}

func (r *DateScheduleResponse) UpdatedAt() time.Time {
	return r.updatedAt
}
