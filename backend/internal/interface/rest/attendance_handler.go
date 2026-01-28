package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/attendance"
	domainAttendance "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/go-chi/chi/v5"
)

// AttendanceHandler handles attendance-related HTTP requests
type AttendanceHandler struct {
	createCollectionUsecase      *attendance.CreateCollectionUsecase
	submitResponseUsecase        *attendance.SubmitResponseUsecase
	closeCollectionUsecase       *attendance.CloseCollectionUsecase
	deleteCollectionUsecase      *attendance.DeleteCollectionUsecase
	updateCollectionUsecase      *attendance.UpdateCollectionUsecase
	getCollectionUsecase         *attendance.GetCollectionUsecase
	getCollectionByTokenUsecase  *attendance.GetCollectionByTokenUsecase
	getResponsesUsecase          *attendance.GetResponsesUsecase
	listCollectionsUsecase       *attendance.ListCollectionsUsecase
	getMemberResponsesUsecase    *attendance.GetMemberResponsesUsecase
	getAllPublicResponsesUsecase *attendance.GetAllPublicResponsesUsecase
}

// NewAttendanceHandler creates a new AttendanceHandler with injected usecases
func NewAttendanceHandler(
	createCollectionUC *attendance.CreateCollectionUsecase,
	submitResponseUC *attendance.SubmitResponseUsecase,
	closeCollectionUC *attendance.CloseCollectionUsecase,
	deleteCollectionUC *attendance.DeleteCollectionUsecase,
	updateCollectionUC *attendance.UpdateCollectionUsecase,
	getCollectionUC *attendance.GetCollectionUsecase,
	getCollectionByTokenUC *attendance.GetCollectionByTokenUsecase,
	getResponsesUC *attendance.GetResponsesUsecase,
	listCollectionsUC *attendance.ListCollectionsUsecase,
	getMemberResponsesUC *attendance.GetMemberResponsesUsecase,
	getAllPublicResponsesUC *attendance.GetAllPublicResponsesUsecase,
) *AttendanceHandler {
	return &AttendanceHandler{
		createCollectionUsecase:      createCollectionUC,
		submitResponseUsecase:        submitResponseUC,
		closeCollectionUsecase:       closeCollectionUC,
		deleteCollectionUsecase:      deleteCollectionUC,
		updateCollectionUsecase:      updateCollectionUC,
		getCollectionUsecase:         getCollectionUC,
		getCollectionByTokenUsecase:  getCollectionByTokenUC,
		getResponsesUsecase:          getResponsesUC,
		listCollectionsUsecase:       listCollectionsUC,
		getMemberResponsesUsecase:    getMemberResponsesUC,
		getAllPublicResponsesUsecase: getAllPublicResponsesUC,
	}
}

// TargetDateRequest represents a target date in API requests
type TargetDateRequest struct {
	TargetDate string  `json:"target_date"` // ISO 8601 format
	StartTime  *string `json:"start_time"`  // HH:MM format (optional)
	EndTime    *string `json:"end_time"`    // HH:MM format (optional)
}

// CreateCollectionRequest represents the request body for creating an attendance collection
type CreateCollectionRequest struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	TargetType  string              `json:"target_type"` // "event" or "business_day"
	TargetID    string              `json:"target_id"`   // optional
	TargetDates []TargetDateRequest `json:"target_dates"`
	Deadline    *time.Time          `json:"deadline"`  // optional
	GroupIDs    []string            `json:"group_ids"` // optional: target group IDs
	RoleIDs     []string            `json:"role_ids"`  // optional: target role IDs
}

// TargetDateResponse represents a target date in API responses
type TargetDateResponse struct {
	TargetDateID string  `json:"target_date_id"`
	TargetDate   string  `json:"target_date"`          // ISO 8601 format
	StartTime    *string `json:"start_time,omitempty"` // HH:MM format (optional)
	EndTime      *string `json:"end_time,omitempty"`   // HH:MM format (optional)
	DisplayOrder int     `json:"display_order"`
}

