package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/role_group"
	"github.com/go-chi/chi/v5"
)

// RoleGroupHandler handles role group-related HTTP requests
type RoleGroupHandler struct {
	createGroupUsecase  *role_group.CreateGroupUsecase
	updateGroupUsecase  *role_group.UpdateGroupUsecase
	getGroupUsecase     *role_group.GetGroupUsecase
	listGroupsUsecase   *role_group.ListGroupsUsecase
	deleteGroupUsecase  *role_group.DeleteGroupUsecase
	assignRolesUsecase  *role_group.AssignRolesUsecase
}

// NewRoleGroupHandler creates a new RoleGroupHandler with injected usecases
func NewRoleGroupHandler(
	createGroupUC *role_group.CreateGroupUsecase,
	updateGroupUC *role_group.UpdateGroupUsecase,
	getGroupUC *role_group.GetGroupUsecase,
	listGroupsUC *role_group.ListGroupsUsecase,
	deleteGroupUC *role_group.DeleteGroupUsecase,
	assignRolesUC *role_group.AssignRolesUsecase,
) *RoleGroupHandler {
	return &RoleGroupHandler{
		createGroupUsecase: createGroupUC,
		updateGroupUsecase: updateGroupUC,
		getGroupUsecase:    getGroupUC,
		listGroupsUsecase:  listGroupsUC,
		deleteGroupUsecase: deleteGroupUC,
		assignRolesUsecase: assignRolesUC,
	}
}

// CreateRoleGroupRequest represents the request body for creating a group
type CreateRoleGroupRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

// UpdateRoleGroupRequest represents the request body for updating a group
type UpdateRoleGroupRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

// AssignRolesRequest represents the request body for assigning roles
type AssignRolesRequest struct {
	RoleIDs []string `json:"role_ids"`
}

// RoleGroupResponse represents a group in API responses
type RoleGroupResponse struct {
	GroupID      string   `json:"group_id"`
	TenantID     string   `json:"tenant_id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Color        string   `json:"color"`
	DisplayOrder int      `json:"display_order"`
	RoleIDs      []string `json:"role_ids,omitempty"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// CreateGroup handles POST /api/v1/role-groups
func (h *RoleGroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	var req CreateRoleGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name is required", nil)
		return
	}

	if len(req.Name) > 100 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name must be 100 characters or less", nil)
		return
	}

	input := role_group.CreateGroupInput{
		TenantID:     tenantID.String(),
		Name:         req.Name,
		Description:  req.Description,
		Color:        req.Color,
		DisplayOrder: req.DisplayOrder,
	}

	output, err := h.createGroupUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("CreateGroup error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to create group", nil)
		return
	}

	resp := RoleGroupResponse{
		GroupID:      output.GroupID,
		TenantID:     output.TenantID,
		Name:         output.Name,
		Description:  output.Description,
		Color:        output.Color,
		DisplayOrder: output.DisplayOrder,
		CreatedAt:    output.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    output.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusCreated, resp)
}

// UpdateGroup handles PUT /api/v1/role-groups/{group_id}
func (h *RoleGroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	groupID := chi.URLParam(r, "group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "group_id is required", nil)
		return
	}

	var req UpdateRoleGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name is required", nil)
		return
	}

	if len(req.Name) > 100 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name must be 100 characters or less", nil)
		return
	}

	input := role_group.UpdateGroupInput{
		TenantID:     tenantID.String(),
		GroupID:      groupID,
		Name:         req.Name,
		Description:  req.Description,
		Color:        req.Color,
		DisplayOrder: req.DisplayOrder,
	}

	output, err := h.updateGroupUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("UpdateGroup error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to update group", nil)
		return
	}

	resp := RoleGroupResponse{
		GroupID:      output.GroupID,
		TenantID:     output.TenantID,
		Name:         output.Name,
		Description:  output.Description,
		Color:        output.Color,
		DisplayOrder: output.DisplayOrder,
		CreatedAt:    output.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    output.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusOK, resp)
}

// GetGroup handles GET /api/v1/role-groups/{group_id}
func (h *RoleGroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	groupID := chi.URLParam(r, "group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "group_id is required", nil)
		return
	}

	input := role_group.GetGroupInput{
		TenantID: tenantID.String(),
		GroupID:  groupID,
	}

	output, err := h.getGroupUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("GetGroup error: %+v", err)
		writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Group not found", nil)
		return
	}

	resp := RoleGroupResponse{
		GroupID:      output.Group.GroupID,
		TenantID:     output.Group.TenantID,
		Name:         output.Group.Name,
		Description:  output.Group.Description,
		Color:        output.Group.Color,
		DisplayOrder: output.Group.DisplayOrder,
		RoleIDs:      output.Group.RoleIDs,
		CreatedAt:    output.Group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    output.Group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusOK, resp)
}

// ListGroups handles GET /api/v1/role-groups
func (h *RoleGroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	input := role_group.ListGroupsInput{
		TenantID: tenantID.String(),
	}

	output, err := h.listGroupsUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("ListGroups error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch groups", nil)
		return
	}

	groupResponses := make([]RoleGroupResponse, 0, len(output.Groups))
	for _, g := range output.Groups {
		groupResponses = append(groupResponses, RoleGroupResponse{
			GroupID:      g.GroupID,
			TenantID:     g.TenantID,
			Name:         g.Name,
			Description:  g.Description,
			Color:        g.Color,
			DisplayOrder: g.DisplayOrder,
			RoleIDs:      g.RoleIDs,
			CreatedAt:    g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"groups": groupResponses,
	})
}

// DeleteGroup handles DELETE /api/v1/role-groups/{group_id}
func (h *RoleGroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	groupID := chi.URLParam(r, "group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "group_id is required", nil)
		return
	}

	input := role_group.DeleteGroupInput{
		TenantID: tenantID.String(),
		GroupID:  groupID,
	}

	output, err := h.deleteGroupUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("DeleteGroup error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to delete group", nil)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"group_id":   output.GroupID,
		"deleted_at": output.DeletedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}

// AssignRoles handles PUT /api/v1/role-groups/{group_id}/roles
func (h *RoleGroupHandler) AssignRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	groupID := chi.URLParam(r, "group_id")
	if groupID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "group_id is required", nil)
		return
	}

	var req AssignRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	input := role_group.AssignRolesInput{
		TenantID: tenantID.String(),
		GroupID:  groupID,
		RoleIDs:  req.RoleIDs,
	}

	output, err := h.assignRolesUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("AssignRoles error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to assign roles", nil)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"group_id": output.GroupID,
		"role_ids": output.RoleIDs,
	})
}
