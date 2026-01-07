package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// GetAllPublicResponsesInput represents the input for getting all responses via public token
type GetAllPublicResponsesInput struct {
	PublicToken string // from URL path
}

// PublicResponseDTO represents a single response for the public table view
type PublicResponseDTO struct {
	MemberID      string  `json:"member_id"`
	MemberName    string  `json:"member_name"`
	TargetDateID  string  `json:"target_date_id"`
	Response      string  `json:"response"`
	Note          string  `json:"note"`
	AvailableFrom *string `json:"available_from,omitempty"`
	AvailableTo   *string `json:"available_to,omitempty"`
}

// GetAllPublicResponsesOutput represents the output for getting all responses
type GetAllPublicResponsesOutput struct {
	Responses []PublicResponseDTO `json:"responses"`
}

// GetAllPublicResponsesUsecase handles getting all responses for a collection via public token
type GetAllPublicResponsesUsecase struct {
	repo       attendance.AttendanceCollectionRepository
	memberRepo member.MemberRepository
}

// NewGetAllPublicResponsesUsecase creates a new GetAllPublicResponsesUsecase
func NewGetAllPublicResponsesUsecase(
	repo attendance.AttendanceCollectionRepository,
	memberRepo member.MemberRepository,
) *GetAllPublicResponsesUsecase {
	return &GetAllPublicResponsesUsecase{
		repo:       repo,
		memberRepo: memberRepo,
	}
}

// Execute retrieves all responses for a collection identified by public token
func (u *GetAllPublicResponsesUsecase) Execute(ctx context.Context, input GetAllPublicResponsesInput) (*GetAllPublicResponsesOutput, error) {
	// 1. Parse public token
	publicToken, err := common.ParsePublicToken(input.PublicToken)
	if err != nil {
		return nil, common.NewNotFoundError("AttendanceCollection", input.PublicToken)
	}

	// 2. Find collection by token
	collection, err := u.repo.FindByToken(ctx, publicToken)
	if err != nil {
		return nil, common.NewNotFoundError("AttendanceCollection", input.PublicToken)
	}

	// 3. Get all responses for this collection
	responses, err := u.repo.FindResponsesByCollectionID(ctx, collection.CollectionID())
	if err != nil {
		return nil, err
	}

	// 4. Convert to DTOs with member names
	responseDTOs := make([]PublicResponseDTO, 0, len(responses))
	for _, resp := range responses {
		// Get member info
		memberName := resp.MemberID().String() // Default to ID
		memberInfo, err := u.memberRepo.FindByID(ctx, collection.TenantID(), resp.MemberID())
		if err == nil {
			memberName = memberInfo.DisplayName()
		}

		responseDTOs = append(responseDTOs, PublicResponseDTO{
			MemberID:      resp.MemberID().String(),
			MemberName:    memberName,
			TargetDateID:  resp.TargetDateID().String(),
			Response:      resp.Response().String(),
			Note:          resp.Note(),
			AvailableFrom: resp.AvailableFrom(),
			AvailableTo:   resp.AvailableTo(),
		})
	}

	return &GetAllPublicResponsesOutput{
		Responses: responseDTOs,
	}, nil
}
