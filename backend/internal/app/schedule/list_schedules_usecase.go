package schedule

import (
	"context"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
)

// ListSchedulesUsecase handles listing schedules for a tenant
type ListSchedulesUsecase struct {
	scheduleRepo schedule.DateScheduleRepository
}

// NewListSchedulesUsecase creates a new ListSchedulesUsecase
func NewListSchedulesUsecase(scheduleRepo schedule.DateScheduleRepository) *ListSchedulesUsecase {
	return &ListSchedulesUsecase{
		scheduleRepo: scheduleRepo,
	}
}

// ListSchedulesInput represents the input for listing schedules
type ListSchedulesInput struct {
	TenantID string
}

// ListSchedulesOutput represents the output for listing schedules
type ListSchedulesOutput struct {
	Schedules []ScheduleSummary `json:"schedules"`
}

// ScheduleSummary represents a summary of a schedule
type ScheduleSummary struct {
	ScheduleID         string     `json:"schedule_id"`
	TenantID           string     `json:"tenant_id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	EventID            *string    `json:"event_id"`
	PublicToken        string     `json:"public_token"`
	Status             string     `json:"status"`
	Deadline           *time.Time `json:"deadline"`
	DecidedCandidateID *string    `json:"decided_candidate_id"`
	CandidateCount     int        `json:"candidate_count"`
	ResponseCount      int        `json:"response_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Execute executes the list schedules use case
func (u *ListSchedulesUsecase) Execute(ctx context.Context, input ListSchedulesInput) (*ListSchedulesOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// 2. Find all schedules for this tenant
	schedules, err := u.scheduleRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 3. Get response counts for each schedule
	summaries := make([]ScheduleSummary, 0, len(schedules))
	for _, s := range schedules {
		// Get responses for this schedule
		responses, err := u.scheduleRepo.FindResponsesByScheduleID(ctx, s.ScheduleID())
		if err != nil {
			return nil, err
		}

		// Count unique member responses
		memberMap := make(map[string]bool)
		for _, resp := range responses {
			memberMap[resp.MemberID().String()] = true
		}

		var eventIDStr *string
		if s.EventID() != nil {
			str := s.EventID().String()
			eventIDStr = &str
		}

		var decidedCandidateIDStr *string
		if s.DecidedCandidateID() != nil {
			str := s.DecidedCandidateID().String()
			decidedCandidateIDStr = &str
		}

		summaries = append(summaries, ScheduleSummary{
			ScheduleID:         s.ScheduleID().String(),
			TenantID:           s.TenantID().String(),
			Title:              s.Title(),
			Description:        s.Description(),
			EventID:            eventIDStr,
			PublicToken:        s.PublicToken().String(),
			Status:             s.Status().String(),
			Deadline:           s.Deadline(),
			DecidedCandidateID: decidedCandidateIDStr,
			CandidateCount:     len(s.Candidates()),
			ResponseCount:      len(memberMap),
			CreatedAt:          s.CreatedAt(),
			UpdatedAt:          s.UpdatedAt(),
		})
	}

	return &ListSchedulesOutput{
		Schedules: summaries,
	}, nil
}
