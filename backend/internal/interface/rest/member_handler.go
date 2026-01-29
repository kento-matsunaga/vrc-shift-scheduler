package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	appmember "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/go-chi/chi/v5"
)

// MemberHandler handles member-related HTTP requests
type MemberHandler struct {
	createMemberUC             *appmember.CreateMemberUsecase
	listMembersUC              *appmember.ListMembersUsecase
	getMemberUC                *appmember.GetMemberUsecase
	deleteMemberUC             *appmember.DeleteMemberUsecase
	updateMemberUsecase        *appmember.UpdateMemberUsecase
	getRecentAttendanceUsecase *appmember.GetRecentAttendanceUsecase
	bulkImportMembersUC        *appmember.BulkImportMembersUsecase
	bulkUpdateRolesUC          *appmember.BulkUpdateRolesUsecase
}

// NewMemberHandler creates a new MemberHandler with injected usecases
func NewMemberHandler(
	createMemberUC *appmember.CreateMemberUsecase,
	listMembersUC *appmember.ListMembersUsecase,
	getMemberUC *appmember.GetMemberUsecase,
	deleteMemberUC *appmember.DeleteMemberUsecase,
	updateMemberUC *appmember.UpdateMemberUsecase,
	getRecentAttendanceUC *appmember.GetRecentAttendanceUsecase,
	bulkImportMembersUC *appmember.BulkImportMembersUsecase,
	bulkUpdateRolesUC *appmember.BulkUpdateRolesUsecase,
) *MemberHandler {
	return &MemberHandler{
		createMemberUC:             createMemberUC,
		listMembersUC:              listMembersUC,
		getMemberUC:                getMemberUC,
		deleteMemberUC:             deleteMemberUC,
		updateMemberUsecase:        updateMemberUC,
		getRecentAttendanceUsecase: getRecentAttendanceUC,
		bulkImportMembersUC:        bulkImportMembersUC,
		bulkUpdateRolesUC:          bulkUpdateRolesUC,
	}
}

// CreateMemberRequest represents the request body for creating a member
type CreateMemberRequest struct {
	DisplayName   string   `json:"display_name"`
	DiscordUserID string   `json:"discord_user_id"`
	Email         string   `json:"email"`
	RoleIDs       []string `json:"role_ids,omitempty"` // Optional role IDs to assign
}

// UpdateMemberRequest represents the request body for updating a member
type UpdateMemberRequest struct {
	DisplayName   string   `json:"display_name"`
	DiscordUserID string   `json:"discord_user_id"`
	Email         string   `json:"email"`
	IsActive      bool     `json:"is_active"`
	RoleIDs       []string `json:"role_ids"` // Role IDs to assign
}

// MemberResponse represents a member in API responses
type MemberResponse struct {
	MemberID      string   `json:"member_id"`
	TenantID      string   `json:"tenant_id"`
	DisplayName   string   `json:"display_name"`
	DiscordUserID string   `json:"discord_user_id,omitempty"`
	Email         string   `json:"email,omitempty"`
	IsActive      bool     `json:"is_active"`
	RoleIDs       []string `json:"role_ids,omitempty"` // Assigned role IDs
	CreatedAt     string   `json:"created_at"`
	UpdatedAt     string   `json:"updated_at"`
}

// CreateMember handles POST /api/v1/members
func (h *MemberHandler) CreateMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// リクエストボディのパース
	var req CreateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "display_name is required", nil)
		return
	}

	if len(req.DisplayName) > 50 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "display_name must be 50 characters or less", nil)
		return
	}

	// Usecaseの実行
	input := appmember.CreateMemberInput{
		TenantID:      tenantID,
		DisplayName:   req.DisplayName,
		DiscordUserID: req.DiscordUserID,
		Email:         req.Email,
		RoleIDs:       req.RoleIDs,
	}

	newMember, err := h.createMemberUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス（リクエストで指定されたロールIDをそのまま返す）
	roleIDs := req.RoleIDs
	if roleIDs == nil {
		roleIDs = []string{}
	}
	resp := MemberResponse{
		MemberID:      newMember.MemberID().String(),
		TenantID:      newMember.TenantID().String(),
		DisplayName:   newMember.DisplayName(),
		DiscordUserID: newMember.DiscordUserID(),
		Email:         newMember.Email(),
		IsActive:      newMember.IsActive(),
		RoleIDs:       roleIDs,
		CreatedAt:     newMember.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     newMember.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusCreated, resp)
}

