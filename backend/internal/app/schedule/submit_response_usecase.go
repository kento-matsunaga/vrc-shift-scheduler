package schedule

import (
	"context"
	"errors"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

type SubmitResponseUsecase struct {
	repo      schedule.DateScheduleRepository
	txManager services.TxManager
	clock     services.Clock
}

func NewSubmitResponseUsecase(repo schedule.DateScheduleRepository, txManager services.TxManager, clk services.Clock) *SubmitResponseUsecase {
	return &SubmitResponseUsecase{repo: repo, txManager: txManager, clock: clk}
}

func (u *SubmitResponseUsecase) Execute(ctx context.Context, input SubmitResponseInput) (*SubmitResponseOutput, error) {
	publicToken, err := common.ParsePublicToken(input.PublicToken)
	if err != nil {
		return nil, ErrScheduleNotFound
	}

	memberID, err := common.ParseMemberID(input.MemberID)
	if err != nil {
		return nil, ErrMemberNotAllowed
	}

	var output *SubmitResponseOutput
	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		sch, err := u.repo.FindByToken(txCtx, publicToken)
		if err != nil {
			var domainErr *common.DomainError
			if errors.As(err, &domainErr) && domainErr.Code() == common.ErrNotFound {
				return ErrScheduleNotFound
			}
			return err
		}

		now := u.clock.Now()
		if err := sch.CanRespond(now); err != nil {
			return err
		}

		// Validate all candidates exist
		validCandidates := make(map[string]bool)
		for _, c := range sch.Candidates() {
			validCandidates[c.CandidateID().String()] = true
		}

		// Upsert each response
		for _, resp := range input.Responses {
			if !validCandidates[resp.CandidateID] {
				continue
			}

			candidateID, _ := common.ParseCandidateID(resp.CandidateID)
			availability, err := schedule.NewAvailability(resp.Availability)
			if err != nil {
				return err
			}

			response, err := schedule.NewDateScheduleResponse(now, sch.ScheduleID(), sch.TenantID(), memberID, candidateID, availability, resp.Note)
			if err != nil {
				return err
			}

			if err := u.repo.UpsertResponse(txCtx, response); err != nil {
				return err
			}
		}

		output = &SubmitResponseOutput{
			ScheduleID:  sch.ScheduleID().String(),
			MemberID:    memberID.String(),
			RespondedAt: now,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return output, nil
}
