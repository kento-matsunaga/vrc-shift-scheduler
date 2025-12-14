package attendance

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// AttendanceCollection は出欠確認の集約ルート
// MVP方針: responses は集約内で保持しない（Repository側UPSERTで管理）
type AttendanceCollection struct {
	collectionID common.CollectionID
	tenantID     common.TenantID
	title        string
	description  string
	targetType   TargetType
	targetID     string // event_id または business_day_id（オプション）
	publicToken  common.PublicToken
	status       Status
	deadline     *time.Time
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewAttendanceCollection creates a new AttendanceCollection entity
// NOTE: now は App層で clock.Now() を呼んで渡す（Domain層で time.Now() を呼ばない）
func NewAttendanceCollection(
	now time.Time,
	tenantID common.TenantID,
	title string,
	description string,
	targetType TargetType,
	targetID string,
	deadline *time.Time,
) (*AttendanceCollection, error) {
	collection := &AttendanceCollection{
		collectionID: common.NewCollectionID(),
		tenantID:     tenantID,
		title:        title,
		description:  description,
		targetType:   targetType,
		targetID:     targetID,
		publicToken:  common.NewPublicToken(),
		status:       StatusOpen,
		deadline:     deadline,
		createdAt:    now,
		updatedAt:    now,
	}

	if err := collection.validate(); err != nil {
		return nil, err
	}

	return collection, nil
}

// ReconstructAttendanceCollection reconstructs an AttendanceCollection entity from persistence
func ReconstructAttendanceCollection(
	collectionID common.CollectionID,
	tenantID common.TenantID,
	title string,
	description string,
	targetType TargetType,
	targetID string,
	publicToken common.PublicToken,
	status Status,
	deadline *time.Time,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*AttendanceCollection, error) {
	collection := &AttendanceCollection{
		collectionID: collectionID,
		tenantID:     tenantID,
		title:        title,
		description:  description,
		targetType:   targetType,
		targetID:     targetID,
		publicToken:  publicToken,
		status:       status,
		deadline:     deadline,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}

	if err := collection.validate(); err != nil {
		return nil, err
	}

	return collection, nil
}

func (c *AttendanceCollection) validate() error {
	// TenantID の必須性チェック
	if err := c.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	// Title の必須性チェック
	if c.title == "" {
		return common.NewValidationError("title is required", nil)
	}
	if len(c.title) > 255 {
		return common.NewValidationError("title must be less than 255 characters", nil)
	}

	// TargetType の検証
	if err := c.targetType.Validate(); err != nil {
		return err
	}

	// PublicToken の検証
	if err := c.publicToken.Validate(); err != nil {
		return err
	}

	// Status の検証
	if err := c.status.Validate(); err != nil {
		return err
	}

	return nil
}

// CanRespond は回答可能かを判定（ドメインルール）
// now は App層から Clock 経由で渡される
func (c *AttendanceCollection) CanRespond(now time.Time) error {
	if c.status != StatusOpen {
		return ErrCollectionClosed
	}
	if c.deadline != nil && now.After(*c.deadline) {
		return ErrDeadlinePassed
	}
	return nil
}

// Close はステータスをclosedに変更（ドメインルール）
// now は App層から Clock 経由で渡される
func (c *AttendanceCollection) Close(now time.Time) error {
	if c.status == StatusClosed {
		return ErrAlreadyClosed
	}
	c.status = StatusClosed
	c.updatedAt = now
	return nil
}

// Getters

func (c *AttendanceCollection) CollectionID() common.CollectionID {
	return c.collectionID
}

func (c *AttendanceCollection) TenantID() common.TenantID {
	return c.tenantID
}

func (c *AttendanceCollection) Title() string {
	return c.title
}

func (c *AttendanceCollection) Description() string {
	return c.description
}

func (c *AttendanceCollection) TargetType() TargetType {
	return c.targetType
}

func (c *AttendanceCollection) TargetID() string {
	return c.targetID
}

func (c *AttendanceCollection) PublicToken() common.PublicToken {
	return c.publicToken
}

func (c *AttendanceCollection) Status() Status {
	return c.status
}

func (c *AttendanceCollection) Deadline() *time.Time {
	return c.deadline
}

func (c *AttendanceCollection) CreatedAt() time.Time {
	return c.createdAt
}

func (c *AttendanceCollection) UpdatedAt() time.Time {
	return c.updatedAt
}

func (c *AttendanceCollection) DeletedAt() *time.Time {
	return c.deletedAt
}

func (c *AttendanceCollection) IsDeleted() bool {
	return c.deletedAt != nil
}
