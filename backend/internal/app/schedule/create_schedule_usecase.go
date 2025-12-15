package schedule

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
)

type CreateScheduleUsecase struct {
	repo  schedule.DateScheduleRepository
	clock clock.Clock
}

func NewCreateScheduleUsecase(repo schedule.DateScheduleRepository, clk clock.Clock) *CreateScheduleUsecase {
	return &CreateScheduleUsecase{repo: repo, clock: clk}
}

func (u *CreateScheduleUsecase) Execute(ctx context.Context, input CreateScheduleInput) (*CreateScheduleOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	var eventID *common.EventID
	if input.EventID != nil {
		eid, err := common.ParseEventID(*input.EventID)
		if err != nil {
			return nil, err
		}
		eventID = &eid
	}

	now := u.clock.Now()
	scheduleID := common.NewScheduleID()

	// Create candidates
	candidates := make([]*schedule.CandidateDate, 0, len(input.Candidates))
	for i, c := range input.Candidates {
		candidate, err := schedule.NewCandidateDate(now, scheduleID, c.Date, c.StartTime, c.EndTime, i)
		if err != nil {
			return nil, err
		}
		candidates = append(candidates, candidate)
	}

	// Create schedule
	sch, err := schedule.NewDateSchedule(now, tenantID, input.Title, input.Description, eventID, candidates, input.Deadline)
	if err != nil {
		return nil, err
	}

	// Save
	if err := u.repo.Save(ctx, sch); err != nil {
		return nil, err
	}

	// Build output
	candidateDTOs := make([]CandidateDTO, len(candidates))
	for i, c := range candidates {
		candidateDTOs[i] = CandidateDTO{
			CandidateID: c.CandidateID().String(),
			Date:        c.CandidateDateValue(),
			StartTime:   c.StartTime(),
			EndTime:     c.EndTime(),
		}
	}

	var eventIDStr *string
	if eventID != nil {
		str := eventID.String()
		eventIDStr = &str
	}

	return &CreateScheduleOutput{
		ScheduleID:  sch.ScheduleID().String(),
		TenantID:    sch.TenantID().String(),
		Title:       sch.Title(),
		Description: sch.Description(),
		EventID:     eventIDStr,
		PublicToken: sch.PublicToken().String(),
		Status:      sch.Status().String(),
		Deadline:    sch.Deadline(),
		Candidates:  candidateDTOs,
		CreatedAt:   sch.CreatedAt(),
		UpdatedAt:   sch.UpdatedAt(),
	}, nil
}
