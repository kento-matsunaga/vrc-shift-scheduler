package role_group

import "time"

// CreateGroupInput is input for creating a group
type CreateGroupInput struct {
	TenantID     string
	Name         string
	Description  string
	Color        string
	DisplayOrder int
}

// CreateGroupOutput is output for creating a group
type CreateGroupOutput struct {
	GroupID      string
	TenantID     string
	Name         string
	Description  string
	Color        string
	DisplayOrder int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UpdateGroupInput is input for updating a group
type UpdateGroupInput struct {
	TenantID     string
	GroupID      string
	Name         string
	Description  string
	Color        string
	DisplayOrder int
}

// UpdateGroupOutput is output for updating a group
type UpdateGroupOutput struct {
	GroupID      string
	TenantID     string
	Name         string
	Description  string
	Color        string
	DisplayOrder int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GetGroupInput is input for getting a group
type GetGroupInput struct {
	TenantID string
	GroupID  string
}

// GroupDTO represents a group in output
type GroupDTO struct {
	GroupID      string
	TenantID     string
	Name         string
	Description  string
	Color        string
	DisplayOrder int
	RoleIDs      []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// GetGroupOutput is output for getting a group
type GetGroupOutput struct {
	Group GroupDTO
}

// ListGroupsInput is input for listing groups
type ListGroupsInput struct {
	TenantID string
}

// ListGroupsOutput is output for listing groups
type ListGroupsOutput struct {
	Groups []GroupDTO
}

// DeleteGroupInput is input for deleting a group
type DeleteGroupInput struct {
	TenantID string
	GroupID  string
}

// DeleteGroupOutput is output for deleting a group
type DeleteGroupOutput struct {
	GroupID   string
	DeletedAt time.Time
}

// AssignRolesInput is input for assigning roles to a group
type AssignRolesInput struct {
	TenantID string
	GroupID  string
	RoleIDs  []string
}

// AssignRolesOutput is output for assigning roles to a group
type AssignRolesOutput struct {
	GroupID string
	RoleIDs []string
}
