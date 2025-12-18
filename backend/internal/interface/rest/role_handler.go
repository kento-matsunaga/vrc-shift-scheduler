package rest

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/role"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RoleHandler handles role-related HTTP requests
type RoleHandler struct {
	createRoleUsecase *role.CreateRoleUsecase
	updateRoleUsecase *role.UpdateRoleUsecase
	getRoleUsecase    *role.GetRoleUsecase
	listRolesUsecase  *role.ListRolesUsecase
	deleteRoleUsecase *role.DeleteRoleUsecase
}

// NewRoleHandler creates a new RoleHandler
func NewRoleHandler(dbPool *pgxpool.Pool) *RoleHandler {
	roleRepo := db.NewRoleRepository(dbPool)

	return &RoleHandler{
		createRoleUsecase: role.NewCreateRoleUsecase(roleRepo),
		updateRoleUsecase: role.NewUpdateRoleUsecase(roleRepo),
		getRoleUsecase:    role.NewGetRoleUsecase(roleRepo),
		listRolesUsecase:  role.NewListRolesUsecase(roleRepo),
		deleteRoleUsecase: role.NewDeleteRoleUsecase(roleRepo),
	}
}

// CreateRoleRequest represents the request body for creating a role
type CreateRoleRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

// UpdateRoleRequest represents the request body for updating a role
type UpdateRoleRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
}

// RoleResponse represents a role in API responses
type RoleResponse struct {
	RoleID       string `json:"role_id"`
	TenantID     string `json:"tenant_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Color        string `json:"color"`
	DisplayOrder int    `json:"display_order"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// CreateRole handles POST /api/v1/roles
func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// リクエストボディのパース
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name is required", nil)
		return
	}

	if len(req.Name) > 100 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name must be 100 characters or less", nil)
		return
	}

	// Usecase実行
	input := role.CreateRoleInput{
		TenantID:     tenantID.String(),
		Name:         req.Name,
		Description:  req.Description,
		Color:        req.Color,
		DisplayOrder: req.DisplayOrder,
	}

	output, err := h.createRoleUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("CreateRole error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to create role", nil)
		return
	}

	// レスポンス
	resp := RoleResponse{
		RoleID:       output.RoleID,
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

// UpdateRole handles PUT /api/v1/roles/{role_id}
func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// URLパラメータからrole_idを取得
	roleID := chi.URLParam(r, "role_id")
	if roleID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "role_id is required", nil)
		return
	}

	// リクエストボディのパース
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name is required", nil)
		return
	}

	if len(req.Name) > 100 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "name must be 100 characters or less", nil)
		return
	}

	// Usecase実行
	input := role.UpdateRoleInput{
		TenantID:     tenantID.String(),
		RoleID:       roleID,
		Name:         req.Name,
		Description:  req.Description,
		Color:        req.Color,
		DisplayOrder: req.DisplayOrder,
	}

	output, err := h.updateRoleUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("UpdateRole error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to update role", nil)
		return
	}

	// レスポンス
	resp := RoleResponse{
		RoleID:       output.RoleID,
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

// GetRole handles GET /api/v1/roles/{role_id}
func (h *RoleHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// URLパラメータからrole_idを取得
	roleID := chi.URLParam(r, "role_id")
	if roleID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "role_id is required", nil)
		return
	}

	// Usecase実行
	input := role.GetRoleInput{
		TenantID: tenantID.String(),
		RoleID:   roleID,
	}

	output, err := h.getRoleUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("GetRole error: %+v", err)
		writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Role not found", nil)
		return
	}

	// レスポンス
	resp := RoleResponse{
		RoleID:       output.Role.RoleID,
		TenantID:     output.Role.TenantID,
		Name:         output.Role.Name,
		Description:  output.Role.Description,
		Color:        output.Role.Color,
		DisplayOrder: output.Role.DisplayOrder,
		CreatedAt:    output.Role.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    output.Role.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusOK, resp)
}

// ListRoles handles GET /api/v1/roles
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// Usecase実行
	input := role.ListRolesInput{
		TenantID: tenantID.String(),
	}

	output, err := h.listRolesUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("ListRoles error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch roles", nil)
		return
	}

	// レスポンス構築
	roleResponses := make([]RoleResponse, 0, len(output.Roles))
	for _, r := range output.Roles {
		roleResponses = append(roleResponses, RoleResponse{
			RoleID:       r.RoleID,
			TenantID:     r.TenantID,
			Name:         r.Name,
			Description:  r.Description,
			Color:        r.Color,
			DisplayOrder: r.DisplayOrder,
			CreatedAt:    r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:    r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"roles": roleResponses,
	})
}

// DeleteRole handles DELETE /api/v1/roles/{role_id}
func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// URLパラメータからrole_idを取得
	roleID := chi.URLParam(r, "role_id")
	if roleID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "role_id is required", nil)
		return
	}

	// Usecase実行
	input := role.DeleteRoleInput{
		TenantID: tenantID.String(),
		RoleID:   roleID,
	}

	output, err := h.deleteRoleUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("DeleteRole error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to delete role", nil)
		return
	}

	// レスポンス
	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"role_id":    output.RoleID,
		"deleted_at": output.DeletedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
