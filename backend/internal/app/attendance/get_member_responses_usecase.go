package attendance

import (
	"context"
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// GetMemberResponsesUsecase handles getting a specific member's responses for a collection
type GetMemberResponsesUsecase struct {
	repo attendance.AttendanceCollectionRepository
}

// NewGetMemberResponsesUsecase creates a new GetMemberResponsesUsecase
func NewGetMemberResponsesUsecase(repo attendance.AttendanceCollectionRepository) *GetMemberResponsesUsecase {
	return &GetMemberResponsesUsecase{repo: repo}
}

// Execute retrieves a member's responses for a collection identified by public token
func (u *GetMemberResponsesUsecase) Execute(ctx context.Context, input GetMemberResponsesInput) (*GetMemberResponsesOutput, error) {
	// Parse public token
	publicToken, err := common.ParsePublicToken(input.PublicToken)
	if err != nil {
		return nil, common.NewValidationError("invalid public token", err)
	}

	// Parse member ID
	memberID, err := common.ParseMemberID(input.MemberID)
	if err != nil {
		return nil, common.NewValidationError("invalid member_id", err)
	}

	// Find collection by token
	collection, err := u.repo.FindByToken(ctx, publicToken)
	if err != nil {
		return nil, common.NewNotFoundError("AttendanceCollection", input.PublicToken)
	}

	// Get responses for this member in this collection
	responses, err := u.repo.FindResponsesByCollectionIDAndMemberID(ctx, collection.CollectionID(), memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to find member responses: %w", err)
	}

	// Convert to DTOs
	responseDTOs := make([]MemberResponseDTO, 0, len(responses))
	for _, resp := range responses {
		responseDTOs = append(responseDTOs, MemberResponseDTO{
			TargetDateID:  resp.TargetDateID().String(),
			Response:      string(resp.Response()),
			Note:          resp.Note(),
			AvailableFrom: resp.AvailableFrom(),
			AvailableTo:   resp.AvailableTo(),
		})
	}

	return &GetMemberResponsesOutput{
		MemberID:  memberID.String(),
		Responses: responseDTOs,
	}, nil
}
