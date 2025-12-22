package member

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// MemberGroup represents a member group entity
type MemberGroup struct {
	groupID      common.MemberGroupID
	tenantID     common.TenantID
	name         string
	description  string
	color        string
	displayOrder int
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// NewMemberGroup creates a new MemberGroup entity
func NewMemberGroup(
	now time.Time,
	tenantID common.TenantID,
	name string,
	description string,
	color string,
	displayOrder int,
) (*MemberGroup, error) {
	group := &MemberGroup{
		groupID:      common.NewMemberGroupID(),
		tenantID:     tenantID,
		name:         name,
		description:  description,
		color:        color,
		displayOrder: displayOrder,
		createdAt:    now,
		updatedAt:    now,
	}

	if err := group.validate(); err != nil {
		return nil, err
	}

	return group, nil
}

// ReconstructMemberGroup reconstructs a MemberGroup entity from persistence
func ReconstructMemberGroup(
	groupID common.MemberGroupID,
	tenantID common.TenantID,
	name string,
	description string,
	color string,
	displayOrder int,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) (*MemberGroup, error) {
	group := &MemberGroup{
		groupID:      groupID,
		tenantID:     tenantID,
		name:         name,
		description:  description,
		color:        color,
		displayOrder: displayOrder,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}

	if err := group.validate(); err != nil {
		return nil, err
	}

	return group, nil
}

func (g *MemberGroup) validate() error {
	if err := g.tenantID.Validate(); err != nil {
		return common.NewValidationError("tenant_id is required", err)
	}

	if g.name == "" {
		return common.NewValidationError("name is required", nil)
	}

	if len(g.name) > 100 {
		return common.NewValidationError("name must be less than 100 characters", nil)
	}

	if g.color != "" && len(g.color) > 7 {
		return common.NewValidationError("color must be 7 characters or less", nil)
	}

	return nil
}

// Getters

func (g *MemberGroup) GroupID() common.MemberGroupID {
	return g.groupID
}

func (g *MemberGroup) TenantID() common.TenantID {
	return g.tenantID
}

func (g *MemberGroup) Name() string {
	return g.name
}

func (g *MemberGroup) Description() string {
	return g.description
}

func (g *MemberGroup) Color() string {
	return g.color
}

func (g *MemberGroup) DisplayOrder() int {
	return g.displayOrder
}

func (g *MemberGroup) CreatedAt() time.Time {
	return g.createdAt
}

func (g *MemberGroup) UpdatedAt() time.Time {
	return g.updatedAt
}

func (g *MemberGroup) DeletedAt() *time.Time {
	return g.deletedAt
}

func (g *MemberGroup) IsDeleted() bool {
	return g.deletedAt != nil
}

// UpdateDetails updates group details
func (g *MemberGroup) UpdateDetails(name, description, color string, displayOrder int) error {
	g.name = name
	g.description = description
	g.color = color
	g.displayOrder = displayOrder
	g.updatedAt = time.Now()

	return g.validate()
}

// Delete marks the group as deleted (soft delete)
func (g *MemberGroup) Delete() {
	now := time.Now()
	g.deletedAt = &now
	g.updatedAt = now
}