// CollectionResponse represents an attendance collection in API responses
type CollectionResponse struct {
	CollectionID string               `json:"collection_id"`
	TenantID     string               `json:"tenant_id"`
	Title        string               `json:"title"`
	Description  string               `json:"description"`
	TargetType   string               `json:"target_type"`
	TargetID     string               `json:"target_id"`
	TargetDates  []TargetDateResponse `json:"target_dates,omitempty"` // Target dates with IDs
	PublicToken  string               `json:"public_token"`
	Status       string               `json:"status"`
	Deadline     *time.Time           `json:"deadline,omitempty"`
	GroupIDs     []string             `json:"group_ids,omitempty"` // Target group IDs
	RoleIDs      []string             `json:"role_ids,omitempty"`  // Target role IDs
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// SubmitResponseRequest represents the request body for submitting an attendance response
type SubmitResponseRequest struct {
	MemberID      string  `json:"member_id"`
	TargetDateID  string  `json:"target_date_id"` // 対象日ID
	Response      string  `json:"response"`       // "attending" or "absent" or "undecided"
	Note          string  `json:"note"`
	AvailableFrom *string `json:"available_from,omitempty"` // 参加可能開始時間 (HH:MM)
	AvailableTo   *string `json:"available_to,omitempty"`   // 参加可能終了時間 (HH:MM)
}

// ResponseDTO represents a single attendance response
type ResponseDTO struct {
	ResponseID    string    `json:"response_id"`
	MemberID      string    `json:"member_id"`
	MemberName    string    `json:"member_name"`    // メンバー表示名
	TargetDateID  string    `json:"target_date_id"` // 対象日ID
	TargetDate    string    `json:"target_date"`    // 対象日（ISO 8601）
	Response      string    `json:"response"`
	Note          string    `json:"note"`
	AvailableFrom *string   `json:"available_from,omitempty"` // 参加可能開始時間
	AvailableTo   *string   `json:"available_to,omitempty"`   // 参加可能終了時間
	RespondedAt   time.Time `json:"responded_at"`
}

// SubmitResponseResponse represents the response for submitting an attendance response
type SubmitResponseResponse struct {
	ResponseID    string    `json:"response_id"`
	CollectionID  string    `json:"collection_id"`
	MemberID      string    `json:"member_id"`
	Response      string    `json:"response"`
	Note          string    `json:"note"`
	AvailableFrom *string   `json:"available_from,omitempty"`
	AvailableTo   *string   `json:"available_to,omitempty"`
	RespondedAt   time.Time `json:"responded_at"`
}

// ResponsesListResponse represents the response for getting responses
type ResponsesListResponse struct {
	CollectionID string        `json:"collection_id"`
	Responses    []ResponseDTO `json:"responses"`
}

// UpdateCollectionRequest represents the request body for updating an attendance collection
type UpdateCollectionRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Deadline    *time.Time `json:"deadline"` // optional
}

