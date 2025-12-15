package schedule

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
)

type GetResponsesUsecase struct {
	repo schedule.DateScheduleRepository
}

func NewGetResponsesUsecase(repo schedule.DateScheduleRepository) *GetResponsesUsecase {
	return &GetResponsesUsecase{repo: repo}
}

func (u *GetResponsesUsecase) Execute(ctx context.Context, input GetResponsesInput) (*GetResponsesOutput, error) {
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, err
	}

	scheduleID, err := common.ParseScheduleID(input.ScheduleID)
	if err != nil {
		return nil, err
	}

	// Verify schedule exists and tenant has access
	_, err = u.repo.FindByID(ctx, tenantID, scheduleID)
	if err != nil {
		return nil, err
	}

	// Get responses
	responses, err := u.repo.FindResponsesByScheduleID(ctx, scheduleID)
	if err != nil {
		return nil, err
	}

	responseDTOs := make([]ScheduleResponseDTO, 0, len(responses))
	for _, resp := range responses {
		responseDTOs = append(responseDTOs, ScheduleResponseDTO{
			ResponseID:   resp.ResponseID().String(),
			MemberID:     resp.MemberID().String(),
			CandidateID:  resp.CandidateID().String(),
			Availability: resp.Availability().String(),
			Note:         resp.Note(),
			RespondedAt:  resp.RespondedAt(),
		})
	}

	return &GetResponsesOutput{
		ScheduleID: scheduleID.String(),
		Responses:  responseDTOs,
	}, nil
}
