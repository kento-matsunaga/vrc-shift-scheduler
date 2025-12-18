package rest

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	domainMember "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/member"
	appMember "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/member"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MemberHandler handles member-related HTTP requests
type MemberHandler struct {
	memberRepo                 *db.MemberRepository
	memberRoleRepo             *db.MemberRoleRepository
	updateMemberUsecase        *appMember.UpdateMemberUsecase
	getRecentAttendanceUsecase *appMember.GetRecentAttendanceUsecase
}

// NewMemberHandler creates a new MemberHandler
func NewMemberHandler(dbPool *pgxpool.Pool) *MemberHandler {
	memberRepo := db.NewMemberRepository(dbPool)
	memberRoleRepo := db.NewMemberRoleRepository(dbPool)
	attendanceRepo := db.NewAttendanceRepository(dbPool)

	return &MemberHandler{
		memberRepo:                 memberRepo,
		memberRoleRepo:             memberRoleRepo,
		updateMemberUsecase:        appMember.NewUpdateMemberUsecase(memberRepo, memberRoleRepo),
		getRecentAttendanceUsecase: appMember.NewGetRecentAttendanceUsecase(memberRepo, attendanceRepo),
	}
}

// CreateMemberRequest represents the request body for creating a member
type CreateMemberRequest struct {
	DisplayName   string `json:"display_name"`
	DiscordUserID string `json:"discord_user_id"`
	Email         string `json:"email"`
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

	// 重複チェック（discord_user_id）
	if req.DiscordUserID != "" {
		existing, err := h.memberRepo.FindByDiscordUserID(ctx, tenantID, req.DiscordUserID)
		if err != nil && err.Error() != "member not found" {
			writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to check discord_user_id duplication", nil)
			return
		}
		if existing != nil {
			writeError(w, http.StatusConflict, "ERR_CONFLICT", "This discord_user_id is already registered", map[string]interface{}{
				"member_id": existing.MemberID().String(),
			})
			return
		}
	}

	// 重複チェック（email）
	if req.Email != "" {
		existing, err := h.memberRepo.FindByEmail(ctx, tenantID, req.Email)
		if err != nil && err.Error() != "member not found" {
			writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to check email duplication", nil)
			return
		}
		if existing != nil {
			writeError(w, http.StatusConflict, "ERR_CONFLICT", "This email is already registered", map[string]interface{}{
				"member_id": existing.MemberID().String(),
			})
			return
		}
	}

	// Member エンティティの作成
	newMember, err := domainMember.NewMember(
		tenantID,
		req.DisplayName,
		req.DiscordUserID,
		req.Email,
	)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", err.Error(), nil)
		return
	}

	// 保存
	if err := h.memberRepo.Save(ctx, newMember); err != nil {
		log.Printf("CreateMember error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to create member", nil)
		return
	}

	// レスポンス
	resp := MemberResponse{
		MemberID:      newMember.MemberID().String(),
		TenantID:      newMember.TenantID().String(),
		DisplayName:   newMember.DisplayName(),
		DiscordUserID: newMember.DiscordUserID(),
		Email:         newMember.Email(),
		IsActive:      newMember.IsActive(),
		RoleIDs:       []string{}, // 新規作成時はロールなし
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
	input := appMember.UpdateMemberInput{
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

	// 更新後のロールIDを取得
	roleIDs, err := h.memberRoleRepo.FindRolesByMemberID(ctx, common.MemberID(memberID))
	if err != nil {
		log.Printf("Failed to fetch roles after update: %v", err)
		roleIDs = []common.RoleID{}
	}

	roleIDStrs := make([]string, len(roleIDs))
	for i, roleID := range roleIDs {
		roleIDStrs[i] = roleID.String()
	}

	// レスポンス
	resp := MemberResponse{
		MemberID:      output.MemberID,
		TenantID:      output.TenantID,
		DisplayName:   output.DisplayName,
		DiscordUserID: output.DiscordUserID,
		Email:         output.Email,
		IsActive:      output.IsActive,
		RoleIDs:       roleIDStrs,
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

	// メンバー一覧を取得
	members, err := h.memberRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch members", nil)
		return
	}

	// is_active フィルタ
	var filteredMembers []*domainMember.Member
	if isActiveStr == "true" {
		for _, m := range members {
			if m.IsActive() {
				filteredMembers = append(filteredMembers, m)
			}
		}
	} else if isActiveStr == "false" {
		for _, m := range members {
			if !m.IsActive() {
				filteredMembers = append(filteredMembers, m)
			}
		}
	} else {
		filteredMembers = members
	}

	// レスポンス構築
	memberResponses := make([]MemberResponse, 0, len(filteredMembers))
	for _, m := range filteredMembers {
		// メンバーのロールIDを取得
		roleIDs, err := h.memberRoleRepo.FindRolesByMemberID(ctx, m.MemberID())
		if err != nil {
			log.Printf("Failed to fetch roles for member %s: %v", m.MemberID().String(), err)
			roleIDs = []common.RoleID{} // エラー時は空配列
		}

		// RoleIDをstringスライスに変換
		roleIDStrs := make([]string, len(roleIDs))
		for i, roleID := range roleIDs {
			roleIDStrs[i] = roleID.String()
		}

		memberResponses = append(memberResponses, MemberResponse{
			MemberID:      m.MemberID().String(),
			TenantID:      m.TenantID().String(),
			DisplayName:   m.DisplayName(),
			DiscordUserID: m.DiscordUserID(),
			Email:         m.Email(),
			IsActive:      m.IsActive(),
			RoleIDs:       roleIDStrs,
			CreatedAt:     m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:     m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
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

	// メンバーの取得
	m, err := h.memberRepo.FindByID(ctx, tenantID, memberID)
	if err != nil {
		if err.Error() == "member not found" {
			writeError(w, http.StatusNotFound, "ERR_NOT_FOUND", "Member not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to fetch member", nil)
		return
	}

	// レスポンス
	resp := MemberResponse{
		MemberID:      m.MemberID().String(),
		TenantID:      m.TenantID().String(),
		DisplayName:   m.DisplayName(),
		DiscordUserID: m.DiscordUserID(),
		Email:         m.Email(),
		IsActive:      m.IsActive(),
		CreatedAt:     m.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     m.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
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
	output, err := h.getRecentAttendanceUsecase.Execute(ctx, appMember.GetRecentAttendanceInput{
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

