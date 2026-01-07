package attendance

import (
	"context"
	"log"

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

	// 4. Get all members at once to avoid N+1 query
	members, err := u.memberRepo.FindByTenantID(ctx, collection.TenantID())
	if err != nil {
		log.Printf("[WARN] Failed to fetch members for tenant %s: %v", collection.TenantID().String(), err)
		members = []*member.Member{} // Continue with empty members
	}

	// Build member name map for O(1) lookup
	memberNameMap := make(map[string]string, len(members))
	for _, m := range members {
		memberNameMap[m.MemberID().String()] = m.DisplayName()
	}

	// 5. Convert to DTOs with member names
	responseDTOs := make([]PublicResponseDTO, 0, len(responses))
	for _, resp := range responses {
		memberIDStr := resp.MemberID().String()
		memberName := memberIDStr // Default to ID
		if name, ok := memberNameMap[memberIDStr]; ok {
			memberName = name
		} else {
			log.Printf("[WARN] Member %s not found in tenant %s", memberIDStr, collection.TenantID().String())
		}

		responseDTOs = append(responseDTOs, PublicResponseDTO{
			MemberID:      memberIDStr,
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
