package rest

import (
	"io"
	"log"
	"net/http"
	"strconv"

	importapp "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/import"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	importjob "github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/import"
	"github.com/go-chi/chi/v5"
)

// ImportHandler handles import-related HTTP requests
type ImportHandler struct {
	importMembersUC   *importapp.ImportMembersUsecase
	getImportStatusUC *importapp.GetImportStatusUsecase
	getImportResultUC *importapp.GetImportResultUsecase
	listImportJobsUC  *importapp.ListImportJobsUsecase
}

// NewImportHandler creates a new ImportHandler
func NewImportHandler(
	importMembersUC *importapp.ImportMembersUsecase,
	getImportStatusUC *importapp.GetImportStatusUsecase,
	getImportResultUC *importapp.GetImportResultUsecase,
	listImportJobsUC *importapp.ListImportJobsUsecase,
) *ImportHandler {
	return &ImportHandler{
		importMembersUC:   importMembersUC,
		getImportStatusUC: getImportStatusUC,
		getImportResultUC: getImportResultUC,
		listImportJobsUC:  listImportJobsUC,
	}
}

// ImportMembersResponse represents the response for member import
type ImportMembersResponse struct {
	ImportJobID  string                `json:"import_job_id"`
	Status       string                `json:"status"`
	TotalRows    int                   `json:"total_rows"`
	SuccessCount int                   `json:"success_count"`
	ErrorCount   int                   `json:"error_count"`
	Errors       []ImportErrorResponse `json:"errors,omitempty"`
}

// ImportErrorResponse represents an import error
type ImportErrorResponse struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}

// ImportStatusResponse represents the import status
type ImportStatusResponse struct {
	ImportJobID   string  `json:"import_job_id"`
	Status        string  `json:"status"`
	ImportType    string  `json:"import_type"`
	FileName      string  `json:"file_name"`
	TotalRows     int     `json:"total_rows"`
	ProcessedRows int     `json:"processed_rows"`
	SuccessCount  int     `json:"success_count"`
	ErrorCount    int     `json:"error_count"`
	Progress      float64 `json:"progress"`
	StartedAt     *string `json:"started_at,omitempty"`
	CompletedAt   *string `json:"completed_at,omitempty"`
	CreatedAt     string  `json:"created_at"`
}

// ImportResultResponse represents the import result
type ImportResultResponse struct {
	ImportJobID  string                `json:"import_job_id"`
	Status       string                `json:"status"`
	TotalRows    int                   `json:"total_rows"`
	SuccessCount int                   `json:"success_count"`
	ErrorCount   int                   `json:"error_count"`
	SkippedCount int                   `json:"skipped_count"`
	Errors       []ImportErrorResponse `json:"errors,omitempty"`
}

// ImportJobListResponse represents the list of import jobs
type ImportJobListResponse struct {
	Jobs       []ImportStatusResponse `json:"jobs"`
	TotalCount int                    `json:"total_count"`
}

// ImportMembers handles POST /api/v1/imports/members
func (h *ImportHandler) ImportMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// AdminIDの取得
	adminID, ok := ctx.Value(ContextKeyAdminID).(common.AdminID)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Admin ID is required", nil)
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Failed to parse form data", nil)
		return
	}

	// ファイルの取得
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "File is required", nil)
		return
	}
	defer file.Close()

	// ファイルデータの読み込み
	fileData, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to read file", nil)
		return
	}

	// オプションの取得
	skipExisting := r.FormValue("skip_existing") == "true"
	updateExisting := r.FormValue("update_existing") == "true"
	fuzzyMatch := r.FormValue("fuzzy_match") == "true"

	// Usecaseの実行
	input := importapp.ImportMembersInput{
		TenantID: tenantID,
		AdminID:  adminID,
		FileName: header.Filename,
		FileData: fileData,
		Options: importjob.ImportOptions{
			SkipExisting:     skipExisting,
			UpdateExisting:   updateExisting,
			FuzzyMemberMatch: fuzzyMatch,
		},
	}

	output, err := h.importMembersUC.Execute(ctx, input)
	if err != nil {
		log.Printf("ImportMembers error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to import members", nil)
		return
	}

	// エラー詳細をレスポンス用に変換
	errors := make([]ImportErrorResponse, len(output.Errors))
	for i, e := range output.Errors {
		errors[i] = ImportErrorResponse{
			Row:     e.Row,
			Message: e.Message,
		}
	}

	resp := ImportMembersResponse{
		ImportJobID:  output.ImportJobID.String(),
		Status:       string(output.Status),
		TotalRows:    output.TotalRows,
		SuccessCount: output.SuccessCount,
		ErrorCount:   output.ErrorCount,
		Errors:       errors,
	}

	writeSuccess(w, http.StatusOK, resp)
}

