package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// GetCollectionUsecase handles getting a single attendance collection
type GetCollectionUsecase struct {
	repo attendance.AttendanceCollectionRepository
}

// NewGetCollectionUsecase creates a new GetCollectionUsecase
func NewGetCollectionUsecase(
	repo attendance.AttendanceCollectionRepository,
) *GetCollectionUsecase {
	return &GetCollectionUsecase{
		repo: repo,
	}
}

// Execute executes the get collection use case
func (u *GetCollectionUsecase) Execute(ctx context.Context, input GetCollectionInput) (*GetCollectionOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	// 2. Parse CollectionID
	collectionID, err := common.ParseCollectionID(input.CollectionID)
	if err != nil {
		return nil, err
	}

	// 3. Find collection by ID (with tenant isolation)
	collection, err := u.repo.FindByID(ctx, tenantID, collectionID)
	if err != nil {
		return nil, err
	}

	// 4. Find target dates
	targetDates, err := u.repo.FindTargetDatesByCollectionID(ctx, collection.CollectionID())
	if err != nil {
		return nil, err
	}

	// Convert target dates to TargetDateDTO array
	var targetDateDTOs []TargetDateDTO
	for _, td := range targetDates {
		targetDateDTOs = append(targetDateDTOs, TargetDateDTO{
			TargetDateID: td.TargetDateID().String(),
			TargetDate:   td.TargetDateValue(),
			DisplayOrder: td.DisplayOrder(),
		})
	}

	// 5. Return output DTO
	return &GetCollectionOutput{
		CollectionID: collection.CollectionID().String(),
		TenantID:     collection.TenantID().String(),
		Title:        collection.Title(),
		Description:  collection.Description(),
		TargetType:   collection.TargetType().String(),
		TargetID:     collection.TargetID(),
		TargetDates:  targetDateDTOs,
		PublicToken:  collection.PublicToken().String(),
		Status:       collection.Status().String(),
		Deadline:     collection.Deadline(),
		CreatedAt:    collection.CreatedAt(),
		UpdatedAt:    collection.UpdatedAt(),
	}, nil
}
