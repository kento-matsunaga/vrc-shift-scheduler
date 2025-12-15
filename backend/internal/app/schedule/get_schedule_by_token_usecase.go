package schedule

import (
	"context"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
)

// GetScheduleByTokenUsecase handles getting a schedule by public token
type GetScheduleByTokenUsecase struct {
	repo schedule.DateScheduleRepository
}

// NewGetScheduleByTokenUsecase creates a new GetScheduleByTokenUsecase
func NewGetScheduleByTokenUsecase(
	repo schedule.DateScheduleRepository,
) *GetScheduleByTokenUsecase {
	return &GetScheduleByTokenUsecase{
		repo: repo,
	}
}

// GetScheduleByTokenInput represents the input for getting a schedule by token
type GetScheduleByTokenInput struct {
	PublicToken string
}

// Execute executes the get schedule by token use case
func (u *GetScheduleByTokenUsecase) Execute(ctx context.Context, input GetScheduleByTokenInput) (*GetScheduleOutput, error) {
	// 1. Parse PublicToken
	token, err := common.ParsePublicToken(input.PublicToken)
	if err != nil {
		return nil, ErrScheduleNotFound
	}

	// 2. Find schedule by token
	sched, err := u.repo.FindByToken(ctx, token)
	if err != nil {
		return nil, ErrScheduleNotFound
	}

	// 3. Get candidates
	candidates, err := u.repo.FindCandidatesByScheduleID(ctx, sched.ScheduleID())
	if err != nil {
		return nil, err
	}

	// 4. Convert to output
	candidateOutputs := make([]CandidateDTO, len(candidates))
	for i, c := range candidates {
		candidateOutputs[i] = CandidateDTO{
			CandidateID: c.CandidateID().String(),
			Date:        c.CandidateDateValue(),
			StartTime:   c.StartTime(),
			EndTime:     c.EndTime(),
		}
	}

	var eventID *string
	if sched.EventID() != nil {
		id := sched.EventID().String()
		eventID = &id
	}

	var decidedCandidateID *string
	if sched.DecidedCandidateID() != nil {
		id := sched.DecidedCandidateID().String()
		decidedCandidateID = &id
	}

	return &GetScheduleOutput{
		ScheduleID:         sched.ScheduleID().String(),
		TenantID:           sched.TenantID().String(),
		Title:              sched.Title(),
		Description:        sched.Description(),
		EventID:            eventID,
		PublicToken:        sched.PublicToken().String(),
		Status:             sched.Status().String(),
		Deadline:           sched.Deadline(),
		DecidedCandidateID: decidedCandidateID,
		Candidates:         candidateOutputs,
		CreatedAt:          sched.CreatedAt(),
		UpdatedAt:          sched.UpdatedAt(),
	}, nil
}

