package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	schedDomain "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/schedule"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
)

type ScheduleHandler struct {
	createScheduleUsecase       *schedule.CreateScheduleUsecase
	submitResponseUsecase       *schedule.SubmitResponseUsecase
	decideScheduleUsecase       *schedule.DecideScheduleUsecase
	closeScheduleUsecase        *schedule.CloseScheduleUsecase
	getScheduleUsecase          *schedule.GetScheduleUsecase
	getScheduleByTokenUsecase   *schedule.GetScheduleByTokenUsecase
	getResponsesUsecase         *schedule.GetResponsesUsecase
	listSchedulesUsecase        *schedule.ListSchedulesUsecase
}

func NewScheduleHandler(pool *pgxpool.Pool) *ScheduleHandler {
	scheduleRepo := db.NewScheduleRepository(pool)
	txManager := db.NewPgxTxManager(pool)
	systemClock := &clock.RealClock{}

	return &ScheduleHandler{
		createScheduleUsecase:       schedule.NewCreateScheduleUsecase(scheduleRepo, systemClock),
		submitResponseUsecase:       schedule.NewSubmitResponseUsecase(scheduleRepo, txManager, systemClock),
		decideScheduleUsecase:       schedule.NewDecideScheduleUsecase(scheduleRepo, systemClock),
		closeScheduleUsecase:        schedule.NewCloseScheduleUsecase(scheduleRepo, systemClock),
		getScheduleUsecase:          schedule.NewGetScheduleUsecase(scheduleRepo),
		getScheduleByTokenUsecase:   schedule.NewGetScheduleByTokenUsecase(scheduleRepo),
		getResponsesUsecase:         schedule.NewGetResponsesUsecase(scheduleRepo),
		listSchedulesUsecase:        schedule.NewListSchedulesUsecase(scheduleRepo),
	}
}

// Management APIs (JWT required)

type ListSchedulesResponse struct {
	Schedules []ScheduleSummaryResponse `json:"schedules"`
}

type ScheduleSummaryResponse struct {
	ScheduleID         string     `json:"schedule_id"`
	TenantID           string     `json:"tenant_id"`
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	EventID            *string    `json:"event_id"`
	PublicToken        string     `json:"public_token"`
	Status             string     `json:"status"`
	Deadline           *time.Time `json:"deadline"`
	DecidedCandidateID *string    `json:"decided_candidate_id"`
	CandidateCount     int        `json:"candidate_count"`
	ResponseCount      int        `json:"response_count"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	input := schedule.ListSchedulesInput{
		TenantID: tenantID.String(),
	}

	output, err := h.listSchedulesUsecase.Execute(ctx, input)
	if err != nil {
		RespondInternalError(w)
		return
	}

	summaries := make([]ScheduleSummaryResponse, len(output.Schedules))
	for i, s := range output.Schedules {
		summaries[i] = ScheduleSummaryResponse{
			ScheduleID:         s.ScheduleID,
			TenantID:           s.TenantID,
			Title:              s.Title,
			Description:        s.Description,
			EventID:            s.EventID,
			PublicToken:        s.PublicToken,
			Status:             s.Status,
			Deadline:           s.Deadline,
			DecidedCandidateID: s.DecidedCandidateID,
			CandidateCount:     s.CandidateCount,
			ResponseCount:      s.ResponseCount,
			CreatedAt:          s.CreatedAt,
			UpdatedAt:          s.UpdatedAt,
		}
	}

	resp := ListSchedulesResponse{
		Schedules: summaries,
	}

	RespondJSON(w, http.StatusOK, resp)
}

type CreateScheduleRequest struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	EventID     *string              `json:"event_id"`
	Candidates  []CandidateRequest   `json:"candidates"`
	Deadline    *time.Time           `json:"deadline"`
}

type CandidateRequest struct {
	Date      time.Time  `json:"date"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
}

