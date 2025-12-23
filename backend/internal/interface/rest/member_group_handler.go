package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/member_group"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemberGroupHandler handles member group-related HTTP requests
type MemberGroupHandler struct {
	createGroupUsecase  *member_group.CreateGroupUsecase
	updateGroupUsecase  *member_group.UpdateGroupUsecase
	getGroupUsecase     *member_group.GetGroupUsecase
	listGroupsUsecase   *member_group.ListGroupsUsecase
	deleteGroupUsecase  *member_group.DeleteGroupUsecase
	assignMembersUsecase *member_group.AssignMembersUsecase
}

// NewMemberGroupHandler creates a new MemberGroupHandler
func NewMemberGroupHandler(dbPool *pgxpool.Pool) *MemberGroupHandler {
	groupRepo := db.NewMemberGroupRepository(dbPool)

	return &MemberGroupHandler{
		createGroupUsecase:  member_group.NewCreateGroupUsecase(groupRepo),
		updateGroupUsecase:  member_group.NewUpdateGroupUsecase(groupRepo),
		getGroupUsecase:     member_group.NewGetGroupUsecase(groupRepo),
		listGroupsUsecase:   member_group.NewListGroupsUsecase(groupRepo),
		deleteGroupUsecase:  member_group.NewDeleteGroupUsecase(groupRepo),
		assignMembersUsecase: member_group.NewAssignMembersUsecase(groupRepo),
	}
}

// CreateGroupRequest represents the request body for creating a group
type CreateGroupRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

// UpdateGroupRequest represents the request body for updating a group
type UpdateGroupRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

// AssignMembersRequest represents the request body for assigning members
type AssignMembersRequest struct {
	MemberIDs []string `json:"member_ids"`
}

// GroupResponse represents a group in API responses
type GroupResponse struct {
	GroupID      string   `json:"group_id"`
	TenantID     string   `json:"tenant_id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Color        string   `json:"color"`
	DisplayOrder int      `json:"display_order"`
	MemberIDs    []string `json:"member_ids,omitempty"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

// CreateGroup handles POST /api/v1/member-groups
func (h *MemberGroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	var req CreateGroupRequest
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

	input := member_group.CreateGroupInput{
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

	resp := GroupResponse{
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

// UpdateGroup handles PUT /api/v1/member-groups/{group_id}
func (h *MemberGroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateGroupRequest
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

	input := member_group.UpdateGroupInput{
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

	resp := GroupResponse{
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

// GetGroup handles GET /api/v1/member-groups/{group_id}
func (h *MemberGroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
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

	input := member_group.GetGroupInput{
		TenantID: tenantID.String(),
		GroupID:  groupID,
	}

	output, err := h.getGroupUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("GetGroup error: %+v", err)
		writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Group not found", nil)
		return
	}

	resp := GroupResponse{
		GroupID:      output.Group.GroupID,
		TenantID:     output.Group.TenantID,
		Name:         output.Group.Name,
		Description:  output.Group.Description,
		Color:        output.Group.Color,
		DisplayOrder: output.Group.DisplayOrder,
		MemberIDs:    output.Group.MemberIDs,
		CreatedAt:    output.Group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    output.Group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusOK, resp)
}

// ListGroups handles GET /api/v1/member-groups
func (h *MemberGroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	input := member_group.ListGroupsInput{
		TenantID: tenantID.String(),
	}

	output, err := h.listGroupsUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("ListGroups error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch groups", nil)
		return
	}

	groupResponses := make([]GroupResponse, 0, len(output.Groups))
	for _, g := range output.Groups {
		groupResponses = append(groupResponses, GroupResponse{
			GroupID:      g.GroupID,
			TenantID:     g.TenantID,
			Name:         g.Name,
			Description:  g.Description,
			Color:        g.Color,
			DisplayOrder: g.DisplayOrder,
			MemberIDs:    g.MemberIDs,
			CreatedAt:    g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"groups": groupResponses,
	})
}

// DeleteGroup handles DELETE /api/v1/member-groups/{group_id}
func (h *MemberGroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
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

	input := member_group.DeleteGroupInput{
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

// AssignMembers handles PUT /api/v1/member-groups/{group_id}/members
func (h *MemberGroupHandler) AssignMembers(w http.ResponseWriter, r *http.Request) {
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

	var req AssignMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	input := member_group.AssignMembersInput{
		TenantID:  tenantID.String(),
		GroupID:   groupID,
		MemberIDs: req.MemberIDs,
	}

	output, err := h.assignMembersUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("AssignMembers error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to assign members", nil)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"group_id":   output.GroupID,
		"member_ids": output.MemberIDs,
	})
}
