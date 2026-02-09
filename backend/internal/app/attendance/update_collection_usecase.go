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
	repo      attendance.AttendanceCollectionRepository
	txManager services.TxManager
	clock     services.Clock
}

func NewUpdateCollectionUsecase(repo attendance.AttendanceCollectionRepository, txManager services.TxManager, clk services.Clock) *UpdateCollectionUsecase {
	return &UpdateCollectionUsecase{repo: repo, txManager: txManager, clock: clk}
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

	if input.TargetDates != nil {
		err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
			if err := u.repo.Save(txCtx, collection); err != nil {
				return fmt.Errorf("出欠確認の保存に失敗: %w", err)
			}

			existingDates, err := u.repo.FindTargetDatesByCollectionID(txCtx, collectionID)
			if err != nil {
				return fmt.Errorf("既存対象日の取得に失敗: %w", err)
			}
			existingMap := make(map[string]*attendance.TargetDate, len(existingDates))
			for _, ed := range existingDates {
				existingMap[ed.TargetDateID().String()] = ed
			}

			allDates := make([]*attendance.TargetDate, 0, len(input.TargetDates))
			for i, tdInput := range input.TargetDates {
				if tdInput.TargetDateID != "" {
					existingTD, ok := existingMap[tdInput.TargetDateID]
					if !ok {
						return common.NewValidationError("invalid target_date_id included", nil)
					}
					if err := existingTD.UpdateFields(tdInput.TargetDate, tdInput.StartTime, tdInput.EndTime, i); err != nil {
						return fmt.Errorf("対象日の更新に失敗: %w", err)
					}
					allDates = append(allDates, existingTD)
				} else {
					newTD, err := attendance.NewTargetDate(now, collectionID, tdInput.TargetDate, tdInput.StartTime, tdInput.EndTime, i)
					if err != nil {
						return fmt.Errorf("対象日の作成に失敗: %w", err)
					}
					allDates = append(allDates, newTD)
				}
			}

			if err := u.repo.ReplaceTargetDates(txCtx, collectionID, allDates); err != nil {
				return fmt.Errorf("対象日の差分更新に失敗: %w", err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := u.repo.Save(ctx, collection); err != nil {
			return nil, fmt.Errorf("出欠確認の保存に失敗: %w", err)
		}
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