// GetImportStatus handles GET /api/v1/imports/{import_job_id}/status
func (h *ImportHandler) GetImportStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得（認可チェック用）
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// import_job_idの取得
	importJobIDStr := chi.URLParam(r, "import_job_id")
	if importJobIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "import_job_id is required", nil)
		return
	}

	importJobID, err := common.ParseImportJobID(importJobIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid import_job_id format", nil)
		return
	}

	// Usecaseの実行
	input := importapp.GetImportStatusInput{
		ImportJobID: importJobID,
		TenantID:    tenantID,
	}

	output, err := h.getImportStatusUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// レスポンス
	var startedAt, completedAt *string
	if output.StartedAt != nil {
		s := output.StartedAt.Format("2006-01-02T15:04:05Z07:00")
		startedAt = &s
	}
	if output.CompletedAt != nil {
		c := output.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		completedAt = &c
	}

	resp := ImportStatusResponse{
		ImportJobID:   output.ImportJobID.String(),
		Status:        string(output.Status),
		ImportType:    string(output.ImportType),
		FileName:      output.FileName,
		TotalRows:     output.TotalRows,
		ProcessedRows: output.ProcessedRows,
		SuccessCount:  output.SuccessCount,
		ErrorCount:    output.ErrorCount,
		Progress:      output.Progress,
		StartedAt:     startedAt,
		CompletedAt:   completedAt,
		CreatedAt:     output.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	writeSuccess(w, http.StatusOK, resp)
}

// GetImportResult handles GET /api/v1/imports/{import_job_id}/result
func (h *ImportHandler) GetImportResult(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得（認可チェック用）
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// import_job_idの取得
	importJobIDStr := chi.URLParam(r, "import_job_id")
	if importJobIDStr == "" {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "import_job_id is required", nil)
		return
	}

	importJobID, err := common.ParseImportJobID(importJobIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid import_job_id format", nil)
		return
	}

	// Usecaseの実行
	input := importapp.GetImportResultInput{
		ImportJobID: importJobID,
		TenantID:    tenantID,
	}

	output, err := h.getImportResultUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	// エラー詳細をレスポンス用に変換
	errors := make([]ImportErrorResponse, len(output.Errors))
	for i, e := range output.Errors {
		errors[i] = ImportErrorResponse{
			Row:     e.Row,
			Message: e.Message,
		}
	}

	resp := ImportResultResponse{
		ImportJobID:  output.ImportJobID.String(),
		Status:       string(output.Status),
		TotalRows:    output.TotalRows,
		SuccessCount: output.SuccessCount,
		ErrorCount:   output.ErrorCount,
		SkippedCount: output.SkippedCount,
		Errors:       errors,
	}

	writeSuccess(w, http.StatusOK, resp)
}

// ListImportJobs handles GET /api/v1/imports
func (h *ImportHandler) ListImportJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// テナントIDの取得
	tenantID, ok := getTenantIDFromContext(ctx)
	if !ok {
		writeError(w, http.StatusForbidden, "ERR_FORBIDDEN", "Tenant ID is required", nil)
		return
	}

	// ページネーションパラメータの取得
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // デフォルト
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Usecaseの実行
	input := importapp.ListImportJobsInput{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	}

	output, err := h.listImportJobsUC.Execute(ctx, input)
	if err != nil {
		log.Printf("ListImportJobs error: %+v", err)
		writeError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Failed to list import jobs", nil)
		return
	}

	// レスポンス構築
	jobs := make([]ImportStatusResponse, len(output.Jobs))
	for i, job := range output.Jobs {
		var startedAt, completedAt *string
		if job.StartedAt != nil {
			s := job.StartedAt.Format("2006-01-02T15:04:05Z07:00")
			startedAt = &s
		}
		if job.CompletedAt != nil {
			c := job.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
			completedAt = &c
		}

		jobs[i] = ImportStatusResponse{
			ImportJobID:   job.ImportJobID.String(),
			Status:        string(job.Status),
			ImportType:    string(job.ImportType),
			FileName:      job.FileName,
			TotalRows:     job.TotalRows,
			ProcessedRows: job.ProcessedRows,
			SuccessCount:  job.SuccessCount,
			ErrorCount:    job.ErrorCount,
			Progress:      job.Progress,
			StartedAt:     startedAt,
			CompletedAt:   completedAt,
			CreatedAt:     job.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	resp := ImportJobListResponse{
		Jobs:       jobs,
		TotalCount: output.TotalCount,
	}

	writeSuccess(w, http.StatusOK, resp)
}
