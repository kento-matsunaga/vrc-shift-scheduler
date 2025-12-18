package role

import "time"

// CreateRoleInput represents the input for creating a role
type CreateRoleInput struct {
	TenantID     string // from JWT context (管理API)
	Name         string
	Description  string
	Color        string
	DisplayOrder int
}

// CreateRoleOutput represents the output for creating a role
type CreateRoleOutput struct {
	RoleID       string    `json:"role_id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Color        string    `json:"color"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// UpdateRoleInput represents the input for updating a role
type UpdateRoleInput struct {
	TenantID     string // from JWT context (管理API)
	RoleID       string
	Name         string
	Description  string
	Color        string
	DisplayOrder int
}

// UpdateRoleOutput represents the output for updating a role
type UpdateRoleOutput struct {
	RoleID       string    `json:"role_id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Color        string    `json:"color"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetRoleInput represents the input for getting a role
type GetRoleInput struct {
	TenantID string // from JWT context (管理API)
	RoleID   string
}

// RoleDTO represents a role in responses
type RoleDTO struct {
	RoleID       string    `json:"role_id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Color        string    `json:"color"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// GetRoleOutput represents the output for getting a role
type GetRoleOutput struct {
	Role RoleDTO `json:"role"`
}

// ListRolesInput represents the input for listing roles
type ListRolesInput struct {
	TenantID string // from JWT context (管理API)
}

// ListRolesOutput represents the output for listing roles
type ListRolesOutput struct {
	Roles []RoleDTO `json:"roles"`
}

// DeleteRoleInput represents the input for deleting a role
type DeleteRoleInput struct {
	TenantID string // from JWT context (管理API)
	RoleID   string
}

// DeleteRoleOutput represents the output for deleting a role
type DeleteRoleOutput struct {
	RoleID    string    `json:"role_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
