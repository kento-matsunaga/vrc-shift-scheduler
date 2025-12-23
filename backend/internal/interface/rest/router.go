package rest

import (
	"context"
	"net/http"
	"os"
	"strings"

	appaudit "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/audit"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
	appevent "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/event"
	applicense "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/license"
	appmember "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/member"
	apppayment "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/payment"
	appshift "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/shift"
	apptenant "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/tenant"
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
	// CORS設定: ALLOWED_ORIGINS環境変数で許可オリジンを指定
	// 未設定の場合は全オリジン許可（開発環境用）
	r.Use(CORSWithOrigins(os.Getenv("ALLOWED_ORIGINS")))

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

	// Billing guard dependencies
	tenantRepo := db.NewTenantRepository(dbPool)
	entitlementRepo := db.NewEntitlementRepository(dbPool)
	billingGuardDeps := BillingGuardDeps{
		TenantRepo:      tenantRepo,
		EntitlementRepo: entitlementRepo,
	}

	// API v1 ルート（認証必要）
	r.Route("/api/v1", func(r chi.Router) {
		// 認証ミドルウェアを適用（JWT優先、X-Tenant-IDフォールバック）
		r.Use(Auth(jwtManager))
		// 課金状態に基づくアクセス制御
		r.Use(BillingGuard(billingGuardDeps))

		// EventHandler dependencies
		eventRepo := db.NewEventRepository(dbPool)
		businessDayRepo := db.NewEventBusinessDayRepository(dbPool)
		groupAssignRepo := db.NewEventGroupAssignmentRepository(dbPool)
		eventHandler := NewEventHandler(
			appevent.NewCreateEventUsecase(eventRepo, businessDayRepo),
			appevent.NewListEventsUsecase(eventRepo),
			appevent.NewGetEventUsecase(eventRepo),
			appevent.NewUpdateEventUsecase(eventRepo),
			appevent.NewDeleteEventUsecase(eventRepo),
			appevent.NewGenerateBusinessDaysUsecase(eventRepo, businessDayRepo),
			appevent.NewGetEventGroupAssignmentsUsecase(eventRepo, groupAssignRepo),
			appevent.NewUpdateEventGroupAssignmentsUsecase(eventRepo, groupAssignRepo),
		)

		// BusinessDayHandler dependencies
		slotRepo := db.NewShiftSlotRepository(dbPool)
		templateRepo := db.NewShiftSlotTemplateRepository(dbPool)
		businessDayHandler := NewBusinessDayHandler(
			appevent.NewCreateBusinessDayUsecase(businessDayRepo, eventRepo, templateRepo, slotRepo),
			appevent.NewListBusinessDaysUsecase(businessDayRepo),
			appevent.NewGetBusinessDayUsecase(businessDayRepo),
			appevent.NewApplyTemplateUsecase(businessDayRepo, templateRepo, slotRepo),
		)

		// MemberHandler dependencies
		memberRepo := db.NewMemberRepository(dbPool)
		memberRoleRepo := db.NewMemberRoleRepository(dbPool)
		attendanceRepo := db.NewAttendanceRepository(dbPool)
		memberHandler := NewMemberHandler(
			appmember.NewCreateMemberUsecase(memberRepo),
			appmember.NewListMembersUsecase(memberRepo, memberRoleRepo),
			appmember.NewGetMemberUsecase(memberRepo, memberRoleRepo),
			appmember.NewDeleteMemberUsecase(memberRepo),
			appmember.NewUpdateMemberUsecase(memberRepo, memberRoleRepo),
			appmember.NewGetRecentAttendanceUsecase(memberRepo, attendanceRepo),
			appmember.NewBulkImportMembersUsecase(memberRepo, memberRoleRepo),
		)
		roleHandler := NewRoleHandler(dbPool)

		// ShiftSlotHandler dependencies (reusing slotRepo, businessDayRepo)
		assignmentRepo := db.NewShiftAssignmentRepository(dbPool)
		shiftSlotHandler := NewShiftSlotHandler(
			appshift.NewCreateShiftSlotUsecase(slotRepo, businessDayRepo),
			appshift.NewListShiftSlotsUsecase(slotRepo, assignmentRepo),
			appshift.NewGetShiftSlotUsecase(slotRepo, assignmentRepo),
		)

		// ShiftTemplateHandler dependencies (reusing templateRepo, slotRepo, businessDayRepo)
		shiftTemplateHandler := NewShiftTemplateHandler(
			appshift.NewCreateShiftTemplateUsecase(templateRepo),
			appshift.NewListShiftTemplatesUsecase(templateRepo),
			appshift.NewGetShiftTemplateUsecase(templateRepo),
			appshift.NewUpdateShiftTemplateUsecase(templateRepo),
			appshift.NewDeleteShiftTemplateUsecase(templateRepo),
			appshift.NewSaveBusinessDayAsTemplateUsecase(templateRepo, businessDayRepo, slotRepo),
		)

		// ShiftAssignmentHandler dependencies (reusing slotRepo, assignmentRepo, memberRepo, businessDayRepo)
		shiftAssignmentHandler := NewShiftAssignmentHandler(
			appshift.NewConfirmManualAssignmentUsecase(slotRepo, assignmentRepo, memberRepo),
			appshift.NewGetAssignmentsUsecase(assignmentRepo, memberRepo, slotRepo, businessDayRepo),
			appshift.NewGetAssignmentDetailUsecase(assignmentRepo, memberRepo, slotRepo, businessDayRepo),
			appshift.NewCancelAssignmentUsecase(assignmentRepo),
		)
		attendanceHandler := NewAttendanceHandler(dbPool)

		// ActualAttendanceHandler dependencies (reusing memberRepo, businessDayRepo, assignmentRepo)
		actualAttendanceHandler := NewActualAttendanceHandler(businessDayRepo, memberRepo, assignmentRepo)
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

			// Event配下のグループ割り当て
			r.Get("/{event_id}/groups", eventHandler.GetGroupAssignments)
			r.Put("/{event_id}/groups", eventHandler.UpdateGroupAssignments)

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
			r.Post("/bulk-import", memberHandler.BulkImportMembers)
			r.Get("/", memberHandler.GetMembers)
			r.Get("/recent-attendance", memberHandler.GetRecentAttendance)
			r.Get("/{member_id}", memberHandler.GetMemberDetail)
			r.Put("/{member_id}", memberHandler.UpdateMember)
			r.Delete("/{member_id}", memberHandler.DeleteMember)
		})

		// Role API
		r.Route("/roles", func(r chi.Router) {
			r.Post("/", roleHandler.CreateRole)
			r.Get("/", roleHandler.ListRoles)
			r.Get("/{role_id}", roleHandler.GetRole)
			r.Put("/{role_id}", roleHandler.UpdateRole)
			r.Delete("/{role_id}", roleHandler.DeleteRole)
		})

		// Member Group API
		memberGroupHandler := NewMemberGroupHandler(dbPool)
		r.Route("/member-groups", func(r chi.Router) {
			r.Post("/", memberGroupHandler.CreateGroup)
			r.Get("/", memberGroupHandler.ListGroups)
			r.Get("/{group_id}", memberGroupHandler.GetGroup)
			r.Put("/{group_id}", memberGroupHandler.UpdateGroup)
			r.Delete("/{group_id}", memberGroupHandler.DeleteGroup)
			r.Put("/{group_id}/members", memberGroupHandler.AssignMembers)
		})

		// Role Group API
		roleGroupHandler := NewRoleGroupHandler(dbPool)
		r.Route("/role-groups", func(r chi.Router) {
			r.Post("/", roleGroupHandler.CreateGroup)
			r.Get("/", roleGroupHandler.ListGroups)
			r.Get("/{group_id}", roleGroupHandler.GetGroup)
			r.Put("/{group_id}", roleGroupHandler.UpdateGroup)
			r.Delete("/{group_id}", roleGroupHandler.DeleteGroup)
			r.Put("/{group_id}/roles", roleGroupHandler.AssignRoles)
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

		// Admin API (テナント管理者のパスワード変更)
		r.Route("/admins", func(r chi.Router) {
			r.Post("/me/change-password", adminHandler.ChangePassword)
		})
	})

	// ============================================================
	// Admin Billing API (Cloudflare Access 保護 - 運営専用)
	// ============================================================
	// NOTE: このルートはテナントJWT認証とは完全に分離されています
	// 本番環境ではCloudflare Accessで保護されます
	r.Route("/api/v1/admin", func(r chi.Router) {
		// Cloudflare Access 認証ミドルウェア
		cfConfig := LoadCloudflareAccessConfig()
		r.Use(CloudflareAccessMiddleware(cfConfig))

		// Initialize dependencies for admin billing
		txManager := db.NewPgxTxManager(dbPool)
		licenseKeyRepo := db.NewLicenseKeyRepository(dbPool)
		billingAuditLogRepo := db.NewBillingAuditLogRepository(dbPool)

		adminLicenseKeyUsecase := applicense.NewAdminLicenseKeyUsecase(
			txManager,
			licenseKeyRepo,
			billingAuditLogRepo,
		)
		adminTenantUsecase := apptenant.NewAdminTenantUsecase(
			txManager,
			tenantRepo,
			entitlementRepo,
			billingAuditLogRepo,
		)
		adminAuditLogUsecase := appaudit.NewAdminAuditLogUsecase(
			billingAuditLogRepo,
		)
		adminBillingHandler := NewAdminBillingHandler(
			adminLicenseKeyUsecase,
			adminTenantUsecase,
			adminAuditLogUsecase,
		)

		// License Key Management
		r.Route("/license-keys", func(r chi.Router) {
			r.Post("/", adminBillingHandler.GenerateLicenseKeys)
			r.Get("/", adminBillingHandler.ListLicenseKeys)
			r.Patch("/{id}", adminBillingHandler.UpdateLicenseKey)
		})

		// Tenant Management
		r.Route("/tenants", func(r chi.Router) {
			r.Get("/", adminBillingHandler.ListTenants)
			r.Get("/{id}", adminBillingHandler.GetTenantDetail)
			r.Patch("/{id}/status", adminBillingHandler.UpdateTenantStatus)
		})

		// Audit Logs
		r.Get("/audit-logs", adminBillingHandler.ListAuditLogs)
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
	// group_ids パラメータで対象グループを指定可能（カンマ区切り）
	publicMemberRepo := db.NewMemberRepository(dbPool)
	publicMemberRoleRepo := db.NewMemberRoleRepository(dbPool)
	publicAttendanceRepo := db.NewAttendanceRepository(dbPool)
	publicMemberGroupRepo := db.NewMemberGroupRepository(dbPool)
	publicMemberHandler := NewMemberHandler(
		appmember.NewCreateMemberUsecase(publicMemberRepo),
		appmember.NewListMembersUsecase(publicMemberRepo, publicMemberRoleRepo),
		appmember.NewGetMemberUsecase(publicMemberRepo, publicMemberRoleRepo),
		appmember.NewDeleteMemberUsecase(publicMemberRepo),
		appmember.NewUpdateMemberUsecase(publicMemberRepo, publicMemberRoleRepo),
		appmember.NewGetRecentAttendanceUsecase(publicMemberRepo, publicAttendanceRepo),
		appmember.NewBulkImportMembersUsecase(publicMemberRepo, publicMemberRoleRepo),
	)
	r.Get("/api/v1/public/members", func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			RespondBadRequest(w, "tenant_id is required")
			return
		}

		// group_ids パラメータの取得（カンマ区切り）
		groupIDsParam := r.URL.Query().Get("group_ids")
		var allowedMemberIDs map[string]struct{}
		if groupIDsParam != "" {
			groupIDStrs := strings.Split(groupIDsParam, ",")
			allowedMemberIDs = make(map[string]struct{})

			// 各グループからメンバーIDを取得
			for _, gidStr := range groupIDStrs {
				gidStr = strings.TrimSpace(gidStr)
				if gidStr == "" {
					continue
				}
				gid, err := common.ParseMemberGroupID(gidStr)
				if err != nil {
					continue
				}
				memberIDs, err := publicMemberGroupRepo.FindMemberIDsByGroupID(r.Context(), gid)
				if err != nil {
					continue
				}
				for _, mid := range memberIDs {
					allowedMemberIDs[mid.String()] = struct{}{}
				}
			}
		}

		// Contextにテナント情報とフィルター用メンバーIDを設定
		ctx := context.WithValue(r.Context(), ContextKeyTenantID, common.TenantID(tenantID))
		if allowedMemberIDs != nil {
			ctx = context.WithValue(ctx, ContextKeyAllowedMemberIDs, allowedMemberIDs)
		}
		r = r.WithContext(ctx)
		publicMemberHandler.GetMembers(w, r)
	})

	// License Claim API（認証不要、レート制限あり）
	r.Route("/api/v1/public/license", func(r chi.Router) {
		// Initialize dependencies for license claim
		txManager := db.NewPgxTxManager(dbPool)
		licenseKeyRepo := db.NewLicenseKeyRepository(dbPool)
		billingAuditLogRepo := db.NewBillingAuditLogRepository(dbPool)
		claimRateLimiter := DefaultClaimRateLimiter()

		claimUsecase := applicense.NewLicenseClaimUsecase(
			txManager,
			tenantRepo,
			adminRepo,
			licenseKeyRepo,
			entitlementRepo,
			billingAuditLogRepo,
			passwordHasher,
		)
		licenseClaimHandler := NewLicenseClaimHandler(claimUsecase, claimRateLimiter)

		r.Post("/claim", licenseClaimHandler.Claim)
	})

	// Stripe Webhook API（認証不要、署名検証のみ）
	r.Route("/api/v1/stripe", func(r chi.Router) {
		// Initialize dependencies for Stripe webhook
		txManager := db.NewPgxTxManager(dbPool)
		subscriptionRepo := db.NewSubscriptionRepository(dbPool)
		webhookEventRepo := db.NewWebhookEventRepository(dbPool)
		billingAuditLogRepo := db.NewBillingAuditLogRepository(dbPool)

		stripeWebhookUsecase := apppayment.NewStripeWebhookUsecase(
			txManager,
			tenantRepo,
			subscriptionRepo,
			entitlementRepo,
			webhookEventRepo,
			billingAuditLogRepo,
		)
		stripeWebhookHandler := NewStripeWebhookHandler(stripeWebhookUsecase)

		r.Post("/webhook", stripeWebhookHandler.HandleWebhook)
	})

	return r
}
