package schedule

import (
	"context"
	"log"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
)

// GetAllPublicResponsesInput represents the input for getting all responses via public token
type GetAllPublicResponsesInput struct {
	PublicToken string // from URL path
}

// PublicScheduleResponseDTO represents a single response for the public table view
type PublicScheduleResponseDTO struct {
	MemberID     string `json:"member_id"`
	MemberName   string `json:"member_name"`
	CandidateID  string `json:"candidate_id"`
	Availability string `json:"availability"`
	Note         string `json:"note"`
}

// GetAllPublicResponsesOutput represents the output for getting all responses
type GetAllPublicResponsesOutput struct {
	Responses []PublicScheduleResponseDTO `json:"responses"`
}

// GetAllPublicResponsesUsecase handles getting all responses for a schedule via public token
type GetAllPublicResponsesUsecase struct {
	repo       schedule.DateScheduleRepository
	memberRepo member.MemberRepository
}

// NewGetAllPublicResponsesUsecase creates a new GetAllPublicResponsesUsecase
func NewGetAllPublicResponsesUsecase(
	repo schedule.DateScheduleRepository,
	memberRepo member.MemberRepository,
) *GetAllPublicResponsesUsecase {
	return &GetAllPublicResponsesUsecase{
		repo:       repo,
		memberRepo: memberRepo,
	}
}

// Execute retrieves all responses for a schedule identified by public token
func (u *GetAllPublicResponsesUsecase) Execute(ctx context.Context, input GetAllPublicResponsesInput) (*GetAllPublicResponsesOutput, error) {
	// 1. Parse public token
	publicToken, err := common.ParsePublicToken(input.PublicToken)
	if err != nil {
		return nil, common.NewNotFoundError("DateSchedule", input.PublicToken)
	}

	// 2. Find schedule by token
	sched, err := u.repo.FindByToken(ctx, publicToken)
	if err != nil {
		return nil, common.NewNotFoundError("DateSchedule", input.PublicToken)
	}

	// 3. Get all responses for this schedule
	responses, err := u.repo.FindResponsesByScheduleID(ctx, sched.ScheduleID())
	if err != nil {
		return nil, err
	}

	// 4. Get all members at once to avoid N+1 query
	members, err := u.memberRepo.FindByTenantID(ctx, sched.TenantID())
	if err != nil {
		log.Printf("[WARN] Failed to fetch members for tenant %s: %v", sched.TenantID().String(), err)
		members = []*member.Member{} // Continue with empty members
	}

	// Build member name map for O(1) lookup
	memberNameMap := make(map[string]string, len(members))
	for _, m := range members {
		memberNameMap[m.MemberID().String()] = m.DisplayName()
	}

	// 5. Convert to DTOs with member names
	responseDTOs := make([]PublicScheduleResponseDTO, 0, len(responses))
	for _, resp := range responses {
		memberIDStr := resp.MemberID().String()
		memberName := memberIDStr // Default to ID
		if name, ok := memberNameMap[memberIDStr]; ok {
			memberName = name
		} else {
			log.Printf("[WARN] Member %s not found in tenant %s", memberIDStr, sched.TenantID().String())
		}

		responseDTOs = append(responseDTOs, PublicScheduleResponseDTO{
			MemberID:     memberIDStr,
			MemberName:   memberName,
			CandidateID:  resp.CandidateID().String(),
			Availability: resp.Availability().String(),
			Note:         resp.Note(),
		})
	}

	return &GetAllPublicResponsesOutput{
		Responses: responseDTOs,
	}, nil
}
