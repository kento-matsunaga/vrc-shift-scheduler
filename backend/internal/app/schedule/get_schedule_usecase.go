package schedule

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
)

type GetScheduleUsecase struct {
	repo schedule.DateScheduleRepository
}

func NewGetScheduleUsecase(repo schedule.DateScheduleRepository) *GetScheduleUsecase {
	return &GetScheduleUsecase{repo: repo}
}

func (u *GetScheduleUsecase) Execute(ctx context.Context, input GetScheduleInput) (*GetScheduleOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	scheduleID, err := common.ParseScheduleID(input.ScheduleID)
	if err != nil {
		return nil, err
	}

	sch, err := u.repo.FindByID(ctx, tenantID, scheduleID)
	if err != nil {
		return nil, err
	}

	candidateDTOs := make([]CandidateDTO, len(sch.Candidates()))
	for i, c := range sch.Candidates() {
		candidateDTOs[i] = CandidateDTO{
			CandidateID: c.CandidateID().String(),
			Date:        c.CandidateDateValue(),
			StartTime:   c.StartTime(),
			EndTime:     c.EndTime(),
		}
	}

	var eventIDStr *string
	if sch.EventID() != nil {
		str := sch.EventID().String()
		eventIDStr = &str
	}

	var decidedCandidateIDStr *string
	if sch.DecidedCandidateID() != nil {
		str := sch.DecidedCandidateID().String()
		decidedCandidateIDStr = &str
	}

	return &GetScheduleOutput{
		ScheduleID:         sch.ScheduleID().String(),
		TenantID:           sch.TenantID().String(),
		Title:              sch.Title(),
		Description:        sch.Description(),
		EventID:            eventIDStr,
		PublicToken:        sch.PublicToken().String(),
		Status:             sch.Status().String(),
		Deadline:           sch.Deadline(),
		DecidedCandidateID: decidedCandidateIDStr,
		Candidates:         candidateDTOs,
		CreatedAt:          sch.CreatedAt(),
		UpdatedAt:          sch.UpdatedAt(),
	}, nil
}
