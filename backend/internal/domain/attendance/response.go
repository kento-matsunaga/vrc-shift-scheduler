package attendance

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AttendanceResponse represents an attendance response entity
type AttendanceResponse struct {
	responseID    common.ResponseID
	tenantID      common.TenantID
	collectionID  common.CollectionID
	memberID      common.MemberID
	targetDateID  common.TargetDateID // 対象日ID
	response      ResponseType
	note          string
	availableFrom *string // 参加可能開始時間（HH:MM形式）
	availableTo   *string // 参加可能終了時間（HH:MM形式）
	respondedAt   time.Time
	createdAt     time.Time
	updatedAt     time.Time
}

// NewAttendanceResponse creates a new AttendanceResponse entity
// NOTE: now は App層で clock.Now() を呼んで渡す（Domain層で time.Now() を呼ばない）
func NewAttendanceResponse(
	now time.Time,
	collectionID common.CollectionID,
	tenantID common.TenantID,
	memberID common.MemberID,
	targetDateID common.TargetDateID,
	responseType ResponseType,
	note string,
	availableFrom *string,
	availableTo *string,
) (*AttendanceResponse, error) {
	response := &AttendanceResponse{
		responseID:    common.NewResponseID(),
		tenantID:      tenantID,
		collectionID:  collectionID,
		memberID:      memberID,
		targetDateID:  targetDateID,
		response:      responseType,
		note:          note,
		availableFrom: availableFrom,
		availableTo:   availableTo,
		respondedAt:   now,
		createdAt:     now,
		updatedAt:     now,
	}

	if err := response.validate(); err != nil {
		return nil, err
	}

	return response, nil
}

// ReconstructAttendanceResponse reconstructs an AttendanceResponse entity from persistence
func ReconstructAttendanceResponse(
	responseID common.ResponseID,
	tenantID common.TenantID,
	collectionID common.CollectionID,
	memberID common.MemberID,
	targetDateID common.TargetDateID,
	responseType ResponseType,
	note string,
	availableFrom *string,
	availableTo *string,
	respondedAt time.Time,
	createdAt time.Time,
	updatedAt time.Time,
) (*AttendanceResponse, error) {
	response := &AttendanceResponse{
		responseID:    responseID,
		tenantID:      tenantID,
		collectionID:  collectionID,
		memberID:      memberID,
		targetDateID:  targetDateID,
		response:      responseType,
		note:          note,
		availableFrom: availableFrom,
		availableTo:   availableTo,
		respondedAt:   respondedAt,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}

	if err := response.validate(); err != nil {
		return nil, err
	}

	return response, nil
}

func (r *AttendanceResponse) validate() error {
	// TenantID の必須性チェック
	if err := r.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// CollectionID の必須性チェック
	if err := r.collectionID.Validate(); err != nil {
		return common.NewValidationError("collection_id is required", err)
	}

	// MemberID の必須性チェック
	if err := r.memberID.Validate(); err != nil {
		return common.NewValidationError("member_id is required", err)
	}

	// TargetDateID の必須性チェック
	if err := r.targetDateID.Validate(); err != nil {
		return common.NewValidationError("target_date_id is required", err)
	}

	// ResponseType の検証
	if err := r.response.Validate(); err != nil {
		return err
	}

	return nil
}

// Getters

func (r *AttendanceResponse) ResponseID() common.ResponseID {
	return r.responseID
}

func (r *AttendanceResponse) TenantID() common.TenantID {
	return r.tenantID
}

func (r *AttendanceResponse) CollectionID() common.CollectionID {
	return r.collectionID
}

func (r *AttendanceResponse) MemberID() common.MemberID {
	return r.memberID
}

func (r *AttendanceResponse) TargetDateID() common.TargetDateID {
	return r.targetDateID
}

func (r *AttendanceResponse) Response() ResponseType {
	return r.response
}

func (r *AttendanceResponse) Note() string {
	return r.note
}

func (r *AttendanceResponse) AvailableFrom() *string {
	return r.availableFrom
}

func (r *AttendanceResponse) AvailableTo() *string {
	return r.availableTo
}

func (r *AttendanceResponse) RespondedAt() time.Time {
	return r.respondedAt
}

func (r *AttendanceResponse) CreatedAt() time.Time {
	return r.createdAt
}

func (r *AttendanceResponse) UpdatedAt() time.Time {
	return r.updatedAt
}