// UpdateCollectionResponse represents an attendance collection update response
type UpdateCollectionResponse struct {
	CollectionID string     `json:"collection_id"`
	TenantID     string     `json:"tenant_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	Deadline     *time.Time `json:"deadline,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// CreateCollection handles POST /api/v1/attendance/collections
func (h *AttendanceHandler) CreateCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TenantIDの取得（JWTから）
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// リクエストボディのパース
	var req CreateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// バリデーション
	if req.Title == "" {
		RespondBadRequest(w, "タイトルを入力してください")
		return
	}
	if req.TargetType == "" {
		RespondBadRequest(w, "対象タイプを選択してください")
		return
	}

	// Parse target dates from request
	var targetDates []attendance.TargetDateInput
	for _, tdReq := range req.TargetDates {
		parsedDate, err := time.Parse(time.RFC3339, tdReq.TargetDate)
		if err != nil {
			RespondBadRequest(w, "対象日の形式が正しくありません: "+tdReq.TargetDate)
			return
		}
		targetDates = append(targetDates, attendance.TargetDateInput{
			TargetDate: parsedDate,
			StartTime:  tdReq.StartTime,
			EndTime:    tdReq.EndTime,
		})
	}

	// Usecase呼び出し
	output, err := h.createCollectionUsecase.Execute(ctx, attendance.CreateCollectionInput{
		TenantID:    tenantID.String(),
		Title:       req.Title,
		Description: req.Description,
		TargetType:  req.TargetType,
		TargetID:    req.TargetID,
		TargetDates: targetDates,
		Deadline:    req.Deadline,
		GroupIDs:    req.GroupIDs,
		RoleIDs:     req.RoleIDs,
	})
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondJSON(w, http.StatusCreated, SuccessResponse{
		Data: CollectionResponse{
			CollectionID: output.CollectionID,
			TenantID:     output.TenantID,
			Title:        output.Title,
			Description:  output.Description,
			TargetType:   output.TargetType,
			TargetID:     output.TargetID,
			PublicToken:  output.PublicToken,
			Status:       output.Status,
			Deadline:     output.Deadline,
			CreatedAt:    output.CreatedAt,
			UpdatedAt:    output.UpdatedAt,
		},
	})
}

// GetCollection handles GET /api/v1/attendance/collections/:id
func (h *AttendanceHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TenantIDの取得（JWTから）
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// URLパラメータからcollection_idを取得
	collectionID := chi.URLParam(r, "collection_id")
	if collectionID == "" {
		RespondBadRequest(w, "collection_id is required")
		return
	}

	// Usecase呼び出し
	output, err := h.getCollectionUsecase.Execute(ctx, attendance.GetCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID,
	})
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// Convert target dates to TargetDateResponse
	var targetDateResponses []TargetDateResponse
	for _, td := range output.TargetDates {
		targetDateResponses = append(targetDateResponses, TargetDateResponse{
			TargetDateID: td.TargetDateID,
			TargetDate:   td.TargetDate.Format(time.RFC3339),
			StartTime:    td.StartTime,
			EndTime:      td.EndTime,
			DisplayOrder: td.DisplayOrder,
		})
	}

	// レスポンス
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: CollectionResponse{
			CollectionID: output.CollectionID,
			TenantID:     output.TenantID,
			Title:        output.Title,
			Description:  output.Description,
			TargetType:   output.TargetType,
			TargetID:     output.TargetID,
			TargetDates:  targetDateResponses,
			PublicToken:  output.PublicToken,
			Status:       output.Status,
			Deadline:     output.Deadline,
			GroupIDs:     output.GroupIDs,
			RoleIDs:      output.RoleIDs,
			CreatedAt:    output.CreatedAt,
			UpdatedAt:    output.UpdatedAt,
		},
	})
}

// CloseCollection handles POST /api/v1/attendance/collections/:id/close
func (h *AttendanceHandler) CloseCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TenantIDの取得（JWTから）
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// URLパラメータからcollection_idを取得
	collectionID := chi.URLParam(r, "collection_id")
	if collectionID == "" {
		RespondBadRequest(w, "collection_id is required")
		return
	}

	// Usecase呼び出し
	output, err := h.closeCollectionUsecase.Execute(ctx, attendance.CloseCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID,
	})
	if err != nil {
		// ドメインエラーハンドリング
		switch {
		case errors.Is(err, domainAttendance.ErrAlreadyClosed):
			RespondConflict(w, "Collection is already closed")
		default:
			RespondDomainError(w, err)
		}
		return
	}

	// レスポンス
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: map[string]interface{}{
			"collection_id": output.CollectionID,
			"status":        output.Status,
			"updated_at":    output.UpdatedAt,
		},
	})
}

