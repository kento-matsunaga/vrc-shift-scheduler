package role

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// RoleGroup represents a group for organizing roles
type RoleGroup struct {
	groupID      common.RoleGroupID
	tenantID     common.TenantID
	name         string
	description  string
	color        string
	displayOrder int
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
	roleIDs      []common.RoleID // Associated role IDs (loaded separately)
}

// NewRoleGroup creates a new RoleGroup
func NewRoleGroup(now time.Time, tenantID common.TenantID, name, description, color string, displayOrder int) (*RoleGroup, error) {
	g := &RoleGroup{
		groupID:      common.NewRoleGroupID(),
		tenantID:     tenantID,
		name:         name,
		description:  description,
		color:        color,
		displayOrder: displayOrder,
		createdAt:    now,
		updatedAt:    now,
	}

	if err := g.validate(); err != nil {
		return nil, err
	}

	return g, nil
}

// ReconstructRoleGroup reconstructs a RoleGroup from persistence
func ReconstructRoleGroup(
	groupID common.RoleGroupID,
	tenantID common.TenantID,
	name, description, color string,
	displayOrder int,
	createdAt, updatedAt time.Time,
	deletedAt *time.Time,
	roleIDs []common.RoleID,
) (*RoleGroup, error) {
	g := &RoleGroup{
		groupID:      groupID,
		tenantID:     tenantID,
		name:         name,
		description:  description,
		color:        color,
		displayOrder: displayOrder,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
		roleIDs:      roleIDs,
	}

	if err := g.validate(); err != nil {
		return nil, err
	}

	return g, nil
}

func (g *RoleGroup) validate() error {
	if g.name == "" {
		return common.NewValidationError("name is required", nil)
	}
	if len(g.name) > 100 {
		return common.NewValidationError("name must be 100 characters or less", nil)
	}
	if len(g.color) > 7 {
		return common.NewValidationError("color must be 7 characters or less", nil)
	}
	return nil
}

// Getters
func (g *RoleGroup) GroupID() common.RoleGroupID { return g.groupID }
func (g *RoleGroup) TenantID() common.TenantID   { return g.tenantID }
func (g *RoleGroup) Name() string                { return g.name }
func (g *RoleGroup) Description() string         { return g.description }
func (g *RoleGroup) Color() string               { return g.color }
func (g *RoleGroup) DisplayOrder() int           { return g.displayOrder }
func (g *RoleGroup) CreatedAt() time.Time        { return g.createdAt }
func (g *RoleGroup) UpdatedAt() time.Time        { return g.updatedAt }
func (g *RoleGroup) DeletedAt() *time.Time       { return g.deletedAt }
func (g *RoleGroup) RoleIDs() []common.RoleID    { return g.roleIDs }

// UpdateDetails updates the group's details
func (g *RoleGroup) UpdateDetails(now time.Time, name, description, color string, displayOrder int) error {
	// Validate before mutating using a temporary copy
	tmp := *g
	tmp.name = name
	tmp.description = description
	tmp.color = color
	tmp.displayOrder = displayOrder
	tmp.updatedAt = now
	if err := tmp.validate(); err != nil {
		return err
	}

	// Apply validated changes
	g.name = name
	g.description = description
	g.color = color
	g.displayOrder = displayOrder
	g.updatedAt = now
	return nil
}

// Delete marks the group as deleted
func (g *RoleGroup) Delete(now time.Time) {
	g.deletedAt = &now
	g.updatedAt = now
}
