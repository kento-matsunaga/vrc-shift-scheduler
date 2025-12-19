package rest

import (
	"context"
	"net/http"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(dbPool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	// グローバルミドルウェア
	r.Use(middleware.RequestID)
	r.Use(Recover)
	r.Use(Logger)
	r.Use(CORS)

	// ヘルスチェック（認証不要）
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		RespondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// 認証基盤の初期化
	jwtManager := security.NewJWTManager()
	adminRepo := db.NewAdminRepository(dbPool)
	passwordHasher := security.NewBcryptHasher()
	loginUsecase := auth.NewLoginUsecase(adminRepo, passwordHasher, jwtManager)
	authHandler := NewAuthHandler(loginUsecase)

	// 招待受理（認証不要）
	invitationHandler := NewInvitationHandler(dbPool)
	r.Post("/api/v1/invitations/accept/{token}", invitationHandler.AcceptInvitation)

	// 認証不要ルート
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
	})

	// API v1 ルート（認証必要）
	r.Route("/api/v1", func(r chi.Router) {
		// 認証ミドルウェアを適用（JWT優先、X-Tenant-IDフォールバック）
		r.Use(Auth(jwtManager))

		// ハンドラの初期化
		eventHandler := NewEventHandler(dbPool)
		businessDayHandler := NewBusinessDayHandler(dbPool)
		memberHandler := NewMemberHandler(dbPool)
		roleHandler := NewRoleHandler(dbPool)
		shiftSlotHandler := NewShiftSlotHandler(dbPool)
		shiftTemplateHandler := NewShiftTemplateHandler(dbPool)
		shiftAssignmentHandler := NewShiftAssignmentHandler(dbPool)
		attendanceHandler := NewAttendanceHandler(dbPool)
		actualAttendanceHandler := NewActualAttendanceHandler(dbPool)
		tenantHandler := NewTenantHandler(dbPool)
		adminHandler := NewAdminHandler(dbPool)

		// Event API
		r.Route("/events", func(r chi.Router) {
			r.Post("/", eventHandler.CreateEvent)
			r.Get("/", eventHandler.ListEvents)

			// 単一イベントのGET/PUT/DELETE
			r.Get("/{event_id}", eventHandler.GetEvent)
			r.MethodFunc("PUT", "/{event_id}", eventHandler.UpdateEvent)
			r.Delete("/{event_id}", eventHandler.DeleteEvent)

			// Event配下のBusinessDay
			r.Post("/{event_id}/business-days", businessDayHandler.CreateBusinessDay)
			r.Get("/{event_id}/business-days", businessDayHandler.ListBusinessDays)

			// Event配下の営業日生成
			r.Post("/{event_id}/generate-business-days", eventHandler.GenerateBusinessDays)

			// Event配下のShiftTemplate
			r.Post("/{event_id}/templates", shiftTemplateHandler.CreateTemplate)
			r.Get("/{event_id}/templates", shiftTemplateHandler.ListTemplates)
			r.Get("/{event_id}/templates/{template_id}", shiftTemplateHandler.GetTemplate)
			r.Put("/{event_id}/templates/{template_id}", shiftTemplateHandler.UpdateTemplate)
			r.Delete("/{event_id}/templates/{template_id}", shiftTemplateHandler.DeleteTemplate)
		})

		// BusinessDay API
		r.Route("/business-days", func(r chi.Router) {
			r.Get("/{business_day_id}", businessDayHandler.GetBusinessDay)

			// BusinessDay配下のShiftSlot
			r.Post("/{business_day_id}/shift-slots", shiftSlotHandler.CreateShiftSlot)
			r.Get("/{business_day_id}/shift-slots", shiftSlotHandler.GetShiftSlots)

			// BusinessDayからShiftTemplateを作成
			r.Post("/{business_day_id}/save-as-template", shiftTemplateHandler.SaveBusinessDayAsTemplate)

			// BusinessDayにShiftTemplateを適用
			r.Post("/{business_day_id}/apply-template", businessDayHandler.ApplyTemplate)
		})

		// Member API
		r.Route("/members", func(r chi.Router) {
			r.Post("/", memberHandler.CreateMember)
			r.Get("/", memberHandler.GetMembers)
			r.Get("/recent-attendance", memberHandler.GetRecentAttendance)
			r.Get("/{member_id}", memberHandler.GetMemberDetail)
			r.Put("/{member_id}", memberHandler.UpdateMember)
		})

		// Role API
		r.Route("/roles", func(r chi.Router) {
			r.Post("/", roleHandler.CreateRole)
			r.Get("/", roleHandler.ListRoles)
			r.Get("/{role_id}", roleHandler.GetRole)
			r.Put("/{role_id}", roleHandler.UpdateRole)
			r.Delete("/{role_id}", roleHandler.DeleteRole)
		})

		// Actual Attendance API（本出席 - 実際のシフト割り当て実績）
		r.Route("/actual-attendance", func(r chi.Router) {
			r.Get("/", actualAttendanceHandler.GetRecentActualAttendance)
		})

		// ShiftSlot API
		r.Route("/shift-slots", func(r chi.Router) {
			r.Get("/{slot_id}", shiftSlotHandler.GetShiftSlotDetail)
		})

		// ShiftAssignment API
		r.Route("/shift-assignments", func(r chi.Router) {
			r.Post("/", shiftAssignmentHandler.ConfirmAssignment)
			r.Get("/", shiftAssignmentHandler.GetAssignments)
			r.Get("/{assignment_id}", shiftAssignmentHandler.GetAssignmentDetail)
			r.Delete("/{assignment_id}", shiftAssignmentHandler.CancelAssignment)
		})

		// Attendance API（管理用）
		r.Route("/attendance/collections", func(r chi.Router) {
			r.Get("/", attendanceHandler.ListCollections)
			r.Post("/", attendanceHandler.CreateCollection)
			r.Get("/{collection_id}", attendanceHandler.GetCollection)
			r.Post("/{collection_id}/close", attendanceHandler.CloseCollection)
			r.Get("/{collection_id}/responses", attendanceHandler.GetResponses)
		})

		// Schedule API（管理用）
		scheduleHandler := NewScheduleHandler(dbPool)
		r.Route("/schedules", func(r chi.Router) {
			r.Get("/", scheduleHandler.ListSchedules)
			r.Post("/", scheduleHandler.CreateSchedule)
			r.Get("/{schedule_id}", scheduleHandler.GetSchedule)
			r.Post("/{schedule_id}/decide", scheduleHandler.DecideSchedule)
			r.Post("/{schedule_id}/close", scheduleHandler.CloseSchedule)
			r.Get("/{schedule_id}/responses", scheduleHandler.GetResponses)
		})

		// Invitation API（管理者のみ）
		r.Route("/invitations", func(r chi.Router) {
			r.Post("/", invitationHandler.InviteAdmin)
		})

		// Tenant API
		r.Route("/tenants", func(r chi.Router) {
			r.Get("/me", tenantHandler.GetCurrentTenant)
			r.Put("/me", tenantHandler.UpdateCurrentTenant)
		})

		// Admin API
		r.Route("/admins", func(r chi.Router) {
			r.Post("/me/change-password", adminHandler.ChangePassword)
		})
	})

	// Public API（認証不要）
	r.Route("/api/v1/public/attendance", func(r chi.Router) {
		attendanceHandler := NewAttendanceHandler(dbPool)
		r.Get("/{token}", attendanceHandler.GetCollectionByToken)
		r.Post("/{token}/responses", attendanceHandler.SubmitResponse)
	})

	r.Route("/api/v1/public/schedules", func(r chi.Router) {
		scheduleHandler := NewScheduleHandler(dbPool)
		r.Get("/{token}", scheduleHandler.GetScheduleByToken)
		r.Post("/{token}/responses", scheduleHandler.SubmitResponse)
	})

	// 公開ページ用メンバー一覧API（認証不要）
	// NOTE: MVPでは簡易実装としてテナントIDを指定してメンバー一覧を取得可能
	r.Get("/api/v1/public/members", func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			RespondBadRequest(w, "tenant_id is required")
			return
		}
		// memberHandler を使用
		memberHandler := NewMemberHandler(dbPool)
		// Contextにテナント情報を設定
		ctx := context.WithValue(r.Context(), ContextKeyTenantID, common.TenantID(tenantID))
		r = r.WithContext(ctx)
		memberHandler.GetMembers(w, r)
	})

	return r
}
