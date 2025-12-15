package schedule

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
)

type CloseScheduleUsecase struct {
	repo  schedule.DateScheduleRepository
	clock clock.Clock
}

func NewCloseScheduleUsecase(repo schedule.DateScheduleRepository, clk clock.Clock) *CloseScheduleUsecase {
	return &CloseScheduleUsecase{repo: repo, clock: clk}
}

func (u *CloseScheduleUsecase) Execute(ctx context.Context, input CloseScheduleInput) (*CloseScheduleOutput, error) {
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

	now := u.clock.Now()
	if err := sch.Close(now); err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, sch); err != nil {
		return nil, err
	}

	return &CloseScheduleOutput{
		ScheduleID: sch.ScheduleID().String(),
		Status:     sch.Status().String(),
		UpdatedAt:  sch.UpdatedAt(),
	}, nil
}
