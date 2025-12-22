package member_group

import "time"

// CreateGroupInput represents the input for creating a member group
type CreateGroupInput struct {
	TenantID     string
	Name         string
	Description  string
	Color        string
	DisplayOrder int
}

// CreateGroupOutput represents the output for creating a member group
type CreateGroupOutput struct {
	GroupID      string    `json:"group_id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Color        string    `json:"color"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpdateGroupInput represents the input for updating a member group
type UpdateGroupInput struct {
	TenantID     string
	GroupID      string
	Name         string
	Description  string
	Color        string
	DisplayOrder int
}

// UpdateGroupOutput represents the output for updating a member group
type UpdateGroupOutput struct {
	GroupID      string    `json:"group_id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Color        string    `json:"color"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetGroupInput represents the input for getting a member group
type GetGroupInput struct {
	TenantID string
	GroupID  string
}

// GroupDTO represents a member group in responses
type GroupDTO struct {
	GroupID      string    `json:"group_id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Color        string    `json:"color"`
	DisplayOrder int       `json:"display_order"`
	MemberIDs    []string  `json:"member_ids,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetGroupOutput represents the output for getting a member group
type GetGroupOutput struct {
	Group GroupDTO `json:"group"`
}

// ListGroupsInput represents the input for listing member groups
type ListGroupsInput struct {
	TenantID string
}

// ListGroupsOutput represents the output for listing member groups
type ListGroupsOutput struct {
	Groups []GroupDTO `json:"groups"`
}

// DeleteGroupInput represents the input for deleting a member group
type DeleteGroupInput struct {
	TenantID string
	GroupID  string
}

// DeleteGroupOutput represents the output for deleting a member group
type DeleteGroupOutput struct {
	GroupID   string    `json:"group_id"`
	DeletedAt time.Time `json:"deleted_at"`
}

// AssignMembersInput represents the input for assigning members to a group
type AssignMembersInput struct {
	TenantID  string
	GroupID   string
	MemberIDs []string
}

// AssignMembersOutput represents the output for assigning members to a group
type AssignMembersOutput struct {
	GroupID   string   `json:"group_id"`
	MemberIDs []string `json:"member_ids"`
}