// DeleteCollection handles DELETE /api/v1/attendance/collections/:id
func (h *AttendanceHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TenantIDの取得（JWTから）
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// URLパラメータからcollection_idを取得
	collectionID := chi.URLParam(r, "collection_id")
	if collectionID == "" {
		RespondBadRequest(w, "collection_id is required")
		return
	}

	// Usecase呼び出し
	output, err := h.deleteCollectionUsecase.Execute(ctx, attendance.DeleteCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainAttendance.ErrAlreadyDeleted):
			RespondConflict(w, "Collection is already deleted")
		default:
			RespondDomainError(w, err)
		}
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: map[string]interface{}{
			"collection_id": output.CollectionID,
			"status":        output.Status,
			"deleted_at":    output.DeletedAt,
			"updated_at":    output.UpdatedAt,
		},
	})
}

// UpdateCollection handles PUT /api/v1/attendance/collections/:id
func (h *AttendanceHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	collectionID := chi.URLParam(r, "collection_id")
	if collectionID == "" {
		RespondBadRequest(w, "collection_id is required")
		return
	}

	var req UpdateCollectionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	if req.Title != "" && len(req.Title) > 255 {
		RespondBadRequest(w, "title must be less than 255 characters")
		return
	}

	output, err := h.updateCollectionUsecase.Execute(ctx, attendance.UpdateCollectionInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID,
		Title:        req.Title,
		Description:  req.Description,
		Deadline:     req.Deadline,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainAttendance.ErrAlreadyDeleted):
			RespondConflict(w, "Collection is already deleted")
		case errors.Is(err, domainAttendance.ErrCollectionClosed):
			RespondBadRequest(w, err.Error())
		default:
			RespondDomainError(w, err)
		}
		return
	}

	resp := UpdateCollectionResponse{
		CollectionID: output.CollectionID,
		TenantID:     output.TenantID,
		Title:        output.Title,
		Description:  output.Description,
		Status:       output.Status,
		Deadline:     output.Deadline,
		UpdatedAt:    output.UpdatedAt,
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: resp})
}

// GetResponses handles GET /api/v1/attendance/collections/:id/responses
func (h *AttendanceHandler) GetResponses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TenantIDの取得（JWTから）
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// URLパラメータからcollection_idを取得
	collectionID := chi.URLParam(r, "collection_id")
	if collectionID == "" {
		RespondBadRequest(w, "collection_id is required")
		return
	}

	// Usecase呼び出し
	output, err := h.getResponsesUsecase.Execute(ctx, attendance.GetResponsesInput{
		TenantID:     tenantID.String(),
		CollectionID: collectionID,
	})
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// DTOの変換
	responses := make([]ResponseDTO, 0, len(output.Responses))
	for _, resp := range output.Responses {
		responses = append(responses, ResponseDTO{
			ResponseID:    resp.ResponseID,
			MemberID:      resp.MemberID,
			MemberName:    resp.MemberName,
			TargetDateID:  resp.TargetDateID,
			TargetDate:    resp.TargetDate.Format(time.RFC3339),
			Response:      resp.Response,
			Note:          resp.Note,
			AvailableFrom: resp.AvailableFrom,
			AvailableTo:   resp.AvailableTo,
			RespondedAt:   resp.RespondedAt,
		})
	}

	// レスポンス
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: ResponsesListResponse{
			CollectionID: output.CollectionID,
			Responses:    responses,
		},
	})
}

// GetCollectionByToken handles GET /api/v1/public/attendance/:token
// Public API（認証不要）
func (h *AttendanceHandler) GetCollectionByToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// URLパラメータからtokenを取得
	token := chi.URLParam(r, "token")
	if token == "" {
		RespondNotFound(w, "出欠確認が見つかりません")
		return
	}

	// Usecase呼び出し
	output, err := h.getCollectionByTokenUsecase.Execute(ctx, attendance.GetCollectionByTokenInput{
		PublicToken: token,
	})
	if err != nil {
		RespondNotFound(w, "出欠確認が見つかりません") // トークンエラー → 404（詳細は返さない）
		return
	}

	// Convert target dates to TargetDateResponse
	var targetDateResponses []TargetDateResponse
	for _, td := range output.TargetDates {
		targetDateResponses = append(targetDateResponses, TargetDateResponse{
			TargetDateID: td.TargetDateID,
			TargetDate:   td.TargetDate.Format(time.RFC3339),
			StartTime:    td.StartTime,
			EndTime:      td.EndTime,
			DisplayOrder: td.DisplayOrder,
		})
	}

	// レスポンス
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: CollectionResponse{
			CollectionID: output.CollectionID,
			TenantID:     output.TenantID,
			Title:        output.Title,
			Description:  output.Description,
			TargetType:   output.TargetType,
			TargetID:     output.TargetID,
			TargetDates:  targetDateResponses,
			PublicToken:  output.PublicToken,
			Status:       output.Status,
			Deadline:     output.Deadline,
			GroupIDs:     output.GroupIDs,
			RoleIDs:      output.RoleIDs,
			CreatedAt:    output.CreatedAt,
			UpdatedAt:    output.UpdatedAt,
		},
	})
}

