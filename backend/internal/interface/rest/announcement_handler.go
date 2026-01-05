package rest

import (
	"encoding/json"
	"net/http"

	"github.com/erenoa/vrc-shift-scheduler/backend/internal/app/announcement"
	"github.com/erenoa/vrc-shift-scheduler/backend/internal/domain/common"
	"github.com/go-chi/chi/v5"
)

type AnnouncementHandler struct {
	listUC        *announcement.ListAnnouncementsUsecase
	getUnreadUC   *announcement.GetUnreadCountUsecase
	markReadUC    *announcement.MarkAsReadUsecase
	markAllReadUC *announcement.MarkAllAsReadUsecase
}

func NewAnnouncementHandler(
	listUC *announcement.ListAnnouncementsUsecase,
	getUnreadUC *announcement.GetUnreadCountUsecase,
	markReadUC *announcement.MarkAsReadUsecase,
	markAllReadUC *announcement.MarkAllAsReadUsecase,
) *AnnouncementHandler {
	return &AnnouncementHandler{
		listUC:        listUC,
		getUnreadUC:   getUnreadUC,
		markReadUC:    markReadUC,
		markAllReadUC: markAllReadUC,
	}
}

// List handles GET /api/v1/announcements
func (h *AnnouncementHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondForbidden(w, "Unauthorized")
		return
	}
	adminID, ok := GetAdminID(ctx)
	if !ok {
		RespondForbidden(w, "Unauthorized")
		return
	}

	announcements, err := h.listUC.Execute(ctx, common.TenantID(tenantID), common.AdminID(adminID))
	if err != nil {
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"announcements": announcements,
	}})
}

// GetUnreadCount handles GET /api/v1/announcements/unread-count
func (h *AnnouncementHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondForbidden(w, "Unauthorized")
		return
	}
	adminID, ok := GetAdminID(ctx)
	if !ok {
		RespondForbidden(w, "Unauthorized")
		return
	}

	count, err := h.getUnreadUC.Execute(ctx, common.AdminID(adminID), common.TenantID(tenantID))
	if err != nil {
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"count": count,
	}})
}

// MarkAsRead handles POST /api/v1/announcements/:id/read
func (h *AnnouncementHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	adminID, ok := GetAdminID(ctx)
	if !ok {
		RespondForbidden(w, "Unauthorized")
		return
	}
	announcementID := chi.URLParam(r, "id")

	err := h.markReadUC.Execute(ctx, announcementID, common.AdminID(adminID))
	if err != nil {
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"success": true,
	}})
}

// MarkAllAsRead handles POST /api/v1/announcements/read-all
func (h *AnnouncementHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondForbidden(w, "Unauthorized")
		return
	}
	adminID, ok := GetAdminID(ctx)
	if !ok {
		RespondForbidden(w, "Unauthorized")
		return
	}

	err := h.markAllReadUC.Execute(ctx, common.AdminID(adminID), common.TenantID(tenantID))
	if err != nil {
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"success": true,
	}})
}

// AdminAnnouncementHandler handles admin operations
type AdminAnnouncementHandler struct {
	listAllUC *announcement.ListAllAnnouncementsUsecase
	createUC  *announcement.CreateAnnouncementUsecase
	updateUC  *announcement.UpdateAnnouncementUsecase
	deleteUC  *announcement.DeleteAnnouncementUsecase
}

func NewAdminAnnouncementHandler(
	listAllUC *announcement.ListAllAnnouncementsUsecase,
	createUC *announcement.CreateAnnouncementUsecase,
	updateUC *announcement.UpdateAnnouncementUsecase,
	deleteUC *announcement.DeleteAnnouncementUsecase,
) *AdminAnnouncementHandler {
	return &AdminAnnouncementHandler{
		listAllUC: listAllUC,
		createUC:  createUC,
		updateUC:  updateUC,
		deleteUC:  deleteUC,
	}
}

// List handles GET /api/v1/admin/announcements
func (h *AdminAnnouncementHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	announcements, err := h.listAllUC.Execute(ctx)
	if err != nil {
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"announcements": announcements,
	}})
}

// Create handles POST /api/v1/admin/announcements
func (h *AdminAnnouncementHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req announcement.CreateAnnouncementInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	result, err := h.createUC.Execute(ctx, req)
	if err != nil {
		RespondBadRequest(w, err.Error())
		return
	}

	RespondJSON(w, http.StatusCreated, SuccessResponse{Data: map[string]interface{}{
		"announcement": result,
	}})
}

// Update handles PUT /api/v1/admin/announcements/:id
func (h *AdminAnnouncementHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	var req announcement.UpdateAnnouncementInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}
	req.ID = id

	result, err := h.updateUC.Execute(ctx, req)
	if err != nil {
		RespondBadRequest(w, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"announcement": result,
	}})
}

// Delete handles DELETE /api/v1/admin/announcements/:id
func (h *AdminAnnouncementHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if err := h.deleteUC.Execute(ctx, id); err != nil {
		RespondInternalError(w)
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{Data: map[string]interface{}{
		"success": true,
	}})
}