// UpdateMember handles PUT /api/v1/members/{member_id}
func (h *MemberHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// URLパラメータからmember_idを取得
	memberID := chi.URLParam(r, "member_id")
	if memberID == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "member_id is required", nil)
		return
	}

	// リクエストボディのパース
	var req UpdateMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if req.DisplayName == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "display_name is required", nil)
		return
	}

	if len(req.DisplayName) > 50 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "display_name must be 50 characters or less", nil)
		return
	}

	// Usecase実行
	input := appmember.UpdateMemberInput{
		TenantID:      tenantID.String(),
		MemberID:      memberID,
		DisplayName:   req.DisplayName,
		DiscordUserID: req.DiscordUserID,
		Email:         req.Email,
		IsActive:      req.IsActive,
		RoleIDs:       req.RoleIDs,
	}

	output, err := h.updateMemberUsecase.Execute(ctx, input)
	if err != nil {
		log.Printf("UpdateMember error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to update member", nil)
		return
	}

	// レスポンス（RoleIDsはUsecaseの出力から取得）
	resp := MemberResponse{
		MemberID:      output.MemberID,
		TenantID:      output.TenantID,
		DisplayName:   output.DisplayName,
		DiscordUserID: output.DiscordUserID,
		Email:         output.Email,
		IsActive:      output.IsActive,
		RoleIDs:       output.RoleIDs,
		CreatedAt:     "", // UpdatedAt is returned, not CreatedAt
		UpdatedAt:     output.UpdatedAt,
	}

	writeSuccess(w, http.StatusOK, resp)
}

// GetMembers handles GET /api/v1/members
func (h *MemberHandler) GetMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// クエリパラメータの取得
	isActiveStr := r.URL.Query().Get("is_active")

	// is_active フィルタのパース
	var isActive *bool
	if isActiveStr == "true" {
		val := true
		isActive = &val
	} else if isActiveStr == "false" {
		val := false
		isActive = &val
	}

	// Usecaseの実行
	input := appmember.ListMembersInput{
		TenantID: tenantID,
		IsActive: isActive,
	}

	membersWithRoles, err := h.listMembersUC.Execute(ctx, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch members", nil)
		return
	}

	// コンテキストから許可されたメンバーIDを取得（公開API用フィルター）
	var allowedMemberIDs map[string]struct{}
	if val := ctx.Value(ContextKeyAllowedMemberIDs); val != nil {
		allowedMemberIDs = val.(map[string]struct{})
	}

	// レスポンス構築
	memberResponses := make([]MemberResponse, 0, len(membersWithRoles))
	for _, mwr := range membersWithRoles {
		memberIDStr := mwr.Member.MemberID().String()

		// グループフィルタが設定されている場合、許可されたメンバーのみを返す
		if allowedMemberIDs != nil {
			if _, ok := allowedMemberIDs[memberIDStr]; !ok {
				continue
			}
		}

		// RoleIDをstringスライスに変換
		roleIDStrs := make([]string, len(mwr.RoleIDs))
		for i, roleID := range mwr.RoleIDs {
			roleIDStrs[i] = roleID.String()
		}

		memberResponses = append(memberResponses, MemberResponse{
			MemberID:      memberIDStr,
			TenantID:      mwr.Member.TenantID().String(),
			DisplayName:   mwr.Member.DisplayName(),
			DiscordUserID: mwr.Member.DiscordUserID(),
			Email:         mwr.Member.Email(),
			IsActive:      mwr.Member.IsActive(),
			RoleIDs:       roleIDStrs,
			CreatedAt:     mwr.Member.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     mwr.Member.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	writeSuccess(w, http.StatusOK, map[string]interface{}{
		"members": memberResponses,
		"count":   len(memberResponses),
	})
}

// GetMemberDetail handles GET /api/v1/members/:member_id
func (h *MemberHandler) GetMemberDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// member_id の取得
	memberIDStr := chi.URLParam(r, "member_id")
	if memberIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "member_id is required", nil)
		return
	}

	memberID, err := common.ParseMemberID(memberIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid member_id format", nil)
		return
	}

	// Usecaseの実行
	input := appmember.GetMemberInput{
		TenantID: tenantID,
		MemberID: memberID,
	}

	result, err := h.getMemberUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// RoleIDをstringスライスに変換
	roleIDStrs := make([]string, len(result.RoleIDs))
	for i, roleID := range result.RoleIDs {
		roleIDStrs[i] = roleID.String()
	}

	// レスポンス
	resp := MemberResponse{
		MemberID:      result.Member.MemberID().String(),
		TenantID:      result.Member.TenantID().String(),
		DisplayName:   result.Member.DisplayName(),
		DiscordUserID: result.Member.DiscordUserID(),
		Email:         result.Member.Email(),
		IsActive:      result.Member.IsActive(),
		RoleIDs:       roleIDStrs,
		CreatedAt:     result.Member.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     result.Member.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusOK, resp)
}