type CreateScheduleResponse struct {
	ScheduleID  string     `json:"schedule_id"`
	TenantID    string     `json:"tenant_id"`
	Title       string     `json:"title"`
	PublicToken string     `json:"public_token"`
	Status      string     `json:"status"`
	Deadline    *time.Time `json:"deadline"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	var req CreateScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "invalid request body")
		return
	}

	if req.Title == "" {
		RespondBadRequest(w, "title is required")
		return
	}

	if len(req.Candidates) == 0 {
		RespondBadRequest(w, "at least one candidate is required")
		return
	}

	candidates := make([]schedule.CandidateInput, len(req.Candidates))
	for i, c := range req.Candidates {
		candidates[i] = schedule.CandidateInput{
			Date:      c.Date,
			StartTime: c.StartTime,
			EndTime:   c.EndTime,
		}
	}

	input := schedule.CreateScheduleInput{
		TenantID:    tenantID.String(),
		Title:       req.Title,
		Description: req.Description,
		EventID:     req.EventID,
		Candidates:  candidates,
		Deadline:    req.Deadline,
	}

	output, err := h.createScheduleUsecase.Execute(ctx, input)
	if err != nil {
		var domainErr *common.DomainError
		if errors.As(err, &domainErr) {
			switch domainErr.Code() {
			case common.ErrNotFound:
				RespondNotFound(w, domainErr.Error())
				return
			case common.ErrInvalidInput:
				RespondBadRequest(w, domainErr.Error())
				return
			}
		}
		RespondInternalError(w)
		return
	}

	resp := CreateScheduleResponse{
		ScheduleID:  output.ScheduleID,
		TenantID:    output.TenantID,
		Title:       output.Title,
		PublicToken: output.PublicToken,
		Status:      output.Status,
		Deadline:    output.Deadline,
		CreatedAt:   output.CreatedAt,
	}

	RespondJSON(w, http.StatusCreated, SuccessResponse{Data: resp})
}

type GetScheduleResponse struct {
	ScheduleID         string              `json:"schedule_id"`
	TenantID           string              `json:"tenant_id"`
	Title              string              `json:"title"`
	Description        string              `json:"description"`
	EventID            *string             `json:"event_id"`
	PublicToken        string              `json:"public_token"`
	Status             string              `json:"status"`
	Deadline           *time.Time          `json:"deadline"`
	DecidedCandidateID *string             `json:"decided_candidate_id"`
	Candidates         []CandidateResponse `json:"candidates"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
}

type CandidateResponse struct {
	CandidateID string     `json:"candidate_id"`
	Date        time.Time  `json:"date"`
	StartTime   *time.Time `json:"start_time"`
	EndTime     *time.Time `json:"end_time"`
}

func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	scheduleID := chi.URLParam(r, "schedule_id")
	if scheduleID == "" {
		RespondBadRequest(w, "schedule_id is required")
		return
	}

	input := schedule.GetScheduleInput{
		TenantID:   tenantID.String(),
		ScheduleID: scheduleID,
	}

	output, err := h.getScheduleUsecase.Execute(ctx, input)
	if err != nil {
		var domainErr *common.DomainError
		if errors.As(err, &domainErr) {
			switch domainErr.Code() {
			case common.ErrNotFound:
				RespondNotFound(w, domainErr.Error())
				return
			case common.ErrInvalidInput:
				RespondBadRequest(w, domainErr.Error())
				return
			}
		}
		RespondInternalError(w)
		return
	}

	candidates := make([]CandidateResponse, len(output.Candidates))
	for i, c := range output.Candidates {
		candidates[i] = CandidateResponse{
			CandidateID: c.CandidateID,
			Date:        c.Date,
			StartTime:   c.StartTime,
			EndTime:     c.EndTime,
		}
	}

	resp := GetScheduleResponse{
		ScheduleID:         output.ScheduleID,
		TenantID:           output.TenantID,
		Title:              output.Title,
		Description:        output.Description,
		EventID:            output.EventID,
		PublicToken:        output.PublicToken,
		Status:             output.Status,
		Deadline:           output.Deadline,
		DecidedCandidateID: output.DecidedCandidateID,
		Candidates:         candidates,
		CreatedAt:          output.CreatedAt,
		UpdatedAt:          output.UpdatedAt,
	}

	RespondJSON(w, http.StatusOK, resp)
}

type DecideScheduleRequest struct {
	CandidateID string `json:"candidate_id"`
}

