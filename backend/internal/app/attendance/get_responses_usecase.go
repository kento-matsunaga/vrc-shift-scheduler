package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// GetResponsesUsecase handles getting all responses for a collection
type GetResponsesUsecase struct {
	repo attendance.AttendanceCollectionRepository
}

// NewGetResponsesUsecase creates a new GetResponsesUsecase
func NewGetResponsesUsecase(
	repo attendance.AttendanceCollectionRepository,
) *GetResponsesUsecase {
	return &GetResponsesUsecase{
		repo: repo,
	}
}

// Execute executes the get responses use case
func (u *GetResponsesUsecase) Execute(ctx context.Context, input GetResponsesInput) (*GetResponsesOutput, error) {
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

	// 3. Verify collection exists and tenant has access
	_, err = u.repo.FindByID(ctx, tenantID, collectionID)
	if err != nil {
		return nil, err
	}

	// 4. Find all responses for the collection
	responses, err := u.repo.FindResponsesByCollectionID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	// 5. Convert to DTOs
	responseDTOs := make([]ResponseDTO, 0, len(responses))
	for _, resp := range responses {
		responseDTOs = append(responseDTOs, ResponseDTO{
			ResponseID:  resp.ResponseID().String(),
			MemberID:    resp.MemberID().String(),
			Response:    resp.Response().String(),
			Note:        resp.Note(),
			RespondedAt: resp.RespondedAt(),
		})
	}

	// 6. Return output DTO
	return &GetResponsesOutput{
		CollectionID: collectionID.String(),
		Responses:    responseDTOs,
	}, nil
}