// SubmitResponse handles POST /api/v1/public/attendance/:token/responses
// Public API（認証不要）
func (h *AttendanceHandler) SubmitResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// URLパラメータからtokenを取得
	token := chi.URLParam(r, "token")
	if token == "" {
		RespondNotFound(w, "出欠確認が見つかりません") // トークンエラー → 404
		return
	}

	// リクエストボディのパース
	var req SubmitResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	// バリデーション
	if req.MemberID == "" {
		RespondBadRequest(w, "メンバーを選択してください")
		return
	}
	if req.TargetDateID == "" {
		RespondBadRequest(w, "対象日を選択してください")
		return
	}
	if req.Response == "" {
		RespondBadRequest(w, "回答を選択してください")
		return
	}

	// Usecase呼び出し
	output, err := h.submitResponseUsecase.Execute(ctx, attendance.SubmitResponseInput{
		PublicToken:   token,
		MemberID:      req.MemberID,
		TargetDateID:  req.TargetDateID,
		Response:      req.Response,
		Note:          req.Note,
		AvailableFrom: req.AvailableFrom,
		AvailableTo:   req.AvailableTo,
	})
	if err != nil {
		// エラーハンドリング（トークンエラー → 404, メンバーエラー → 400）
		switch {
		case errors.Is(err, attendance.ErrCollectionNotFound):
			RespondNotFound(w, "出欠確認が見つかりません") // トークンエラー → 404（詳細は返さない）
		case errors.Is(err, attendance.ErrMemberNotAllowed):
			RespondBadRequest(w, "このメンバーは回答できません") // メンバーエラー → 400（詳細は返さない）
		case errors.Is(err, domainAttendance.ErrCollectionClosed):
			RespondConflict(w, "この出欠確認は締め切られています")
		case errors.Is(err, domainAttendance.ErrDeadlinePassed):
			RespondConflict(w, "回答期限が過ぎています")
		default:
			RespondDomainError(w, err)
		}
		return
	}

	// レスポンス
	RespondJSON(w, http.StatusCreated, SuccessResponse{
		Data: SubmitResponseResponse{
			ResponseID:    output.ResponseID,
			CollectionID:  output.CollectionID,
			MemberID:      output.MemberID,
			Response:      output.Response,
			Note:          output.Note,
			AvailableFrom: output.AvailableFrom,
			AvailableTo:   output.AvailableTo,
			RespondedAt:   output.RespondedAt,
		},
	})
}

// ListCollections handles GET /api/v1/attendance/collections
func (h *AttendanceHandler) ListCollections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TenantIDの取得（JWTから）
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// Usecase呼び出し
	output, err := h.listCollectionsUsecase.Execute(ctx, attendance.ListCollectionsInput{
		TenantID: tenantID.String(),
	})
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: map[string]interface{}{
			"collections": output.Collections,
		},
	})
}

// MemberResponseDTO represents a single response for a member (public API)
type MemberResponseDTO struct {
	TargetDateID  string  `json:"target_date_id"`
	Response      string  `json:"response"`
	Note          string  `json:"note"`
	AvailableFrom *string `json:"available_from,omitempty"`
	AvailableTo   *string `json:"available_to,omitempty"`
}

