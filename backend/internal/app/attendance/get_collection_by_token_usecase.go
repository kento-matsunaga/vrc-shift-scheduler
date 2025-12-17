package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// GetCollectionByTokenUsecase handles getting a collection by public token
type GetCollectionByTokenUsecase struct {
	repo attendance.AttendanceCollectionRepository
}

// NewGetCollectionByTokenUsecase creates a new GetCollectionByTokenUsecase
func NewGetCollectionByTokenUsecase(
	repo attendance.AttendanceCollectionRepository,
) *GetCollectionByTokenUsecase {
	return &GetCollectionByTokenUsecase{
		repo: repo,
	}
}

// GetCollectionByTokenInput represents the input for getting a collection by token
type GetCollectionByTokenInput struct {
	PublicToken string
}

// Execute executes the get collection by token use case
func (u *GetCollectionByTokenUsecase) Execute(ctx context.Context, input GetCollectionByTokenInput) (*GetCollectionOutput, error) {
	// 1. Parse PublicToken
	token, err := common.ParsePublicToken(input.PublicToken)
	if err != nil {
		return nil, ErrCollectionNotFound
	}

	// 2. Find collection by token
	collection, err := u.repo.FindByToken(ctx, token)
	if err != nil {
		return nil, ErrCollectionNotFound
	}

	// 3. Find target dates
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

	// 4. Return output DTO
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


