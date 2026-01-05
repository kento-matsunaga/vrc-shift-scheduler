package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

type DeleteCollectionUsecase struct {
	repo  attendance.AttendanceCollectionRepository
	clock services.Clock
}

func NewDeleteCollectionUsecase(repo attendance.AttendanceCollectionRepository, clk services.Clock) *DeleteCollectionUsecase {
	return &DeleteCollectionUsecase{repo: repo, clock: clk}
}

func (u *DeleteCollectionUsecase) Execute(ctx context.Context, input DeleteCollectionInput) (*DeleteCollectionOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	collectionID, err := common.ParseCollectionID(input.CollectionID)
	if err != nil {
		return nil, err
	}

	collection, err := u.repo.FindByID(ctx, tenantID, collectionID)
	if err != nil {
		return nil, err
	}

	now := u.clock.Now()
	if err := collection.Delete(now); err != nil {
		return nil, err
	}

	if err := u.repo.Save(ctx, collection); err != nil {
		return nil, err
	}

	return &DeleteCollectionOutput{
		CollectionID: collection.CollectionID().String(),
		Status:       collection.Status().String(),
		DeletedAt:    collection.DeletedAt(),
		UpdatedAt:    collection.UpdatedAt(),
	}, nil
}
