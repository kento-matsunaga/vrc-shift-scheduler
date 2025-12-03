package shift

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Position represents a position/role entity
// 役職（例: スタッフ、警備、受付）
type Position struct {
	positionID   PositionID
	tenantID     common.TenantID
	positionName string
	description  string
	displayOrder int
	isActive     bool
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewPosition creates a new Position entity
func NewPosition(
	tenantID common.TenantID,
	positionName string,
	description string,
	displayOrder int,
) (*Position, error) {
	position := &Position{
		positionID:   NewPositionID(),
		tenantID:     tenantID,
		positionName: positionName,
		description:  description,
		displayOrder: displayOrder,
		isActive:     true,
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
	}

	if err := position.validate(); err != nil {
		return nil, err
	}

	return position, nil
}

// ReconstructPosition reconstructs a Position from persistence
func ReconstructPosition(
	positionID PositionID,
	tenantID common.TenantID,
	positionName string,
	description string,
	displayOrder int,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*Position, error) {
	position := &Position{
		positionID:   positionID,
		tenantID:     tenantID,
		positionName: positionName,
		description:  description,
		displayOrder: displayOrder,
		isActive:     isActive,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}

	if err := position.validate(); err != nil {
		return nil, err
	}

	return position, nil
}

func (p *Position) validate() error {
	if err := p.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	if p.positionName == "" {
		return common.NewValidationError("position_name is required", nil)
	}

	if len(p.positionName) > 255 {
		return common.NewValidationError("position_name must be less than 255 characters", nil)
	}

	return nil
}

// Getters

func (p *Position) PositionID() PositionID {
	return p.positionID
}

func (p *Position) TenantID() common.TenantID {
	return p.tenantID
}

func (p *Position) PositionName() string {
	return p.positionName
}

func (p *Position) Description() string {
	return p.description
}

func (p *Position) DisplayOrder() int {
	return p.displayOrder
}

func (p *Position) IsActive() bool {
	return p.isActive
}

func (p *Position) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Position) UpdatedAt() time.Time {
	return p.updatedAt
}

func (p *Position) DeletedAt() *time.Time {
	return p.deletedAt
}

func (p *Position) IsDeleted() bool {
	return p.deletedAt != nil
}

// UpdatePositionName updates the position name
func (p *Position) UpdatePositionName(positionName string) error {
	if positionName == "" {
		return common.NewValidationError("position_name is required", nil)
	}
	if len(positionName) > 255 {
		return common.NewValidationError("position_name must be less than 255 characters", nil)
	}

	p.positionName = positionName
	p.updatedAt = time.Now()
	return nil
}

// UpdateDescription updates the description
func (p *Position) UpdateDescription(description string) {
	p.description = description
	p.updatedAt = time.Now()
}

// UpdateDisplayOrder updates the display order
func (p *Position) UpdateDisplayOrder(displayOrder int) {
	p.displayOrder = displayOrder
	p.updatedAt = time.Now()
}

// Activate activates the position
func (p *Position) Activate() {
	p.isActive = true
	p.updatedAt = time.Now()
}

// Deactivate deactivates the position
func (p *Position) Deactivate() {
	p.isActive = false
	p.updatedAt = time.Now()
}

// Delete marks the position as deleted (soft delete)
func (p *Position) Delete() {
	now := time.Now()
	p.deletedAt = &now
	p.updatedAt = now
}

