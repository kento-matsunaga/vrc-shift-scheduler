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
		return nil, fmt.Errorf("failed to parse tenant ID: %w", err)
	}

	collectionID, err := common.ParseCollectionID(input.CollectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to parse collection ID: %w", err)
	}

	collection, err := u.repo.FindByID(ctx, tenantID, collectionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find collection: %w", err)
	}

	now := u.clock.Now()
	if err := collection.Update(now, input.Title, input.Description, input.Deadline); err != nil {
		return nil, fmt.Errorf("failed to update collection: %w", err)
	}

	if input.TargetDates != nil {
		err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
			if err := u.repo.Save(txCtx, collection); err != nil {
				return fmt.Errorf("failed to save collection: %w", err)
			}

			existingDates, err := u.repo.FindTargetDatesByCollectionID(txCtx, collectionID)
			if err != nil {
				return fmt.Errorf("failed to find existing target dates: %w", err)
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
						return fmt.Errorf("failed to update target date: %w", err)
					}
					allDates = append(allDates, existingTD)
				} else {
					newTD, err := attendance.NewTargetDate(now, collectionID, tdInput.TargetDate, tdInput.StartTime, tdInput.EndTime, i)
					if err != nil {
						return fmt.Errorf("failed to create target date: %w", err)
					}
					allDates = append(allDates, newTD)
				}
			}

			if err := u.repo.ReplaceTargetDates(txCtx, collectionID, allDates); err != nil {
				return fmt.Errorf("failed to replace target dates: %w", err)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		if err := u.repo.Save(ctx, collection); err != nil {
			return nil, fmt.Errorf("failed to save collection: %w", err)
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
