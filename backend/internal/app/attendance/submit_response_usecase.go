package attendance

import (
	"context"
	"errors"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// SubmitResponseUsecase handles submitting an attendance response
type SubmitResponseUsecase struct {
	repo      attendance.AttendanceCollectionRepository
	txManager services.TxManager
	clock     services.Clock
}

// NewSubmitResponseUsecase creates a new SubmitResponseUsecase
func NewSubmitResponseUsecase(
	repo attendance.AttendanceCollectionRepository,
	txManager services.TxManager,
	clock services.Clock,
) *SubmitResponseUsecase {
	return &SubmitResponseUsecase{
		repo:      repo,
		txManager: txManager,
		clock:     clock,
	}
}

// Execute executes the submit response use case
func (u *SubmitResponseUsecase) Execute(ctx context.Context, input SubmitResponseInput) (*SubmitResponseOutput, error) {
	// 1. Parse PublicToken
	publicToken, err := common.ParsePublicToken(input.PublicToken)
	if err != nil {
		// トークンエラー → 404
		return nil, ErrCollectionNotFound
	}

	// 2. Parse MemberID
	memberID, err := common.ParseMemberID(input.MemberID)
	if err != nil {
		// メンバーエラー → 400（詳細は返さない）
		return nil, ErrMemberNotAllowed
	}

	// 3. Parse TargetDateID
	targetDateID, err := common.ParseTargetDateID(input.TargetDateID)
	if err != nil {
		return nil, common.NewValidationError("対象日IDが無効です", err)
	}

	// 4. Parse ResponseType
	responseType, err := attendance.NewResponseType(input.Response)
	if err != nil {
		return nil, err
	}

	// 4. Use transaction to ensure atomicity
	var output *SubmitResponseOutput
	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// a. Find collection by token
		collection, err := u.repo.FindByToken(txCtx, publicToken)
		if err != nil {
			// NotFoundError → ErrCollectionNotFound (404)
			var domainErr *common.DomainError
			if errors.As(err, &domainErr) && domainErr.Code() == common.ErrNotFound {
				return ErrCollectionNotFound
			}
			return err
		}

		// b. Check if response is allowed (domain rule)
		now := u.clock.Now()
		if err := collection.CanRespond(now); err != nil {
			// ErrCollectionClosed or ErrDeadlinePassed
			return err
		}

		// c. Create AttendanceResponse entity
		response, err := attendance.NewAttendanceResponse(
			now,
			collection.CollectionID(),
			collection.TenantID(),
			memberID,
			targetDateID,
			responseType,
			input.Note,
			input.AvailableFrom,
			input.AvailableTo,
		)
		if err != nil {
			return err
		}

		// d. Upsert response (ON CONFLICT DO UPDATE)
		if err := u.repo.UpsertResponse(txCtx, response); err != nil {
			return err
		}

		// e. Build output
		output = &SubmitResponseOutput{
			ResponseID:    response.ResponseID().String(),
			CollectionID:  response.CollectionID().String(),
			MemberID:      response.MemberID().String(),
			Response:      response.Response().String(),
			Note:          response.Note(),
			AvailableFrom: response.AvailableFrom(),
			AvailableTo:   response.AvailableTo(),
			RespondedAt:   response.RespondedAt(),
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