// GetRecentAttendance handles GET /api/v1/members/recent-attendance
func (h *MemberHandler) GetRecentAttendance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// クエリパラメータからlimitを取得（デフォルト10）
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 10
		}
	}

	// Usecaseの実行
	output, err := h.getRecentAttendanceUsecase.Execute(ctx, appmember.GetRecentAttendanceInput{
		TenantID: tenantID.String(),
		Limit:    limit,
	})
	if err != nil {
		log.Printf("GetRecentAttendance error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch recent attendance", nil)
		return
	}

	// レスポンス
	writeSuccess(w, http.StatusOK, output)
}

// DeleteMember handles DELETE /api/v1/members/{member_id}
func (h *MemberHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// member_id の取得
	memberIDStr := chi.URLParam(r, "member_id")
	if memberIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "member_id is required", nil)
		return
	}

	memberID, err := common.ParseMemberID(memberIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid member_id format", nil)
		return
	}

	// Usecaseの実行
	input := appmember.DeleteMemberInput{
		TenantID: tenantID,
		MemberID: memberID,
	}

	if err := h.deleteMemberUC.Execute(ctx, input); err != nil {
		RespondDomainError(w, err)
		return
	}

	// 成功レスポンス（No Content）
	w.WriteHeader(http.StatusNoContent)
}

// BulkImportMembersRequest represents the request body for bulk importing members
type BulkImportMembersRequest struct {
	Members []BulkImportMemberRequest `json:"members"`
}

// BulkImportMemberRequest represents a single member in bulk import request
type BulkImportMemberRequest struct {
	DisplayName string   `json:"display_name"`
	RoleIDs     []string `json:"role_ids,omitempty"`
}

// BulkImportMembers handles POST /api/v1/members/bulk-import
func (h *MemberHandler) BulkImportMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// リクエストボディのパース
	var req BulkImportMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if len(req.Members) == 0 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "members array is required and must not be empty", nil)
		return
	}

	// 上限チェック（DoS対策）
	if len(req.Members) > 100 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Maximum 100 members can be imported at once", nil)
		return
	}

	// Usecaseの入力を構築
	memberInputs := make([]appmember.BulkImportMemberInput, len(req.Members))
	for i, m := range req.Members {
		memberInputs[i] = appmember.BulkImportMemberInput{
			DisplayName: m.DisplayName,
			RoleIDs:     m.RoleIDs,
		}
	}

	input := appmember.BulkImportMembersInput{
		TenantID: tenantID,
		Members:  memberInputs,
	}

	// Usecaseの実行
	output, err := h.bulkImportMembersUC.Execute(ctx, input)
	if err != nil {
		log.Printf("BulkImportMembers error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to import members", nil)
		return
	}

	// レスポンス
	writeSuccess(w, http.StatusOK, output)
}

// BulkUpdateRolesRequest represents the request body for bulk updating roles
type BulkUpdateRolesRequest struct {
	MemberIDs     []string `json:"member_ids"`
	AddRoleIDs    []string `json:"add_role_ids,omitempty"`
	RemoveRoleIDs []string `json:"remove_role_ids,omitempty"`
}

// BulkUpdateRoles handles POST /api/v1/members/bulk-update-roles
func (h *MemberHandler) BulkUpdateRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// 管理者IDの取得（監査ログ用）
	adminID, ok := GetAdminID(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Admin ID is required", nil)
		return
	}

	// リクエストボディのパース
	var req BulkUpdateRolesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// バリデーション
	if len(req.MemberIDs) == 0 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "member_ids is required and must not be empty", nil)
		return
	}

	if len(req.MemberIDs) > appmember.MaxBulkUpdateMembers {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", fmt.Sprintf("Maximum %d members can be updated at once", appmember.MaxBulkUpdateMembers), nil)
		return
	}

	if len(req.AddRoleIDs) == 0 && len(req.RemoveRoleIDs) == 0 {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "At least one of add_role_ids or remove_role_ids is required", nil)
		return
	}

	// Usecaseの入力を構築
	input := appmember.BulkUpdateRolesInput{
		TenantID:      tenantID,
		AdminID:       adminID,
		MemberIDs:     req.MemberIDs,
		AddRoleIDs:    req.AddRoleIDs,
		RemoveRoleIDs: req.RemoveRoleIDs,
	}

	// Usecaseの実行
	output, err := h.bulkUpdateRolesUC.Execute(ctx, input)
	if err != nil {
		log.Printf("BulkUpdateRoles error: %+v", err)
		// バリデーションエラーの場合は400を返す
		if domainErr, ok := err.(*common.DomainError); ok && domainErr.Code() == common.ErrInvalidInput {
			writeError(w, http.StatusBadRequest, "ERR_VALIDATION", domainErr.Message, nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to update roles", nil)
		return
	}

	// レスポンス
	writeSuccess(w, http.StatusOK, output)
}
