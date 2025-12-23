package attendance

import (
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// CollectionGroupAssignment represents the association between an attendance collection and a member group
// This is used to restrict which members can respond to an attendance collection
type CollectionGroupAssignment struct {
	collectionID common.CollectionID
	groupID      common.MemberGroupID
	createdAt    time.Time
}

// NewCollectionGroupAssignment creates a new CollectionGroupAssignment
func NewCollectionGroupAssignment(
	now time.Time,
	collectionID common.CollectionID,
	groupID common.MemberGroupID,
) (*CollectionGroupAssignment, error) {
	assignment := &CollectionGroupAssignment{
		collectionID: collectionID,
		groupID:      groupID,
		createdAt:    now,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

// ReconstructCollectionGroupAssignment reconstructs from persistence
func ReconstructCollectionGroupAssignment(
	collectionID common.CollectionID,
	groupID common.MemberGroupID,
	createdAt time.Time,
) (*CollectionGroupAssignment, error) {
	assignment := &CollectionGroupAssignment{
		collectionID: collectionID,
		groupID:      groupID,
		createdAt:    createdAt,
	}

	if err := assignment.validate(); err != nil {
		return nil, err
	}

	return assignment, nil
}

func (a *CollectionGroupAssignment) validate() error {
	if err := a.collectionID.Validate(); err != nil {
		return common.NewValidationError("collection_id is required", err)
	}
	if err := a.groupID.Validate(); err != nil {
		return common.NewValidationError("group_id is required", err)
	}
	return nil
}

// Getters

func (a *CollectionGroupAssignment) CollectionID() common.CollectionID {
	return a.collectionID
}

func (a *CollectionGroupAssignment) GroupID() common.MemberGroupID {
	return a.groupID
}

func (a *CollectionGroupAssignment) CreatedAt() time.Time {
	return a.createdAt
}