type DecideScheduleResponse struct {
	ScheduleID         string    `json:"schedule_id"`
	Status             string    `json:"status"`
	DecidedCandidateID string    `json:"decided_candidate_id"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (h *ScheduleHandler) DecideSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	scheduleID := chi.URLParam(r, "schedule_id")
	if scheduleID == "" {
		RespondBadRequest(w, "schedule_id is required")
		return
	}

	var req DecideScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "invalid request body")
		return
	}

	if req.CandidateID == "" {
		RespondBadRequest(w, "candidate_id is required")
		return
	}

	input := schedule.DecideScheduleInput{
		TenantID:    tenantID.String(),
		ScheduleID:  scheduleID,
		CandidateID: req.CandidateID,
	}

	output, err := h.decideScheduleUsecase.Execute(ctx, input)
	if err != nil {
		// Check for domain errors from schedule domain
		if errors.Is(err, schedDomain.ErrAlreadyDecided) || errors.Is(err, schedDomain.ErrCandidateNotFound) {
			RespondBadRequest(w, err.Error())
			return
		}
		var domainErr *common.DomainError
		if errors.As(err, &domainErr) {
			switch domainErr.Code() {
			case common.ErrNotFound:
				RespondNotFound(w, domainErr.Error())
				return
			case common.ErrInvalidInput:
				RespondBadRequest(w, domainErr.Error())
				return
			}
		}
		RespondInternalError(w)
		return
	}

	resp := DecideScheduleResponse{
		ScheduleID:         output.ScheduleID,
		Status:             output.Status,
		DecidedCandidateID: output.DecidedCandidateID,
		UpdatedAt:          output.UpdatedAt,
	}

	RespondJSON(w, http.StatusOK, resp)
}

type CloseScheduleResponse struct {
	ScheduleID string    `json:"schedule_id"`
	Status     string    `json:"status"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (h *ScheduleHandler) CloseSchedule(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	scheduleID := chi.URLParam(r, "schedule_id")
	if scheduleID == "" {
		RespondBadRequest(w, "schedule_id is required")
		return
	}

	input := schedule.CloseScheduleInput{
		TenantID:   tenantID.String(),
		ScheduleID: scheduleID,
	}

	output, err := h.closeScheduleUsecase.Execute(ctx, input)
	if err != nil {
		// Check for domain errors from schedule domain
		if errors.Is(err, schedDomain.ErrAlreadyClosed) {
			RespondBadRequest(w, err.Error())
			return
		}
		var domainErr *common.DomainError
		if errors.As(err, &domainErr) {
			switch domainErr.Code() {
			case common.ErrNotFound:
				RespondNotFound(w, domainErr.Error())
				return
			case common.ErrInvalidInput:
				RespondBadRequest(w, domainErr.Error())
				return
			}
		}
		RespondInternalError(w)
		return
	}

	resp := CloseScheduleResponse{
		ScheduleID: output.ScheduleID,
		Status:     output.Status,
		UpdatedAt:  output.UpdatedAt,
	}

	RespondJSON(w, http.StatusOK, resp)
}

type GetResponsesResponse struct {
	ScheduleID string                       `json:"schedule_id"`
	Responses  []ScheduleResponseResponse   `json:"responses"`
}

type ScheduleResponseResponse struct {
	ResponseID   string    `json:"response_id"`
	MemberID     string    `json:"member_id"`
	CandidateID  string    `json:"candidate_id"`
	Availability string    `json:"availability"`
	Note         string    `json:"note"`
	RespondedAt  time.Time `json:"responded_at"`
}

func (h *ScheduleHandler) GetResponses(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	scheduleID := chi.URLParam(r, "schedule_id")
	if scheduleID == "" {
		RespondBadRequest(w, "schedule_id is required")
		return
	}

	input := schedule.GetResponsesInput{
		TenantID:   tenantID.String(),
		ScheduleID: scheduleID,
	}

	output, err := h.getResponsesUsecase.Execute(ctx, input)
	if err != nil {
		var domainErr *common.DomainError
		if errors.As(err, &domainErr) {
			switch domainErr.Code() {
			case common.ErrNotFound:
				RespondNotFound(w, domainErr.Error())
				return
			case common.ErrInvalidInput:
				RespondBadRequest(w, domainErr.Error())
				return
			}
		}
		RespondInternalError(w)
		return
	}

	responses := make([]ScheduleResponseResponse, len(output.Responses))
	for i, r := range output.Responses {
		responses[i] = ScheduleResponseResponse{
			ResponseID:   r.ResponseID,
			MemberID:     r.MemberID,
			CandidateID:  r.CandidateID,
			Availability: r.Availability,
			Note:         r.Note,
			RespondedAt:  r.RespondedAt,
		}
	}

	resp := GetResponsesResponse{
		ScheduleID: output.ScheduleID,
		Responses:  responses,
	}

	RespondJSON(w, http.StatusOK, resp)
}

// Public API (No authentication)

// GetScheduleByToken handles GET /api/v1/public/schedules/:token
func (h *ScheduleHandler) GetScheduleByToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")
	if token == "" {
		RespondNotFound(w, "Schedule not found")
		return
	}

	output, err := h.getScheduleByTokenUsecase.Execute(ctx, schedule.GetScheduleByTokenInput{
		PublicToken: token,
	})
	if err != nil {
		RespondNotFound(w, "Schedule not found")
		return
	}

	candidates := make([]CandidateResponse, len(output.Candidates))
	for i, c := range output.Candidates {
		candidates[i] = CandidateResponse{
			CandidateID: c.CandidateID,
			Date:        c.Date,
			StartTime:   c.StartTime,
			EndTime:     c.EndTime,
		}
	}

	resp := GetScheduleResponse{
		ScheduleID:         output.ScheduleID,
		TenantID:           output.TenantID,
		Title:              output.Title,
		Description:        output.Description,
		EventID:            output.EventID,
		PublicToken:        output.PublicToken,
		Status:             output.Status,
		Deadline:           output.Deadline,
		DecidedCandidateID: output.DecidedCandidateID,
		Candidates:         candidates,
		CreatedAt:          output.CreatedAt,
		UpdatedAt:          output.UpdatedAt,
	}

	// フロントエンドが { data: ... } 形式を期待しているため、SuccessResponseでラップ
	RespondJSON(w, http.StatusOK, SuccessResponse{Data: resp})
}

