package auth

import (
	"fmt"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Role represents an admin's role in the system
type Role string

const (
	RoleOwner   Role = "owner"   // 店長
	RoleManager Role = "manager" // 副店長
)

// NewRole creates a new Role from a string
func NewRole(role string) (Role, error) {
	r := Role(role)
	if err := r.Validate(); err != nil {
		return "", err
	}
	return r, nil
}

// Validate validates the role
func (r Role) Validate() error {
	switch r {
	case RoleOwner, RoleManager:
		return nil
	default:
		return common.NewValidationError(
			fmt.Sprintf("invalid role: must be 'owner' or 'manager', got: %s", r),
			nil,
		)
	}
}

func (r Role) String() string {
	return string(r)
}

// IsOwner returns true if the role is owner
func (r Role) IsOwner() bool {
	return r == RoleOwner
}

// IsManager returns true if the role is manager
func (r Role) IsManager() bool {
	return r == RoleManager
}
