package schedule

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
)

type DecideScheduleUsecase struct {
	repo  schedule.DateScheduleRepository
	clock clock.Clock
}

func NewDecideScheduleUsecase(repo schedule.DateScheduleRepository, clk clock.Clock) *DecideScheduleUsecase {
	return &DecideScheduleUsecase{repo: repo, clock: clk}
}

func (u *DecideScheduleUsecase) Execute(ctx context.Context, input DecideScheduleInput) (*DecideScheduleOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	scheduleID, err := common.ParseScheduleID(input.ScheduleID)
	if err != nil {
		return nil, err
	}

	candidateID, err := common.ParseCandidateID(input.CandidateID)
	if err != nil {
		return nil, err
	}

	sch, err := u.repo.FindByID(ctx, tenantID, scheduleID)
	if err != nil {
		return nil, err
	}

	now := u.clock.Now()
	if err := sch.Decide(candidateID, now); err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, sch); err != nil {
		return nil, err
	}

	return &DecideScheduleOutput{
		ScheduleID:         sch.ScheduleID().String(),
		Status:             sch.Status().String(),
		DecidedCandidateID: sch.DecidedCandidateID().String(),
		UpdatedAt:          sch.UpdatedAt(),
	}, nil
}