type ScheduleSubmitResponseRequest struct {
	MemberID  string                   `json:"member_id"`
	Responses []ScheduleResponseInput  `json:"responses"`
}

type ScheduleResponseInput struct {
	CandidateID  string `json:"candidate_id"`
	Availability string `json:"availability"`
	Note         string `json:"note"`
}

type ScheduleSubmitResponseResponse struct {
	ScheduleID  string    `json:"schedule_id"`
	MemberID    string    `json:"member_id"`
	RespondedAt time.Time `json:"responded_at"`
}

func (h *ScheduleHandler) SubmitResponse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")
	if token == "" {
		RespondNotFound(w, "schedule not found")
		return
	}

	var req ScheduleSubmitResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		println("ERROR: Failed to decode request body:", err.Error())
		RespondBadRequest(w, "invalid request body")
		return
	}

	println("DEBUG: SubmitResponse request - MemberID:", req.MemberID, "ResponseCount:", len(req.Responses))

	if req.MemberID == "" {
		println("ERROR: member_id is empty")
		RespondBadRequest(w, "member_id is required")
		return
	}

	if len(req.Responses) == 0 {
		println("ERROR: responses array is empty")
		RespondBadRequest(w, "at least one response is required")
		return
	}

	responses := make([]schedule.ResponseInput, len(req.Responses))
	for i, r := range req.Responses {
		responses[i] = schedule.ResponseInput{
			CandidateID:  r.CandidateID,
			Availability: r.Availability,
			Note:         r.Note,
		}
	}

	input := schedule.SubmitResponseInput{
		PublicToken: token,
		MemberID:    req.MemberID,
		Responses:   responses,
	}

	output, err := h.submitResponseUsecase.Execute(ctx, input)
	if err != nil {
		if errors.Is(err, schedule.ErrScheduleNotFound) {
			RespondNotFound(w, "schedule not found")
			return
		}
		if errors.Is(err, schedule.ErrMemberNotAllowed) {
			RespondBadRequest(w, "invalid member")
			return
		}
		// Check for domain errors from schedule domain
		if errors.Is(err, schedDomain.ErrScheduleClosed) || errors.Is(err, schedDomain.ErrDeadlinePassed) {
			RespondBadRequest(w, err.Error())
			return
		}
		var domainErr *common.DomainError
		if errors.As(err, &domainErr) {
			switch domainErr.Code() {
			case common.ErrInvalidInput:
				RespondBadRequest(w, domainErr.Error())
				return
			}
		}
		RespondInternalError(w)
		return
	}

	resp := ScheduleSubmitResponseResponse{
		ScheduleID:  output.ScheduleID,
		MemberID:    output.MemberID,
		RespondedAt: output.RespondedAt,
	}

	RespondJSON(w, http.StatusOK, resp)
}
