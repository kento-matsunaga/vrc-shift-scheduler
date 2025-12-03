package rest

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// Context keys
type contextKey string

const (
	contextKeyTenantID contextKey = "tenant_id"
	contextKeyMemberID contextKey = "member_id"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// ListResponse represents a paginated list response
type ListResponse struct {
	Data       interface{} `json:"data"`
	TotalCount int         `json:"total_count"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// RespondSuccess sends a successful JSON response
func RespondSuccess(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusOK, SuccessResponse{Data: data})
}

// RespondCreated sends a created response
func RespondCreated(w http.ResponseWriter, data interface{}) {
	RespondJSON(w, http.StatusCreated, SuccessResponse{Data: data})
}

// RespondList sends a paginated list response
func RespondList(w http.ResponseWriter, data interface{}, totalCount, page, limit int) {
	RespondJSON(w, http.StatusOK, ListResponse{
		Data:       data,
		TotalCount: totalCount,
		Page:       page,
		Limit:      limit,
	})
}

// RespondNoContent sends a no content response
func RespondNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	RespondJSON(w, statusCode, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// RespondDomainError sends an error response based on domain error
func RespondDomainError(w http.ResponseWriter, err error) {
	if domainErr, ok := err.(*common.DomainError); ok {
		switch domainErr.Code() {
		case common.ErrNotFound:
			RespondError(w, http.StatusNotFound, "ERR_NOT_FOUND", domainErr.Message, nil)
		case common.ErrInvalidInput:
			RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", domainErr.Message, nil)
		case common.ErrConflict:
			RespondError(w, http.StatusConflict, "ERR_CONFLICT", domainErr.Message, nil)
		case "INVARIANT_VIOLATION":
			RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", domainErr.Message, nil)
		case common.ErrUnauthorized:
			RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", domainErr.Message, nil)
		default:
			RespondError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Internal server error", nil)
		}
	} else {
		RespondError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Internal server error", nil)
	}
}

// RespondBadRequest sends a bad request error
func RespondBadRequest(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", message, nil)
}

// RespondNotFound sends a not found error
func RespondNotFound(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusNotFound, "ERR_NOT_FOUND", message, nil)
}

// RespondConflict sends a conflict error
func RespondConflict(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusConflict, "ERR_CONFLICT", message, nil)
}

// RespondForbidden sends a forbidden error
func RespondForbidden(w http.ResponseWriter, message string) {
	RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", message, nil)
}

// RespondInternalError sends an internal server error
func RespondInternalError(w http.ResponseWriter) {
	RespondError(w, http.StatusInternalServerError, "ERR_INTERNAL", "Internal server error", nil)
}

// writeError is an alias for RespondError (for consistency with handlers)
func writeError(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
	RespondError(w, statusCode, code, message, details)
}

// writeSuccess is an alias for RespondJSON with SuccessResponse (for consistency with handlers)
func writeSuccess(w http.ResponseWriter, statusCode int, data interface{}) {
	RespondJSON(w, statusCode, SuccessResponse{Data: data})
}

// getTenantIDFromContext retrieves the tenant ID from the request context
func getTenantIDFromContext(ctx context.Context) (common.TenantID, bool) {
	tenantIDStr, ok := ctx.Value(contextKeyTenantID).(string)
	if !ok || tenantIDStr == "" {
		return "", false
	}

	tenantID, err := common.ParseTenantID(tenantIDStr)
	if err != nil {
		return "", false
	}

	return tenantID, true
}

// getMemberIDFromContext retrieves the member ID from the request context
func getMemberIDFromContext(ctx context.Context) (common.MemberID, bool) {
	memberIDStr, ok := ctx.Value(contextKeyMemberID).(string)
	if !ok || memberIDStr == "" {
		return "", false
	}

	memberID, err := common.ParseMemberID(memberIDStr)
	if err != nil {
		return "", false
	}

	return memberID, true
}

