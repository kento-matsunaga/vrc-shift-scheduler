package attendance

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// CollectionRoleAssignment represents the association between an attendance collection and a role
// This is used to restrict which members can respond to an attendance collection based on their roles
type CollectionRoleAssignment struct {
	collectionID common.CollectionID
	roleID       common.RoleID
	createdAt    time.Time
}

// NewCollectionRoleAssignment creates a new CollectionRoleAssignment
func NewCollectionRoleAssignment(
	now time.Time,
	collectionID common.CollectionID,
	roleID common.RoleID,
) (*CollectionRoleAssignment, error) {
	assignment := &CollectionRoleAssignment{
		collectionID: collectionID,
		roleID:       roleID,
		createdAt:    now,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

// ReconstructCollectionRoleAssignment reconstructs from persistence
func ReconstructCollectionRoleAssignment(
	collectionID common.CollectionID,
	roleID common.RoleID,
	createdAt time.Time,
) (*CollectionRoleAssignment, error) {
	assignment := &CollectionRoleAssignment{
		collectionID: collectionID,
		roleID:       roleID,
		createdAt:    createdAt,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (a *CollectionRoleAssignment) validate() error {
	if err := a.collectionID.Validate(); err != nil {
		return common.NewValidationError("collection_id is required", err)
	}
	if err := a.roleID.Validate(); err != nil {
		return common.NewValidationError("role_id is required", err)
	}
	return nil
}

// Getters

func (a *CollectionRoleAssignment) CollectionID() common.CollectionID {
	return a.collectionID
}

func (a *CollectionRoleAssignment) RoleID() common.RoleID {
	return a.roleID
}

func (a *CollectionRoleAssignment) CreatedAt() time.Time {
	return a.createdAt
}
