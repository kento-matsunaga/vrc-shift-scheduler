package schedule

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

type DeleteScheduleUsecase struct {
	repo  schedule.DateScheduleRepository
	clock services.Clock
}

func NewDeleteScheduleUsecase(repo schedule.DateScheduleRepository, clk services.Clock) *DeleteScheduleUsecase {
	return &DeleteScheduleUsecase{repo: repo, clock: clk}
}

func (u *DeleteScheduleUsecase) Execute(ctx context.Context, input DeleteScheduleInput) (*DeleteScheduleOutput, error) {
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
	if err := sch.Delete(now); err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, sch); err != nil {
		return nil, err
	}

	return &DeleteScheduleOutput{
		ScheduleID: sch.ScheduleID().String(),
		Status:     sch.Status().String(),
		DeletedAt:  sch.DeletedAt(),
		UpdatedAt:  sch.UpdatedAt(),
	}, nil
}
