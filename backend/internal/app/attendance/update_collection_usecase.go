package attendance

import (
	"context"
	"fmt"
	"log"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

type UpdateCollectionUsecase struct {
	repo  attendance.AttendanceCollectionRepository
	clock services.Clock
}

func NewUpdateCollectionUsecase(repo attendance.AttendanceCollectionRepository, clk services.Clock) *UpdateCollectionUsecase {
	return &UpdateCollectionUsecase{repo: repo, clock: clk}
}

func (u *UpdateCollectionUsecase) Execute(ctx context.Context, input UpdateCollectionInput) (*UpdateCollectionOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant ID のパースに失敗: %w", err)
	}

	collectionID, err := common.ParseCollectionID(input.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("collection ID のパースに失敗: %w", err)
	}

	collection, err := u.repo.FindByID(ctx, tenantID, collectionID)
	if err != nil {
		return nil, fmt.Errorf("出欠確認の取得に失敗: %w", err)
	}

	now := u.clock.Now()
	if err := collection.Update(now, input.Title, input.Description, input.Deadline); err != nil {
		return nil, fmt.Errorf("出欠確認の更新に失敗: %w", err)
	}

	if err := u.repo.Save(ctx, collection); err != nil {
		return nil, fmt.Errorf("出欠確認の保存に失敗: %w", err)
	}

	log.Printf("[AUDIT] UpdateCollection: tenant=%s collection=%s", collection.TenantID().String(), collection.CollectionID().String())

	return &UpdateCollectionOutput{
		CollectionID: collection.CollectionID().String(),
		TenantID:     collection.TenantID().String(),
		Title:        collection.Title(),
		Description:  collection.Description(),
		Status:       collection.Status().String(),
		Deadline:     collection.Deadline(),
		UpdatedAt:    collection.UpdatedAt(),
	}, nil
}
