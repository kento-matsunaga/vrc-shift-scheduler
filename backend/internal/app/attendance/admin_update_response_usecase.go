package attendance

import (
	"context"
	"errors"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
)

// AdminUpdateResponseUsecase handles admin updating an attendance response
// 管理者による出欠回答の更新（締め切りチェックをスキップ）
type AdminUpdateResponseUsecase struct {
	repo      attendance.AttendanceCollectionRepository
	txManager services.TxManager
	clock     services.Clock
}

// NewAdminUpdateResponseUsecase creates a new AdminUpdateResponseUsecase
func NewAdminUpdateResponseUsecase(
	repo attendance.AttendanceCollectionRepository,
	txManager services.TxManager,
	clock services.Clock,
) *AdminUpdateResponseUsecase {
	return &AdminUpdateResponseUsecase{
		repo:      repo,
		txManager: txManager,
		clock:     clock,
	}
}

// Execute executes the admin update response use case
// 締め切り後でも回答を更新可能（管理者権限）
func (u *AdminUpdateResponseUsecase) Execute(ctx context.Context, input AdminUpdateResponseInput) (*AdminUpdateResponseOutput, error) {
	// 1. Parse TenantID
	tenantID, err := common.ParseTenantID(input.TenantID)
	if err != nil {
		return nil, common.NewValidationError("invalid tenant_id", err)
	}

	// 2. Parse CollectionID
	collectionID, err := common.ParseCollectionID(input.CollectionID)
	if err != nil {
		return nil, common.NewValidationError("invalid collection_id", err)
	}

	// 3. Parse MemberID
	memberID, err := common.ParseMemberID(input.MemberID)
	if err != nil {
		return nil, common.NewValidationError("invalid member_id", err)
	}

	// 4. Parse TargetDateID
	targetDateID, err := common.ParseTargetDateID(input.TargetDateID)
	if err != nil {
		return nil, common.NewValidationError("invalid target_date_id", err)
	}

	// 5. Parse ResponseType
	responseType, err := attendance.NewResponseType(input.Response)
	if err != nil {
		return nil, err
	}

	// 6. Use transaction to ensure atomicity
	var output *AdminUpdateResponseOutput
	err = u.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// a. Find collection by ID (tenant scoped)
		collection, err := u.repo.FindByID(txCtx, tenantID, collectionID)
		if err != nil {
			var domainErr *common.DomainError
			if errors.As(err, &domainErr) && domainErr.Code() == common.ErrNotFound {
				return ErrCollectionNotFound
			}
			return err
		}

		// NOTE: 管理者による更新のため、CanRespond() をスキップ
		// 締め切り後でも回答を更新可能

		// b. Create AttendanceResponse entity
		now := u.clock.Now()
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

		// c. Upsert response (ON CONFLICT DO UPDATE)
		if err := u.repo.UpsertResponse(txCtx, response); err != nil {
			return err
		}

		// d. Build output
		output = &AdminUpdateResponseOutput{
			ResponseID:    response.ResponseID().String(),
			CollectionID:  response.CollectionID().String(),
			MemberID:      response.MemberID().String(),
			TargetDateID:  response.TargetDateID().String(),
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
