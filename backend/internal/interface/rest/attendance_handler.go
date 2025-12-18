package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/attendance"
	domainAttendance "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/attendance"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AttendanceHandler handles attendance-related HTTP requests
type AttendanceHandler struct {
	createCollectionUsecase       *attendance.CreateCollectionUsecase
	submitResponseUsecase         *attendance.SubmitResponseUsecase
	closeCollectionUsecase        *attendance.CloseCollectionUsecase
	getCollectionUsecase          *attendance.GetCollectionUsecase
	getCollectionByTokenUsecase   *attendance.GetCollectionByTokenUsecase
	getResponsesUsecase           *attendance.GetResponsesUsecase
	listCollectionsUsecase        *attendance.ListCollectionsUsecase
}

// NewAttendanceHandler creates a new AttendanceHandler
func NewAttendanceHandler(dbPool *pgxpool.Pool) *AttendanceHandler {
	// Repository, Clock, TxManagerの初期化
	repo := db.NewAttendanceRepository(dbPool)
	memberRepo := db.NewMemberRepository(dbPool)
	clk := &clock.RealClock{}
	txManager := db.NewPgxTxManager(dbPool)

	return &AttendanceHandler{
		createCollectionUsecase:       attendance.NewCreateCollectionUsecase(repo, clk),
		submitResponseUsecase:         attendance.NewSubmitResponseUsecase(repo, txManager, clk),
		closeCollectionUsecase:        attendance.NewCloseCollectionUsecase(repo, clk),
		getCollectionUsecase:          attendance.NewGetCollectionUsecase(repo),
		getCollectionByTokenUsecase:   attendance.NewGetCollectionByTokenUsecase(repo),
		getResponsesUsecase:           attendance.NewGetResponsesUsecase(repo, memberRepo),
		listCollectionsUsecase:        attendance.NewListCollectionsUsecase(repo),
	}
}

// CreateCollectionRequest represents the request body for creating an attendance collection
type CreateCollectionRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	TargetType  string     `json:"target_type"`  // "event" or "business_day"
	TargetID    string     `json:"target_id"`    // optional
	TargetDates []string   `json:"target_dates"` // ISO 8601 format array
	Deadline    *time.Time `json:"deadline"`     // optional
}

// TargetDateResponse represents a target date in API responses
type TargetDateResponse struct {
	TargetDateID string `json:"target_date_id"`
	TargetDate   string `json:"target_date"`   // ISO 8601 format
	DisplayOrder int    `json:"display_order"`
}

