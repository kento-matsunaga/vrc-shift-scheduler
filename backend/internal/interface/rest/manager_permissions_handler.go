package rest

import (
	"encoding/json"
	"net/http"

	apptenant "github.com/erenoa/vrc-shift-scheduler/backend/internal/app/tenant"
)

// ManagerPermissionsHandler handles manager permissions HTTP requests
type ManagerPermissionsHandler struct {
	getPermissionsUC    *apptenant.GetManagerPermissionsUsecase
	updatePermissionsUC *apptenant.UpdateManagerPermissionsUsecase
}

// NewManagerPermissionsHandler creates a new ManagerPermissionsHandler
func NewManagerPermissionsHandler(
	getPermissionsUC *apptenant.GetManagerPermissionsUsecase,
	updatePermissionsUC *apptenant.UpdateManagerPermissionsUsecase,
) *ManagerPermissionsHandler {
	return &ManagerPermissionsHandler{
		getPermissionsUC:    getPermissionsUC,
		updatePermissionsUC: updatePermissionsUC,
	}
}

// ManagerPermissionsResponse represents the response for manager permissions
type ManagerPermissionsResponse struct {
	CanAddMember        bool `json:"can_add_member"`
	CanEditMember       bool `json:"can_edit_member"`
	CanDeleteMember     bool `json:"can_delete_member"`
	CanCreateEvent      bool `json:"can_create_event"`
	CanEditEvent        bool `json:"can_edit_event"`
	CanDeleteEvent      bool `json:"can_delete_event"`
	CanAssignShift      bool `json:"can_assign_shift"`
	CanEditShift        bool `json:"can_edit_shift"`
	CanCreateAttendance bool `json:"can_create_attendance"`
	CanCreateSchedule   bool `json:"can_create_schedule"`
	CanManageRoles      bool `json:"can_manage_roles"`
	CanManagePositions  bool `json:"can_manage_positions"`
	CanManageGroups     bool `json:"can_manage_groups"`
	CanInviteManager    bool `json:"can_invite_manager"`
}

// UpdateManagerPermissionsRequest represents the request body for updating permissions
type UpdateManagerPermissionsRequest struct {
	CanAddMember        bool `json:"can_add_member"`
	CanEditMember       bool `json:"can_edit_member"`
	CanDeleteMember     bool `json:"can_delete_member"`
	CanCreateEvent      bool `json:"can_create_event"`
	CanEditEvent        bool `json:"can_edit_event"`
	CanDeleteEvent      bool `json:"can_delete_event"`
	CanAssignShift      bool `json:"can_assign_shift"`
	CanEditShift        bool `json:"can_edit_shift"`
	CanCreateAttendance bool `json:"can_create_attendance"`
	CanCreateSchedule   bool `json:"can_create_schedule"`
	CanManageRoles      bool `json:"can_manage_roles"`
	CanManagePositions  bool `json:"can_manage_positions"`
	CanManageGroups     bool `json:"can_manage_groups"`
	CanInviteManager    bool `json:"can_invite_manager"`
}

// GetManagerPermissions handles GET /api/v1/settings/manager-permissions
func (h *ManagerPermissionsHandler) GetManagerPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	input := apptenant.GetManagerPermissionsInput{
		TenantID: tenantID,
	}

	output, err := h.getPermissionsUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	response := ManagerPermissionsResponse{
		CanAddMember:        output.CanAddMember,
		CanEditMember:       output.CanEditMember,
		CanDeleteMember:     output.CanDeleteMember,
		CanCreateEvent:      output.CanCreateEvent,
		CanEditEvent:        output.CanEditEvent,
		CanDeleteEvent:      output.CanDeleteEvent,
		CanAssignShift:      output.CanAssignShift,
		CanEditShift:        output.CanEditShift,
		CanCreateAttendance: output.CanCreateAttendance,
		CanCreateSchedule:   output.CanCreateSchedule,
		CanManageRoles:      output.CanManageRoles,
		CanManagePositions:  output.CanManagePositions,
		CanManageGroups:     output.CanManageGroups,
		CanInviteManager:    output.CanInviteManager,
	}

	RespondSuccess(w, response)
}

// UpdateManagerPermissions handles PUT /api/v1/settings/manager-permissions
func (h *ManagerPermissionsHandler) UpdateManagerPermissions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID, ok := GetTenantID(ctx)
	if !ok {
		RespondBadRequest(w, "tenant_id is required")
		return
	}

	// ownerのみ更新可能
	role, ok := GetRole(ctx)
	if !ok || role != "owner" {
		RespondError(w, http.StatusForbidden, "ERR_FORBIDDEN", "オーナーのみがマネージャー権限を変更できます", nil)
		return
	}

	var req UpdateManagerPermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondBadRequest(w, "Invalid request body")
		return
	}

	input := apptenant.UpdateManagerPermissionsInput{
		TenantID:            tenantID,
		CanAddMember:        req.CanAddMember,
		CanEditMember:       req.CanEditMember,
		CanDeleteMember:     req.CanDeleteMember,
		CanCreateEvent:      req.CanCreateEvent,
		CanEditEvent:        req.CanEditEvent,
		CanDeleteEvent:      req.CanDeleteEvent,
		CanAssignShift:      req.CanAssignShift,
		CanEditShift:        req.CanEditShift,
		CanCreateAttendance: req.CanCreateAttendance,
		CanCreateSchedule:   req.CanCreateSchedule,
		CanManageRoles:      req.CanManageRoles,
		CanManagePositions:  req.CanManagePositions,
		CanManageGroups:     req.CanManageGroups,
		CanInviteManager:    req.CanInviteManager,
	}

	output, err := h.updatePermissionsUC.Execute(ctx, input)
	if err != nil {
		RespondDomainError(w, err)
		return
	}

	response := ManagerPermissionsResponse{
		CanAddMember:        output.CanAddMember,
		CanEditMember:       output.CanEditMember,
		CanDeleteMember:     output.CanDeleteMember,
		CanCreateEvent:      output.CanCreateEvent,
		CanEditEvent:        output.CanEditEvent,
		CanDeleteEvent:      output.CanDeleteEvent,
		CanAssignShift:      output.CanAssignShift,
		CanEditShift:        output.CanEditShift,
		CanCreateAttendance: output.CanCreateAttendance,
		CanCreateSchedule:   output.CanCreateSchedule,
		CanManageRoles:      output.CanManageRoles,
		CanManagePositions:  output.CanManagePositions,
		CanManageGroups:     output.CanManageGroups,
		CanInviteManager:    output.CanInviteManager,
	}

	RespondSuccess(w, response)
}
