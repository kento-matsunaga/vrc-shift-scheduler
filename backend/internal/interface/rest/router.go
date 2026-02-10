package rest

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	appannouncement "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/announcement"
	appattendance "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/attendance"
	appaudit "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/audit"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/auth"
	appcalendar "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/calendar"
	appevent "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/event"
	appimport "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/import"
	applicense "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/license"
	appmember "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/member"
	appmembergroup "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/member_group"
	apppayment "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/payment"
	approle "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/role"
	approlegroup "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/role_group"
	appschedule "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/schedule"
	appshift "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/shift"
	appsystem "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/system"
	apptenant "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/tenant"
	apptutorial "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/tutorial"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/services"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/tenant"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/clock"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/db"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/email"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/security"
	infrastripe "github.com/erenoa/vrc-shift-scheduler/backend/internal/infra/stripe"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// initEmailService creates an email service based on environment configuration
// If Resend is configured, it returns ResendEmailService; otherwise MockEmailService
func initEmailService() services.EmailService {
	baseURL := os.Getenv("INVITATION_BASE_URL")
	if baseURL == "" {
		baseURL = "https://vrcshift.com"
	}

	// Check if Resend is configured
	apiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")
	if apiKey == "" || fromEmail == "" {
		slog.Info("Resend not configured, using mock email service")
		return email.NewMockEmailService(baseURL)
	}

	// Validate API key format
	if len(apiKey) < 3 || apiKey[:3] != "re_" {
		slog.Warn("RESEND_API_KEY does not start with 're_', may be invalid")
	}

	// Validate email format (basic check)
	if !strings.Contains(fromEmail, "@") || !strings.Contains(fromEmail, ".") {
		slog.Warn("RESEND_FROM_EMAIL appears to be invalid", "from_email", fromEmail)
	}

	slog.Info("Resend configured", "from_email", fromEmail)
	return email.NewResendEmailService(apiKey, fromEmail, baseURL)
}

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

	// InvitationHandler dependencies
	invitationRepo := db.NewInvitationRepository(dbPool)
	invitationClock := &clock.RealClock{}
	invitationTenantRepo := db.NewTenantRepository(dbPool)
	invitationEmailService := initEmailService()
	invitationHandler := NewInvitationHandler(
		auth.NewInviteAdminUsecase(adminRepo, invitationRepo, invitationTenantRepo, invitationEmailService, invitationClock),
		auth.NewAcceptInvitationUsecase(adminRepo, invitationRepo, passwordHasher, invitationClock),
	)

	// 招待受理（認証不要）
	r.Post("/api/v1/invitations/accept/{token}", invitationHandler.AcceptInvitation)

	// PasswordResetHandler dependencies (public endpoints)
	passwordResetClock := &clock.RealClock{}
	licenseKeyRepo := db.NewLicenseKeyRepository(dbPool)
	billingAuditLogRepo := db.NewBillingAuditLogRepository(dbPool)
	passwordResetTokenRepo := db.NewPasswordResetTokenRepository(dbPool)
	passwordResetTxManager := db.NewPgxTxManager(dbPool)
	checkPasswordResetStatusUsecase := auth.NewCheckPasswordResetStatusUsecase(adminRepo, passwordResetClock)
	verifyAndResetPasswordUsecase := auth.NewVerifyAndResetPasswordUsecase(adminRepo, licenseKeyRepo, passwordHasher, passwordResetClock, billingAuditLogRepo)
	requestPasswordResetUsecase := auth.NewRequestPasswordResetUsecase(adminRepo, passwordResetTokenRepo, invitationEmailService, passwordResetClock)
	resetPasswordWithTokenUsecase := auth.NewResetPasswordWithTokenUsecase(adminRepo, passwordResetTokenRepo, passwordHasher, passwordResetClock, passwordResetTxManager)
	passwordResetRateLimiter := DefaultPasswordResetRateLimiter()

	// 認証不要ルート
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Post("/login", authHandler.Login)
		// Password reset public endpoints (with rate limiting)
		passwordResetHandler := NewPasswordResetHandler(nil, checkPasswordResetStatusUsecase, verifyAndResetPasswordUsecase, requestPasswordResetUsecase, resetPasswordWithTokenUsecase, passwordResetRateLimiter)
		r.Get("/password-reset-status", passwordResetHandler.CheckPasswordResetStatus)
		r.Post("/reset-password", passwordResetHandler.ResetPassword)
		// New email-based password reset endpoints
		r.Post("/forgot-password", passwordResetHandler.ForgotPassword)
		r.Post("/reset-password-with-token", passwordResetHandler.ResetPasswordWithToken)
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
		// テナントステータスチェック（suspended状態はアクセス拒否）
		r.Use(TenantStatusMiddleware(tenantRepo))
		// 課金状態に基づくアクセス制御
		r.Use(BillingGuard(billingGuardDeps))

		// ManagerPermissions repository (shared by permission checker and handler)
		managerPermissionsRepo := db.NewManagerPermissionsRepository(dbPool)

		// PermissionChecker for manager permission enforcement
		permissionChecker := NewPermissionChecker(
			apptenant.NewCheckManagerPermissionUsecase(managerPermissionsRepo),
		)

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
		instanceRepo := db.NewInstanceRepository(dbPool)
		businessDayHandler := NewBusinessDayHandler(
			appevent.NewCreateBusinessDayUsecase(businessDayRepo, eventRepo, templateRepo, slotRepo, instanceRepo),
			appevent.NewListBusinessDaysUsecase(businessDayRepo),
			appevent.NewGetBusinessDayUsecase(businessDayRepo),
			appevent.NewApplyTemplateUsecase(businessDayRepo, templateRepo, slotRepo, instanceRepo),
			appevent.NewDeleteBusinessDayUsecase(businessDayRepo),
		)

		// InstanceHandler dependencies
		assignmentRepo := db.NewShiftAssignmentRepository(dbPool)
		instanceTxManager := db.NewPgxTxManager(dbPool)
		instanceHandler := NewInstanceHandler(
			appshift.NewCreateInstanceUsecase(instanceRepo, eventRepo),
			appshift.NewListInstancesUsecase(instanceRepo),
			appshift.NewGetInstanceUsecase(instanceRepo),
			appshift.NewUpdateInstanceUsecase(instanceRepo),
			appshift.NewDeleteInstanceUsecase(instanceTxManager, instanceRepo, slotRepo, assignmentRepo),
		)

		// RoleHandler dependencies (needed by MemberHandler too)
		roleRepo := db.NewRoleRepository(dbPool)

		// MemberHandler dependencies
		memberRepo := db.NewMemberRepository(dbPool)
		memberRoleRepo := db.NewMemberRoleRepository(dbPool)
		attendanceRepo := db.NewAttendanceRepository(dbPool)
		memberTxManager := db.NewPgxTxManager(dbPool)
		memberHandler := NewMemberHandler(
			appmember.NewCreateMemberUsecase(memberRepo, memberRoleRepo),
			appmember.NewListMembersUsecase(memberRepo, memberRoleRepo),
			appmember.NewGetMemberUsecase(memberRepo, memberRoleRepo),
			appmember.NewDeleteMemberUsecase(memberRepo),
			appmember.NewUpdateMemberUsecase(memberRepo, memberRoleRepo),
			appmember.NewGetRecentAttendanceUsecase(memberRepo, attendanceRepo),
			appmember.NewBulkImportMembersUsecase(memberRepo, memberRoleRepo),
			appmember.NewBulkUpdateRolesUsecase(memberRepo, memberRoleRepo, roleRepo, memberTxManager),
		)

		// RoleHandler
		roleHandler := NewRoleHandler(
			approle.NewCreateRoleUsecase(roleRepo),
			approle.NewUpdateRoleUsecase(roleRepo),
			approle.NewGetRoleUsecase(roleRepo),
			approle.NewListRolesUsecase(roleRepo),
			approle.NewDeleteRoleUsecase(roleRepo),
		)

		// ShiftSlotHandler dependencies (reusing slotRepo, businessDayRepo, instanceRepo, assignmentRepo)
		slotTxManager := db.NewPgxTxManager(dbPool)
		shiftSlotHandler := NewShiftSlotHandler(
			appshift.NewCreateShiftSlotUsecase(slotRepo, businessDayRepo, instanceRepo),
			appshift.NewListShiftSlotsUsecase(slotRepo, assignmentRepo),
			appshift.NewGetShiftSlotUsecase(slotRepo, assignmentRepo),
			appshift.NewDeleteShiftSlotUsecase(slotRepo, assignmentRepo),
			appshift.NewDeleteSlotsByInstanceUsecase(slotTxManager, slotRepo, assignmentRepo),
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

		// AttendanceHandler dependencies (reusing attendanceRepo, memberRepo, roleRepo)
		systemClock := &clock.RealClock{}
		txManager := db.NewPgxTxManager(dbPool)
		attendanceHandler := NewAttendanceHandler(
			appattendance.NewCreateCollectionUsecase(attendanceRepo, roleRepo, txManager, systemClock),
			appattendance.NewSubmitResponseUsecase(attendanceRepo, txManager, systemClock),
			appattendance.NewCloseCollectionUsecase(attendanceRepo, systemClock),
			appattendance.NewDeleteCollectionUsecase(attendanceRepo, systemClock),
			appattendance.NewUpdateCollectionUsecase(attendanceRepo, txManager, systemClock),
			appattendance.NewGetCollectionUsecase(attendanceRepo),
			appattendance.NewGetCollectionByTokenUsecase(attendanceRepo),
			appattendance.NewGetResponsesUsecase(attendanceRepo, memberRepo),
			appattendance.NewListCollectionsUsecase(attendanceRepo),
			appattendance.NewGetMemberResponsesUsecase(attendanceRepo),
			appattendance.NewGetAllPublicResponsesUsecase(attendanceRepo, memberRepo),
			appattendance.NewAdminUpdateResponseUsecase(attendanceRepo, memberRepo, txManager, systemClock),
		)

		// ActualAttendanceHandler dependencies (reusing memberRepo, businessDayRepo, assignmentRepo)
		actualAttendanceHandler := NewActualAttendanceHandler(businessDayRepo, memberRepo, assignmentRepo)

		// TenantHandler dependencies (reusing tenantRepo from billing guard)
		tenantHandler := NewTenantHandler(
			apptenant.NewGetTenantUsecase(tenantRepo),
			apptenant.NewUpdateTenantUsecase(tenantRepo),
		)

		// AdminHandler dependencies (reusing adminRepo and passwordHasher from auth setup)
		adminHandler := NewAdminHandler(
			auth.NewChangePasswordUsecase(adminRepo, passwordHasher),
			auth.NewChangeEmailUsecase(adminRepo, passwordHasher),
		)

		// PasswordResetHandler dependencies (authenticated endpoint - no rate limiting needed)
		allowPasswordResetUsecase := auth.NewAllowPasswordResetUsecase(adminRepo, systemClock)
		authPasswordResetHandler := NewPasswordResetHandler(allowPasswordResetUsecase, nil, nil, nil, nil, nil)

		// Event API
		r.Route("/events", func(r chi.Router) {
			// 権限チェック付きルート
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateEvent)).Post("/", eventHandler.CreateEvent)
			r.Get("/", eventHandler.ListEvents)

			// 単一イベントのGET/PUT/DELETE
			r.Get("/{event_id}", eventHandler.GetEvent)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Put("/{event_id}", eventHandler.UpdateEvent)
			r.With(permissionChecker.RequirePermission(tenant.PermissionDeleteEvent)).Delete("/{event_id}", eventHandler.DeleteEvent)

			// Event配下のBusinessDay
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateEvent)).Post("/{event_id}/business-days", businessDayHandler.CreateBusinessDay)
			r.Get("/{event_id}/business-days", businessDayHandler.ListBusinessDays)

			// Event配下の営業日生成
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateEvent)).Post("/{event_id}/generate-business-days", eventHandler.GenerateBusinessDays)

			// Event配下のグループ割り当て
			r.Get("/{event_id}/groups", eventHandler.GetGroupAssignments)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Put("/{event_id}/groups", eventHandler.UpdateGroupAssignments)

			// Event配下のShiftTemplate
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Post("/{event_id}/templates", shiftTemplateHandler.CreateTemplate)
			r.Get("/{event_id}/templates", shiftTemplateHandler.ListTemplates)
			r.Get("/{event_id}/templates/{template_id}", shiftTemplateHandler.GetTemplate)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Put("/{event_id}/templates/{template_id}", shiftTemplateHandler.UpdateTemplate)
			r.With(permissionChecker.RequirePermission(tenant.PermissionDeleteEvent)).Delete("/{event_id}/templates/{template_id}", shiftTemplateHandler.DeleteTemplate)

			// Event配下のInstance
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Post("/{event_id}/instances", instanceHandler.CreateInstance)
			r.Get("/{event_id}/instances", instanceHandler.GetInstances)
		})

		// Instance API
		r.Route("/instances", func(r chi.Router) {
			r.Get("/{instance_id}", instanceHandler.GetInstance)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Put("/{instance_id}", instanceHandler.UpdateInstance)
			r.Get("/{instance_id}/deletable", instanceHandler.CheckInstanceDeletable)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Delete("/{instance_id}", instanceHandler.DeleteInstance)
		})

		// BusinessDay API
		r.Route("/business-days", func(r chi.Router) {
			r.Get("/{business_day_id}", businessDayHandler.GetBusinessDay)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Delete("/{business_day_id}", businessDayHandler.DeleteBusinessDay)

			// BusinessDay配下のShiftSlot
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditShift)).Post("/{business_day_id}/shift-slots", shiftSlotHandler.CreateShiftSlot)
			r.Get("/{business_day_id}/shift-slots", shiftSlotHandler.GetShiftSlots)

			// BusinessDay配下のインスタンス別シフト枠一括削除
			r.Get("/{business_day_id}/instances/{instance_id}/slots/deletable", shiftSlotHandler.CheckSlotsByInstanceDeletable)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditShift)).Delete("/{business_day_id}/instances/{instance_id}/slots", shiftSlotHandler.DeleteSlotsByInstance)

			// BusinessDayからShiftTemplateを作成
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Post("/{business_day_id}/save-as-template", shiftTemplateHandler.SaveBusinessDayAsTemplate)

			// BusinessDayにShiftTemplateを適用
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditEvent)).Post("/{business_day_id}/apply-template", businessDayHandler.ApplyTemplate)
		})

		// Member API
		r.Route("/members", func(r chi.Router) {
			r.With(permissionChecker.RequirePermission(tenant.PermissionAddMember)).Post("/", memberHandler.CreateMember)
			r.With(permissionChecker.RequirePermission(tenant.PermissionAddMember)).Post("/bulk-import", memberHandler.BulkImportMembers)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditMember)).Post("/bulk-update-roles", memberHandler.BulkUpdateRoles)
			r.Get("/", memberHandler.GetMembers)
			r.Get("/recent-attendance", memberHandler.GetRecentAttendance)
			r.Get("/{member_id}", memberHandler.GetMemberDetail)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditMember)).Put("/{member_id}", memberHandler.UpdateMember)
			r.With(permissionChecker.RequirePermission(tenant.PermissionDeleteMember)).Delete("/{member_id}", memberHandler.DeleteMember)
		})

		// Role API
		r.Route("/roles", func(r chi.Router) {
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageRoles)).Post("/", roleHandler.CreateRole)
			r.Get("/", roleHandler.ListRoles)
			r.Get("/{role_id}", roleHandler.GetRole)
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageRoles)).Put("/{role_id}", roleHandler.UpdateRole)
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageRoles)).Delete("/{role_id}", roleHandler.DeleteRole)
		})

		// MemberGroupHandler dependencies
		memberGroupRepo := db.NewMemberGroupRepository(dbPool)
		memberGroupHandler := NewMemberGroupHandler(
			appmembergroup.NewCreateGroupUsecase(memberGroupRepo),
			appmembergroup.NewUpdateGroupUsecase(memberGroupRepo),
			appmembergroup.NewGetGroupUsecase(memberGroupRepo),
			appmembergroup.NewListGroupsUsecase(memberGroupRepo),
			appmembergroup.NewDeleteGroupUsecase(memberGroupRepo),
			appmembergroup.NewAssignMembersUsecase(memberGroupRepo),
		)
		r.Route("/member-groups", func(r chi.Router) {
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageGroups)).Post("/", memberGroupHandler.CreateGroup)
			r.Get("/", memberGroupHandler.ListGroups)
			r.Get("/{group_id}", memberGroupHandler.GetGroup)
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageGroups)).Put("/{group_id}", memberGroupHandler.UpdateGroup)
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageGroups)).Delete("/{group_id}", memberGroupHandler.DeleteGroup)
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageGroups)).Put("/{group_id}/members", memberGroupHandler.AssignMembers)
		})

		// RoleGroupHandler dependencies
		roleGroupRepo := db.NewRoleGroupRepository(dbPool)
		roleGroupHandler := NewRoleGroupHandler(
			approlegroup.NewCreateGroupUsecase(roleGroupRepo),
			approlegroup.NewUpdateGroupUsecase(roleGroupRepo),
			approlegroup.NewGetGroupUsecase(roleGroupRepo),
			approlegroup.NewListGroupsUsecase(roleGroupRepo),
			approlegroup.NewDeleteGroupUsecase(roleGroupRepo),
			approlegroup.NewAssignRolesUsecase(roleGroupRepo),
		)
		r.Route("/role-groups", func(r chi.Router) {
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageGroups)).Post("/", roleGroupHandler.CreateGroup)
			r.Get("/", roleGroupHandler.ListGroups)
			r.Get("/{group_id}", roleGroupHandler.GetGroup)
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageGroups)).Put("/{group_id}", roleGroupHandler.UpdateGroup)
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageGroups)).Delete("/{group_id}", roleGroupHandler.DeleteGroup)
			r.With(permissionChecker.RequirePermission(tenant.PermissionManageGroups)).Put("/{group_id}/roles", roleGroupHandler.AssignRoles)
		})

		// Actual Attendance API（本出席 - 実際のシフト割り当て実績）
		r.Route("/actual-attendance", func(r chi.Router) {
			r.Get("/", actualAttendanceHandler.GetRecentActualAttendance)
		})

		// ShiftSlot API
		r.Route("/shift-slots", func(r chi.Router) {
			r.Get("/{slot_id}", shiftSlotHandler.GetShiftSlotDetail)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditShift)).Delete("/{slot_id}", shiftSlotHandler.DeleteShiftSlot)
		})

		// ShiftAssignment API
		r.Route("/shift-assignments", func(r chi.Router) {
			r.With(permissionChecker.RequirePermission(tenant.PermissionAssignShift)).Post("/", shiftAssignmentHandler.ConfirmAssignment)
			r.Get("/", shiftAssignmentHandler.GetAssignments)
			r.Get("/{assignment_id}", shiftAssignmentHandler.GetAssignmentDetail)
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditShift)).Delete("/{assignment_id}", shiftAssignmentHandler.CancelAssignment)
		})

		// Attendance API（管理用）
		r.Route("/attendance/collections", func(r chi.Router) {
			r.Get("/", attendanceHandler.ListCollections)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateAttendance)).Post("/", attendanceHandler.CreateCollection)
			r.Get("/{collection_id}", attendanceHandler.GetCollection)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateAttendance)).Post("/{collection_id}/close", attendanceHandler.CloseCollection)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateAttendance)).Delete("/{collection_id}", attendanceHandler.DeleteCollection)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateAttendance)).Put("/{collection_id}", attendanceHandler.UpdateCollection)
			r.Get("/{collection_id}/responses", attendanceHandler.GetResponses)
			// 管理者による出欠回答の更新（締め切り後も可能）
			r.With(permissionChecker.RequirePermission(tenant.PermissionEditMember)).Put("/{collection_id}/responses", attendanceHandler.AdminUpdateResponse)
		})

		// Schedule API（管理用）
		scheduleRepo := db.NewScheduleRepository(dbPool)
		scheduleHandler := NewScheduleHandler(
			appschedule.NewCreateScheduleUsecase(scheduleRepo, systemClock),
			appschedule.NewSubmitResponseUsecase(scheduleRepo, txManager, systemClock),
			appschedule.NewDecideScheduleUsecase(scheduleRepo, systemClock),
			appschedule.NewCloseScheduleUsecase(scheduleRepo, systemClock),
			appschedule.NewDeleteScheduleUsecase(scheduleRepo, systemClock),
			appschedule.NewUpdateScheduleUsecase(scheduleRepo, systemClock),
			appschedule.NewGetScheduleUsecase(scheduleRepo),
			appschedule.NewGetScheduleByTokenUsecase(scheduleRepo),
			appschedule.NewGetResponsesUsecase(scheduleRepo),
			appschedule.NewListSchedulesUsecase(scheduleRepo),
			appschedule.NewGetAllPublicResponsesUsecase(scheduleRepo, memberRepo),
			appschedule.NewConvertToAttendanceUsecase(scheduleRepo, attendanceRepo, memberGroupRepo, txManager, systemClock),
		)
		r.Route("/schedules", func(r chi.Router) {
			r.Get("/", scheduleHandler.ListSchedules)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateSchedule)).Post("/", scheduleHandler.CreateSchedule)
			r.Get("/{schedule_id}", scheduleHandler.GetSchedule)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateSchedule)).Post("/{schedule_id}/decide", scheduleHandler.DecideSchedule)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateSchedule)).Post("/{schedule_id}/close", scheduleHandler.CloseSchedule)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateSchedule)).Delete("/{schedule_id}", scheduleHandler.DeleteSchedule)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateSchedule)).Put("/{schedule_id}", scheduleHandler.UpdateSchedule)
			r.Get("/{schedule_id}/responses", scheduleHandler.GetResponses)
			r.With(permissionChecker.RequirePermission(tenant.PermissionCreateSchedule)).Post("/{schedule_id}/convert-to-attendance", scheduleHandler.ConvertToAttendance)
		})

		// Invitation API（管理者のみ - マネージャー招待権限が必要）
		r.Route("/invitations", func(r chi.Router) {
			r.With(permissionChecker.RequirePermission(tenant.PermissionInviteManager)).Post("/", invitationHandler.InviteAdmin)
		})

		// Tenant API
		r.Route("/tenants", func(r chi.Router) {
			r.Get("/me", tenantHandler.GetCurrentTenant)
			r.Put("/me", tenantHandler.UpdateCurrentTenant)
		})

		// Admin API (テナント管理者のパスワード変更、メールアドレス変更、PWリセット許可)
		r.Route("/admins", func(r chi.Router) {
			r.Post("/me/change-password", adminHandler.ChangePassword)
			r.Post("/me/change-email", adminHandler.ChangeEmail)
			// PWリセット許可（Ownerのみ実行可能 - Usecase内でチェック）
			r.Post("/{admin_id}/allow-password-reset", authPasswordResetHandler.AllowPasswordReset)
		})

		// ManagerPermissionsHandler dependencies (reusing managerPermissionsRepo)
		managerPermissionsHandler := NewManagerPermissionsHandler(
			apptenant.NewGetManagerPermissionsUsecase(managerPermissionsRepo),
			apptenant.NewUpdateManagerPermissionsUsecase(managerPermissionsRepo),
		)

		// Settings API
		r.Route("/settings", func(r chi.Router) {
			r.Get("/manager-permissions", managerPermissionsHandler.GetManagerPermissions)
			r.Put("/manager-permissions", managerPermissionsHandler.UpdateManagerPermissions)
		})

		// Import API（一括取り込み機能）
		importJobRepo := db.NewImportJobRepository(dbPool)
		importHandler := NewImportHandler(
			appimport.NewImportMembersUsecase(importJobRepo, memberRepo),
			appimport.NewGetImportStatusUsecase(importJobRepo),
			appimport.NewGetImportResultUsecase(importJobRepo),
			appimport.NewListImportJobsUsecase(importJobRepo),
		)
		r.Route("/imports", func(r chi.Router) {
			r.Get("/", importHandler.ListImportJobs)
			r.With(permissionChecker.RequirePermission(tenant.PermissionAddMember)).Post("/members", importHandler.ImportMembers)
			r.Get("/{import_job_id}/status", importHandler.GetImportStatus)
			r.Get("/{import_job_id}/result", importHandler.GetImportResult)
		})

		// Announcement API（お知らせ機能）
		announcementRepo := db.NewAnnouncementRepository(dbPool)
		announcementReadRepo := db.NewAnnouncementReadRepository(dbPool)
		announcementHandler := NewAnnouncementHandler(
			appannouncement.NewListAnnouncementsUsecase(announcementRepo, announcementReadRepo),
			appannouncement.NewGetUnreadCountUsecase(announcementReadRepo),
			appannouncement.NewMarkAsReadUsecase(announcementReadRepo),
			appannouncement.NewMarkAllAsReadUsecase(announcementReadRepo),
		)
		r.Route("/announcements", func(r chi.Router) {
			r.Get("/", announcementHandler.List)
			r.Get("/unread-count", announcementHandler.GetUnreadCount)
			r.Post("/{id}/read", announcementHandler.MarkAsRead)
			r.Post("/read-all", announcementHandler.MarkAllAsRead)
		})

		// Tutorial API（チュートリアル機能）
		tutorialRepo := db.NewTutorialRepository(dbPool)
		tutorialHandler := NewTutorialHandler(
			apptutorial.NewListTutorialsUsecase(tutorialRepo),
			apptutorial.NewGetTutorialUsecase(tutorialRepo),
		)
		r.Route("/tutorials", func(r chi.Router) {
			r.Get("/", tutorialHandler.List)
			r.Get("/{id}", tutorialHandler.Get)
		})

		// Calendar API（カレンダー機能）
		calendarRepo := db.NewCalendarRepository(dbPool)
		calendarEntryRepo := db.NewCalendarEntryRepository(dbPool)
		calendarHandler := NewCalendarHandler(
			appcalendar.NewCreateCalendarUsecase(calendarRepo, eventRepo),
			appcalendar.NewGetCalendarUsecase(calendarRepo, eventRepo, businessDayRepo),
			appcalendar.NewListCalendarsUsecase(calendarRepo),
			appcalendar.NewUpdateCalendarUsecase(calendarRepo, eventRepo),
			appcalendar.NewDeleteCalendarUsecase(calendarRepo),
			appcalendar.NewGetCalendarByTokenUsecase(calendarRepo, eventRepo, businessDayRepo, calendarEntryRepo),
		)
		calendarEntryHandler := NewCalendarEntryHandler(
			appcalendar.NewCreateCalendarEntryUsecase(calendarRepo, calendarEntryRepo),
			appcalendar.NewListCalendarEntriesUsecase(calendarEntryRepo),
			appcalendar.NewUpdateCalendarEntryUsecase(calendarEntryRepo),
			appcalendar.NewDeleteCalendarEntryUsecase(calendarEntryRepo),
		)
		r.Route("/calendars", func(r chi.Router) {
			r.Post("/", calendarHandler.Create)
			r.Get("/", calendarHandler.List)
			r.Get("/{id}", calendarHandler.GetByID)
			r.Put("/{id}", calendarHandler.Update)
			r.Delete("/{id}", calendarHandler.Delete)

			// Calendar Entry routes
			r.Route("/{calendar_id}/entries", func(r chi.Router) {
				r.Post("/", calendarEntryHandler.CreateCalendarEntry)
				r.Get("/", calendarEntryHandler.ListCalendarEntries)
				r.Put("/{entry_id}", calendarEntryHandler.UpdateCalendarEntry)
				r.Delete("/{entry_id}", calendarEntryHandler.DeleteCalendarEntry)
			})
		})

		// Billing API（課金管理 - Stripeカスタマーポータル、課金状態）
		stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
		billingPortalReturnURL := os.Getenv("BILLING_PORTAL_RETURN_URL")
		if billingPortalReturnURL == "" {
			billingPortalReturnURL = "https://vrcshift.com/admin/settings"
		}

		billingSubscriptionRepo := db.NewSubscriptionRepository(dbPool)
		billingStatusUsecase := apppayment.NewBillingStatusUsecase(
			billingSubscriptionRepo,
			entitlementRepo,
		)

		var billingHandler *BillingHandler
		if stripeSecretKey != "" {
			billingStripeClient := infrastripe.NewClient(stripeSecretKey)
			billingPaymentGateway := infrastripe.NewStripePaymentGateway(billingStripeClient)
			billingPortalUsecase := apppayment.NewBillingPortalUsecase(
				billingSubscriptionRepo,
				billingPaymentGateway,
				billingPortalReturnURL,
			)
			billingHandler = NewBillingHandler(billingPortalUsecase, billingStatusUsecase)
		} else {
			billingHandler = NewBillingHandler(nil, billingStatusUsecase)
		}

		r.Route("/billing", func(r chi.Router) {
			// 課金状態取得
			r.Get("/status", billingHandler.GetStatus)
			// カスタマーポータルセッション作成（カード変更、解約など）- Stripe設定時のみ
			if stripeSecretKey != "" {
				r.Post("/portal", billingHandler.CreatePortalSession)
			}
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
		subscriptionRepo := db.NewSubscriptionRepository(dbPool)
		billingAuditLogRepo := db.NewBillingAuditLogRepository(dbPool)

		adminLicenseKeyUsecase := applicense.NewAdminLicenseKeyUsecase(
			txManager,
			licenseKeyRepo,
			billingAuditLogRepo,
		)
		adminTenantUsecase := apptenant.NewAdminTenantUsecase(
			txManager,
			tenantRepo,
			adminRepo,
			entitlementRepo,
			subscriptionRepo,
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

		// Admin Auth (Password Reset Allowance)
		adminAuthClock := &clock.RealClock{}
		adminAllowPasswordResetUsecase := auth.NewAdminAllowPasswordResetUsecase(adminRepo, adminAuthClock)
		adminAuthHandler := NewAdminAuthHandler(adminAllowPasswordResetUsecase)

		r.Route("/admins", func(r chi.Router) {
			r.Post("/{admin_id}/allow-password-reset", adminAuthHandler.AllowPasswordReset)
		})

		// Admin Announcement Management
		adminAnnouncementRepo := db.NewAnnouncementRepository(dbPool)
		adminAnnouncementHandler := NewAdminAnnouncementHandler(
			appannouncement.NewListAllAnnouncementsUsecase(adminAnnouncementRepo),
			appannouncement.NewCreateAnnouncementUsecase(adminAnnouncementRepo),
			appannouncement.NewUpdateAnnouncementUsecase(adminAnnouncementRepo),
			appannouncement.NewDeleteAnnouncementUsecase(adminAnnouncementRepo),
		)
		r.Route("/announcements", func(r chi.Router) {
			r.Get("/", adminAnnouncementHandler.List)
			r.Post("/", adminAnnouncementHandler.Create)
			r.Put("/{id}", adminAnnouncementHandler.Update)
			r.Delete("/{id}", adminAnnouncementHandler.Delete)
		})

		// Admin Tutorial Management
		adminTutorialRepo := db.NewTutorialRepository(dbPool)
		adminTutorialHandler := NewAdminTutorialHandler(
			apptutorial.NewListAllTutorialsUsecase(adminTutorialRepo),
			apptutorial.NewCreateTutorialUsecase(adminTutorialRepo),
			apptutorial.NewUpdateTutorialUsecase(adminTutorialRepo),
			apptutorial.NewDeleteTutorialUsecase(adminTutorialRepo),
		)
		r.Route("/tutorials", func(r chi.Router) {
			r.Get("/", adminTutorialHandler.List)
			r.Post("/", adminTutorialHandler.Create)
			r.Put("/{id}", adminTutorialHandler.Update)
			r.Delete("/{id}", adminTutorialHandler.Delete)
		})

		// Admin System Settings (Release Status Toggle)
		adminSystemSettingRepo := db.NewSystemSettingRepository(dbPool)
		adminSystemUsecase := appsystem.NewUsecase(adminSystemSettingRepo)
		adminSystemHandler := NewSystemHandler(adminSystemUsecase)
		r.Route("/system", func(r chi.Router) {
			r.Get("/release-status", adminSystemHandler.GetReleaseStatusAdmin)
			r.Put("/release-status", adminSystemHandler.UpdateReleaseStatus)
		})
	})

	// Public API（認証不要）
	// Shared dependencies for public handlers
	publicClock := &clock.RealClock{}
	publicTxManager := db.NewPgxTxManager(dbPool)

	// Rate limiters for public attendance/schedules API
	publicReadRL := PublicAPIReadRateLimiter()   // 60 requests/minute/IP for GET
	publicWriteRL := PublicAPIWriteRateLimiter() // 10 requests/minute/IP for POST

	r.Route("/api/v1/public/attendance", func(r chi.Router) {
		publicAttendanceRepoForHandler := db.NewAttendanceRepository(dbPool)
		publicMemberRepoForAttendance := db.NewMemberRepository(dbPool)
		publicRoleRepoForAttendance := db.NewRoleRepository(dbPool)
		publicAttendanceHandler := NewAttendanceHandler(
			appattendance.NewCreateCollectionUsecase(publicAttendanceRepoForHandler, publicRoleRepoForAttendance, publicTxManager, publicClock),
			appattendance.NewSubmitResponseUsecase(publicAttendanceRepoForHandler, publicTxManager, publicClock),
			appattendance.NewCloseCollectionUsecase(publicAttendanceRepoForHandler, publicClock),
			appattendance.NewDeleteCollectionUsecase(publicAttendanceRepoForHandler, publicClock),
			nil,
			appattendance.NewGetCollectionUsecase(publicAttendanceRepoForHandler),
			appattendance.NewGetCollectionByTokenUsecase(publicAttendanceRepoForHandler),
			appattendance.NewGetResponsesUsecase(publicAttendanceRepoForHandler, publicMemberRepoForAttendance),
			appattendance.NewListCollectionsUsecase(publicAttendanceRepoForHandler),
			appattendance.NewGetMemberResponsesUsecase(publicAttendanceRepoForHandler),
			appattendance.NewGetAllPublicResponsesUsecase(publicAttendanceRepoForHandler, publicMemberRepoForAttendance),
			nil, // AdminUpdateResponseUsecase は公開APIでは使用しない
		)
		// GET endpoints: 60 requests/minute/IP
		r.With(RateLimitMiddleware(publicReadRL)).Get("/{token}", publicAttendanceHandler.GetCollectionByToken)
		r.With(RateLimitMiddleware(publicReadRL)).Get("/{token}/members/{member_id}/responses", publicAttendanceHandler.GetMemberResponses)
		r.With(RateLimitMiddleware(publicReadRL)).Get("/{token}/responses", publicAttendanceHandler.GetAllPublicResponses)
		// POST endpoints: 10 requests/minute/IP
		r.With(RateLimitMiddleware(publicWriteRL)).Post("/{token}/responses", publicAttendanceHandler.SubmitResponse)
	})

	r.Route("/api/v1/public/schedules", func(r chi.Router) {
		publicScheduleRepo := db.NewScheduleRepository(dbPool)
		publicScheduleMemberRepo := db.NewMemberRepository(dbPool)
		publicScheduleHandler := NewScheduleHandler(
			appschedule.NewCreateScheduleUsecase(publicScheduleRepo, publicClock),
			appschedule.NewSubmitResponseUsecase(publicScheduleRepo, publicTxManager, publicClock),
			appschedule.NewDecideScheduleUsecase(publicScheduleRepo, publicClock),
			appschedule.NewCloseScheduleUsecase(publicScheduleRepo, publicClock),
			appschedule.NewDeleteScheduleUsecase(publicScheduleRepo, publicClock),
			nil,
			appschedule.NewGetScheduleUsecase(publicScheduleRepo),
			appschedule.NewGetScheduleByTokenUsecase(publicScheduleRepo),
			appschedule.NewGetResponsesUsecase(publicScheduleRepo),
			appschedule.NewListSchedulesUsecase(publicScheduleRepo),
			appschedule.NewGetAllPublicResponsesUsecase(publicScheduleRepo, publicScheduleMemberRepo),
			nil, // ConvertToAttendance は public API では使用しない
		)
		// GET endpoints: 60 requests/minute/IP
		r.With(RateLimitMiddleware(publicReadRL)).Get("/{token}", publicScheduleHandler.GetScheduleByToken)
		r.With(RateLimitMiddleware(publicReadRL)).Get("/{token}/responses", publicScheduleHandler.GetAllPublicResponses)
		// POST endpoints: 10 requests/minute/IP
		r.With(RateLimitMiddleware(publicWriteRL)).Post("/{token}/responses", publicScheduleHandler.SubmitResponse)
	})

	// 公開カレンダーAPI（認証不要）
	r.Route("/api/v1/public/calendar", func(r chi.Router) {
		publicCalendarRepo := db.NewCalendarRepository(dbPool)
		publicEventRepo := db.NewEventRepository(dbPool)
		publicBusinessDayRepo := db.NewEventBusinessDayRepository(dbPool)
		publicCalendarEntryRepo := db.NewCalendarEntryRepository(dbPool)
		publicCalendarHandler := NewCalendarHandler(
			nil, // Create not needed for public handler
			nil, // Get not needed for public handler
			nil, // List not needed for public handler
			nil, // Update not needed for public handler
			nil, // Delete not needed for public handler
			appcalendar.NewGetCalendarByTokenUsecase(publicCalendarRepo, publicEventRepo, publicBusinessDayRepo, publicCalendarEntryRepo),
		)
		r.Get("/{token}", publicCalendarHandler.GetByPublicToken)
	})

	// 公開ページ用メンバー一覧API（認証不要）
	// NOTE: MVPでは簡易実装としてテナントIDを指定してメンバー一覧を取得可能
	// group_ids パラメータで対象グループを指定可能（カンマ区切り）
	// role_ids パラメータで対象ロールを指定可能（カンマ区切り）
	publicMemberRepo := db.NewMemberRepository(dbPool)
	publicMemberRoleRepo := db.NewMemberRoleRepository(dbPool)
	publicAttendanceRepo := db.NewAttendanceRepository(dbPool)
	publicMemberGroupRepo := db.NewMemberGroupRepository(dbPool)
	publicMemberHandler := NewMemberHandler(
		appmember.NewCreateMemberUsecase(publicMemberRepo, publicMemberRoleRepo),
		appmember.NewListMembersUsecase(publicMemberRepo, publicMemberRoleRepo),
		appmember.NewGetMemberUsecase(publicMemberRepo, publicMemberRoleRepo),
		appmember.NewDeleteMemberUsecase(publicMemberRepo),
		appmember.NewUpdateMemberUsecase(publicMemberRepo, publicMemberRoleRepo),
		appmember.NewGetRecentAttendanceUsecase(publicMemberRepo, publicAttendanceRepo),
		appmember.NewBulkImportMembersUsecase(publicMemberRepo, publicMemberRoleRepo),
		nil, // BulkUpdateRoles not needed for public handler
	)
	r.Get("/api/v1/public/members", func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.URL.Query().Get("tenant_id")
		if tenantID == "" {
			RespondBadRequest(w, "tenant_id is required")
			return
		}

		var allowedMemberIDsByGroup map[string]struct{}
		var allowedMemberIDsByRole map[string]struct{}

		// group_ids パラメータの取得（カンマ区切り）
		groupIDsParam := r.URL.Query().Get("group_ids")
		if groupIDsParam != "" {
			groupIDStrs := strings.Split(groupIDsParam, ",")
			allowedMemberIDsByGroup = make(map[string]struct{})

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
					allowedMemberIDsByGroup[mid.String()] = struct{}{}
				}
			}
		}

		// role_ids パラメータの取得（カンマ区切り）
		roleIDsParam := r.URL.Query().Get("role_ids")
		if roleIDsParam != "" {
			roleIDStrs := strings.Split(roleIDsParam, ",")
			allowedMemberIDsByRole = make(map[string]struct{})

			// 各ロールからメンバーIDを取得
			for _, ridStr := range roleIDStrs {
				ridStr = strings.TrimSpace(ridStr)
				if ridStr == "" {
					continue
				}
				rid, err := common.ParseRoleID(ridStr)
				if err != nil {
					continue
				}
				memberIDs, err := publicMemberRoleRepo.FindMemberIDsByRoleID(r.Context(), rid)
				if err != nil {
					continue
				}
				for _, mid := range memberIDs {
					allowedMemberIDsByRole[mid.String()] = struct{}{}
				}
			}
		}

		// グループとロールの両方が指定されている場合は交差（AND条件）
		// どちらか一方のみ指定の場合はそのフィルタを使用
		var allowedMemberIDs map[string]struct{}
		if allowedMemberIDsByGroup != nil && allowedMemberIDsByRole != nil {
			// 両方指定：交差を取る（AND条件）
			allowedMemberIDs = make(map[string]struct{})
			for mid := range allowedMemberIDsByGroup {
				if _, ok := allowedMemberIDsByRole[mid]; ok {
					allowedMemberIDs[mid] = struct{}{}
				}
			}
		} else if allowedMemberIDsByGroup != nil {
			allowedMemberIDs = allowedMemberIDsByGroup
		} else if allowedMemberIDsByRole != nil {
			allowedMemberIDs = allowedMemberIDsByRole
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

	// Subscribe API（Stripe Checkout経由での新規登録、認証不要、レート制限あり）
	r.Route("/api/v1/public/subscribe", func(r chi.Router) {
		// Initialize dependencies for subscribe
		txManager := db.NewPgxTxManager(dbPool)
		subscribeRateLimiter := DefaultClaimRateLimiter() // 同じレート制限を使用

		// Stripe client configuration from environment
		stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
		stripePriceID := os.Getenv("STRIPE_PRICE_SUB_200")
		successURL := os.Getenv("STRIPE_SUCCESS_URL")
		cancelURL := os.Getenv("STRIPE_CANCEL_URL")

		// Default URLs if not configured
		if successURL == "" {
			successURL = "https://vrcshift.com/subscribe/complete"
		}
		if cancelURL == "" {
			cancelURL = "https://vrcshift.com/subscribe/cancel"
		}

		// Only register route if Stripe is configured
		if stripeSecretKey != "" && stripePriceID != "" {
			stripeClient := infrastripe.NewClient(stripeSecretKey)
			paymentGateway := infrastripe.NewStripePaymentGateway(stripeClient)
			subscribeClock := &clock.RealClock{}

			// Read checkout session expiration from environment variable
			// Valid range: 30-1440 minutes (Stripe API constraint)
			checkoutExpireMinutes := 0 // 0 means use default (24 hours)
			if envExpire := os.Getenv("CHECKOUT_SESSION_EXPIRE_MINUTES"); envExpire != "" {
				if minutes, err := strconv.Atoi(envExpire); err == nil {
					if minutes >= services.MinCheckoutExpireMinutes && minutes <= services.MaxCheckoutExpireMinutes {
						checkoutExpireMinutes = minutes
						slog.Info("Checkout session expiration configured from environment", "minutes", minutes)
					} else {
						slog.Warn("CHECKOUT_SESSION_EXPIRE_MINUTES out of valid range, using default",
							"value", minutes,
							"validRange", "30-1440")
					}
				} else {
					slog.Warn("Invalid CHECKOUT_SESSION_EXPIRE_MINUTES format, using default", "value", envExpire)
				}
			}

			subscribeUsecase := apppayment.NewSubscribeUsecase(
				txManager,
				tenantRepo,
				adminRepo,
				passwordHasher,
				paymentGateway,
				subscribeClock,
				successURL,
				cancelURL,
				stripePriceID,
				checkoutExpireMinutes,
			)
			subscribeHandler := NewSubscribeHandler(subscribeUsecase, subscribeRateLimiter)

			r.Post("/", subscribeHandler.Subscribe)
		}
	})

	// System Settings Public API（認証不要）
	// リリース状態などのシステム設定を公開
	r.Route("/api/v1/public/system", func(r chi.Router) {
		systemSettingRepo := db.NewSystemSettingRepository(dbPool)
		systemUsecase := appsystem.NewUsecase(systemSettingRepo)
		systemHandler := NewSystemHandler(systemUsecase)

		r.Get("/release-status", systemHandler.GetReleaseStatus)
	})

	// Stripe Webhook API（認証不要、署名検証のみ）
	r.Route("/api/v1/stripe", func(r chi.Router) {
		// Initialize dependencies for Stripe webhook
		txManager := db.NewPgxTxManager(dbPool)
		subscriptionRepo := db.NewSubscriptionRepository(dbPool)
		webhookEventRepo := db.NewWebhookEventRepository(dbPool)
		billingAuditLogRepo := db.NewBillingAuditLogRepository(dbPool)

		// Read grace period from environment variable (default: 14 days, max: 90 days)
		const maxGracePeriodDays = 90
		gracePeriodDays := tenant.DefaultGracePeriodDays
		if envGracePeriod := os.Getenv("GRACE_PERIOD_DAYS"); envGracePeriod != "" {
			if days, err := strconv.Atoi(envGracePeriod); err == nil && days > 0 && days <= maxGracePeriodDays {
				gracePeriodDays = days
				slog.Info("Grace period configured from environment", "days", days)
			} else {
				slog.Warn("Invalid GRACE_PERIOD_DAYS value, using default",
					"value", envGracePeriod,
					"default", tenant.DefaultGracePeriodDays,
					"validRange", "1-90")
			}
		}

		stripeWebhookUsecase := apppayment.NewStripeWebhookUsecase(
			txManager,
			tenantRepo,
			subscriptionRepo,
			entitlementRepo,
			webhookEventRepo,
			billingAuditLogRepo,
			gracePeriodDays,
		)
		stripeWebhookHandler := NewStripeWebhookHandler(stripeWebhookUsecase)

		r.Post("/webhook", stripeWebhookHandler.HandleWebhook)
	})

	return r
}