// GetMemberResponses handles GET /api/v1/public/attendance/:token/members/:member_id/responses
// Public API（認証不要）- 特定メンバーの回答一覧を取得
func (h *AttendanceHandler) GetMemberResponses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// URLパラメータからtokenとmember_idを取得
	token := chi.URLParam(r, "token")
	if token == "" {
		RespondNotFound(w, "出欠確認が見つかりません")
		return
	}

	memberID := chi.URLParam(r, "member_id")
	if memberID == "" {
		RespondBadRequest(w, "member_id is required")
		return
	}

	// Usecase呼び出し
	output, err := h.getMemberResponsesUsecase.Execute(ctx, attendance.GetMemberResponsesInput{
		PublicToken: token,
		MemberID:    memberID,
	})
	if err != nil {
		// エラー種別に応じて適切なレスポンスを返す
		var domainErr *common.DomainError
		if errors.As(err, &domainErr) {
			switch domainErr.Code() {
			case common.ErrInvalidInput:
				RespondBadRequest(w, domainErr.Message)
				return
			case common.ErrNotFound:
				RespondNotFound(w, "出欠確認が見つかりません")
				return
			}
		}
		// その他のエラーは内部エラーとして処理
		RespondInternalError(w)
		return
	}

	// DTOの変換
	responses := make([]MemberResponseDTO, 0, len(output.Responses))
	for _, resp := range output.Responses {
		responses = append(responses, MemberResponseDTO{
			TargetDateID:  resp.TargetDateID,
			Response:      resp.Response,
			Note:          resp.Note,
			AvailableFrom: resp.AvailableFrom,
			AvailableTo:   resp.AvailableTo,
		})
	}

	// レスポンス
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: map[string]interface{}{
			"member_id": output.MemberID,
			"responses": responses,
		},
	})
}

// PublicResponseDTO represents a single response for the public table view
type PublicResponseDTO struct {
	MemberID      string  `json:"member_id"`
	MemberName    string  `json:"member_name"`
	TargetDateID  string  `json:"target_date_id"`
	Response      string  `json:"response"`
	Note          string  `json:"note"`
	AvailableFrom *string `json:"available_from,omitempty"`
	AvailableTo   *string `json:"available_to,omitempty"`
}

// GetAllPublicResponses handles GET /api/v1/public/attendance/:token/responses
// Public API（認証不要）- 全回答一覧を取得（調整さん形式の表示用）
func (h *AttendanceHandler) GetAllPublicResponses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// URLパラメータからtokenを取得
	token := chi.URLParam(r, "token")
	if token == "" {
		RespondNotFound(w, "出欠確認が見つかりません")
		return
	}

	// Usecase呼び出し
	output, err := h.getAllPublicResponsesUsecase.Execute(ctx, attendance.GetAllPublicResponsesInput{
		PublicToken: token,
	})
	if err != nil {
		// エラー種別に応じて適切なレスポンスを返す
		var domainErr *common.DomainError
		if errors.As(err, &domainErr) {
			switch domainErr.Code() {
			case common.ErrInvalidInput:
				RespondBadRequest(w, domainErr.Message)
				return
			case common.ErrNotFound:
				RespondNotFound(w, "出欠確認が見つかりません")
				return
			}
		}
		// その他のエラーは内部エラーとして処理
		RespondInternalError(w)
		return
	}

	// DTOの変換
	responses := make([]PublicResponseDTO, 0, len(output.Responses))
	for _, resp := range output.Responses {
		responses = append(responses, PublicResponseDTO{
			MemberID:      resp.MemberID,
			MemberName:    resp.MemberName,
			TargetDateID:  resp.TargetDateID,
			Response:      resp.Response,
			Note:          resp.Note,
			AvailableFrom: resp.AvailableFrom,
			AvailableTo:   resp.AvailableTo,
		})
	}

	// レスポンス
	RespondJSON(w, http.StatusOK, SuccessResponse{
		Data: map[string]interface{}{
			"responses": responses,
		},
	})
}
