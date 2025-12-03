package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewRouter creates a new HTTP router with all routes configured
func NewRouter(db *pgxpool.Pool) http.Handler {
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

	// API v1 ルート
	r.Route("/api/v1", func(r chi.Router) {
		// 認証ミドルウェアを適用
		r.Use(Auth)

		// ハンドラの初期化
		eventHandler := NewEventHandler(db)
		businessDayHandler := NewBusinessDayHandler(db)
		memberHandler := NewMemberHandler(db)
		shiftSlotHandler := NewShiftSlotHandler(db)
		shiftAssignmentHandler := NewShiftAssignmentHandler(db)

		// Event API
		r.Route("/events", func(r chi.Router) {
			r.Post("/", eventHandler.CreateEvent)
			r.Get("/", eventHandler.ListEvents)
			r.Get("/{event_id}", eventHandler.GetEvent)

			// Event配下のBusinessDay
			r.Post("/{event_id}/business-days", businessDayHandler.CreateBusinessDay)
			r.Get("/{event_id}/business-days", businessDayHandler.ListBusinessDays)
		})

		// BusinessDay API
		r.Route("/business-days", func(r chi.Router) {
			r.Get("/{business_day_id}", businessDayHandler.GetBusinessDay)

			// BusinessDay配下のShiftSlot
			r.Post("/{business_day_id}/shift-slots", shiftSlotHandler.CreateShiftSlot)
			r.Get("/{business_day_id}/shift-slots", shiftSlotHandler.GetShiftSlots)
		})

		// Member API
		r.Route("/members", func(r chi.Router) {
			r.Post("/", memberHandler.CreateMember)
			r.Get("/", memberHandler.GetMembers)
			r.Get("/{member_id}", memberHandler.GetMemberDetail)
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
		})
	})

	return r
}

