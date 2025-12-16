package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
)

// CreateCollectionUsecase handles creating an attendance collection
type CreateCollectionUsecase struct {
	repo  attendance.AttendanceCollectionRepository
	clock clock.Clock
}

// NewCreateCollectionUsecase creates a new CreateCollectionUsecase
func NewCreateCollectionUsecase(
	repo attendance.AttendanceCollectionRepository,
	clock clock.Clock,
) *CreateCollectionUsecase {
	return &CreateCollectionUsecase{
		repo:  repo,
		clock: clock,
	}
}

// Execute executes the create collection use case
func (u *CreateCollectionUsecase) Execute(ctx context.Context, input CreateCollectionInput) (*CreateCollectionOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// 2. Parse TargetType
	targetType, err := attendance.NewTargetType(input.TargetType)
	if err != nil {
		return nil, err
	}

	// 3. Create AttendanceCollection entity (Domain層)
	// Clock から now を取得して渡す
	now := u.clock.Now()
	collection, err := attendance.NewAttendanceCollection(
		now,
		tenantID,
		input.Title,
		input.Description,
		targetType,
		input.TargetID,
		input.Deadline,
	)
	if err != nil {
		return nil, err
	}

	// 4. Save to repository
	if err := u.repo.Save(ctx, collection); err != nil {
		return nil, err
	}

	// 5. Save target dates if provided
	if len(input.TargetDates) > 0 {
		var targetDates []*attendance.TargetDate
		for i, date := range input.TargetDates {
			td, err := attendance.NewTargetDate(now, collection.CollectionID(), date, i)
			if err != nil {
				return nil, err
			}
			targetDates = append(targetDates, td)
		}

		if err := u.repo.SaveTargetDates(ctx, collection.CollectionID(), targetDates); err != nil {
			return nil, err
		}
	}

	// 6. Return output DTO
	return &CreateCollectionOutput{
		CollectionID: collection.CollectionID().String(),
		TenantID:     collection.TenantID().String(),
		Title:        collection.Title(),
		Description:  collection.Description(),
		TargetType:   collection.TargetType().String(),
		TargetID:     collection.TargetID(),
		PublicToken:  collection.PublicToken().String(),
		Status:       collection.Status().String(),
		Deadline:     collection.Deadline(),
		CreatedAt:    collection.CreatedAt(),
		UpdatedAt:    collection.UpdatedAt(),
	}, nil
}
