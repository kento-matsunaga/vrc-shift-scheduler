package rest

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// ContextKeyTenantID is the context key for tenant ID
	ContextKeyTenantID ContextKey = "tenant_id"
	// ContextKeyMemberID is the context key for member ID
	ContextKeyMemberID ContextKey = "member_id"
	// ContextKeyAdminID is the context key for admin ID (JWT認証時)
	ContextKeyAdminID ContextKey = "admin_id"
	// ContextKeyRole is the context key for admin role (JWT認証時)
	ContextKeyRole ContextKey = "role"
	// ContextKeyAllowedMemberIDs is the context key for allowed member IDs filter (map[string]struct{})
	ContextKeyAllowedMemberIDs ContextKey = "allowed_member_ids"
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

// CORSWithOrigins creates a CORS middleware with specified allowed origins
// If allowedOrigins is empty, it falls back to allowing all origins (development mode)
func CORSWithOrigins(allowedOrigins string) func(http.Handler) http.Handler {
	// Parse allowed origins into a map for fast lookup
	origins := make(map[string]bool)
	if allowedOrigins != "" {
		for _, origin := range strings.Split(allowedOrigins, ",") {
			origins[strings.TrimSpace(origin)] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if the origin is allowed
			if len(origins) > 0 {
				// Production mode: check against allowed origins
				if origins[origin] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
				}
				// If origin not in allowed list, don't set CORS headers (request will be blocked by browser)
			} else {
				// Development mode: allow all origins
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tenant-ID, X-Member-ID, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORS is a middleware that handles CORS headers (allows all origins - for backward compatibility)
// Deprecated: Use CORSWithOrigins with ALLOWED_ORIGINS environment variable instead
func CORS(next http.Handler) http.Handler {
	return CORSWithOrigins("")(next)
}

// Auth is a middleware that extracts tenant and member IDs from headers
// JWT認証優先、フォールバックで v1 簡易認証（X-Tenant-ID, X-Member-ID）
func Auth(tokenVerifier security.TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Authorization: Bearer があればJWT検証
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
				token := authHeader[7:]
				claims, err := tokenVerifier.Verify(token)
				if err != nil {
					// JWT検証失敗 → 401
					RespondError(w, http.StatusUnauthorized, "ERR_UNAUTHORIZED", "Invalid or expired token", nil)
					return
				}

				// JWT検証成功 → context に tenant_id, admin_id, role をセット
				ctx := r.Context()
				ctx = context.WithValue(ctx, ContextKeyTenantID, common.TenantID(claims.TenantID))
				ctx = context.WithValue(ctx, ContextKeyAdminID, common.AdminID(claims.AdminID))
				ctx = context.WithValue(ctx, ContextKeyRole, claims.Role)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// JWT がない → 従来の X-Tenant-ID 認証にフォールバック（段階移行）
			tenantIDStr := r.Header.Get("X-Tenant-ID")
			if tenantIDStr == "" {
				RespondError(w, http.StatusBadRequest, "ERR_INVALID_REQUEST", "X-Tenant-ID header or Authorization header is required", nil)
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

// GetAdminID extracts admin ID from context
func GetAdminID(ctx context.Context) (common.AdminID, bool) {
	adminID, ok := ctx.Value(ContextKeyAdminID).(common.AdminID)
	return adminID, ok
}

// GetRole extracts admin role from context
func GetRole(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(ContextKeyRole).(string)
	return role, ok
}

// TenantStatusMiddleware はテナントのステータスをチェックし、
// suspended状態の場合はアクセスを拒否するミドルウェア
// grace状態はログイン可能、suspended状態はログイン不可
func TenantStatusMiddleware(tenantRepo tenant.TenantRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// コンテキストからテナントIDを取得
			tenantID, ok := GetTenantID(r.Context())
			if !ok {
				// テナントIDがない場合はスキップ（Auth middleware で既に処理済みのはず）
				next.ServeHTTP(w, r)
				return
			}

			// テナントを取得
			t, err := tenantRepo.FindByID(r.Context(), tenantID)
			if err != nil {
				// テナント取得失敗時は後続の処理に任せる
				log.Printf("[TenantStatusMiddleware] Failed to find tenant %s: %v", tenantID, err)
				next.ServeHTTP(w, r)
				return
			}

			// suspended状態の場合はアクセス拒否
			if t.Status() == tenant.TenantStatusSuspended {
				RespondError(w, http.StatusForbidden, "ERR_TENANT_SUSPENDED",
					"お支払いが確認できないため、サービスを一時停止しています。", nil)
				return
			}

			// grace, active, pending_payment は通過
			next.ServeHTTP(w, r)
		})
	}
}
