package rest

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// ContextKeyTenantID is the context key for tenant ID
	ContextKeyTenantID ContextKey = "tenant_id"
	// ContextKeyMemberID is the context key for member ID
	ContextKeyMemberID ContextKey = "member_id"
)

// Logger is a middleware that logs HTTP requests
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf(
			"%s %s %d %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// CORS is a middleware that handles CORS headers
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tenant-ID, X-Member-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Auth is a middleware that extracts tenant and member IDs from headers
// v1 簡易認証: X-Tenant-ID, X-Member-ID ヘッダーを使用
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract tenant ID from header
		tenantIDStr := r.Header.Get("X-Tenant-ID")
		if tenantIDStr == "" {
			RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "X-Tenant-ID header is required", nil)
			return
		}

		tenantID := common.TenantID(tenantIDStr)
		if err := tenantID.Validate(); err != nil {
			RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid X-Tenant-ID format", nil)
			return
		}

		// Extract member ID from header (optional for some endpoints)
		memberIDStr := r.Header.Get("X-Member-ID")
		var memberID common.MemberID
		if memberIDStr != "" {
			memberID = common.MemberID(memberIDStr)
			if err := memberID.Validate(); err != nil {
				RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "Invalid X-Member-ID format", nil)
				return
			}
		}

		// Add IDs to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ContextKeyTenantID, tenantID)
		if memberIDStr != "" {
			ctx = context.WithValue(ctx, ContextKeyMemberID, memberID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Recover is a middleware that recovers from panics
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				RespondInternalError(w)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) (common.TenantID, bool) {
	tenantID, ok := ctx.Value(ContextKeyTenantID).(common.TenantID)
	return tenantID, ok
}

// GetMemberID extracts member ID from context
func GetMemberID(ctx context.Context) (common.MemberID, bool) {
	memberID, ok := ctx.Value(ContextKeyMemberID).(common.MemberID)
	return memberID, ok
}