// CollectionResponse represents an attendance collection in API responses
type CollectionResponse struct {
	CollectionID string                `json:"collection_id"`
	TenantID     string                `json:"tenant_id"`
	Title        string                `json:"title"`
	Description  string                `json:"description"`
	TargetType   string                `json:"target_type"`
	TargetID     string                `json:"target_id"`
	TargetDates  []TargetDateResponse  `json:"target_dates,omitempty"` // Target dates with IDs
	PublicToken  string                `json:"public_token"`
	Status       string                `json:"status"`
	Deadline     *time.Time            `json:"deadline,omitempty"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
}

// SubmitResponseRequest represents the request body for submitting an attendance response
type SubmitResponseRequest struct {
	MemberID     string `json:"member_id"`
	TargetDateID string `json:"target_date_id"` // 対象日ID
	Response     string `json:"response"`       // "attending" or "absent"
	Note         string `json:"note"`
}

// ResponseDTO represents a single attendance response
type ResponseDTO struct {
	ResponseID   string    `json:"response_id"`
	MemberID     string    `json:"member_id"`
	MemberName   string    `json:"member_name"`    // メンバー表示名
	TargetDateID string    `json:"target_date_id"` // 対象日ID
	TargetDate   string    `json:"target_date"`    // 対象日（ISO 8601）
	Response     string    `json:"response"`
	Note         string    `json:"note"`
	RespondedAt  time.Time `json:"responded_at"`
}

// SubmitResponseResponse represents the response for submitting an attendance response
type SubmitResponseResponse struct {
	ResponseID   string    `json:"response_id"`
	CollectionID string    `json:"collection_id"`
	MemberID     string    `json:"member_id"`
	Response     string    `json:"response"`
	Note         string    `json:"note"`
	RespondedAt  time.Time `json:"responded_at"`
}

// ResponsesListResponse represents the response for getting responses
type ResponsesListResponse struct {
	CollectionID string        `json:"collection_id"`
	Responses    []ResponseDTO `json:"responses"`
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
		RespondBadRequest(w, "title is required")
		return
	}
	if req.TargetType == "" {
		RespondBadRequest(w, "target_type is required")
		return
	}

	// Parse target dates from strings to time.Time
	var targetDates []time.Time
	for _, dateStr := range req.TargetDates {
		parsedDate, err := time.Parse(time.RFC3339, dateStr)
		if err != nil {
			RespondBadRequest(w, "invalid target_date format: "+dateStr)
			return
		}
		targetDates = append(targetDates, parsedDate)
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
			ResponseID:   resp.ResponseID,
			MemberID:     resp.MemberID,
			MemberName:   resp.MemberName,
			TargetDateID: resp.TargetDateID,
			TargetDate:   resp.TargetDate.Format(time.RFC3339),
			Response:     resp.Response,
			Note:         resp.Note,
			RespondedAt:  resp.RespondedAt,
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
		RespondNotFound(w, "Collection not found")
		return
	}

	// Usecase呼び出し
	output, err := h.getCollectionByTokenUsecase.Execute(ctx, attendance.GetCollectionByTokenInput{
		PublicToken: token,
	})
	if err != nil {
		RespondNotFound(w, "Collection not found") // トークンエラー → 404（詳細は返さない）
		return
	}

	// Convert target dates to TargetDateResponse
	var targetDateResponses []TargetDateResponse
	for _, td := range output.TargetDates {
		targetDateResponses = append(targetDateResponses, TargetDateResponse{
			TargetDateID: td.TargetDateID,
			TargetDate:   td.TargetDate.Format(time.RFC3339),
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
		RespondNotFound(w, "Collection not found") // トークンエラー → 404
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
		RespondBadRequest(w, "member_id is required")
		return
	}
	if req.TargetDateID == "" {
		RespondBadRequest(w, "target_date_id is required")
		return
	}
	if req.Response == "" {
		RespondBadRequest(w, "response is required")
		return
	}

	// Usecase呼び出し
	output, err := h.submitResponseUsecase.Execute(ctx, attendance.SubmitResponseInput{
		PublicToken:  token,
		MemberID:     req.MemberID,
		TargetDateID: req.TargetDateID,
		Response:     req.Response,
		Note:         req.Note,
	})
	if err != nil {
		// エラーハンドリング（トークンエラー → 404, メンバーエラー → 400）
		switch {
		case errors.Is(err, attendance.ErrCollectionNotFound):
			RespondNotFound(w, "Collection not found") // トークンエラー → 404（詳細は返さない）
		case errors.Is(err, attendance.ErrMemberNotAllowed):
			RespondBadRequest(w, "Member not allowed") // メンバーエラー → 400（詳細は返さない）
		case errors.Is(err, domainAttendance.ErrCollectionClosed):
			RespondConflict(w, "Collection is closed")
		case errors.Is(err, domainAttendance.ErrDeadlinePassed):
			RespondConflict(w, "Deadline has passed")
		default:
			RespondDomainError(w, err)
		}
		return
	}

	// レスポンス
	RespondJSON(w, http.StatusCreated, SuccessResponse{
		Data: SubmitResponseResponse{
			ResponseID:   output.ResponseID,
			CollectionID: output.CollectionID,
			MemberID:     output.MemberID,
			Response:     output.Response,
			Note:         output.Note,
			RespondedAt:  output.RespondedAt,
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
