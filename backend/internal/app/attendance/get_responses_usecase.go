package attendance

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
)

// GetResponsesUsecase handles getting all responses for a collection
type GetResponsesUsecase struct {
	repo       attendance.AttendanceCollectionRepository
	memberRepo member.MemberRepository
}

// NewGetResponsesUsecase creates a new GetResponsesUsecase
func NewGetResponsesUsecase(
	repo attendance.AttendanceCollectionRepository,
	memberRepo member.MemberRepository,
) *GetResponsesUsecase {
	return &GetResponsesUsecase{
		repo:       repo,
		memberRepo: memberRepo,
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

	// 5. Find all target dates for this collection
	targetDates, err := u.repo.FindTargetDatesByCollectionID(ctx, collectionID)
	if err != nil {
		return nil, err
	}

	// Create a map of target_date_id -> target_date for quick lookup
	targetDateMap := make(map[string]*attendance.TargetDate)
	for _, td := range targetDates {
		targetDateMap[td.TargetDateID().String()] = td
	}

	// 6. Convert to DTOs with member names and target dates
	responseDTOs := make([]ResponseDTO, 0, len(responses))
	for _, resp := range responses {
		// Get member info
		memberInfo, err := u.memberRepo.FindByID(ctx, tenantID, resp.MemberID())
		if err != nil {
			// If member not found, use ID as name
			responseDTOs = append(responseDTOs, ResponseDTO{
				ResponseID:   resp.ResponseID().String(),
				MemberID:     resp.MemberID().String(),
				MemberName:   resp.MemberID().String(),
				TargetDateID: resp.TargetDateID().String(),
				TargetDate:   targetDateMap[resp.TargetDateID().String()].TargetDateValue(),
				Response:     resp.Response().String(),
				Note:         resp.Note(),
				RespondedAt:  resp.RespondedAt(),
			})
			continue
		}

		// Get target date info
		targetDate := targetDateMap[resp.TargetDateID().String()]
		if targetDate == nil {
			continue // Skip if target date not found
		}

		responseDTOs = append(responseDTOs, ResponseDTO{
			ResponseID:   resp.ResponseID().String(),
			MemberID:     resp.MemberID().String(),
			MemberName:   memberInfo.DisplayName(),
			TargetDateID: resp.TargetDateID().String(),
			TargetDate:   targetDate.TargetDateValue(),
			Response:     resp.Response().String(),
			Note:         resp.Note(),
			RespondedAt:  resp.RespondedAt(),
		})
	}

	// 7. Return output DTO
	return &GetResponsesOutput{
		CollectionID: collectionID.String(),
		Responses:    responseDTOs,
	}, nil
}
